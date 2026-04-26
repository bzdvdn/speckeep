package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"speckeep/src/internal/project"
)

func newRemoveSkillCmd() *cobra.Command {
	var id string
	var noInstall bool

	cmd := &cobra.Command{
		Use:   "remove-skill [path]",
		Short: "Remove a skill from the project skills manifest",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}

			result, err := project.RemoveSkill(root, project.RemoveSkillOptions{
				ID:        id,
				NoInstall: noInstall,
			})
			if err != nil {
				return err
			}

			for _, line := range result.Messages {
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "skill id to remove")
	cmd.Flags().BoolVar(&noInstall, "no-install", false, "do not reconcile installed skills in agent folders")
	cmd.MarkFlagRequired("id")

	return cmd
}
