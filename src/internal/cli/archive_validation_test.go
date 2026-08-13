package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"speckeep/src/internal/config"
)

func TestArchiveCommandRequiresReasonForNonCompletedStatuses(t *testing.T) {
	root := t.TempDir()

	_, stderr, err := executeRoot(t, "archive", "demo", root, "--status", "deferred")
	if err == nil {
		t.Fatalf("expected archive command to return an error")
	}
	if !strings.Contains(stderr, "archive reason is required for non-completed statuses") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestArchiveCommandRejectsInvalidStatus(t *testing.T) {
	root := t.TempDir()

	_, stderr, err := executeRoot(t, "archive", "demo", root, "--status", "bogus")
	if err == nil {
		t.Fatalf("expected archive command to return an error")
	}
	if !strings.Contains(stderr, "invalid archive status: bogus") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestArchiveCommandAllowsCompletedWithoutReason(t *testing.T) {
	root := t.TempDir()

	_, stderr, err := executeRoot(t, "archive", "demo", root, "--status", "completed")
	if err == nil {
		t.Fatalf("expected archive command to return an error")
	}
	if strings.Contains(stderr, "archive reason is required") {
		t.Fatalf("did not expect a reason validation error, got: %s", stderr)
	}
	if !strings.Contains(stderr, "no spec or hotfix found for demo") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestArchiveAllowsWithoutVerifyWhenOptional(t *testing.T) {
	root := t.TempDir()
	initTestProject(t, root)
	writeMinimalFeature(t, root, "demo")

	_, stderr, err := executeRoot(t, "archive", "demo", root, "--status", "completed")
	if err != nil {
		t.Fatalf("expected archive to succeed without verify.md, got: %v\nstderr: %s", err, stderr)
	}

	cfg, err := config.Load(context.Background(), root)
	if err != nil {
		t.Fatalf("config.Load returned error: %v", err)
	}
	archiveDir, err := cfg.ArchiveDir(root)
	if err != nil {
		t.Fatalf("cfg.ArchiveDir returned error: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(archiveDir, "demo"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected one archive snapshot for demo, got entries=%v err=%v", entries, err)
	}
}

func TestArchiveRequiresVerifyWhenConfiguredRequired(t *testing.T) {
	root := t.TempDir()
	initTestProject(t, root)

	cfg, err := config.Load(context.Background(), root)
	if err != nil {
		t.Fatalf("config.Load returned error: %v", err)
	}
	cfg.Workflow.Verify = "required"
	if err := config.Save(context.Background(), root, cfg); err != nil {
		t.Fatalf("config.Save returned error: %v", err)
	}

	writeMinimalFeature(t, root, "demo")

	_, stderr, err := executeRoot(t, "archive", "demo", root, "--status", "completed")
	if err == nil {
		t.Fatalf("expected archive to require verify.md, but it succeeded")
	}
	if !strings.Contains(stderr, "verify.md not found - run verify before archiving") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestArchiveBlocksOnNonPassVerifyReport(t *testing.T) {
	root := t.TempDir()
	initTestProject(t, root)
	specDir := writeMinimalFeature(t, root, "demo")

	verifyContent := "---\nreport_type: verify\nslug: demo\nstatus: concerns\ngenerated_at: 2026-08-13\n---\n# Verify Report: demo\n"
	if err := os.WriteFile(filepath.Join(specDir, "verify.md"), []byte(verifyContent), 0o644); err != nil {
		t.Fatalf("WriteFile(verify) returned error: %v", err)
	}

	_, stderr, err := executeRoot(t, "archive", "demo", root, "--status", "completed")
	if err == nil {
		t.Fatalf("expected archive to block on a non-pass verify report")
	}
	if !strings.Contains(stderr, "verify status is concerns - fix before archiving") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func initTestProject(t *testing.T, root string) {
	t.Helper()
	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init returned error: %v", err)
	}
}

func writeMinimalFeature(t *testing.T, root, slug string) string {
	t.Helper()
	specDir := ensureSpecDir(t, root, slug)
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# "+slug+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}
	return specDir
}
