package templates

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func TestLanguageAssetSetsMatch(t *testing.T) {
	enFiles := languageAssetSet(t, "en")
	ruFiles := languageAssetSet(t, "ru")

	if len(enFiles) == 0 {
		t.Fatal("expected English language assets to be non-empty")
	}
	if len(ruFiles) == 0 {
		t.Fatal("expected Russian language assets to be non-empty")
	}

	if !reflect.DeepEqual(enFiles, ruFiles) {
		t.Fatalf("language asset sets differ\n\nen: %v\n\nru: %v", enFiles, ruFiles)
	}
}

func TestFilesBuildForSupportedLanguages(t *testing.T) {
	testCases := []struct {
		name     string
		settings LanguageSettings
	}{
		{
			name: "english",
			settings: LanguageSettings{
				Default:  "en",
				Docs:     "en",
				Agent:    "en",
				Comments: "en",
				Shell:    "sh",
			},
		},
		{
			name: "russian",
			settings: LanguageSettings{
				Default:  "ru",
				Docs:     "ru",
				Agent:    "ru",
				Comments: "ru",
				Shell:    "sh",
			},
		},
		{
			name: "mixed",
			settings: LanguageSettings{
				Default:  "en",
				Docs:     "ru",
				Agent:    "en",
				Comments: "ru",
				Shell:    "sh",
			},
		},
		{
			name: "powershell",
			settings: LanguageSettings{
				Default:  "en",
				Docs:     "en",
				Agent:    "en",
				Comments: "en",
				Shell:    "powershell",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files, err := Files(tc.settings)
			if err != nil {
				t.Fatalf("Files() returned error: %v", err)
			}
			if len(files) == 0 {
				t.Fatal("expected generated file set to be non-empty")
			}

			targets := make(map[string]struct{}, len(files))
			for _, file := range files {
				if file.TargetPath == "" {
					t.Fatal("expected file target path to be non-empty")
				}
				if file.Content == "" {
					t.Fatalf("expected generated content for %s to be non-empty", file.TargetPath)
				}
				targets[file.TargetPath] = struct{}{}
			}

			requiredFiles := []string{
				"speckeep.yaml",
				"constitution.md",
				"templates/spec.md",
				"templates/plan.md",
				"templates/research.md",
				"templates/tasks.md",
				"templates/data-model.md",
				"templates/inspect-report.md",
				"templates/verify-report.md",
				"templates/archive/summary.md",
				"templates/prompts/spec.md",
				"templates/prompts/inspect.md",
				"templates/prompts/plan.md",
				"templates/prompts/tasks.md",
				"templates/prompts/implement.md",
				"templates/prompts/archive.md",
				"templates/prompts/verify.md",
			}
			ext := ".sh"
			if tc.settings.Shell == "powershell" {
				ext = ".ps1"
			}
			requiredFiles = append(requiredFiles,
				"scripts/run-speckeep"+ext,
				"scripts/check-inspect-ready"+ext,
				"scripts/check-archive-ready"+ext,
				"scripts/check-verify-ready"+ext,
				"scripts/verify-task-state"+ext,
			)
			for _, required := range requiredFiles {
				if _, ok := targets[required]; !ok {
					t.Fatalf("expected generated file set to include %s", required)
				}
			}
		})
	}
}

func TestInspectPromptDefinesCheapScopeAndVerdictRules(t *testing.T) {
	files, err := Files(LanguageSettings{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
		Shell:    "sh",
	})
	if err != nil {
		t.Fatalf("Files() returned error: %v", err)
	}

	content := fileContentByTarget(t, files, "templates/prompts/inspect.md")
	requiredSnippets := []string{
		"Always read these first:",
		"Read these only when they exist and the inspection requires cross-artifact consistency checks",
		"Do Not Read By Default",
		"Prefer the cheapest inspection scope first",
		"Default to a compact report in conversation output",
		"Produce the full sectioned report only when the user explicitly asks for a full report",
		"Verify `constitution <-> spec`",
		"Treat technology names, framework choices, library lists, or version pins in the spec as a `Warning` unless they clearly represent a user requirement, repository constraint, or external compatibility contract.",
		"Verify `spec <-> plan`",
		"verify `plan <-> tasks`",
		"The `## Verdict` section MUST use one of: `pass`, `concerns`, `blocked`.",
		"machine-readable metadata block",
		"major `spec <-> plan` contradictions",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected inspect prompt to contain %q", snippet)
		}
	}
}

