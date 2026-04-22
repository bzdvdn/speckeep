package agents

type CommandDefinition struct {
	Name        string
	Description string
	PromptPath  string
	Extras      []string
	Optional    bool
	Category    string
}

func DefaultCommands(shell string) []CommandDefinition {
	normalizedShell := normalizeShell(shell)

	return []CommandDefinition{
		{
			Name:        "constitution",
			Description: "Create or update the project constitution",
			PromptPath:  ".speckeep/templates/prompts/constitution.md",
			Extras:      []string{scriptPath("check-constitution", normalizedShell)},
			Category:    "workflow",
		},
		{
			Name:        "spec",
			Description: "Create or update one feature spec",
			PromptPath:  ".speckeep/templates/prompts/spec.md",
			Extras:      []string{scriptPath("check-spec-ready", normalizedShell)},
			Category:    "workflow",
		},
		{
			Name:        "inspect",
			Description: "Inspect one feature for consistency and quality",
			PromptPath:  ".speckeep/templates/prompts/inspect.md",
			Extras: []string{
				scriptPath("check-inspect-ready", normalizedShell),
				scriptPath("inspect-spec", normalizedShell),
			},
			Category: "workflow",
		},
		{
			Name:        "plan",
			Description: "Create or update one feature plan package",
			PromptPath:  ".speckeep/templates/prompts/plan.md",
			Extras:      []string{scriptPath("check-plan-ready", normalizedShell)},
			Category:    "workflow",
		},
		{
			Name:        "tasks",
			Description: "Create or update tasks for one feature",
			PromptPath:  ".speckeep/templates/prompts/tasks.md",
			Extras:      []string{scriptPath("check-tasks-ready", normalizedShell)},
			Category:    "workflow",
		},
		{
			Name:        "implement",
			Description: "Implement one feature from tasks",
			PromptPath:  ".speckeep/templates/prompts/implement.md",
			Extras: []string{
				scriptPath("check-implement-ready", normalizedShell),
				scriptPath("list-open-tasks", normalizedShell),
			},
			Category: "workflow",
		},
		{
			Name:        "verify",
			Description: "Verify one implemented feature package",
			PromptPath:  ".speckeep/templates/prompts/verify.md",
			Extras: []string{
				scriptPath("check-verify-ready", normalizedShell),
				scriptPath("verify-task-state", normalizedShell),
			},
			Category: "workflow",
		},
		{
			Name:        "archive",
			Description: "Archive one feature package",
			PromptPath:  ".speckeep/templates/prompts/archive.md",
			Extras: []string{
				scriptPath("check-archive-ready", normalizedShell),
				scriptPath("archive-feature", normalizedShell),
			},
			Category: "workflow",
		},
		{
			Name:        "handoff",
			Description: "Generate a session handoff document for one feature",
			PromptPath:  ".speckeep/templates/prompts/handoff.md",
			Extras:      []string{scriptPath("list-open-tasks", normalizedShell)},
			Category:    "workflow",
		},
		{
			Name:        "challenge",
			Description: "Adversarial review of a feature spec or plan",
			PromptPath:  ".speckeep/templates/prompts/challenge.md",
			Extras:      nil,
			Optional:    true,
			Category:    "workflow",
		},
		{
			Name:        "scope",
			Description: "Quick scope boundary check for a feature",
			PromptPath:  ".speckeep/templates/prompts/scope.md",
			Extras:      nil,
			Optional:    true,
			Category:    "workflow",
		},
		{
			Name:        "recap",
			Description: "Project-level overview of all active features and their current phase",
			PromptPath:  ".speckeep/templates/prompts/recap.md",
			Extras:      []string{scriptPath("list-specs", normalizedShell)},
			Optional:    true,
			Category:    "workflow",
		},
		{
			Name:        "hotfix",
			Description: "Create emergency fix outside the standard phase chain",
			PromptPath:  ".speckeep/templates/prompts/hotfix.md",
			Extras:      nil,
			Optional:    true,
			Category:    "workflow",
		},
	}
}
