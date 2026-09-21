package project

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"speckeep/src/internal/agents"
	"speckeep/src/internal/config"
)

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

func AddAgents(root string, options AddAgentsOptions) (AddAgentsResult, error) {
	root, cfg, err := loadInitializedProject(root)
	if err != nil {
		return AddAgentsResult{}, err
	}

	deprecated := agents.DeprecatedTargetsRequested(options.Targets)
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
	if err := config.Save(context.Background(), root, cfg); err != nil {
		return AddAgentsResult{}, err
	}

	messages := []string{"updated .speckeep/speckeep.yaml with agent targets"}
	for _, target := range deprecated {
		messages = append(messages, fmt.Sprintf("skipped %q: no longer a supported agent target", target))
	}
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
	if err := config.Save(context.Background(), root, cfg); err != nil {
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

	// Also clean up old per-command wrapper files: pre-rename (speckeep.*)
	// and pre-skills-first (spk.*) — both eras are superseded by the
	// composite sdd skill pack, regardless of which target is enabled.
	commands := agents.DefaultCommands(cfg.Runtime.Shell)
	oldPaths := append(agents.LegacyPrefixPaths(commands), agents.LegacyCommandWrapperPaths(commands)...)
	oldPaths = append(oldPaths, agents.LegacySkillPhasePaths(commands)...)
	oldPaths = append(oldPaths, agents.LegacyOpenCodeCommandPaths(commands)...)
	oldPaths = append(oldPaths, agents.LegacyAmazonQSkillPaths(commands)...)
	oldPaths = append(oldPaths, agents.LegacyContinueSkillPaths(commands)...)
	for _, relPath := range oldPaths {
		fullPath := filepath.Join(root, filepath.FromSlash(relPath))
		if _, err := os.Stat(fullPath); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return CleanupAgentsResult{}, err
		}
		// Check if the parent dir exists and matches a disabled target or any target
		// Only clean up if the new-style file doesn't exist (its target is unknown)
		// or if the file is orphaned from a disabled target
		if err := os.Remove(fullPath); err != nil {
			return CleanupAgentsResult{}, err
		}
		messages = append(messages, fmt.Sprintf("removed orphaned agent artifact %s", rel(root, fullPath)))
		removedAny = true
	}
	pruneEmptyOpenCodeCommandsDir(root)

	if !removedAny {
		messages = append(messages, "no orphaned agent artifacts found")
	}

	return CleanupAgentsResult{Messages: messages}, nil
}

// pruneEmptyOpenCodeCommandsDir best-effort removes the now-unused
// `.opencode/commands/` directory once every speckeep slash-command file has
// been cleaned up (OpenCode ships skills-only). A non-empty directory — e.g.
// the user's own OpenCode commands — is left untouched, as are any errors.
func pruneEmptyOpenCodeCommandsDir(root string) {
	dir := filepath.Join(root, ".opencode", "commands")
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		return
	}
	_ = os.Remove(dir)
}
