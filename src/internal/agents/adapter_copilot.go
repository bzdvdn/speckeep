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
%s
`, spec.Name, spec.PromptPath, commandHint(spec.Name, lang), specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang))
	}

	return fmt.Sprintf(`# SpecKeep %s

Use %q as the primary workflow prompt.

%s

%s
%s
`, spec.Name, spec.PromptPath, commandHint(spec.Name, lang), specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang))
}
