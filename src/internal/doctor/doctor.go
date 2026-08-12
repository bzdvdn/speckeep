package doctor

import (
	"context"
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
	"speckeep/src/internal/project"
	"speckeep/src/internal/skills"
	"speckeep/src/internal/trace"
	"speckeep/src/internal/workflow"
)

var placeholderPattern = regexp.MustCompile(`\[[A-Z][A-Z0-9_]*\]`)

// deprecatedCommandPattern matches deprecated /speckeep.* slash-command references
// without matching the config path ".speckeep/speckeep.yaml" (the /speckeep. there is
// a path separator, not a command). Commands are slash-command-like tokens such as
// /speckeep.archive /speckeep.spec; looking for a known command suffix keeps the check precise.
var deprecatedCommandPattern = regexp.MustCompile(`/?speckeep\.(?:archive|spec|plan|tasks|inspect|implement|verify|constitution|rollback|recap|repo-map|challenge)\b`)

type Finding struct {
	Level   string
	Message string
}

type Result struct {
	Findings []Finding
}

func Check(ctx context.Context, root string) (Result, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Result{}, err
	}

	if _, err := os.Stat(filepath.Join(root, ".speckeep", "speckeep.yaml")); err != nil {
		if os.IsNotExist(err) {
			return Result{Findings: []Finding{
				{Level: "error", Message: "speckeep project is not initialized — run `speckeep init` in this directory first"},
			}}, nil
		}
		return Result{}, err
	}

	cfg, err := config.Load(ctx, root)
	if err != nil {
		return Result{}, err
	}

	var findings []Finding

	migrationResult, err := workflow.MigrateProject(ctx, root, false, false)
	if err != nil {
		return Result{}, err
	}
	for _, repair := range migrationResult.Results {
		for _, action := range repair.Actions {
			findings = append(findings, Finding{Level: "ok", Message: action})
		}
		for _, warning := range repair.Warnings {
			if warning == "no safe migrations were needed" || strings.HasPrefix(warning, "no safe migrations were needed for slug ") {
				continue
			}
			findings = append(findings, Finding{Level: "warning", Message: warning})
		}
	}
	for _, warning := range migrationResult.Warnings {
		if warning == "no safe migrations were needed" || strings.HasPrefix(warning, "no safe migrations were needed for slug ") {
			continue
		}
		findings = append(findings, Finding{Level: "warning", Message: warning})
	}

	if migrationResult.Changed {
		if _, refreshErr := project.Refresh(root, project.RefreshOptions{}); refreshErr != nil {
			findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("refresh after migration failed: %v", refreshErr)})
		}
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
		filepath.Join(templatesDir, cfg.Templates.VerifyPrompt),
		filepath.Join(scriptsDir, cfg.Scripts.RunSpeckeep),
		filepath.Join(scriptsDir, cfg.Scripts.CheckConstitution),
		filepath.Join(scriptsDir, cfg.Scripts.CheckReady),
		filepath.Join(scriptsDir, cfg.Scripts.VerifyTaskState),
	} {
		checkPath(&findings, path, false)
	}

	agentsPath := filepath.Join(root, cfg.Agents.AgentsFile)
	if content, err := os.ReadFile(agentsPath); err == nil {
		text := string(content)
		if !strings.Contains(text, "/spk.repo-map") {
			findings = append(findings, Finding{
				Level:   "warning",
				Message: "AGENTS.md is missing /spk.repo-map guidance — run `speckeep refresh .` to sync the managed SpecKeep block",
			})
		}
		if deprecatedCommandPattern.MatchString(text) {
			findings = append(findings, Finding{
				Level:   "warning",
				Message: "AGENTS.md still references deprecated /speckeep.* commands — run `speckeep refresh .` to update to /spk.*",
			})
		}
	}

	legacyArchivePrompt := filepath.Join(templatesDir, "prompts", "archive.md")
	if _, err := os.Stat(legacyArchivePrompt); err == nil {
		findings = append(findings, Finding{
			Level:   "warning",
			Message: fmt.Sprintf("legacy archive prompt is still present at %s — archive is CLI-only now; run `speckeep refresh .` to remove it", legacyArchivePrompt),
		})
	}

	constitutionPath := filepath.Join(root, cfg.Project.ConstitutionFile)
	if content, err := os.ReadFile(constitutionPath); err == nil {
		if placeholderPattern.Match(content) {
			findings = append(findings, Finding{
				Level:   "warning",
				Message: "constitution.md contains unfilled placeholder content — run /spk.constitution to complete setup",
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
					Message: "constitution.summary.md not found — run /spk.constitution to generate the compact summary used by spec, inspect, plan, tasks, implement, verify, and hotfix phases",
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
	if warning := layoutWarning(cfg); warning != "" {
		findings = append(findings, Finding{Level: "warning", Message: warning})
	}
	findings = append(findings, legacyNestedPlanFindings(specsDir)...)

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
	for _, relPath := range agents.LegacyArchivePaths() {
		fullPath := filepath.Join(root, filepath.FromSlash(relPath))
		if _, err := os.Stat(fullPath); err == nil {
			findings = append(findings, Finding{
				Level:   "warning",
				Message: fmt.Sprintf("legacy archive agent artifact is no longer needed: %s (run `speckeep refresh .`)", fullPath),
			})
		}
	}

	shell := cfg.Runtime.Shell
	oldPrefixPaths := agents.LegacyPrefixPaths(agents.DefaultCommands(shell))
	oldPrefixSeen := make(map[string]struct{})
	for _, relPath := range oldPrefixPaths {
		normalized := filepath.FromSlash(relPath)
		if _, seen := oldPrefixSeen[normalized]; seen {
			continue
		}
		oldPrefixSeen[normalized] = struct{}{}
		fullPath := filepath.Join(root, normalized)
		if _, err := os.Stat(fullPath); err == nil {
			findings = append(findings, Finding{
				Level:   "warning",
				Message: fmt.Sprintf("deprecated /speckeep.* agent artifact found, rename to /spk.*; run `speckeep refresh .`: %s", fullPath),
			})
		}
	}

	skillsManifest, err := skills.Load(ctx, root)
	if err != nil {
		findings = append(findings, Finding{Level: "error", Message: err.Error()})
	} else {
		skillErrors, skillWarnings := skills.ValidateManifest(context.Background(), root, skillsManifest)
		for _, message := range skillErrors {
			findings = append(findings, Finding{Level: "error", Message: message})
		}
		for _, message := range skillWarnings {
			findings = append(findings, Finding{Level: "warning", Message: message})
		}
	}

	workflowFindings, err := workflow.ValidateProject(ctx, root)
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
	traceFindings, err := traceabilityChecks(ctx, root)
	if err == nil {
		findings = append(findings, traceFindings...)
	}

	// Branching checks
	if branch, err := gitutil.CurrentBranch(ctx, root); err == nil {
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

func layoutWarning(cfg config.Config) string {
	specsDir := strings.TrimSpace(cfg.Paths.SpecsDir)
	archiveDir := strings.TrimSpace(cfg.Paths.ArchiveDir)

	switch {
	case specsDir == "specs" && archiveDir == "archive":
		return "workspace still uses legacy default layout specs/ + archive/ — run `speckeep refresh .` to migrate to specs/active + specs/archived"
	case specsDir == "specs" || archiveDir == "archive":
		return fmt.Sprintf("workspace uses mixed old/new feature layout (specs_dir=%s, archive_dir=%s) — prefer specs/active + specs/archived; run `speckeep refresh .` or set both paths explicitly", specsDir, archiveDir)
	default:
		return ""
	}
}

func legacyNestedPlanFindings(specsDir string) []Finding {
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil
	}

	var findings []Finding
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		legacyPlanDir := filepath.Join(specsDir, entry.Name(), "plan")
		if info, err := os.Stat(legacyPlanDir); err == nil && info.IsDir() {
			findings = append(findings, Finding{
				Level:   "error",
				Message: fmt.Sprintf("legacy nested plan layout still present for slug %s at %s — run `speckeep refresh .` or `speckeep feature repair %s`", entry.Name(), legacyPlanDir, entry.Name()),
			})
		}
	}
	return findings
}

func traceabilityChecks(ctx context.Context, root string) ([]Finding, error) {
	var findings []Finding

	markers, err := trace.FindLegacyMarkers(ctx, root)
	if err != nil {
		return nil, err
	}
	if len(markers) > 0 {
		findings = append(findings, Finding{
			Level:   "warning",
			Message: fmt.Sprintf("%d deprecated traceability marker(s) found in code — remove @sk-task/@sk-test/@ds-* and record evidence as tasks.md Proof entries (speckeep trace reads Proof)", len(markers)),
		})
	}

	entries, err := trace.ParseAll(ctx, root)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return findings, nil
	}

	cfg, err := config.Load(ctx, root)
	if err != nil {
		return nil, err
	}
	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return nil, err
	}

	grouped := map[string][]trace.Entry{}
	for _, entry := range entries {
		grouped[entry.Slug] = append(grouped[entry.Slug], entry)
	}
	slugs := make([]string, 0, len(grouped))
	for slug := range grouped {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	for _, slug := range slugs {
		tasksPath, _ := featurepaths.ResolveTasks(specsDir, slug)
		taskIDs := map[string]struct{}{}
		if content, err := os.ReadFile(tasksPath); err == nil {
			for _, id := range taskIDsFromFile(string(content)) {
				taskIDs[id] = struct{}{}
			}
		}
		specPath, _ := featurepaths.ResolveSpec(specsDir, slug)
		specContent, _ := os.ReadFile(specPath)

		for _, entry := range grouped[slug] {
			name := slug + "#" + entry.TaskID
			if _, ok := taskIDs[entry.TaskID]; !ok {
				findings = append(findings, Finding{
					Level:   "warning",
					Message: fmt.Sprintf("orphaned proof entry %s: task not found in tasks.md", name),
				})
			}
			if entry.ACID != "" && !strings.Contains(string(specContent), entry.ACID) {
				findings = append(findings, Finding{
					Level:   "warning",
					Message: fmt.Sprintf("invalid proof entry %s: AC %s not found in spec for slug %s", name, entry.ACID, slug),
				})
			}
			for _, problem := range trace.CheckEntry(root, entry) {
				findings = append(findings, Finding{Level: "warning", Message: problem.Message})
			}
		}
	}

	return findings, nil
}

var taskIDRegex = regexp.MustCompile(`(T[0-9]+(?:\.[0-9]+)*)`)

func taskIDsFromFile(content string) []string {
	return taskIDRegex.FindAllString(content, -1)
}

func speckeepEntrypointWarning(root string) string {
	configured := strings.TrimSpace(os.Getenv("SPECKEEP_BIN"))
	if configured != "" {
		if _, err := resolveSpecgateBinary(root, configured); err != nil {
			return fmt.Sprintf("SPECKEEP_BIN could not be resolved: %s", configured)
		}
		return ""
	}

	if _, err := exec.LookPath("speckeep"); err == nil {
		return ""
	}
	return "speckeep CLI entrypoint not found; set SPECKEEP_BIN or add speckeep to PATH"
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