func TestReportTemplatesIncludeMetadataFrontmatter(t *testing.T) {
	files, err := Files(LanguageSettings{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
		Shell:    "sh",
	})
	if err != nil {
		t.Fatalf("Files() returned error: %v", err)
	}

	testCases := []struct {
		target string
		want   []string
	}{
		{
			target: "templates/inspect-report.md",
			want: []string{
				"report_type: inspect",
				"slug: <slug>",
				"status: pass",
				"generated_at: <YYYY-MM-DD>",
			},
		},
		{
			target: "templates/verify-report.md",
			want: []string{
				"report_type: verify",
				"slug: <slug>",
				"status: pass",
				"generated_at: <YYYY-MM-DD>",
			},
		},
	}

	for _, tc := range testCases {
		content := fileContentByTarget(t, files, tc.target)
		for _, snippet := range tc.want {
			if !strings.Contains(content, snippet) {
				t.Fatalf("expected %s to contain %q", tc.target, snippet)
			}
		}
	}
}

func TestGeneratedAgentSnippetMentionsDraftspecLauncher(t *testing.T) {
	files, err := Files(LanguageSettings{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
		Shell:    "powershell",
	})
	if err != nil {
		t.Fatalf("Files() returned error: %v", err)
	}

	content := fileContentByTarget(t, files, "templates/agents-snippet.md")
	requiredSnippets := []string{
		"./.speckeep/scripts/run-speckeep.ps1",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected agents snippet to contain %q", snippet)
		}
	}
}

func TestCoreTemplatesPreferDetailedButTightArtifacts(t *testing.T) {
	files, err := Files(LanguageSettings{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
		Shell:    "sh",
	})
	if err != nil {
		t.Fatalf("Files() returned error: %v", err)
	}

	testCases := []struct {
		target string
		want   []string
	}{
		{
			target: "templates/spec.md",
			want: []string{
				"## Primary User Flow",
				"who benefits, what changes for them",
				"Evidence:",
			},
		},
		{
			target: "templates/plan.md",
			want: []string{
				"Tradeoff:",
				"## Rollout and Compatibility",
				"proof that the result will be observable",
			},
		},
		{
			target: "templates/tasks.md",
			want: []string{
				"Goal: establish the minimum structure",
				"Goal: deliver the primary feature behavior",
				"Separate validation work from broad implementation work",
			},
		},
		{
			target: "templates/data-model.md",
			want: []string{
				"Source of truth:",
				"Failure or consistency notes:",
				"## State Transitions",
			},
		},
	}

	for _, tc := range testCases {
		content := fileContentByTarget(t, files, tc.target)
		for _, snippet := range tc.want {
			if !strings.Contains(content, snippet) {
				t.Fatalf("expected %s to contain %q", tc.target, snippet)
			}
		}
	}
}

func TestImplementPromptSupportsFullRunAndScopedExecution(t *testing.T) {
	files, err := Files(LanguageSettings{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
		Shell:    "sh",
	})
	if err != nil {
		t.Fatalf("Files() returned error: %v", err)
	}

	content := fileContentByTarget(t, files, "templates/prompts/implement.md")
	requiredSnippets := []string{
		"Default behavior: if the user does not restrict scope, execute only the first unfinished phase.",
		"`--phase <number>`: execute only the specified phase.",
		"`--tasks <task-id-list>`: execute only the specified task IDs.",
		"Do not accept `--phase` and `--tasks` together in the same run.",
		"`--continue`: resume mode",
		"the selected work would force changes across another feature package or slug",
		"the next safe step would require inventing new tasks or acceptance coverage",
		"Leave the feature in a state that the next verify pass can inspect without guessing",
		"Before marking a task done, confirm that the observable outcome named in the task text is actually present.",
		"[T1.1] started",
		"[T1.1] done",
		"[T1.1] blocked: <reason>",
		"[Phase 1] done: T1.1, T1.2",
		"do not claim coverage that was not implemented",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected implement prompt to contain %q", snippet)
		}
	}
}

