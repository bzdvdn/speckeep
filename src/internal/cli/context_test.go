package cli

import (
	"encoding/json"
	"testing"
)

func TestContextJSONOutput(t *testing.T) {
	root := t.TempDir()
	setupTraceCLIFeature(t, root, "demo")

	stdout, stderr, err := executeRoot(t, "context", "demo", root, "--json")
	if err != nil {
		t.Fatalf("context --json: %v\n%s", err, stderr)
	}

	var payload contextReport
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("decode context json: %v\n%s", err, stdout)
	}
	if payload.Slug != "demo" {
		t.Fatalf("expected slug demo, got %q", payload.Slug)
	}
	if payload.Phase == "" {
		t.Fatal("expected a phase in context output")
	}
	if len(payload.Phases) == 0 {
		t.Fatal("expected per-artifact phases in context output")
	}
	if payload.TotalTokens <= 0 {
		t.Fatalf("expected positive total tokens, got %d", payload.TotalTokens)
	}

	var specFound bool
	for _, phase := range payload.Phases {
		if phase.Name == "Spec" && phase.Exists && phase.EstimatedTokens > 0 {
			specFound = true
		}
	}
	if !specFound {
		t.Fatalf("expected the spec artifact to be counted, got %+v", payload.Phases)
	}
}
