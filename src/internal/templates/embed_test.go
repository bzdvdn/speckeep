package templates

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
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
				"templates/prompts/verify.md",
			}
			ext := ".sh"
			if tc.settings.Shell == "powershell" {
				ext = ".ps1"
			}
			requiredFiles = append(requiredFiles,
				"scripts/run-speckeep"+ext,
				"scripts/check-ready"+ext,
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

func TestPromptContentAppliesShellPlaceholders(t *testing.T) {
	for _, language := range []string{"en", "ru"} {
		t.Run(language, func(t *testing.T) {
			shContent, err := PromptContent(language, "propose", "sh")
			if err != nil {
				t.Fatalf("PromptContent(sh) returned error: %v", err)
			}
			if !strings.Contains(shContent, "check-ready.sh propose") {
				t.Fatalf("sh prompt must reference check-ready.sh, got:\n%s", shContent)
			}

			psContent, err := PromptContent(language, "propose", "powershell")
			if err != nil {
				t.Fatalf("PromptContent(powershell) returned error: %v", err)
			}
			if !strings.Contains(psContent, "check-ready.ps1 propose") {
				t.Fatalf("powershell prompt must reference check-ready.ps1, got:\n%s", psContent)
			}
			if strings.Contains(psContent, "check-ready.sh") {
				t.Fatalf("powershell prompt must not reference check-ready.sh, got:\n%s", psContent)
			}

			psVerify, err := PromptContent(language, "verify", "powershell")
			if err != nil {
				t.Fatalf("PromptContent(verify, powershell) returned error: %v", err)
			}
			for _, want := range []string{"verify-task-state.ps1", "trace.ps1"} {
				if !strings.Contains(psVerify, want) {
					t.Fatalf("powershell verify prompt must reference %s, got:\n%s", want, psVerify)
				}
			}
			for _, stale := range []string{"verify-task-state.sh", "trace.sh"} {
				if strings.Contains(psVerify, stale) {
					t.Fatalf("powershell verify prompt must not reference %s, got:\n%s", stale, psVerify)
				}
			}
		})
	}
}

func TestShellScriptReplacementCoversGeneratedScripts(t *testing.T) {
	replacements := scriptReplacements("powershell")
	for _, name := range shellScriptNames {
		got, ok := replacements[name+".sh"]
		if !ok {
			t.Fatalf("missing shell replacement for generated script %s.sh", name)
		}
		if want := name + ".ps1"; got != want {
			t.Fatalf("replacement for %s.sh = %q, want %q", name, got, want)
		}
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
	if !strings.Contains(content, `spec_path="$(CDPATH= cd -- "$(dirname -- "$input")" && pwd)/$(basename -- "$input")"`) {
		t.Fatalf("expected inspect-spec.sh to absolutize CWD-relative spec paths\ncontent:\n%s", content)
	}

	psFiles, err := Files(LanguageSettings{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
		Shell:    "powershell",
	})
	if err != nil {
		t.Fatalf("Files(powershell) returned error: %v", err)
	}
	psContent := fileContentByTarget(t, psFiles, "scripts/inspect-spec.ps1")
	if !strings.Contains(psContent, "$specPath = (Resolve-Path -LiteralPath $specPath).Path") {
		t.Fatalf("expected inspect-spec.ps1 to absolutize CWD-relative spec paths\ncontent:\n%s", psContent)
	}
}

func TestLanguageTemplatesUseShellPlaceholder(t *testing.T) {
	for _, language := range []string{"en", "ru"} {
		for _, target := range []string{
			"templates/agents-snippet.md",
			"templates/commands.md",
			"templates/prompts/propose.md",
			"templates/prompts/verify.md",
			"templates/prompts/converge.md",
			"templates/prompts/recap.md",
			"templates/prompts/spec.md",
			"templates/prompts/inspect.md",
		} {
			content, err := localizedFileContent(language, target)
			if err != nil {
				t.Fatalf("%s/%s: %v", language, target, err)
			}
			for _, ext := range []string{".sh", ".ps1"} {
				if strings.Contains(content, ext) {
					t.Fatalf("%s/%s hardcodes %s; use [SCRIPT_EXT] instead", language, target, ext)
				}
			}
		}
	}
}

func TestPromptScriptRefsResolveToGeneratedWrappers(t *testing.T) {
	refRe := regexp.MustCompile(`scripts/([a-z0-9-]+)\[SCRIPT_EXT\]`)
	targets := []string{
		"templates/agents-snippet.md",
		"templates/commands.md",
		"templates/prompts/propose.md",
		"templates/prompts/verify.md",
		"templates/prompts/converge.md",
		"templates/prompts/recap.md",
		"templates/prompts/spec.md",
		"templates/prompts/inspect.md",
	}

	for _, language := range []string{"en", "ru"} {
		for _, target := range targets {
			content, err := localizedFileContent(language, target)
			if err != nil {
				t.Fatalf("%s/%s: %v", language, target, err)
			}
			refs := refRe.FindAllStringSubmatch(content, -1)
			for _, shell := range []string{"sh", "powershell"} {
				ext := ".sh"
				if shell == "powershell" {
					ext = ".ps1"
				}
				files, err := Files(LanguageSettings{
					Default:  language,
					Docs:     language,
					Agent:    language,
					Comments: language,
					Shell:    shell,
				})
				if err != nil {
					t.Fatalf("%s/%s: Files(%s) returned error: %v", language, target, shell, err)
				}
				existing := make(map[string]struct{}, len(files))
				for _, file := range files {
					existing[file.TargetPath] = struct{}{}
				}
				for _, ref := range refs {
					want := "scripts/" + ref[1] + ext
					if _, ok := existing[want]; !ok {
						t.Fatalf("%s/%s references %s, but %s is not generated for shell %s", language, target, ref[0], want, shell)
					}
				}
			}
		}
	}
}

func TestGeneratedFilesResolveScriptExtPlaceholder(t *testing.T) {
	for _, shell := range []string{"sh", "powershell"} {
		files, err := Files(LanguageSettings{
			Default:  "en",
			Docs:     "en",
			Agent:    "en",
			Comments: "en",
			Shell:    shell,
		})
		if err != nil {
			t.Fatalf("%s: Files() returned error: %v", shell, err)
		}
		for _, file := range files {
			if strings.Contains(file.Content, "[SCRIPT_EXT]") {
				t.Fatalf("%s: %s still contains unresolved [SCRIPT_EXT]", shell, file.TargetPath)
			}
		}
	}
}

func TestPowerShellGeneratedFilesHaveNoShellScriptRefs(t *testing.T) {
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

	for _, file := range files {
		for _, name := range shellScriptNames {
			stale := name + ".sh"
			if strings.Contains(file.Content, stale) {
				t.Fatalf("%s still references %s in a powershell project", file.TargetPath, stale)
			}
		}
	}
}

func TestWrapperScriptsUseMatchingRunnerAndCommand(t *testing.T) {
	expected := map[string]string{
		"run-speckeep":       "",
		"inspect-spec":       "__internal inspect-spec",
		"check-constitution": "__internal check-constitution",
		"check-ready":        "__internal check-ready",
		"archive-feature":    " archive ",
		"trace":              " trace ",
		"verify-task-state":  "__internal verify-task-state",
		"list-open-tasks":    "__internal list-open-tasks",
		"link-agents":        "__internal link-agents",
		"list-specs":         "__internal list-specs",
		"show-spec":          "__internal show-spec",
	}

	for _, shell := range []string{"sh", "powershell"} {
		ext := ".sh"
		if shell == "powershell" {
			ext = ".ps1"
		}
		files, err := Files(LanguageSettings{
			Default:  "en",
			Docs:     "en",
			Agent:    "en",
			Comments: "en",
			Shell:    shell,
		})
		if err != nil {
			t.Fatalf("%s: Files() returned error: %v", shell, err)
		}
		runner := "run-speckeep" + ext
		for name, command := range expected {
			if name == "run-speckeep" {
				continue
			}
			content := fileContentByTarget(t, files, "scripts/"+name+ext)
			if !strings.Contains(content, runner) {
				t.Fatalf("%s: scripts/%s%s must delegate through %s", shell, name, ext, runner)
			}
			if !strings.Contains(content, command) {
				t.Fatalf("%s: scripts/%s%s must invoke %q", shell, name, ext, command)
			}
		}
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
		"check-ready.sh inspect",
		".speckeep/constitution.summary.md",
		"Return to: /spk-spec <slug>",
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
		"Proof plan:",
		"Proof: <kind> <path> [<anchor>]",
		"Ready for: speckeep converge <slug>",
		"If open tasks remain",
		"Do not emit `/spk-implement <slug>` when no tasks remain",
		"End with the canonical end block from `AGENTS.md`",
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
		".speckeep/constitution.summary.md",
		"/spk-tasks <slug>",
		"/spk-inspect <slug>",
		"/spk-plan <slug>",
		"Final line (mandatory): `Ready for:",
		"End with the canonical end block from `AGENTS.md`",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected spec prompt to contain %q", snippet)
		}
	}
}

func TestProposePromptDefinesOneShotLaneAndExpressDefault(t *testing.T) {
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

	content := fileContentByTarget(t, files, "templates/prompts/propose.md")
	requiredSnippets := []string{
		"one-shot",
		"spec.md` + `tasks.md",
		"check-ready.sh propose",
		"express",
		"Ready for: /spk-implement <slug>",
		".speckeep/constitution.summary.md",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected propose prompt to contain %q", snippet)
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
		"Create `research.md` only when needed",
		"Do not create `research.md` for generic brainstorming",
		".speckeep/constitution.summary.md",
		"Ready for: /spk-tasks <slug>",
		"End with the canonical end block from `AGENTS.md`",
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
				"Implementation Context",
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
				"Read code only to confirm",
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
		"no public `speckeep verify-task-state` command",
		"Return to: /spk-<phase> <slug>",
		"Ready for: speckeep archive <slug> .",
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
				".speckeep/constitution.summary.md",
				"/spk-tasks <slug>",
				"/spk-inspect <slug>",
				"/spk-plan <slug>",
				"Final line (mandatory): `Ready for:",
				"End with the canonical end block from `AGENTS.md`",
			},
		},
		{
			target: "templates/prompts/inspect.md",
			want: []string{
				".speckeep/constitution.summary.md",
				"Ready for: /spk-tasks <slug>",
				"/spk-plan <slug>",
				"End with the canonical end block from `AGENTS.md`",
			},
		},
		{
			target: "templates/prompts/plan.md",
			want: []string{
				".speckeep/constitution.summary.md",
				"Ready for: /spk-tasks <slug>",
				"End with the canonical end block from `AGENTS.md`",
			},
		},
		{
			target: "templates/prompts/tasks.md",
			want: []string{
				".speckeep/constitution.summary.md",
				"Ready for: /spk-implement <slug>",
				"End with the canonical end block from `AGENTS.md`",
			},
		},
		{
			target: "templates/prompts/implement.md",
			want: []string{
				"Default scope: only the **first unfinished phase**",
				".speckeep/constitution.summary.md",
				"End with the canonical end block from `AGENTS.md`",
				"Proof: <kind> <path> [<anchor>]",
				"Ready for: speckeep converge <slug>",
			},
		},
		{
			target: "templates/prompts/verify.md",
			want: []string{
				".speckeep/constitution.summary.md",
				"End with the canonical end block from `AGENTS.md`",
				"Return to: /spk-<phase> <slug>",
				"Ready for: speckeep archive <slug> .",
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
				"If `blocked`, final line: `Return to: /spk-<phase> <slug>`",
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
			name:   "sh check ready",
			shell:  "sh",
			target: "scripts/check-ready.sh",
			want: []string{
				"run-speckeep.sh",
				"__internal check-ready \"$PHASE\" --root \"$ROOT_DIR\"",
			},
		},
		{
			name:   "powershell check ready",
			shell:  "powershell",
			target: "scripts/check-ready.ps1",
			want: []string{
				"run-speckeep.ps1",
				"__internal check-ready $Phase --root $RootDir",
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

func TestPromptReferencesToTemplatesExist(t *testing.T) {
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

	existing := make(map[string]bool, len(files))
	for _, f := range files {
		existing[f.TargetPath] = true
	}

	// Prompts point at the artifact skeleton via `.speckeep/templates/<name>.md`;
	// that reference must resolve to a real embedded template or the phase has no
	// format authority and silently degrades. Keep this contract in sync on rename.
	refRe := regexp.MustCompile(`\.speckeep/templates/([A-Za-z0-9_.\-/]+\.md)`)
	for _, f := range files {
		if !strings.HasPrefix(f.TargetPath, "templates/prompts/") {
			continue
		}
		for _, m := range refRe.FindAllStringSubmatch(f.Content, -1) {
			target := "templates/" + m[1]
			if !existing[target] {
				t.Errorf("%s references missing template %s", f.TargetPath, target)
			}
		}
	}
}

func TestRepoAgentsMdTracksSnippetCanon(t *testing.T) {
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
	snippet := fileContentByTarget(t, files, "templates/agents-snippet.md")

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller information")
	}
	repoRoot := filepath.Join(filepath.Dir(filename), "..", "..", "..")
	agentsBytes, err := os.ReadFile(filepath.Join(repoRoot, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read repo AGENTS.md: %v", err)
	}
	agentsMd := string(agentsBytes)

	// AGENTS.md locates the cross-phase canon inside the snippet; that pointer
	// must resolve to a snippet that really carries the claimed canonical topics.
	if !strings.Contains(agentsMd, ".speckeep/templates/agents-snippet.md") {
		t.Error("repo AGENTS.md must point to .speckeep/templates/agents-snippet.md as the single source of truth")
	}
	for _, topic := range []string{"End block", "Verify gate policy"} {
		if !strings.Contains(snippet, topic) {
			t.Errorf("agents-snippet.md no longer carries the canonical %q that AGENTS.md points to", topic)
		}
	}

	// Hard invariants both top-level rule sources must agree on; drift here means
	// agent guidance and human-facing rules silently diverge.
	for _, invariant := range []string{"workflow.verify", "feature/<slug>"} {
		if !strings.Contains(agentsMd, invariant) {
			t.Errorf("repo AGENTS.md lost shared invariant %q", invariant)
		}
		if !strings.Contains(snippet, invariant) {
			t.Errorf("agents-snippet.md lost shared invariant %q", invariant)
		}
	}

	// Phase chain core tokens must appear in the same order in both sources
	// (optional inspect/verify may differ between them).
	coreChain := []string{"constitution", "spec", "plan", "tasks", "implement", "archive"}
	for _, src := range []struct {
		name string
		text string
	}{
		{name: "repo AGENTS.md", text: agentsMd},
		{name: "agents-snippet.md", text: snippet},
	} {
		pos := 0
		for _, token := range coreChain {
			idx := strings.Index(src.text[pos:], token)
			if idx < 0 {
				t.Errorf("%s phase chain lost token %q", src.name, token)
				break
			}
			pos += idx + len(token)
		}
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

// TestContextFootprintBudgets guards the always-on and per-phase prompt
// footprint. AGENTS.md (agents-snippet) is loaded every session, and the core
// phase prompts are loaded once per phase; silent growth here is the main
// regression risk for token usage.
func TestContextFootprintBudgets(t *testing.T) {
	core := []string{"spec", "inspect", "plan", "tasks", "implement", "verify"}
	budgets := map[string]struct {
		snippet     int
		corePrompts int
	}{
		"en": {snippet: 5500, corePrompts: 30500},
		"ru": {snippet: 7200, corePrompts: 39700},
	}

	for lang, budget := range budgets {
		snippet, err := localizedFileContent(lang, "templates/agents-snippet.md")
		if err != nil {
			t.Fatalf("%s agents-snippet: %v", lang, err)
		}
		if len(snippet) > budget.snippet {
			t.Errorf("%s agents-snippet is %d bytes, over budget %d — trim always-on context", lang, len(snippet), budget.snippet)
		}

		total := 0
		for _, name := range core {
			content, err := localizedFileContent(lang, "templates/prompts/"+name+".md")
			if err != nil {
				t.Fatalf("%s prompt %s: %v", lang, name, err)
			}
			total += len(content)
		}
		if total > budget.corePrompts {
			t.Errorf("%s core prompts total %d bytes, over budget %d — dedupe against AGENTS.md", lang, total, budget.corePrompts)
		}
	}
}

// TestAgentsSnippetRoutesNewFeaturesToPropose guards the default entry point:
// a new-feature request must route directly to /spk-propose, without first
// loading the sdd router.
func TestAgentsSnippetRoutesNewFeaturesToPropose(t *testing.T) {
	wantByLang := map[string]string{
		"en": "skip the `sdd` router",
		"ru": "не грузи `sdd`-роутер",
	}
	// The default entry is propose only for small/low-risk work; design-heavy
	// ideas must be routed to /spk-spec, and propose must not claim to write plans.
	designByLang := map[string][]string{
		"en": {"several features", "never writes `plan.md`"},
		"ru": {"несколько фич", "не пишет"},
	}
	for lang, want := range wantByLang {
		content, err := localizedFileContent(lang, "templates/agents-snippet.md")
		if err != nil {
			t.Fatalf("%s agents-snippet: %v", lang, err)
		}
		if !strings.Contains(content, "/spk-propose") {
			t.Errorf("%s agents-snippet must route new features to /spk-propose", lang)
		}
		if !strings.Contains(content, want) {
			t.Errorf("%s agents-snippet must tell the agent to %q for a new feature", lang, want)
		}
		for _, needle := range designByLang[lang] {
			if !strings.Contains(content, needle) {
				t.Errorf("%s agents-snippet must route design-heavy ideas to /spk-spec (%q missing)", lang, needle)
			}
		}
	}
}

// TestProposeExpressOnlyContract guards the single propose decision rule:
// propose is the express lane, hands design-heavy work to /spk-spec, and never
// writes plan.md (that lane belongs to /spk-spec -> /spk-plan).
func TestProposeExpressOnlyContract(t *testing.T) {
	cases := []struct {
		lang      string
		max       int
		required  []string
		forbidden []string
	}{
		{
			lang: "en", max: 6200,
			required: []string{
				"Escalation (rare)", "default is to stay here", "breaking change or a schema migration",
				"coordinated across services or repos", "If unsure",
				"Not signals", "two endpoints", "data-model.md", "contracts/api.md",
				"Return to: /spk-spec <slug>", "Ready for: /spk-plan <slug>",
				"Never create `plan.md`", "Stop if",
			},
			forbidden: []string{"planned"},
		},
		{
			lang: "ru", max: 9600,
			required: []string{
				"Эскалация (редко)", "По умолчанию оставайся здесь", "ломающее изменение или миграция схемы",
				"согласовывать между сервисами", "Если сомневаешься",
				"Не сигналы", "две ручки", "data-model.md", "contracts/api.md",
				"Вернуться к: /spk-spec <slug>", "Готово к: /spk-plan <slug>",
				"Никогда не создавай `plan.md`", "Стоп, если",
			},
			forbidden: []string{"planned"},
		},
	}
	for _, tc := range cases {
		content, err := localizedFileContent(tc.lang, "templates/prompts/propose.md")
		if err != nil {
			t.Fatalf("%s propose prompt: %v", tc.lang, err)
		}
		if len(content) > tc.max {
			t.Errorf("%s propose prompt is %d bytes, over budget %d", tc.lang, len(content), tc.max)
		}
		for _, want := range tc.required {
			if !strings.Contains(content, want) {
				t.Errorf("%s propose prompt missing %q", tc.lang, want)
			}
		}
		for _, banned := range tc.forbidden {
			if strings.Contains(content, banned) {
				t.Errorf("%s propose prompt still contains %q (planned lane must be gone)", tc.lang, banned)
			}
		}
	}
}

// TestArchiveReadyPromptsOfferFinishAlt keeps the local-close alternative in
// every prompt that suggests archive: archive stays primary, finish is the
// explicit alt for local merge.
func TestArchiveReadyPromptsOfferFinishAlt(t *testing.T) {
	prompts := []string{"converge", "verify", "hotfix", "handoff", "module-context", "implement"}
	for _, lang := range []string{"en", "ru"} {
		for _, name := range prompts {
			content, err := localizedFileContent(lang, "templates/prompts/"+name+".md")
			if err != nil {
				t.Fatalf("%s prompt %s: %v", lang, name, err)
			}
			for _, want := range []string{"speckeep archive <slug>", "speckeep finish <slug>"} {
				if !strings.Contains(content, want) {
					t.Errorf("%s %s.md missing %q", lang, name, want)
				}
			}
		}
	}
}

// TestCloseGuidanceMentionsFinishTargetBranch keeps the agent guidance explicit
// about merging into a user-named branch (e.g. --into dev).
func TestCloseGuidanceMentionsFinishTargetBranch(t *testing.T) {
	for _, lang := range []string{"en", "ru"} {
		for _, target := range []string{"templates/agents-snippet.md", "templates/commands.md"} {
			content, err := localizedFileContent(lang, target)
			if err != nil {
				t.Fatalf("%s/%s: %v", lang, target, err)
			}
			if !strings.Contains(content, "--into") {
				t.Errorf("%s/%s must mention --into for a user-named target branch", lang, target)
			}
		}
	}
}

// TestPromptContracts guards the meaning of the core phase prompts: required
// sections, safety anchors, final routing commands, and per-prompt budgets.
// It catches both accidental rule removal during compression and silent growth.
func TestPromptContracts(t *testing.T) {
	type contract struct {
		name          string
		maxEN, maxRU  int
		required      []string
		finalCommands []string
	}

	common := []string{"## Phase Contract", "Stop if", "AGENTS.md"}
	contracts := []contract{
		{
			name: "spec", maxEN: 5200, maxRU: 7300,
			required:      append(append([]string{}, common...), "## Self-Check", "## Alignment Pass", "AC-", "/spk-inspect <slug>", "/spk-plan <slug>"),
			finalCommands: []string{"/spk-tasks <slug>"},
		},
		{
			name: "inspect", maxEN: 4600, maxRU: 6150,
			required:      append(append([]string{}, common...), "## Self-Check", "/spk-plan <slug>"),
			finalCommands: []string{"/spk-tasks <slug>"},
		},
		{
			name: "plan", maxEN: 4400, maxRU: 6000,
			required:      append(append([]string{}, common...), "## Self-Check", "DEC-", "data-model"),
			finalCommands: []string{"/spk-tasks <slug>"},
		},
		{
			name: "tasks", maxEN: 4100, maxRU: 5500,
			required:      append(append([]string{}, common...), "## Self-Check", "Touches:", "Surface Map", "Implementation Context"),
			finalCommands: []string{"/spk-implement <slug>"},
		},
		{
			name: "implement", maxEN: 7000, maxRU: 7250,
			required:      append(append([]string{}, common...), "Proof:", "Touches:", "scope violation", "## Output expectations"),
			finalCommands: []string{"speckeep converge <slug>", "/spk-verify <slug>"},
		},
		{
			name: "verify", maxEN: 6700, maxRU: 8100,
			required:      append(append([]string{}, common...), "## Self-Check", "Proof:", "Verification Matrix", "## Review", "blocker", "Not Verified", "concerns", "blocked", "speckeep archive <slug>", "speckeep finish <slug>"),
			finalCommands: []string{"speckeep archive <slug>", "speckeep finish <slug>"},
		},
	}

	for _, lang := range []string{"en", "ru"} {
		finalMarker := "Ready for"
		if lang == "ru" {
			finalMarker = "Готово к"
		}
		for _, tc := range contracts {
			content, err := localizedFileContent(lang, "templates/prompts/"+tc.name+".md")
			if err != nil {
				t.Fatalf("%s prompt %s: %v", lang, tc.name, err)
			}
			limit := tc.maxEN
			if lang == "ru" {
				limit = tc.maxRU
			}
			if len(content) > limit {
				t.Errorf("%s %s prompt is %d bytes, over budget %d", lang, tc.name, len(content), limit)
			}
			if !strings.Contains(content, finalMarker) {
				t.Errorf("%s %s prompt lost the %q final marker", lang, tc.name, finalMarker)
			}
			for _, want := range append(append([]string{}, tc.required...), tc.finalCommands...) {
				if !strings.Contains(content, want) {
					t.Errorf("%s %s prompt missing %q", lang, tc.name, want)
				}
			}
		}
	}
}

// TestGrillPromptContract guards the optional pre-spec interview: design-tree
// rounds, a hard cap, no artifacts by default, and a clear next lane.
func TestGrillPromptContract(t *testing.T) {
	type contract struct {
		lang     string
		max      int
		required []string
	}
	contracts := []contract{
		{
			lang: "en", max: 3000,
			required: []string{
				"frontier", "Recommended", "3 rounds / 12 questions",
				"no spec, tasks, or code", "Do not create `grill.md`",
				"Stop if", "Ready for: /spk-propose <slug>",
			},
		},
		{
			lang: "ru", max: 4300,
			required: []string{
				"frontier", "Рекомендация", "3 раунда / 12 вопросов",
				"без spec/tasks/code", "Не создавай `grill.md`",
				"Stop if", "Готово к: /spk-propose <slug>",
			},
		},
	}
	for _, tc := range contracts {
		content, err := localizedFileContent(tc.lang, "templates/prompts/grill.md")
		if err != nil {
			t.Fatalf("%s grill prompt: %v", tc.lang, err)
		}
		if len(content) > tc.max {
			t.Errorf("%s grill prompt is %d bytes, over budget %d", tc.lang, len(content), tc.max)
		}
		for _, want := range tc.required {
			if !strings.Contains(content, want) {
				t.Errorf("%s grill prompt missing %q", tc.lang, want)
			}
		}
	}
}

// TestResumePromptContract guards the thin in-session resume: minimal packet,
// no artifacts, and an explicit boundary against /spk-handoff.
func TestResumePromptContract(t *testing.T) {
	cases := []struct {
		lang     string
		max      int
		required []string
	}{
		{
			lang: "en", max: 1800,
			required: []string{
				"not `/spk-handoff`", "at most 8 lines", "No files created",
				"next open task", "Blockers:", "Stop if", "Ready for: /spk-",
			},
		},
		{
			lang: "ru", max: 2800,
			required: []string{
				"не `/spk-handoff`", "не более 8 строк", "Файлы не создаются",
				"следующую открытую задачу", "Blockers:", "Stop if", "Готово к: /spk-",
			},
		},
	}
	for _, tc := range cases {
		content, err := localizedFileContent(tc.lang, "templates/prompts/resume.md")
		if err != nil {
			t.Fatalf("%s resume prompt: %v", tc.lang, err)
		}
		if len(content) > tc.max {
			t.Errorf("%s resume prompt is %d bytes, over budget %d", tc.lang, len(content), tc.max)
		}
		for _, want := range tc.required {
			if !strings.Contains(content, want) {
				t.Errorf("%s resume prompt missing %q", tc.lang, want)
			}
		}
	}
}

// TestResearchAndDiagnosePromptContracts guards the two optional discipline
// loops: cited research and a bounded diagnosis cycle.
func TestResearchAndDiagnosePromptContracts(t *testing.T) {
	cases := []struct {
		lang, name string
		max        int
		required   []string
	}{
		{
			lang: "en", name: "research", max: 2200,
			required: []string{
				"research.md", "primary source", "Source:", "1–5 specific unknowns",
				"Do not implement code", "Stop if", "Ready for: /spk-plan <slug>",
			},
		},
		{
			lang: "ru", name: "research", max: 3300,
			required: []string{
				"research.md", "primary source", "Источник:", "1–5 конкретных unknowns",
				"Не реализуй код", "Stop if", "Готово к: /spk-plan <slug>",
			},
		},
		{
			lang: "en", name: "diagnose", max: 2900,
			required: []string{
				"Reproduce", "Minimize", "Hypothesize", "Instrument", "Regression",
				"3 hypotheses / 2 fix cycles", "Stop if", "/spk-hotfix <slug>",
			},
		},
		{
			lang: "ru", name: "diagnose", max: 4300,
			required: []string{
				"Reproduce", "Minimize", "Hypothesize", "Instrument", "Regression",
				"3 гипотезы / 2 цикла фиксов", "Stop if", "/spk-hotfix <slug>",
			},
		},
	}
	for _, tc := range cases {
		content, err := localizedFileContent(tc.lang, "templates/prompts/"+tc.name+".md")
		if err != nil {
			t.Fatalf("%s %s prompt: %v", tc.lang, tc.name, err)
		}
		if len(content) > tc.max {
			t.Errorf("%s %s prompt is %d bytes, over budget %d", tc.lang, tc.name, len(content), tc.max)
		}
		for _, want := range tc.required {
			if !strings.Contains(content, want) {
				t.Errorf("%s %s prompt missing %q", tc.lang, tc.name, want)
			}
		}
	}
}
