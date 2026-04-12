package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"speckeep/src/internal/doctor"
	"speckeep/src/internal/workflow"
)

func newDoctorCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "doctor [path]",
		Short: "Check SpecKeep workspace health and agent target consistency",
		Long:  "Runs a health check over the SpecKeep workspace structure, required files, and feature lifecycle readiness.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}
			result, err := doctor.Check(root)
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
			errorCount, warningCount, okCount := doctorFindingCounts(result.Findings)
			printPanel(w, "speckeep doctor", []string{
				fmt.Sprintf("errors: %d", errorCount),
				fmt.Sprintf("warnings: %d", warningCount),
				fmt.Sprintf("ok: %d", okCount),
			})

			fmt.Fprintf(w, "summary: %d error(s), %d warning(s), %d ok\n", errorCount, warningCount, okCount)
			for _, group := range []string{"error", "warning", "ok"} {
				lines := doctorLinesForLevel(result.Findings, group)
				if len(lines) == 0 {
					continue
				}
				fmt.Fprintf(w, "%ss:\n", group)
				for _, line := range lines {
					switch group {
					case "error":
						fmt.Fprintf(w, "- %s\n", styleError(w, line))
					case "warning":
						fmt.Fprintf(w, "- %s\n", styleWarn(w, line))
					default:
						fmt.Fprintf(w, "- %s\n", styleOK(w, line))
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output doctor findings as JSON")
	return cmd
}

func doctorFindingCounts(findings []doctor.Finding) (errors int, warnings int, oks int) {
	for _, finding := range findings {
		switch finding.Level {
		case "error":
			errors++
		case "warning":
			warnings++
		case "ok":
			oks++
		}
	}
	return errors, warnings, oks
}

func doctorLinesForLevel(findings []doctor.Finding, level string) []string {
	lines := make([]string, 0, len(findings))
	for _, finding := range findings {
		if finding.Level == level {
			lines = append(lines, renderDoctorFinding(finding))
		}
	}
	sort.Strings(lines)
	return lines
}

func renderDoctorFinding(finding doctor.Finding) string {
	slug := workflow.FindingSlug(workflow.Finding{Message: finding.Message})
	if slug == "" {
		return "[workspace] " + finding.Message
	}
	return fmt.Sprintf("[%s] %s", slug, displayFeatureFinding(slug, finding.Message))
}
