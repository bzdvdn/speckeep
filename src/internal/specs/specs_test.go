package specs

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"speckeep/src/internal/project"
)

func TestListReturnsSortedMarkdownSpecsOnly(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specsDir := filepath.Join(root, "specs")
	if err := os.MkdirAll(filepath.Join(specsDir, "zeta"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "zeta", "spec.md"), []byte("# Zeta"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(specsDir, "alpha"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "alpha", "spec.md"), []byte("# Alpha"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "notes.txt"), []byte("ignore"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(specsDir, "nested"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	got, err := List(root)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	want := []string{"alpha", "zeta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("List = %#v, want %#v", got, want)
	}
}

func TestShowReturnsSpecContent(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specDir := filepath.Join(root, "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	specPath := filepath.Join(specDir, "spec.md")
	content := "# Demo\n\nHello"
	if err := os.WriteFile(specPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	got, err := Show(root, "demo")
	if err != nil {
		t.Fatalf("Show returned error: %v", err)
	}
	if got != content {
		t.Fatalf("Show = %q, want %q", got, content)
	}
}

func TestCreateGeneratesSpecAndTasksFromTemplates(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	result, err := Create(root, "Partner Scheduling", CreateOptions{CreateBranch: false})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if len(result.Messages) == 0 || result.Messages[0] != "skipped feature branch creation" {
		t.Fatalf("unexpected messages: %v", result.Messages)
	}

	specPath := filepath.Join(root, "specs", "partner-scheduling", "spec.md")
	tasksPath := filepath.Join(root, "specs", "partner-scheduling", "plan", "tasks.md")

	specContent, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("ReadFile spec returned error: %v", err)
	}
	tasksContent, err := os.ReadFile(tasksPath)
	if err != nil {
		t.Fatalf("ReadFile tasks returned error: %v", err)
	}

	if !strings.Contains(string(specContent), "Partner Scheduling") {
		t.Fatalf("expected spec to contain filled title, got: %s", string(specContent))
	}
	if !strings.Contains(string(tasksContent), "Partner Scheduling") {
		t.Fatalf("expected tasks to contain filled title, got: %s", string(tasksContent))
	}
}

func TestCreateFailsOnEmptySlug(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	_, err = Create(root, "---", CreateOptions{CreateBranch: false})
	if err == nil {
		t.Fatal("expected error for empty slug, got nil")
	}
}

func TestResolveInputFromPlainText(t *testing.T) {
	resolved, err := ResolveInput("Add dark mode")
	if err != nil {
		t.Fatalf("ResolveInput returned error: %v", err)
	}
	if got, want := resolved.Title, "Add Dark Mode"; got != want {
		t.Fatalf("Title = %q, want %q", got, want)
	}
	if got, want := resolved.Slug, "add-dark-mode"; got != want {
		t.Fatalf("Slug = %q, want %q", got, want)
	}
}

func TestResolveInputFromFileMetadataName(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "spec_prompt.md")
	content := "name: Add dark mode\n\nUse the new theme tokens.\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	resolved, err := ResolveInput(path)
	if err != nil {
		t.Fatalf("ResolveInput returned error: %v", err)
	}
	if got, want := resolved.Title, "Add dark mode"; got != want {
		t.Fatalf("Title = %q, want %q", got, want)
	}
	if got, want := resolved.Slug, "add-dark-mode"; got != want {
		t.Fatalf("Slug = %q, want %q", got, want)
	}
}

func TestResolveInputFromFileMetadataSlug(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "spec_prompt.md")
	content := "name: Add dark mode\nslug: ui-dark-mode\n\nUse the new theme tokens.\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	resolved, err := ResolveInput(path)
	if err != nil {
		t.Fatalf("ResolveInput returned error: %v", err)
	}
	if got, want := resolved.Title, "Add dark mode"; got != want {
		t.Fatalf("Title = %q, want %q", got, want)
	}
	if got, want := resolved.Slug, "ui-dark-mode"; got != want {
		t.Fatalf("Slug = %q, want %q", got, want)
	}
}

func TestResolveInputRejectsGenericPromptFileWithoutMetadata(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "spec_prompt.md")
	if err := os.WriteFile(path, []byte("Add dark mode.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	_, err := ResolveInput(path)
	if err == nil {
		t.Fatal("expected error for generic prompt file name, got nil")
	}
	if !strings.Contains(err.Error(), "needs a top-level name: or slug:") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveInputRejectsURLInput(t *testing.T) {
	_, err := ResolveInput("https://example.com/spec_prompt.md")
	if err == nil {
		t.Fatal("expected error for URL input, got nil")
	}
	if !strings.Contains(err.Error(), "looks like a URL") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateCreatesAndSwitchesFeatureBranch(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: true, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	result, err := Create(root, "Partner Scheduling", CreateOptions{CreateBranch: true})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if len(result.Messages) == 0 || !strings.Contains(result.Messages[0], "feature/partner-scheduling") {
		t.Fatalf("unexpected messages: %v", result.Messages)
	}

	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse returned error: %v", err)
	}
	if got, want := strings.TrimSpace(string(output)), "feature/partner-scheduling"; got != want {
		t.Fatalf("HEAD = %q, want %q", got, want)
	}
}
