package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"speckeep/src/internal/config"
	"speckeep/src/internal/gitutil"
	"speckeep/src/internal/trace"
)

const traceSpec = "# demo\n\n## Goal\nx\n\n## Acceptance Criteria\n### AC-001 Demo\n- Given x\n- When y\n- Then z\n"

const traceTasks = `# Tasks

## Phase 1

- [x] T1.1 Add export (AC-001). Touches: src/export.go
  Proof: code src/export.go RunExport
`

func setupTraceCLIFeature(t *testing.T, root, slug string) {
	t.Helper()
	if err := runGit(root, "init", "-q", "-b", "main"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := runGit(root, "config", "user.email", "test@example.com"); err != nil {
		t.Fatal(err)
	}
	if err := runGit(root, "config", "user.name", "Test"); err != nil {
		t.Fatal(err)
	}
	if _, stderr, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init: %v\n%s", err, stderr)
	}
	if err := runGit(root, "add", "-A"); err != nil {
		t.Fatal(err)
	}
	if err := runGit(root, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-q", "-m", "init"); err != nil {
		t.Fatal(err)
	}
	if err := runGit(root, "checkout", "-q", "-b", "feature/"+slug); err != nil {
		t.Fatal(err)
	}

	specDir := ensureSpecDir(t, root, slug)
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(traceSpec), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "tasks.md"), []byte(traceTasks), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, stderr, err := executeRoot(t, "trace", "init", slug, root); err != nil {
		t.Fatalf("trace init: %v\n%s", err, stderr)
	}
	if err := runGit(root, "add", "-A"); err != nil {
		t.Fatal(err)
	}
	if err := runGit(root, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-q", "-m", "spec"); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "export.go"), []byte("package src\n\nfunc RunExport() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "other.go"), []byte("package src\n\nfunc Other() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runGit(root, "add", "-A"); err != nil {
		t.Fatal(err)
	}
	if err := runGit(root, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-q", "-m", "implement"); err != nil {
		t.Fatal(err)
	}
}

func decodeChangeSet(t *testing.T, stdout string) trace.ChangeSet {
	t.Helper()
	var changes trace.ChangeSet
	if err := json.Unmarshal([]byte(stdout), &changes); err != nil {
		t.Fatalf("unmarshal changes: %v\n%s", err, stdout)
	}
	return changes
}

func TestTraceFilesCommandMapsTasks(t *testing.T) {
	root := t.TempDir()
	setupTraceCLIFeature(t, root, "demo")

	stdout, stderr, err := executeRoot(t, "trace", "demo", root, "--files", "--json")
	if err != nil {
		t.Fatalf("trace --files: %v\n%s", err, stderr)
	}
	changes := decodeChangeSet(t, stdout)
	if changes.Source != "git" {
		t.Fatalf("source = %s, want git", changes.Source)
	}
	var export *trace.ChangedFile
	for i := range changes.Files {
		if changes.Files[i].Path == "src/export.go" {
			export = &changes.Files[i]
		}
	}
	if export == nil {
		t.Fatalf("src/export.go missing: %+v", changes.Files)
	}
	if !reflect.DeepEqual(export.Tasks, []string{"T1.1"}) || !reflect.DeepEqual(export.ACs, []string{"AC-001"}) {
		t.Errorf("export attribution = %+v", export)
	}
	if !reflect.DeepEqual(changes.Unmapped, []string{"src/other.go"}) {
		t.Errorf("unmapped = %v", changes.Unmapped)
	}
}

