package featurepaths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSpecReturnsCanonicalPath(t *testing.T) {
	got := Spec("specs", "my-feature")
	want := filepath.Join("specs", "my-feature", "spec.md")
	if got != want {
		t.Errorf("Spec() = %q, want %q", got, want)
	}
}

func TestInspectReturnsCanonicalPath(t *testing.T) {
	got := Inspect("specs", "my-feature")
	want := filepath.Join("specs", "my-feature", "inspect.md")
	if got != want {
		t.Errorf("Inspect() = %q, want %q", got, want)
	}
}

func TestSummaryReturnsCanonicalPath(t *testing.T) {
	got := Summary("specs", "my-feature")
	want := filepath.Join("specs", "my-feature", "summary.md")
	if got != want {
		t.Errorf("Summary() = %q, want %q", got, want)
	}
}

func TestPlanDirReturnsCorrectPath(t *testing.T) {
	got := PlanDir("specs", "my-feature")
	want := filepath.Join("specs", "my-feature", "plan")
	if got != want {
		t.Errorf("PlanDir() = %q, want %q", got, want)
	}
}

func TestTasksReturnsCorrectPath(t *testing.T) {
	got := Tasks("specs", "my-feature")
	want := filepath.Join("specs", "my-feature", "plan", "tasks.md")
	if got != want {
		t.Errorf("Tasks() = %q, want %q", got, want)
	}
}

func TestLegacySpecReturnsCorrectPath(t *testing.T) {
	got := LegacySpec("specs", "my-feature")
	want := filepath.Join("specs", "my-feature.md")
	if got != want {
		t.Errorf("LegacySpec() = %q, want %q", got, want)
	}
}

