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
				"templates/inspect.md",
				"templates/verify.md",
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
				"scripts/inspect-spec"+ext,
				"scripts/trace"+ext,
			)
			for _, required := range requiredFiles {
				if _, ok := targets[required]; !ok {
					t.Fatalf("expected generated file set to include %s", required)
				}
			}
		})
	}
}

func TestInspectSpecScriptResolvesSlugViaConfig(t *testing.T) {
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
	content := fileContentByTarget(t, files, "scripts/inspect-spec.sh")
	if !strings.Contains(content, "specs_dir") {
		t.Fatalf("expected inspect-spec.sh to reference specs_dir for non-default layouts")
	}
	if !strings.Contains(content, "<spec-file|slug>") {
		t.Fatalf("expected inspect-spec.sh usage to accept slug\ncontent:\n%s", content)
	}
}

func TestTraceScriptPinsRootDir(t *testing.T) {
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
	content := fileContentByTarget(t, files, "scripts/trace.sh")
	if !strings.Contains(content, "ROOT_DIR") {
		t.Fatalf("expected trace.sh to pin ROOT_DIR so it can run from any cwd")
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
		"## Phase Contract",
		"pass|concerns|blocked",
		"check-inspect-ready",
		"summary.md",
		"Ready for: /speckeep.plan <slug>",
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
			target: "templates/inspect.md",
			want: []string{
				"report_type: inspect",
				"slug: <slug>",
				"status: pass",
				"generated_at: <YYYY-MM-DD>",
			},
		},
		{
			target: "templates/verify.md",
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
		"Default scope: only the **first unfinished phase**",
		"`--continue`",
		"`--phase <N>`",
		"`--tasks <list>`",
		"Do not use `--phase` and `--tasks` together.",
		"`Touches:`",
		"Do not assume `research.md` should exist;",
		"Ready for: /speckeep.verify <slug>",
		"`Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`",
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
		"treat the next non-command user message as the continuation",
		"staged mode is canceled",
		"Do not pin technologies/versions unless required",
		"Ready for: /speckeep.inspect <slug>",
		"`Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`",
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
		"Create `plan/research.md` only when needed",
		"Do not create `research.md` for generic brainstorming",
		"Ready for: /speckeep.tasks <slug>",
		"`Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`",
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
				"preserve spec intent",
				"DEC-*",
				"Minimum context",
			},
		},
		{
			target: "templates/prompts/tasks.md",
			want: []string{
				"Touches:",
				"Surface Map",
				"Acceptance Coverage",
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
		"Do not pin technologies/versions unless required",
	} {
		if !strings.Contains(specContent, snippet) {
			t.Fatalf("expected spec prompt to contain %q", snippet)
		}
	}

	planContent := fileContentByTarget(t, files, "templates/prompts/plan.md")
	for _, snippet := range []string{
		"trade-offs",
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
				"Stop if",
			},
		},
		{
			target: "templates/prompts/plan.md",
			want: []string{
				"Stop if",
			},
		},
		{
			target: "templates/prompts/tasks.md",
			want: []string{
				"cannot be mapped to executable work without guessing",
			},
		},
		{
			target: "templates/prompts/verify.md",
			want: []string{
				"Deep code reads only",
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

	templateContent := fileContentByTarget(t, files, "templates/verify.md")
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
		"evidence log",
		"verify-task-state",
		"Return to: /speckeep.<phase> <slug>",
		"Ready for: /speckeep.archive <slug>",
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
				"Default scope: only the **first unfinished phase**",
				"`Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`",
				"Ready for: /speckeep.verify <slug>",
			},
		},
		{
			target: "templates/prompts/verify.md",
			want: []string{
				"`Slug`, `Status`, `Artifacts`, `Blockers`, and either `Ready for` or `Return to`",
				"Return to: /speckeep.<phase> <slug>",
				"Ready for: /speckeep.archive <slug>",
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
				"Spec captures intent, not plan/tasks.",
			},
		},
		{
			target: "templates/prompts/plan.md",
			want: []string{
				"Plan must preserve spec intent",
			},
		},
		{
			target: "templates/prompts/tasks.md",
			want: []string{
				"Do not implement or edit source code in the tasks phase.",
			},
		},
		{
			target: "templates/prompts/implement.md",
			want: []string{
				"No redesign / re-planning.",
			},
		},
		{
			target: "templates/prompts/inspect.md",
			want: []string{
				"pass|concerns|blocked",
			},
		},
		{
			target: "templates/prompts/verify.md",
			want: []string{
				"If `blocked`, end with `Return to: /speckeep.<phase> <slug>`.",
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
		{name: "sh-archive", shell: "sh", target: "scripts/archive-feature.sh", want: []string{"run-speckeep.sh", "archive \"$slug\" \"$ROOT_DIR\""}},
		{name: "powershell-archive", shell: "powershell", target: "scripts/archive-feature.ps1", want: []string{"run-speckeep.ps1", "archive $slug $RootDir"}},
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
