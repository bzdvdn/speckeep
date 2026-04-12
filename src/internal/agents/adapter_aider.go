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
		sections = append(sections, "Используйте `.speckeep/` как основной источник проектного контекста. Следуйте соответствующим prompt-файлам в `.speckeep/templates/prompts/`.")
		sections = append(sections, "")
		sections = append(sections, "Загружайте этот файл через `--read .aider/CONVENTIONS.md` или добавьте `read: .aider/CONVENTIONS.md` в `.aider.conf.yml`.")
		sections = append(sections, "")
		sections = append(sections, workflowChainHint(lang))
		for _, cmd := range commands {
			sections = append(sections, "")
			sections = append(sections, fmt.Sprintf("## /speckeep.%s", cmd.Name))
			sections = append(sections, fmt.Sprintf("- Основной prompt: %s", cmd.PromptPath))
			sections = append(sections, fmt.Sprintf("- %s", commandHint(cmd.Name, lang)))
			sections = append(sections, "- Используйте только минимально нужный контекст репозитория")
			if hint := specBranchFirstBullet(cmd.Name, lang); hint != "" {
				sections = append(sections, hint)
			}
			sections = append(sections, fmt.Sprintf("- %s", scriptExecutionHint(lang)))
			sections = append(sections, fmt.Sprintf("- %s", helpDiscoveryHint(lang)))
			if len(cmd.Extras) > 0 {
				sections = append(sections, scriptListBlock(cmd.Extras, lang))
			}
		}
		sections = append(sections, "")
		sections = append(sections, antiPatternHint(lang))
		return strings.Join(sections, "\n") + "\n"
	}

	var sections []string
	sections = append(sections, "# SpecKeep Conventions")
	sections = append(sections, "")
	sections = append(sections, "Use `.speckeep/` as the primary source of project context. Follow the matching prompt files under `.speckeep/templates/prompts/`.")
	sections = append(sections, "")
	sections = append(sections, "Load this file via `--read .aider/CONVENTIONS.md` or add `read: .aider/CONVENTIONS.md` to `.aider.conf.yml`.")
	sections = append(sections, "")
	sections = append(sections, workflowChainHint(lang))
	for _, cmd := range commands {
		sections = append(sections, "")
		sections = append(sections, fmt.Sprintf("## /speckeep.%s", cmd.Name))
		sections = append(sections, fmt.Sprintf("- Primary prompt: %s", cmd.PromptPath))
		sections = append(sections, fmt.Sprintf("- %s", commandHint(cmd.Name, lang)))
		sections = append(sections, "- Use only the minimum repository context required")
		if hint := specBranchFirstBullet(cmd.Name, lang); hint != "" {
			sections = append(sections, hint)
		}
		sections = append(sections, fmt.Sprintf("- %s", scriptExecutionHint(lang)))
		sections = append(sections, fmt.Sprintf("- %s", helpDiscoveryHint(lang)))
		if len(cmd.Extras) > 0 {
			sections = append(sections, scriptListBlock(cmd.Extras, lang))
		}
	}
	sections = append(sections, "")
	sections = append(sections, antiPatternHint(lang))
	return strings.Join(sections, "\n") + "\n"
}
