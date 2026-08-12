package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"speckeep/src/internal/trace"
)

func newTraceCmd() *cobra.Command {
	var jsonOutput bool
	var testsOnly bool

	cmd := &cobra.Command{
		Use:   "trace [slug] [path]",
		Short: "Trace requirements and tasks to Proof entries in tasks.md",
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

			var entries []trace.Entry
			var err error
			if slug != "" {
				entries, err = trace.ParseTasks(context.Background(), root, slug)
			} else {
				entries, err = trace.ParseAll(context.Background(), root)
			}
			if err != nil {
				return err
			}

			if testsOnly {
				var filtered []trace.Entry
				for _, e := range entries {
					if e.Kind == "test" {
						filtered = append(filtered, e)
					}
				}
				entries = filtered
			}

			if jsonOutput {
				payload, err := json.MarshalIndent(entries, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
				return nil
			}

			if len(entries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No Proof entries found in tasks.md files.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Found %d Proof entries:\n\n", len(entries))
			for _, e := range entries {
				anchorPart := ""
				if e.Anchor != "" {
					anchorPart = ":" + e.Anchor
				}
				acPart := ""
				if e.ACID != "" {
					acPart = fmt.Sprintf(" (%s)", e.ACID)
				}
				slugPart := ""
				if slug == "" {
					slugPart = e.Slug + "#"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "- %s%s%s -> %s%s [%s]%s\n", slugPart, e.TaskID, acPart, e.File, anchorPart, e.Kind, traceProblemsSuffix(cmd, root, e))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	cmd.Flags().BoolVar(&testsOnly, "tests", false, "Show only test-kind Proof entries")

	return cmd
}

func traceProblemsSuffix(cmd *cobra.Command, root string, e trace.Entry) string {
	if !trace.FileExists(root, e) {
		return " (file missing)"
	}
	if e.Anchor == "" {
		return ""
	}
	for _, problem := range trace.CheckEntry(root, e) {
		if problem.Kind == "anchor-missing" {
			return " (anchor missing)"
		}
	}
	return ""
}
