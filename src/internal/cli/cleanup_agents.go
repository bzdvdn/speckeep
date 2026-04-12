package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"speckeep/src/internal/project"
)

func newCleanupAgentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cleanup-agents [path]",
		Short: "Remove orphaned agent artifacts for targets that are no longer enabled",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}
			result, err := project.CleanupAgents(root)
			if err != nil {
				return err
			}
			for _, line := range result.Messages {
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			return nil
		},
	}
}
