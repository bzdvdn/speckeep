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

func TestNormalizeTargetsDropsDeprecatedSilently(t *testing.T) {
	// "continue" (Continue.dev) is discontinued; a project whose
	// speckeep.yaml still lists it should not hard-fail every command —
	// it should just quietly stop being generated for.
	targets, err := NormalizeTargets([]string{"claude", "continue", "cursor"})
	if err != nil {
		t.Fatalf("NormalizeTargets returned error: %v", err)
	}
	for _, target := range targets {
		if target == "continue" {
			t.Fatalf("expected \"continue\" to be dropped, got %#v", targets)
		}
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets after dropping continue, got %#v", targets)
	}

	if _, ok := targetSkillDirs["continue"]; ok {
		t.Fatal("continue should no longer be a generated target")
	}
}

func TestNormalizeTargetsAll(t *testing.T) {
	targets, err := NormalizeTargets([]string{"all"})
	if err != nil {
		t.Fatalf("NormalizeTargets returned error: %v", err)
	}

	if len(targets) != 19 {
		t.Fatalf("expected 19 targets for all, got %#v", targets)
	}
}

func TestFlatCommandTargetsGetCommandsAlongsideSkills(t *testing.T) {
	// Windsurf, OpenCode, Cline, and Amazon Q Skills are model/tool-invoked
	// only, never invocable by name there — real commands for those
	// targets come from a flat file per phase in their own directory,
	// generated in addition to the skill pack.
	for target, flat := range targetFlatCommandDirs {
		files, err := FilesForTarget(target, "en", "sh")
		if err != nil {
			t.Fatalf("FilesForTarget(%q) returned error: %v", target, err)
		}
		want := flat.dir + "/spk-spec.md"
		wantInvoke := flat.invokePrefix + "spk-spec"
		var found bool
		for _, f := range files {
			if f.Path != want {
				continue
			}
			found = true
			if strings.Contains(f.Content, "name:") && strings.HasPrefix(f.Content, "---") {
				t.Fatalf("flat command %s should not carry skill YAML frontmatter, got: %s", f.Path, f.Content[:min(80, len(f.Content))])
			}
			if !strings.Contains(f.Content, wantInvoke) {
				t.Fatalf("flat command %s missing its own invocation name %q", f.Path, wantInvoke)
			}
		}
		if !found {
			t.Fatalf("target %q missing flat command %s", target, want)
		}
	}
}

func TestGeminiGetsTomlCommandsAlongsideSkills(t *testing.T) {
	// Gemini CLI Skills are auto-discovered but not slash-invocable; its
	// real custom-command mechanism is TOML, not markdown.
	files, err := FilesForTarget("gemini", "en", "sh")
	if err != nil {
		t.Fatalf("FilesForTarget(gemini) returned error: %v", err)
	}
	want := ".gemini/commands/spk-spec.toml"
	var found bool
	for _, f := range files {
		if f.Path != want {
			continue
		}
		found = true
		if !strings.Contains(f.Content, "description = ") || !strings.Contains(f.Content, "prompt = ") {
			t.Fatalf("gemini command %s missing required TOML fields, got: %s", f.Path, f.Content)
		}
	}
	if !found {
		t.Fatalf("gemini missing TOML command %s", want)
	}
}

func TestAmazonQSkillsMovedOffQDir(t *testing.T) {
	if got := targetSkillDirs["amazonq"]; got != ".amazonq/skills" {
		t.Fatalf("amazonq skill dir = %q, want .amazonq/skills (the .q/ root was never confirmed read by Amazon Q)", got)
	}

	legacy := LegacyAmazonQSkillPaths(DefaultCommands("sh"))
	if len(legacy) == 0 {
		t.Fatal("expected legacy .q/skills paths for cleanup")
	}
	for _, p := range legacy {
		if !strings.HasPrefix(p, ".q/skills/") {
			t.Fatalf("unexpected legacy amazonq path %q, want .q/skills/ prefix", p)
		}
	}
}

