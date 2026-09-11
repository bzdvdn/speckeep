package workflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"speckeep/src/internal/config"
	"speckeep/src/internal/project"
)

func TestCheckConvergeReadyPassesOnProvenFeature(t *testing.T) {
	root := t.TempDir()
	setupConvergeProject(t, root, false)

	result, err := CheckConvergeReady(context.Background(), config.Default(), root, "demo")
	if err != nil {
		t.Fatalf("CheckConvergeReady returned error: %v", err)
	}
	if result.Failed {
		t.Fatalf("converge should pass on a proven feature, got %+v", result.Lines)
	}
	if !containsRaw(result.Lines, "CONVERGED=1") {
		t.Fatalf("expected CONVERGED=1, got %+v", result.Lines)
	}
}

func TestCheckConvergeReadyFailsOnOpenTasks(t *testing.T) {
	root := t.TempDir()
	setupConvergeProject(t, root, true)

	result, err := CheckConvergeReady(context.Background(), config.Default(), root, "demo")
	if err != nil {
		t.Fatalf("CheckConvergeReady returned error: %v", err)
	}
	if !result.Failed {
		t.Fatalf("converge should fail with open tasks, got %+v", result.Lines)
	}
	if !containsRaw(result.Lines, "CONVERGED=0") {
		t.Fatalf("expected CONVERGED=0, got %+v", result.Lines)
	}
}

func TestCheckTasksReadyAndImplementReadyAllowExpressMode(t *testing.T) {
	root := t.TempDir()
	setupConvergeProject(t, root, false)
	// Remove plan.md and data-model.md to force express mode (spec + tasks only).
	specDir := filepath.Join(checksSpecsDir(t, root), "demo")
	if err := os.Remove(filepath.Join(specDir, "plan", "plan.md")); err != nil {
		t.Fatalf("Remove(plan.md) returned error: %v", err)
	}
	if err := os.Remove(filepath.Join(specDir, "plan", "data-model.md")); err != nil {
		t.Fatalf("Remove(data-model.md) returned error: %v", err)
	}

	tasksResult, err := CheckTasksReady(context.Background(), config.Default(), root, "demo")
	if err != nil {
		t.Fatalf("CheckTasksReady returned error: %v", err)
	}
	if tasksResult.Failed {
		t.Fatalf("CheckTasksReady should not fail without plan/data-model (express mode), got %+v", tasksResult.Lines)
	}

	implResult, err := CheckImplementReady(context.Background(), config.Default(), root, "demo")
	if err != nil {
		t.Fatalf("CheckImplementReady returned error: %v", err)
	}
	if implResult.Failed {
		t.Fatalf("CheckImplementReady should not fail without plan/data-model (express mode), got %+v", implResult.Lines)
	}
}

// setupConvergeProject initializes a project with a spec + tasks (+ plan) for slug demo.
func setupConvergeProject(t *testing.T, root string, openTask bool) {
	t.Helper()
	if _, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	}); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specDir := filepath.Join(checksSpecsDir(t, root), "demo")
	if err := os.MkdirAll(filepath.Join(specDir, "plan"), 0o755); err != nil {
		t.Fatalf("MkdirAll(specDir) returned error: %v", err)
	}
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 First\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}
	planContent := "# Demo Plan\n\n## Implementation Surfaces\n- src/demo.go\n\n## Data and Contracts\n- Data model: no change\n\n## Acceptance Approach\n- AC-001 -> src/demo.go\n\n## Constitution Compliance\n- no conflicts\n"
	if err := os.WriteFile(filepath.Join(specDir, "plan", "plan.md"), []byte(planContent), 0o644); err != nil {
		t.Fatalf("WriteFile(plan) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "plan", "data-model.md"), []byte("# Data Model\n- no change\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(data-model) returned error: %v", err)
	}
	tasksContent := "# Tasks\n\n## Implementation Context\n- MVP: demo hook\n\n## Surface Map\n| Surface | Tasks |\n|---------|-------|\n| src/demo.go | T1.1 |\n\n## Phase 1: Foundation\n- [x] T1.1 implement demo. Touches: src/demo.go\n  Proof: code src/demo.go Demo\n"
	if openTask {
		tasksContent += "- [ ] T1.2 open follow-up. Touches: src/demo.go\n"
	}
	tasksContent += "\n## Acceptance Coverage\n- AC-001 -> T1.1\n"
	if err := os.WriteFile(filepath.Join(specDir, "plan", "tasks.md"), []byte(tasksContent), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatalf("MkdirAll(src) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "demo.go"), []byte("package main\nfunc Demo() {}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(demo.go) returned error: %v", err)
	}
}

func containsRaw(lines []string, needle string) bool {
	for _, line := range lines {
		if line == needle {
			return true
		}
	}
	return false
}
