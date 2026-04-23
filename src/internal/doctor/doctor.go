package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"speckeep/src/internal/agents"
	"speckeep/src/internal/config"
	"speckeep/src/internal/featurepaths"
	"speckeep/src/internal/gitutil"
	"speckeep/src/internal/trace"
	"speckeep/src/internal/workflow"
)

var placeholderPattern = regexp.MustCompile(`\[[A-Z][A-Z0-9_]*\]`)

type Finding struct {
	Level   string
	Message string
}

type Result struct {
	Findings []Finding
}

func Check(root string) (Result, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Result{}, err
	}

	cfg, err := config.Load(root)
	if err != nil {
		return Result{}, err
	}

	var findings []Finding

	migrationResult, err := workflow.MigrateProject(root, false, false)
	if err != nil {
		return Result{}, err
	}
	for _, repair := range migrationResult.Results {
		for _, action := range repair.Actions {
			findings = append(findings, Finding{Level: "ok", Message: action})
		}
		for _, warning := range repair.Warnings {
			findings = append(findings, Finding{Level: "warning", Message: warning})
		}
	}
	for _, warning := range migrationResult.Warnings {
		if warning == "no safe migrations were needed" {
			continue
		}
		findings = append(findings, Finding{Level: "warning", Message: warning})
	}

	draftspecDir, err := cfg.DraftspecDir(root)
	if err != nil {
		return Result{}, err
	}
	configPath, err := cfg.ConfigPath(root)
	if err != nil {
		return Result{}, err
	}
	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return Result{}, err
	}
	archiveDir, err := cfg.ArchiveDir(root)
	if err != nil {
		return Result{}, err
	}
	templatesDir, err := cfg.TemplatesDir(root)
	if err != nil {
		return Result{}, err
	}
	scriptsDir, err := cfg.ScriptsDir(root)
	if err != nil {
		return Result{}, err
	}

	for _, path := range []string{draftspecDir, specsDir, archiveDir, templatesDir, scriptsDir} {
		checkPath(&findings, path, true)
	}
	for _, path := range []string{
		configPath,
		filepath.Join(root, cfg.Project.ConstitutionFile),
		filepath.Join(templatesDir, cfg.Templates.Spec),
		filepath.Join(templatesDir, cfg.Templates.Plan),
		filepath.Join(templatesDir, cfg.Templates.Tasks),
		filepath.Join(templatesDir, cfg.Templates.InspectReport),
		filepath.Join(templatesDir, cfg.Templates.VerifyReport),
		filepath.Join(templatesDir, cfg.Templates.ConstitutionPrompt),
		filepath.Join(templatesDir, cfg.Templates.SpecPrompt),
		filepath.Join(templatesDir, cfg.Templates.InspectPrompt),
		filepath.Join(templatesDir, cfg.Templates.PlanPrompt),
		filepath.Join(templatesDir, cfg.Templates.TasksPrompt),
		filepath.Join(templatesDir, cfg.Templates.ImplementPrompt),
		filepath.Join(templatesDir, cfg.Templates.ArchivePrompt),
		filepath.Join(templatesDir, cfg.Templates.VerifyPrompt),
		filepath.Join(scriptsDir, cfg.Scripts.RunSpeckeep),
		filepath.Join(scriptsDir, cfg.Scripts.CheckConstitution),
		filepath.Join(scriptsDir, cfg.Scripts.CheckSpecReady),
		filepath.Join(scriptsDir, cfg.Scripts.CheckInspectReady),
		filepath.Join(scriptsDir, cfg.Scripts.CheckPlanReady),
		filepath.Join(scriptsDir, cfg.Scripts.CheckTasksReady),
		filepath.Join(scriptsDir, cfg.Scripts.CheckImplementReady),
		filepath.Join(scriptsDir, cfg.Scripts.CheckArchiveReady),
		filepath.Join(scriptsDir, cfg.Scripts.CheckVerifyReady),
		filepath.Join(scriptsDir, cfg.Scripts.VerifyTaskState),
	} {
		checkPath(&findings, path, false)
	}

	constitutionPath := filepath.Join(root, cfg.Project.ConstitutionFile)
	if content, err := os.ReadFile(constitutionPath); err == nil {
		if placeholderPattern.Match(content) {
			findings = append(findings, Finding{
				Level:   "warning",
				Message: "constitution.md contains unfilled placeholder content — run /speckeep.constitution to complete setup",
			})
		}
		summaryPath := filepath.Join(draftspecDir, "constitution.summary.md")
		hasActiveSpecs, err := hasActiveSpecs(specsDir)
		if err != nil {
			findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("read specs directory: %v", err)})
		} else if hasActiveSpecs {
			if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
				findings = append(findings, Finding{
					Level:   "warning",
					Message: "constitution.summary.md not found — run /speckeep.constitution to generate the compact summary used by spec, inspect, plan, tasks, implement, verify, and hotfix phases",
				})
			}
		}
	}

	if cfg.Language.Default != "en" && cfg.Language.Default != "ru" {
		findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("unsupported default language: %s", cfg.Language.Default)})
	}
	for _, value := range []string{cfg.Language.Docs, cfg.Language.Agent, cfg.Language.Comments} {
		if value != "en" && value != "ru" {
			findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("unsupported configured language: %s", value)})
		}
	}
	if _, err := config.NormalizeShell(cfg.Runtime.Shell); err != nil {
		findings = append(findings, Finding{Level: "error", Message: err.Error()})
	}
	if warning := speckeepEntrypointWarning(root); warning != "" {
		findings = append(findings, Finding{Level: "warning", Message: warning})
	}

	enabledTargets := map[string]struct{}{}
	for _, target := range cfg.Agents.Targets {
		enabledTargets[target] = struct{}{}
		paths, err := agents.PathsForTarget(target)
		if err != nil {
			findings = append(findings, Finding{Level: "error", Message: err.Error()})
			continue
		}
		for _, relPath := range paths {
			checkPath(&findings, filepath.Join(root, filepath.FromSlash(relPath)), false)
		}
	}

	for _, target := range agents.SupportedTargets() {
		if _, ok := enabledTargets[target]; ok {
			continue
		}
		paths, err := agents.PathsForTarget(target)
		if err != nil {
			continue
		}
		for _, relPath := range paths {
			fullPath := filepath.Join(root, filepath.FromSlash(relPath))
			if _, err := os.Stat(fullPath); err == nil {
				findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("orphaned agent artifact for disabled target %s: %s", target, fullPath)})
			}
		}
	}

	workflowFindings, err := workflow.ValidateProject(root)
	if err != nil {
		findings = append(findings, Finding{Level: "error", Message: err.Error()})
	} else {
		for _, finding := range workflowFindings {
			findings = append(findings, Finding{Level: finding.Level, Message: finding.Message})
		}
	}

	hasErrors := false
	for _, finding := range findings {
		if finding.Level == "error" {
			hasErrors = true
			break
		}
	}
	if hasErrors {
		findings = append(findings, Finding{Level: "error", Message: "speckeep workspace has critical errors"})
	} else {
		findings = append(findings, Finding{Level: "ok", Message: "speckeep workspace looks healthy"})
	}

	// Traceability checks
	traceFindings, err := traceabilityChecks(root)
	if err == nil {
		findings = append(findings, traceFindings...)
	}

	// Branching checks
	if branch, err := gitutil.CurrentBranch(root); err == nil {
		if branch != "main" && branch != "master" && !strings.HasPrefix(branch, "feature/") && !strings.HasPrefix(branch, "hotfix/") {
			findings = append(findings, Finding{
				Level:   "warning",
				Message: fmt.Sprintf("working on non-standard branch: %s (expected main, master, feature/*, or hotfix/*)", branch),
			})
		}
	}

	sort.Slice(findings, func(i, j int) bool {
		ri := severityRank(findings[i].Level)
		rj := severityRank(findings[j].Level)
		if ri != rj {
			return ri < rj
		}
		if findings[i].Message != findings[j].Message {
			return findings[i].Message < findings[j].Message
		}
		return findings[i].Level < findings[j].Level
	})
	return Result{Findings: findings}, nil
}

