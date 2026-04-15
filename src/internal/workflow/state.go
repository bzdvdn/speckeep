package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"speckeep/src/internal/config"
	"speckeep/src/internal/featurepaths"
	"speckeep/src/internal/gitutil"
)

type FeatureState struct {
	Slug           string `json:"slug"`
	Phase          string `json:"phase"`
	SpecExists     bool   `json:"spec_exists"`
	HotfixExists   bool   `json:"hotfix_exists"`
	InspectExists  bool   `json:"inspect_exists"`
	PlanExists     bool   `json:"plan_exists"`
	TasksExists    bool   `json:"tasks_exists"`
	VerifyExists   bool   `json:"verify_exists"`
	Archived       bool   `json:"archived"`
	InspectPath    string `json:"inspect_path,omitempty"`
	InspectLegacy  bool   `json:"inspect_legacy,omitempty"`
	VerifyPath     string `json:"verify_path,omitempty"`
	InspectStatus  string `json:"inspect_status,omitempty"`
	VerifyStatus   string `json:"verify_status,omitempty"`
	TasksTotal     int    `json:"tasks_total"`
	TasksCompleted int    `json:"tasks_completed"`
	TasksOpen      int    `json:"tasks_open"`
	ReadyFor       string `json:"ready_for,omitempty"`
	Blocked        bool   `json:"blocked"`
	CurrentBranch  string `json:"current_branch,omitempty"`
	BranchMismatch bool   `json:"branch_mismatch,omitempty"`
}

var checkboxPattern = regexp.MustCompile(`^\s*- \[([ x])\]`)

func State(root, slug string) (FeatureState, error) {
	if slug == "" {
		return FeatureState{}, fmt.Errorf("slug cannot be empty")
	}

	cfg, err := config.Load(root)
	if err != nil {
		return FeatureState{}, err
	}

	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return FeatureState{}, err
	}
	archiveDir, err := cfg.ArchiveDir(root)
	if err != nil {
		return FeatureState{}, err
	}

	specPath, _ := featurepaths.ResolveSpec(specsDir, slug)
	hotfixPath, _ := featurepaths.ResolveHotfix(specsDir, slug)
	inspectPath, inspectLegacyFlat := featurepaths.ResolveInspect(specsDir, slug)
	legacyInspectPath := featurepaths.Inspect(specsDir, slug)
	planPath := featurepaths.Plan(specsDir, slug)
	tasksPath := featurepaths.Tasks(specsDir, slug)
	verifyPath := featurepaths.Verify(specsDir, slug)
	archiveSlugDir := filepath.Join(archiveDir, slug)

	state := FeatureState{
		Slug:         slug,
		SpecExists:   fileExists(specPath),
		HotfixExists: fileExists(hotfixPath),
		PlanExists:   fileExists(planPath),
		TasksExists:  fileExists(tasksPath),
		VerifyPath:   verifyPath,
	}
	state.InspectExists, state.InspectPath, state.InspectLegacy = existingInspectReportPath(inspectPath, legacyInspectPath)
	if state.InspectExists && inspectLegacyFlat && state.InspectPath == inspectPath {
		state.InspectLegacy = true
	}
	state.VerifyExists = fileExists(verifyPath)

	if state.TasksExists {
		total, completed, open, err := taskCounts(tasksPath)
		if err != nil {
			return FeatureState{}, err
		}
		state.TasksTotal = total
		state.TasksCompleted = completed
		state.TasksOpen = open
	}

	state.Archived = archiveExists(archiveSlugDir)
	state.InspectStatus, _ = reportStatus(state.InspectPath)
	state.VerifyStatus, _ = reportStatus(state.VerifyPath)

	if branch, err := gitutil.CurrentBranch(root); err == nil {
		state.CurrentBranch = branch
		if !state.Archived && (state.SpecExists || state.HotfixExists) {
			expected := "feature/" + slug
			if !state.SpecExists && state.HotfixExists {
				expected = "hotfix/" + slug
			}
			if branch != expected && branch != "main" && branch != "master" {
				state.BranchMismatch = true
			}
		}
	}

	inferLifecycle(&state)

	return state, nil
}

func States(root string) ([]FeatureState, error) {
	cfg, err := config.Load(root)
	if err != nil {
		return nil, err
	}

	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return nil, err
	}
	archiveDir, err := cfg.ArchiveDir(root)
	if err != nil {
		return nil, err
	}

	slugSet := map[string]struct{}{}
	collectSpecSlugs(slugSet, specsDir)
	collectDirSlugs(slugSet, archiveDir)

	slugs := make([]string, 0, len(slugSet))
	for slug := range slugSet {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	results := make([]FeatureState, 0, len(slugs))
	for _, slug := range slugs {
		state, err := State(root, slug)
		if err != nil {
			return nil, err
		}
		results = append(results, state)
	}

	return results, nil
}

