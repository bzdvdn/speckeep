package workflow

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"speckeep/src/internal/config"
	"speckeep/src/internal/featurepaths"
	"speckeep/src/internal/project"
)

type RepairResult struct {
	Slug     string   `json:"slug"`
	DryRun   bool     `json:"dry_run"`
	Changed  bool     `json:"changed"`
	Actions  []string `json:"actions,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

type MigrationResult struct {
	DryRun   bool           `json:"dry_run"`
	Changed  bool           `json:"changed"`
	Results  []RepairResult `json:"results,omitempty"`
	Warnings []string       `json:"warnings,omitempty"`
}

func RepairFeature(root, slug string, dryRun bool) (RepairResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return RepairResult{}, err
	}
	cfg, err := config.Load(root)
	if err != nil {
		return RepairResult{}, err
	}

	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return RepairResult{}, err
	}

	result := RepairResult{Slug: slug, DryRun: dryRun}

	if changed, warnings, err := migrateFlatSpecArtifacts(root, specsDir, slug, dryRun, &result.Actions); err != nil {
		return RepairResult{}, err
	} else {
		result.Changed = result.Changed || changed
		result.Warnings = append(result.Warnings, warnings...)
	}

	workspaceDir := firstWorkspaceWithPlans(root)
	if workspaceDir == "" {
		workspaceDir, err = cfg.DraftspecDir(root)
		if err != nil {
			return RepairResult{}, err
		}
	}

	if changed, warnings, err := migrateLegacyPlanDir(root, workspaceDir, specsDir, slug, dryRun, &result.Actions); err != nil {
		return RepairResult{}, err
	} else {
		result.Changed = result.Changed || changed
		result.Warnings = append(result.Warnings, warnings...)
	}

	if !result.Changed && len(result.Warnings) == 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("no safe migrations were needed for slug %s", slug))
	}
	return result, nil
}

func MigrateProject(root string, dryRun bool, copyWorkspace bool) (MigrationResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return MigrationResult{}, err
	}

	result := MigrationResult{DryRun: dryRun}
	workspace := RepairResult{Slug: "__workspace__", DryRun: dryRun}
	workspaceChanged, workspaceWarnings, err := migrateLegacyDraftspecWorkspace(root, dryRun, copyWorkspace, &workspace.Actions)
	if err != nil {
		return MigrationResult{}, err
	}
	workspace.Warnings = append(workspace.Warnings, workspaceWarnings...)
	if workspaceChanged || len(workspace.Warnings) > 0 {
		result.Results = append(result.Results, workspace)
		if workspaceChanged {
			result.Changed = true
		}
	}

	if workspaceChanged && !dryRun {
		if refreshResult, refreshErr := project.Refresh(root, project.RefreshOptions{}); refreshErr == nil {
			if len(refreshResult.Created)+len(refreshResult.Updated)+len(refreshResult.Removed) > 0 {
				result.Results = append(result.Results, RepairResult{
					Slug:    "__refresh__",
					DryRun:  false,
					Changed: true,
					Actions: append([]string(nil), refreshResult.Messages...),
				})
				result.Changed = true
			}
		} else {
			result.Warnings = append(result.Warnings, fmt.Sprintf("refresh after workspace migration failed: %v", refreshErr))
		}
	}

	states, err := States(root)
	if err != nil {
		return MigrationResult{}, err
	}

	slugSet := map[string]struct{}{}
	for _, s := range states {
		slugSet[s.Slug] = struct{}{}
	}
	for _, plansDir := range legacyPlansDirs(root) {
		if entries, err := os.ReadDir(plansDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					slugSet[entry.Name()] = struct{}{}
				}
			}
		}
	}

	slugs := make([]string, 0, len(slugSet))
	for slug := range slugSet {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	for _, slug := range slugs {
		repair, err := RepairFeature(root, slug, dryRun)
		if err != nil {
			return MigrationResult{}, err
		}
		if repair.Changed || len(repair.Warnings) > 0 {
			result.Results = append(result.Results, repair)
		}
		if repair.Changed {
			result.Changed = true
		}
	}
	sort.Slice(result.Results, func(i, j int) bool {
		return result.Results[i].Slug < result.Results[j].Slug
	})
	if len(result.Results) == 0 {
		result.Warnings = append(result.Warnings, "no safe migrations were needed")
	}
	return result, nil
}

func migrateLegacyDraftspecWorkspace(root string, dryRun bool, copyWorkspace bool, actions *[]string) (bool, []string, error) {
	legacyDir := filepath.Join(root, ".draftspec")
	targetDir := filepath.Join(root, ".speckeep")

	if !isDir(legacyDir) {
		return false, nil, nil
	}
	if isDir(targetDir) {
		return false, []string{"both .draftspec and .speckeep exist; workspace migration skipped"}, nil
	}

	mode := "move"
	if copyWorkspace {
		mode = "copy"
	}
	*actions = append(*actions, fmt.Sprintf("%s legacy workspace %s to %s", mode, displayPath(root, legacyDir), displayPath(root, targetDir)))
	if dryRun {
		return true, nil, nil
	}

	if copyWorkspace {
		if err := copyDir(legacyDir, targetDir); err != nil {
			return false, nil, err
		}
	} else {
		if err := os.Rename(legacyDir, targetDir); err != nil {
			// fallback to copy+remove (for cross-device scenarios)
			if err := copyDir(legacyDir, targetDir); err != nil {
				return false, nil, err
			}
			_ = os.RemoveAll(legacyDir)
		}
	}

	// Rename legacy config file name if present.
	legacyCfg := filepath.Join(targetDir, "draftspec.yaml")
	canonicalCfg := filepath.Join(targetDir, "speckeep.yaml")
	if fileExists(legacyCfg) && !fileExists(canonicalCfg) {
		*actions = append(*actions, fmt.Sprintf("rename %s to %s", displayPath(root, legacyCfg), displayPath(root, canonicalCfg)))
		if err := os.Rename(legacyCfg, canonicalCfg); err != nil {
			return true, nil, err
		}
	}

	// Canonicalize config paths from .draftspec -> .speckeep when possible.
	if fileExists(canonicalCfg) {
		cfg, err := config.Load(root)
		if err != nil {
			return true, nil, err
		}
		cfg.Project.ConstitutionFile = rewriteLegacyWorkspacePrefix(cfg.Project.ConstitutionFile)
		cfg.Paths.SpecsDir = rewriteLegacyWorkspacePrefix(cfg.Paths.SpecsDir)
		cfg.Paths.ArchiveDir = rewriteLegacyWorkspacePrefix(cfg.Paths.ArchiveDir)
		cfg.Paths.TemplatesDir = rewriteLegacyWorkspacePrefix(cfg.Paths.TemplatesDir)
		cfg.Paths.ScriptsDir = rewriteLegacyWorkspacePrefix(cfg.Paths.ScriptsDir)

		if err := config.Save(root, cfg); err != nil {
			return true, nil, err
		}
	}

	return true, nil, nil
}

func rewriteLegacyWorkspacePrefix(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, ".draftspec/") {
		return ".speckeep/" + strings.TrimPrefix(value, ".draftspec/")
	}
	if value == ".draftspec" {
		return ".speckeep"
	}
	return value
}

func legacyPlansDirs(root string) []string {
	return []string{
		filepath.Join(root, ".speckeep", "plans"),
		filepath.Join(root, ".draftspec", "plans"),
	}
}

func firstWorkspaceWithPlans(root string) string {
	for _, plansDir := range legacyPlansDirs(root) {
		if isDir(plansDir) {
			return filepath.Dir(plansDir)
		}
	}
	return ""
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func copyDir(src string, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}
		content, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dstPath, content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func displayPath(root, path string) string {
	if rel, err := filepath.Rel(root, path); err == nil {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(path)
}

func migrateFlatSpecArtifacts(root, specsDir, slug string, dryRun bool, actions *[]string) (bool, []string, error) {
	changed := false
	var warnings []string

	for _, artifact := range featurepaths.Artifacts(specsDir, slug) {
		canonicalExists := fileExists(artifact.CanonicalPath)
		legacyExists := fileExists(artifact.LegacyPath)

		switch {
		case !canonicalExists && !legacyExists:
			continue
		case !canonicalExists && legacyExists:
			*actions = append(*actions, fmt.Sprintf("move legacy %s from %s to %s", artifact.Name, displayPath(root, artifact.LegacyPath), displayPath(root, artifact.CanonicalPath)))
			changed = true
			if dryRun {
				continue
			}
			if err := os.MkdirAll(filepath.Dir(artifact.CanonicalPath), 0o755); err != nil {
				return false, nil, err
			}
			if err := os.Rename(artifact.LegacyPath, artifact.CanonicalPath); err != nil {
				return false, nil, fmt.Errorf("move legacy %s for slug %s: %w", artifact.Name, slug, err)
			}
		case canonicalExists && legacyExists:
			canonicalContent, err := os.ReadFile(artifact.CanonicalPath)
			if err != nil {
				return false, nil, fmt.Errorf("read canonical %s for slug %s: %w", artifact.Name, slug, err)
			}
			legacyContent, err := os.ReadFile(artifact.LegacyPath)
			if err != nil {
				return false, nil, fmt.Errorf("read legacy %s for slug %s: %w", artifact.Name, slug, err)
			}
			if !bytes.Equal(canonicalContent, legacyContent) {
				warnings = append(warnings, fmt.Sprintf("canonical and legacy %s differ for slug %s; resolve manually", artifact.Name, slug))
				continue
			}
			*actions = append(*actions, fmt.Sprintf("remove duplicate legacy %s %s", artifact.Name, displayPath(root, artifact.LegacyPath)))
			changed = true
			if dryRun {
				continue
			}
			if err := os.Remove(artifact.LegacyPath); err != nil {
				return false, nil, fmt.Errorf("remove duplicate legacy %s for slug %s: %w", artifact.Name, slug, err)
			}
		}
	}

	if !dryRun {
		_ = removeEmptyDir(featurepaths.SpecDir(specsDir, slug))
	}
	return changed, warnings, nil
}

func removeEmptyDir(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return nil
	}
	return os.Remove(path)
}

// migrateLegacyPlanDir moves plan artifacts from the old workspace plans/<slug>/
// directory layout to the new .speckeep/specs/<slug>/plan/ layout.
func migrateLegacyPlanDir(root, workspaceDir, specsDir, slug string, dryRun bool, actions *[]string) (bool, []string, error) {
	oldPlanDir := filepath.Join(workspaceDir, "plans", slug)
	newPlanDir := featurepaths.PlanDir(specsDir, slug)

	entries, err := os.ReadDir(oldPlanDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("read legacy plan dir for slug %s: %w", slug, err)
	}
	if len(entries) == 0 {
		return false, nil, nil
	}

	changed := false
	var warnings []string

	for _, entry := range entries {
		if entry.IsDir() {
			if entry.Name() == "contracts" {
				c, w, err := migrateLegacyContractsDir(
					root, slug,
					filepath.Join(oldPlanDir, "contracts"),
					filepath.Join(newPlanDir, "contracts"),
					dryRun, actions,
				)
				if err != nil {
					return false, nil, err
				}
				changed = changed || c
				warnings = append(warnings, w...)
			}
			continue
		}

		src := filepath.Join(oldPlanDir, entry.Name())
		dst := filepath.Join(newPlanDir, entry.Name())

		if fileExists(dst) {
			srcContent, err := os.ReadFile(src)
			if err != nil {
				return false, nil, fmt.Errorf("read legacy plan %s for slug %s: %w", entry.Name(), slug, err)
			}
			dstContent, err := os.ReadFile(dst)
			if err != nil {
				return false, nil, fmt.Errorf("read plan %s for slug %s: %w", entry.Name(), slug, err)
			}
			if !bytes.Equal(srcContent, dstContent) {
				warnings = append(warnings, fmt.Sprintf("plan %s already exists at new location for slug %s and differs; resolve manually", entry.Name(), slug))
				continue
			}
			*actions = append(*actions, fmt.Sprintf("remove duplicate legacy plan %s %s", entry.Name(), displayPath(root, src)))
			changed = true
			if !dryRun {
				if err := os.Remove(src); err != nil {
					return false, nil, fmt.Errorf("remove duplicate legacy plan %s for slug %s: %w", entry.Name(), slug, err)
				}
			}
			continue
		}

		*actions = append(*actions, fmt.Sprintf("move legacy plan %s from %s to %s", entry.Name(), displayPath(root, src), displayPath(root, dst)))
		changed = true
		if !dryRun {
			if err := os.MkdirAll(newPlanDir, 0o755); err != nil {
				return false, nil, err
			}
			if err := os.Rename(src, dst); err != nil {
				return false, nil, fmt.Errorf("move legacy plan %s for slug %s: %w", entry.Name(), slug, err)
			}
		}
	}

	if !dryRun {
		_ = removeEmptyDir(filepath.Join(oldPlanDir, "contracts"))
		_ = removeEmptyDir(oldPlanDir)
		_ = removeEmptyDir(filepath.Join(workspaceDir, "plans"))
	}
	return changed, warnings, nil
}

func migrateLegacyContractsDir(root, slug, src, dst string, dryRun bool, actions *[]string) (bool, []string, error) {
	entries, err := os.ReadDir(src)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("read legacy contracts dir for slug %s: %w", slug, err)
	}

	changed := false
	var warnings []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		srcFile := filepath.Join(src, entry.Name())
		dstFile := filepath.Join(dst, entry.Name())

		if fileExists(dstFile) {
			srcContent, _ := os.ReadFile(srcFile)
			dstContent, _ := os.ReadFile(dstFile)
			if !bytes.Equal(srcContent, dstContent) {
				warnings = append(warnings, fmt.Sprintf("plan contracts/%s already exists at new location for slug %s and differs; resolve manually", entry.Name(), slug))
				continue
			}
			*actions = append(*actions, fmt.Sprintf("remove duplicate legacy plan contracts/%s %s", entry.Name(), displayPath(root, srcFile)))
			changed = true
			if !dryRun {
				if err := os.Remove(srcFile); err != nil {
					return false, nil, fmt.Errorf("remove duplicate legacy plan contracts/%s for slug %s: %w", entry.Name(), slug, err)
				}
			}
			continue
		}

		*actions = append(*actions, fmt.Sprintf("move legacy plan contracts/%s from %s to %s", entry.Name(), displayPath(root, srcFile), displayPath(root, dstFile)))
		changed = true
		if !dryRun {
			if err := os.MkdirAll(dst, 0o755); err != nil {
				return false, nil, err
			}
			if err := os.Rename(srcFile, dstFile); err != nil {
				return false, nil, fmt.Errorf("move legacy plan contracts/%s for slug %s: %w", entry.Name(), slug, err)
			}
		}
	}
	return changed, warnings, nil
}
