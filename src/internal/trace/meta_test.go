package trace

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"speckeep/src/internal/config"
	"speckeep/src/internal/gitutil"
	"speckeep/src/internal/project"
)

func initTraceProject(t *testing.T, branch string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}
	root := t.TempDir()
	traceRunGit(t, root, "init", "-q", "-b", branch)
	traceRunGit(t, root, "config", "user.email", "test@example.com")
	traceRunGit(t, root, "config", "user.name", "Test")
	if _, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"}); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	traceCommit(t, root, "init")
	return root
}

func TestResolveBaseDeclaredWithoutGit(t *testing.T) {
	root := t.TempDir()
	if _, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"}); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	cfg := specsDirForTest(t, root)
	meta, err := ResolveBase(context.Background(), root, cfg, "demo")
	if err != nil {
		t.Fatalf("ResolveBase: %v", err)
	}
	if meta.BaseSHA != "" || meta.BaseSource != "declared" {
		t.Errorf("meta = %+v, want declared without base", meta)
	}
}

func TestResolveBaseInfersFromMain(t *testing.T) {
	root := initTraceProject(t, "main")
	base, err := gitutil.Head(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gitutil.EnsureBranch(context.Background(), root, "feature/demo"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "feature.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	traceCommit(t, root, "feature")

	cfg := specsDirForTest(t, root)
	meta, err := ResolveBase(context.Background(), root, cfg, "demo")
	if err != nil {
		t.Fatalf("ResolveBase: %v", err)
	}
	if meta.BaseSHA != base {
		t.Errorf("base = %s, want %s", meta.BaseSHA, base)
	}
	if meta.BaseBranch != "main" || meta.BaseSource != "inferred" {
		t.Errorf("meta = %+v", meta)
	}
}

func TestResolveBaseFallsBackToMaster(t *testing.T) {
	root := initTraceProject(t, "master")
	if _, err := gitutil.EnsureBranch(context.Background(), root, "feature/demo"); err != nil {
		t.Fatal(err)
	}
	cfg := specsDirForTest(t, root)
	meta, err := ResolveBase(context.Background(), root, cfg, "demo")
	if err != nil {
		t.Fatalf("ResolveBase: %v", err)
	}
	if meta.BaseBranch != "master" {
		t.Errorf("base branch = %q, want master", meta.BaseBranch)
	}
}

func TestResolveBaseUsesExistingMeta(t *testing.T) {
	root := initTraceProject(t, "main")
	cfg := specsDirForTest(t, root)
	want := Meta{Version: 1, Slug: "demo", BaseSHA: "abc123", BaseSource: "explicit"}
	if err := WriteMeta(cfg, "demo", want); err != nil {
		t.Fatal(err)
	}
	meta, err := ResolveBase(context.Background(), root, cfg, "demo")
	if err != nil {
		t.Fatalf("ResolveBase: %v", err)
	}
	if meta.BaseSHA != want.BaseSHA {
		t.Errorf("base = %s, want %s", meta.BaseSHA, want.BaseSHA)
	}
}

func TestInitMetaExplicitBaseAndForce(t *testing.T) {
	root := initTraceProject(t, "main")
	first, err := gitutil.Head(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "second.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	traceCommit(t, root, "second")
	second, err := gitutil.Head(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}

	meta, changed, err := InitMeta(context.Background(), root, "demo", first, false)
	if err != nil || !changed {
		t.Fatalf("InitMeta: changed=%v err=%v", changed, err)
	}
	if meta.BaseSHA != first || meta.BaseSource != "explicit" {
		t.Errorf("meta = %+v, want explicit base %s", meta, first)
	}

	meta, changed, err = InitMeta(context.Background(), root, "demo", second, true)
	if err != nil || !changed {
		t.Fatalf("InitMeta(force): changed=%v err=%v", changed, err)
	}
	if meta.BaseSHA != second {
		t.Errorf("forced base = %s, want %s", meta.BaseSHA, second)
	}

	// Without force the existing pin is kept.
	meta, changed, err = InitMeta(context.Background(), root, "demo", first, false)
	if err != nil {
		t.Fatal(err)
	}
	if changed || meta.BaseSHA != second {
		t.Errorf("idempotent call changed=%v base=%s, want unchanged %s", changed, meta.BaseSHA, second)
	}
}

func TestWriteArchiveMetaFileRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace-meta.json")
	want := ArchiveMeta{
		Meta:    Meta{Version: 1, Slug: "demo", BaseSHA: "base", BaseSource: "inferred", CreatedAt: "now"},
		HeadSHA: "head",
		Branch:  "feature/demo",
	}
	if err := WriteArchiveMetaFile(path, want); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got ArchiveMeta
	if err := json.Unmarshal(content, &got); err != nil {
		t.Fatal(err)
	}
	if got.BaseSHA != want.BaseSHA || got.HeadSHA != want.HeadSHA || got.Branch != want.Branch {
		t.Errorf("roundtrip = %+v, want %+v", got, want)
	}
}

// specsDirForTest returns the absolute specs dir resolved from the project config.
func specsDirForTest(t *testing.T, root string) string {
	t.Helper()
	cfg, err := config.Load(context.Background(), root)
	if err != nil {
		t.Fatalf("config load: %v", err)
	}
	dir, err := cfg.SpecsDir(root)
	if err != nil {
		t.Fatalf("SpecsDir: %v", err)
	}
	return dir
}

func TestInitMetaRejectsInvalidBase(t *testing.T) {
	root := initTraceProject(t, "main")
	if _, _, err := InitMeta(context.Background(), root, "demo", "does-not-exist", true); err == nil {
		t.Fatal("expected error for unresolvable base")
	}
}

func TestResolveBaseWithoutConventionalBranch(t *testing.T) {
	root := initTraceProject(t, "trunk")
	head, err := gitutil.Head(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	dir := specsDirForTest(t, root)

	// A single-commit repo falls back to its root commit.
	meta, err := ResolveBase(context.Background(), root, dir, "demo")
	if err != nil {
		t.Fatalf("ResolveBase: %v", err)
	}
	if meta.BaseSHA != head || meta.BaseSource != "inferred" {
		t.Errorf("meta = %+v, want root commit %s", meta, head)
	}

	// Multiple commits without main/master cannot be attributed safely.
	if err := os.WriteFile(filepath.Join(root, "second.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	traceCommit(t, root, "second")
	meta, err = ResolveBase(context.Background(), root, dir, "demo")
	if err != nil {
		t.Fatalf("ResolveBase: %v", err)
	}
	if meta.BaseSHA != "" || meta.BaseSource != "declared" {
		t.Errorf("meta = %+v, want declared", meta)
	}
}
