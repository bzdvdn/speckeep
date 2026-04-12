package agents

import (
	"fmt"
	"path/filepath"
)

type kilocodeAdapter struct{}

func (kilocodeAdapter) Target() string { return "kilocode" }

func (kilocodeAdapter) Render(commands []CommandDefinition, language string) ([]File, error) {
	lang := normalizeLanguage(language)
	files := make([]File, 0, len(commands))
	for _, command := range commands {
		files = append(files, File{
			Path:    filepath.ToSlash(filepath.Join(".kilocode", "workflows", fmt.Sprintf("speckeep.%s.md", command.Name))),
			Content: renderKilo(command, lang),
			Mode:    0o644,
		})
	}
	return files, nil
}

func (kilocodeAdapter) Paths(commands []CommandDefinition, language string) ([]string, error) {
	files, err := kilocodeAdapter{}.Render(commands, language)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	return paths, nil
}

func renderKilo(spec CommandDefinition, lang string) string {
	if lang == "ru" {
		return fmt.Sprintf(`# SpecKeep %s

Следуйте файлу %q.

Запуск (Kilo workflow): %q

%s

Используйте этот workflow, когда запрос относится к фазе %q.

%s

- %s
%s
%s

%s
`, spec.Name, spec.PromptPath, "/speckeep."+spec.Name+".md", workflowChainHint(lang), spec.Name, scriptExecutionHint(lang), helpDiscoveryHint(lang), specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang), antiPatternHint(lang))
	}

	return fmt.Sprintf(`# SpecKeep %s

Follow %q.

Trigger (Kilo workflow): %q

%s

Use this workflow when the request maps to the %q phase.

%s

- %s
%s
%s

%s
`, spec.Name, spec.PromptPath, "/speckeep."+spec.Name+".md", workflowChainHint(lang), spec.Name, scriptExecutionHint(lang), helpDiscoveryHint(lang), specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang), antiPatternHint(lang))
}
