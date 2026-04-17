package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
	"speckeep/src/internal/config"
)

//go:embed assets/scripts/* assets/scripts/powershell/* assets/lang/* assets/lang/*/* assets/lang/*/templates/* assets/lang/*/templates/prompts/* assets/lang/*/templates/contracts/* assets/lang/*/templates/archive/* assets/demo/* assets/demo/specs/* assets/demo/specs/export-report/* assets/demo/specs/export-report/plan/*
var embedded embed.FS

type File struct {
	TargetPath string
	Content    string
	Mode       fs.FileMode
}

type LanguageSettings struct {
	Default      string
	Docs         string
	Agent        string
	Comments     string
	Shell        string
	AgentTargets []string
}

func ResolveLanguageSettings(defaultLang, docsLang, agentLang, commentsLang string) (LanguageSettings, error) {
	base, err := normalizeLanguage(defaultLang)
	if err != nil {
		return LanguageSettings{}, fmt.Errorf("resolve default language: %w", err)
	}
	docs := base
	if strings.TrimSpace(docsLang) != "" {
		docs, err = normalizeLanguage(docsLang)
		if err != nil {
			return LanguageSettings{}, fmt.Errorf("resolve docs language: %w", err)
		}
	}
	agent := base
	if strings.TrimSpace(agentLang) != "" {
		agent, err = normalizeLanguage(agentLang)
		if err != nil {
			return LanguageSettings{}, fmt.Errorf("resolve agent language: %w", err)
		}
	}
	comments := base
	if strings.TrimSpace(commentsLang) != "" {
		comments, err = normalizeLanguage(commentsLang)
		if err != nil {
			return LanguageSettings{}, fmt.Errorf("resolve comments language: %w", err)
		}
	}
	return LanguageSettings{Default: base, Docs: docs, Agent: agent, Comments: comments}, nil
}

