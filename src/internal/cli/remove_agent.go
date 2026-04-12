package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"speckeep/src/internal/project"
)

func newRemoveAgentCmd() *cobra.Command {
	var agentTargets []string
	var legacyAgentTargets []string

	cmd := &cobra.Command{
		Use:   "remove-agent [path]",
		Short: "Remove agent targets and their generated files from an existing SpecKeep project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}
			result, err := project.RemoveAgents(root, project.RemoveAgentsOptions{Targets: append(agentTargets, legacyAgentTargets...)})
			if err != nil {
				return err
			}
			for _, line := range result.Messages {
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&agentTargets, "agents", nil, "remove one or more agent targets: claude, codex, copilot, cursor, kilocode, trae, all")
	cmd.Flags().StringSliceVar(&legacyAgentTargets, "agent", nil, "deprecated alias for --agents")
	cmd.Flags().MarkHidden("agent")

	return cmd
}
