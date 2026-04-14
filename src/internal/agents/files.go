package agents

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type File struct {
	Path    string
	Content string
	Mode    os.FileMode
}

func NormalizeTargets(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}

	seen := map[string]struct{}{}
	var out []string
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			target := strings.ToLower(strings.TrimSpace(part))
			if target == "" {
				continue
			}
			if target == "all" {
				for _, candidate := range SupportedTargets() {
					if _, ok := seen[candidate]; ok {
						continue
					}
					seen[candidate] = struct{}{}
					out = append(out, candidate)
				}
				continue
			}
			if _, ok := adapterRegistry[target]; !ok {
				return nil, fmt.Errorf("unsupported agent target %q, expected one of: aider, claude, codex, copilot, cursor, kilocode, roocode, trae, windsurf, all", target)
			}
			if _, ok := seen[target]; ok {
				continue
			}
			seen[target] = struct{}{}
			out = append(out, target)
		}
	}

	sort.Strings(out)
	return out, nil
}

func Files(targets []string, language string, shell string) ([]File, error) {
	normalized, err := NormalizeTargets(targets)
	if err != nil {
		return nil, err
	}

	commands := DefaultCommands(shell)
	var files []File
	for _, target := range normalized {
		adapter, err := adapterForTarget(target)
		if err != nil {
			return nil, err
		}
		targetFiles, err := adapter.Render(commands, language)
		if err != nil {
			return nil, err
		}
		files = append(files, targetFiles...)
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

func FilesForTarget(target, language, shell string) ([]File, error) {
	normalized, err := NormalizeTargets([]string{target})
	if err != nil {
		return nil, err
	}
	if len(normalized) == 0 {
		return nil, nil
	}

	adapter, err := adapterForTarget(normalized[0])
	if err != nil {
		return nil, err
	}
	return adapter.Render(DefaultCommands(shell), language)
}

func PathsForTarget(target string) ([]string, error) {
	adapter, err := adapterForTarget(target)
	if err != nil {
		return nil, err
	}
	paths, err := adapter.Paths(DefaultCommands("sh"), "en")
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

// commandSpec and commandSpecs are kept as compatibility shims while tests and
// callers continue to use the previous names.
type commandSpec = CommandDefinition

func commandSpecs(shell string) []commandSpec {
	return DefaultCommands(shell)
}

// render is kept as a narrow compatibility shim for single-command rendering.
func render(target, language string, spec CommandDefinition) (string, string, error) {
	adapter, err := adapterForTarget(target)
	if err != nil {
		return "", "", err
	}
	files, err := adapter.Render([]CommandDefinition{spec}, language)
	if err != nil {
		return "", "", err
	}
	if len(files) != 1 {
		return "", "", fmt.Errorf("expected one rendered file for target %q, got %d", target, len(files))
	}
	return files[0].Path, files[0].Content, nil
}

func normalizeLanguage(language string) string {
	lang := strings.ToLower(strings.TrimSpace(language))
	if lang == "ru" {
		return "ru"
	}
	return "en"
}

func normalizeShell(shell string) string {
	if strings.EqualFold(strings.TrimSpace(shell), "powershell") {
		return "powershell"
	}
	return "sh"
}

func scriptPath(name, shell string) string {
	ext := ".sh"
	if shell == "powershell" {
		ext = ".ps1"
	}
	return "./.speckeep/scripts/" + name + ext
}

func commandHint(name, lang string) string {
	if lang == "ru" {
		return fmt.Sprintf("Команда: `/speckeep.%s [request]`", name)
	}
	return fmt.Sprintf("Command: `/speckeep.%s [request]`", name)
}

func toolInvocationHint(lang string) string {
	if lang == "ru" {
		return "Используйте инструменты напрямую через runtime агента; не печатайте raw JSON/XML/tool-call payloads и не выводите внутренние рассуждения о выборе инструмента."
	}
	return "Use tools directly through the agent runtime; do not print raw JSON/XML/tool-call payloads or expose internal reasoning about tool choice."
}

func helpDiscoveryHint(lang string) string {
	if lang == "ru" {
		return "Не запускайте `speckeep ... --help`/`speckeep help` для «разведки»; вместо этого опирайтесь на prompt-файл и readiness scripts."
	}
	return "Do not run `speckeep ... --help`/`speckeep help` for discovery; rely on the prompt file and readiness scripts instead."
}

func specBranchFirstBullet(commandName, lang string) string {
	if commandName != "spec" {
		return ""
	}
	if lang == "ru" {
		return "- Для `/speckeep.spec`: до записи любого файла обязательно переключиться/создать feature-ветку `feature/<slug>` (или явное значение `--branch`). Если git недоступен или вы в detached HEAD — остановитесь и сообщите причину."
	}
	return "- For `/speckeep.spec`: before writing any file, you must switch/create the feature branch `feature/<slug>` (or the explicit `--branch` value). If git is unavailable or you are in detached HEAD, stop and report the reason."
}

func titleCase(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func workflowChainHint(lang string) string {
	if lang == "ru" {
		return "Цепочка workflow: constitution → spec → inspect → plan → tasks → implement → verify → archive. Не пропускайте фазы и не забегайте вперёд."
	}
	return "Workflow chain: constitution → spec → inspect → plan → tasks → implement → verify → archive. Do not skip phases or jump ahead."
}

func antiPatternHint(lang string) string {
	if lang == "ru" {
		return `Запрещено:
- пропускать readiness scripts и переходить к фазе напрямую
- читать или анализировать исходный код scripts
- перепланировать или редизайнить во время implement
- отмечать таск завершённым без observable proof
- выполнять git commit/git push/git tag или создавать PR без явной просьбы пользователя (коммиты — ответственность разработчика)
- читать весь репозиторий, когда промпт говорит "минимальный контекст"`
	}
	return `Do not:
- skip readiness scripts and proceed to the phase directly
- read or inspect script source code
- re-plan or re-design during implement
- mark a task as done without observable proof
- run git commit/git push/git tag or open a PR unless the user explicitly asks (commits are the developer's responsibility)
- read the full repository when the prompt says "minimum context"`
}

func scriptExecutionHint(lang string) string {
	if lang == "ru" {
		return "Когда для фазы есть связанные scripts — выполняйте их как shell-команды (например `bash ./path/to/script.sh`). Доверяйте stdout и exit-коду скрипта. Не читайте, не анализируйте и не модифицируйте исходный код скриптов. Если скрипт завершился с ошибкой (exit code ≠ 0), сообщите пользователю вывод ошибки и остановитесь."
	}
	return "When related scripts are listed for a phase, execute them as shell commands (e.g. `bash ./path/to/script.sh`). Trust the script stdout and exit code as-is. Do not read, inspect, or modify the script source. If a script exits with a non-zero code, report the error output to the user and stop."
}

func windsurfWorkspaceHint(lang string) string {
	if lang == "ru" {
		return "Примечание (Windsurf): убедитесь, что hidden/dotfiles индексируются и видны (папка `.speckeep/`). Перед запуском scripts работайте из корня репозитория (где лежит `.speckeep/`): `cd \"$(git rev-parse --show-toplevel 2>/dev/null || pwd)\"`."
	}
	return "Note (Windsurf): ensure hidden/dotfiles are indexed and visible (the `.speckeep/` folder). Before running scripts, work from the repo root (where `.speckeep/` lives): `cd \"$(git rev-parse --show-toplevel 2>/dev/null || pwd)\"`."
}

func scriptListBlock(items []string, lang string) string {
	if len(items) == 0 {
		return ""
	}
	header := "Scripts to execute:"
	if lang == "ru" {
		header = "Scripts для выполнения (запускать через shell):"
	}
	lines := []string{"- " + header}
	for _, item := range items {
		display := item
		switch {
		case strings.Contains(item, "check-spec-ready"):
			display = item + " <slug>"
		case strings.Contains(item, "check-inspect-ready"):
			display = item + " <slug>"
		case strings.Contains(item, "check-plan-ready"):
			display = item + " <slug>"
		case strings.Contains(item, "check-tasks-ready"):
			display = item + " <slug>"
		case strings.Contains(item, "check-implement-ready"):
			display = item + " <slug>"
		case strings.Contains(item, "check-verify-ready"):
			display = item + " <slug>"
		case strings.Contains(item, "check-archive-ready"):
			display = item + " <slug> completed"
		case strings.Contains(item, "verify-task-state"):
			display = item + " <slug>"
		case strings.Contains(item, "list-open-tasks"):
			display = item + " <slug>"
		}
		lines = append(lines, fmt.Sprintf("  - `%s`", display))
	}
	return strings.Join(lines, "\n")
}
