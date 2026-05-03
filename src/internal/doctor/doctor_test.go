package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"speckeep/src/internal/config"
	"speckeep/src/internal/project"
)

func doctorSpecsDir(t *testing.T, root string) string {
	t.Helper()
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("config.Load returned error: %v", err)
	}
	dir, err := cfg.SpecsDir(root)
	if err != nil {
		t.Fatalf("cfg.SpecsDir returned error: %v", err)
	}
	return dir
}

func TestCheckHealthyWorkspace(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh", AgentTargets: []string{"claude"}})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	if len(result.Findings) == 0 {
		t.Fatal("expected findings, got none")
	}
	if result.Findings[len(result.Findings)-1].Level != "ok" {
		t.Fatalf("last finding level = %q, want ok", result.Findings[len(result.Findings)-1].Level)
	}
}

func TestCheckErrorsWhenPlanSkipsMandatoryInspect(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specDir := filepath.Join(doctorSpecsDir(t, root), "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(specDir) returned error: %v", err)
	}
	specPath := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specPath, []byte("# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 Demo\n- **Given** x\n- **When** y\n- **Then** z\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	planDir := filepath.Join(doctorSpecsDir(t, root), "demo", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(planDir) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "plan.md"), []byte("# Demo Plan\n\n- DEC-001 x\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(plan) returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	var found bool
	for _, finding := range result.Findings {
		if finding.Level == "error" && strings.Contains(finding.Message, "mandatory inspect report") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected missing inspect error, got %+v", result.Findings)
	}
}

