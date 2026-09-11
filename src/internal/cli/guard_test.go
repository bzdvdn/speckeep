package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGuardPassesWhenAllFeaturesCloseable(t *testing.T) {
	root := t.TempDir()
	writeConvergeFeature(t, root, "alpha", false)

	stdout, stderr, err := executeRoot(t, "guard", root)
	if err != nil {
		t.Fatalf("guard should pass with a closeable feature, got: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "pass") {
		t.Fatalf("expected pass verdict, got: %s", stdout)
	}
}

func TestGuardFailsOnInFlightFeature(t *testing.T) {
	root := t.TempDir()
	writeConvergeFeature(t, root, "alpha", false)
	// Not yet implemented: open tasks.
	specDir := ensureSpecDir(t, root, "beta")
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# beta\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec beta): %v", err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "tasks.md"), []byte("# Tasks\n\n## Phase 1\n- [ ] T1.1 todo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks beta): %v", err)
	}

	stdout, _, err := executeRoot(t, "guard", root)
	if err == nil {
		t.Fatalf("guard should fail when a feature is in flight")
	}
	if !strings.Contains(stdout, "not ready to close") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestGuardSlugFilter(t *testing.T) {
	root := t.TempDir()
	writeConvergeFeature(t, root, "alpha", false)
	specDir := ensureSpecDir(t, root, "beta")
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# beta\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec beta): %v", err)
	}

	// beta is not closeable (only a spec) — but filtering on alpha must pass.
	stdout, stderr, err := executeRoot(t, "guard", root, "--slug", "alpha")
	if err != nil {
		t.Fatalf("guard --slug alpha should pass, got: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "pass") {
		t.Fatalf("expected pass verdict, got: %s", stdout)
	}
}

// A feature whose tasks are all checked but carry no Proof: must FAIL guard.
func TestGuardFailsOnMissingProofDespiteCompleteTasks(t *testing.T) {
	root := t.TempDir()
	writeConvergeFeature(t, root, "alpha", false)
	specDir := ensureSpecDir(t, root, "alpha")
	tasksWithoutProof := "# Tasks\n\n## Phase 1\n- [x] T1.1 do it\n\n## Acceptance Coverage\n- AC-001 -> T1.1\n"
	if err := os.WriteFile(filepath.Join(specDir, "tasks.md"), []byte(tasksWithoutProof), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks): %v", err)
	}

	stdout, _, err := executeRoot(t, "guard", root)
	if err == nil {
		t.Fatalf("guard should fail on tasks without Proof")
	}
	if !strings.Contains(stdout, "not ready to close") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}
