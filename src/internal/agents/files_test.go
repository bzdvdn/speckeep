package agents

import (
	"strings"
	"testing"
)

func TestNormalizeTargets(t *testing.T) {
	targets, err := NormalizeTargets([]string{"claude", "cursor,kilocode", "claude", "trae"})
	if err != nil {
		t.Fatalf("NormalizeTargets returned error: %v", err)
	}

	if len(targets) != 4 || targets[0] != "claude" || targets[1] != "cursor" || targets[2] != "kilocode" || targets[3] != "trae" {
		t.Fatalf("unexpected normalized targets: %#v", targets)
	}
}

func TestNormalizeTargetsAll(t *testing.T) {
	targets, err := NormalizeTargets([]string{"all"})
	if err != nil {
		t.Fatalf("NormalizeTargets returned error: %v", err)
	}

	if len(targets) != 9 {
		t.Fatalf("expected 9 targets for all, got %#v", targets)
	}
}

func TestFiles(t *testing.T) {
	files, err := Files([]string{"aider", "claude", "codex", "copilot", "cursor", "kilocode", "roocode", "trae", "windsurf"}, "en", "sh")
	if err != nil {
		t.Fatalf("Files returned error: %v", err)
	}

	if len(files) != 86 {
		t.Fatalf("expected 86 generated agent files, got %d", len(files))
	}

	required := map[string]bool{
		".aider/CONVENTIONS.md":                     false,
		".claude/commands/speckeep.inspect.md":      false,
		".claude/commands/speckeep.verify.md":       false,
		".codex/prompts/speckeep.plan.md":           false,
		".github/prompts/speckeep-spec.prompt.md":   false,
		".github/prompts/speckeep-verify.prompt.md": false,
		".cursor/rules/speckeep-implement.mdc":      false,
		".cursor/rules/speckeep-verify.mdc":         false,
		".kilocode/workflows/speckeep.verify.md":    false,
		".roo/rules/speckeep-spec.md":               false,
		".roo/rules/speckeep-plan.md":               false,
		".trae/project_rules.md":                    false,
		".windsurf/workflows/speckeep.implement.md": false,
		".windsurf/workflows/speckeep.verify.md":    false,
		".claude/commands/speckeep.recap.md":        false,
		".claude/commands/speckeep.hotfix.md":       false,
		".cursor/rules/speckeep-recap.mdc":          false,
	}

	for _, file := range files {
		if _, ok := required[file.Path]; ok {
			required[file.Path] = true
		}
		if file.Content == "" {
			t.Fatalf("expected non-empty content for %s", file.Path)
		}
	}

	for path, found := range required {
		if !found {
			t.Fatalf("missing generated agent file %s", path)
		}
	}
}

func TestRenderEmphasizesRunningScriptsFirst(t *testing.T) {
	// trae and aider are standalone agents that do not load AGENTS.md,
	// so script execution rules must appear in their generated files.
	tests := []struct {
		name string
		lang string
		want string
	}{
		{name: "en", lang: "en", want: "run it as a shell command"},
		{name: "ru", lang: "ru", want: "выполните его как shell-команду"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := renderTrae(tt.lang, "sh")
			if !strings.Contains(content, tt.want) {
				t.Fatalf("expected trae rules for %s to contain %q\ncontent:\n%s", tt.lang, tt.want, content)
			}
		})
	}
}

func TestRenderWindsurfMentionsHiddenDirsAndRepoRoot(t *testing.T) {
	spec := commandSpecs("sh")[3] // plan

	_, content, err := render("windsurf", "en", spec)
	if err != nil {
		t.Fatalf("render returned error: %v", err)
	}

	if !strings.Contains(content, "hidden/dotfiles") {
		t.Fatalf("expected windsurf output to mention hidden/dotfiles\ncontent:\n%s", content)
	}
	if !strings.Contains(content, "git rev-parse --show-toplevel") {
		t.Fatalf("expected windsurf output to mention git rev-parse --show-toplevel\ncontent:\n%s", content)
	}
	if !strings.Contains(content, "require `<slug>` as the first argument") {
		t.Fatalf("expected windsurf output to mention passing slug to readiness scripts\ncontent:\n%s", content)
	}
}

