package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConvergeFeature(t *testing.T, root, slug string, openTask bool) {
	t.Helper()
	initTestProject(t, root)
	specDir := ensureSpecDir(t, root, slug)
	specContent := "# " + slug + "\n\n## Goal\nx\n\n## Requirements\n- RQ-001 demo\n\n## Acceptance Criteria\n### AC-001 Demo\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec): %v", err)
	}
	tasksContent := "# Tasks\n\n## Phase 1\n- [x] T1.1 implement demo\n  Proof: code src/demo.go Demo\n"
	if openTask {
		tasksContent += "- [ ] T1.2 open follow-up\n  Touches: src/demo.go\n"
	}
	tasksContent += "\n## Acceptance Coverage\n- AC-001 -> T1.1\n"
	if err := os.WriteFile(filepath.Join(specDir, "tasks.md"), []byte(tasksContent), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks): %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatalf("MkdirAll(src): %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "demo.go"), []byte("package main\nfunc Demo() {}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(demo.go): %v", err)
	}
}

func TestConvergePassesWhenTasksAreProven(t *testing.T) {
	root := t.TempDir()
	writeConvergeFeature(t, root, "demo", false)

	stdout, stderr, err := executeRoot(t, "converge", "demo", root)
	if err != nil {
		t.Fatalf("converge should pass, got: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "converged") {
		t.Fatalf("expected converged verdict, got stdout: %s", stdout)
	}
}

func TestConvergeFailsOnOpenTasks(t *testing.T) {
	root := t.TempDir()
	writeConvergeFeature(t, root, "demo", true)

	stdout, _, err := executeRoot(t, "converge", "demo", root)
	if err == nil {
		t.Fatalf("converge should fail on open tasks")
	}
	if !strings.Contains(stdout, "open tasks remain") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestConvergeFailsOnMissingProofFile(t *testing.T) {
	root := t.TempDir()
	writeConvergeFeature(t, root, "demo", false)
	// Remove the file referenced by the Proof line.
	if err := os.Remove(filepath.Join(root, "src", "demo.go")); err != nil {
		t.Fatalf("Remove(demo.go): %v", err)
	}
	stdout, _, err := executeRoot(t, "converge", "demo", root)
	if err == nil {
		t.Fatalf("converge should fail on a missing proof file")
	}
	if !strings.Contains(stdout, "references missing file") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestConvergeJSONOutputShape(t *testing.T) {
	root := t.TempDir()
	writeConvergeFeature(t, root, "demo", false)

	stdout, _, err := executeRoot(t, "converge", "demo", root, "--json")
	if err != nil {
		t.Fatalf("converge --json should pass, got %v", err)
	}
	if !strings.Contains(stdout, `"converged": true`) {
		t.Fatalf("expected converged:true in JSON, got: %s", stdout)
	}
}
