package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"speckeep/src/internal/config"
)

func newRiskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "risk <slug> [path]",
		Short: "Assess feature risk level and suggest spec rigor",
		Long: `Analyze a feature to determine its risk level and recommend spec rigor.

Low risk features can use lightweight specs, while high risk features need full specs.

Examples:
  speckeep risk my-feature
  speckeep risk my-feature --json
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

			risk := assessRisk(root, slug, cfg)

			fmt.Fprintf(cmd.OutOrStdout(), "Risk Assessment: %s\n\n", slug)
			fmt.Fprintf(cmd.OutOrStdout(), "Overall risk level: %s\n\n", risk.Level)

			fmt.Fprintf(cmd.OutOrStdout(), "Risk factors:\n")
			for _, factor := range risk.Factors {
				level := "low"
				if factor.Score > 3 {
					level = "medium"
				}
				if factor.Score > 6 {
					level = "high"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s (score: %d/10)\n", level, factor.Name, factor.Score)
				if factor.Reason != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "         %s\n", factor.Reason)
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\nTotal score: %d/50\n", risk.TotalScore)
			fmt.Fprintf(cmd.OutOrStdout(), "Recommended rigor: %s\n", risk.RecommendedRigor)

			if risk.RecommendedRigor == "full" {
				fmt.Fprintf(cmd.OutOrStdout(), "\nThis feature requires a full spec with inspect gate.\n")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "\nThis feature can use a lightweight spec.\n")
			}

			return nil
		},
	}

	return cmd
}

type RiskFactor struct {
	Name   string
	Score  int
	Reason string
}

type RiskAssessment struct {
	Level            string
	TotalScore       int
	RecommendedRigor string
	Factors          []RiskFactor
}

func assessRisk(root, slug string, cfg config.Config) RiskAssessment {
	specsDir, _ := cfg.SpecsDir(root)

	specPath := filepath.Join(specsDir, slug, "spec.md")

	factors := []RiskFactor{
		assessComplexity(specPath),
		assessSecurityImpact(root, slug),
		assessDataChanges(root, slug),
		assessAPIChanges(root, slug),
		assessDependencies(root, slug),
	}

	totalScore := 0
	for _, f := range factors {
		totalScore += f.Score
	}

	level := "low"
	if totalScore > 15 {
		level = "medium"
	}
	if totalScore > 30 {
		level = "high"
	}

	rigor := "lite"
	if totalScore > 20 {
		rigor = "full"
	}

	return RiskAssessment{
		Level:            level,
		TotalScore:       totalScore,
		RecommendedRigor: rigor,
		Factors:          factors,
	}
}

func assessComplexity(specPath string) RiskFactor {
	if !fileExists(specPath) {
		return RiskFactor{Name: "Complexity", Score: 5, Reason: "Spec not yet created, assuming medium complexity"}
	}

	content, err := os.ReadFile(specPath)
	if err != nil {
		return RiskFactor{Name: "Complexity", Score: 5, Reason: "Cannot read spec"}
	}

	acceptanceCount := strings.Count(string(content), "AC-")
	requirementCount := strings.Count(string(content), "RQ-")

	total := acceptanceCount + requirementCount
	if total == 0 {
		return RiskFactor{Name: "Complexity", Score: 3, Reason: "No acceptance criteria or requirements found"}
	}
	if total > 10 {
		return RiskFactor{Name: "Complexity", Score: 8, Reason: fmt.Sprintf("High complexity: %d criteria/requirements", total)}
	}
	if total > 5 {
		return RiskFactor{Name: "Complexity", Score: 5, Reason: fmt.Sprintf("Medium complexity: %d criteria/requirements", total)}
	}
	return RiskFactor{Name: "Complexity", Score: 2, Reason: fmt.Sprintf("Low complexity: %d criteria/requirements", total)}
}

func assessSecurityImpact(root, slug string) RiskFactor {
	specsDir := filepath.Join(root, ".speckeep", "specs")
	specPath := filepath.Join(specsDir, slug, "spec.md")

	if !fileExists(specPath) {
		return RiskFactor{Name: "Security Impact", Score: 3, Reason: "Spec not yet created"}
	}

	content, err := os.ReadFile(specPath)
	if err != nil {
		return RiskFactor{Name: "Security Impact", Score: 3}
	}

	securityKeywords := []string{"auth", "permission", "token", "password", "secret", "encrypt", "access control", "CORS", "CSRF"}
	found := 0
	for _, keyword := range securityKeywords {
		if strings.Contains(strings.ToLower(string(content)), keyword) {
			found++
		}
	}

	if found > 2 {
		return RiskFactor{Name: "Security Impact", Score: 8, Reason: fmt.Sprintf("Security-related: %d keywords found", found)}
	}
	if found > 0 {
		return RiskFactor{Name: "Security Impact", Score: 5, Reason: fmt.Sprintf("Some security concerns: %d keywords found", found)}
	}
	return RiskFactor{Name: "Security Impact", Score: 1, Reason: "No obvious security concerns"}
}

func assessDataChanges(root, slug string) RiskFactor {
	dataModelPath := filepath.Join(root, ".speckeep", "specs", slug, "plan", "data-model.md")

	if !fileExists(dataModelPath) {
		return RiskFactor{Name: "Data Changes", Score: 3, Reason: "Data model not yet created"}
	}

	content, err := os.ReadFile(dataModelPath)
	if err != nil {
		return RiskFactor{Name: "Data Changes", Score: 3}
	}

	entityCount := strings.Count(string(content), "DM-")
	if entityCount > 5 {
		return RiskFactor{Name: "Data Changes", Score: 8, Reason: fmt.Sprintf("Many entities: %d", entityCount)}
	}
	if entityCount > 0 {
		return RiskFactor{Name: "Data Changes", Score: 4, Reason: fmt.Sprintf("Some entities: %d", entityCount)}
	}
	return RiskFactor{Name: "Data Changes", Score: 1, Reason: "No data model changes"}
}

func assessAPIChanges(root, slug string) RiskFactor {
	contractsDir := filepath.Join(root, ".speckeep", "specs", slug, "plan", "contracts")

	if !dirExists(contractsDir) {
		return RiskFactor{Name: "API Changes", Score: 2, Reason: "No contracts directory"}
	}

	entries, err := os.ReadDir(contractsDir)
	if err != nil {
		return RiskFactor{Name: "API Changes", Score: 2}
	}

	apiCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			apiCount++
		}
	}

	if apiCount > 2 {
		return RiskFactor{Name: "API Changes", Score: 7, Reason: fmt.Sprintf("Multiple API contracts: %d", apiCount)}
	}
	if apiCount > 0 {
		return RiskFactor{Name: "API Changes", Score: 4, Reason: fmt.Sprintf("Some API contracts: %d", apiCount)}
	}
	return RiskFactor{Name: "API Changes", Score: 1, Reason: "No API changes"}
}

func assessDependencies(root, slug string) RiskFactor {
	specsDir := filepath.Join(root, ".speckeep", "specs")
	specPath := filepath.Join(specsDir, slug, "spec.md")

	if !fileExists(specPath) {
		return RiskFactor{Name: "Dependencies", Score: 3, Reason: "Spec not yet created"}
	}

	content, err := os.ReadFile(specPath)
	if err != nil {
		return RiskFactor{Name: "Dependencies", Score: 3}
	}

	depKeywords := []string{"external", "third-party", "integration", "API", "webhook", "service"}
	found := 0
	for _, keyword := range depKeywords {
		if strings.Contains(strings.ToLower(string(content)), keyword) {
			found++
		}
	}

	if found > 2 {
		return RiskFactor{Name: "Dependencies", Score: 7, Reason: fmt.Sprintf("Many external dependencies: %d keywords", found)}
	}
	if found > 0 {
		return RiskFactor{Name: "Dependencies", Score: 4, Reason: fmt.Sprintf("Some external dependencies: %d keywords", found)}
	}
	return RiskFactor{Name: "Dependencies", Score: 1, Reason: "No external dependencies"}
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
