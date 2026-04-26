package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"speckeep/src/internal/agents"
	"speckeep/src/internal/config"
	"speckeep/src/internal/gitutil"
	"speckeep/src/internal/templates"
)

type InitOptions struct {
	InitGit          bool
	DefaultLang      string
	DocsLang         string
	AgentLang        string
	CommentsLang     string
	Shell            string
	SpecsDir         string
	ArchiveDir       string
	ConstitutionFile string
	AgentTargets     []string
}

type InitResult struct {
	Messages []string

	RootAbs string

	Shell            string
	DocsLang         string
	AgentLang        string
	CommentsLang     string
	AgentTargets     []string
	SpecsDir         string
	ArchiveDir       string
	ConstitutionFile string

	GitRepoStatus string // initialized, kept, skipped

	EnsuredDirs []string
	Created     []string
	Kept        []string

	AgentsSnippetChanged  bool
	AgentArtifactMessages []string
}

type AddAgentsOptions struct {
	Targets   []string
	AgentLang string
}

type AddAgentsResult struct{ Messages []string }

type RemoveAgentsOptions struct {
	Targets []string
}

type RemoveAgentsResult struct{ Messages []string }

type ListAgentsResult struct {
	Targets []string
}

type CleanupAgentsResult struct{ Messages []string }

