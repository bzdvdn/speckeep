package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"speckeep/src/internal/project"
)

func newAddAgentCmd() *cobra.Command {
	var agentTargets []string
	var legacyAgentTargets []string
	var agentLang string

	cmd := &cobra.Command{
		Use:   "add-agent [path]",
		Short: "Generate additional agent command files for an existing SpecKeep project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}

			result, err := project.AddAgents(root, project.AddAgentsOptions{
				Targets:   append(agentTargets, legacyAgentTargets...),
				AgentLang: agentLang,
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

	cmd.Flags().StringSliceVar(&agentTargets, "agents", nil, "generate project-local agent command files for one or more targets: claude, codex, copilot, cursor, kilocode, trae, all")
	cmd.Flags().StringSliceVar(&legacyAgentTargets, "agent", nil, "deprecated alias for --agents")
	cmd.Flags().MarkHidden("agent")
	cmd.Flags().StringVar(&agentLang, "agent-lang", "", "override the configured agent language for generated agent files: en or ru")

	return cmd
}
