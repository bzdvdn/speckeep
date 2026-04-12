package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"speckeep/src/internal/workflow"
)

type checkResult struct {
	Slug          string                  `json:"slug"`
	Phase         string                  `json:"phase"`
	Verdict       string                  `json:"verdict"`
	NextCommand   string                  `json:"next_command,omitempty"`
	Blocked       bool                    `json:"blocked"`
	Artifacts     checkArtifacts          `json:"artifacts"`
	CheckSummary  *checkFindingSummary    `json:"check_summary,omitempty"`
	CheckFindings []workflow.CheckFinding `json:"check_findings,omitempty"`
}

type checkArtifacts struct {
	Spec    checkArtifact `json:"spec"`
	Inspect checkArtifact `json:"inspect"`
	Plan    checkArtifact `json:"plan"`
	Tasks   checkArtifact `json:"tasks"`
	Verify  checkArtifact `json:"verify"`
}

type checkArtifact struct {
	Present bool   `json:"present"`
	Detail  string `json:"detail,omitempty"`
}

type checkAllResult struct {
	Blocked  bool          `json:"blocked"`
	Features []checkResult `json:"features"`
}

type checkFindingSummary struct {
	Errors            int            `json:"errors"`
	Warnings          int            `json:"warnings"`
	ErrorCategories   map[string]int `json:"error_categories,omitempty"`
	WarningCategories map[string]int `json:"warning_categories,omitempty"`
}

