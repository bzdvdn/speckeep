package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"speckeep/src/internal/config"
	"speckeep/src/internal/featurepaths"
)

func newExportCmd() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "export <slug> [path]",
		Short: "Bundle all artifacts for one feature into a single markdown document",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 2 {
				root = args[1]
			}

			content, err := exportFeature(root, args[0])
			if err != nil {
				return err
			}

			if outputPath != "" {
				return os.WriteFile(outputPath, []byte(content), 0o644)
			}
			fmt.Fprint(cmd.OutOrStdout(), content)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "write to file instead of stdout")
	return cmd
}

func exportFeature(root, slug string) (string, error) {
	cfg, err := config.Load(root)
	if err != nil {
		return "", err
	}

	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return "", err
	}

	specPath, _ := featurepaths.ResolveSpec(specsDir, slug)
	inspectPath, _ := featurepaths.ResolveInspect(specsDir, slug)
	summaryPath, _ := featurepaths.ResolveSummary(specsDir, slug)
	hotfixPath, _ := featurepaths.ResolveHotfix(specsDir, slug)

	artifacts := []struct {
		path    string
		heading string
	}{
		{specPath, "Spec"},
		{inspectPath, "Inspect Report"},
		{summaryPath, "Spec Summary"},
		{hotfixPath, "Hotfix"},
		{featurepaths.Plan(specsDir, slug), "Plan"},
		{featurepaths.Tasks(specsDir, slug), "Tasks"},
		{featurepaths.DataModel(specsDir, slug), "Data Model"},
		{featurepaths.Research(specsDir, slug), "Research"},
		{filepath.Join(featurepaths.PlanDir(specsDir, slug), "challenge.md"), "Challenge Report"},
		{featurepaths.Verify(specsDir, slug), "Verify Report"},
	}

	var sections []string
	for _, a := range artifacts {
		content, err := os.ReadFile(a.path)
		if err != nil {
			continue
		}
		rel, _ := filepath.Rel(root, a.path)
		sections = append(sections, fmt.Sprintf("<!-- %s: %s -->\n\n%s",
			a.heading, filepath.ToSlash(rel), strings.TrimRight(string(content), "\n")))
	}

	if len(sections) == 0 {
		return "", fmt.Errorf("no artifacts found for slug %q", slug)
	}

	return strings.Join(sections, "\n\n---\n\n") + "\n", nil
}
