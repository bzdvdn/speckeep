package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"speckeep/src/internal/config"
	"speckeep/src/internal/workflow"
)

type guardResult struct {
	Passed   bool           `json:"passed"`
	Blocked  int            `json:"blocked"`
	Features []guardFeature `json:"features,omitempty"`
}

type guardFeature struct {
	Slug        string `json:"slug"`
	Phase       string `json:"phase"`
	ReadyFor    string `json:"ready_for"`
	ReadyToArch bool   `json:"ready_to_close"`
	Blocked     bool   `json:"blocked"`
}

// newGuardCmd exposes a CI-friendly gate: it fails when any active feature is
// not closeable right now (open tasks, missing Proof, or a non-pass verify veto).
// Exit code 0 means the workspace holds only features that can be archived.
func newGuardCmd() *cobra.Command {
	var (
		jsonOutput bool
		slugFilter string
	)

	cmd := &cobra.Command{
		Use:   "guard [path]",
		Short: "CI gate: fail unless every feature is closeable (archive-ready)",
		Long: `Guard is the deterministic CI gate for closing features fast.

It fails with exit code 1 when at least one active feature is not archive-ready:
open tasks, missing Proof lines, blocked inspect/verify state, or branch mismatch.
Pass --slug to limit the check to one feature.

This lets a PR/CI pipeline answer "can this feature be closed now?" without
reading spec prose — only structure and deterministic proof.`,
		Example: `  speckeep guard .
  speckeep guard . --slug my-feature
  speckeep guard . --json`,
		SilenceUsage: true,
		Args:         cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}

			cfg, err := config.Load(context.Background(), root)
			if err != nil {
				return err
			}

			states, err := workflow.States(context.Background(), root)
			if err != nil {
				return err
			}

			out := guardResult{}
			for _, state := range states {
				if slugFilter != "" && state.Slug != slugFilter {
					continue
				}
				gf := guardFeature{
					Slug:        state.Slug,
					Phase:       state.Phase,
					ReadyFor:    state.ReadyFor,
					ReadyToArch: state.ReadyFor == "archive" && !state.Blocked,
					Blocked:     state.Blocked,
				}
				out.Features = append(out.Features, gf)
				// Run the phase-level structural check so proof/coverage holes
				// are caught even when the lifecycle phase looks complete.
				switch state.ReadyFor {
				case "archive":
					archiveResult, aerr := workflow.CheckArchiveReady(context.Background(), cfg, root, state.Slug, "completed", "")
					if aerr == nil && archiveResult.Failed {
						gf.ReadyToArch = false
					}
				case "inspect", "plan", "tasks", "implement", "verify":
					phaseResult, perr := phaseCheckResult(context.Background(), cfg, root, state)
					if perr == nil && phaseResult.Failed {
						gf.ReadyToArch = false
					}
				}
				if !gf.ReadyToArch {
					out.Blocked++
					out.Features[len(out.Features)-1] = gf
				}
			}

			if out.Blocked == 0 {
				out.Blocked = 0
			}
			out.Passed = out.Blocked == 0

			if jsonOutput {
				payload, err := json.MarshalIndent(out, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
			} else {
				printGuard(cmd, out)
			}

			if !out.Passed {
				return newExitError(1, "")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON; exits with code 1 when any feature is not closeable")
	cmd.Flags().StringVar(&slugFilter, "slug", "", "Limit the gate to one feature slug")
	return cmd
}

func printGuard(cmd *cobra.Command, result guardResult) {
	w := cmd.OutOrStdout()
	if result.Passed {
		printPanel(w, "speckeep guard", []string{
			fmt.Sprintf("verdict: %s — all %d features are closeable", styleOK(w, "pass"), len(result.Features)),
		})
		return
	}
	printPanel(w, "speckeep guard", []string{
		fmt.Sprintf("verdict: %s — %d feature(s) are not ready to close", styleError(w, "fail"), result.Blocked),
	})
	for _, gf := range result.Features {
		status := styleOK(w, "ready")
		if !gf.ReadyToArch {
			status = styleError(w, "blocked")
		}
		fmt.Fprintf(w, "- %-24s %-12s %-12s %s\n", gf.Slug, gf.Phase, gf.ReadyFor, status)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "run `speckeep check <slug> .` per feature for the next command")
}
