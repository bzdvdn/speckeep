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
	planPath, _ := featurepaths.ResolvePlan(specsDir, slug)
	tasksPath, _ := featurepaths.ResolveTasks(specsDir, slug)
	dataModelPath, _ := featurepaths.ResolveDataModel(specsDir, slug)
	researchPath, _ := featurepaths.ResolveResearch(specsDir, slug)
	verifyPath, _ := featurepaths.ResolveVerify(specsDir, slug)
	challengePath := filepath.Join(featurepaths.SpecDir(specsDir, slug), "challenge.md")
	if _, err := os.Stat(challengePath); err != nil {
		challengePath = filepath.Join(featurepaths.PlanDir(specsDir, slug), "challenge.md")
	}

	artifacts := []struct {
		path    string
		heading string
	}{
		{specPath, "Spec"},
		{inspectPath, "Inspect Report"},
		{summaryPath, "Spec Summary"},
		{hotfixPath, "Hotfix"},
		{planPath, "Plan"},
		{tasksPath, "Tasks"},
		{dataModelPath, "Data Model"},
		{researchPath, "Research"},
		{challengePath, "Challenge Report"},
		{verifyPath, "Verify Report"},
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