func hasActiveSpecs(specsDir string) (bool, error) {
	info, err := os.Stat(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !info.IsDir() {
		return false, nil
	}
	slugs, err := featurepaths.ListSpecSlugs(specsDir)
	if err != nil {
		return false, err
	}
	return len(slugs) > 0, nil
}

func checkPath(findings *[]Finding, path string, expectDir bool) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			*findings = append(*findings, Finding{Level: "error", Message: fmt.Sprintf("missing %s", path)})
			return
		}
		*findings = append(*findings, Finding{Level: "error", Message: fmt.Sprintf("failed to stat %s: %v", path, err)})
		return
	}
	if expectDir && !info.IsDir() {
		*findings = append(*findings, Finding{Level: "error", Message: fmt.Sprintf("expected directory: %s", path)})
		return
	}
	if !expectDir && info.IsDir() {
		*findings = append(*findings, Finding{Level: "error", Message: fmt.Sprintf("expected file: %s", path)})
	}
}

func severityRank(level string) int {
	switch level {
	case "error":
		return 1
	case "warning":
		return 2
	case "ok":
		return 3
	default:
		return 4
	}
}

func traceabilityChecks(root string) ([]Finding, error) {
	var findings []Finding

	traceResult, err := trace.Scan(root)
	if err != nil {
		return nil, err
	}

	if len(traceResult.Findings) == 0 {
		return []Finding{{Level: "warning", Message: "no traceability annotations (@sk-task / @sk-test) found in codebase"}}, nil
	}

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

	states, err := workflow.States(root)
	if err != nil {
		return nil, err
	}

	allTaskIDs := make(map[string]string) // taskID -> slug
	for _, state := range states {
		if state.TasksExists {
			tasksPath := featurepaths.Tasks(specsDir, state.Slug)
			taskIDs, err := taskIDsFromFile(tasksPath)
			if err == nil {
				for _, id := range taskIDs {
					allTaskIDs[id] = state.Slug
				}
			}
		}
		if state.Archived {
			taskIDs, err := taskIDsFromLatestArchive(archiveDir, state.Slug)
			if err == nil {
				for _, id := range taskIDs {
					allTaskIDs[id] = state.Slug
				}
			}
		}
	}

	for _, f := range traceResult.Findings {
		slug, ok := allTaskIDs[f.TaskID]
		if !ok {
			findings = append(findings, Finding{
				Level:   "warning",
				Message: fmt.Sprintf("orphaned traceability annotation in %s:%d: task %s not found in any tasks.md", f.File, f.Line, f.TaskID),
			})
			continue
		}

		if f.ACID != "" {
			// Check if AC ID exists in the spec for this slug
			specPath, _ := featurepaths.ResolveSpec(specsDir, slug)
			if content, err := os.ReadFile(specPath); err == nil {
				if !strings.Contains(string(content), f.ACID) {
					findings = append(findings, Finding{
						Level:   "warning",
						Message: fmt.Sprintf("invalid traceability annotation in %s:%d: AC %s not found in spec for slug %s", f.File, f.Line, f.ACID, slug),
					})
				}
			}
		}
	}

	return findings, nil
}