func TestTraceInitIsIdempotent(t *testing.T) {
	root := t.TempDir()
	setupTraceCLIFeature(t, root, "demo")
	cfg, err := config.Load(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	specsDir, _ := cfg.SpecsDir(root)
	before, ok, err := trace.LoadMeta(specsDir, "demo")
	if err != nil || !ok {
		t.Fatalf("LoadMeta: ok=%v err=%v", ok, err)
	}

	stdout, stderr, err := executeRoot(t, "trace", "init", "demo", root)
	if err != nil {
		t.Fatalf("trace init (again): %v\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "already pinned") {
		t.Errorf("expected idempotent message, got %q", stdout)
	}
	after, _, err := trace.LoadMeta(specsDir, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if after.BaseSHA != before.BaseSHA {
		t.Errorf("base changed: %s -> %s", before.BaseSHA, after.BaseSHA)
	}
}

func TestTraceArchivePersistsAndReverseLookup(t *testing.T) {
	root := t.TempDir()
	setupTraceCLIFeature(t, root, "demo")

	if _, stderr, err := executeRoot(t, "archive", "demo", root, "--status", "completed"); err != nil {
		t.Fatalf("archive: %v\n%s", err, stderr)
	}

	stdout, stderr, err := executeRoot(t, "trace", "demo", root, "--files", "--json")
	if err != nil {
		t.Fatalf("trace --files after archive: %v\n%s", err, stderr)
	}
	changes := decodeChangeSet(t, stdout)
	if changes.Source != "git" {
		t.Fatalf("source = %s, want archived git changes", changes.Source)
	}
	found := false
	for _, file := range changes.Files {
		if file.Path == "src/export.go" && reflect.DeepEqual(file.Tasks, []string{"T1.1"}) {
			found = true
		}
	}
	if !found {
		t.Errorf("archived changes lost task mapping: %+v", changes.Files)
	}

	stdout, stderr, err = executeRoot(t, "trace", "--file", "src/export.go", root, "--json")
	if err != nil {
		t.Fatalf("trace --file: %v\n%s", err, stderr)
	}
	var hits []TraceHit
	if err := json.Unmarshal([]byte(stdout), &hits); err != nil {
		t.Fatalf("unmarshal hits: %v\n%s", err, stdout)
	}
	if len(hits) == 0 {
		t.Fatal("no reverse lookup hits")
	}
	if hits[0].Source != "archived" || hits[0].Slug != "demo" || !reflect.DeepEqual(hits[0].Tasks, []string{"T1.1"}) {
		t.Errorf("unexpected hit: %+v", hits[0])
	}
}

func TestTraceRebuildBackfillsChanges(t *testing.T) {
	root := t.TempDir()
	setupTraceCLIFeature(t, root, "demo")
	if _, stderr, err := executeRoot(t, "archive", "demo", root, "--status", "completed"); err != nil {
		t.Fatalf("archive: %v\n%s", err, stderr)
	}

	cfg, err := config.Load(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	archiveDir, _ := cfg.ArchiveDir(root)
	var changesPath string
	if err := filepath.WalkDir(archiveDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "changes.json" {
			changesPath = path
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if changesPath == "" {
		t.Fatal("changes.json not created by archive")
	}
	if err := os.Remove(changesPath); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, err := executeRoot(t, "trace", "rebuild", "demo", root, "--json")
	if err != nil {
		t.Fatalf("trace rebuild: %v\n%s", err, stderr)
	}
	var results []rebuildResult
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Fatalf("unmarshal rebuild results: %v\n%s", err, stdout)
	}
	if len(results) != 1 || results[0].Status != "rebuilt" {
		t.Fatalf("unexpected rebuild results: %+v", results)
	}
	if _, err := os.Stat(changesPath); err != nil {
		t.Fatalf("changes.json not rebuilt: %v", err)
	}

	// The rebuilt file keeps attribution.
	changes, err := trace.ReadChanges(changesPath)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, file := range changes.Files {
		if file.Path == "src/export.go" && reflect.DeepEqual(file.Tasks, []string{"T1.1"}) {
			found = true
		}
	}
	if !found {
		t.Errorf("rebuilt changes lost attribution: %+v", changes.Files)
	}
}

func TestTraceFilesHumanOutput(t *testing.T) {
	root := t.TempDir()
	setupTraceCLIFeature(t, root, "demo")

	stdout, stderr, err := executeRoot(t, "trace", "demo", root, "--files")
	if err != nil {
		t.Fatalf("trace --files: %v\n%s", err, stderr)
	}
	for _, want := range []string{"Files:", "src/export.go", "[T1.1]", "AC: AC-001", "Unmapped", "src/other.go"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("output missing %q:\n%s", want, stdout)
		}
	}
}

func TestTraceFileNoHits(t *testing.T) {
	root := t.TempDir()
	setupTraceCLIFeature(t, root, "demo")

	stdout, stderr, err := executeRoot(t, "trace", "--file", "src/none.go", root)
	if err != nil {
		t.Fatalf("trace --file: %v\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "No spec or task references src/none.go") {
		t.Errorf("unexpected output: %s", stdout)
	}
}

func TestTraceProofProblemSuffixes(t *testing.T) {
	root := t.TempDir()
	setupTraceCLIFeature(t, root, "demo")

	tasksPath := filepath.Join(root, "specs", "active", "demo", "tasks.md")
	tasks := `# Tasks

- [x] T1.1 Missing file. Touches: src/export.go
  Proof: code src/missing.go Missing
- [x] T1.2 Missing anchor. Touches: src/export.go
  Proof: code src/export.go DoesNotExist
`
	if err := os.WriteFile(tasksPath, []byte(tasks), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, err := executeRoot(t, "trace", "demo", root)
	if err != nil {
		t.Fatalf("trace: %v\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "(file missing)") {
		t.Errorf("missing file suffix absent:\n%s", stdout)
	}
	if !strings.Contains(stdout, "(anchor missing)") {
		t.Errorf("missing anchor suffix absent:\n%s", stdout)
	}
}

func TestTraceParseArgsAndShortSHA(t *testing.T) {
	root := t.TempDir()
	if gotRoot, gotSlug := parseTraceArgs(nil); gotRoot != "." || gotSlug != "" {
		t.Errorf("parseTraceArgs(nil) = %q, %q", gotRoot, gotSlug)
	}
	if gotRoot, gotSlug := parseTraceArgs([]string{"demo"}); gotRoot != "." || gotSlug != "demo" {
		t.Errorf("parseTraceArgs(slug) = %q, %q", gotRoot, gotSlug)
	}
	if gotRoot, gotSlug := parseTraceArgs([]string{"demo", root}); gotRoot != root || gotSlug != "demo" {
		t.Errorf("parseTraceArgs(slug, root) = %q, %q", gotRoot, gotSlug)
	}
	if gotRoot, gotSlug := parseTraceArgs([]string{root}); gotRoot != root || gotSlug != "" {
		t.Errorf("parseTraceArgs(root) = %q, %q", gotRoot, gotSlug)
	}
	if shortSHA("") != "-" || shortSHA("abc") != "abc" || shortSHA("1234567890") != "1234567" {
		t.Errorf("shortSHA values = %q, %q, %q", shortSHA(""), shortSHA("abc"), shortSHA("1234567890"))
	}
}

func TestTraceRebuildSkipsSnapshotsWithoutPointer(t *testing.T) {
	root := t.TempDir()
	if _, stderr, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init: %v\n%s", err, stderr)
	}
	cfg, err := config.Load(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	archiveDir, _ := cfg.ArchiveDir(root)
	dir := filepath.Join(archiveDir, "demo", "2026-01-01")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "summary.md"), []byte("# summary\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, err := executeRoot(t, "trace", "rebuild", "demo", root, "--json")
	if err != nil {
		t.Fatalf("rebuild: %v\n%s", err, stderr)
	}
	var results []rebuildResult
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, stdout)
	}
	if len(results) != 1 || results[0].Status != "skipped: no commit pointer and no archived base" {
		t.Errorf("results = %+v", results)
	}
}

func TestArchiveWarnsOnDirtyTree(t *testing.T) {
	root := t.TempDir()
	setupTraceCLIFeature(t, root, "demo")

	// Modify a tracked file after the last commit.
	if err := os.WriteFile(filepath.Join(root, "src", "other.go"), []byte("package src\n\nfunc Other() {}\n\nfunc Extra() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, err := executeRoot(t, "archive", "demo", root, "--status", "completed")
	if err != nil {
		t.Fatalf("archive: %v\n%s", err, stderr)
	}
	if !strings.Contains(stdout, "uncommitted changes") {
		t.Errorf("missing dirty-tree warning:\n%s", stdout)
	}

	changesPath := findFile(t, root, "changes.json")
	changes, err := trace.ReadChanges(changesPath)
	if err != nil {
		t.Fatal(err)
	}
	if !changes.Dirty {
		t.Error("changes.json dirty flag not set")
	}
	for _, file := range changes.Files {
		if file.Path == "src/other.go" && file.Status == "M" {
			t.Errorf("uncommitted modification leaked into committed changes: %+v", file)
		}
	}
}

func TestTraceRebuildCompactSnapshot(t *testing.T) {
	root := t.TempDir()
	setupTraceCLIFeature(t, root, "demo")
	if _, stderr, err := executeRoot(t, "archive", "demo", root, "--compact"); err != nil {
		t.Fatalf("compact archive: %v\n%s", err, stderr)
	}

	changesPath := findFile(t, root, "changes.json")
	if err := os.Remove(changesPath); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, err := executeRoot(t, "trace", "rebuild", "demo", root, "--json")
	if err != nil {
		t.Fatalf("rebuild: %v\n%s", err, stderr)
	}
	var results []rebuildResult
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, stdout)
	}
	if len(results) != 1 || results[0].Status != "rebuilt" {
		t.Fatalf("results = %+v", results)
	}
	changes, err := trace.ReadChanges(changesPath)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, file := range changes.Files {
		if file.Path == "src/export.go" && reflect.DeepEqual(file.Tasks, []string{"T1.1"}) {
			found = true
		}
	}
	if !found {
		t.Errorf("compact rebuild lost attribution: %+v", changes.Files)
	}
}

func TestArchiveRestoreKeepsTraceMeta(t *testing.T) {
	root := t.TempDir()
	setupTraceCLIFeature(t, root, "demo")
	if _, stderr, err := executeRoot(t, "archive", "demo", root, "--status", "completed"); err != nil {
		t.Fatalf("archive: %v\n%s", err, stderr)
	}
	if _, stderr, err := executeRoot(t, "archive", "demo", root, "--restore"); err != nil {
		t.Fatalf("restore: %v\n%s", err, stderr)
	}

	cfg, err := config.Load(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	specsDir, _ := cfg.SpecsDir(root)
	if _, ok, err := trace.LoadMeta(specsDir, "demo"); err != nil || !ok {
		t.Fatalf("trace meta not restored: ok=%v err=%v", ok, err)
	}
	stdout, stderr, err := executeRoot(t, "trace", "demo", root, "--files", "--json")
	if err != nil {
		t.Fatalf("trace --files: %v\n%s", err, stderr)
	}
	changes := decodeChangeSet(t, stdout)
	if len(changes.Files) == 0 {
		t.Error("restored feature has no change attribution")
	}
}

func findFile(t *testing.T, root, name string) string {
	t.Helper()
	var found string
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == name {
			found = path
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if found == "" {
		t.Fatalf("%s not found under %s", name, root)
	}
	return found
}

func TestInferBaseFor(t *testing.T) {
	root := t.TempDir()
	if err := runGit(root, "init", "-q", "-b", "main"); err != nil {
		t.Fatal(err)
	}
	if err := runGit(root, "config", "user.email", "test@example.com"); err != nil {
		t.Fatal(err)
	}
	if err := runGit(root, "config", "user.name", "Test"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runGit(root, "add", "-A"); err != nil {
		t.Fatal(err)
	}
	if err := runGit(root, "-c", "user.email=test@example.com", "-c", "user.name=Test", "commit", "-q", "-m", "init"); err != nil {
		t.Fatal(err)
	}
	head, err := gitutil.Head(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if got := inferBaseFor(context.Background(), root, head); got != head {
		t.Errorf("inferBaseFor = %q, want %q", got, head)
	}
	if got := inferBaseFor(context.Background(), t.TempDir(), head); got != "" {
		t.Errorf("inferBaseFor(non-git) = %q, want empty", got)
	}
}

func TestTraceTestsFilter(t *testing.T) {
	root := t.TempDir()
	setupTraceCLIFeature(t, root, "demo")
	tasksPath := filepath.Join(root, "specs", "active", "demo", "tasks.md")
	tasks := `# Tasks

- [x] T1.1 Code. Touches: src/export.go
  Proof: code src/export.go RunExport
- [x] T1.2 Test. Touches: src/export_test.go
  Proof: test src/export_test.go TestRunExport
`
	if err := os.WriteFile(tasksPath, []byte(tasks), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, err := executeRoot(t, "trace", "demo", root, "--tests", "--json")
	if err != nil {
		t.Fatalf("trace --tests: %v\n%s", err, stderr)
	}
	var entries []trace.Entry
	if err := json.Unmarshal([]byte(stdout), &entries); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, stdout)
	}
	if len(entries) != 1 || entries[0].Kind != "test" || entries[0].TaskID != "T1.2" {
		t.Errorf("entries = %+v, want one test proof for T1.2", entries)
	}
}
