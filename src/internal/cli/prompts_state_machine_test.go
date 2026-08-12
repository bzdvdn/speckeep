package cli

import (
	"strings"
	"testing"

	"speckeep/src/internal/templates"
	"speckeep/src/internal/workflow"
)

// TestPromptsNextCommandsMatchStateMachine guards the contract between phase
// prompts ("Ready for: <command>"/"Return to: <command>") and the CLI state
// machine (nextCommand). If the CLI mapping changes, this test fails until the
// promoted final lines follow suit — prompts must never silently drift.
func TestPromptsNextCommandsMatchStateMachine(t *testing.T) {
	files, err := templates.Files(templates.LanguageSettings{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
		Shell:    "sh",
	})
	if err != nil {
		t.Fatalf("templates.Files() returned error: %v", err)
	}

	contents := make(map[string]string, len(files))
	for _, f := range files {
		contents[f.TargetPath] = f.Content
	}

	prompt := func(name string) string {
		target := "templates/prompts/" + name + ".md"
		content, ok := contents[target]
		if !ok {
			t.Fatalf("prompt file %s not present in embedded template set", target)
		}
		return content
	}

	// For each state the CLI can resolve as "ready for", list the prompts that
	// must emit that phase's command as their final line. `<slug>` is the exact
	// placeholder the prompts use, so nextCommand output matches prompt literals.
	cases := []struct {
		phase   string
		prompts []string
	}{
		{phase: "spec", prompts: []string{"inspect"}},              // blocked → "Return to: /spk.spec <slug>"
		{phase: "inspect", prompts: []string{"spec"}},              // optional deep review branch
		{phase: "plan", prompts: []string{"spec", "inspect"}},      // spec + inspect next command
		{phase: "tasks", prompts: []string{"plan"}},                // post-plan
		{phase: "implement", prompts: []string{"tasks", "verify"}}, // post-tasks + verify concerns path
		{phase: "verify", prompts: []string{"implement", "handoff", "hotfix"}},
		{phase: "archive", prompts: []string{"implement", "handoff", "hotfix", "verify"}},
	}

	for _, tc := range cases {
		want := nextCommand(workflow.FeatureState{ReadyFor: tc.phase, Slug: "<slug>"})
		if want == "" {
			t.Fatalf("nextCommand(/%s) mapped to an empty command — state machine has no %q phase", tc.phase, tc.phase)
		}
		for _, name := range tc.prompts {
			if !strings.Contains(prompt(name), want) {
				t.Errorf("prompt %s.md is missing the CLI-mapped command %q for phase %q — prompts and the CLI state machine drifted",
					name, want, tc.phase)
			}
		}
	}
}