func TestSpecPromptDefinesDeterministicStagedMode(t *testing.T) {
	files, err := Files(LanguageSettings{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
		Shell:    "sh",
	})
	if err != nil {
		t.Fatalf("Files() returned error: %v", err)
	}

	content := fileContentByTarget(t, files, "templates/prompts/spec.md")
	requiredSnippets := []string{
		"keep staged mode active for the next non-command user message",
		"If the next user message begins with `/speckeep.`, staged mode is canceled",
		"If the next user message does not begin with `/speckeep.`, treat it as the continuation of the staged spec request.",
		"The spec should be detailed enough that both an agent and a human reviewer can understand the user flow",
		"`## Primary User Flow` should describe the main path in 3-5 concrete steps",
		"prefer a tiny structured clarify pass instead of a broad open-ended interview",
		"`## Change Delta` should make it obvious what becomes newly possible, what changes, and what stays unchanged.",
		"Prefer density over length",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected spec prompt to contain %q", snippet)
		}
	}
}

func TestPlanPromptDefinesConcreteResearchTriggers(t *testing.T) {
	files, err := Files(LanguageSettings{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
		Shell:    "sh",
	})
	if err != nil {
		t.Fatalf("Files() returned error: %v", err)
	}

	content := fileContentByTarget(t, files, "templates/prompts/plan.md")
	requiredSnippets := []string{
		"Create `.speckeep/specs/<slug>/plan/research.md` only when at least one of these is true:",
		"external system, API, or dependency",
		"multiple realistic implementation options",
		"Before creating `research.md`, write down the concrete unknowns first:",
		"Do not create `research.md` for generic brainstorming",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected plan prompt to contain %q", snippet)
		}
	}
}

func TestTasksAndImplementPromptsDoNotAssumeResearchArtifact(t *testing.T) {
	files, err := Files(LanguageSettings{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
		Shell:    "sh",
	})
	if err != nil {
		t.Fatalf("Files() returned error: %v", err)
	}

	for _, target := range []string{
		"templates/prompts/tasks.md",
		"templates/prompts/implement.md",
	} {
		content := fileContentByTarget(t, files, target)
		if !strings.Contains(content, "Do not assume `research.md` should exist;") {
			t.Fatalf("expected %s to explain that research.md is not assumed by default", target)
		}
	}
}

func TestPlanAndTasksPromptsReinforceDetailedButTightArtifacts(t *testing.T) {
	files, err := Files(LanguageSettings{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
		Shell:    "sh",
	})
	if err != nil {
		t.Fatalf("Files() returned error: %v", err)
	}

	testCases := []struct {
		target string
		want   []string
	}{
		{
			target: "templates/prompts/plan.md",
			want: []string{
				"The plan should be specific enough that both an agent and a human reviewer can see the intended implementation shape",
				"Each significant `DEC-*` should capture `Why`, `Tradeoff`, `Affects`, and `Validation`.",
				"Add a short `Unknowns First` pass before finalizing the plan",
				"`## Rollout and Compatibility` should be explicit",
				"Record technologies, libraries, framework choices, or version constraints only when they materially affect",
			},
		},
		{
			target: "templates/prompts/tasks.md",
			want: []string{
				"The task list should be readable to both an implementation agent and a human reviewer",
				"Each phase should have a short goal",
				"Touches:`",
				"Could another developer execute these tasks in order without guessing what `done` means",
			},
		},
	}

	for _, tc := range testCases {
		content := fileContentByTarget(t, files, tc.target)
		for _, snippet := range tc.want {
			if !strings.Contains(content, snippet) {
				t.Fatalf("expected %s to contain %q", tc.target, snippet)
			}
		}
	}
}