func newCheckCmd() *cobra.Command {
	var jsonOutput bool
	var allFeatures bool

	cmd := &cobra.Command{
		Use:   "check <slug> [path]",
		Short: "Check feature readiness and show next action",
		Args: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			if all {
				if len(args) > 1 {
					return fmt.Errorf("accepts at most 1 arg (path) when --all is set, received %d", len(args))
				}
				return nil
			}
			if len(args) < 1 || len(args) > 2 {
				return fmt.Errorf("accepts 1 or 2 args (slug [path]), received %d", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if allFeatures {
				root := "."
				if len(args) == 1 {
					root = args[0]
				}
				return runCheckAll(cmd, root, jsonOutput)
			}

			root := "."
			if len(args) == 2 {
				root = args[1]
			}

			state, err := workflow.State(root, args[0])
			if err != nil {
				return err
			}

			result, err := buildCheckResult(root, state)
			if err != nil {
				return err
			}

			if jsonOutput {
				payload, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
				if result.Blocked {
					return fmt.Errorf("blocked")
				}
				return nil
			}

			printCheck(cmd, state, result)
			if result.Blocked {
				return fmt.Errorf("blocked")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON; exits with code 1 when blocked")
	cmd.Flags().BoolVar(&allFeatures, "all", false, "check all features; exits with code 1 if any are blocked")
	return cmd
}

func runCheckAll(cmd *cobra.Command, root string, jsonOutput bool) error {
	states, err := workflow.States(root)
	if err != nil {
		return err
	}

	results := make([]checkResult, 0, len(states))
	anyBlocked := false
	for _, state := range states {
		r, err := buildCheckResult(root, state)
		if err != nil {
			return err
		}
		results = append(results, r)
		if r.Blocked {
			anyBlocked = true
		}
	}

	if jsonOutput {
		payload, err := json.MarshalIndent(checkAllResult{Blocked: anyBlocked, Features: results}, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(payload))
		if anyBlocked {
			return fmt.Errorf("blocked")
		}
		return nil
	}

	printCheckAll(cmd, states, results)
	if anyBlocked {
		return fmt.Errorf("blocked")
	}
	return nil
}

func buildCheckResult(root string, state workflow.FeatureState) (checkResult, error) {
	result := checkResult{
		Slug:    state.Slug,
		Phase:   state.Phase,
		Blocked: state.Blocked,
		Artifacts: checkArtifacts{
			Spec:    checkArtifact{Present: state.SpecExists},
			Inspect: checkArtifact{Present: state.InspectExists},
			Plan:    checkArtifact{Present: state.PlanExists},
			Tasks:   checkArtifact{Present: state.TasksExists},
			Verify:  checkArtifact{Present: state.VerifyExists},
		},
	}

	if state.InspectExists && state.InspectStatus != "" {
		result.Artifacts.Inspect.Detail = state.InspectStatus
	}

	if state.TasksExists {
		result.Artifacts.Tasks.Detail = fmt.Sprintf("%d/%d done", state.TasksCompleted, state.TasksTotal)
	}

	if state.VerifyExists && state.VerifyStatus != "" {
		result.Artifacts.Verify.Detail = state.VerifyStatus
	}

	if state.Blocked {
		result.Verdict = "blocked"
	} else {
		result.Verdict = "ready"
	}

	result.NextCommand = nextCommand(state)

	phaseResult, err := phaseCheckResult(root, state)
	if err != nil {
		return checkResult{}, err
	}
	if len(phaseResult.Findings) > 0 {
		result.CheckFindings = phaseResult.Findings
	}
	if summary := summarizeCheckFindings(phaseResult.Findings); summary != nil {
		result.CheckSummary = summary
	}

	return result, nil
}

func nextCommand(state workflow.FeatureState) string {
	switch state.ReadyFor {
	case "spec":
		return "/speckeep.spec " + state.Slug
	case "inspect":
		return "/speckeep.inspect " + state.Slug
	case "plan":
		return "/speckeep.plan " + state.Slug
	case "tasks":
		return "/speckeep.tasks " + state.Slug
	case "implement":
		return "/speckeep.implement " + state.Slug
	case "verify":
		return "/speckeep.verify " + state.Slug
	case "archive":
		return "/speckeep.archive " + state.Slug
	default:
		return ""
	}
}

func printCheck(cmd *cobra.Command, state workflow.FeatureState, result checkResult) {
	w := cmd.OutOrStdout()

	nextLine := "-"
	if result.NextCommand != "" {
		nextLine = styleCmd(w, result.NextCommand)
	}
	verdictLine := styleOK(w, "ready")
	if state.Blocked {
		verdictLine = styleError(w, "blocked")
	}

	printPanel(w, "speckeep check", []string{
		"feature: " + state.Slug,
		"phase: " + strings.ToUpper(state.Phase),
		"verdict: " + verdictLine,
		"next: " + nextLine,
	})

	printPanel(w, "Artifacts", []string{
		"spec: " + artifactLine(w, state.SpecExists, ""),
		"inspect: " + artifactLine(w, state.InspectExists, state.InspectStatus),
		"plan: " + artifactLine(w, state.PlanExists, ""),
		"tasks: " + artifactLine(w, state.TasksExists, taskDetail(state)),
		"verify: " + artifactLine(w, state.VerifyExists, state.VerifyStatus),
	})

	if state.BranchMismatch {
		printPanel(w, "Branch Mismatch", []string{
			"current: " + state.CurrentBranch,
			"expected: feature/" + state.Slug,
		})
	}

	// Keep stable lines for grep/tests and scripting.
	fmt.Fprintln(w)
	fmt.Fprintf(w, "verdict:  %s\n", result.Verdict)
	if result.NextCommand != "" {
		fmt.Fprintf(w, "next:     %s\n", result.NextCommand)
	}
	if result.CheckSummary != nil && (result.CheckSummary.Errors > 0 || result.CheckSummary.Warnings > 0) {
		fmt.Fprintf(w, "checks:   %s\n", renderCheckSummary(*result.CheckSummary))
	}
	if len(result.CheckFindings) > 0 {
		for _, line := range topFindingLines(result.CheckFindings, 3) {
			fmt.Fprintf(w, "detail:   %s\n", line)
		}
	}

	if state.InspectLegacy {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "warning:  inspect report is at legacy path — run `speckeep feature repair "+state.Slug+"`")
	}
}

func taskDetail(state workflow.FeatureState) string {
	if !state.TasksExists {
		return ""
	}
	tasksDetail := fmt.Sprintf("%d/%d done", state.TasksCompleted, state.TasksTotal)
	if state.TasksOpen > 0 {
		tasksDetail += fmt.Sprintf("  (%d open)", state.TasksOpen)
	}
	return tasksDetail
}

func artifactLine(w io.Writer, present bool, detail string) string {
	if !present {
		return styleError(w, "missing")
	}

	parts := []string{styleOK(w, "✓")}
	if detail != "" {
		parts = append(parts, " ", styleMuted(w, detail))
	}
	return strings.Join(parts, "")
}

func printCheckAll(cmd *cobra.Command, states []workflow.FeatureState, results []checkResult) {
	w := cmd.OutOrStdout()

	if len(results) == 0 {
		printPanel(w, "speckeep check --all", []string{"no features found"})
		return
	}

	blockedCount := 0
	for i, r := range results {
		_ = states[i]
		if r.Blocked {
			blockedCount++
		}
	}

	if blockedCount > 0 {
		printPanel(w, "speckeep check --all", []string{
			fmt.Sprintf("verdict: %s", styleError(w, fmt.Sprintf("%d of %d features blocked", blockedCount, len(results)))),
		})
	} else {
		printPanel(w, "speckeep check --all", []string{
			fmt.Sprintf("verdict: %s", styleOK(w, fmt.Sprintf("all %d features ready", len(results)))),
		})
	}

	slugWidth := visibleRuneLen("feature")
	phaseWidth := visibleRuneLen("phase")
	for _, r := range results {
		if l := visibleRuneLen(r.Slug); l > slugWidth {
			slugWidth = l
		}
		if l := visibleRuneLen(r.Phase); l > phaseWidth {
			phaseWidth = l
		}
	}

	header := fmt.Sprintf("%s  %s  %-8s  %s",
		padRightVisible("feature", slugWidth),
		padRightVisible("phase", phaseWidth),
		"verdict",
		"next",
	)
	fmt.Fprintln(w, header)
	fmt.Fprintln(w, fmt.Sprintf("%s  %s  %-8s  %s",
		strings.Repeat("-", slugWidth),
		strings.Repeat("-", phaseWidth),
		"-------",
		"----",
	))

	for _, r := range results {
		verdict := styleOK(w, "ready")
		if r.Blocked {
			verdict = styleError(w, "blocked")
		}
		next := r.NextCommand
		if next != "" {
			next = styleCmd(w, next)
		}
		fmt.Fprintf(w, "%s  %s  %-8s  %s\n",
			padRightVisible(r.Slug, slugWidth),
			padRightVisible(r.Phase, phaseWidth),
			padRightVisible(verdict, 8),
			next,
		)
	}
	fmt.Fprintln(w)
}

func phaseCheckResult(root string, state workflow.FeatureState) (workflow.CheckResult, error) {
	switch state.ReadyFor {
	case "inspect":
		return workflow.CheckInspectReady(root, state.Slug)
	case "plan":
		return workflow.CheckPlanReady(root, state.Slug)
	case "tasks":
		return workflow.CheckTasksReady(root, state.Slug)
	case "implement":
		return workflow.CheckImplementReady(root, state.Slug)
	case "verify":
		return workflow.CheckVerifyReady(root, state.Slug)
	default:
		return workflow.CheckResult{}, nil
	}
}

func summarizeCheckFindings(findings []workflow.CheckFinding) *checkFindingSummary {
	if len(findings) == 0 {
		return nil
	}

	summary := &checkFindingSummary{
		ErrorCategories:   map[string]int{},
		WarningCategories: map[string]int{},
	}
	for _, finding := range findings {
		switch finding.Severity {
		case workflow.SeverityError:
			summary.Errors++
			summary.ErrorCategories[string(finding.Category)]++
		case workflow.SeverityWarning:
			summary.Warnings++
			summary.WarningCategories[string(finding.Category)]++
		}
	}
	if summary.Errors == 0 && summary.Warnings == 0 {
		return nil
	}
	if len(summary.ErrorCategories) == 0 {
		summary.ErrorCategories = nil
	}
	if len(summary.WarningCategories) == 0 {
		summary.WarningCategories = nil
	}
	return summary
}

func renderCheckSummary(summary checkFindingSummary) string {
	parts := make([]string, 0, 2)
	if summary.Errors > 0 {
		parts = append(parts, fmt.Sprintf("errors=%d", summary.Errors))
	}
	if summary.Warnings > 0 {
		parts = append(parts, fmt.Sprintf("warnings=%d", summary.Warnings))
	}
	if cats := renderCategoryCounts(summary.ErrorCategories); cats != "" {
		parts = append(parts, "error_categories="+cats)
	}
	if cats := renderCategoryCounts(summary.WarningCategories); cats != "" {
		parts = append(parts, "warning_categories="+cats)
	}
	return strings.Join(parts, "  ")
}

func renderCategoryCounts(counts map[string]int) string {
	if len(counts) == 0 {
		return ""
	}
	order := []string{"structure", "traceability", "ambiguity", "consistency", "readiness"}
	parts := make([]string, 0, len(counts))
	for _, key := range order {
		if counts[key] > 0 {
			parts = append(parts, fmt.Sprintf("%s=%d", key, counts[key]))
		}
	}
	for key, count := range counts {
		known := false
		for _, ordered := range order {
			if key == ordered {
				known = true
				break
			}
		}
		if !known {
			parts = append(parts, fmt.Sprintf("%s=%d", key, count))
		}
	}
	return strings.Join(parts, ",")
}

func topFindingLines(findings []workflow.CheckFinding, limit int) []string {
	lines := make([]string, 0, limit)
	for _, finding := range findings {
		if finding.Severity != workflow.SeverityError && finding.Severity != workflow.SeverityWarning {
			continue
		}
		line := string(finding.Severity) + " [" + string(finding.Category) + "] " + finding.Message
		if len(finding.Refs) > 0 {
			line += " (" + strings.Join(finding.Refs, ", ") + ")"
		}
		lines = append(lines, line)
		if len(lines) == limit {
			break
		}
	}
	return lines
}