func Initialize(root string, options InitOptions) (InitResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return InitResult{}, err
	}

	configPath := filepath.Join(root, ".speckeep", "speckeep.yaml")
	configExists := fileExists(configPath)

	languages, err := templates.ResolveLanguageSettings(options.DefaultLang, options.DocsLang, options.AgentLang, options.CommentsLang)
	if err != nil {
		return InitResult{}, err
	}
	normalizedAgentTargets, err := agents.NormalizeTargets(options.AgentTargets)
	if err != nil {
		return InitResult{}, err
	}
	shell, err := config.NormalizeShell(options.Shell)
	if err != nil {
		return InitResult{}, err
	}
	languages.AgentTargets = normalizedAgentTargets
	languages.Shell = shell

	cfg := config.Default()
	if configExists {
		loaded, err := config.Load(root)
		if err != nil {
			return InitResult{}, err
		}
		cfg = loaded
		languages.Default = cfg.Language.Default
		languages.Docs = cfg.Language.Docs
		languages.Agent = cfg.Language.Agent
		languages.Comments = cfg.Language.Comments
		languages.Shell = cfg.Runtime.Shell
		languages.AgentTargets = cfg.Agents.Targets
	} else {
		cfg.Language.Default = languages.Default
		cfg.Language.Docs = languages.Docs
		cfg.Language.Agent = languages.Agent
		cfg.Language.Comments = languages.Comments
		cfg.Runtime.Shell = shell
		cfg.Scripts = config.ScriptDefaultsForShell(shell)
		cfg.Agents.Targets = normalizedAgentTargets
		if strings.TrimSpace(options.SpecsDir) != "" {
			cfg.Paths.SpecsDir = strings.TrimSpace(options.SpecsDir)
		}
		if strings.TrimSpace(options.ArchiveDir) != "" {
			cfg.Paths.ArchiveDir = strings.TrimSpace(options.ArchiveDir)
		}
		if strings.TrimSpace(options.ConstitutionFile) != "" {
			value := strings.TrimSpace(options.ConstitutionFile)
			if filepath.IsAbs(value) {
				return InitResult{}, fmt.Errorf("constitution-file must be a relative path, got %q", value)
			}
			cfg.Project.ConstitutionFile = value
		}
	}

	result := InitResult{
		RootAbs:      root,
		Shell:        cfg.Runtime.Shell,
		DocsLang:     cfg.Language.Docs,
		AgentLang:    cfg.Language.Agent,
		CommentsLang: cfg.Language.Comments,
		AgentTargets: cfg.Agents.Targets,
	}

	var messages []string
	if options.InitGit {
		created, err := gitutil.EnsureRepository(root)
		if err != nil {
			return InitResult{}, err
		}
		if created {
			result.GitRepoStatus = "initialized"
			messages = append(messages, "initialized git repository")
		} else {
			result.GitRepoStatus = "kept"
			messages = append(messages, "kept existing git repository")
		}
	} else {
		result.GitRepoStatus = "skipped"
		messages = append(messages, "skipped git repository initialization")
	}
	draftspecDir, err := cfg.DraftspecDir(root)
	if err != nil {
		return InitResult{}, err
	}
	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return InitResult{}, err
	}
	archiveDir, err := cfg.ArchiveDir(root)
	if err != nil {
		return InitResult{}, err
	}
	templatesDir, err := cfg.TemplatesDir(root)
	if err != nil {
		return InitResult{}, err
	}
	scriptsDir, err := cfg.ScriptsDir(root)
	if err != nil {
		return InitResult{}, err
	}
	constitutionAbs := filepath.Clean(filepath.Join(root, cfg.Project.ConstitutionFile))
	result.SpecsDir = rel(root, specsDir)
	result.ArchiveDir = rel(root, archiveDir)
	result.ConstitutionFile = rel(root, constitutionAbs)
	subdirs := []string{
		draftspecDir,
		filepath.Join(draftspecDir, "skills"),
		specsDir,
		archiveDir,
		templatesDir,
		filepath.Join(templatesDir, "prompts"),
		filepath.Join(templatesDir, "contracts"),
		filepath.Join(templatesDir, "archive"),
		scriptsDir,
		filepath.Dir(constitutionAbs),
	}
	for _, dir := range subdirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return InitResult{}, err
		}
		relPath := rel(root, dir)
		result.EnsuredDirs = append(result.EnsuredDirs, relPath)
		messages = append(messages, fmt.Sprintf("ensured directory %s", relPath))
	}
	files, err := templates.Files(languages)
	if err != nil {
		return InitResult{}, err
	}
	for _, file := range files {
		target := filepath.Join(draftspecDir, file.TargetPath)
		if file.TargetPath == "constitution.md" {
			target = constitutionAbs
		}
		written, err := writeIfMissing(target, file.Content, file.Mode)
		if err != nil {
			return InitResult{}, err
		}
		if written {
			result.Created = append(result.Created, rel(root, target))
			messages = append(messages, fmt.Sprintf("created %s", rel(root, target)))
		} else {
			result.Kept = append(result.Kept, rel(root, target))
			messages = append(messages, fmt.Sprintf("kept existing %s", rel(root, target)))
		}

		if file.TargetPath == "speckeep.yaml" && written && !configExists {
			if err := config.Save(root, cfg); err != nil {
				return InitResult{}, err
			}
		}
	}
	messages = append(messages, fmt.Sprintf("configured languages: docs=%s agent=%s comments=%s", cfg.Language.Docs, cfg.Language.Agent, cfg.Language.Comments))
	messages = append(messages, fmt.Sprintf("configured shell: %s", cfg.Runtime.Shell))
	agentsPath := filepath.Join(root, "AGENTS.md")
	snippetPath := filepath.Join(templatesDir, "agents-snippet.md")
	changed, err := ensureAgentsSnippet(root, agentsPath, snippetPath)
	if err != nil {
		return InitResult{}, err
	}
	result.AgentsSnippetChanged = changed
	if changed {
		messages = append(messages, "updated AGENTS.md with SpecKeep guidance")
	} else {
		messages = append(messages, "kept existing AGENTS.md SpecKeep guidance")
	}
	result.AgentArtifactMessages = ensureAgentFiles(root, normalizedAgentTargets, languages.Agent, cfg.Runtime.Shell)
	messages = append(messages, result.AgentArtifactMessages...)
	if len(normalizedAgentTargets) > 0 {
		messages = append(messages, fmt.Sprintf("enabled agent targets: %s", strings.Join(normalizedAgentTargets, ", ")))
	} else {
		messages = append(messages, "enabled agent targets: none")
	}
	result.Messages = messages

	return result, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func AddAgents(root string, options AddAgentsOptions) (AddAgentsResult, error) {
	root, cfg, err := loadInitializedProject(root)
	if err != nil {
		return AddAgentsResult{}, err
	}

	requested, err := agents.NormalizeTargets(options.Targets)
	if err != nil {
		return AddAgentsResult{}, err
	}
	combined, err := agents.NormalizeTargets(append(cfg.Agents.Targets, requested...))
	if err != nil {
		return AddAgentsResult{}, err
	}

	agentLanguage := cfg.Language.Agent
	if strings.TrimSpace(options.AgentLang) != "" {
		agentLanguage = strings.TrimSpace(options.AgentLang)
	}

	cfg.Agents.Targets = combined
	if err := config.Save(root, cfg); err != nil {
		return AddAgentsResult{}, err
	}

	messages := []string{"updated .speckeep/speckeep.yaml with agent targets"}
	messages = append(messages, ensureAgentFiles(root, requested, agentLanguage, cfg.Runtime.Shell)...)
	messages = append(messages, fmt.Sprintf("enabled agent targets: %s", strings.Join(combined, ", ")))
	return AddAgentsResult{Messages: messages}, nil
}

func RemoveAgents(root string, options RemoveAgentsOptions) (RemoveAgentsResult, error) {
	root, cfg, err := loadInitializedProject(root)
	if err != nil {
		return RemoveAgentsResult{}, err
	}

	requested, err := agents.NormalizeTargets(options.Targets)
	if err != nil {
		return RemoveAgentsResult{}, err
	}
	removeSet := make(map[string]struct{}, len(requested))
	for _, target := range requested {
		removeSet[target] = struct{}{}
	}

	var remaining []string
	for _, target := range cfg.Agents.Targets {
		if _, ok := removeSet[target]; ok {
			continue
		}
		remaining = append(remaining, target)
	}
	sort.Strings(remaining)
	cfg.Agents.Targets = remaining
	if err := config.Save(root, cfg); err != nil {
		return RemoveAgentsResult{}, err
	}

	messages := []string{"updated .speckeep/speckeep.yaml with agent targets"}
	messages = append(messages, removeAgentFiles(root, requested)...)
	if len(remaining) > 0 {
		messages = append(messages, fmt.Sprintf("enabled agent targets: %s", strings.Join(remaining, ", ")))
	} else {
		messages = append(messages, "enabled agent targets: none")
	}
	return RemoveAgentsResult{Messages: messages}, nil
}

