package workflow

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"speckeep/src/internal/gitutil"
	"speckeep/src/internal/project"
	"speckeep/src/internal/trace"
)

func runTraceDriftGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func commitTraceDrift(t *testing.T, root, message string) {
	t.Helper()
	runTraceDriftGit(t, root, "add", "-A")
	runTraceDriftGit(t, root, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-q", "-m", message)
}

func setupTraceDriftProject(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}
	root := t.TempDir()
	runTraceDriftGit(t, root, "init", "-q", "-b", "main")
	runTraceDriftGit(t, root, "config", "user.email", "test@example.com")
	runTraceDriftGit(t, root, "config", "user.name", "Test")
	if _, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"}); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	commitTraceDrift(t, root, "init")
	if _, err := gitutil.EnsureBranch(context.Background(), root, "feature/demo"); err != nil {
		t.Fatal(err)
	}

	specDir := filepath.Join(root, "specs", "active", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	tasks := "# Tasks\n\n## Phase 1\n- [x] T1.1 Add demo (AC-001). Touches: src/a.go\n  Proof: code src/a.go A\n"
	if err := os.WriteFile(filepath.Join(specDir, "tasks.md"), []byte(tasks), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := trace.InitMeta(context.Background(), root, "demo", "main", false); err != nil {
		t.Fatalf("InitMeta: %v", err)
	}
	commitTraceDrift(t, root, "spec")

	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "a.go"), []byte("package src\n\nfunc A() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	commitTraceDrift(t, root, "implement")
	return root
}

func TestCheckTraceDriftCleanAndWarnings(t *testing.T) {
	root := setupTraceDriftProject(t)

	var clean CheckResult
	checkTraceDrift(context.Background(), &clean, root, "demo")
	if !containsRaw(clean.Lines, "OK: trace: change attribution matches declared surfaces") {
		t.Fatalf("expected clean trace result, got %+v", clean.Lines)
	}

	// An untracked file outside Touches must produce an Unmapped warning.
	if err := os.WriteFile(filepath.Join(root, "src", "extra.go"), []byte("package src\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var drift CheckResult
	checkTraceDrift(context.Background(), &drift, root, "demo")
	if !hasLineContaining(drift.Lines, "WARN: trace: 1 changed file(s) are outside declared Touches") {
		t.Fatalf("expected unmapped warning, got %+v", drift.Lines)
	}

	// A closed task that declares an untouched surface must warn too.
	tasksPath := filepath.Join(root, "specs", "active", "demo", "tasks.md")
	content, err := os.ReadFile(tasksPath)
	if err != nil {
		t.Fatal(err)
	}
	content = append(content, []byte("- [x] T1.2 Docs (AC-002). Touches: docs/planned.md\n  Proof: docs docs/planned.md\n")...)
	if err := os.WriteFile(tasksPath, content, 0o644); err != nil {
		t.Fatal(err)
	}
	var notChanged CheckResult
	checkTraceDrift(context.Background(), &notChanged, root, "demo")
	if !hasLineContaining(notChanged.Lines, "declared surface(s) were not changed") {
		t.Fatalf("expected declared-not-changed warning, got %+v", notChanged.Lines)
	}
}

func TestCheckTraceDriftSkipsWithoutMeta(t *testing.T) {
	root := t.TempDir()
	if _, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"}); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	var result CheckResult
	checkTraceDrift(context.Background(), &result, root, "demo")
	if len(result.Lines) != 0 {
		t.Errorf("expected no findings without meta, got %+v", result.Lines)
	}
}

func hasLineContaining(lines []string, needle string) bool {
	for _, line := range lines {
		if strings.Contains(line, needle) {
			return true
		}
	}
	return false
}
