package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"speckeep/src/internal/config"
	"speckeep/src/internal/gitutil"
)

func setupFinishRepo(t *testing.T, slug string, withProof bool) (string, string) {
	t.Helper()

	root := t.TempDir()
	if _, _, err := executeRoot(t, "init", root, "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init returned error: %v", err)
	}
	if err := runGit(root, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config email: %v", err)
	}
	if err := runGit(root, "config", "user.name", "Test"); err != nil {
		t.Fatalf("git config name: %v", err)
	}
	if err := gitutil.CommitAll(context.Background(), root, "init"); err != nil {
		t.Fatalf("initial commit: %v", err)
	}

	target, err := gitutil.CurrentBranch(context.Background(), root)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if err := runGit(root, "checkout", "-b", "feature/"+slug); err != nil {
		t.Fatalf("create feature branch: %v", err)
	}

	specDir := ensureSpecDir(t, root, slug)
	specContent := "# " + slug + "\n\n## Goal\nx\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 Demo\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec): %v", err)
	}
	tasksContent := "# Tasks\n\n## Phase 1\n- [x] T1.1 implement demo\n  Proof: code src/demo.go Demo\n\n## Acceptance Coverage\n- AC-001 -> T1.1\n"
	if !withProof {
		tasksContent = "# Tasks\n\n## Phase 1\n- [x] T1.1 implement demo\n\n## Acceptance Coverage\n- AC-001 -> T1.1\n"
	}
	if err := os.WriteFile(filepath.Join(specDir, "tasks.md"), []byte(tasksContent), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks): %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatalf("MkdirAll(src): %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "demo.go"), []byte("package main\n\nfunc Demo() {}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(demo.go): %v", err)
	}
	if err := gitutil.CommitAll(context.Background(), root, "feature: "+slug); err != nil {
		t.Fatalf("feature commit: %v", err)
	}

	return root, target
}

