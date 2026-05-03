package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"speckeep/src/internal/config"
	"speckeep/src/internal/project"
)

func repairSpecsDir(t *testing.T, root string) string {
	t.Helper()
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("config.Load returned error: %v", err)
	}
	dir, err := cfg.SpecsDir(root)
	if err != nil {
		t.Fatalf("cfg.SpecsDir returned error: %v", err)
	}
	return dir
}

func TestRepairFeatureMigratesLegacyFlatSpec(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specsDir := repairSpecsDir(t, root)
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(specsDir) returned error: %v", err)
	}
	legacyPath := filepath.Join(specsDir, "demo.md")
	content := "# Demo Spec\n"
	if err := os.WriteFile(legacyPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	result, err := RepairFeature(root, "demo", false)
	if err != nil {
		t.Fatalf("RepairFeature returned error: %v", err)
	}
	if !result.Changed {
		t.Fatalf("expected repair to change files, got %+v", result)
	}

	canonicalPath := filepath.Join(specsDir, "demo", "spec.md")
	if _, err := os.Stat(canonicalPath); err != nil {
		t.Fatalf("expected canonical spec to exist: %v", err)
	}
	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Fatalf("expected legacy spec to be moved, got err=%v", err)
	}
}

func TestRepairFeatureRemovesDuplicateLegacyFlatSpec(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specsDir := repairSpecsDir(t, root)
	content := "# Demo Spec\n"
	canonicalPath := filepath.Join(specsDir, "demo", "spec.md")
	if err := os.MkdirAll(filepath.Dir(canonicalPath), 0o755); err != nil {
		t.Fatalf("MkdirAll(canonical) returned error: %v", err)
	}
	if err := os.WriteFile(canonicalPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(canonical) returned error: %v", err)
	}
	legacyPath := filepath.Join(specsDir, "demo.md")
	if err := os.WriteFile(legacyPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(legacy) returned error: %v", err)
	}

	result, err := RepairFeature(root, "demo", false)
	if err != nil {
		t.Fatalf("RepairFeature returned error: %v", err)
	}
	if !result.Changed {
		t.Fatalf("expected duplicate cleanup to change files, got %+v", result)
	}
	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Fatalf("expected duplicate legacy spec to be removed, got err=%v", err)
	}
}

func TestRepairFeatureWarnsWhenCanonicalAndLegacyDiffer(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specsDir := repairSpecsDir(t, root)
	canonicalPath := filepath.Join(specsDir, "demo", "spec.md")
	if err := os.MkdirAll(filepath.Dir(canonicalPath), 0o755); err != nil {
		t.Fatalf("MkdirAll(canonical) returned error: %v", err)
	}
	if err := os.WriteFile(canonicalPath, []byte("canonical"), 0o644); err != nil {
		t.Fatalf("WriteFile(canonical) returned error: %v", err)
	}
	legacyPath := filepath.Join(specsDir, "demo.md")
	if err := os.WriteFile(legacyPath, []byte("legacy"), 0o644); err != nil {
		t.Fatalf("WriteFile(legacy) returned error: %v", err)
	}

	result, err := RepairFeature(root, "demo", false)
	if err != nil {
		t.Fatalf("RepairFeature returned error: %v", err)
	}
	if result.Changed {
		t.Fatalf("expected conflicting repair to avoid changes, got %+v", result)
	}
	if len(result.Warnings) == 0 || !strings.Contains(result.Warnings[0], "differ") {
		t.Fatalf("expected warning about differing reports, got %+v", result)
	}
}
