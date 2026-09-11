package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"speckeep/src/internal/config"
	"speckeep/src/internal/gitutil"
)

// setupCompactFeature initializes a git-backed project with a minimal closeable
// feature: spec.md + tasks.md (+ plan + data-model) fully checked with Proof.
func setupCompactFeature(t *testing.T, root, slug string) {
	t.Helper()
	if _, _, err := executeRoot(t, "init", root, "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init returned error: %v", err)
	}
	if ok, err := gitutil.EnsureRepository(context.Background(), root); err != nil {
		t.Fatalf("EnsureRepository returned error: %v", err)
	} else if !ok {
		t.Logf("git repo already present")
	}
	specDir := ensureSpecDir(t, root, slug)
	specContent := "# " + slug + "\n\n## Goal\nx\n\n## Acceptance Criteria\n### AC-001 Demo\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec): %v", err)
	}
	tasksContent := "# Tasks\n\n## Phase 1\n- [x] T1.1 implement demo\n  Proof: code src/demo.go Demo\n\n## Acceptance Coverage\n- AC-001 -> T1.1\n"
	if err := os.WriteFile(filepath.Join(specDir, "tasks.md"), []byte(tasksContent), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks): %v", err)
	}
	proofSrc := filepath.Join(root, "src")
	if err := os.MkdirAll(proofSrc, 0o755); err != nil {
		t.Fatalf("MkdirAll(src): %v", err)
	}
	if err := os.WriteFile(filepath.Join(proofSrc, "demo.go"), []byte("package main\nfunc Demo() {}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(demo.go): %v", err)
	}
	// Commit everything so --compact has a valid git pointer.
	for _, cmdArgs := range [][]string{
		{"add", "-A"},
		{"-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-m", "feature: " + slug},
	} {
		if err := runGit(root, cmdArgs...); err != nil {
			t.Fatalf("git %v returned error: %v", cmdArgs, err)
		}
	}
}

func runGit(root string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	_ = out
	return nil
}

func TestArchiveCompactRoundtrip(t *testing.T) {
	root := t.TempDir()
	setupCompactFeature(t, root, "demo")

	cfg, err := config.Load(context.Background(), root)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	specsDir, _ := cfg.SpecsDir(root)
	archiveDir, _ := cfg.ArchiveDir(root)

	// Full closeable path needs tasks open count 0 and Proof present; archive it compact.
	_, stderr, err := executeRoot(t, "archive", "demo", root, "--compact", "--status", "completed")
	if err != nil {
		t.Fatalf("compact archive should succeed, got: %v\nstderr: %s", err, stderr)
	}

	// Active feature artifacts must be gone.
	if _, err := os.Stat(filepath.Join(specsDir, "demo", "spec.md")); !os.IsNotExist(err) {
		t.Fatalf("expected active spec.md to be removed, stat err=%v", err)
	}

	// Archive holds only summary.md + snapshot.sha (no artifact copies).
	entries, err := os.ReadDir(filepath.Join(archiveDir, "demo"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected one snapshot, got entries=%v err=%v", entries, err)
	}
	snapshotDir := filepath.Join(archiveDir, "demo", entries[0].Name())
	snapEntries, _ := os.ReadDir(snapshotDir)
	if len(snapEntries) != 2 {
		t.Fatalf("expected compact snapshot to hold exactly summary.md + snapshot.sha, got %d files", len(snapEntries))
	}
	if _, err := os.Stat(filepath.Join(snapshotDir, "snapshot.sha")); err != nil {
		t.Fatalf("expected snapshot.sha, got %v", err)
	}

	// Restore from the compact snapshot.
	_, restoreErr, err := executeRoot(t, "archive", "demo", root, "--restore")
	if err != nil {
		t.Fatalf("compact restore should succeed, got: %v\nstderr: %s", err, restoreErr)
	}
	restoredSpec, err := os.ReadFile(filepath.Join(specsDir, "demo", "spec.md"))
	if err != nil {
		t.Fatalf("expected restored spec.md, got %v", err)
	}
	if !strings.Contains(string(restoredSpec), "# demo") {
		t.Fatalf("unexpected restored spec content: %s", restoredSpec)
	}
	restoredTasks, err := os.ReadFile(filepath.Join(specsDir, "demo", "tasks.md"))
	if err != nil {
		t.Fatalf("expected restored tasks.md, got %v", err)
	}
	if !strings.Contains(string(restoredTasks), "T1.1") {
		t.Fatalf("unexpected restored tasks content: %s", restoredTasks)
	}
}

func TestArchiveCompactRequiresGitRepo(t *testing.T) {
	root := t.TempDir()
	// init without git
	initTestProject(t, root)
	writeMinimalFeature(t, root, "demo")

	_, stderr, err := executeRoot(t, "archive", "demo", root, "--compact", "--status", "completed")
	if err == nil {
		t.Fatalf("expected --compact to fail without a git repo")
	}
	if !strings.Contains(stderr, "--compact requires a git repository") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}
