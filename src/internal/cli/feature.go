package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"speckeep/src/internal/workflow"
)

func newFeatureCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "feature <slug> [path]",
		Short: "Show a detailed workflow view for one feature",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 2 {
				root = args[1]
			}

			state, err := workflow.State(root, args[0])
			if err != nil {
				return err
			}
			findings, err := workflow.ValidateFeature(root, args[0])
			if err != nil {
				return err
			}
			phaseResult, err := phaseCheckResult(root, state)
			if err != nil {
				return err
			}
			checkSummary := summarizeCheckFindings(phaseResult.Findings)

			if jsonOutput {
				payload, err := json.MarshalIndent(struct {
					State         workflow.FeatureState   `json:"state"`
					Findings      []workflow.Finding      `json:"findings,omitempty"`
					CheckSummary  *checkFindingSummary    `json:"check_summary,omitempty"`
					CheckFindings []workflow.CheckFinding `json:"check_findings,omitempty"`
				}{
					State:         state,
					Findings:      findings,
					CheckSummary:  checkSummary,
					CheckFindings: phaseResult.Findings,
				}, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "slug: %s\n", state.Slug)
			fmt.Fprintf(cmd.OutOrStdout(), "phase: %s\n", state.Phase)
			if state.ReadyFor != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "ready_for: %s\n", state.ReadyFor)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "blocked: %t\n", state.Blocked)
			fmt.Fprintf(cmd.OutOrStdout(), "spec: %t\n", state.SpecExists)
			fmt.Fprintf(cmd.OutOrStdout(), "inspect: %t\n", state.InspectExists)
			if state.InspectStatus != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "inspect_status: %s\n", state.InspectStatus)
			}
			if state.InspectExists && state.InspectPath != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "inspect_path: %s\n", state.InspectPath)
				if state.InspectLegacy {
					fmt.Fprintf(cmd.OutOrStdout(), "inspect_legacy: true\n")
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "plan: %t\n", state.PlanExists)
			fmt.Fprintf(cmd.OutOrStdout(), "tasks: %t\n", state.TasksExists)
			if state.TasksExists {
				fmt.Fprintf(cmd.OutOrStdout(), "tasks_progress: %d/%d complete (%d open)\n", state.TasksCompleted, state.TasksTotal, state.TasksOpen)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "verify: %t\n", state.VerifyExists)
			if state.VerifyStatus != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "verify_status: %s\n", state.VerifyStatus)
			}
			if state.VerifyExists && state.VerifyPath != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "verify_path: %s\n", state.VerifyPath)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "archived: %t\n", state.Archived)
			errorCount, warningCount := countFindings(findings)
			fmt.Fprintf(cmd.OutOrStdout(), "issues: %s\n", renderIssueSummary(errorCount, warningCount))
			if checkSummary != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "check_issues: %s\n", renderCheckSummary(*checkSummary))
			}
			if focus := featureFocusLine(state, errorCount, warningCount); focus != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "focus: %s\n", focus)
			}
			if len(phaseResult.Findings) > 0 {
				for _, line := range topFindingLines(phaseResult.Findings, 3) {
					fmt.Fprintf(cmd.OutOrStdout(), "check_detail: %s\n", line)
				}
			}
			if len(findings) > 0 {
				errors, warnings := splitFindings(findings)
				if len(errors) > 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "errors:")
					for _, finding := range errors {
						fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", displayFeatureFinding(args[0], finding.Message))
					}
				}
				if len(warnings) > 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "warnings:")
					for _, finding := range warnings {
						fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", displayFeatureFinding(args[0], finding.Message))
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output feature view as JSON")
	cmd.AddCommand(newFeatureRepairCmd())
	return cmd
}

func countFindings(findings []workflow.Finding) (errors int, warnings int) {
	for _, finding := range findings {
		switch finding.Level {
		case "error":
			errors++
		case "warning":
			warnings++
		}
	}
	return errors, warnings
}

func renderIssueSummary(errors, warnings int) string {
	switch {
	case errors == 0 && warnings == 0:
		return "none"
	case errors > 0 && warnings > 0:
		return fmt.Sprintf("%d error(s), %d warning(s)", errors, warnings)
	case errors > 0:
		return fmt.Sprintf("%d error(s)", errors)
	default:
		return fmt.Sprintf("%d warning(s)", warnings)
	}
}

func featureFocusLine(state workflow.FeatureState, errors, warnings int) string {
	switch {
	case errors > 0:
		return "resolve blocking workflow issues before moving forward"
	case state.Blocked && state.Phase == "inspect":
		return "update the spec or inspect findings, then rerun inspect"
	case state.Blocked && state.Phase == "verify":
		return "resolve verify blockers before archive"
	case state.ReadyFor == "inspect":
		return "run inspect for this feature"
	case state.ReadyFor == "plan":
		return "write the plan package"
	case state.ReadyFor == "tasks":
		return "decompose the plan into executable tasks"
	case state.ReadyFor == "implement":
		if state.TasksOpen > 0 {
			return "finish the remaining open tasks"
		}
		return "continue implementation work"
	case state.ReadyFor == "verify":
		return "run verify and capture evidence"
	case state.ReadyFor == "archive" && warnings > 0:
		return "review warnings before archive"
	case state.ReadyFor == "archive":
		return "ready to archive when the result is accepted"
	default:
		return ""
	}
}

func splitFindings(findings []workflow.Finding) (errors []workflow.Finding, warnings []workflow.Finding) {
	for _, finding := range findings {
		switch finding.Level {
		case "error":
			errors = append(errors, finding)
		case "warning":
			warnings = append(warnings, finding)
		}
	}
	return errors, warnings
}

func displayFeatureFinding(slug, message string) string {
	return strings.ReplaceAll(message, " for slug "+slug, "")
}

func newFeatureRepairCmd() *cobra.Command {
	var jsonOutput bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "repair <slug> [path]",
		Short: "Repair safe feature-local SpecKeep issues such as legacy inspect paths",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 2 {
				root = args[1]
			}

			result, err := workflow.RepairFeature(root, args[0], dryRun)
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

			fmt.Fprintf(cmd.OutOrStdout(), "slug: %s\n", result.Slug)
			fmt.Fprintf(cmd.OutOrStdout(), "dry_run: %t\n", result.DryRun)
			fmt.Fprintf(cmd.OutOrStdout(), "changed: %t\n", result.Changed)
			for _, action := range result.Actions {
				fmt.Fprintf(cmd.OutOrStdout(), "action: %s\n", action)
			}
			for _, warning := range result.Warnings {
				fmt.Fprintf(cmd.OutOrStdout(), "warning: %s\n", warning)
			}
			if !result.Changed && len(result.Warnings) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "status: no safe repairs were needed")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show repairs without applying them")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output repair result as JSON")
	return cmd
}
