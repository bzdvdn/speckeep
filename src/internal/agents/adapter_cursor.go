package agents

import (
	"fmt"
	"path/filepath"
)

type cursorAdapter struct{}

func (cursorAdapter) Target() string { return "cursor" }

func (cursorAdapter) Render(commands []CommandDefinition, language string) ([]File, error) {
	lang := normalizeLanguage(language)
	files := make([]File, 0, len(commands))
	for _, command := range commands {
		files = append(files, File{
			Path:    filepath.ToSlash(filepath.Join(".cursor", "rules", fmt.Sprintf("speckeep-%s.mdc", command.Name))),
			Content: renderCursor(command, lang),
			Mode:    0o644,
		})
	}
	return files, nil
}

func (cursorAdapter) Paths(commands []CommandDefinition, language string) ([]string, error) {
	files, err := cursorAdapter{}.Render(commands, language)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	return paths, nil
}

func renderCursor(spec CommandDefinition, lang string) string {
	if lang == "ru" {
		return fmt.Sprintf(`---
description: SpecKeep %s workflow
alwaysApply: false
---

Следуйте файлу %q.

%s

%s

Используйте эту rule, когда запрос явно относится к фазе %q или к команде /speckeep.%s.

%s

- %s
%s
%s

%s
`, spec.Name, spec.PromptPath, commandHint(spec.Name, lang), workflowChainHint(lang), spec.Name, spec.Name, scriptExecutionHint(lang), helpDiscoveryHint(lang), specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang), antiPatternHint(lang))
	}

	return fmt.Sprintf(`---
description: SpecKeep %s workflow
alwaysApply: false
---

Follow %q.

%s

%s

Use this rule when the request clearly maps to the %q phase or the /speckeep.%s command.

%s

- %s
%s
%s

%s
`, spec.Name, spec.PromptPath, commandHint(spec.Name, lang), workflowChainHint(lang), spec.Name, spec.Name, scriptExecutionHint(lang), helpDiscoveryHint(lang), specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang), antiPatternHint(lang))
}
