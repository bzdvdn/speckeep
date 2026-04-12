package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"speckeep/src/internal/workflow"
)

func newDashboardCmd() *cobra.Command {
	var includeArchived bool
	var archivedOnly bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "dashboard [path]",
		Short: "Visual dashboard of all features and project health",
		Long: `Shows an at-a-glance dashboard of feature lifecycle state across the project.

Includes phase, task progress, readiness, and branch mismatch hints.`,
		Example: "  speckeep dashboard .\n  speckeep dashboard . --all\n  speckeep dashboard . --archived\n  speckeep dashboard . --json\n  speckeep dashboard /path/to/project",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}

			if includeArchived && archivedOnly {
				return newExitError(1, "flags --all and --archived are mutually exclusive")
			}

			states, err := workflow.States(root)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			active, archived := splitStates(states)
			displayed := filterDisplayedStates(states, includeArchived, archivedOnly)
			summary := computeDashboardSummary(active, archived, displayed, includeArchived, archivedOnly)

			if jsonOutput {
				payload, err := json.MarshalIndent(summary, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(w, string(payload))
				return nil
			}

			printDashboardSummary(w, summary)

			if len(displayed) == 0 {
				if len(active) == 0 && len(archived) > 0 && !includeArchived && !archivedOnly {
					printPanel(w, "No Active Features Found", []string{
						"Use " + styleCmd(w, "speckeep dashboard . --archived") + " to view archived features.",
						"Or use " + styleCmd(w, "speckeep dashboard . --all") + " to view everything.",
					})
					return nil
				}
				printPanel(w, "No Features Found", []string{
					"Create your first spec with " + styleCmd(w, "/speckeep.spec") + ".",
					"Or run " + styleCmd(w, "speckeep demo") + " to explore an example workspace.",
				})
				return nil
			}

			printDashboardTable(w, displayed)
			printDashboardLegend(w)

			return nil
		},
	}

	cmd.Flags().BoolVar(&includeArchived, "all", false, "Include archived features in the dashboard")
	cmd.Flags().BoolVar(&archivedOnly, "archived", false, "Show only archived features")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output dashboard data as JSON")

	return cmd
}

func splitStates(states []workflow.FeatureState) (active []workflow.FeatureState, archived []workflow.FeatureState) {
	for _, state := range states {
		if state.Archived {
			archived = append(archived, state)
			continue
		}
		active = append(active, state)
	}
	return active, archived
}

type dashboardSummary struct {
	ActiveCount         int                     `json:"active_count"`
	ArchivedCount       int                     `json:"archived_count"`
	DisplayedCount      int                     `json:"displayed_count"`
	Showing             string                  `json:"showing"`
	BlockedCount        int                     `json:"blocked_count"`
	BranchMismatchCount int                     `json:"branch_mismatch_count"`
	PhaseCounts         map[string]int          `json:"phase_counts,omitempty"`
	Features            []workflow.FeatureState `json:"features,omitempty"`
}

func filterDisplayedStates(states []workflow.FeatureState, includeArchived bool, archivedOnly bool) []workflow.FeatureState {
	if includeArchived {
		return append([]workflow.FeatureState(nil), states...)
	}
	displayed := make([]workflow.FeatureState, 0, len(states))
	for _, state := range states {
		if archivedOnly && !state.Archived {
			continue
		}
		if !archivedOnly && state.Archived {
			continue
		}
		displayed = append(displayed, state)
	}
	return displayed
}

func computeDashboardSummary(active []workflow.FeatureState, archived []workflow.FeatureState, displayed []workflow.FeatureState, includeArchived bool, archivedOnly bool) dashboardSummary {
	blocked := 0
	branchMismatch := 0
	phaseCounts := map[string]int{}
	for _, state := range displayed {
		if state.Blocked {
			blocked++
		}
		if state.BranchMismatch {
			branchMismatch++
		}
		phaseCounts[state.Phase]++
	}

	phases := make([]string, 0, len(phaseCounts))
	for phase := range phaseCounts {
		phases = append(phases, phase)
	}
	sort.Strings(phases)

	var phaseLineParts []string
	for _, phase := range phases {
		phaseLineParts = append(phaseLineParts, fmt.Sprintf("%s=%d", strings.ToUpper(phase), phaseCounts[phase]))
	}

	showing := "active"
	switch {
	case archivedOnly:
		showing = "archived"
	case includeArchived:
		showing = "active+archived"
	}

	return dashboardSummary{
		ActiveCount:         len(active),
		ArchivedCount:       len(archived),
		DisplayedCount:      len(displayed),
		Showing:             showing,
		BlockedCount:        blocked,
		BranchMismatchCount: branchMismatch,
		PhaseCounts:         phaseCounts,
		Features:            displayed,
	}
}