func ListAgents(root string) (ListAgentsResult, error) {
	_, cfg, err := loadInitializedProject(root)
	if err != nil {
		return ListAgentsResult{}, err
	}
	return ListAgentsResult{Targets: append([]string(nil), cfg.Agents.Targets...)}, nil
}

func CleanupAgents(root string) (CleanupAgentsResult, error) {
	root, cfg, err := loadInitializedProject(root)
	if err != nil {
		return CleanupAgentsResult{}, err
	}

	enabledTargets := make(map[string]struct{}, len(cfg.Agents.Targets))
	for _, target := range cfg.Agents.Targets {
		enabledTargets[target] = struct{}{}
	}

	var messages []string
	removedAny := false
	for _, target := range agents.SupportedTargets() {
		if _, ok := enabledTargets[target]; ok {
			continue
		}
		paths, err := agents.PathsForTarget(target)
		if err != nil {
			return CleanupAgentsResult{}, err
		}
		for _, relPath := range paths {
			fullPath := filepath.Join(root, filepath.FromSlash(relPath))
			if _, err := os.Stat(fullPath); errors.Is(err, os.ErrNotExist) {
				continue
			} else if err != nil {
				return CleanupAgentsResult{}, err
			}
			if err := os.Remove(fullPath); err != nil {
				return CleanupAgentsResult{}, err
			}
			messages = append(messages, fmt.Sprintf("removed orphaned agent artifact %s", rel(root, fullPath)))
			removedAny = true
		}
	}

	if !removedAny {
		messages = append(messages, "no orphaned agent artifacts found")
	}

	return CleanupAgentsResult{Messages: messages}, nil
}

func writeIfMissing(path, content string, mode os.FileMode) (bool, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	return true, os.WriteFile(path, []byte(content), mode)
}

func ensureAgentsSnippet(root, path, snippetPath string) (bool, error) {
	snippetBytes, err := os.ReadFile(snippetPath)
	if err != nil {
		return false, err
	}
	block, err := renderManagedAgentsBlockForRoot(root, string(snippetBytes))
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return true, os.WriteFile(path, []byte(block), 0o644)
	} else if err != nil {
		return false, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	updated := upsertManagedAgentsBlock(string(content), block)
	if updated == string(content) {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(updated), 0o644)
}

func ensureAgentFiles(root string, targets []string, language string, shell string) []string {
	agentFiles, err := agents.Files(targets, language, shell)
	if err != nil {
		return []string{fmt.Sprintf("skipped agent files: %v", err)}
	}
	messages := make([]string, 0, len(agentFiles))
	for _, file := range agentFiles {
		target := filepath.Join(root, filepath.FromSlash(file.Path))
		written, err := writeIfMissing(target, file.Content, file.Mode)
		if err != nil {
			messages = append(messages, fmt.Sprintf("failed %s: %v", rel(root, target), err))
			continue
		}
		if written {
			messages = append(messages, fmt.Sprintf("created %s", rel(root, target)))
		} else {
			messages = append(messages, fmt.Sprintf("kept existing %s", rel(root, target)))
		}
	}
	return messages
}

func removeAgentFiles(root string, targets []string) []string {
	messages := []string{}
	for _, target := range targets {
		paths, err := agents.PathsForTarget(target)
		if err != nil {
			messages = append(messages, fmt.Sprintf("skipped removing %s: %v", target, err))
			continue
		}
		for _, relPath := range paths {
			fullPath := filepath.Join(root, filepath.FromSlash(relPath))
			if err := os.Remove(fullPath); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					messages = append(messages, fmt.Sprintf("missing %s", rel(root, fullPath)))
					continue
				}
				messages = append(messages, fmt.Sprintf("failed %s: %v", rel(root, fullPath), err))
				continue
			}
			messages = append(messages, fmt.Sprintf("removed %s", rel(root, fullPath)))
		}
	}
	return messages
}

func loadInitializedProject(root string) (string, config.Config, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", config.Config{}, err
	}
	cfgPath := filepath.Join(absRoot, ".speckeep", "speckeep.yaml")
	if _, err := os.Stat(cfgPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", config.Config{}, fmt.Errorf("speckeep project is not initialized at %s", absRoot)
		}
		return "", config.Config{}, err
	}
	cfg, err := config.Load(absRoot)
	if err != nil {
		return "", config.Config{}, err
	}
	return absRoot, cfg, nil
}

func rel(root, target string) string {
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return target
	}
	return relative
}
