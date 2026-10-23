package gitutil

import (
	"context"
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

	created, err := EnsureRepository(context.Background(), dir)
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

	created, err := EnsureRepository(context.Background(), dir)
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

	branch, err := CurrentBranch(context.Background(), dir)
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

	branch, err := CurrentBranch(context.Background(), dir)
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

	branch, err := CurrentBranch(context.Background(), dir)
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

	msg, err := EnsureBranch(context.Background(), dir, "feature/new-feature")
	if err != nil {
		t.Fatalf("EnsureBranch: %v", err)
	}
	if !strings.Contains(msg, "created") {
		t.Errorf("expected message to mention 'created', got %q", msg)
	}

	branch, err := CurrentBranch(context.Background(), dir)
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
	if _, err := EnsureBranch(context.Background(), dir, "feature/existing"); err != nil {
		t.Fatalf("EnsureBranch (create): %v", err)
	}
	// Switch back to main
	cmd := exec.Command("git", "checkout", "main")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git checkout main: %v\n%s", err, out)
	}
	// Now EnsureBranch should switch to existing
	msg, err := EnsureBranch(context.Background(), dir, "feature/existing")
	if err != nil {
		t.Fatalf("EnsureBranch (switch): %v", err)
	}
	if !strings.Contains(msg, "switched") {
		t.Errorf("expected message to mention 'switched', got %q", msg)
	}
	branch, err := CurrentBranch(context.Background(), dir)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if branch != "feature/existing" {
		t.Errorf("CurrentBranch() = %q, want %q", branch, "feature/existing")
	}
}

func testGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func writeAndCommit(t *testing.T, dir, path, content, message string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	testGit(t, dir, "add", "-A")
	testGit(t, dir, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-m", message)
}

func TestMergeBaseAndDiffFiles(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	initRepo(t, dir)
	seedCommit(t, dir)

	base, err := ResolveRef(context.Background(), dir, "HEAD")
	if err != nil {
		t.Fatalf("ResolveRef: %v", err)
	}
	if _, err := EnsureBranch(context.Background(), dir, "feature/x"); err != nil {
		t.Fatalf("EnsureBranch: %v", err)
	}

	writeAndCommit(t, dir, "src/a.go", "package src\n", "add a")
	writeAndCommit(t, dir, "src/b.go", "package src\n", "add b")

	mergeBase, err := MergeBase(context.Background(), dir, "main", "HEAD")
	if err != nil {
		t.Fatalf("MergeBase: %v", err)
	}
	if mergeBase != base {
		t.Errorf("MergeBase = %s, want %s", mergeBase, base)
	}

	files, err := DiffFiles(context.Background(), dir, base, "HEAD")
	if err != nil {
		t.Fatalf("DiffFiles: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("DiffFiles returned %d files, want 2: %+v", len(files), files)
	}
	for _, file := range files {
		if file.Status != "A" {
			t.Errorf("%s status = %s, want A", file.Path, file.Status)
		}
		if file.Additions == 0 {
			t.Errorf("%s additions = 0, want > 0", file.Path)
		}
	}

	// Rename an unchanged file: git should report R with both paths.
	base2, err := ResolveRef(context.Background(), dir, "HEAD")
	if err != nil {
		t.Fatalf("ResolveRef: %v", err)
	}
	testGit(t, dir, "mv", "src/b.go", "src/c.go")
	testGit(t, dir, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-m", "rename b")
	files, err = DiffFiles(context.Background(), dir, base2, "HEAD")
	if err != nil {
		t.Fatalf("DiffFiles: %v", err)
	}
	foundRename := false
	for _, file := range files {
		if file.Status == "R" && file.OldPath == "src/b.go" && file.Path == "src/c.go" {
			foundRename = true
		}
	}
	if !foundRename {
		t.Errorf("rename not detected: %+v", files)
	}

	// Modify and delete files.
	base3, err := ResolveRef(context.Background(), dir, "HEAD")
	if err != nil {
		t.Fatalf("ResolveRef: %v", err)
	}
	writeAndCommit(t, dir, "src/a.go", "package src\n\nfunc A() {}\n", "change a")
	testGit(t, dir, "rm", "-q", "src/c.go")
	testGit(t, dir, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-m", "remove c")

	files, err = DiffFiles(context.Background(), dir, base3, "HEAD")
	if err != nil {
		t.Fatalf("DiffFiles: %v", err)
	}
	status := map[string]string{}
	for _, file := range files {
		status[file.Path] = file.Status
	}
	if status["src/a.go"] != "M" || status["src/c.go"] != "D" {
		t.Errorf("statuses = %v, want a=M c=D", status)
	}
}

func TestTrackedDirty(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	initRepo(t, dir)
	seedCommit(t, dir)

	if TrackedDirty(context.Background(), dir) {
		t.Fatal("fresh repo reported dirty")
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !TrackedDirty(context.Background(), dir) {
		t.Error("modified tracked file not reported dirty")
	}
}

func TestRepositoryHelpers(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	if IsRepository(context.Background(), dir) {
		t.Fatal("plain directory reported as git repository")
	}
	initRepo(t, dir)
	seedCommit(t, dir)
	if !IsRepository(context.Background(), dir) {
		t.Fatal("git repository not detected")
	}

	rootSHA, err := Head(context.Background(), dir)
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	if count, err := CommitCount(context.Background(), dir, "HEAD"); err != nil || count != 1 {
		t.Errorf("CommitCount = %d, %v; want 1", count, err)
	}
	if got, err := RootCommit(context.Background(), dir, "HEAD"); err != nil || got != rootSHA {
		t.Errorf("RootCommit = %q, %v; want %q", got, err, rootSHA)
	}
	if !IsAncestor(context.Background(), dir, rootSHA, "HEAD") {
		t.Error("root should be an ancestor of HEAD")
	}
	if content, err := FileAt(context.Background(), dir, "HEAD", "README.md"); err != nil || !strings.Contains(string(content), "# test") {
		t.Errorf("FileAt = %q, %v", content, err)
	}

	writeAndCommit(t, dir, "src/a.go", "package src\n", "add a")
	if IsAncestor(context.Background(), dir, "HEAD", rootSHA) {
		t.Error("HEAD must not be an ancestor of root after a second commit")
	}
	if count, err := CommitCount(context.Background(), dir, "HEAD"); err != nil || count != 2 {
		t.Errorf("CommitCount after second commit = %d, %v; want 2", count, err)
	}
	commits, err := Commits(context.Background(), dir, rootSHA, "HEAD")
	if err != nil {
		t.Fatalf("Commits: %v", err)
	}
	if len(commits) != 1 || commits[0].Subject != "add a" || commits[0].SHA == "" {
		t.Errorf("Commits = %+v, want one 'add a' commit", commits)
	}

	if err := os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	untracked, err := UntrackedFiles(context.Background(), dir)
	if err != nil {
		t.Fatalf("UntrackedFiles: %v", err)
	}
	if len(untracked) != 1 || untracked[0] != "untracked.txt" {
		t.Errorf("UntrackedFiles = %v", untracked)
	}
}

func TestEnsureBranchCreatesOrphanBranchWithoutCommits(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	initRepo(t, dir)

	msg, err := EnsureBranch(context.Background(), dir, "feature/orphan")
	if err != nil {
		t.Fatalf("EnsureBranch: %v", err)
	}
	if !strings.Contains(msg, "orphan") {
		t.Errorf("message = %q, want orphan", msg)
	}
	out, _, err := run(context.Background(), dir, "git", "symbolic-ref", "--short", "HEAD")
	if err != nil {
		t.Fatalf("symbolic-ref: %v", err)
	}
	if strings.TrimSpace(out) != "feature/orphan" {
		t.Errorf("HEAD = %q, want feature/orphan", strings.TrimSpace(out))
	}
}

func TestFileAtMissingPathReturnsError(t *testing.T) {
	skipIfNoGit(t)
	dir := t.TempDir()
	initRepo(t, dir)
	seedCommit(t, dir)

	if _, err := FileAt(context.Background(), dir, "HEAD", "missing.txt"); err == nil {
		t.Fatal("expected error for missing path")
	}
}
