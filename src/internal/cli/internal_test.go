package cli

import (
	"strings"
	"testing"
)

func TestInternalCheckConstitutionReadyAcceptsRootFlag(t *testing.T) {
	root := t.TempDir()
	initTestProject(t, root)

	stdout, stderr, err := executeRoot(t, "__internal", "check-constitution-ready", "--root", root)
	if err != nil {
		t.Fatalf("check-constitution-ready failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
	if strings.Contains(stderr, "unknown flag") || strings.Contains(stdout, "unknown flag") {
		t.Fatalf("unexpected unknown flag error: stdout=%s stderr=%s", stdout, stderr)
	}
}
