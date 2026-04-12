package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"speckeep/src/internal/config"
)

func TestInitializeCreatesWorkspaceAndAgentTargets(t *testing.T) {
	root := t.TempDir()

	_, err := Initialize(root, InitOptions{
		InitGit:      false,
		DefaultLang:  "en",
		Shell:        "sh",
		AgentTargets: []string{"claude", "cursor"},
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("config.Load returned error: %v", err)
	}

	if got, want := cfg.Language.Docs, "en"; got != want {
		t.Fatalf("docs language = %q, want %q", got, want)
	}
	if got, want := strings.Join(cfg.Agents.Targets, ","), "claude,cursor"; got != want {
		t.Fatalf("agent targets = %q, want %q", got, want)
	}
	if got, want := cfg.Runtime.Shell, "sh"; got != want {
		t.Fatalf("shell = %q, want %q", got, want)
	}

	required := []string{
		filepath.Join(root, ".speckeep", "speckeep.yaml"),
		filepath.Join(root, ".speckeep", "constitution.md"),
		filepath.Join(root, ".speckeep", "scripts", "run-speckeep.sh"),
		filepath.Join(root, "AGENTS.md"),
		filepath.Join(root, ".claude", "commands", "speckeep.inspect.md"),
		filepath.Join(root, ".cursor", "rules", "speckeep-inspect.mdc"),
	}
	for _, path := range required {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}

	agentsContent, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("ReadFile(AGENTS.md) returned error: %v", err)
	}
	if !strings.Contains(string(agentsContent), "<!-- speckeep:start -->") {
		t.Fatalf("expected AGENTS.md to contain managed speckeep block, got %q", string(agentsContent))
	}
}

