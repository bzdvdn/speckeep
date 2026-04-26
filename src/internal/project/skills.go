package project

import (
	"fmt"
	"path/filepath"
	"strings"

	"speckeep/src/internal/config"
	"speckeep/src/internal/skills"
)

type AddSkillOptions struct {
	ID        string
	FromLocal string
	FromGit   string
	Ref       string
	Path      string
	Version   string
	Enabled   bool
	NoInstall bool
}

type AddSkillResult struct {
	Messages []string
	Entry    skills.Entry
}

type RemoveSkillOptions struct {
	ID        string
	NoInstall bool
}

type RemoveSkillResult struct {
	Messages []string
	Removed  bool
}

type ListSkillsResult struct {
	Skills []skills.Entry
}

type SyncSkillsOptions struct {
	DryRun bool
}

type SyncSkillsResult struct {
	DryRun    bool     `json:"dry_run"`
	Created   []string `json:"created,omitempty"`
	Updated   []string `json:"updated,omitempty"`
	Unchanged []string `json:"unchanged,omitempty"`
	Messages  []string `json:"messages,omitempty"`
}

func AddSkill(root string, options AddSkillOptions) (AddSkillResult, error) {
	root, cfg, err := loadInitializedProject(root)
	if err != nil {
		return AddSkillResult{}, err
	}

	result, err := skills.Add(root, skills.AddOptions{
		ID:        options.ID,
		FromLocal: options.FromLocal,
		FromGit:   options.FromGit,
		Ref:       options.Ref,
		Path:      options.Path,
		Version:   options.Version,
		Enabled:   options.Enabled,
	})
	if err != nil {
		return AddSkillResult{}, err
	}

	locationKind := result.Entry.Source
	location := result.Entry.Location
	if result.Entry.Ref != "" {
		location = fmt.Sprintf("%s@%s", location, result.Entry.Ref)
	}

	verb := "updated"
	if result.Created {
		verb = "added"
	}
	messages := []string{
		fmt.Sprintf("%s skill %q (%s: %s)", verb, result.Entry.ID, locationKind, location),
		"updated .speckeep/skills/manifest.yaml",
	}
	if strings.TrimSpace(result.Entry.Path) != "" {
		messages = append(messages, fmt.Sprintf("skill path: %s", result.Entry.Path))
	}
	if strings.TrimSpace(result.Entry.CheckoutDir) != "" {
		messages = append(messages, fmt.Sprintf("skill checkout: %s", result.Entry.CheckoutDir))
	}
	if strings.TrimSpace(result.Entry.ResolvedCommit) != "" {
		messages = append(messages, fmt.Sprintf("resolved commit: %s", result.Entry.ResolvedCommit))
	}
	if err := refreshAgentsSnippetFromConfig(root, cfg); err != nil {
		return AddSkillResult{}, err
	}
	messages = append(messages, "updated managed SpecKeep block in AGENTS.md")
	if options.NoInstall {
		messages = append(messages, "skipped skill installation into agent folders (--no-install)")
	} else {
		installResult, err := InstallSkills(root, InstallSkillsOptions{Targets: cfg.Agents.Targets})
		if err != nil {
			return AddSkillResult{}, err
		}
		messages = append(messages, skillInstallMessages(installResult)...)
	}

	return AddSkillResult{
		Messages: messages,
		Entry:    result.Entry,
	}, nil
}

func RemoveSkill(root string, options RemoveSkillOptions) (RemoveSkillResult, error) {
	root, cfg, err := loadInitializedProject(root)
	if err != nil {
		return RemoveSkillResult{}, err
	}

	removed, err := skills.Remove(root, options.ID)
	if err != nil {
		return RemoveSkillResult{}, err
	}

	if !removed.Removed {
		return RemoveSkillResult{
			Removed:  false,
			Messages: []string{fmt.Sprintf("skill %q is not configured", strings.TrimSpace(options.ID))},
		}, nil
	}
	if err := refreshAgentsSnippetFromConfig(root, cfg); err != nil {
		return RemoveSkillResult{}, err
	}
	messages := []string{
		fmt.Sprintf("removed skill %q", strings.TrimSpace(options.ID)),
		"updated .speckeep/skills/manifest.yaml",
		"updated managed SpecKeep block in AGENTS.md",
	}
	if options.NoInstall {
		messages = append(messages, "skipped skill installation into agent folders (--no-install)")
	} else {
		installResult, err := InstallSkills(root, InstallSkillsOptions{Targets: cfg.Agents.Targets})
		if err != nil {
			return RemoveSkillResult{}, err
		}
		messages = append(messages, skillInstallMessages(installResult)...)
	}
	return RemoveSkillResult{
		Removed:  true,
		Messages: messages,
	}, nil
}

func ListSkills(root string) (ListSkillsResult, error) {
	root, _, err := loadInitializedProject(root)
	if err != nil {
		return ListSkillsResult{}, err
	}

	manifest, err := skills.Load(root)
	if err != nil {
		return ListSkillsResult{}, err
	}

	return ListSkillsResult{
		Skills: append([]skills.Entry(nil), manifest.Skills...),
	}, nil
}

func SyncSkills(root string, options SyncSkillsOptions) (SyncSkillsResult, error) {
	root, cfg, err := loadInitializedProject(root)
	if err != nil {
		return SyncSkillsResult{}, err
	}

	internalResult := RefreshResult{DryRun: options.DryRun}
	if err := syncSkillsManifest(root, options.DryRun, &internalResult); err != nil {
		return SyncSkillsResult{}, err
	}
	if err := refreshAgentsSnippetFromConfigWithDryRun(root, cfg, options.DryRun, &internalResult); err != nil {
		return SyncSkillsResult{}, err
	}
	internalResult.Messages = buildRefreshMessages(internalResult)

	return SyncSkillsResult{
		DryRun:    internalResult.DryRun,
		Created:   append([]string(nil), internalResult.Created...),
		Updated:   append([]string(nil), internalResult.Updated...),
		Unchanged: append([]string(nil), internalResult.Unchanged...),
		Messages:  append([]string(nil), internalResult.Messages...),
	}, nil
}

func refreshAgentsSnippetFromConfig(root string, cfg config.Config) error {
	var syncResult RefreshResult
	return refreshAgentsSnippetFromConfigWithDryRun(root, cfg, false, &syncResult)
}

func refreshAgentsSnippetFromConfigWithDryRun(root string, cfg config.Config, dryRun bool, result *RefreshResult) error {
	templatesDir, err := cfg.TemplatesDir(root)
	if err != nil {
		return err
	}

	return syncAgentsSnippet(
		root,
		filepath.Join(root, cfg.Agents.AgentsFile),
		filepath.Join(templatesDir, "agents-snippet.md"),
		dryRun,
		result,
	)
}

func skillInstallMessages(result InstallSkillsResult) []string {
	messages := []string{
		fmt.Sprintf(
			"installed skills into agent folders (created=%d updated=%d removed=%d unchanged=%d)",
			len(result.Created),
			len(result.Updated),
			len(result.Removed),
			len(result.Unchanged),
		),
	}
	if len(result.Warnings) > 0 {
		messages = append(messages, result.Warnings...)
	}
	return messages
}