func Files(settings LanguageSettings) ([]File, error) {
	definitions := []struct {
		RelativePath string
		TargetPath   string
		Mode         fs.FileMode
		Language     string
	}{
		{RelativePath: "constitution.md", TargetPath: "constitution.md", Mode: 0o644, Language: settings.Docs},
		{RelativePath: "templates/constitution.md", TargetPath: "templates/constitution.md", Mode: 0o644, Language: settings.Docs},
		{RelativePath: "templates/spec.md", TargetPath: "templates/spec.md", Mode: 0o644, Language: settings.Docs},
		{RelativePath: "templates/plan.md", TargetPath: "templates/plan.md", Mode: 0o644, Language: settings.Docs},
		{RelativePath: "templates/research.md", TargetPath: "templates/research.md", Mode: 0o644, Language: settings.Docs},
		{RelativePath: "templates/tasks.md", TargetPath: "templates/tasks.md", Mode: 0o644, Language: settings.Docs},
		{RelativePath: "templates/data-model.md", TargetPath: "templates/data-model.md", Mode: 0o644, Language: settings.Docs},
		{RelativePath: "templates/contracts/api.md", TargetPath: "templates/contracts/api.md", Mode: 0o644, Language: settings.Docs},
		{RelativePath: "templates/contracts/events.md", TargetPath: "templates/contracts/events.md", Mode: 0o644, Language: settings.Docs},
		{RelativePath: "templates/archive/summary.md", TargetPath: "templates/archive/summary.md", Mode: 0o644, Language: settings.Docs},
		{RelativePath: "templates/inspect.md", TargetPath: "templates/inspect.md", Mode: 0o644, Language: settings.Docs},
		{RelativePath: "templates/verify.md", TargetPath: "templates/verify.md", Mode: 0o644, Language: settings.Docs},
		{RelativePath: "templates/agents-snippet.md", TargetPath: "templates/agents-snippet.md", Mode: 0o644, Language: settings.Agent},
		{RelativePath: "templates/prompts/constitution.md", TargetPath: "templates/prompts/constitution.md", Mode: 0o644, Language: settings.Agent},
		{RelativePath: "templates/prompts/spec.md", TargetPath: "templates/prompts/spec.md", Mode: 0o644, Language: settings.Agent},
		{RelativePath: "templates/prompts/inspect.md", TargetPath: "templates/prompts/inspect.md", Mode: 0o644, Language: settings.Agent},
		{RelativePath: "templates/prompts/plan.md", TargetPath: "templates/prompts/plan.md", Mode: 0o644, Language: settings.Agent},
		{RelativePath: "templates/prompts/tasks.md", TargetPath: "templates/prompts/tasks.md", Mode: 0o644, Language: settings.Agent},
		{RelativePath: "templates/prompts/implement.md", TargetPath: "templates/prompts/implement.md", Mode: 0o644, Language: settings.Agent},
		{RelativePath: "templates/prompts/archive.md", TargetPath: "templates/prompts/archive.md", Mode: 0o644, Language: settings.Agent},
		{RelativePath: "templates/prompts/verify.md", TargetPath: "templates/prompts/verify.md", Mode: 0o644, Language: settings.Agent},
		{RelativePath: "templates/prompts/handoff.md", TargetPath: "templates/prompts/handoff.md", Mode: 0o644, Language: settings.Agent},
		{RelativePath: "templates/prompts/challenge.md", TargetPath: "templates/prompts/challenge.md", Mode: 0o644, Language: settings.Agent},
		{RelativePath: "templates/prompts/scope.md", TargetPath: "templates/prompts/scope.md", Mode: 0o644, Language: settings.Agent},
		{RelativePath: "templates/prompts/recap.md", TargetPath: "templates/prompts/recap.md", Mode: 0o644, Language: settings.Agent},
		{RelativePath: "templates/prompts/hotfix.md", TargetPath: "templates/prompts/hotfix.md", Mode: 0o644, Language: settings.Agent},
	}
	files := make([]File, 0, len(definitions)+9)
	configContent, err := generateConfig(settings)
	if err != nil {
		return nil, err
	}
	files = append(files, File{TargetPath: "speckeep.yaml", Content: configContent, Mode: 0o644})
	for _, definition := range definitions {
		content, err := localizedFileContent(definition.Language, definition.RelativePath)
		if err != nil {
			return nil, err
		}
		content = applyLanguagePlaceholders(content, definition.Language, settings)
		content = applyShellPlaceholders(content, settings)
		files = append(files, File{TargetPath: definition.TargetPath, Content: content, Mode: definition.Mode})
	}
	for _, definition := range shellScriptDefinitions(settings.Shell) {
		content, err := FileContent(definition.AssetPath)
		if err != nil {
			return nil, err
		}
		content = applyShellPlaceholders(content, settings)
		files = append(files, File{TargetPath: definition.TargetPath, Content: content, Mode: definition.Mode})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].TargetPath < files[j].TargetPath })
	return files, nil
}

func shellScriptDefinitions(shell string) []struct {
	AssetPath, TargetPath string
	Mode                  fs.FileMode
} {
	normalizedShell := normalizeShellValue(shell)
	ext := ".sh"
	mode := fs.FileMode(0o755)
	if normalizedShell == "powershell" {
		ext = ".ps1"
		mode = 0o644
	}
	names := []string{
		"run-speckeep",
		"inspect-spec",
		"check-constitution",
		"check-spec-ready",
		"check-inspect-ready",
		"check-plan-ready",
		"check-tasks-ready",
		"check-implement-ready",
		"check-archive-ready",
		"check-verify-ready",
		"archive-feature",
		"trace",
		"verify-task-state",
		"list-open-tasks",
		"link-agents",
		"list-specs",
		"show-spec",
	}
	definitions := make([]struct {
		AssetPath, TargetPath string
		Mode                  fs.FileMode
	}, 0, len(names))
	for _, name := range names {
		assetPath := fmt.Sprintf("assets/scripts/%s%s", name, ext)
		if normalizedShell == "powershell" {
			assetPath = fmt.Sprintf("assets/scripts/powershell/%s%s", name, ext)
		}
		definitions = append(definitions, struct {
			AssetPath, TargetPath string
			Mode                  fs.FileMode
		}{
			AssetPath:  assetPath,
			TargetPath: fmt.Sprintf("scripts/%s%s", name, ext),
			Mode:       mode,
		})
	}
	return definitions
}

func FileContent(path string) (string, error) {
	content, err := embedded.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read embedded asset %s: %w", path, err)
	}
	return string(content), nil
}