func TestEveryTargetGetsCompositeSDDPack(t *testing.T) {
	commands := DefaultCommands("sh")
	wantFiles := len(commands) + 1 // root SKILL.md + one phase file per command

	for _, target := range SupportedTargets() {
		files, err := FilesForTarget(target, "en", "sh")
		if err != nil {
			t.Fatalf("FilesForTarget(%q) returned error: %v", target, err)
		}
		extra := 0
		if target == "aider" {
			extra = 1 // CONVENTIONS.md pointer
		}
		if _, ok := targetFlatCommandDirs[target]; ok {
			extra += len(commands) // one flat command file per phase, since Skills aren't invocable-by-name there
		}
		if target == "gemini" {
			extra += len(commands) // one TOML command file per phase
		}
		if len(files) != wantFiles+extra {
			t.Fatalf("expected %d files for %q, got %d", wantFiles+extra, target, len(files))
		}

		base := targetSkillDirs[target]
		flat, hasFlatDir := targetFlatCommandDirs[target]
		rootSeen := false
		for _, f := range files {
			switch {
			case f.Path == base+"/sdd/SKILL.md":
				rootSeen = true
				if !strings.Contains(f.Content, "speckeep check") {
					t.Fatalf("root skill %s missing CLI gate guidance", f.Path)
				}
				if !strings.Contains(f.Content, "/spk-") {
					t.Fatalf("root skill %s missing direct phase-skill invocation guidance", f.Path)
				}
			case strings.HasSuffix(f.Path, "/SKILL.md") && strings.Contains(f.Path, base+"/spk-"):
				if f.Content == "" {
					t.Fatalf("phase skill %s has no content", f.Path)
				}
				if !strings.Contains(f.Content, ".speckeep/templates/prompts/") {
					t.Fatalf("phase skill %s must reference the canonical prompt", f.Path)
				}
			case hasFlatDir && strings.HasPrefix(f.Path, flat.dir+"/spk-"):
				if f.Content == "" {
					t.Fatalf("flat command %s has no content", f.Path)
				}
				if !strings.Contains(f.Content, ".speckeep/templates/prompts/") {
					t.Fatalf("flat command %s must reference the canonical prompt", f.Path)
				}
			case target == "gemini" && strings.HasPrefix(f.Path, ".gemini/commands/spk-") && strings.HasSuffix(f.Path, ".toml"):
				if f.Content == "" {
					t.Fatalf("gemini command %s has no content", f.Path)
				}
			default:
				// aider CONVENTIONS pointer is the only other file.
				if f.Path != ".aider/CONVENTIONS.md" {
					t.Fatalf("unexpected file %q for target %q", f.Path, target)
				}
			}
		}
		if !rootSeen {
			t.Fatalf("target %q missing root SKILL.md", target)
		}
	}
}

func TestFiles(t *testing.T) {
	files, err := Files(SupportedTargets(), "en", "sh")
	if err != nil {
		t.Fatalf("Files returned error: %v", err)
	}

	commands := DefaultCommands("sh")
	// +1 for aider CONVENTIONS pointer; +len(commands) per target with a flat
	// command dir (windsurf, opencode — Skills aren't slash-invocable there).
	want := len(SupportedTargets())*(len(commands)+1) + 1 + len(targetFlatCommandDirs)*len(commands) + len(commands) // +len(commands) for gemini's TOML commands
	if len(files) != want {
		t.Fatalf("expected %d generated files, got %d", want, len(files))
	}

	required := map[string]bool{
		".claude/skills/sdd/SKILL.md":           false,
		".claude/skills/spk-spec/SKILL.md":      false,
		".claude/skills/spk-implement/SKILL.md": false,
		".opencode/skills/spk-verify/SKILL.md":  false,
		".cursor/skills/sdd/SKILL.md":           false,
		".gemini/skills/spk-plan/SKILL.md":      false,
		".amazonq/skills/spk-propose/SKILL.md":  false,
		".github/skills/sdd/SKILL.md":           false,
		".aider/CONVENTIONS.md":                 false,
		".windsurf/workflows/spk-spec.md":       false,
		".opencode/commands/spk-implement.md":   false,
		".clinerules/workflows/spk-tasks.md":    false,
		".amazonq/prompts/spk-verify.md":        false,
		".gemini/commands/spk-plan.toml":        false,
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

func TestPhaseSkillNamesCoverEveryCommand(t *testing.T) {
	base := targetSkillDirs["claude"]
	files, err := FilesForTarget("claude", "en", "sh")
	if err != nil {
		t.Fatalf("FilesForTarget returned error: %v", err)
	}
	for _, cmd := range DefaultCommands("sh") {
		want := base + "/spk-" + cmd.Name + "/SKILL.md"
		var found bool
		for _, f := range files {
			if f.Path == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing phase skill %s", want)
		}
	}
}

func TestAiderPointerReferencesSkillPack(t *testing.T) {
	files, err := FilesForTarget("aider", "en", "sh")
	if err != nil {
		t.Fatalf("FilesForTarget(aider) returned error: %v", err)
	}
	for _, f := range files {
		if f.Path == ".aider/CONVENTIONS.md" {
			if !strings.Contains(f.Content, ".aider/skills/sdd/SKILL.md") {
				t.Fatalf("aider CONVENTIONS must point at the skill pack")
			}
			return
		}
	}
	t.Fatal("aider CONVENTIONS.md pointer not generated")
}

func TestEveryTargetHasASkillDir(t *testing.T) {
	if len(targetSkillDirs) != 19 {
		t.Fatalf("expected 19 targets with skill dirs, got %d", len(targetSkillDirs))
	}
	for target, dir := range targetSkillDirs {
		if !strings.HasPrefix(dir, ".") || !strings.HasSuffix(dir, "skills") {
			t.Fatalf("target %q has unexpected skill dir %q", target, dir)
		}
	}
}

func TestLegacySkillPhasePathsCoverEveryTargetAndCommand(t *testing.T) {
	commands := DefaultCommands("sh")
	paths := LegacySkillPhasePaths(commands)

	want := len(targetSkillDirs) * len(commands)
	if len(paths) != want {
		t.Fatalf("expected %d legacy skill phase paths, got %d", want, len(paths))
	}

	pathSet := make(map[string]bool, len(paths))
	for _, p := range paths {
		pathSet[p] = true
	}

	base := targetSkillDirs["claude"]
	for _, cmd := range commands {
		legacy := base + "/sdd/phases/" + cmd.Name + ".md"
		if !pathSet[legacy] {
			t.Fatalf("expected legacy path %s to be present", legacy)
		}
	}
}
