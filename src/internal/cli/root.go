package cli

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "speckeep",
		Short: "A lightweight project context kit for development agents and humans",
		Long: `speckeep — specification-driven development kit for agents and humans

Strict phase chain: constitution → spec → inspect → plan → tasks → implement → verify → archive

Quick start:
  speckeep init . --lang en --shell sh --agents codex
  speckeep doctor .
  speckeep list-specs .

For agents (Kilocode/Claude/Cursor):
  /speckeep.constitution                        — create a constitution
  /speckeep.spec --name "feature name"          — create a spec
  /speckeep.spec --amend                        — targeted spec edit
  /speckeep.plan <slug> [--research|--update]   — create a plan
  /speckeep.tasks <slug>                        — decompose into tasks
  /speckeep.implement <slug>                    — implement tasks
  /speckeep.verify <slug> [--deep]              — verify AC coverage
  /speckeep.archive <slug> [--restore]          — archive/restore

Optional commands (any phase):
  /speckeep.challenge <slug> [--spec|--plan]    — adversarial review
  /speckeep.handoff [slug]                      — session handoff doc
  /speckeep.hotfix <slug>                       — emergency fix
  /speckeep.scope <slug>                        — scope boundary check
  /speckeep.recap                               — recap active features

CLI commands:
  speckeep doctor .                      — workspace health check
  speckeep list-specs .                  — list active specs
  speckeep check <slug> . [--json]       — feature status
  speckeep check . --all                 — all features table
  speckeep dashboard .                   — visual dashboard
  speckeep trace <slug> . [--tests]      — code traceability
  speckeep export <slug> . --output f.md — export artifacts
  speckeep list-archive [path]           — list archived features

Documentation:
  README.md — overview and examples
  docs/en/ or docs/ru/ — extended documentation`,
		Version: Version,
	}

	cmd.SetHelpTemplate(`{{speckeepBanner .}}{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`)

	cmd.SetHelpCommand(newHelpCmd())

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newRefreshCmd())
	cmd.AddCommand(newAddAgentCmd())
	cmd.AddCommand(newListAgentsCmd())
	cmd.AddCommand(newRemoveAgentCmd())
	cmd.AddCommand(newCleanupAgentsCmd())
	cmd.AddCommand(newDoctorCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newDashboardCmd())
	cmd.AddCommand(newFeatureCmd())
	cmd.AddCommand(newFeaturesCmd())
	cmd.AddCommand(newMigrateCmd())
	cmd.AddCommand(newListSpecsCmd())
	cmd.AddCommand(newShowSpecCmd())
	cmd.AddCommand(newCheckCmd())
	cmd.AddCommand(newTraceCmd())
	cmd.AddCommand(newDemoCmd())
	cmd.AddCommand(newExportCmd())
	cmd.AddCommand(newExploreCmd())
	cmd.AddCommand(newContextCmd())
	cmd.AddCommand(newSchemaCmd())
	cmd.AddCommand(newRiskCmd())
	cmd.AddCommand(newInternalCmd())
	cmd.AddCommand(newArchiveCmd())
	cmd.AddCommand(newListArchiveCmd())

	return cmd
}
