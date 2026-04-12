package agents

import (
	"fmt"
	"path/filepath"
)

type claudeAdapter struct{}

func (claudeAdapter) Target() string { return "claude" }

func (claudeAdapter) Render(commands []CommandDefinition, language string) ([]File, error) {
	lang := normalizeLanguage(language)
	files := make([]File, 0, len(commands))
	for _, command := range commands {
		files = append(files, File{
			Path:    filepath.ToSlash(filepath.Join(".claude", "commands", fmt.Sprintf("speckeep.%s.md", command.Name))),
			Content: renderClaude(command, lang),
			Mode:    0o644,
		})
	}
	return files, nil
}

func (claudeAdapter) Paths(commands []CommandDefinition, language string) ([]string, error) {
	files, err := claudeAdapter{}.Render(commands, language)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	return paths, nil
}

func renderClaude(spec CommandDefinition, lang string) string {
	if lang == "ru" {
		return fmt.Sprintf(`---
description: %s
argument-hint: [request]
---

Следуйте файлу %q.

%s

%s

Аргументы пользователя:
{{arguments}}

Требования:
- сначала прочитайте .speckeep/constitution.md, если это требуется prompt-файлом
- используйте только минимально нужный контекст репозитория
- %s
%s
%s
- %s
- обновляйте только релевантные артефакты и кратко сообщайте об итогах и блокерах

%s
`, spec.Description, spec.PromptPath, commandHint(spec.Name, lang), workflowChainHint(lang), scriptExecutionHint(lang), specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang), helpDiscoveryHint(lang), antiPatternHint(lang))
	}

	return fmt.Sprintf(`---
description: %s
argument-hint: [request]
---

Follow %q.

%s

%s

User arguments:
{{arguments}}

Requirements:
- read .speckeep/constitution.md first when the prompt requires it
- use only the minimum repository context needed
- %s
%s
%s
- %s
- update only the relevant artifacts and report outcomes and blockers briefly

%s
`, spec.Description, spec.PromptPath, commandHint(spec.Name, lang), workflowChainHint(lang), scriptExecutionHint(lang), specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang), helpDiscoveryHint(lang), antiPatternHint(lang))
}
