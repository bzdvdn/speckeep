package agents

import (
	"fmt"
	"path/filepath"
)

type windsurfAdapter struct{}

func (windsurfAdapter) Target() string { return "windsurf" }

func (windsurfAdapter) Render(commands []CommandDefinition, language string) ([]File, error) {
	lang := normalizeLanguage(language)
	files := make([]File, 0, len(commands))
	for _, command := range commands {
		files = append(files, File{
			Path:    filepath.ToSlash(filepath.Join(".windsurf", "workflows", fmt.Sprintf("speckeep.%s.md", command.Name))),
			Content: renderWindsurf(command, lang),
			Mode:    0o644,
		})
	}
	return files, nil
}

func (windsurfAdapter) Paths(commands []CommandDefinition, language string) ([]string, error) {
	files, err := windsurfAdapter{}.Render(commands, language)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	return paths, nil
}

func renderWindsurf(spec CommandDefinition, lang string) string {
	if lang == "ru" {
		return fmt.Sprintf(`---
trigger: manual
---

Сначала откройте и прочитайте файл %q (обязательно). Затем строго следуйте его разделу "Output expectations", включая финальную строку `+"`Готово к: ...`"+` (без замены на «Следующий шаг», без альтернативных фаз, без выдуманных флагов).

%s

%s

Используйте этот workflow, когда запрос явно относится к фазе %q или команде /speckeep.%s.

%s

- %s
- %s
%s
%s

%s
`, spec.PromptPath, commandHint(spec.Name, lang), workflowChainHint(lang), spec.Name, spec.Name, scriptExecutionHint(lang), windsurfWorkspaceHint(lang), helpDiscoveryHint(lang), specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang), antiPatternHint(lang))
	}

	return fmt.Sprintf(`---
trigger: manual
---

First, open and read %q (mandatory). Then follow its "Output expectations" section strictly, including the final `+"`Ready for: ...`"+` line (do not replace it with "Next step", do not jump phases, do not invent flags).

%s

%s

Use this workflow when the request clearly maps to the %q phase or the /speckeep.%s command.

%s

- %s
- %s
%s
%s

%s
`, spec.PromptPath, commandHint(spec.Name, lang), workflowChainHint(lang), spec.Name, spec.Name, scriptExecutionHint(lang), windsurfWorkspaceHint(lang), helpDiscoveryHint(lang), specBranchFirstBullet(spec.Name, lang), scriptListBlock(spec.Extras, lang), antiPatternHint(lang))
}