func printDashboardSummary(w io.Writer, summary dashboardSummary) {
	phases := make([]string, 0, len(summary.PhaseCounts))
	for phase := range summary.PhaseCounts {
		phases = append(phases, phase)
	}
	sort.Strings(phases)

	var phaseLineParts []string
	for _, phase := range phases {
		phaseLineParts = append(phaseLineParts, fmt.Sprintf("%s=%d", strings.ToUpper(phase), summary.PhaseCounts[phase]))
	}
	phaseLine := "phases: " + strings.Join(phaseLineParts, "  ")
	if len(phaseLineParts) == 0 {
		phaseLine = "phases: none"
	}

	printPanel(w, "speckeep dashboard", []string{
		fmt.Sprintf("features: %d active, %d archived", summary.ActiveCount, summary.ArchivedCount),
		fmt.Sprintf("showing: %s (%d)", summary.Showing, summary.DisplayedCount),
		fmt.Sprintf("blocked: %d", summary.BlockedCount),
		fmt.Sprintf("branch mismatches: %d", summary.BranchMismatchCount),
		phaseLine,
	})
}

func printDashboardTable(w io.Writer, states []workflow.FeatureState) {
	color := useColor(w)

	const (
		slugW   = 22
		phaseW  = 10
		readyW  = 12
		tasksW  = 9
		barW    = 12
		statusW = 8
		branchW = 24
	)

	fmt.Fprintln(w, "  SLUG                   PHASE       READY FOR     TASKS     PROGRESS      STATUS    BRANCH")
	fmt.Fprintln(w, "  ────────────────────── ──────────  ────────────  ────────  ────────────  ────────  ────────────────────────")

	for _, state := range states {
		slug := truncateRunes(state.Slug, slugW)

		phase := strings.ToUpper(state.Phase)
		readyFor := "-"
		if strings.TrimSpace(state.ReadyFor) != "" {
			readyFor = strings.ToUpper(state.ReadyFor)
		}

		tasks := "-"
		pct := 0
		if state.TasksTotal > 0 {
			tasks = fmt.Sprintf("%d/%d", state.TasksCompleted, state.TasksTotal)
			pct = (state.TasksCompleted * 100) / state.TasksTotal
		}
		progress := renderProgressBar(pct, barW, color)

		status := "READY"
		if state.Blocked {
			status = "BLOCKED"
		}
		status = styleStatus(status, state.Blocked, color)

		branch := state.CurrentBranch
		if strings.TrimSpace(branch) == "" {
			branch = "-"
		}
		branch = truncateRunes(branch, branchW)
		if state.BranchMismatch {
			branch = styleWarn(w, "!!") + " " + branch
		}

		fmt.Fprintf(w, "  %s  %s  %s  %s  %s  %s  %s\n",
			padRightVisible(slug, slugW),
			padRightVisible(stylePhase(phase, color), phaseW),
			padRightVisible(stylePhase(readyFor, color), readyW),
			padRightVisible(tasks, tasksW),
			padRightVisible(progress, barW),
			padRightVisible(status, statusW),
			padRightVisible(branch, branchW+3), // allow optional "!! "
		)
	}
	fmt.Fprintln(w)
}

func printDashboardLegend(w io.Writer) {
	color := useColor(w)
	printPanel(w, "Legend", []string{
		styleWarn(w, "!!") + " prefix in branch column = branch mismatch (expected feature/<slug>)",
		styleStatus("READY", false, color) + " / " + styleStatus("BLOCKED", true, color) + " = readiness status",
		"progress bar uses task checklist counts from specs/<slug>/plan/tasks.md",
	})
}

func padRightVisible(s string, width int) string {
	visible := visibleRuneLen(s)
	if visible >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visible)
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 3 {
		return string(r[:max])
	}
	return string(r[:max-3]) + "..."
}

func renderProgressBar(percent int, width int, color bool) string {
	if width < 4 {
		return fmt.Sprintf("%d%%", percent)
	}
	barWidth := width - 5 // " [bar] 100%"
	if barWidth < 4 {
		barWidth = 4
	}

	filled := (percent * barWidth) / 100
	if filled < 0 {
		filled = 0
	}
	if filled > barWidth {
		filled = barWidth
	}

	fill := strings.Repeat("█", filled)
	empty := strings.Repeat("░", barWidth-filled)
	if color {
		fill = colorize(fill, "\x1b[32m", true)
		empty = colorize(empty, "\x1b[90m", true)
	}
	return fmt.Sprintf("%s%s %3d%%", fill, empty, percent)
}

func styleStatus(s string, blocked bool, color bool) string {
	if !color {
		return s
	}
	if blocked {
		return colorize(s, ansiRed, true)
	}
	return colorize(s, ansiGreen, true)
}

func stylePhase(s string, color bool) string {
	return colorize(s, ansiCyan, color)
}
