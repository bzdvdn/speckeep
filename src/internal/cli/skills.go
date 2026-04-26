package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"speckeep/src/internal/project"
)

func newSkillsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Manage project skills and skills integration",
	}
	cmd.AddCommand(newSkillsInstallSubCmd())
	cmd.AddCommand(newSkillsSyncSubCmd())
	return cmd
}

func newSyncSkillsCmd() *cobra.Command {
	return newSkillsSyncCmd("sync-skills [path]", "Sync skills manifest and AGENTS.md skills block")
}

func newSkillsSyncSubCmd() *cobra.Command {
	return newSkillsSyncCmd("sync [path]", "Sync skills manifest and AGENTS.md skills block")
}

func newSkillsSyncCmd(use, short string) *cobra.Command {
	var dryRun bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}

			result, err := project.SyncSkills(root, project.SyncSkillsOptions{
				DryRun: dryRun,
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
			printPanel(cmd.OutOrStdout(), "skills sync summary", []string{
				fmt.Sprintf("created: %d", len(result.Created)),
				fmt.Sprintf("updated: %d", len(result.Updated)),
				fmt.Sprintf("unchanged: %d", len(result.Unchanged)),
				fmt.Sprintf("dry-run: %t", result.DryRun),
			})
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show which skills-managed files would change without writing them")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output sync result as JSON")
	return cmd
}