var taskIDRegex = regexp.MustCompile(`(T[0-9]+(?:\.[0-9]+)*)`)

func taskIDsFromFile(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return taskIDRegex.FindAllString(string(content), -1), nil
}

func taskIDsFromLatestArchive(archiveDir, slug string) ([]string, error) {
	slugDir := filepath.Join(archiveDir, slug)
	entries, err := os.ReadDir(slugDir)
	if err != nil {
		return nil, err
	}

	latest := ""
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() > latest {
			latest = entry.Name()
		}
	}
	if latest == "" {
		return nil, os.ErrNotExist
	}

	tasksPath := filepath.Join(slugDir, latest, "plan", "tasks.md")
	return taskIDsFromFile(tasksPath)
}

func speckeepEntrypointWarning(root string) string {
	configured := strings.TrimSpace(os.Getenv("SPECKEEP_BIN"))
	if configured != "" {
		if _, err := resolveSpecgateBinary(root, configured); err != nil {
			return fmt.Sprintf("SPECKEEP_BIN could not be resolved: %s", configured)
		}
		return ""
	}

	legacyConfigured := strings.TrimSpace(os.Getenv("DRAFTSPEC_BIN"))
	if legacyConfigured != "" {
		if _, err := resolveSpecgateBinary(root, legacyConfigured); err != nil {
			return fmt.Sprintf("DRAFTSPEC_BIN could not be resolved: %s", legacyConfigured)
		}
		return ""
	}

	if _, err := exec.LookPath("speckeep"); err == nil {
		return ""
	}
	if _, err := exec.LookPath("draftspec"); err == nil {
		return ""
	}
	return "speckeep CLI entrypoint not found; set SPECKEEP_BIN (or legacy DRAFTSPEC_BIN) or add speckeep to PATH"
}

func resolveSpecgateBinary(root, value string) (string, error) {
	if strings.ContainsAny(value, `/\`) || filepath.IsAbs(value) {
		candidate := value
		if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(root, candidate)
		}
		info, err := os.Stat(candidate)
		if err != nil {
			return "", err
		}
		if info.IsDir() {
			return "", fmt.Errorf("configured path is a directory")
		}
		if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
			return "", fmt.Errorf("configured path is not executable")
		}
		return candidate, nil
	}
	return exec.LookPath(value)
}
