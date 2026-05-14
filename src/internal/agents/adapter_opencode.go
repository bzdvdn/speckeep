package agents

import (
	"fmt"
	"path/filepath"
)

type opencodeAdapter struct{}

func (opencodeAdapter) Target() string { return "opencode" }

func (opencodeAdapter) Render(commands []CommandDefinition, language string) ([]File, error) {
	lang := normalizeLanguage(language)
	files := make([]File, 0, len(commands))
	for _, command := range commands {
		files = append(files, File{
			Path:    filepath.ToSlash(filepath.Join(".opencode", "commands", fmt.Sprintf("speckeep.%s.md", command.Name))),
			Content: renderOpencode(command, lang),
			Mode:    0o644,
		})
	}
	return files, nil
}

func (opencodeAdapter) Paths(commands []CommandDefinition, language string) ([]string, error) {
	files, err := opencodeAdapter{}.Render(commands, language)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	return paths, nil
}

func renderOpencode(spec CommandDefinition, lang string) string {
	if lang == "ru" {
		return fmt.Sprintf(`---
description: %s
argument-hint: [request]
---

Следуйте файлу %q.

%s

Аргументы пользователя:
{{arguments}}

Требования:
- сначала прочитайте project.constitution_file (по умолчанию CONSTITUTION.md), если это требуется prompt-файлом
- %s
- используйте только минимально нужный контекст репозитория
- %s
%s
%s
`, spec.Description, spec.PromptPath, commandHint(spec.Name, lang), constitutionSummaryHint(lang), finalLineHint(lang), specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang))
	}

	return fmt.Sprintf(`---
description: %s
argument-hint: [request]
---

Follow %q.

%s

User arguments:
{{arguments}}

Requirements:
- read project.constitution_file (default: CONSTITUTION.md) first when the prompt requires it
- %s
- use only the minimum repository context needed
- %s
%s
%s
`, spec.Description, spec.PromptPath, commandHint(spec.Name, lang), constitutionSummaryHint(lang), finalLineHint(lang), specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang))
}
