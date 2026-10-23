package trace

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"speckeep/src/internal/gitutil"
	"speckeep/src/internal/project"
)

const changesTasks = `# Tasks

## Phase 1

- [x] T1.1 Add export (AC-001). Touches: src/export.go
  Proof: code src/export.go RunExport
- [x] T1.2 Plan docs (AC-002). Touches: src/planned.go
  Proof: docs docs/planned.md
- [ ] T2.1 Later work. Touches: src/later.go
`

func traceRunGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func traceCommit(t *testing.T, root, message string) {
	t.Helper()
	traceRunGit(t, root, "add", "-A")
	traceRunGit(t, root, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-m", message)
}

func setupTraceFeature(t *testing.T) (root, base string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}
	root = t.TempDir()
	traceRunGit(t, root, "init", "-b", "main")
	traceRunGit(t, root, "config", "user.email", "test@example.com")
	traceRunGit(t, root, "config", "user.name", "Test")

	if _, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"}); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "base.go"), []byte("package src\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	traceCommit(t, root, "init")

	sha, err := gitutil.Head(context.Background(), root)
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	base = sha

	if _, err := gitutil.EnsureBranch(context.Background(), root, "feature/demo"); err != nil {
		t.Fatalf("EnsureBranch: %v", err)
	}
	specDir := filepath.Join(root, "specs", "active", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "tasks.md"), []byte(changesTasks), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := InitMeta(context.Background(), root, "demo", "main", false); err != nil {
		t.Fatalf("InitMeta: %v", err)
	}
	traceCommit(t, root, "spec")

	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "export.go"), []byte("package src\n\nfunc RunExport() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "other.go"), []byte("package src\n\nfunc Other() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	traceCommit(t, root, "implement")
	return root, base
}

func TestBuildChangesMapsTasks(t *testing.T) {
	root, base := setupTraceFeature(t)

	changes, err := BuildChanges(context.Background(), root, "demo")
	if err != nil {
		t.Fatalf("BuildChanges: %v", err)
	}
	if changes.Source != "git" {
		t.Fatalf("source = %s, want git", changes.Source)
	}
	if changes.BaseSHA != base {
		t.Errorf("base = %s, want %s", changes.BaseSHA, base)
	}
	if changes.HeadSHA == "" || len(changes.Commits) == 0 {
		t.Errorf("expected head and commits: %+v", changes)
	}

	byPath := map[string]ChangedFile{}
	for _, file := range changes.Files {
		byPath[file.Path] = file
	}
	export, ok := byPath["src/export.go"]
	if !ok {
		t.Fatalf("src/export.go missing from changes: %+v", changes.Files)
	}
	if !reflect.DeepEqual(export.Tasks, []string{"T1.1"}) {
		t.Errorf("export tasks = %v, want T1.1", export.Tasks)
	}
	if !reflect.DeepEqual(export.ACs, []string{"AC-001"}) {
		t.Errorf("export ACs = %v", export.ACs)
	}
	if !reflect.DeepEqual(export.Anchors, []string{"RunExport"}) {
		t.Errorf("export anchors = %v", export.Anchors)
	}
	if !reflect.DeepEqual(changes.Unmapped, []string{"src/other.go"}) {
		t.Errorf("unmapped = %v", changes.Unmapped)
	}
	if !reflect.DeepEqual(changes.NotChanged, []string{"docs/planned.md", "src/planned.go"}) {
		t.Errorf("declared not changed = %v", changes.NotChanged)
	}
	for _, ref := range changes.NotChanged {
		if ref == "src/later.go" {
			t.Errorf("open task surface reported as drift: %v", changes.NotChanged)
		}
	}
}

func TestBuildChangesAtDeclaredFallback(t *testing.T) {
	changes, err := BuildChangesAt(context.Background(), t.TempDir(), "demo", "specs/active", "", "", changesTasks, Meta{Slug: "demo"})
	if err != nil {
		t.Fatalf("BuildChangesAt: %v", err)
	}
	if changes.Source != "declared" {
		t.Fatalf("source = %s, want declared", changes.Source)
	}
	if len(changes.Files) == 0 {
		t.Fatal("expected declared files")
	}
	for _, file := range changes.Files {
		if file.Status != "declared" {
			t.Errorf("%s status = %s, want declared", file.Path, file.Status)
		}
	}
}

func TestWriteAndReadChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "changes.json")
	want := ChangeSet{Version: 1, Slug: "demo", Source: "git", Files: []ChangedFile{{Path: "src/a.go", Tasks: []string{"T1.1"}}}}
	if err := WriteChanges(path, want); err != nil {
		t.Fatalf("WriteChanges: %v", err)
	}
	got, err := ReadChanges(path)
	if err != nil {
		t.Fatalf("ReadChanges: %v", err)
	}
	if !reflect.DeepEqual(got.Files, want.Files) || got.Slug != want.Slug {
		t.Errorf("roundtrip = %+v, want %+v", got, want)
	}
}

func TestBuildChangesIncludesWorktreeAndUntracked(t *testing.T) {
	root, _ := setupTraceFeature(t)

	if err := os.WriteFile(filepath.Join(root, "src", "base.go"), []byte("package src\n\nfunc Base() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "new.go"), []byte("package src\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	changes, err := BuildChanges(context.Background(), root, "demo")
	if err != nil {
		t.Fatalf("BuildChanges: %v", err)
	}
	if !changes.Worktree {
		t.Fatal("expected worktree mode for active feature")
	}
	status := map[string]string{}
	for _, file := range changes.Files {
		status[file.Path] = file.Status
	}
	if status["src/base.go"] != "M" {
		t.Errorf("modified file status = %q, want M", status["src/base.go"])
	}
	if status["src/new.go"] != "?" {
		t.Errorf("untracked file status = %q, want ?", status["src/new.go"])
	}
	found := false
	for _, file := range changes.Unmapped {
		if file == "src/new.go" {
			found = true
		}
	}
	if !found {
		t.Errorf("untracked file not reported as unmapped: %v", changes.Unmapped)
	}
}

func TestBuildChangesCommittedExcludesWorktree(t *testing.T) {
	root, _ := setupTraceFeature(t)

	if err := os.WriteFile(filepath.Join(root, "src", "base.go"), []byte("package src\n\nfunc Base() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "new.go"), []byte("package src\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	changes, err := BuildChangesCommitted(context.Background(), root, "demo")
	if err != nil {
		t.Fatalf("BuildChangesCommitted: %v", err)
	}
	if changes.Worktree {
		t.Error("committed changes must not be marked as worktree")
	}
	if !changes.Dirty {
		t.Error("expected dirty=true with uncommitted changes")
	}
	paths := map[string]bool{}
	for _, file := range changes.Files {
		paths[file.Path] = true
	}
	if paths["src/base.go"] || paths["src/new.go"] {
		t.Errorf("uncommitted files leaked into committed changes: %+v", changes.Files)
	}
	if !paths["src/export.go"] {
		t.Errorf("committed file missing: %+v", changes.Files)
	}
}

func TestBuildChangesHonoursIgnoreConfig(t *testing.T) {
	root, _ := setupTraceFeature(t)
	configPath := filepath.Join(root, ".speckeep", "speckeep.yaml")
	if err := os.WriteFile(configPath, []byte("version: 1\ntrace:\n  ignore:\n    - src/\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	changes, err := BuildChangesCommitted(context.Background(), root, "demo")
	if err != nil {
		t.Fatalf("BuildChangesCommitted: %v", err)
	}
	if len(changes.Files) != 0 || len(changes.Unmapped) != 0 {
		t.Errorf("ignored paths leaked into changes: files=%v unmapped=%v", changes.Files, changes.Unmapped)
	}
}

func TestMatchesRef(t *testing.T) {
	cases := []struct {
		file, ref string
		want      bool
	}{
		{"src/a.go", "src/a.go", true},
		{"src/a.go", "src/", true},
		{"src/dir/a.go", "src/*.go", false},
		{"src/dir/a.go", "src/dir/*.go", true},
		{"src/other.go", "src/a.go", false},
		{"src/a.go.bak", "src/a.go", false},
		{"src/a.go", "", false},
	}
	for _, tc := range cases {
		if got := MatchesRef(tc.file, tc.ref); got != tc.want {
			t.Errorf("MatchesRef(%q, %q) = %v, want %v", tc.file, tc.ref, got, tc.want)
		}
	}
}
