package agents

import (
	"fmt"
	"path/filepath"
)

type roocodeAdapter struct{}

func (roocodeAdapter) Target() string { return "roocode" }

func (roocodeAdapter) Render(commands []CommandDefinition, language string) ([]File, error) {
	lang := normalizeLanguage(language)
	files := make([]File, 0, len(commands))
	for _, command := range commands {
		files = append(files, File{
			Path:    filepath.ToSlash(filepath.Join(".roo", "rules", fmt.Sprintf("speckeep-%s.md", command.Name))),
			Content: renderRoocode(command, lang),
			Mode:    0o644,
		})
	}
	return files, nil
}

func (roocodeAdapter) Paths(commands []CommandDefinition, language string) ([]string, error) {
	files, err := roocodeAdapter{}.Render(commands, language)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	return paths, nil
}

func renderRoocode(spec CommandDefinition, lang string) string {
	if lang == "ru" {
		return fmt.Sprintf(`# SpecKeep %s

Следуйте файлу %q.

%s

Используйте это rule, когда запрос относится к фазе %q.

%s
%s
`, spec.Name, spec.PromptPath, commandHint(spec.Name, lang), spec.Name, specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang))
	}

	return fmt.Sprintf(`# SpecKeep %s

Follow %q.

%s

Use this rule when the request maps to the %q phase.

%s
%s
`, spec.Name, spec.PromptPath, commandHint(spec.Name, lang), spec.Name, specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang))
}