func normalizeLanguage(language string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(language))
	if value == "" {
		value = "en"
	}
	switch value {
	case "en", "ru":
		return value, nil
	default:
		return "", fmt.Errorf("unsupported language %q, expected en or ru", language)
	}
}

func localizedFileContent(language, relativePath string) (string, error) {
	language, err := normalizeLanguage(language)
	if err != nil {
		return "", err
	}
	return FileContent(fmt.Sprintf("assets/lang/%s/%s", language, relativePath))
}

func applyLanguagePlaceholders(content, outputLanguage string, settings LanguageSettings) string {
	return strings.NewReplacer(
		"[DEFAULT_LANGUAGE]", settings.Default,
		"[DOCS_LANGUAGE]", languageLabel(settings.Docs, outputLanguage),
		"[AGENT_LANGUAGE]", languageLabel(settings.Agent, outputLanguage),
		"[COMMENTS_LANGUAGE]", languageLabel(settings.Comments, outputLanguage),
	).Replace(content)
}

func applyShellPlaceholders(content string, settings LanguageSettings) string {
	replacements := scriptReplacements(settings.Shell)
	oldNew := make([]string, 0, len(replacements)*2)
	for oldValue, newValue := range replacements {
		oldNew = append(oldNew, oldValue, newValue)
	}
	return strings.NewReplacer(oldNew...).Replace(content)
}

func scriptReplacements(shell string) map[string]string {
	normalizedShell := normalizeShellValue(shell)
	ext := ".sh"
	if normalizedShell == "powershell" {
		ext = ".ps1"
	}
	names := []string{
		"run-speckeep",
		"inspect-spec",
		"check-constitution",
		"check-spec-ready",
		"check-inspect-ready",
		"check-plan-ready",
		"check-tasks-ready",
		"check-implement-ready",
		"check-archive-ready",
		"check-verify-ready",
		"archive-feature",
		"verify-task-state",
		"list-open-tasks",
		"link-agents",
		"list-specs",
		"show-spec",
	}
	replacements := make(map[string]string, len(names))
	for _, name := range names {
		replacements[name+".sh"] = name + ext
	}
	return replacements
}

func normalizeShellValue(shell string) string {
	if strings.EqualFold(strings.TrimSpace(shell), "powershell") {
		return "powershell"
	}
	return "sh"
}

// DemoFiles returns the pre-populated example artifacts for the demo workspace.
// All paths are relative to the .speckeep directory.
func DemoFiles() ([]File, error) {
	entries := []struct {
		assetPath  string
		targetPath string
	}{
		{"assets/demo/constitution.md", "constitution.md"},
		{"assets/demo/specs/export-report.md", "specs/export-report/spec.md"},
		{"assets/demo/specs/export-report.inspect.md", "specs/export-report/inspect.md"},
		{"assets/demo/specs/export-report/plan/plan.md", "specs/export-report/plan/plan.md"},
		{"assets/demo/specs/export-report/plan/tasks.md", "specs/export-report/plan/tasks.md"},
		{"assets/demo/specs/export-report/plan/data-model.md", "specs/export-report/plan/data-model.md"},
	}
	files := make([]File, 0, len(entries))
	for _, e := range entries {
		content, err := FileContent(e.assetPath)
		if err != nil {
			return nil, err
		}
		files = append(files, File{TargetPath: e.targetPath, Content: content, Mode: 0o644})
	}
	return files, nil
}

func languageLabel(code, outputLanguage string) string {
	switch strings.ToLower(strings.TrimSpace(outputLanguage)) {
	case "ru":
		switch code {
		case "ru":
			return "русский"
		case "en":
			return "английский"
		}
	default:
		switch code {
		case "ru":
			return "Russian"
		case "en":
			return "English"
		}
	}
	return code
}

func generateConfig(settings LanguageSettings) (string, error) {
	cfg := config.Default()
	cfg.Runtime.Shell = normalizeShellValue(settings.Shell)
	cfg.Scripts = config.ScriptDefaultsForShell(cfg.Runtime.Shell)
	cfg.Language.Default = settings.Default
	cfg.Language.Docs = settings.Docs
	cfg.Language.Agent = settings.Agent
	cfg.Language.Comments = settings.Comments
	cfg.Agents.Targets = settings.AgentTargets
	content, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshal generated speckeep config: %w", err)
	}
	return string(content), nil
}
