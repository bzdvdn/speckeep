package status

import (
	"os"
	"path/filepath"
	"testing"

	"speckeep/src/internal/project"
)

func TestCheckInfersPhaseAcrossFeatureLifecycle(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	check := func(wantPhase, wantReadyFor string, wantBlocked bool) {
		t.Helper()
		result, err := Check(root, "demo")
		if err != nil {
			t.Fatalf("Check returned error: %v", err)
		}
		if result.Phase != wantPhase {
			t.Fatalf("phase = %q, want %q", result.Phase, wantPhase)
		}
		if result.ReadyFor != wantReadyFor {
			t.Fatalf("ready_for = %q, want %q", result.ReadyFor, wantReadyFor)
		}
		if result.Blocked != wantBlocked {
			t.Fatalf("blocked = %v, want %v", result.Blocked, wantBlocked)
		}
	}

	check("constitution", "spec", true)

	specDir := filepath.Join(root, ".speckeep", "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(specDir) returned error: %v", err)
	}
	specPath := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specPath, []byte("# Demo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}
	check("spec", "inspect", false)

	inspectPath := filepath.Join(specDir, "inspect.md")
	inspectContent := "---\nreport_type: inspect\nslug: demo\nstatus: pass\ndocs_language: en\ngenerated_at: 2026-03-30\n---\n# Inspect Report: demo\n\n## Verdict\n\n- status: pass\n"
	if err := os.WriteFile(inspectPath, []byte(inspectContent), 0o644); err != nil {
		t.Fatalf("WriteFile(inspect) returned error: %v", err)
	}
	check("inspect", "plan", false)

	planDir := filepath.Join(root, ".speckeep", "specs", "demo", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(planDir) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "plan.md"), []byte("# Demo Plan\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(plan) returned error: %v", err)
	}
	check("plan", "tasks", false)

	tasksContent := "# Demo Tasks\n\n- [x] T1.1 Done task\n- [ ] T1.2 Open task\n"
	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte(tasksContent), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}

	result, err := Check(root, "demo")
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if result.Phase != "implement" || result.ReadyFor != "implement" {
		t.Fatalf("unexpected implement status: %+v", result)
	}
	if result.TasksTotal != 2 || result.TasksCompleted != 1 || result.TasksOpen != 1 {
		t.Fatalf("unexpected task counts: %+v", result)
	}

	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte("# Demo Tasks\n\n- [x] T1.1 Done\n- [x] T1.2 Done\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks complete) returned error: %v", err)
	}
	check("verify", "verify", false)

	verifyContent := "---\nreport_type: verify\nslug: demo\nstatus: pass\ndocs_language: en\ngenerated_at: 2026-03-30\n---\n# Verify Report: demo\n\n## Verdict\n\n- status: pass\n"
	if err := os.WriteFile(filepath.Join(planDir, "verify.md"), []byte(verifyContent), 0o644); err != nil {
		t.Fatalf("WriteFile(verify) returned error: %v", err)
	}
	check("verify", "archive", false)

	archiveDir := filepath.Join(root, ".speckeep", "archive", "demo", "2026-03-30")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(archiveDir) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(archiveDir, "summary.md"), []byte("# Summary\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(summary) returned error: %v", err)
	}
	check("archive", "", false)
}
