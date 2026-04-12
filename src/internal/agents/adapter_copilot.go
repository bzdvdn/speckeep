package agents

import (
	"fmt"
	"path/filepath"
)

type copilotAdapter struct{}

func (copilotAdapter) Target() string { return "copilot" }

func (copilotAdapter) Render(commands []CommandDefinition, language string) ([]File, error) {
	lang := normalizeLanguage(language)
	files := make([]File, 0, len(commands))
	for _, command := range commands {
		files = append(files, File{
			Path:    filepath.ToSlash(filepath.Join(".github", "prompts", fmt.Sprintf("speckeep-%s.prompt.md", command.Name))),
			Content: renderCopilot(command, lang),
			Mode:    0o644,
		})
	}
	return files, nil
}

func (copilotAdapter) Paths(commands []CommandDefinition, language string) ([]string, error) {
	files, err := copilotAdapter{}.Render(commands, language)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	return paths, nil
}

func renderCopilot(spec CommandDefinition, lang string) string {
	if lang == "ru" {
		return fmt.Sprintf(`# SpecKeep %s

Используйте %q как основной workflow prompt.

%s

%s

Что нужно сделать:
- обработать запрос пользователя для одной фазы %q
- применять только минимально нужный контекст репозитория
- %s
%s
%s
- %s
- кратко сообщить о результатах и блокерах

%s
`, spec.Name, spec.PromptPath, commandHint(spec.Name, lang), workflowChainHint(lang), spec.Name, scriptExecutionHint(lang), specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang), helpDiscoveryHint(lang), antiPatternHint(lang))
	}

	return fmt.Sprintf(`# SpecKeep %s

Use %q as the primary workflow prompt.

%s

%s

What to do:
- handle the user request for the %q phase
- use only the minimum repository context required
- %s
%s
%s
- %s
- report outcomes and blockers briefly

%s
`, spec.Name, spec.PromptPath, commandHint(spec.Name, lang), workflowChainHint(lang), spec.Name, scriptExecutionHint(lang), specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang), helpDiscoveryHint(lang), antiPatternHint(lang))
}
