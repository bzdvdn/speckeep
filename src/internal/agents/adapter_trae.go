package agents

import (
	"fmt"
	"strings"
)

type traeAdapter struct{}

func (traeAdapter) Target() string { return "trae" }

func (traeAdapter) Render(commands []CommandDefinition, language string) ([]File, error) {
	return []File{{
		Path:    ".trae/project_rules.md",
		Content: renderTraeCommands(commands, language),
		Mode:    0o644,
	}}, nil
}

func (traeAdapter) Paths(commands []CommandDefinition, language string) ([]string, error) {
	return []string{".trae/project_rules.md"}, nil
}

func renderTrae(language, shell string) string {
	return renderTraeCommands(DefaultCommands(shell), language)
}

func renderTraeCommands(commands []CommandDefinition, language string) string {
	lang := normalizeLanguage(language)
	if lang == "ru" {
		var sections []string
		sections = append(sections, "# SpecKeep Project Rules")
		sections = append(sections, "")
		sections = append(sections, "Используйте `.speckeep/` как основной источник контекста. Для каждой фазы открывайте prompt в `.speckeep/templates/prompts/<phase>.md` и следуйте ему.")
		sections = append(sections, "")
		sections = append(sections, workflowChainHint(lang))
		sections = append(sections, "")
		sections = append(sections, "Команды:")
		for _, cmd := range commands {
			sections = append(sections, fmt.Sprintf("- `/speckeep.%s` → %s", cmd.Name, cmd.PromptPath))
		}
		sections = append(sections, "")
		sections = append(sections, "Правила:")
		sections = append(sections, "- Минимальный контекст: текущий slug и surfaces из `Touches:`.")
		sections = append(sections, "- "+scriptExecutionHint(lang))
		sections = append(sections, "- "+helpDiscoveryHint(lang))
		if hint := specBranchFirstBullet("spec", lang); hint != "" {
			sections = append(sections, hint)
		}
		sections = append(sections, "")
		sections = append(sections, antiPatternHint(lang))
		return strings.Join(sections, "\n") + "\n"
	}

	var sections []string
	sections = append(sections, "# SpecKeep Project Rules")
	sections = append(sections, "")
	sections = append(sections, "Use `.speckeep/` as the primary context. For each phase, open the prompt in `.speckeep/templates/prompts/<phase>.md` and follow it.")
	sections = append(sections, "")
	sections = append(sections, workflowChainHint(lang))
	sections = append(sections, "")
	sections = append(sections, "Commands:")
	for _, cmd := range commands {
		sections = append(sections, fmt.Sprintf("- `/speckeep.%s` → %s", cmd.Name, cmd.PromptPath))
	}
	sections = append(sections, "")
	sections = append(sections, "Rules:")
	sections = append(sections, "- Minimum context: current slug and surfaces from `Touches:`.")
	sections = append(sections, "- "+scriptExecutionHint(lang))
	sections = append(sections, "- "+helpDiscoveryHint(lang))
	if hint := specBranchFirstBullet("spec", lang); hint != "" {
		sections = append(sections, hint)
	}
	sections = append(sections, "")
	sections = append(sections, antiPatternHint(lang))
	return strings.Join(sections, "\n") + "\n"
}
