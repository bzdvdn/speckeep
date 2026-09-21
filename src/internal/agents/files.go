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

// NormalizeTargets validates and deduplicates agent target names.
//
// Deprecated targets (see deprecatedTargets) are silently dropped rather
// than rejected: they may still be sitting in an existing project's
// speckeep.yaml, and a discontinued target should quietly stop being
// generated for on the next refresh, not hard-fail every command that
// touches that project. speckeep doctor separately surfaces their now-stale
// generated files so the drop stays discoverable.
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
			if _, ok := deprecatedTargets[target]; ok {
				continue
			}
			if _, ok := targetSkillDirs[target]; !ok {
				return nil, fmt.Errorf("unsupported agent target %q, expected one of: %s, all", target, TargetOptionsText())
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

// Files returns all skill-pack files for the given targets.
func Files(targets []string, language string, shell string) ([]File, error) {
	normalized, err := NormalizeTargets(targets)
	if err != nil {
		return nil, err
	}
	commands := DefaultCommands(shell)
	var files []File
	for _, target := range normalized {
		targetFiles, err := skillPackFiles(target, language, shell, commands)
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
	return skillPackFiles(normalized[0], language, shell, DefaultCommands(shell))
}

func PathsForTarget(target string) ([]string, error) {
	if _, ok := targetSkillDirs[target]; !ok {
		return nil, fmt.Errorf("unsupported agent target %q: %w", target, ErrUnsupportedTarget)
	}
	paths := skillPackPaths(target, DefaultCommands("sh"))
	sort.Strings(paths)
	return paths, nil
}

// TargetOptionsText returns the comma-joined target list used in CLI flag help.
func TargetOptionsText() string {
	return strings.Join(SupportedTargets(), ", ")
}

// SupportedTargets lists every agent target that can receive the sdd skill pack.
func SupportedTargets() []string {
	targets := make([]string, 0, len(targetSkillDirs))
	for target := range targetSkillDirs {
		targets = append(targets, target)
	}
	sort.Strings(targets)
	return targets
}

// LegacyArchivePaths are leftover /speckeep.archive wrappers removed on refresh.
func LegacyArchivePaths() []string {
	return []string{
		".claude/commands/speckeep.archive.md",
		".codex/prompts/speckeep.archive.md",
		".github/prompts/speckeep-archive.prompt.md",
		".cursor/rules/speckeep-archive.mdc",
		".kilocode/workflows/speckeep.archive.md",
		".opencode/commands/speckeep.archive.md",
		".roo/rules/speckeep-archive.md",
		".windsurf/workflows/speckeep.archive.md",
	}
}

// commandWrapperDirPatterns enumerates the per-command wrapper file
// conventions used by every target before skills-first generation replaced
// them with one composite `sdd` skill pack. Shared by LegacyPrefixPaths and
// LegacyCommandWrapperPaths so both prefix eras stay in sync.
func commandWrapperDirPatterns() []struct {
	dir string
	sep string
	ext string
} {
	return []struct {
		dir string
		sep string
		ext string
	}{
		{dir: ".claude/commands", sep: ".", ext: ".md"},
		{dir: ".opencode/commands", sep: ".", ext: ".md"},
		{dir: ".kilocode/workflows", sep: ".", ext: ".md"},
		{dir: ".windsurf/workflows", sep: ".", ext: ".md"},
		{dir: ".trae/rules", sep: ".", ext: ".md"},
		{dir: ".codex/prompts", sep: ".", ext: ".md"},
		{dir: ".cursor/rules", sep: "-", ext: ".mdc"},
		{dir: ".roo/rules", sep: "-", ext: ".md"},
		{dir: ".github/prompts", sep: "-", ext: ".prompt.md"},
	}
}

// LegacyPrefixPaths returns pre-rename `/speckeep.*` per-command wrapper
// paths (from before the `/speckeep.* -> /spk.*` command shortening) that
// refresh and cleanup remove.
func LegacyPrefixPaths(commands []CommandDefinition) []string {
	var paths []string
	for _, cmd := range commands {
		for _, pattern := range commandWrapperDirPatterns() {
			paths = append(paths, pattern.dir+"/"+"speckeep"+pattern.sep+cmd.Name+pattern.ext)
		}
	}
	return paths
}

// LegacyCommandWrapperPaths returns `/spk.*` per-command wrapper paths from
// before skills-first generation (one file per command per target) was
// introduced. These are superseded by the composite `sdd` skill pack, not
// renamed, so refresh and cleanup remove them outright.
func LegacyCommandWrapperPaths(commands []CommandDefinition) []string {
	var paths []string
	for _, cmd := range commands {
		for _, pattern := range commandWrapperDirPatterns() {
			paths = append(paths, pattern.dir+"/"+"spk"+pattern.sep+cmd.Name+pattern.ext)
		}
	}
	return paths
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

func proofHint(lang string) string {
	if lang == "ru" {
		return "Доказанность: каждая закрытая задача в `tasks.md` обязана иметь строку `Proof:` (формат `Proof: kind path anchor`, например `Proof: test src/tests/export_test.go TestRunExport`). Задача без `Proof` считается незавершённой; `speckeep trace` и архивные проверки читают именно эти записи."
	}
	return "Evidence: every completed task in `tasks.md` must carry a `Proof:` line (format `Proof: kind path anchor`, e.g. `Proof: test src/tests/export_test.go TestRunExport`). A task without `Proof` is not complete; `speckeep trace` and archive gates read exactly these records."
}

func antiPatternHint(lang string) string {
	if lang == "ru" {
		return `Запрещено:
- пропускать readiness scripts
- расширять scope / перепланировать во время implement
- отмечать done без observable proof
- делать git commit/push/tag или PR без явной просьбы
- читать весь репозиторий вместо минимального среза`
	}
	return `Do not:
- skip readiness scripts
- expand scope / re-plan during implement
- mark done without observable proof
- run git commit/push/tag or open a PR unless explicitly asked
- read the full repo instead of the minimum slice`
}

func workflowChainHint(lang string) string {
	if lang == "ru" {
		return "Цепочка workflow: constitution → spec → [inspect, опционально] → plan → tasks → implement → archive; verify — опциональный on-demand аудит; propose — one-shot быстрая полоса; converge — быстрый цикл закрытия. Уважайте `workflow.verify` в `.speckeep/speckeep.yaml`. Archive — CLI-only: `speckeep archive <slug> .`."
	}
	return "Workflow chain: constitution → spec → [inspect, optional] → plan → tasks → implement → archive; verify is an optional on-demand audit; propose is the one-shot fast lane; converge is the fast closing loop. Respect `workflow.verify` in `.speckeep/speckeep.yaml`. Archive is CLI-only: `speckeep archive <slug> .`."
}

func titleCase(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
