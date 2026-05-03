package cli

import (
	"strings"

	"github.com/spf13/cobra"
	"speckeep/src/internal/project"
)

func newDemoCmd() *cobra.Command {
	var shell string
	var agentTargets []string

	cmd := &cobra.Command{
		Use:   "demo [path]",
		Short: "Create a demo workspace with pre-populated example artifacts",
		Long: `Create a demo workspace at the given path (default: ./speckeep-demo).

The demo workspace contains a fully worked example feature (export-report) at
the implement phase — spec, inspect report, plan, tasks, and data model are all
populated under specs/active/ so you can explore the workflow and try slash
commands immediately.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "./speckeep-demo"
			if len(args) == 1 {
				root = args[0]
			}

			result, err := project.Demo(root, project.DemoOptions{
				Shell:        shell,
				AgentTargets: agentTargets,
			})
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			printPanel(w, "speckeep demo", []string{
				"path: " + result.RootAbs,
				"script type: " + result.Shell,
				"agent targets: " + formatTargets(result.AgentTargets),
				"example feature: " + result.ExampleSlug,
			})

			if len(result.Created) > 0 {
				var lines []string
				for _, p := range result.Created {
					lines = append(lines, stylePath(w, p))
				}
				printPanel(w, "Created Demo Artifacts", lines)
			}

			printPanel(w, "Try It", []string{
				styleCmd(w, "cd "+result.RootAbs),
				styleCmd(w, "speckeep dashboard ."),
				styleCmd(w, "speckeep check "+result.ExampleSlug+" ."),
				"agent flow:",
				"  " + styleCmd(w, "/speckeep.challenge "+result.ExampleSlug) + "  (optional)",
				"  " + styleCmd(w, "/speckeep.scope "+result.ExampleSlug) + "      (optional)",
				"  " + styleCmd(w, "/speckeep.handoff "+result.ExampleSlug) + "    (optional)",
			})

			return nil
		},
	}

	cmd.Flags().StringVar(&shell, "shell", "sh", "shell for generated workflow scripts: sh or powershell")
	cmd.Flags().StringSliceVar(&agentTargets, "agents", nil, "generate agent command files for one or more targets: claude, codex, copilot, cursor, kilocode, trae, windsurf, roocode, aider, all")

	return cmd
}

func formatTargets(targets []string) string {
	if len(targets) == 0 {
		return "none"
	}
	return strings.Join(targets, ", ")
}
