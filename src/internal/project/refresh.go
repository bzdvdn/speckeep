package project

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"speckeep/src/internal/agents"
	"speckeep/src/internal/config"
	"speckeep/src/internal/templates"
)

const (
	agentsBlockStart = "<!-- speckeep:start -->"
	agentsBlockEnd   = "<!-- speckeep:end -->"
)

type RefreshOptions struct {
	DefaultLang  string
	DocsLang     string
	AgentLang    string
	CommentsLang string
	Shell        string
	AgentTargets []string
	DryRun       bool
	RewriteTrace bool
}

type RefreshResult struct {
	DryRun    bool     `json:"dry_run"`
	Created   []string `json:"created,omitempty"`
	Updated   []string `json:"updated,omitempty"`
	Rewritten []string `json:"rewritten,omitempty"`
	Unchanged []string `json:"unchanged,omitempty"`
	Removed   []string `json:"removed,omitempty"`
	Messages  []string `json:"messages,omitempty"`
}

func Refresh(root string, options RefreshOptions) (RefreshResult, error) {
	root, cfg, err := loadInitializedProject(root)
	if err != nil {
		return RefreshResult{}, err
	}

	result := RefreshResult{DryRun: options.DryRun}

	languages, shell, agentTargets, err := resolveRefreshSettings(cfg, options)
	if err != nil {
		return RefreshResult{}, err
	}
	languages.AgentTargets = agentTargets
	languages.Shell = shell

	cfg.Language.Default = languages.Default
	cfg.Language.Docs = languages.Docs
	cfg.Language.Agent = languages.Agent
	cfg.Language.Comments = languages.Comments
	cfg.Runtime.Shell = shell
	cfg.Scripts = config.ScriptDefaultsForShell(shell)
	cfg.Agents.Targets = agentTargets

	if err := syncConfig(root, cfg, options.DryRun, &result); err != nil {
		return RefreshResult{}, err
	}

	draftspecDir, err := cfg.DraftspecDir(root)
	if err != nil {
		return RefreshResult{}, err
	}
	templateFiles, err := templates.Files(languages)
	if err != nil {
		return RefreshResult{}, err
	}
	for _, file := range templateFiles {
		if file.TargetPath == "speckeep.yaml" || file.TargetPath == "constitution.md" {
			continue
		}
		target := filepath.Join(draftspecDir, file.TargetPath)
		if err := syncManagedFile(root, target, file.Content, file.Mode, options.DryRun, &result); err != nil {
			return RefreshResult{}, err
		}
	}

	templatesDir, err := cfg.TemplatesDir(root)
	if err != nil {
		return RefreshResult{}, err
	}
	agentsPath := filepath.Join(root, cfg.Agents.AgentsFile)
	snippetPath := filepath.Join(templatesDir, "agents-snippet.md")
	if err := syncAgentsSnippet(root, agentsPath, snippetPath, options.DryRun, &result); err != nil {
		return RefreshResult{}, err
	}

	if err := syncAgentFiles(root, agentTargets, languages.Agent, shell, options.DryRun, &result); err != nil {
		return RefreshResult{}, err
	}
	if err := removeDisabledAgentArtifacts(root, agentTargets, options.DryRun, &result); err != nil {
		return RefreshResult{}, err
	}

	if options.RewriteTrace {
		if err := rewriteTraceAnnotations(root, options.DryRun, &result); err != nil {
			return RefreshResult{}, err
		}
	}

	result.Messages = buildRefreshMessages(result)
	return result, nil
}

