package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
	"speckeep/src/internal/config"
	"speckeep/src/internal/trace"
)

func newTraceCmd() *cobra.Command {
	var jsonOutput bool
	var testsOnly bool

	cmd := &cobra.Command{
		Use:   "trace [slug] [path]",
		Short: "Trace requirements and tasks to code annotations",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			slug := ""
			if len(args) > 0 {
				if len(args) == 2 {
					slug = args[0]
					root = args[1]
				} else {
					// Check if first arg is a slug or a path
					if _, err := os.Stat(args[0]); err == nil {
						root = args[0]
					} else {
						slug = args[0]
					}
				}
			}

			traceResult, err := trace.Scan(root)
			if err != nil {
				return err
			}

			if testsOnly {
				var filtered []trace.Finding
				for _, f := range traceResult.Findings {
					if f.Type == "test" {
						filtered = append(filtered, f)
					}
				}
				traceResult.Findings = filtered
			}

			if slug != "" {
				taskIDs, err := getTaskIDsForSlug(root, slug)
				if err != nil {
					return err
				}
				traceResult.Findings = trace.FilterBySlug(traceResult.Findings, taskIDs)
			}

			if jsonOutput {
				payload, err := json.MarshalIndent(traceResult, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
				return nil
			}

			if len(traceResult.Findings) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No traceability annotations found.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Found %d traceability annotations:\n\n", len(traceResult.Findings))
			for _, f := range traceResult.Findings {
				acPart := ""
				if f.ACID != "" {
					acPart = fmt.Sprintf(" (%s)", f.ACID)
				}
				typePart := ""
				if f.Type == "test" {
					typePart = "[TEST] "
				}
				fmt.Fprintf(cmd.OutOrStdout(), "- %s:%d: %s%s%s %s\n", f.File, f.Line, typePart, f.TaskID, acPart, f.Description)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	cmd.Flags().BoolVar(&testsOnly, "tests", false, "Show only test-related annotations (@sk-test)")

	return cmd
}

func getTaskIDsForSlug(root, slug string) (map[string]struct{}, error) {
	cfg, err := config.Load(root)
	if err != nil {
		return nil, err
	}

	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return nil, err
	}

	tasksPath := filepath.Join(specsDir, slug, "plan", "tasks.md")
	content, err := os.ReadFile(tasksPath)
	if err != nil {
		return nil, fmt.Errorf("could not read tasks file for slug %s: %w", slug, err)
	}

	taskIDs := make(map[string]struct{})
	// Match task IDs like T1.1, T2.2, etc.
	re := regexp.MustCompile(`(T[0-9]+(?:\.[0-9]+)*)`)
	matches := re.FindAllString(string(content), -1)
	for _, m := range matches {
		taskIDs[m] = struct{}{}
	}

	return taskIDs, nil
}
