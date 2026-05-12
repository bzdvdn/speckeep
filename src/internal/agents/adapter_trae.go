package agents

import (
	"fmt"
	"path/filepath"
)

type traeAdapter struct{}

func (traeAdapter) Target() string { return "trae" }

func (traeAdapter) Render(commands []CommandDefinition, language string) ([]File, error) {
	lang := normalizeLanguage(language)
	files := make([]File, 0, len(commands))
	for _, command := range commands {
		files = append(files, File{
			Path:    filepath.ToSlash(filepath.Join(".trae", "rules", fmt.Sprintf("speckeep.%s.md", command.Name))),
			Content: renderTrae(command, lang),
			Mode:    0o644,
		})
	}
	return files, nil
}

func (traeAdapter) Paths(commands []CommandDefinition, language string) ([]string, error) {
	files, err := traeAdapter{}.Render(commands, language)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	return paths, nil
}

func renderTrae(spec CommandDefinition, lang string) string {
	if lang == "ru" {
		return fmt.Sprintf(`# SpecKeep %s

Следуйте файлу %q.

%s

%s

Используйте это rule, когда запрос явно относится к фазе %q или к команде /speckeep.%s.

Правила:
- %s
- %s
- Минимальный контекст: текущий slug и surfaces из `+"`Touches:`"+`.
%s

%s
%s
`, titleCase(spec.Name), spec.PromptPath, workflowChainHint(lang), commandHint(spec.Name, lang), spec.Name, spec.Name, scriptExecutionHint(lang), helpDiscoveryHint(lang), specBranchFirstBullet(spec.Name, lang), antiPatternHint(lang), scriptListBlock(spec.Extras, lang))
	}

	return fmt.Sprintf(`# SpecKeep %s

Follow %q.

%s

%s

Use this rule when the request clearly maps to the %q phase or the /speckeep.%s command.

Rules:
- %s
- %s
- Minimum context: current slug and surfaces from `+"`Touches:`"+`.
%s

%s
%s
`, titleCase(spec.Name), spec.PromptPath, workflowChainHint(lang), commandHint(spec.Name, lang), spec.Name, spec.Name, scriptExecutionHint(lang), helpDiscoveryHint(lang), specBranchFirstBullet(spec.Name, lang), antiPatternHint(lang), scriptListBlock(spec.Extras, lang))
}