func TestSpecAndPlanSeparateProductIntentFromTechChoices(t *testing.T) {
	files, err := Files(LanguageSettings{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
		Shell:    "sh",
	})
	if err != nil {
		t.Fatalf("Files() returned error: %v", err)
	}

	specContent := fileContentByTarget(t, files, "templates/prompts/spec.md")
	for _, snippet := range []string{
		"Do not lock in technologies, libraries, framework choices, or version details by default.",
		"If a technology choice matters only as an implementation preference, record it in `plan`, not in `spec`.",
		"do not add library lists, framework choices, SDK names, or version pins to the spec unless they are product or repository constraints",
	} {
		if !strings.Contains(specContent, snippet) {
			t.Fatalf("expected spec prompt to contain %q", snippet)
		}
	}

	planContent := fileContentByTarget(t, files, "templates/prompts/plan.md")
	for _, snippet := range []string{
		"Record technologies, libraries, framework choices, or version constraints only when they materially affect implementation shape, integration boundaries, validation, or risk.",
		"If a version or dependency is named, explain why it matters for this feature",
		"Do not enumerate stack details for completeness; capture only technical constraints that reduce downstream guesswork.",
	} {
		if !strings.Contains(planContent, snippet) {
			t.Fatalf("expected plan prompt to contain %q", snippet)
		}
	}
}

func TestPromptsDefineScopeTripwiresForRefinement(t *testing.T) {
	files, err := Files(LanguageSettings{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
		Shell:    "sh",
	})
	if err != nil {
		t.Fatalf("Files() returned error: %v", err)
	}

	testCases := []struct {
		target string
		want   []string
	}{
		{
			target: "templates/prompts/spec.md",
			want: []string{
				"multiple feature slugs or multiple independent specs",
			},
		},
		{
			target: "templates/prompts/plan.md",
			want: []string{
				"cross an unclear integration or architectural boundary",
				"multiple feature packages were planned together",
			},
		},
		{
			target: "templates/prompts/tasks.md",
			want: []string{
				"span multiple feature slugs or unrelated change sets",
				"cannot be mapped to executable work without guessing",
			},
		},
		{
			target: "templates/prompts/verify.md",
			want: []string{
				"broad repository sweep instead of focused evidence",
				"cannot be confirmed from the current tasks, plan artifacts, and targeted code inspection",
			},
		},
	}

	for _, tc := range testCases {
		content := fileContentByTarget(t, files, tc.target)
		for _, snippet := range tc.want {
			if !strings.Contains(content, snippet) {
				t.Fatalf("expected %s to contain %q", tc.target, snippet)
			}
		}
	}
}

func TestVerifyTemplateAndPromptPreferEvidenceScopedVerification(t *testing.T) {
	files, err := Files(LanguageSettings{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
		Shell:    "sh",
	})
	if err != nil {
		t.Fatalf("Files() returned error: %v", err)
	}

	templateContent := fileContentByTarget(t, files, "templates/verify-report.md")
	for _, snippet := range []string{
		"verification_mode: default | deep",
		"archive_readiness: safe",
		"acceptance_evidence:",
		"implementation_alignment:",
		"## Not Verified",
	} {
		if !strings.Contains(templateContent, snippet) {
			t.Fatalf("expected verify report template to contain %q", snippet)
		}
	}

	promptContent := fileContentByTarget(t, files, "templates/prompts/verify.md")
	for _, snippet := range []string{
		"Treat verify as an evidence log, not a reassurance ritual.",
		"Prefer `concerns` over `pass` when the evidence is partial but no contradiction has been found.",
		"`acceptance_evidence` for the `AC-*` items you actually confirmed",
		"`## Not Verified`",
		"Keep claims scoped.",
		"send the feature back to the narrowest earlier phase that can honestly fix it",
		"Do not use `pass` unless the completed task state is confirmed",
	} {
		if !strings.Contains(promptContent, snippet) {
			t.Fatalf("expected verify prompt to contain %q", snippet)
		}
	}
}

