package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"speckeep/src/internal/project"
)

func newListAgentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-agents [path]",
		Short: "List enabled agent targets for an existing SpecKeep project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}
			result, err := project.ListAgents(root)
			if err != nil {
				return err
			}
			for _, target := range result.Targets {
				fmt.Fprintln(cmd.OutOrStdout(), target)
			}
			return nil
		},
	}
}