func resolveRefreshSettings(cfg config.Config, options RefreshOptions) (templates.LanguageSettings, string, []string, error) {
	defaultLang := cfg.Language.Default
	if strings.TrimSpace(options.DefaultLang) != "" {
		defaultLang = strings.TrimSpace(options.DefaultLang)
	}
	docsLang := cfg.Language.Docs
	if strings.TrimSpace(options.DocsLang) != "" {
		docsLang = strings.TrimSpace(options.DocsLang)
	}
	agentLang := cfg.Language.Agent
	if strings.TrimSpace(options.AgentLang) != "" {
		agentLang = strings.TrimSpace(options.AgentLang)
	}
	commentsLang := cfg.Language.Comments
	if strings.TrimSpace(options.CommentsLang) != "" {
		commentsLang = strings.TrimSpace(options.CommentsLang)
	}

	languages, err := templates.ResolveLanguageSettings(defaultLang, docsLang, agentLang, commentsLang)
	if err != nil {
		return templates.LanguageSettings{}, "", nil, err
	}

	shell := cfg.Runtime.Shell
	if strings.TrimSpace(options.Shell) != "" {
		shell = strings.TrimSpace(options.Shell)
	}
	shell, err = config.NormalizeShell(shell)
	if err != nil {
		return templates.LanguageSettings{}, "", nil, err
	}

	targets := cfg.Agents.Targets
	if len(options.AgentTargets) > 0 {
		targets, err = agents.NormalizeTargets(options.AgentTargets)
		if err != nil {
			return templates.LanguageSettings{}, "", nil, err
		}
	}

	return languages, shell, targets, nil
}

func syncConfig(root string, cfg config.Config, dryRun bool, result *RefreshResult) error {
	content, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal speckeep config: %w", err)
	}
	path, err := cfg.ConfigPath(root)
	if err != nil {
		return err
	}
	return syncManagedFile(root, path, string(content), 0o644, dryRun, result)
}

func syncManagedFile(root, path, content string, mode os.FileMode, dryRun bool, result *RefreshResult) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	current, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		recordRefreshAction(result, "created", rel(root, path))
		if dryRun {
			return nil
		}
		return os.WriteFile(path, []byte(content), mode)
	case err != nil:
		return err
	}

	if bytes.Equal(current, []byte(content)) {
		recordRefreshAction(result, "unchanged", rel(root, path))
		return nil
	}

	recordRefreshAction(result, "updated", rel(root, path))
	if dryRun {
		return nil
	}
	return os.WriteFile(path, []byte(content), mode)
}

func syncAgentsSnippet(root, path, snippetPath string, dryRun bool, result *RefreshResult) error {
	snippetBytes, err := os.ReadFile(snippetPath)
	if err != nil {
		return err
	}
	block := renderManagedAgentsBlock(string(snippetBytes))

	current, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		recordRefreshAction(result, "created", rel(root, path))
		if dryRun {
			return nil
		}
		return os.WriteFile(path, []byte(block), 0o644)
	case err != nil:
		return err
	}

	updated := upsertManagedAgentsBlock(string(current), block)
	if updated == string(current) {
		recordRefreshAction(result, "unchanged", rel(root, path))
		return nil
	}

	recordRefreshAction(result, "updated", rel(root, path))
	if dryRun {
		return nil
	}
	return os.WriteFile(path, []byte(updated), 0o644)
}

func renderManagedAgentsBlock(snippet string) string {
	trimmed := strings.TrimSpace(snippet)
	return agentsBlockStart + "\n" + trimmed + "\n" + agentsBlockEnd + "\n"
}

func upsertManagedAgentsBlock(current, block string) string {
	start := strings.Index(current, agentsBlockStart)
	end := strings.Index(current, agentsBlockEnd)
	if start >= 0 && end > start {
		end += len(agentsBlockEnd)
		updated := current[:start] + block + current[end:]
		return normalizeTrailingWhitespace(updated)
	}

	for _, legacy := range []struct {
		start string
		end   string
	}{
		{start: "<!-- draftspec:start -->", end: "<!-- draftspec:end -->"},
	} {
		legacyStart := strings.Index(current, legacy.start)
		legacyEnd := strings.Index(current, legacy.end)
		if legacyStart >= 0 && legacyEnd > legacyStart {
			legacyEnd += len(legacy.end)
			updated := current[:legacyStart] + block + current[legacyEnd:]
			return normalizeTrailingWhitespace(updated)
		}
	}

	for _, header := range []string{"## SpecKeep", "## Draftspec"} {
		if legacy := strings.Index(current, header); legacy >= 0 {
			updated := current[:legacy] + block
			return normalizeTrailingWhitespace(updated)
		}
	}

	if strings.TrimSpace(current) == "" {
		return block
	}

	updated := strings.TrimRight(current, "\n") + "\n\n" + block
	return normalizeTrailingWhitespace(updated)
}