func TestCheckWarnsAboutOrphanedAgentArtifact(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh", AgentTargets: []string{"claude", "cursor"}})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}
	_, err = project.RemoveAgents(root, project.RemoveAgentsOptions{Targets: []string{"cursor"}})
	if err != nil {
		t.Fatalf("RemoveAgents returned error: %v", err)
	}

	orphanPath := filepath.Join(root, ".cursor", "rules", "speckeep-inspect.mdc")
	if err := os.MkdirAll(filepath.Dir(orphanPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(orphanPath, []byte("orphan"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	var foundWarning bool
	for _, finding := range result.Findings {
		if finding.Level == "warning" && strings.Contains(finding.Message, "orphaned agent artifact") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Fatalf("expected orphaned artifact warning, got %+v", result.Findings)
	}
}

func TestCheckWarnsAboutLegacyArchiveArtifacts(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh", AgentTargets: []string{"claude"}})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	legacyPrompt := filepath.Join(root, ".speckeep", "templates", "prompts", "archive.md")
	if err := os.MkdirAll(filepath.Dir(legacyPrompt), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(legacyPrompt, []byte("legacy archive prompt"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	legacyAgent := filepath.Join(root, ".claude", "commands", "speckeep.archive.md")
	if err := os.MkdirAll(filepath.Dir(legacyAgent), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(legacyAgent, []byte("legacy archive command"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	var foundPromptWarning bool
	var foundAgentWarning bool
	for _, finding := range result.Findings {
		if finding.Level == "warning" && strings.Contains(finding.Message, "legacy archive prompt") {
			foundPromptWarning = true
		}
		if finding.Level == "warning" && strings.Contains(finding.Message, "legacy archive agent artifact") {
			foundAgentWarning = true
		}
	}
	if !foundPromptWarning || !foundAgentWarning {
		t.Fatalf("expected legacy archive warnings, got %+v", result.Findings)
	}
}

func TestCheckErrorsWhenRequiredFileIsMissing(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh", AgentTargets: []string{"claude"}})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	missingPath := filepath.Join(root, "CONSTITUTION.md")
	if err := os.Remove(missingPath); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	if len(result.Findings) == 0 || result.Findings[0].Level != "error" {
		t.Fatalf("expected first finding to be error, got %+v", result.Findings)
	}

	var foundMissing bool
	for _, finding := range result.Findings {
		if finding.Level == "error" && strings.Contains(finding.Message, "missing") && strings.Contains(finding.Message, "CONSTITUTION.md") {
			foundMissing = true
			break
		}
	}
	if !foundMissing {
		t.Fatalf("expected missing constitution error, got %+v", result.Findings)
	}
}

func TestCheckHealthyPowerShellWorkspace(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "powershell"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if result.Findings[len(result.Findings)-1].Level != "ok" {
		t.Fatalf("last finding level = %q, want ok", result.Findings[len(result.Findings)-1].Level)
	}
}

func TestCheckErrorsOnUnsupportedShell(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("config.Load returned error: %v", err)
	}
	cfg.Runtime.Shell = "fish"
	if err := config.Save(root, cfg); err != nil {
		t.Fatalf("config.Save returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	var found bool
	for _, finding := range result.Findings {
		if finding.Level == "error" && strings.Contains(finding.Message, "unsupported shell") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected unsupported shell error, got %+v", result.Findings)
	}
}

func TestCheckWarnsWhenConstitutionHasPlaceholders(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	var found bool
	for _, finding := range result.Findings {
		if finding.Level == "warning" && strings.Contains(finding.Message, "unfilled placeholder") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected unfilled placeholder warning after init, got %+v", result.Findings)
	}
}

func TestCheckNoPlaceholderWarnWhenConstitutionIsFilled(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	constitutionPath := filepath.Join(root, "CONSTITUTION.md")
	filled := "# My Project Constitution\n\n## Purpose\nBuild a great product.\n\n## Core Principles\n\n### Simplicity\nKeep it simple.\n\n## Constraints\nNo magic.\n\n## Decision Priorities\n- Correctness first\n\n## Key Quality Dimensions\n- Tested\n\n## Language Policy\n- Documentation language: English\n- Agent interaction language: English\n- Code comment language: English\n\n## Development Workflow\nUse feature branches.\n\n## Governance\nConstitution is authoritative.\n\n## Exceptions Protocol\nRecord deviations explicitly.\n\n## Last Updated\n2026-04-03\n"
	if err := os.WriteFile(constitutionPath, []byte(filled), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	for _, finding := range result.Findings {
		if finding.Level == "warning" && strings.Contains(finding.Message, "unfilled placeholder") {
			t.Fatalf("unexpected unfilled placeholder warning when constitution is filled: %+v", result.Findings)
		}
	}
}

func TestCheckWarnsDuplicateStableIDsAcrossSpecs(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specsDir := doctorSpecsDir(t, root)
	specA := "# Feature A\n\n## Goal\nDo A.\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 A\n- **Given** x\n- **When** y\n- **Then** z\n"
	specB := "# Feature B\n\n## Goal\nDo B.\n\n## Requirements\n- RQ-001 y\n\n## Acceptance Criteria\n### AC-001 B\n- **Given** a\n- **When** b\n- **Then** c\n"

	if err := os.MkdirAll(filepath.Join(specsDir, "feature-a"), 0o755); err != nil {
		t.Fatalf("MkdirAll(feature-a) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "feature-a", "spec.md"), []byte(specA), 0o644); err != nil {
		t.Fatalf("WriteFile(feature-a) returned error: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(specsDir, "feature-b"), 0o755); err != nil {
		t.Fatalf("MkdirAll(feature-b) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "feature-b", "spec.md"), []byte(specB), 0o644); err != nil {
		t.Fatalf("WriteFile(feature-b) returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	var foundAC, foundRQ bool
	for _, finding := range result.Findings {
		if finding.Level == "warning" && strings.Contains(finding.Message, "AC-001") && strings.Contains(finding.Message, "multiple specs") {
			foundAC = true
		}
		if finding.Level == "warning" && strings.Contains(finding.Message, "RQ-001") && strings.Contains(finding.Message, "multiple specs") {
			foundRQ = true
		}
	}
	if !foundAC {
		t.Fatalf("expected AC-001 duplicate warning, got %+v", result.Findings)
	}
	if !foundRQ {
		t.Fatalf("expected RQ-001 duplicate warning, got %+v", result.Findings)
	}
}

func TestCheckErrorsOnInvalidSkillsManifestEntry(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	manifestPath := filepath.Join(root, ".speckeep", "skills", "manifest.yaml")
	manifest := "version: 1\nskills:\n  - id: invalid-skill\n    enabled: true\n    source: git\n    location: https://example.com/skills.git\n    ref: main\n"
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o644); err != nil {
		t.Fatalf("WriteFile(manifest) returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	var found bool
	for _, finding := range result.Findings {
		if finding.Level == "error" && strings.Contains(finding.Message, "floating git ref") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected floating git ref error, got %+v", result.Findings)
	}
}

func TestCheckWarnsWhenSkillMissingResolvedCommit(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	manifestPath := filepath.Join(root, ".speckeep", "skills", "manifest.yaml")
	manifest := "version: 1\nskills:\n  - id: git-skill\n    enabled: true\n    source: git\n    location: https://example.com/skills.git\n    ref: v1.0.0\n"
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o644); err != nil {
		t.Fatalf("WriteFile(manifest) returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	var found bool
	for _, finding := range result.Findings {
		if finding.Level == "warning" && strings.Contains(finding.Message, "no resolved_commit") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected resolved_commit warning, got %+v", result.Findings)
	}
}

func TestCheckDoesNotWarnOrphanedTraceabilityWhenTaskIsInArchive(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	// Code annotation for an archived feature.
	codePath := filepath.Join(root, "pkg", "demo", "demo_test.go")
	if err := os.MkdirAll(filepath.Dir(codePath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(codePath, []byte("package demo\n\n// @sk-task T4.1: Archived implementation (AC-001)\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	// Archive snapshot containing the task ID.
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("config.Load returned error: %v", err)
	}
	archiveDir, err := cfg.ArchiveDir(root)
	if err != nil {
		t.Fatalf("cfg.ArchiveDir returned error: %v", err)
	}
	archiveTasks := filepath.Join(archiveDir, "demo", "2026-01-01", "plan", "tasks.md")
	if err := os.MkdirAll(filepath.Dir(archiveTasks), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(archiveTasks, []byte("# Tasks\n\n- [x] T4.1 done\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	for _, finding := range result.Findings {
		if finding.Level == "warning" && strings.Contains(finding.Message, "orphaned traceability annotation") {
			t.Fatalf("unexpected orphaned traceability warning: %+v", finding)
		}
	}
}

func TestCheckWarnsWhenSpecgateEntrypointCannotBeResolved(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PATH", "")
	t.Setenv("DRAFTSPEC_BIN", "")
	t.Setenv("SPECKEEP_BIN", "")

	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	var foundWarning bool
	for _, finding := range result.Findings {
		if finding.Level == "warning" && strings.Contains(finding.Message, "set SPECKEEP_BIN") && strings.Contains(finding.Message, "add speckeep to PATH") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Fatalf("expected missing entrypoint warning, got %+v", result.Findings)
	}
}

func TestCheckDoesNotWarnMissingConstitutionSummaryWhenNoActiveSpecsAndCustomPaths(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:          false,
		DefaultLang:      "en",
		Shell:            "sh",
		SpecsDir:         ".speckeep/specifications",
		ArchiveDir:       ".speckeep/artifacts/archive",
		ConstitutionFile: "docs/project-constitution.md",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	for _, finding := range result.Findings {
		if finding.Level == "warning" && strings.Contains(finding.Message, "constitution.summary.md not found") {
			t.Fatalf("unexpected constitution summary warning with no active specs: %+v", result.Findings)
		}
	}
}

func TestCheckUsesWorkspaceConstitutionSummaryPathForCustomConstitutionFile(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:          false,
		DefaultLang:      "en",
		Shell:            "sh",
		SpecsDir:         ".speckeep/specifications",
		ArchiveDir:       ".speckeep/artifacts/archive",
		ConstitutionFile: "docs/project-constitution.md",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("config.Load returned error: %v", err)
	}
	customSpecsDir, err := cfg.SpecsDir(root)
	if err != nil {
		t.Fatalf("cfg.SpecsDir returned error: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(customSpecsDir, "demo"), 0o755); err != nil {
		t.Fatalf("MkdirAll(spec dir) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(customSpecsDir, "demo", "spec.md"), []byte("# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 Demo\n- **Given** x\n- **When** y\n- **Then** z\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".speckeep", "constitution.summary.md"), []byte("## Purpose\n- x\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(summary) returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	for _, finding := range result.Findings {
		if finding.Level == "warning" && strings.Contains(finding.Message, "constitution.summary.md not found") {
			t.Fatalf("unexpected constitution summary warning when workspace summary exists: %+v", result.Findings)
		}
	}
}

func TestCheckWarnsAboutLegacyDefaultLayout(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
		SpecsDir:    "specs",
		ArchiveDir:  "archive",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	var found bool
	for _, finding := range result.Findings {
		if finding.Level == "warning" && strings.Contains(finding.Message, "legacy default layout") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected legacy layout warning, got %+v", result.Findings)
	}
}

func TestCheckWarnsAboutMixedLayout(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
		SpecsDir:    "specs",
		ArchiveDir:  "specs/archived",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	result, err := Check(root)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	var found bool
	for _, finding := range result.Findings {
		if finding.Level == "warning" && strings.Contains(finding.Message, "mixed old/new feature layout") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected mixed layout warning, got %+v", result.Findings)
	}
}
