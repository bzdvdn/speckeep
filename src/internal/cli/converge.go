package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"speckeep/src/internal/config"
	"speckeep/src/internal/workflow"
)

type convergeResult struct {
	Slug      string                  `json:"slug"`
	Converged bool                    `json:"converged"`
	ExitHint  string                  `json:"exit_hint"`
	Findings  []workflow.CheckFinding `json:"findings,omitempty"`
}

func newConvergeCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "converge <slug> [path]",
		Short: "Check that an implemented feature converges toward its tasks/spec",
		Long: `Run the cheap convergence loop for one feature: confirms every completed
task carries valid Proof, every Proof file exists, touched surfaces exist, and
AC-* coverage holds. When gaps remain it reports them as structured findings and
exits with code 1 — the agent should append follow-up tasks and re-run converge.

Converge is lighter than verify: it is the "implement -> assessed -> fixed"
loop that closes features fast without producing a verify report.`,
		Example: `  speckeep converge my-feature .
  speckeep converge my-feature . --json`,
		SilenceUsage: true,
		Args:         cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 2 {
				root = args[1]
			}
			slug := args[0]

			cfg, err := config.Load(context.Background(), root)
			if err != nil {
				return err
			}
			result, err := workflow.CheckConvergeReady(context.Background(), cfg, root, slug)
			if err != nil {
				return err
			}

			summary := summarizeCheckFindings(result.Findings)
			exitHint := ""
			if summary != nil {
				exitHint = renderCheckSummary(*summary)
			}
			converged := !result.Failed

			out := convergeResult{
				Slug:      slug,
				Converged: converged,
				ExitHint:  exitHint,
				Findings:  result.Findings,
			}

			if jsonOutput {
				payload, err := json.MarshalIndent(out, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
			} else {
				printConverge(cmd, out, result)
			}

			if !converged {
				return newExitError(1, "")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON; exits with code 1 when the feature is not converged")
	return cmd
}

func printConverge(cmd *cobra.Command, out convergeResult, result workflow.CheckResult) {
	w := cmd.OutOrStdout()
	printPanel(w, "speckeep converge", []string{
		"feature: " + out.Slug,
		"verdict: " + convergeVerdictLine(w, out.Converged),
	})
	if len(result.Lines) == 0 {
		fmt.Fprintln(w, "no findings")
		return
	}
	for _, line := range result.Lines {
		fmt.Fprintln(w, line)
	}
	fmt.Fprintf(w, "\n%s\n", styleCmd(w, nextConvergeCommand(out.Converged)))
}

func convergeVerdictLine(w interface{ Write([]byte) (int, error) }, converged bool) string {
	if converged {
		return styleOK(w, "converged")
	}
	return styleError(w, "gaps-found")
}

func nextConvergeCommand(converged bool) string {
	if converged {
		return "Ready for: speckeep archive <slug> ."
	}
	return "Next: append follow-up tasks, then re-run speckeep converge <slug>"
}
