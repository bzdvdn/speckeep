package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"speckeep/src/internal/status"
)

func newStatusCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:     "status <slug> [path]",
		Short:   "Show feature workflow status for one slug",
		Long:    "Shows the current workflow phase and artifact readiness for a single feature.",
		Example: "  speckeep status export-report .\n  speckeep status export-report . --json",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 2 {
				root = args[1]
			}

			result, err := status.Check(root, args[0])
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

			w := cmd.OutOrStdout()
			next := "-"
			if result.ReadyFor != "" {
				next = styleCmd(w, "/speckeep."+result.ReadyFor+" "+result.Slug)
			}
			blocked := styleOK(w, "false")
			if result.Blocked {
				blocked = styleError(w, "true")
			}

			printPanel(w, "speckeep status", []string{
				"slug: " + result.Slug,
				"phase: " + result.Phase,
				"ready_for: " + result.ReadyFor,
				"blocked: " + blocked,
				"next: " + next,
			})

			printPanel(w, "Artifacts", []string{
				"spec: " + boolArtifact(w, result.SpecExists),
				"inspect: " + boolArtifact(w, result.InspectExists),
				"plan: " + boolArtifact(w, result.PlanExists),
				"tasks: " + boolArtifact(w, result.TasksExists),
				"verify: " + boolArtifact(w, result.VerifyExists),
				"archived: " + boolArtifact(w, result.Archived),
			})

			if result.TasksExists {
				printPanel(w, "Tasks", []string{
					fmt.Sprintf("total: %d", result.TasksTotal),
					fmt.Sprintf("completed: %d", result.TasksCompleted),
					fmt.Sprintf("open: %d", result.TasksOpen),
				})
			}

			if result.InspectPath != "" || result.VerifyPath != "" {
				lines := []string{}
				if result.InspectPath != "" {
					lines = append(lines, "inspect_path: "+stylePath(w, result.InspectPath))
				}
				if result.VerifyPath != "" {
					lines = append(lines, "verify_path: "+stylePath(w, result.VerifyPath))
				}
				printPanel(w, "Paths", lines)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output feature status as JSON")
	return cmd
}

func boolArtifact(w io.Writer, present bool) string {
	if present {
		return styleOK(w, "present")
	}
	return styleError(w, "missing")
}