func TestPhasePromptsIncludeExplicitNextCommandGuidance(t *testing.T) {
	files, err := Files(LanguageSettings{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
		Shell:    "sh",
	})
	if err != nil {
		t.Fatalf("Files() returned error: %v", err)
	}

	testCases := []struct {
		target string
		want   []string
	}{
		{
			target: "templates/prompts/spec.md",
			want: []string{
				"Ready for: /speckeep.inspect <slug>",
				"`Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`",
			},
		},
		{
			target: "templates/prompts/plan.md",
			want: []string{
				"Ready for: /speckeep.tasks <slug>",
				"`Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`",
			},
		},
		{
			target: "templates/prompts/tasks.md",
			want: []string{
				"Ready for: /speckeep.implement <slug>",
				"`Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`",
			},
		},
		{
			target: "templates/prompts/implement.md",
			want: []string{
				"execute only the first unfinished phase",
				"`Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`",
				"Ready for: /speckeep.verify <slug>",
				"do not claim coverage that was not implemented",
			},
		},
		{
			target: "templates/prompts/verify.md",
			want: []string{
				"`Slug`, `Status`, `Artifacts`, `Blockers`, and either `Ready for` or `Return to`",
				"Return to: /speckeep.<phase> <slug>",
				"Ready for: /speckeep.archive <slug>",
				"name it explicitly with its slash command",
			},
		},
		{
			target: "templates/prompts/archive.md",
			want: []string{
				"terminal workflow step for this feature",
			},
		},
	}

	for _, tc := range testCases {
		content := fileContentByTarget(t, files, tc.target)
		for _, snippet := range tc.want {
			if !strings.Contains(content, snippet) {
				t.Fatalf("expected %s to contain %q", tc.target, snippet)
			}
		}
	}
}

func TestPromptsEnforcePhaseBoundaries(t *testing.T) {
	files, err := Files(LanguageSettings{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
		Shell:    "sh",
	})
	if err != nil {
		t.Fatalf("Files() returned error: %v", err)
	}

	testCases := []struct {
		target string
		want   []string
	}{
		{
			target: "templates/prompts/spec.md",
			want: []string{
				"Do not write planning decisions, task decomposition, or implementation steps in the spec itself.",
			},
		},
		{
			target: "templates/prompts/plan.md",
			want: []string{
				"Do not write the task checklist, edit implementation code, or emit verify/archive conclusions during planning.",
			},
		},
		{
			target: "templates/prompts/tasks.md",
			want: []string{
				"Do not start implementation work, edit source code, or claim tasks are already done during the tasks phase.",
			},
		},
		{
			target: "templates/prompts/implement.md",
			want: []string{
				"Do not re-plan the feature, emit a verify verdict, or silently complete neighboring tasks",
			},
		},
		{
			target: "templates/prompts/inspect.md",
			want: []string{
				"For `blocked`, do not suggest the next phase command; state which refinement is required first.",
			},
		},
		{
			target: "templates/prompts/verify.md",
			want: []string{
				"For `blocked`, do not suggest archive; end with `Return to: /speckeep.<phase> <slug>`",
			},
		},
	}

	for _, tc := range testCases {
		content := fileContentByTarget(t, files, tc.target)
		for _, snippet := range tc.want {
			if !strings.Contains(content, snippet) {
				t.Fatalf("expected %s to contain %q", tc.target, snippet)
			}
		}
	}
}

func TestInspectHelperScriptsDelegateToInternalCLI(t *testing.T) {
	testCases := []struct {
		name   string
		shell  string
		target string
		want   []string
	}{
		{name: "sh", shell: "sh", target: "scripts/inspect-spec.sh", want: []string{"run-speckeep.sh", "__internal inspect-spec --root \"$ROOT_DIR\""}},
		{name: "powershell", shell: "powershell", target: "scripts/inspect-spec.ps1", want: []string{"run-speckeep.ps1", "__internal inspect-spec --root $RootDir"}},
		{name: "sh-archive", shell: "sh", target: "scripts/archive-feature.sh", want: []string{"run-speckeep.sh", "archive --root \"$ROOT_DIR\""}},
		{name: "powershell-archive", shell: "powershell", target: "scripts/archive-feature.ps1", want: []string{"run-speckeep.ps1", "archive --root $RootDir"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files, err := Files(LanguageSettings{
				Default:  "en",
				Docs:     "en",
				Agent:    "en",
				Comments: "en",
				Shell:    tc.shell,
			})
			if err != nil {
				t.Fatalf("Files() returned error: %v", err)
			}

			content := fileContentByTarget(t, files, tc.target)
			for _, snippet := range tc.want {
				if !strings.Contains(content, snippet) {
					t.Fatalf("expected %s to contain %q", tc.target, snippet)
				}
			}
		})
	}
}

