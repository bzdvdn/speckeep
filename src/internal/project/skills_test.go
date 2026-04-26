package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddListRemoveSkills(t *testing.T) {
	root := t.TempDir()

	_, err := Initialize(root, InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	localSkillDir := filepath.Join(root, "skills", "architecture")
	if err := os.MkdirAll(localSkillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	added, err := AddSkill(root, AddSkillOptions{
		ID:        "architecture",
		FromLocal: "skills/architecture",
		Path:      ".",
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("AddSkill(local) returned error: %v", err)
	}
	if len(added.Messages) == 0 {
		t.Fatalf("expected non-empty add messages")
	}

	added, err = AddSkill(root, AddSkillOptions{
		ID:      "openai-docs",
		FromGit: createGitSkillRepo(t, root),
		Ref:     "v1.2.3",
		Path:    "skills/openai-docs",
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("AddSkill(git) returned error: %v", err)
	}
	if got, want := added.Entry.Ref, "v1.2.3"; got != want {
		t.Fatalf("Entry.Ref = %q, want %q", got, want)
	}
	if strings.TrimSpace(added.Entry.CheckoutDir) == "" {
		t.Fatalf("expected checkout dir to be populated")
	}

	list, err := ListSkills(root)
	if err != nil {
		t.Fatalf("ListSkills returned error: %v", err)
	}
	if len(list.Skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(list.Skills))
	}

	removed, err := RemoveSkill(root, RemoveSkillOptions{ID: "architecture"})
	if err != nil {
		t.Fatalf("RemoveSkill returned error: %v", err)
	}
	if !removed.Removed {
		t.Fatalf("expected removed=true")
	}

	list, err = ListSkills(root)
	if err != nil {
		t.Fatalf("ListSkills returned error: %v", err)
	}
	if len(list.Skills) != 1 || list.Skills[0].ID != "openai-docs" {
		t.Fatalf("unexpected skills after remove: %+v", list.Skills)
	}

	manifestPath := filepath.Join(root, ".speckeep", "skills", "manifest.yaml")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("ReadFile(manifest) returned error: %v", err)
	}
	if !strings.Contains(string(content), "openai-docs") {
		t.Fatalf("expected manifest to contain remaining skill, got %q", string(content))
	}

	agentsPath := filepath.Join(root, "AGENTS.md")
	agentsContent, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("ReadFile(AGENTS.md) returned error: %v", err)
	}
	if !strings.Contains(string(agentsContent), "openai-docs") {
		t.Fatalf("expected AGENTS.md to contain skill listing, got %q", string(agentsContent))
	}
}

func TestAddSkillRejectsFloatingGitRef(t *testing.T) {
	root := t.TempDir()

	_, err := Initialize(root, InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	_, err = AddSkill(root, AddSkillOptions{
		ID:      "test-skill",
		FromGit: createGitSkillRepo(t, root),
		Ref:     "main",
		Enabled: true,
	})
	if err == nil {
		t.Fatalf("expected AddSkill to fail for floating ref")
	}
}

func createGitSkillRepo(t *testing.T, root string) string {
	t.Helper()

	repoDir := filepath.Join(root, "git-skill-repo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "SKILL.md"), []byte("# skill\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(SKILL.md) returned error: %v", err)
	}
	runGitTest(t, repoDir, "init")
	runGitTest(t, repoDir, "config", "user.email", "test@example.com")
	runGitTest(t, repoDir, "config", "user.name", "Test User")
	runGitTest(t, repoDir, "add", "SKILL.md")
	runGitTest(t, repoDir, "commit", "-m", "init skill")
	runGitTest(t, repoDir, "tag", "v1.2.3")
	return repoDir
}

func runGitTest(t *testing.T, dir string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v, output=%s", args, err, string(output))
	}
}

func TestRefreshUpdatesAgentsBlockWithSkills(t *testing.T) {
	root := t.TempDir()

	_, err := Initialize(root, InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	localSkillDir := filepath.Join(root, "skills", "architecture")
	if err := os.MkdirAll(localSkillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	if _, err := AddSkill(root, AddSkillOptions{
		ID:        "architecture",
		FromLocal: "skills/architecture",
		Enabled:   true,
	}); err != nil {
		t.Fatalf("AddSkill returned error: %v", err)
	}

	agentsPath := filepath.Join(root, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("manual header\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(AGENTS.md) returned error: %v", err)
	}

	if _, err := Refresh(root, RefreshOptions{}); err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	content, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("ReadFile(AGENTS.md) returned error: %v", err)
	}
	if !strings.Contains(string(content), "architecture") {
		t.Fatalf("expected AGENTS.md to include skill after refresh, got %q", string(content))
	}
}

func TestInstallSkillsCopiesToAgentSkillDirsAndCleansStale(t *testing.T) {
	root := t.TempDir()

	_, err := Initialize(root, InitOptions{
		InitGit:      false,
		DefaultLang:  "en",
		Shell:        "sh",
		AgentTargets: []string{"codex", "claude", "windsurf", "kilocode", "trae", "cursor"},
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	localSkillDir := filepath.Join(root, "skills", "architecture")
	if err := os.MkdirAll(filepath.Join(localSkillDir, "assets"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localSkillDir, "SKILL.md"), []byte("# Architecture\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(SKILL.md) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localSkillDir, "assets", "checklist.md"), []byte("- done\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(checklist) returned error: %v", err)
	}

	if _, err := AddSkill(root, AddSkillOptions{
		ID:        "architecture",
		FromLocal: "skills/architecture",
		Enabled:   true,
	}); err != nil {
		t.Fatalf("AddSkill returned error: %v", err)
	}

	installed, err := InstallSkills(root, InstallSkillsOptions{})
	if err != nil {
		t.Fatalf("InstallSkills returned error: %v", err)
	}
	if len(installed.Warnings) == 0 {
		t.Fatalf("expected warning for unsupported target")
	}

	for _, rel := range []string{
		".codex/skills/architecture/SKILL.md",
		".claude/skills/architecture/SKILL.md",
		".windsurf/skills/architecture/SKILL.md",
		".kilocode/skills/architecture/SKILL.md",
		".trae/skills/architecture/SKILL.md",
	} {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected installed skill file %s: %v", rel, err)
		}
	}

	if _, err := RemoveSkill(root, RemoveSkillOptions{ID: "architecture"}); err != nil {
		t.Fatalf("RemoveSkill returned error: %v", err)
	}
	if _, err := InstallSkills(root, InstallSkillsOptions{}); err != nil {
		t.Fatalf("InstallSkills after remove returned error: %v", err)
	}

	for _, rel := range []string{
		".codex/skills/architecture",
		".claude/skills/architecture",
		".windsurf/skills/architecture",
		".kilocode/skills/architecture",
		".trae/skills/architecture",
	} {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected stale installed skill to be removed for %s, got err=%v", rel, err)
		}
	}
}
