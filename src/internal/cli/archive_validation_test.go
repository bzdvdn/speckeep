package cli

import (
	"strings"
	"testing"
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
	if !strings.Contains(stderr, "no spec found for demo") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}
