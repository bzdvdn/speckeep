package agents

import (
	"fmt"
	"path/filepath"
)

type codexAdapter struct{}

func (codexAdapter) Target() string { return "codex" }

func (codexAdapter) Render(commands []CommandDefinition, language string) ([]File, error) {
	lang := normalizeLanguage(language)
	files := make([]File, 0, len(commands))
	for _, command := range commands {
		files = append(files, File{
			Path:    filepath.ToSlash(filepath.Join(".codex", "prompts", fmt.Sprintf("speckeep.%s.md", command.Name))),
			Content: renderCodex(command, lang),
			Mode:    0o644,
		})
	}
	return files, nil
}

func (codexAdapter) Paths(commands []CommandDefinition, language string) ([]string, error) {
	files, err := codexAdapter{}.Render(commands, language)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	return paths, nil
}

func renderCodex(spec CommandDefinition, lang string) string {
	title := titleCase(spec.Name)
	if lang == "ru" {
		return fmt.Sprintf(`# SpecKeep %s

Сначала откройте и прочитайте файл %q (обязательно). Затем строго следуйте его разделу "Output expectations", включая финальную строку `+"`Готово к: ...`"+`.

%s

%s

Вход пользователя: {{arguments}}

Дополнительно:
- %s
- %s
- %s
%s
%s

%s
`, title, spec.PromptPath, commandHint(spec.Name, lang), workflowChainHint(lang), scriptExecutionHint(lang), toolInvocationHint(lang), helpDiscoveryHint(lang), specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang), antiPatternHint(lang))
	}

	return fmt.Sprintf(`# SpecKeep %s

First, open and read %q (mandatory). Then follow its "Output expectations" section strictly, including the final `+"`Ready for: ...`"+` line.

%s

%s

User input: {{arguments}}

Additional context:
- %s
- %s
- %s
%s
%s

%s
`, title, spec.PromptPath, commandHint(spec.Name, lang), workflowChainHint(lang), scriptExecutionHint(lang), toolInvocationHint(lang), helpDiscoveryHint(lang), specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang), antiPatternHint(lang))
}