func TestAddRemoveAndCleanupAgents(t *testing.T) {
	root := t.TempDir()

	_, err := Initialize(root, InitOptions{
		InitGit:      false,
		DefaultLang:  "en",
		Shell:        "sh",
		AgentTargets: []string{"claude"},
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	_, err = AddAgents(root, AddAgentsOptions{Targets: []string{"cursor"}})
	if err != nil {
		t.Fatalf("AddAgents returned error: %v", err)
	}

	cursorPath := filepath.Join(root, ".cursor", "rules", "speckeep-inspect.mdc")
	if _, err := os.Stat(cursorPath); err != nil {
		t.Fatalf("expected cursor agent file after AddAgents: %v", err)
	}

	_, err = RemoveAgents(root, RemoveAgentsOptions{Targets: []string{"cursor"}})
	if err != nil {
		t.Fatalf("RemoveAgents returned error: %v", err)
	}

	if _, err := os.Stat(cursorPath); !os.IsNotExist(err) {
		t.Fatalf("expected cursor agent file to be removed, got err=%v", err)
	}

	if err := os.MkdirAll(filepath.Dir(cursorPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(cursorPath, []byte("orphan"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	result, err := CleanupAgents(root)
	if err != nil {
		t.Fatalf("CleanupAgents returned error: %v", err)
	}
	if len(result.Messages) == 0 || !strings.Contains(strings.Join(result.Messages, "\n"), "removed orphaned agent artifact") {
		t.Fatalf("expected cleanup message, got %v", result.Messages)
	}
	if _, err := os.Stat(cursorPath); !os.IsNotExist(err) {
		t.Fatalf("expected orphaned cursor file to be removed, got err=%v", err)
	}

	list, err := ListAgents(root)
	if err != nil {
		t.Fatalf("ListAgents returned error: %v", err)
	}
	if got, want := strings.Join(list.Targets, ","), "claude"; got != want {
		t.Fatalf("enabled targets = %q, want %q", got, want)
	}
}

func TestCleanupAgentsNoop(t *testing.T) {
	root := t.TempDir()

	_, err := Initialize(root, InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	result, err := CleanupAgents(root)
	if err != nil {
		t.Fatalf("CleanupAgents returned error: %v", err)
	}
	if len(result.Messages) != 1 || result.Messages[0] != "no orphaned agent artifacts found" {
		t.Fatalf("unexpected cleanup messages: %v", result.Messages)
	}
}

func TestInitializeWithPowerShellGeneratesPS1Scripts(t *testing.T) {
	root := t.TempDir()

	_, err := Initialize(root, InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "powershell",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	required := []string{
		filepath.Join(root, ".speckeep", "scripts", "run-speckeep.ps1"),
		filepath.Join(root, ".speckeep", "scripts", "check-spec-ready.ps1"),
		filepath.Join(root, ".speckeep", "scripts", "verify-task-state.ps1"),
	}
	for _, path := range required {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}
}

func TestInitializeGeneratesReadinessScriptsWithTraceabilityChecks(t *testing.T) {
	root := t.TempDir()

	_, err := Initialize(root, InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	testCases := []struct {
		path string
		want []string
	}{
		{
			path: filepath.Join(root, ".speckeep", "scripts", "check-spec-ready.sh"),
			want: []string{
				"run-speckeep.sh",
				"__internal check-spec-ready --root \"$ROOT_DIR\"",
			},
		},
		{
			path: filepath.Join(root, ".speckeep", "scripts", "check-implement-ready.sh"),
			want: []string{
				"run-speckeep.sh",
				"__internal check-implement-ready --root \"$ROOT_DIR\"",
			},
		},
		{
			path: filepath.Join(root, ".speckeep", "scripts", "check-verify-ready.sh"),
			want: []string{
				"run-speckeep.sh",
				"__internal check-verify-ready --root \"$ROOT_DIR\"",
			},
		},
	}

	for _, tc := range testCases {
		content, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatalf("ReadFile(%s) returned error: %v", tc.path, err)
		}
		for _, want := range tc.want {
			if !strings.Contains(string(content), want) {
				t.Fatalf("expected %s to contain %q", tc.path, want)
			}
		}
	}
}

func TestInitializeGeneratesDraftspecLauncher(t *testing.T) {
	root := t.TempDir()

	_, err := Initialize(root, InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	launcherPath := filepath.Join(root, ".speckeep", "scripts", "run-speckeep.sh")
	content, err := os.ReadFile(launcherPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) returned error: %v", launcherPath, err)
	}

	for _, want := range []string{"SPECKEEP_BIN", "speckeep CLI not found", "add speckeep to PATH"} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("expected %s to contain %q", launcherPath, want)
		}
	}
}

func TestRefreshUpdatesManagedFilesWithoutTouchingAuthoredArtifacts(t *testing.T) {
	root := t.TempDir()

	_, err := Initialize(root, InitOptions{
		InitGit:      false,
		DefaultLang:  "en",
		Shell:        "sh",
		AgentTargets: []string{"claude"},
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	constitutionPath := filepath.Join(root, ".speckeep", "constitution.md")
	customConstitution := "# custom constitution\n"
	if err := os.WriteFile(constitutionPath, []byte(customConstitution), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	promptPath := filepath.Join(root, ".speckeep", "templates", "prompts", "inspect.md")
	if err := os.WriteFile(promptPath, []byte("stale prompt"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	agentPath := filepath.Join(root, ".claude", "commands", "speckeep.inspect.md")
	if err := os.WriteFile(agentPath, []byte("stale agent file"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	agentsPath := filepath.Join(root, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("Project notes\n\n## SpecKeep\nold guidance\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	result, err := Refresh(root, RefreshOptions{
		Shell:        "powershell",
		AgentTargets: []string{"claude"},
	})
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if len(result.Updated) == 0 {
		t.Fatal("expected refresh to update managed files")
	}

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("config.Load returned error: %v", err)
	}
	if got, want := cfg.Runtime.Shell, "powershell"; got != want {
		t.Fatalf("shell = %q, want %q", got, want)
	}

	constitutionContent, err := os.ReadFile(constitutionPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(constitutionContent) != customConstitution {
		t.Fatalf("constitution content was unexpectedly changed: %q", string(constitutionContent))
	}

	promptContent, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(promptContent) == "stale prompt" {
		t.Fatal("expected refresh to overwrite managed prompt file")
	}

	if _, err := os.Stat(filepath.Join(root, ".speckeep", "scripts", "check-spec-ready.ps1")); err != nil {
		t.Fatalf("expected refreshed powershell script to exist: %v", err)
	}

	agentContent, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(agentContent) == "stale agent file" {
		t.Fatal("expected refresh to overwrite managed agent file")
	}

	agentsContent, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.Contains(string(agentsContent), "<!-- speckeep:start -->") {
		t.Fatalf("expected AGENTS.md to contain managed speckeep block, got %q", string(agentsContent))
	}
}

func TestRefreshDryRunDoesNotWriteChanges(t *testing.T) {
	root := t.TempDir()

	_, err := Initialize(root, InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	promptPath := filepath.Join(root, ".speckeep", "templates", "prompts", "inspect.md")
	if err := os.WriteFile(promptPath, []byte("stale prompt"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	result, err := Refresh(root, RefreshOptions{DryRun: true, Shell: "powershell"})
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if !result.DryRun {
		t.Fatal("expected dry-run refresh result")
	}
	if len(result.Updated) == 0 && len(result.Created) == 0 {
		t.Fatal("expected dry-run refresh to report pending changes")
	}

	content, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(content) != "stale prompt" {
		t.Fatalf("expected dry-run not to change managed file, got %q", string(content))
	}

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("config.Load returned error: %v", err)
	}
	if got, want := cfg.Runtime.Shell, "sh"; got != want {
		t.Fatalf("shell after dry-run = %q, want %q", got, want)
	}
}

func TestRefreshDryRunAfterInitializeDoesNotReportAgentsDrift(t *testing.T) {
	root := t.TempDir()

	_, err := Initialize(root, InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
		AgentTargets: []string{
			"codex",
		},
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	result, err := Refresh(root, RefreshOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	for _, path := range result.Updated {
		if path == "AGENTS.md" {
			t.Fatalf("did not expect AGENTS.md drift immediately after initialize, got updated=%v", result.Updated)
		}
	}
}
