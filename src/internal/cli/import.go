package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"speckeep/src/internal/config"
	"speckeep/src/internal/importer"
)

func newImportCmd() *cobra.Command {
	var (
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "import <openspec|speckit> [path]",
		Short: "Import feature packages from OpenSpec or Spec Kit workspaces",
		Long: `Import feature packages from another spec system into the current speckeep
workspace, converting their artifacts to speckeep layout:

  - openspec: reads openspec/changes/<slug>/ (proposal.md, specs/, design.md, tasks.md)
    and rebuilds spec.md (RQ-*/AC-* from Requirement/Scenario blocks), plan.md from
    design.md, and copies tasks.md (best-effort).
  - speckit:   reads specs/<slug>/ and copies spec.md / plan.md / tasks.md.

Existing speckeep feature directories are never overwritten — they are reported
as skipped. Imported tasks.md usually need a /spk.tasks regeneration to add
Touches:, Surface Map, and Acceptance Coverage.`,
		Example: `  speckeep import openspec .
  speckeep import openspec ./existing-openspec-repo
  speckeep import speckit .
  speckeep import openspec . --json`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			source, err := importer.ParseSource(args[0])
			if err != nil {
				return err
			}
			root := "."
			if len(args) == 2 {
				root = args[1]
			}

			cfg, err := config.Load(context.Background(), root)
			if err != nil {
				return err
			}
			specsDir, err := cfg.SpecsDir(root)
			if err != nil {
				return err
			}

			result, err := importer.Import(context.Background(), source, root, specsDir)
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
			printImportResult(cmd, result)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func printImportResult(cmd *cobra.Command, result importer.Result) {
	w := cmd.OutOrStdout()
	if len(result.Imported) == 0 {
		printPanel(w, "speckeep import "+string(result.Source), []string{"no feature packages found to import"})
		return
	}
	printPanel(w, "speckeep import "+string(result.Source), []string{
		fmt.Sprintf("imported: %d feature(s)", len(result.Imported)),
	})
	for _, feat := range result.Imported {
		printPanel(w, feat.Slug, feat.Files)
		if len(feat.Notes) > 0 {
			for _, note := range feat.Notes {
				fmt.Fprintf(w, "  note: %s\n", note)
			}
		}
	}
	if len(result.Skipped) > 0 {
		printPanel(w, "Skipped", result.Skipped)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "tip:     run `/spk.tasks <slug>` to regenerate imported tasks.md with Touches/Surface Map/Acceptance Coverage")
}
