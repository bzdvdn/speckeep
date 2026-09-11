package workflow

import (
	"context"
	"testing"

	"speckeep/src/internal/config"
	"speckeep/src/internal/project"
)

func TestCheckProposeReadyRequiresConstitutionAndTemplates(t *testing.T) {
	root := t.TempDir()
	if _, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	}); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	result, err := CheckProposeReady(context.Background(), config.Default(), root, "")
	if err != nil {
		t.Fatalf("CheckProposeReady returned error: %v", err)
	}
	if result.Failed {
		t.Fatalf("propose should be ready on a fresh initialized project, got %+v", result.Lines)
	}
	if len(result.Lines) < 5 {
		t.Fatalf("expected ok lines for each required file, got %+v", result.Lines)
	}
}

func TestCheckProposeReadyFlagsMissingConstitution(t *testing.T) {
	root := t.TempDir()
	if _, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	}); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}
	cfg, err := config.Load(context.Background(), root)
	if err != nil {
		t.Fatalf("config.Load returned error: %v", err)
	}
	cfg.Project.ConstitutionFile = "MISSING_CONSTITUTION.md"

	result, err := CheckProposeReady(context.Background(), cfg, root, "")
	if err != nil {
		t.Fatalf("CheckProposeReady returned error: %v", err)
	}
	if !result.Failed {
		t.Fatalf("expected propose to fail on missing constitution, got %+v", result.Lines)
	}
}
