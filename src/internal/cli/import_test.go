package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportOpenSpecCommand(t *testing.T) {
	root := t.TempDir()
	initTestProject(t, root)

	changeDir := filepath.Join(root, "openspec", "changes", "add-dark-mode")
	if err := os.MkdirAll(filepath.Join(changeDir, "specs"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Why\n\nAdd theming.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile proposal: %v", err)
	}
	spec := `## ADDED Requirements

### Requirement: Theme selection
The app SHALL let users switch themes.

#### Scenario: User toggles dark mode
- **WHEN** the user clicks the toggle
- **THEN** the app switches to dark mode
`
	if err := os.WriteFile(filepath.Join(changeDir, "specs", "theme.md"), []byte(spec), 0o644); err != nil {
		t.Fatalf("WriteFile spec: %v", err)
	}
	if err := os.WriteFile(filepath.Join(changeDir, "design.md"), []byte("# Design\n"), 0o644); err != nil {
		t.Fatalf("WriteFile design: %v", err)
	}
	if err := os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("- [ ] T1 add toggle\n"), 0o644); err != nil {
		t.Fatalf("WriteFile tasks: %v", err)
	}

	stdout, stderr, err := executeRoot(t, "import", "openspec", root)
	if err != nil {
		t.Fatalf("import openspec should succeed, got: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "add-dark-mode") {
		t.Fatalf("expected imported slug in output, got: %s", stdout)
	}

	specDir := ensureSpecDir(t, root, "add-dark-mode")
	if _, err := os.Stat(filepath.Join(specDir, "spec.md")); err != nil {
		t.Fatalf("expected imported spec.md, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(specDir, "plan.md")); err != nil {
		t.Fatalf("expected imported plan.md, got %v", err)
	}
}

func TestImportRejectsUnknownSource(t *testing.T) {
	root := t.TempDir()
	initTestProject(t, root)
	_, stderr, err := executeRoot(t, "import", "jira", root)
	if err == nil {
		t.Fatalf("import jira should fail")
	}
	if !strings.Contains(stderr, "unsupported import source") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}
