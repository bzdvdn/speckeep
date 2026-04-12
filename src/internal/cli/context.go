package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"speckeep/src/internal/config"
	"speckeep/src/internal/workflow"
)

func newContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context <slug> [path]",
		Short: "Show context budget for a feature",
		Long: `Display how many tokens each phase loads for a feature.
Helps teams optimize context usage and identify phases that consume too much.

Examples:
  speckeep context my-feature
  speckeep context my-feature --json
`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 2 {
				root = args[1]
			}
			slug := args[0]

			cfg, err := config.Load(root)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			state, err := workflow.State(root, slug)
			if err != nil {
				return fmt.Errorf("get feature state: %w", err)
			}

			contextInfo := calculateContextBudget(root, slug, state, cfg)

			fmt.Fprintf(cmd.OutOrStdout(), "Context Budget for: %s\n", slug)
			fmt.Fprintf(cmd.OutOrStdout(), "Phase: %s\n\n", state.Phase)

			totalTokens := 0
			for _, phase := range contextInfo.Phases {
				tokens := phase.EstimatedTokens
				status := "✓"
				if !phase.Exists {
					status = " "
					tokens = 0
				}
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %-20s %6d tokens  %s\n", status, phase.Name, tokens, phase.Path)
				totalTokens += tokens
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\nTotal active context: %d tokens\n", totalTokens)
			fmt.Fprintf(cmd.OutOrStdout(), "Recommended limit: 50000 tokens (OpenSpec-style)\n")

			if totalTokens > 50000 {
				fmt.Fprintf(cmd.OutOrStdout(), "\n⚠️  WARNING: Context exceeds 50KB limit\n")
			}

			return nil
		},
	}

	return cmd
}

type PhaseContext struct {
	Name            string
	Path            string
	Exists          bool
	EstimatedTokens int
	Description     string
}

type ContextBudget struct {
	Phases []PhaseContext
}

func calculateContextBudget(root, slug string, state workflow.FeatureState, cfg config.Config) ContextBudget {
	specsDir, _ := cfg.SpecsDir(root)

	specPath := filepath.Join(specsDir, slug, "spec.md")
	inspectPath := filepath.Join(specsDir, slug, "inspect.md")
	summaryPath := filepath.Join(specsDir, slug, "summary.md")
	planPath := filepath.Join(specsDir, slug, "plan", "plan.md")
	tasksPath := filepath.Join(specsDir, slug, "plan", "tasks.md")
	dataModelPath := filepath.Join(specsDir, slug, "plan", "data-model.md")

	return ContextBudget{
		Phases: []PhaseContext{
			{
				Name:            "Spec",
				Path:            relPath(root, specPath),
				Exists:          state.SpecExists,
				EstimatedTokens: estimateTokens(specPath),
				Description:     "Feature requirements and acceptance criteria",
			},
			{
				Name:            "Inspect Report",
				Path:            relPath(root, inspectPath),
				Exists:          state.InspectExists,
				EstimatedTokens: estimateTokens(inspectPath),
				Description:     "Quality gate verification",
			},
			{
				Name:            "Summary",
				Path:            relPath(root, summaryPath),
				Exists:          fileExists(summaryPath),
				EstimatedTokens: estimateTokens(summaryPath),
				Description:     "Compact AC table for implement/verify",
			},
			{
				Name:            "Plan",
				Path:            relPath(root, planPath),
				Exists:          state.PlanExists,
				EstimatedTokens: estimateTokens(planPath),
				Description:     "Implementation decisions and surfaces",
			},
			{
				Name:            "Tasks",
				Path:            relPath(root, tasksPath),
				Exists:          state.TasksExists,
				EstimatedTokens: estimateTokens(tasksPath),
				Description:     "Task breakdown with AC coverage",
			},
			{
				Name:            "Data Model",
				Path:            relPath(root, dataModelPath),
				Exists:          fileExists(dataModelPath),
				EstimatedTokens: estimateTokens(dataModelPath),
				Description:     "Entity definitions",
			},
		},
	}
}

func estimateTokens(path string) int {
	if !fileExists(path) {
		return 0
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	words := len(strings.Fields(string(content)))

	avgTokensPerWord := 1.3
	return int(float64(words) * avgTokensPerWord)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