func TestFinishArchivesMergesAndDeletesBranch(t *testing.T) {
	root, target := setupFinishRepo(t, "demo", true)

	stdout, stderr, err := executeRoot(t, "finish", "demo", root)
	if err != nil {
		t.Fatalf("finish returned error: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "Finished demo") {
		t.Fatalf("unexpected finish output: %s", stdout)
	}

	current, err := gitutil.CurrentBranch(context.Background(), root)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if current != target {
		t.Fatalf("expected current branch %s, got %s", target, current)
	}
	exists, err := gitutil.BranchExists(context.Background(), root, "feature/demo")
	if err != nil {
		t.Fatalf("BranchExists: %v", err)
	}
	if exists {
		t.Fatalf("expected feature/demo to be deleted")
	}

	cfg, err := config.Load(context.Background(), root)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	specsDir, _ := cfg.SpecsDir(root)
	archiveDir, _ := cfg.ArchiveDir(root)
	if _, err := os.Stat(filepath.Join(specsDir, "demo", "spec.md")); !os.IsNotExist(err) {
		t.Fatalf("expected active spec to be archived, stat err=%v", err)
	}
	if entries, err := os.ReadDir(filepath.Join(archiveDir, "demo")); err != nil || len(entries) != 1 {
		t.Fatalf("expected one archive snapshot, entries=%v err=%v", entries, err)
	}
	if _, err := os.Stat(filepath.Join(root, "src", "demo.go")); err != nil {
		t.Fatalf("expected merged code on target: %v", err)
	}
}

func TestFinishDryRunMakesNoChanges(t *testing.T) {
	root, _ := setupFinishRepo(t, "demo", true)

	stdout, stderr, err := executeRoot(t, "finish", "demo", root, "--dry-run")
	if err != nil {
		t.Fatalf("finish --dry-run returned error: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "Dry run") {
		t.Fatalf("unexpected dry-run output: %s", stdout)
	}
	exists, _ := gitutil.BranchExists(context.Background(), root, "feature/demo")
	if !exists {
		t.Fatalf("dry-run must not delete the feature branch")
	}
	cfg, _ := config.Load(context.Background(), root)
	specsDir, _ := cfg.SpecsDir(root)
	if _, err := os.Stat(filepath.Join(specsDir, "demo", "spec.md")); err != nil {
		t.Fatalf("dry-run must not archive the feature: %v", err)
	}
}

func TestFinishRefusesDirtyWorktree(t *testing.T) {
	root, _ := setupFinishRepo(t, "demo", true)
	if err := os.WriteFile(filepath.Join(root, "dirty.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile(dirty): %v", err)
	}

	_, stderr, err := executeRoot(t, "finish", "demo", root)
	if err == nil {
		t.Fatalf("expected finish to refuse a dirty worktree")
	}
	if !strings.Contains(stderr, "worktree has uncommitted changes") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestFinishRefusesMissingProof(t *testing.T) {
	root, _ := setupFinishRepo(t, "demo", false)

	_, stderr, err := executeRoot(t, "finish", "demo", root)
	if err == nil {
		t.Fatalf("expected finish to refuse missing Proof")
	}
	if !strings.Contains(stderr, "without Proof") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestFinishAbortsMergeConflictAndKeepsBranch(t *testing.T) {
	root, target := setupFinishRepo(t, "demo", true)

	if err := runGit(root, "checkout", target); err != nil {
		t.Fatalf("checkout target: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatalf("MkdirAll(target src): %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "demo.go"), []byte("package main\n\nfunc Demo() { /* target */ }\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(target demo.go): %v", err)
	}
	if err := gitutil.CommitAll(context.Background(), root, "target change"); err != nil {
		t.Fatalf("target commit: %v", err)
	}
	if err := runGit(root, "checkout", "feature/demo"); err != nil {
		t.Fatalf("checkout feature: %v", err)
	}

	_, stderr, err := executeRoot(t, "finish", "demo", root)
	if err == nil {
		t.Fatalf("expected finish to fail on merge conflict")
	}
	if !strings.Contains(stderr, "merge failed") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
	current, _ := gitutil.CurrentBranch(context.Background(), root)
	if current != "feature/demo" {
		t.Fatalf("expected conflict abort to return to feature/demo, got %s", current)
	}
	exists, _ := gitutil.BranchExists(context.Background(), root, "feature/demo")
	if !exists {
		t.Fatalf("conflict must keep the feature branch")
	}
}

func TestFinishAfterManualArchive(t *testing.T) {
	root, target := setupFinishRepo(t, "demo", true)

	if _, stderr, err := executeRoot(t, "archive", "demo", root); err != nil {
		t.Fatalf("archive returned error: %v\nstderr: %s", err, stderr)
	}
	if err := gitutil.CommitAll(context.Background(), root, "chore: archive demo"); err != nil {
		t.Fatalf("archive commit: %v", err)
	}

	if _, stderr, err := executeRoot(t, "finish", "demo", root); err != nil {
		t.Fatalf("finish after archive returned error: %v\nstderr: %s", err, stderr)
	}
	current, _ := gitutil.CurrentBranch(context.Background(), root)
	if current != target {
		t.Fatalf("expected current branch %s, got %s", target, current)
	}
	exists, _ := gitutil.BranchExists(context.Background(), root, "feature/demo")
	if exists {
		t.Fatalf("expected feature/demo to be deleted")
	}
}

func TestFinishKeepBranch(t *testing.T) {
	root, _ := setupFinishRepo(t, "demo", true)

	if _, stderr, err := executeRoot(t, "finish", "demo", root, "--keep-branch"); err != nil {
		t.Fatalf("finish --keep-branch returned error: %v\nstderr: %s", err, stderr)
	}
	exists, _ := gitutil.BranchExists(context.Background(), root, "feature/demo")
	if !exists {
		t.Fatalf("--keep-branch must keep feature/demo")
	}
}

func TestFinishRequiresFeatureBranchCheckedOut(t *testing.T) {
	root, target := setupFinishRepo(t, "demo", true)
	if err := runGit(root, "checkout", target); err != nil {
		t.Fatalf("checkout target: %v", err)
	}

	_, stderr, err := executeRoot(t, "finish", "demo", root)
	if err == nil {
		t.Fatalf("expected finish to require the feature branch")
	}
	if !strings.Contains(stderr, "switch to feature/demo") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestFinishRequiresGitRepository(t *testing.T) {
	root := t.TempDir()
	initTestProject(t, root)
	writeMinimalFeature(t, root, "demo")

	_, stderr, err := executeRoot(t, "finish", "demo", root)
	if err == nil {
		t.Fatalf("expected finish to require git")
	}
	if !strings.Contains(stderr, "requires a git repository") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestFinishRejectsMissingTarget(t *testing.T) {
	root, _ := setupFinishRepo(t, "demo", true)

	_, stderr, err := executeRoot(t, "finish", "demo", root, "--into", "missing")
	if err == nil {
		t.Fatalf("expected finish to reject a missing target")
	}
	if !strings.Contains(stderr, "target branch missing does not exist") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestCheckShowsFinishAltForArchiveReadyFeature(t *testing.T) {
	root, _ := setupFinishRepo(t, "demo", true)

	stdout, _, err := executeRoot(t, "check", "demo", root, "--json")
	if err != nil {
		t.Fatalf("check --json returned error: %v", err)
	}
	var payload struct {
		NextCommand string `json:"next_command"`
		AltCommand  string `json:"alt_command"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("parse check json: %v\n%s", err, stdout)
	}
	if payload.NextCommand != "speckeep archive demo ." {
		t.Fatalf("next_command = %q, want speckeep archive demo .", payload.NextCommand)
	}
	if payload.AltCommand != "speckeep finish demo ." {
		t.Fatalf("alt_command = %q, want speckeep finish demo .", payload.AltCommand)
	}

	text, _, err := executeRoot(t, "check", "demo", root)
	if err != nil {
		t.Fatalf("check returned error: %v", err)
	}
	if !strings.Contains(text, "alt:      speckeep finish demo .") {
		t.Fatalf("expected finish alt in check output, got %s", text)
	}
}

func TestFinishMergesIntoNamedBranch(t *testing.T) {
	root, target := setupFinishRepo(t, "demo", true)
	if err := runGit(root, "branch", "dev", target); err != nil {
		t.Fatalf("create dev branch: %v", err)
	}

	stdout, stderr, err := executeRoot(t, "finish", "demo", root, "--into", "dev")
	if err != nil {
		t.Fatalf("finish --into dev returned error: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "merged into dev") {
		t.Fatalf("expected output to mention dev merge, got %s", stdout)
	}
	current, _ := gitutil.CurrentBranch(context.Background(), root)
	if current != "dev" {
		t.Fatalf("expected current branch dev, got %s", current)
	}
	exists, _ := gitutil.BranchExists(context.Background(), root, "feature/demo")
	if exists {
		t.Fatalf("expected feature/demo to be deleted")
	}
}
