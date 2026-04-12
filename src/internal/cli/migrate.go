package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"speckeep/src/internal/workflow"
)

func newMigrateCmd() *cobra.Command {
	var dryRun bool
	var jsonOutput bool
	var copyWorkspace bool

	cmd := &cobra.Command{
		Use:   "migrate [path]",
		Short: "Migrate safe legacy SpecKeep artifacts to canonical paths",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}

			result, err := workflow.MigrateProject(root, dryRun, copyWorkspace)
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

			fmt.Fprintf(cmd.OutOrStdout(), "dry_run: %t\n", result.DryRun)
			fmt.Fprintf(cmd.OutOrStdout(), "changed: %t\n", result.Changed)
			for _, repair := range result.Results {
				fmt.Fprintf(cmd.OutOrStdout(), "slug: %s\n", repair.Slug)
				for _, action := range repair.Actions {
					fmt.Fprintf(cmd.OutOrStdout(), "action: %s\n", action)
				}
				for _, warning := range repair.Warnings {
					fmt.Fprintf(cmd.OutOrStdout(), "warning: %s\n", warning)
				}
			}
			for _, warning := range result.Warnings {
				fmt.Fprintf(cmd.OutOrStdout(), "warning: %s\n", warning)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show migrations without applying them")
	cmd.Flags().BoolVar(&copyWorkspace, "copy", false, "Copy legacy .draftspec workspace into .speckeep instead of moving it (when applicable)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output migration result as JSON")
	return cmd
}