func TestResolveSpecPrefersCanonical(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	if err := os.MkdirAll(filepath.Join(specsDir, "demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Create both canonical and legacy
	if err := os.WriteFile(Spec(specsDir, "demo"), []byte("# spec"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(LegacySpec(specsDir, "demo"), []byte("# legacy spec"), 0o644); err != nil {
		t.Fatal(err)
	}

	path, isLegacy := ResolveSpec(specsDir, "demo")
	if path != Spec(specsDir, "demo") {
		t.Errorf("expected canonical path, got %q", path)
	}
	if isLegacy {
		t.Error("expected isLegacy=false when canonical exists")
	}
}

func TestResolveSpecFallsBackToLegacy(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Only legacy exists
	if err := os.WriteFile(LegacySpec(specsDir, "demo"), []byte("# legacy spec"), 0o644); err != nil {
		t.Fatal(err)
	}

	path, isLegacy := ResolveSpec(specsDir, "demo")
	if path != LegacySpec(specsDir, "demo") {
		t.Errorf("expected legacy path, got %q", path)
	}
	if !isLegacy {
		t.Error("expected isLegacy=true when only legacy exists")
	}
}

func TestResolveSpecReturnsCanonicalWhenNeitherExists(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")

	path, isLegacy := ResolveSpec(specsDir, "missing")
	if path != Spec(specsDir, "missing") {
		t.Errorf("expected canonical path even when missing, got %q", path)
	}
	if isLegacy {
		t.Error("expected isLegacy=false when neither file exists")
	}
}

func TestResolveSummaryPrefersCanonical(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	if err := os.MkdirAll(filepath.Join(specsDir, "demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Summary(specsDir, "demo"), []byte("# summary"), 0o644); err != nil {
		t.Fatal(err)
	}

	path, isLegacy := ResolveSummary(specsDir, "demo")
	if path != Summary(specsDir, "demo") {
		t.Errorf("expected canonical summary path, got %q", path)
	}
	if isLegacy {
		t.Error("expected isLegacy=false")
	}
}

func TestListSpecSlugsFindsCanonicalDirs(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")

	for _, slug := range []string{"alpha", "beta", "gamma"} {
		if err := os.MkdirAll(filepath.Join(specsDir, slug), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(specsDir, slug, "spec.md"), []byte("# spec"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	slugs, err := ListSpecSlugs(specsDir)
	if err != nil {
		t.Fatalf("ListSpecSlugs: %v", err)
	}
	if len(slugs) != 3 {
		t.Fatalf("expected 3 slugs, got %d: %v", len(slugs), slugs)
	}
	if slugs[0] != "alpha" || slugs[1] != "beta" || slugs[2] != "gamma" {
		t.Errorf("unexpected slug order: %v", slugs)
	}
}

func TestListSpecSlugsFindsLegacyFlatFiles(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "my-feature.md"), []byte("# spec"), 0o644); err != nil {
		t.Fatal(err)
	}

	slugs, err := ListSpecSlugs(specsDir)
	if err != nil {
		t.Fatalf("ListSpecSlugs: %v", err)
	}
	if len(slugs) != 1 || slugs[0] != "my-feature" {
		t.Errorf("expected [my-feature], got %v", slugs)
	}
}

func TestListSpecSlugsExcludesNonSpecMarkdown(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// These should NOT be treated as spec slugs
	for _, name := range []string{
		"feature.inspect.md",
		"feature.summary.md",
		"feature.hotfix.md",
	} {
		if err := os.WriteFile(filepath.Join(specsDir, name), []byte("# doc"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// This is the only real spec
	if err := os.WriteFile(filepath.Join(specsDir, "real-feature.md"), []byte("# spec"), 0o644); err != nil {
		t.Fatal(err)
	}

	slugs, err := ListSpecSlugs(specsDir)
	if err != nil {
		t.Fatalf("ListSpecSlugs: %v", err)
	}
	if len(slugs) != 1 || slugs[0] != "real-feature" {
		t.Errorf("expected only [real-feature], got %v", slugs)
	}
}

func TestListSpecSlugsIgnoresDirWithoutSpec(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	// Directory exists but has no spec.md
	if err := os.MkdirAll(filepath.Join(specsDir, "empty-feature"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Another dir with spec.md
	if err := os.MkdirAll(filepath.Join(specsDir, "real-feature"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "real-feature", "spec.md"), []byte("# spec"), 0o644); err != nil {
		t.Fatal(err)
	}

	slugs, err := ListSpecSlugs(specsDir)
	if err != nil {
		t.Fatalf("ListSpecSlugs: %v", err)
	}
	if len(slugs) != 1 || slugs[0] != "real-feature" {
		t.Errorf("expected only [real-feature], got %v", slugs)
	}
}

func TestListSpecSlugsDeduplicatesCanonicalAndLegacy(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	if err := os.MkdirAll(filepath.Join(specsDir, "demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Both canonical and legacy exist for same slug
	if err := os.WriteFile(filepath.Join(specsDir, "demo", "spec.md"), []byte("# spec"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "demo.md"), []byte("# spec legacy"), 0o644); err != nil {
		t.Fatal(err)
	}

	slugs, err := ListSpecSlugs(specsDir)
	if err != nil {
		t.Fatalf("ListSpecSlugs: %v", err)
	}
	if len(slugs) != 1 || slugs[0] != "demo" {
		t.Errorf("expected deduplication to one slug, got %v", slugs)
	}
}

func TestListSpecSlugsReturnsEmptyForMissingDir(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "nonexistent", "specs")

	_, err := ListSpecSlugs(specsDir)
	if err == nil {
		t.Error("expected error for missing directory, got nil")
	}
}

func TestListSpecSlugsWorksWithNestedActiveRoot(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs", "active")
	if err := os.MkdirAll(filepath.Join(specsDir, "demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "demo", "spec.md"), []byte("# spec"), 0o644); err != nil {
		t.Fatal(err)
	}

	slugs, err := ListSpecSlugs(specsDir)
	if err != nil {
		t.Fatalf("ListSpecSlugs: %v", err)
	}
	if len(slugs) != 1 || slugs[0] != "demo" {
		t.Fatalf("expected [demo], got %v", slugs)
	}
}

func TestArtifactsReturnsAllExpectedNames(t *testing.T) {
	artifacts := Artifacts("specs", "demo")
	names := make(map[string]struct{}, len(artifacts))
	for _, a := range artifacts {
		names[a.Name] = struct{}{}
	}
	for _, want := range []string{"spec", "inspect report", "summary", "hotfix"} {
		if _, ok := names[want]; !ok {
			t.Errorf("Artifacts() missing %q, got %v", want, artifacts)
		}
	}
}
