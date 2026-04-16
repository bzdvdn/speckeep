package gitutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// skipIfNoGit skips the test when git is not available in PATH.
func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}
}

// initRepo creates a bare git repo with a first commit so branch commands work.
func initRepo(t *testing.T, dir string) {
	t.Helper()
	cmds := [][]string{
		{"git", "init", "-b", "main"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
}

// seedCommit creates an initial commit so the repo has a HEAD.
func seedCommit(t *testing.T, dir string) {
	t.Helper()
	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("# test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmds := [][]string{
		{"git", "add", "README.md"},
		{"git", "commit", "-m", "init"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
}

func TestEnsureRepositoryCreatesGitDir(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()

	created, err := EnsureRepository(dir)
	if err != nil {
		t.Fatalf("EnsureRepository: %v", err)
	}
	if !created {
		t.Error("expected created=true for new repo")
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		t.Errorf(".git directory not created: %v", err)
	}
}

func TestEnsureRepositoryNoOpsWhenAlreadyExists(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	initRepo(t, dir)

	created, err := EnsureRepository(dir)
	if err != nil {
		t.Fatalf("EnsureRepository: %v", err)
	}
	if created {
		t.Error("expected created=false when .git already exists")
	}
}

func TestCurrentBranchReturnsName(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	initRepo(t, dir)
	seedCommit(t, dir)

	branch, err := CurrentBranch(dir)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if branch == "" {
		t.Error("expected non-empty branch name")
	}
}

func TestCurrentBranchReturnsCorrectNameAfterCheckout(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	initRepo(t, dir)
	seedCommit(t, dir)

	// Create and switch to a feature branch
	cmd := exec.Command("git", "checkout", "-b", "feature/my-feature")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git checkout -b: %v\n%s", err, out)
	}

	branch, err := CurrentBranch(dir)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if branch != "feature/my-feature" {
		t.Errorf("CurrentBranch() = %q, want %q", branch, "feature/my-feature")
	}
}

func TestCurrentBranchReturnsHEADWhenDetached(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	initRepo(t, dir)
	seedCommit(t, dir)

	// Get commit hash and detach HEAD
	out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}
	hash := strings.TrimSpace(string(out))

	cmd := exec.Command("git", "checkout", hash)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git checkout (detach): %v\n%s", err, out)
	}

	branch, err := CurrentBranch(dir)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if branch != "HEAD" {
		t.Errorf("CurrentBranch() = %q in detached state, want %q", branch, "HEAD")
	}
}

func TestEnsureBranchCreatesNewBranch(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	initRepo(t, dir)
	seedCommit(t, dir)

	msg, err := EnsureBranch(dir, "feature/new-feature")
	if err != nil {
		t.Fatalf("EnsureBranch: %v", err)
	}
	if !strings.Contains(msg, "created") {
		t.Errorf("expected message to mention 'created', got %q", msg)
	}

	branch, err := CurrentBranch(dir)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if branch != "feature/new-feature" {
		t.Errorf("CurrentBranch() = %q, want %q", branch, "feature/new-feature")
	}
}

func TestEnsureBranchSwitchesToExistingBranch(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	initRepo(t, dir)
	seedCommit(t, dir)

	// Create branch first
	if _, err := EnsureBranch(dir, "feature/existing"); err != nil {
		t.Fatalf("EnsureBranch (create): %v", err)
	}
	// Switch back to main
	cmd := exec.Command("git", "checkout", "main")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git checkout main: %v\n%s", err, out)
	}
	// Now EnsureBranch should switch to existing
	msg, err := EnsureBranch(dir, "feature/existing")
	if err != nil {
		t.Fatalf("EnsureBranch (switch): %v", err)
	}
	if !strings.Contains(msg, "switched") {
		t.Errorf("expected message to mention 'switched', got %q", msg)
	}
	branch, err := CurrentBranch(dir)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if branch != "feature/existing" {
		t.Errorf("CurrentBranch() = %q, want %q", branch, "feature/existing")
	}
}
