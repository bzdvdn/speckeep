package cli

import (
	"strings"

	"github.com/spf13/cobra"
	"speckeep/src/internal/agents"
	"speckeep/src/internal/project"
)

func newDemoCmd() *cobra.Command {
	var shell string
	var agentTargets []string

	cmd := &cobra.Command{
		Use:   "demo [path]",
		Short: "Create a demo workspace with pre-populated example artifacts",
		Long: `Create a demo workspace at the given path (default: ./speckeep-demo).

The demo workspace contains a worked example feature (export-report) at the
implement phase — spec, inspect report, plan, tasks, and data model are all
populated under specs/active/, and the first tasks carry Proof entries so
traceability can be explored immediately. Open tasks are intentional: they
show what /spk-implement would pick up next.`,
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
				styleCmd(w, "speckeep trace "+result.ExampleSlug+" ."),
				styleCmd(w, "speckeep trace "+result.ExampleSlug+" . --files"),
				"agent flow:",
				"  " + styleCmd(w, "/spk-challenge "+result.ExampleSlug) + "  (optional)",
				"  " + styleCmd(w, "/spk-scope "+result.ExampleSlug) + "      (optional)",
				"  " + styleCmd(w, "/spk-handoff "+result.ExampleSlug) + "    (optional)",
			})

			return nil
		},
	}

	cmd.Flags().StringVar(&shell, "shell", "sh", "shell for generated workflow scripts: sh or powershell")
	cmd.Flags().StringSliceVar(&agentTargets, "agents", nil, "generate agent command files for one or more targets: "+agents.TargetOptionsText()+", all")

	return cmd
}

func formatTargets(targets []string) string {
	if len(targets) == 0 {
		return "none"
	}
	return strings.Join(targets, ", ")
}
