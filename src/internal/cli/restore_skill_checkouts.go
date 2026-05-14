package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"speckeep/src/internal/project"
)

func newRestoreSkillCheckoutsCmd() *cobra.Command {
	return newSkillsRestoreCmd("skills-restore [path]", "Restore git-backed skill checkouts from the skills manifest")
}

func newSkillsRestoreSubCmd() *cobra.Command {
	return newSkillsRestoreCmd("restore [path]", "Restore git-backed skill checkouts from the skills manifest")
}

func newSkillsRestoreCmd(use, short string) *cobra.Command {
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

			result, err := project.RestoreSkillCheckouts(root)
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
			printPanel(cmd.OutOrStdout(), "skills restore summary", []string{
				fmt.Sprintf("restored: %d", len(result.Restored)),
				fmt.Sprintf("unchanged: %t", result.Unchanged),
			})
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output restore result as JSON")
	return cmd
}