func inferLifecycle(state *FeatureState) {
	hasValidInspect := state.InspectExists && ValidStatus(state.InspectStatus)
	hasValidVerify := state.VerifyExists && ValidStatus(state.VerifyStatus)

	switch {
	case state.Archived:
		state.Phase = "archive"
	case !state.SpecExists && state.HotfixExists:
		state.Phase = "hotfix"
		switch {
		case !hasValidVerify:
			state.ReadyFor = "hotfix"
		case state.VerifyStatus == StatusBlocked:
			state.ReadyFor = "hotfix"
			state.Blocked = true
		default:
			state.ReadyFor = "archive"
		}
	case !state.SpecExists:
		state.Phase = "constitution"
		state.ReadyFor = "spec"
		state.Blocked = true
	case !hasValidInspect:
		state.Phase = "spec"
		state.ReadyFor = "inspect"
	case state.InspectStatus == StatusBlocked:
		state.Phase = "inspect"
		state.ReadyFor = "inspect"
		state.Blocked = true
	case !state.PlanExists:
		state.Phase = "inspect"
		state.ReadyFor = "plan"
	case !state.TasksExists:
		state.Phase = "plan"
		state.ReadyFor = "tasks"
	case state.TasksTotal == 0:
		// tasks.md exists but contains no checkboxes — treat as empty/incomplete
		state.Phase = "plan"
		state.ReadyFor = "tasks"
	case state.TasksOpen > 0:
		state.Phase = "implement"
		state.ReadyFor = "implement"
	case !hasValidVerify:
		state.Phase = "verify"
		state.ReadyFor = "verify"
	case state.VerifyStatus == StatusBlocked:
		state.Phase = "verify"
		state.ReadyFor = "verify"
		state.Blocked = true
	default:
		state.Phase = "verify"
		state.ReadyFor = "archive"
	}

	if state.BranchMismatch && !state.Archived {
		state.Blocked = true
	}
}

func reportStatus(path string) (string, error) {
	if path == "" || !fileExists(path) {
		return "", nil
	}

	report, err := ParseReport(path)
	if err != nil {
		return "", err
	}
	return report.Status, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func archiveExists(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	return len(entries) > 0
}

func taskCounts(path string) (total int, completed int, open int, err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("read tasks file: %w", err)
	}

	lines := regexp.MustCompile(`\r?\n`).Split(string(content), -1)
	for _, line := range lines {
		match := checkboxPattern.FindStringSubmatch(line)
		if len(match) == 0 {
			continue
		}
		total++
		if match[1] == "x" {
			completed++
		}
		if match[1] == " " {
			open++
		}
	}

	return total, completed, open, nil
}

func existingInspectReportPath(canonicalPath, legacyPath string) (bool, string, bool) {
	switch {
	case fileExists(canonicalPath):
		return true, canonicalPath, false
	case fileExists(legacyPath):
		return true, legacyPath, true
	default:
		return false, canonicalPath, false
	}
}

func collectSpecSlugs(slugSet map[string]struct{}, dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if specDirHasArtifacts(filepath.Join(dir, entry.Name())) {
				slugSet[entry.Name()] = struct{}{}
			}
			continue
		}
		name := entry.Name()
		if slug, ok := legacySpecArtifactSlug(name); ok {
			slugSet[slug] = struct{}{}
		}
	}
}

func specDirHasArtifacts(dir string) bool {
	for _, name := range []string{"spec.md", "inspect.md", "summary.md", "hotfix.md"} {
		if fileExists(filepath.Join(dir, name)) {
			return true
		}
	}
	return false
}

func legacySpecArtifactSlug(name string) (string, bool) {
	if !strings.HasSuffix(name, ".md") {
		return "", false
	}
	switch {
	case strings.HasSuffix(name, ".inspect.md"):
		return strings.TrimSuffix(name, ".inspect.md"), true
	case strings.HasSuffix(name, ".summary.md"):
		return strings.TrimSuffix(name, ".summary.md"), true
	case strings.HasSuffix(name, ".hotfix.md"):
		return strings.TrimSuffix(name, ".hotfix.md"), true
	default:
		return strings.TrimSuffix(name, ".md"), true
	}
}

func collectDirSlugs(slugSet map[string]struct{}, dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		slugSet[entry.Name()] = struct{}{}
	}
}