func TestReadinessScriptsDelegateToInternalCLI(t *testing.T) {
	testCases := []struct {
		name   string
		shell  string
		target string
		want   []string
	}{
		{
			name:   "sh spec ready",
			shell:  "sh",
			target: "scripts/check-spec-ready.sh",
			want: []string{
				"run-speckeep.sh",
				"__internal check-spec-ready --root \"$ROOT_DIR\"",
			},
		},
		{
			name:   "sh plan ready",
			shell:  "sh",
			target: "scripts/check-plan-ready.sh",
			want: []string{
				"run-speckeep.sh",
				"__internal check-plan-ready --root \"$ROOT_DIR\"",
			},
		},
		{
			name:   "sh tasks ready",
			shell:  "sh",
			target: "scripts/check-tasks-ready.sh",
			want: []string{
				"run-speckeep.sh",
				"__internal check-tasks-ready --root \"$ROOT_DIR\"",
			},
		},
		{
			name:   "sh implement ready",
			shell:  "sh",
			target: "scripts/check-implement-ready.sh",
			want: []string{
				"run-speckeep.sh",
				"__internal check-implement-ready --root \"$ROOT_DIR\"",
			},
		},
		{
			name:   "sh verify ready",
			shell:  "sh",
			target: "scripts/check-verify-ready.sh",
			want: []string{
				"run-speckeep.sh",
				"__internal check-verify-ready --root \"$ROOT_DIR\"",
			},
		},
		{
			name:   "powershell verify task state",
			shell:  "powershell",
			target: "scripts/verify-task-state.ps1",
			want: []string{
				"run-speckeep.ps1",
				"__internal verify-task-state --root $RootDir",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files, err := Files(LanguageSettings{
				Default:  "en",
				Docs:     "en",
				Agent:    "en",
				Comments: "en",
				Shell:    tc.shell,
			})
			if err != nil {
				t.Fatalf("Files() returned error: %v", err)
			}

			content := fileContentByTarget(t, files, tc.target)
			for _, snippet := range tc.want {
				if !strings.Contains(content, snippet) {
					t.Fatalf("expected %s to contain %q", tc.target, snippet)
				}
			}
		})
	}
}

func TestUtilityScriptsDelegateToCLIBackends(t *testing.T) {
	testCases := []struct {
		name   string
		shell  string
		target string
		want   []string
	}{
		{
			name:   "sh list-open-tasks",
			shell:  "sh",
			target: "scripts/list-open-tasks.sh",
			want:   []string{"run-speckeep.sh", "__internal list-open-tasks --root \"$ROOT_DIR\""},
		},
		{
			name:   "sh list-specs",
			shell:  "sh",
			target: "scripts/list-specs.sh",
			want:   []string{"run-speckeep.sh", "__internal list-specs --root \"$ROOT_DIR\""},
		},
		{
			name:   "sh show-spec",
			shell:  "sh",
			target: "scripts/show-spec.sh",
			want:   []string{"run-speckeep.sh", "__internal show-spec --root \"$ROOT_DIR\""},
		},
		{
			name:   "sh link-agents",
			shell:  "sh",
			target: "scripts/link-agents.sh",
			want:   []string{"run-speckeep.sh", "__internal link-agents --root \"$ROOT_DIR\""},
		},
		{
			name:   "powershell list-specs",
			shell:  "powershell",
			target: "scripts/list-specs.ps1",
			want:   []string{"run-speckeep.ps1", "__internal list-specs --root $RootDir"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files, err := Files(LanguageSettings{
				Default:  "en",
				Docs:     "en",
				Agent:    "en",
				Comments: "en",
				Shell:    tc.shell,
			})
			if err != nil {
				t.Fatalf("Files() returned error: %v", err)
			}

			content := fileContentByTarget(t, files, tc.target)
			for _, snippet := range tc.want {
				if !strings.Contains(content, snippet) {
					t.Fatalf("expected %s to contain %q", tc.target, snippet)
				}
			}
		})
	}
}

func fileContentByTarget(t *testing.T, files []File, target string) string {
	t.Helper()

	for _, file := range files {
		if file.TargetPath == target {
			return file.Content
		}
	}

	t.Fatalf("expected generated file set to include %s", target)
	return ""
}

func languageAssetSet(t *testing.T, language string) []string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller information")
	}

	root := filepath.Join(filepath.Dir(filename), "assets", "lang", language)
	var files []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("walk language assets for %s: %v", language, err)
	}

	sort.Strings(files)
	return files
}
