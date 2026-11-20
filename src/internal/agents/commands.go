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
			Extras:      []string{scriptPath("check-ready", normalizedShell) + " constitution"},
			Category:    "workflow",
		},
		{
			Name:        "spec",
			Description: "Create or update one feature spec",
			PromptPath:  ".speckeep/templates/prompts/spec.md",
			Extras:      []string{scriptPath("check-ready", normalizedShell) + " spec"},
			Category:    "workflow",
		},
		{
			Name:        "propose",
			Description: "Recommended default for a new feature: idea → spec + tasks (plan optional), straight to implement",
			PromptPath:  ".speckeep/templates/prompts/propose.md",
			Extras:      []string{scriptPath("check-ready", normalizedShell) + " propose"},
			Category:    "workflow",
		},
		{
			Name:        "inspect",
			Description: "Inspect one feature for consistency and quality",
			PromptPath:  ".speckeep/templates/prompts/inspect.md",
			Extras: []string{
				scriptPath("check-ready", normalizedShell) + " inspect",
				scriptPath("inspect-spec", normalizedShell),
			},
			Category: "workflow",
		},
		{
			Name:        "plan",
			Description: "Optional design artifacts for one feature (design-heavy work only)",
			PromptPath:  ".speckeep/templates/prompts/plan.md",
			Extras:      []string{scriptPath("check-ready", normalizedShell) + " plan"},
			Category:    "workflow",
		},
		{
			Name:        "research",
			Description: "Optional technical research: investigate a question and write a cited research.md",
			PromptPath:  ".speckeep/templates/prompts/research.md",
			Extras:      nil,
			Optional:    true,
			Category:    "workflow",
		},
		{
			Name:        "tasks",
			Description: "Create or update tasks for one feature",
			PromptPath:  ".speckeep/templates/prompts/tasks.md",
			Extras:      []string{scriptPath("check-ready", normalizedShell) + " tasks"},
			Category:    "workflow",
		},
		{
			Name:        "implement",
			Description: "Implement one feature from tasks",
			PromptPath:  ".speckeep/templates/prompts/implement.md",
			Extras: []string{
				scriptPath("check-ready", normalizedShell) + " implement",
				scriptPath("list-open-tasks", normalizedShell),
			},
			Category: "workflow",
		},
		{
			Name:        "verify",
			Description: "Verify one implemented feature package",
			PromptPath:  ".speckeep/templates/prompts/verify.md",
			Extras: []string{
				scriptPath("check-ready", normalizedShell) + " verify",
				scriptPath("verify-task-state", normalizedShell),
			},
			Category: "workflow",
		},
		{
			Name:        "converge",
			Description: "Close a feature fast: re-check tasks/proofs, append follow-up tasks, repeat until converged",
			PromptPath:  ".speckeep/templates/prompts/converge.md",
			Extras:      []string{scriptPath("check-ready", normalizedShell) + " converge"},
			Optional:    true,
			Category:    "workflow",
		},
		{
			Name:        "handoff",
			Description: "Generate a session handoff document for one feature",
			PromptPath:  ".speckeep/templates/prompts/handoff.md",
			Extras:      []string{scriptPath("list-open-tasks", normalizedShell)},
			Category:    "workflow",
		},
		{
			Name:        "resume",
			Description: "Minimal in-session resume: state packet + next task, then hand over to the phase skill",
			PromptPath:  ".speckeep/templates/prompts/resume.md",
			Extras:      []string{scriptPath("list-open-tasks", normalizedShell)},
			Optional:    true,
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
			Name:        "grill",
			Description: "Optional pre-spec interview: resolve the design tree before writing a spec",
			PromptPath:  ".speckeep/templates/prompts/grill.md",
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
			Name:        "glossary",
			Description: "Create or update the shared domain-language glossary",
			PromptPath:  ".speckeep/templates/prompts/glossary.md",
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
		{
			Name:        "diagnose",
			Description: "Optional bug diagnosis loop: reproduce, minimize, hypothesize, instrument, fix, regression-test",
			PromptPath:  ".speckeep/templates/prompts/diagnose.md",
			Extras:      nil,
			Optional:    true,
			Category:    "workflow",
		},
		{
			Name:        "repo-map",
			Description: "Update REPOSITORY_MAP.md navigation index",
			PromptPath:  ".speckeep/templates/prompts/repo-map.md",
			Extras:      nil,
			Optional:    true,
			Category:    "workflow",
		},
		{
			Name:        "module-context",
			Description: "Document a module: write/refresh CONTEXT.md for a folder (purpose, how it works, contracts)",
			PromptPath:  ".speckeep/templates/prompts/module-context.md",
			Extras:      nil,
			Optional:    true,
			Category:    "workflow",
		},
		{
			Name:        "rollback",
			Description: "Roll back completed tasks for a feature, returning them to unfinished state",
			PromptPath:  ".speckeep/templates/prompts/rollback.md",
			Extras:      []string{scriptPath("list-open-tasks", normalizedShell)},
			Optional:    true,
			Category:    "workflow",
		},
	}
}