func normalizeTrailingWhitespace(content string) string {
	return strings.TrimRight(content, "\n") + "\n"
}

func syncAgentFiles(root string, targets []string, language string, shell string, dryRun bool, result *RefreshResult) error {
	files, err := agents.Files(targets, language, shell)
	if err != nil {
		return err
	}
	for _, file := range files {
		target := filepath.Join(root, filepath.FromSlash(file.Path))
		if err := syncManagedFile(root, target, file.Content, file.Mode, dryRun, result); err != nil {
			return err
		}
	}
	return nil
}

func removeDisabledAgentArtifacts(root string, enabled []string, dryRun bool, result *RefreshResult) error {
	enabledSet := make(map[string]struct{}, len(enabled))
	for _, target := range enabled {
		enabledSet[target] = struct{}{}
	}

	for _, target := range agents.SupportedTargets() {
		if _, ok := enabledSet[target]; ok {
			continue
		}
		paths, err := agents.PathsForTarget(target)
		if err != nil {
			return err
		}
		for _, relPath := range paths {
			fullPath := filepath.Join(root, filepath.FromSlash(relPath))
			if _, err := os.Stat(fullPath); errors.Is(err, os.ErrNotExist) {
				continue
			} else if err != nil {
				return err
			}
			recordRefreshAction(result, "removed", rel(root, fullPath))
			if dryRun {
				continue
			}
			if err := os.Remove(fullPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func recordRefreshAction(result *RefreshResult, action string, path string) {
	switch action {
	case "created":
		result.Created = append(result.Created, path)
	case "updated":
		result.Updated = append(result.Updated, path)
	case "rewritten":
		result.Rewritten = append(result.Rewritten, path)
	case "unchanged":
		result.Unchanged = append(result.Unchanged, path)
	case "removed":
		result.Removed = append(result.Removed, path)
	}
}

func buildRefreshMessages(result RefreshResult) []string {
	var messages []string
	prefix := ""
	if result.DryRun {
		prefix = "would "
	}
	for _, path := range result.Created {
		messages = append(messages, prefix+"create "+path)
	}
	for _, path := range result.Updated {
		messages = append(messages, prefix+"update "+path)
	}
	for _, path := range result.Rewritten {
		messages = append(messages, prefix+"rewrite "+path)
	}
	for _, path := range result.Removed {
		messages = append(messages, prefix+"remove "+path)
	}
	for _, path := range result.Unchanged {
		messages = append(messages, "unchanged "+path)
	}
	return messages
}

func rewriteTraceAnnotations(root string, dryRun bool, result *RefreshResult) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if shouldSkipTraceRewrite(path, entry) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if bytes.IndexByte(content, 0) >= 0 {
			return nil
		}
		if !bytes.Contains(content, []byte("@ds-")) {
			return nil
		}

		updated := bytes.ReplaceAll(content, []byte("@ds-task"), []byte("@sk-task"))
		updated = bytes.ReplaceAll(updated, []byte("@ds-test"), []byte("@sk-test"))
		if bytes.Equal(updated, content) {
			return nil
		}

		recordRefreshAction(result, "rewritten", rel(root, path))
		if dryRun {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return nil
		}
		mode := info.Mode() & os.ModePerm
		return os.WriteFile(path, updated, mode)
	})
}

func shouldSkipTraceRewrite(path string, entry os.DirEntry) bool {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") && base != "." {
		return true
	}
	if entry.IsDir() {
		switch base {
		case "node_modules", "vendor", "dist", "bin", "obj", ".git", ".speckeep":
			return true
		}
	}
	return false
}

func (r RefreshResult) MarshalJSON() ([]byte, error) {
	type alias RefreshResult
	return json.Marshal(alias(r))
}
