package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultAppliesExpectedDefaults(t *testing.T) {
	cfg := Default()

	if cfg.Version != 1 {
		t.Fatalf("Version = %d, want 1", cfg.Version)
	}
	if cfg.Runtime.Shell != "sh" {
		t.Fatalf("Runtime.Shell = %q, want sh", cfg.Runtime.Shell)
	}
	if cfg.Language.Default != "en" || cfg.Language.Docs != "en" || cfg.Language.Agent != "en" || cfg.Language.Comments != "en" {
		t.Fatalf("unexpected default languages: %+v", cfg.Language)
	}
	if cfg.Paths.SpecsDir != "specs" {
		t.Fatalf("SpecsDir = %q, want %q", cfg.Paths.SpecsDir, "specs")
	}
	if cfg.Templates.SpecPrompt != "prompts/spec.md" {
		t.Fatalf("SpecPrompt = %q, want %q", cfg.Templates.SpecPrompt, "prompts/spec.md")
	}
	if cfg.Scripts.CheckInspectReady != "check-inspect-ready.sh" {
		t.Fatalf("CheckInspectReady = %q, want %q", cfg.Scripts.CheckInspectReady, "check-inspect-ready.sh")
	}
	if cfg.Scripts.RunSpeckeep != "run-speckeep.sh" {
		t.Fatalf("RunSpeckeep = %q, want %q", cfg.Scripts.RunSpeckeep, "run-speckeep.sh")
	}
	if cfg.Scripts.VerifyTaskState != "verify-task-state.sh" {
		t.Fatalf("VerifyTaskState = %q, want %q", cfg.Scripts.VerifyTaskState, "verify-task-state.sh")
	}
}

func TestLoadReturnsDefaultsWhenConfigDoesNotExist(t *testing.T) {
	root := t.TempDir()

	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Language.Docs != "en" {
		t.Fatalf("Docs language = %q, want en", cfg.Language.Docs)
	}
	if cfg.Project.ConstitutionFile != "CONSTITUTION.md" {
		t.Fatalf("ConstitutionFile = %q", cfg.Project.ConstitutionFile)
	}
}

func TestSaveAndLoadPreserveConfigAndApplyDefaults(t *testing.T) {
	root := t.TempDir()

	cfg := Config{}
	cfg.Project.Name = "demo"
	cfg.Language.Default = "ru"
	cfg.Language.Docs = "ru"
	cfg.Paths.SpecsDir = "workspace/specs"
	cfg.Agents.Targets = []string{"claude", "cursor"}

	if err := Save(root, cfg); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	loaded, err := Load(root)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if loaded.Project.Name != "demo" {
		t.Fatalf("Project.Name = %q, want demo", loaded.Project.Name)
	}
	if loaded.Runtime.Shell != "sh" {
		t.Fatalf("Runtime.Shell = %q, want sh", loaded.Runtime.Shell)
	}
	if loaded.Language.Default != "ru" || loaded.Language.Docs != "ru" {
		t.Fatalf("unexpected loaded languages: %+v", loaded.Language)
	}
	if loaded.Language.Agent != "ru" || loaded.Language.Comments != "ru" {
		t.Fatalf("expected missing language values to inherit default, got %+v", loaded.Language)
	}
	if got, want := loaded.Paths.SpecsDir, "workspace/specs"; got != want {
		t.Fatalf("SpecsDir = %q, want %q", got, want)
	}
	if len(loaded.Agents.Targets) != 2 || loaded.Agents.Targets[0] != "claude" || loaded.Agents.Targets[1] != "cursor" {
		t.Fatalf("unexpected agent targets: %+v", loaded.Agents.Targets)
	}
	if loaded.Templates.Spec == "" || loaded.Scripts.ShowSpec == "" {
		t.Fatalf("expected template and script defaults to be applied: templates=%+v scripts=%+v", loaded.Templates, loaded.Scripts)
	}
}

func TestScriptDefaultsForShell(t *testing.T) {
	ps := ScriptDefaultsForShell("powershell")
	if ps.CheckSpecReady != "check-spec-ready.ps1" {
		t.Fatalf("CheckSpecReady = %q, want check-spec-ready.ps1", ps.CheckSpecReady)
	}
	if ps.RunSpeckeep != "run-speckeep.ps1" {
		t.Fatalf("RunSpeckeep = %q, want run-speckeep.ps1", ps.RunSpeckeep)
	}
	if ps.VerifyTaskState != "verify-task-state.ps1" {
		t.Fatalf("VerifyTaskState = %q, want verify-task-state.ps1", ps.VerifyTaskState)
	}
}

func TestResolvePathHelpersRespectConfiguredPaths(t *testing.T) {
	root := t.TempDir()
	cfg := Default()
	cfg.Paths.SpecsDir = "workspace/specs"
	cfg.Paths.ArchiveDir = "workspace/archive"
	cfg.Paths.TemplatesDir = "workspace/templates"
	cfg.Paths.ScriptsDir = "workspace/scripts"

	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		t.Fatalf("SpecsDir returned error: %v", err)
	}
	scriptsDir, err := cfg.ScriptsDir(root)
	if err != nil {
		t.Fatalf("ScriptsDir returned error: %v", err)
	}

	if specsDir != filepath.Join(root, "workspace", "specs") {
		t.Fatalf("SpecsDir resolved to %q", specsDir)
	}
	if scriptsDir != filepath.Join(root, "workspace", "scripts") {
		t.Fatalf("ScriptsDir resolved to %q", scriptsDir)
	}
}
