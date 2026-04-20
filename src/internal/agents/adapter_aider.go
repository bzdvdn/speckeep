package agents

import (
	"fmt"
	"strings"
)

type aiderAdapter struct{}

func (aiderAdapter) Target() string { return "aider" }

func (aiderAdapter) Render(commands []CommandDefinition, language string) ([]File, error) {
	return []File{{
		Path:    ".aider/CONVENTIONS.md",
		Content: renderAiderCommands(commands, language),
		Mode:    0o644,
	}}, nil
}

func (aiderAdapter) Paths(commands []CommandDefinition, language string) ([]string, error) {
	return []string{".aider/CONVENTIONS.md"}, nil
}

func renderAiderCommands(commands []CommandDefinition, language string) string {
	lang := normalizeLanguage(language)
	if lang == "ru" {
		var sections []string
		sections = append(sections, "# SpecKeep Conventions")
		sections = append(sections, "")
		sections = append(sections, "Используйте `.speckeep/` как основной источник контекста. Для каждой фазы открывайте prompt в `.speckeep/templates/prompts/<phase>.md` и следуйте ему.")
		sections = append(sections, "")
		sections = append(sections, "Загружайте этот файл через `--read .aider/CONVENTIONS.md` или добавьте `read: .aider/CONVENTIONS.md` в `.aider.conf.yml`.")
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
	sections = append(sections, "# SpecKeep Conventions")
	sections = append(sections, "")
	sections = append(sections, "Use `.speckeep/` as the primary context. For each phase, open the prompt in `.speckeep/templates/prompts/<phase>.md` and follow it.")
	sections = append(sections, "")
	sections = append(sections, "Load this file via `--read .aider/CONVENTIONS.md` or add `read: .aider/CONVENTIONS.md` to `.aider.conf.yml`.")
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
