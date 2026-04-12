package cli

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"speckeep/src/internal/status"
	"speckeep/src/internal/workflow"
)

func newFeaturesCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "features [path]",
		Short: "List feature workflow status across the project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}

			results, err := status.List(root)
			if err != nil {
				return err
			}
			issueCounts, err := featureIssueCounts(root)
			if err != nil {
				return err
			}

			if jsonOutput {
				payload, err := json.MarshalIndent(results, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
				return nil
			}

			if len(results) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no features found")
				return nil
			}

			stateBySlug, err := featureStatesBySlug(root)
			if err != nil {
				return err
			}

			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "slug\tphase\tready_for\tblocked\tinspect\tverify\ttasks\tissues\tartifacts")
			for _, result := range results {
				counts := issueCounts[result.Slug]
				state := stateBySlug[result.Slug]
				fmt.Fprintf(
					tw,
					"%s\t%s\t%s\t%t\t%s\t%s\t%s\t%s\t%s\n",
					result.Slug,
					result.Phase,
					result.ReadyFor,
					result.Blocked,
					renderArtifactStatus(state.InspectExists, state.InspectStatus, state.InspectLegacy),
					renderArtifactStatus(state.VerifyExists, state.VerifyStatus, false),
					renderTaskProgress(state),
					renderIssueCounts(counts.errors, counts.warnings),
					artifactSummary(result),
				)
			}
			return tw.Flush()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output feature statuses as JSON")
	return cmd
}

func artifactSummary(result status.Result) string {
	var artifacts []string
	if result.SpecExists {
		artifacts = append(artifacts, "spec")
	}
	if result.InspectExists {
		artifacts = append(artifacts, "inspect")
	}
	if result.PlanExists {
		artifacts = append(artifacts, "plan")
	}
	if result.TasksExists {
		artifacts = append(artifacts, "tasks")
	}
	if result.VerifyExists {
		artifacts = append(artifacts, "verify")
	}
	if result.Archived {
		artifacts = append(artifacts, "archive")
	}
	if len(artifacts) == 0 {
		return "-"
	}
	return stringsJoin(artifacts, ",")
}

func stringsJoin(values []string, sep string) string {
	if len(values) == 0 {
		return ""
	}
	out := values[0]
	for _, value := range values[1:] {
		out += sep + value
	}
	return out
}

type issueCount struct {
	errors   int
	warnings int
}

func featureIssueCounts(root string) (map[string]issueCount, error) {
	findings, err := workflow.ValidateProject(root)
	if err != nil {
		return nil, err
	}

	counts := map[string]issueCount{}
	for _, finding := range findings {
		slug := workflow.FindingSlug(finding)
		if slug == "" {
			continue
		}
		current := counts[slug]
		switch finding.Level {
		case "error":
			current.errors++
		case "warning":
			current.warnings++
		}
		counts[slug] = current
	}
	return counts, nil
}

func renderIssueCounts(errors, warnings int) string {
	if errors == 0 && warnings == 0 {
		return "-"
	}
	if warnings == 0 {
		return fmt.Sprintf("%de", errors)
	}
	if errors == 0 {
		return fmt.Sprintf("%dw", warnings)
	}
	return fmt.Sprintf("%de/%dw", errors, warnings)
}

func renderArtifactStatus(exists bool, status string, legacy bool) string {
	if !exists {
		return "-"
	}
	if status == "" {
		if legacy {
			return "present*"
		}
		return "present"
	}
	if legacy {
		return status + "*"
	}
	return status
}

func renderTaskProgress(state workflow.FeatureState) string {
	if !state.TasksExists {
		return "-"
	}
	return fmt.Sprintf("%d/%d", state.TasksCompleted, state.TasksTotal)
}

func featureStatesBySlug(root string) (map[string]workflow.FeatureState, error) {
	states, err := workflow.States(root)
	if err != nil {
		return nil, err
	}

	out := make(map[string]workflow.FeatureState, len(states))
	for _, state := range states {
		out[state.Slug] = state
	}
	return out, nil
}
