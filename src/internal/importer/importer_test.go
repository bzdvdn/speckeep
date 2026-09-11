package importer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportOpenSpecWorkspace(t *testing.T) {
	root := t.TempDir()
	specsDir := filepath.Join(root, "specs", "active")
	writeOpenSpecChange(t, filepath.Join(root, "openspec", "changes", "add-dark-mode"), "add-dark-mode")

	result, err := Import(context.Background(), SourceOpenSpec, root, specsDir)
	if err != nil {
		t.Fatalf("Import returned error: %v", err)
	}
	if len(result.Imported) != 1 {
		t.Fatalf("expected 1 imported feature, got %d (%+v)", len(result.Imported), result.Imported)
	}
	feat := result.Imported[0]
	if feat.Slug != "add-dark-mode" {
		t.Fatalf("unexpected slug %q", feat.Slug)
	}

	specContent, err := os.ReadFile(filepath.Join(specsDir, "add-dark-mode", "spec.md"))
	if err != nil {
		t.Fatalf("read spec.md: %v", err)
	}
	spec := string(specContent)
	for _, want := range []string{"RQ-001", "AC-001", "**Given**", "**When**", "**Then**", "## Assumptions", "## Context", "clicks the theme toggle"} {
		if !strings.Contains(spec, want) {
			t.Fatalf("imported spec.md missing %q:\n%s", want, spec)
		}
	}

	plan, err := os.ReadFile(filepath.Join(specsDir, "add-dark-mode", "plan.md"))
	if err != nil {
		t.Fatalf("read plan.md: %v", err)
	}
	if !strings.Contains(string(plan), "## Constitution Compliance") {
		t.Fatalf("imported plan.md missing Constitution Compliance")
	}

	tasks, err := os.ReadFile(filepath.Join(specsDir, "add-dark-mode", "tasks.md"))
	if err != nil {
		t.Fatalf("read tasks.md: %v", err)
	}
	if !strings.Contains(string(tasks), "IMPORTED from") {
		t.Fatalf("imported tasks.md missing import note")
	}
}

func TestImportOpenSpecSkipsArchiveAndExisting(t *testing.T) {
	root := t.TempDir()
	specsDir := filepath.Join(root, "specs", "active")
	writeOpenSpecChange(t, filepath.Join(root, "openspec", "changes", "alpha"), "alpha")
	writeOpenSpecChange(t, filepath.Join(root, "openspec", "changes", "archive", "old"), "old")

	// Pre-create an existing feature for "beta" that also exists in changes.
	writeOpenSpecChange(t, filepath.Join(root, "openspec", "changes", "beta"), "beta")
	if err := os.MkdirAll(filepath.Join(specsDir, "beta"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "beta", "spec.md"), []byte("# existing\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	result, err := Import(context.Background(), SourceOpenSpec, root, specsDir)
	if err != nil {
		t.Fatalf("Import returned error: %v", err)
	}
	if len(result.Imported) != 1 {
		t.Fatalf("expected only 'alpha' to import, got %+v", result.Imported)
	}
	if !strings.Contains(strings.Join(result.Skipped, ","), "beta") {
		t.Fatalf("expected beta to be skipped (exists), skipped=%v", result.Skipped)
	}
}

func TestImportSpecKitWorkspace(t *testing.T) {
	root := t.TempDir()
	specsDir := filepath.Join(root, "specs", "active")
	writeSpecKitPackage(t, filepath.Join(root, "specs", "login-flow"))

	result, err := Import(context.Background(), SourceSpecKit, root, specsDir)
	if err != nil {
		t.Fatalf("Import returned error: %v", err)
	}
	if len(result.Imported) != 1 || result.Imported[0].Slug != "login-flow" {
		t.Fatalf("unexpected import result: %+v", result.Imported)
	}
	spec, err := os.ReadFile(filepath.Join(specsDir, "login-flow", "spec.md"))
	if err != nil {
		t.Fatalf("read spec.md: %v", err)
	}
	if !strings.Contains(string(spec), "imported from Spec Kit") {
		t.Fatalf("unexpected spec content: %s", spec)
	}
	if _, err := os.Stat(filepath.Join(specsDir, "login-flow", "plan.md")); err != nil {
		t.Fatalf("expected plan.md, got %v", err)
	}
}

func TestParseSource(t *testing.T) {
	if _, err := ParseSource("openspec"); err != nil {
		t.Fatalf("openspec should parse, got %v", err)
	}
	if _, err := ParseSource("SPECKit"); err != nil {
		t.Fatalf("specKit (case-insensitive) should parse, got %v", err)
	}
	if _, err := ParseSource("jira"); err == nil {
		t.Fatalf("jira should not parse")
	}
}

func writeOpenSpecChange(t *testing.T, changeDir, slug string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(changeDir, "specs"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	proposal := "---\nslug: " + slug + "\n---\n# Why\n\nWe want polished UX.\n\n# What Changes\n\nAdd theming support.\n"
	if err := os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte(proposal), 0o644); err != nil {
		t.Fatalf("WriteFile proposal: %v", err)
	}
	spec := `## ADDED Requirements

### Requirement: Theme selection
The app SHALL let users switch between light and dark themes, defaulting to the system preference.

#### Scenario: User toggles dark mode
- **WHEN** the user clicks the theme toggle
- **THEN** the app switches to dark mode and persists the choice

#### Scenario: Empty state handled
- **GIVEN** no saved preference exists
- **WHEN** the app first loads
- **THEN** it shows the system default theme
`
	if err := os.WriteFile(filepath.Join(changeDir, "specs", "theme.md"), []byte(spec), 0o644); err != nil {
		t.Fatalf("WriteFile spec: %v", err)
	}
	design := "# Design\n\nUse CSS variables and a ThemeContext.\n"
	if err := os.WriteFile(filepath.Join(changeDir, "design.md"), []byte(design), 0o644); err != nil {
		t.Fatalf("WriteFile design: %v", err)
	}
	tasks := "- [ ] T1 add theme context\n- [x] T2 wire toggle\n"
	if err := os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte(tasks), 0o644); err != nil {
		t.Fatalf("WriteFile tasks: %v", err)
	}
}

func writeSpecKitPackage(t *testing.T, pkgDir string) {
	t.Helper()
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	spec := "# Login Flow\n\n## Goal\nAllow users to authenticate.\n"
	if err := os.WriteFile(filepath.Join(pkgDir, "spec.md"), []byte(spec), 0o644); err != nil {
		t.Fatalf("WriteFile spec: %v", err)
	}
	plan := "# Plan\n\nUse OAuth2 PKCE.\n"
	if err := os.WriteFile(filepath.Join(pkgDir, "plan.md"), []byte(plan), 0o644); err != nil {
		t.Fatalf("WriteFile plan: %v", err)
	}
	tasks := "- [ ] T1 add PKCE flow\n"
	if err := os.WriteFile(filepath.Join(pkgDir, "tasks.md"), []byte(tasks), 0o644); err != nil {
		t.Fatalf("WriteFile tasks: %v", err)
	}
}
