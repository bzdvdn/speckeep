package project

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"speckeep/src/internal/agents"
	"speckeep/src/internal/skills"
)

type InstallSkillsOptions struct {
	Targets         []string
	DryRun          bool
	IncludeDisabled bool
}

type InstallSkillsResult struct {
	DryRun    bool     `json:"dry_run"`
	Created   []string `json:"created,omitempty"`
	Updated   []string `json:"updated,omitempty"`
	Unchanged []string `json:"unchanged,omitempty"`
	Removed   []string `json:"removed,omitempty"`
	Messages  []string `json:"messages,omitempty"`
	Warnings  []string `json:"warnings,omitempty"`
}

func InstallSkills(root string, options InstallSkillsOptions) (InstallSkillsResult, error) {
	root, cfg, err := loadInitializedProject(root)
	if err != nil {
		return InstallSkillsResult{}, err
	}

	targets := cfg.Agents.Targets
	if len(options.Targets) > 0 {
		targets, err = agents.NormalizeTargets(options.Targets)
		if err != nil {
			return InstallSkillsResult{}, err
		}
	}

	manifest, err := skills.Load(root)
	if err != nil {
		return InstallSkillsResult{}, err
	}

	refresh := RefreshResult{DryRun: options.DryRun}
	warnings, err := installSkillsForTargets(root, targets, manifest.Skills, options.IncludeDisabled, options.DryRun, &refresh)
	if err != nil {
		return InstallSkillsResult{}, err
	}

	refresh.Messages = buildRefreshMessages(refresh)
	refresh.Messages = append(refresh.Messages, warnings...)

	return InstallSkillsResult{
		DryRun:    refresh.DryRun,
		Created:   append([]string(nil), refresh.Created...),
		Updated:   append([]string(nil), refresh.Updated...),
		Unchanged: append([]string(nil), refresh.Unchanged...),
		Removed:   append([]string(nil), refresh.Removed...),
		Messages:  append([]string(nil), refresh.Messages...),
		Warnings:  append([]string(nil), warnings...),
	}, nil
}

func installSkillsForTargets(root string, targets []string, entries []skills.Entry, includeDisabled bool, dryRun bool, result *RefreshResult) ([]string, error) {
	var warnings []string
	for _, target := range targets {
		basePath, supported := agents.SkillBasePath(target)
		if !supported {
			warnings = append(warnings, fmt.Sprintf("skip skill install for target %q: no known skill directory", target))
			continue
		}

		baseAbs := filepath.Join(root, filepath.FromSlash(basePath))
		expected := map[string]struct{}{}
		for _, entry := range entries {
			if !entry.Enabled && !includeDisabled {
				continue
			}

			srcDir, err := resolveSkillSourceDir(root, entry)
			if err != nil {
				return warnings, err
			}
			dstDir := filepath.Join(baseAbs, entry.ID)
			expected[entry.ID] = struct{}{}

			if err := syncManagedDir(root, dstDir, srcDir, dryRun, result); err != nil {
				return warnings, err
			}
		}

		if err := removeStaleSkillDirs(root, baseAbs, expected, dryRun, result); err != nil {
			return warnings, err
		}
	}

	return warnings, nil
}

func resolveSkillSourceDir(root string, entry skills.Entry) (string, error) {
	base := strings.TrimSpace(entry.Location)
	switch entry.Source {
	case "local":
		if base == "" {
			return "", fmt.Errorf("skill %q has empty local location", entry.ID)
		}
		if !filepath.IsAbs(base) {
			base = filepath.Join(root, filepath.FromSlash(base))
		}
	case "git":
		base = strings.TrimSpace(entry.CheckoutDir)
		if base == "" {
			return "", fmt.Errorf("skill %q has no checkout_dir; run add-skill again", entry.ID)
		}
		if !filepath.IsAbs(base) {
			base = filepath.Join(root, filepath.FromSlash(base))
		}
	default:
		return "", fmt.Errorf("skill %q has unsupported source %q", entry.ID, entry.Source)
	}

	subPath := normalizeSkillSubPath(entry.Path)
	if subPath != "." {
		base = filepath.Join(base, filepath.FromSlash(subPath))
	}

	info, err := os.Stat(base)
	if err != nil {
		return "", fmt.Errorf("skill %q source path is not accessible: %w", entry.ID, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("skill %q source path must be a directory: %s", entry.ID, base)
	}
	return base, nil
}

func normalizeSkillSubPath(path string) string {
	cleaned := filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
	if cleaned == "" || cleaned == "." {
		return "."
	}
	return cleaned
}

func removeStaleSkillDirs(root, baseAbs string, expected map[string]struct{}, dryRun bool, result *RefreshResult) error {
	entries, err := os.ReadDir(baseAbs)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, ok := expected[entry.Name()]; ok {
			continue
		}
		target := filepath.Join(baseAbs, entry.Name())
		recordRefreshAction(result, "removed", rel(root, target))
		if dryRun {
			continue
		}
		if err := os.RemoveAll(target); err != nil {
			return err
		}
	}
	return nil
}

func syncManagedDir(root, destination, source string, dryRun bool, result *RefreshResult) error {
	srcInfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", source)
	}

	dstInfo, err := os.Stat(destination)
	if os.IsNotExist(err) {
		recordRefreshAction(result, "created", rel(root, destination))
		if dryRun {
			return nil
		}
		return copyDir(source, destination)
	}
	if err != nil {
		return err
	}
	if !dstInfo.IsDir() {
		return fmt.Errorf("destination path is not a directory: %s", destination)
	}

	equal, err := dirsEqual(source, destination)
	if err != nil {
		return err
	}
	if equal {
		recordRefreshAction(result, "unchanged", rel(root, destination))
		return nil
	}

	recordRefreshAction(result, "updated", rel(root, destination))
	if dryRun {
		return nil
	}
	if err := os.RemoveAll(destination); err != nil {
		return err
	}
	return copyDir(source, destination)
}

type dirEntryFingerprint struct {
	IsDir bool
	Mode  os.FileMode
	Hash  []byte
}

func dirsEqual(left, right string) (bool, error) {
	leftSnapshot, err := snapshotDir(left)
	if err != nil {
		return false, err
	}
	rightSnapshot, err := snapshotDir(right)
	if err != nil {
		return false, err
	}

	if len(leftSnapshot) != len(rightSnapshot) {
		return false, nil
	}

	for path, leftEntry := range leftSnapshot {
		rightEntry, ok := rightSnapshot[path]
		if !ok {
			return false, nil
		}
		if leftEntry.IsDir != rightEntry.IsDir {
			return false, nil
		}
		if leftEntry.Mode != rightEntry.Mode {
			return false, nil
		}
		if !bytes.Equal(leftEntry.Hash, rightEntry.Hash) {
			return false, nil
		}
	}
	return true, nil
}

func snapshotDir(root string) (map[string]dirEntryFingerprint, error) {
	out := map[string]dirEntryFingerprint{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink is not supported in skill directory: %s", path)
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		normalizedRel := filepath.ToSlash(relPath)
		if entry.IsDir() {
			out[normalizedRel] = dirEntryFingerprint{
				IsDir: true,
				Mode:  info.Mode() & os.ModePerm,
			}
			return nil
		}

		hash, err := fileSHA256(path)
		if err != nil {
			return err
		}
		out[normalizedRel] = dirEntryFingerprint{
			IsDir: false,
			Mode:  info.Mode() & os.ModePerm,
			Hash:  hash,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(out))
	for key := range out {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return out, nil
}

func fileSHA256(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return nil, err
	}
	return digest.Sum(nil), nil
}