func TestRenderIncludesNoCommitRule(t *testing.T) {
	// trae and aider are standalone agents that do not load AGENTS.md,
	// so the no-commit rule must appear in their generated files.
	content := renderTrae("en", "sh")
	if !strings.Contains(content, "git commit") {
		t.Fatalf("expected trae rules to contain no-commit rule\ncontent:\n%s", content)
	}
}

func TestRenderTraeEmphasizesRunningScriptsFirst(t *testing.T) {
	tests := []struct {
		name string
		lang string
		want string
	}{
		{
			name: "en",
			lang: "en",
			want: "run it as a shell command",
		},
		{
			name: "ru",
			lang: "ru",
			want: "выполните его как shell-команду",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := renderTrae(tt.lang, "sh")
			if !strings.Contains(content, tt.want) {
				t.Fatalf("expected trae rules for %s to contain %q\ncontent:\n%s", tt.lang, tt.want, content)
			}
		})
	}
}

func TestRenderIncludesCommandHints(t *testing.T) {
	specs := map[string]commandSpec{}
	for _, spec := range commandSpecs("sh") {
		specs[spec.Name] = spec
	}

	tests := []struct {
		name   string
		target string
		lang   string
		spec   string
		want   string
	}{
		{name: "claude spec en", target: "claude", lang: "en", spec: "spec", want: "Command: `/speckeep.spec [request]`"},
		{name: "codex tasks en", target: "codex", lang: "en", spec: "tasks", want: "Command: `/speckeep.tasks [request]`"},
		{name: "copilot implement ru", target: "copilot", lang: "ru", spec: "implement", want: "Команда: `/speckeep.implement [request]`"},
		{name: "cursor verify en", target: "cursor", lang: "en", spec: "verify", want: "Command: `/speckeep.verify [request]`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, content, _ := render(tt.target, tt.lang, specs[tt.spec])
			if !strings.Contains(content, tt.want) {
				t.Fatalf("expected rendered content for %s/%s/%s to contain %q\ncontent:\n%s", tt.target, tt.lang, tt.spec, tt.want, content)
			}
		})
	}
}

func TestRenderTraeIncludesCommandHints(t *testing.T) {
	tests := []struct {
		name string
		lang string
		want string
	}{
		{name: "en", lang: "en", want: "- `/speckeep.verify` → .speckeep/templates/prompts/verify.md"},
		{name: "ru", lang: "ru", want: "- `/speckeep.verify` → .speckeep/templates/prompts/verify.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := renderTrae(tt.lang, "sh")
			if !strings.Contains(content, tt.want) {
				t.Fatalf("expected trae rules for %s to contain %q\ncontent:\n%s", tt.lang, tt.want, content)
			}
		})
	}
}

func TestRenderCodexDisallowsRawToolPayloads(t *testing.T) {
	specs := map[string]commandSpec{}
	for _, spec := range commandSpecs("sh") {
		specs[spec.Name] = spec
	}

	tests := []struct {
		name string
		lang string
		want string
	}{
		{
			name: "en",
			lang: "en",
			want: "Use tools directly through the agent runtime; do not print raw JSON/XML/tool-call payloads or expose internal reasoning about tool choice.",
		},
		{
			name: "ru",
			lang: "ru",
			want: "Используйте инструменты напрямую через runtime агента; не печатайте raw JSON/XML/tool-call payloads и не выводите внутренние рассуждения о выборе инструмента.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, content, _ := render("codex", tt.lang, specs["plan"])
			if !strings.Contains(content, tt.want) {
				t.Fatalf("expected codex rendered content for %s to contain %q\ncontent:\n%s", tt.lang, tt.want, content)
			}
		})
	}
}
