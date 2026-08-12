package trace

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"speckeep/src/internal/project"
)

const tasksContent = `# Tasks

## Phase 1

- [x] T1.1 Implement export (AC-001). Touches: src/internal/export.go
      Proof: code src/internal/export.go RunExport
- [x] T1.2 Add test (AC-001). Touches: src/tests/export_test.go
      Proof: test src/tests/export_test.go TestRunExport
- [ ] T2.1 Open task without proof. Touches: src/internal/export.go
- [x] T3.1 Docs note
      Proof: docs docs/export.md
`

func TestParseTasks(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specDir := filepath.Join(root, "specs", "active", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	tasksPath := filepath.Join(specDir, "tasks.md")
	if err := os.WriteFile(tasksPath, []byte(tasksContent), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	entries, err := ParseTasks(context.Background(), root, "demo")
	if err != nil {
		t.Fatalf("ParseTasks returned error: %v", err)
	}

	// Only closed tasks carry Proof entries: T1.1, T1.2, T3.1.
	want := []Entry{
		{Slug: "demo", TaskID: "T1.1", ACID: "AC-001", Kind: "code", File: "src/internal/export.go", Anchor: "RunExport"},
		{Slug: "demo", TaskID: "T1.2", ACID: "AC-001", Kind: "test", File: "src/tests/export_test.go", Anchor: "TestRunExport"},
		{Slug: "demo", TaskID: "T3.1", ACID: "", Kind: "docs", File: "docs/export.md"},
	}
	if !reflect.DeepEqual(entries, want) {
		t.Errorf("ParseTasks() = %+v, want %+v", entries, want)
	}
}

func TestCompletedTaskLines(t *testing.T) {
	ids, withProof := CompletedTaskLines(tasksContent)

	if len(ids) != 3 {
		t.Fatalf("CompletedTaskLines() ids = %v, want 3", ids)
	}
	if !withProof["T1.1"] || !withProof["T1.2"] || !withProof["T3.1"] {
		t.Errorf("CompletedTaskLines() withProof = %v", withProof)
	}
	for _, id := range []string{"T1.1", "T1.2", "T3.1"} {
		found := false
		for _, got := range ids {
			if got == id {
				found = true
			}
		}
		if !found {
			t.Errorf("CompletedTaskLines() missing id %s", id)
		}
	}
}

func TestCheckEntry(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src", "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "internal", "export.go"), []byte("package export\n\nfunc RunExport() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ok := Entry{TaskID: "T1.1", Kind: "code", File: "src/internal/export.go", Anchor: "RunExport"}
	if problems := CheckEntry(root, ok); len(problems) != 0 {
		t.Errorf("CheckEntry(ok) = %v, want none", problems)
	}

	missingAnchor := Entry{TaskID: "T1.1", Kind: "code", File: "src/internal/export.go", Anchor: "DoesNotExist"}
	problems := CheckEntry(root, missingAnchor)
	if len(problems) != 1 || problems[0].Kind != "anchor-missing" {
		t.Errorf("CheckEntry(missing anchor) = %v, want anchor-missing", problems)
	}

	missingFile := Entry{TaskID: "T1.1", Kind: "code", File: "src/internal/gone.go", Anchor: "RunExport"}
	problems = CheckEntry(root, missingFile)
	if len(problems) != 1 || problems[0].Kind != "file-missing" {
		t.Errorf("CheckEntry(missing file) = %v, want file-missing", problems)
	}
}

func TestFindLegacyMarkers(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "service.go"), []byte("package service\n\n// @sk-task T1.1: Legacy\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	markers, err := FindLegacyMarkers(context.Background(), root)
	if err != nil {
		t.Fatalf("FindLegacyMarkers returned error: %v", err)
	}
	if len(markers) != 1 {
		t.Errorf("FindLegacyMarkers() len = %d, want 1", len(markers))
	}
}

func TestShouldSkip(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"node_modules", true},
		{".git", true},
		{".speckeep", true},
		{"src", false},
		{"vendor", true},
		{".hidden", true},
	}

	for _, tt := range tests {
		if got := shouldSkip(tt.path, nil); got != tt.want {
			t.Errorf("shouldSkip(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
