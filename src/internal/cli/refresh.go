package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"speckeep/src/internal/project"
)

func newRefreshCmd() *cobra.Command {
	var defaultLang string
	var docsLang string
	var agentLang string
	var commentsLang string
	var shell string
	var constitutionFile string
	var specsDir string
	var archiveDir string
	var jsonOutput bool
	var dryRun bool
	var rewriteTrace bool
	var agentTargets []string
	var legacyAgentTargets []string

	cmd := &cobra.Command{
		Use:   "refresh [path]",
		Short: "Refresh generated SpecKeep artifacts for an existing project without touching authored feature state",
		Long: `Refreshes SpecKeep-managed artifacts inside an already initialized project.

Synchronizes:
  - .speckeep/speckeep.yaml (config)
  - managed templates/prompts/scripts inside .speckeep/
  - the managed SpecKeep block in AGENTS.md
  - agent-target artifacts (.claude/, .cursor/, etc.)

	Does not touch authored feature state:
	  - feature contents under specs_dir are not modified (but can be moved with --specs-dir).
	  - archive contents under archive_dir are not modified (but can be moved with --archive-dir).

	Tip: use --dry-run to preview changes without writing.`,
		Example: "  speckeep refresh .\n  speckeep refresh . --dry-run\n  speckeep refresh . --agents claude,cursor --agent-lang en",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}

			if !jsonOutput {
				printPanel(cmd.OutOrStdout(), "speckeep refresh", []string{
					"Sync managed files (without modifying specs/plans).",
					"Tip: add --dry-run to preview changes.",
				})
			}

			result, err := project.Refresh(root, project.RefreshOptions{
				DefaultLang:      defaultLang,
				DocsLang:         docsLang,
				AgentLang:        agentLang,
				CommentsLang:     commentsLang,
				Shell:            shell,
				ConstitutionFile: constitutionFile,
				SpecsDir:         specsDir,
				ArchiveDir:       archiveDir,
				AgentTargets:     append(agentTargets, legacyAgentTargets...),
				DryRun:           dryRun,
				RewriteTrace:     rewriteTrace,
			})
			if err != nil {
				return err
			}

			if jsonOutput {
				payload, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
				return nil
			}

			for _, line := range result.Messages {
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}

			printPanel(cmd.OutOrStdout(), "refresh summary", []string{
				fmt.Sprintf("created: %d", len(result.Created)),
				fmt.Sprintf("updated: %d", len(result.Updated)),
				fmt.Sprintf("rewritten: %d", len(result.Rewritten)),
				fmt.Sprintf("removed: %d", len(result.Removed)),
				fmt.Sprintf("unchanged: %d", len(result.Unchanged)),
				fmt.Sprintf("dry-run: %t", result.DryRun),
			})
			return nil
		},
	}

	cmd.Flags().StringVar(&defaultLang, "lang", "", "override the base language for generated docs and prompts: en or ru")
	cmd.Flags().StringVar(&docsLang, "docs-lang", "", "override the generated documentation language: en or ru")
	cmd.Flags().StringVar(&agentLang, "agent-lang", "", "override the generated prompt and AGENTS guidance language: en or ru")
	cmd.Flags().StringVar(&commentsLang, "comments-lang", "", "override the preferred code comment language: en or ru")
	cmd.Flags().StringVar(&shell, "shell", "", "override the generated workflow script family: sh or powershell")
	cmd.Flags().StringVar(&constitutionFile, "constitution-file", "", "override the constitution file path and (safely) move the existing file when possible")
	cmd.Flags().StringVar(&specsDir, "specs-dir", "", "override paths.specs_dir and (safely) move the existing specs directory when possible")
	cmd.Flags().StringVar(&archiveDir, "archive-dir", "", "override paths.archive_dir and (safely) move the existing archive directory when possible")
	cmd.Flags().StringSliceVar(&agentTargets, "agents", nil, "override enabled project-local agent targets: claude, codex, copilot, cursor, kilocode, trae, all")
	cmd.Flags().StringSliceVar(&legacyAgentTargets, "agent", nil, "deprecated alias for --agents")
	cmd.Flags().MarkHidden("agent")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show which managed files would change without writing them")
	cmd.Flags().BoolVar(&rewriteTrace, "rewrite-trace", false, "Rewrite legacy trace annotations in code: @ds-task/@ds-test → @sk-task/@sk-test")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output refresh results as JSON")

	return cmd
}
