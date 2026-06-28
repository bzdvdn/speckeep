package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"speckeep/src/internal/config"
)

func docsSections(language string) docSections {
	if strings.EqualFold(strings.TrimSpace(language), "ru") {
		return docSections{
			Goal:         "Цель",
			Context:      "Контекст",
			Requirements: "Требования",
			Acceptance:   "Критерии приемки",
			Questions:    "Открытые вопросы",
			Coverage:     "Покрытие критериев приемки",
			Assumptions:  "Допущения",
		}
	}
	return docSections{
		Goal:         "Goal",
		Context:      "Context",
		Requirements: "Requirements",
		Acceptance:   "Acceptance Criteria",
		Questions:    "Open Questions",
		Coverage:     "Acceptance Coverage",
		Assumptions:  "Assumptions",
	}
}

func constitutionSections(language string) []string {
	if strings.EqualFold(strings.TrimSpace(language), "ru") {
		return []string{
			"Назначение",
			"Ключевые принципы",
			"Ограничения",
			"Языковая политика",
			"Процесс разработки",
			"Управление",
			"Последнее обновление",
		}
	}
	return []string{
		"Purpose",
		"Core Principles",
		"Constraints",
		"Language Policy",
		"Development Workflow",
		"Governance",
		"Last Updated",
	}
}

func CheckConstitution(ctx context.Context, cfg config.Config, root, constitutionPath string) (CheckResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return CheckResult{}, err
	}
	displayPath := constitutionPath
	if strings.TrimSpace(displayPath) == "" {
		displayPath = cfg.Project.ConstitutionFile
	}
	displayPath, absolutePath := resolveUserPath(root, displayPath)
	result := CheckResult{}
	if !fileExists(absolutePath) {
		result.AddError(fmt.Sprintf("constitution file not found: %s", displayPath))
		result.Failed = false
		return result, nil
	}
	content, err := os.ReadFile(absolutePath)
	if err != nil {
		return CheckResult{}, fmt.Errorf("read constitution %s: %w", displayPath, err)
	}
	sections := constitutionSections(cfg.Language.Docs)
	for _, section := range sections {
		if hasHeading(string(content), section) {
			result.AddOK(section)
		} else {
			result.AddError(fmt.Sprintf("missing section: %s", section))
		}
	}
	principlesCount := countMatchingLines(string(content), `(?m)^### `)
	if principlesCount >= 5 {
		result.AddOK(fmt.Sprintf("principles count is %d", principlesCount))
	} else {
		result.AddError(fmt.Sprintf("expected at least 5 principles, found %d", principlesCount))
	}
	if placeholderPattern.Match(content) {
		result.AddWarn("placeholder tokens remain in constitution")
	} else {
		result.AddOK("no placeholder tokens detected")
	}
	result.Failed = false
	return result, nil
}

func CheckConstitutionReady(ctx context.Context, cfg config.Config, root string) (CheckResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return CheckResult{}, err
	}
	result := CheckResult{}
	configPath, err := cfg.ConfigPath(root)
	if err == nil {
		checkFile(&result, ".speckeep/speckeep.yaml", configPath)
	}
	templateDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.Constitution)
	promptDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.ConstitutionPrompt)
	agentsDisplay := cfg.Agents.AgentsFile
	checkFile(&result, templateDisplay, absFromRoot(root, templateDisplay))
	checkFile(&result, promptDisplay, absFromRoot(root, promptDisplay))
	if cfg.Agents.UpdateAgentsMD {
		checkFile(&result, agentsDisplay, absFromRoot(root, agentsDisplay))
	}
	return result, nil
}
