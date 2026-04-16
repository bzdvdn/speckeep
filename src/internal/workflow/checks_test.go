package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"speckeep/src/internal/project"
)

func TestCheckResultAddFindingTracksStructuredState(t *testing.T) {
	var result CheckResult

	result.AddFinding(CheckFinding{
		Code:     "ok_code",
		Severity: SeverityOK,
		Category: CategoryReadiness,
		Message:  "all good",
	})
	result.AddFinding(CheckFinding{
		Code:     "warn_code",
		Severity: SeverityWarning,
		Category: CategoryAmbiguity,
		Message:  "something unclear",
	})
	result.AddFinding(CheckFinding{
		Code:     "error_code",
		Severity: SeverityError,
		Category: CategoryStructure,
		Message:  "something missing",
	})

	if len(result.Findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(result.Findings))
	}
	if result.Warnings != 1 {
		t.Fatalf("expected 1 warning, got %d", result.Warnings)
	}
	if result.Errors != 1 {
		t.Fatalf("expected 1 error, got %d", result.Errors)
	}
	if !result.Failed {
		t.Fatalf("expected result to be marked failed")
	}

	joined := strings.Join(result.Lines, "\n")
	if !strings.Contains(joined, "OK: all good") {
		t.Fatalf("expected OK line, got %s", joined)
	}
	if !strings.Contains(joined, "WARN: something unclear") {
		t.Fatalf("expected WARN line, got %s", joined)
	}
	if !strings.Contains(joined, "ERROR: something missing") {
		t.Fatalf("expected ERROR line, got %s", joined)
	}
}

func TestCheckResultMergeIncludesFindings(t *testing.T) {
	left := CheckResult{}
	left.AddWarn("left warning")

	right := CheckResult{}
	right.AddError("right error")

	left.Merge(right)

	if len(left.Findings) != 2 {
		t.Fatalf("expected merged findings, got %d", len(left.Findings))
	}
	if left.Warnings != 1 || left.Errors != 1 {
		t.Fatalf("unexpected counters after merge: warnings=%d errors=%d", left.Warnings, left.Errors)
	}
	if !left.Failed {
		t.Fatalf("expected merged result to be failed")
	}
}

func TestInspectSpecValidatesAcceptanceCoverageInGo(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specDir := filepath.Join(root, ".speckeep", "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(specDir) returned error: %v", err)
	}
	specPath := filepath.Join(specDir, "spec.md")
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 First\n- Given x\n- When y\n- Then z\n\n### AC-002 Second\n- Given a\n- When b\n- Then c\n"
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	tasksPath := filepath.Join(root, ".speckeep", "specs", "demo", "plan", "tasks.md")
	if err := os.MkdirAll(filepath.Dir(tasksPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	tasksContent := "# Tasks\n\n## Acceptance Coverage\n- AC-001 -> T1.1\n"
	if err := os.WriteFile(tasksPath, []byte(tasksContent), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}

	result, err := InspectSpec(root, ".speckeep/specs/demo/spec.md", ".speckeep/specs/demo/plan/tasks.md")
	if err != nil {
		t.Fatalf("InspectSpec returned error: %v", err)
	}
	if !result.Failed {
		t.Fatalf("expected InspectSpec to fail, got %+v", result)
	}
	joined := strings.Join(result.Lines, "\n")
	if !strings.Contains(joined, "acceptance coverage entries (1) are fewer than acceptance criteria (2)") {
		t.Fatalf("expected coverage mismatch in output, got %s", joined)
	}
	if !strings.Contains(joined, "SUMMARY: errors=") {
		t.Fatalf("expected summary line in output, got %s", joined)
	}
	if !containsFinding(result.Findings, "acceptance_not_covered") {
		t.Fatalf("expected structured acceptance_not_covered finding, got %+v", result.Findings)
	}
}

func TestInspectSpecDetectsAmbiguousLanguage(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specDir := filepath.Join(root, ".speckeep", "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(specDir) returned error: %v", err)
	}
	specPath := filepath.Join(specDir, "spec.md")
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 The flow should feel fast.\n\n## Acceptance Criteria\n### AC-001 First\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	result, err := InspectSpec(root, ".speckeep/specs/demo/spec.md", "")
	if err != nil {
		t.Fatalf("InspectSpec returned error: %v", err)
	}

	if !containsFinding(result.Findings, "ambiguous_wording") {
		t.Fatalf("expected ambiguous_wording finding, got %+v", result.Findings)
	}
	if !strings.Contains(strings.Join(result.Lines, "\n"), `ambiguous wording detected in Requirements: "should"`) {
		t.Fatalf("expected ambiguity warning in lines, got %+v", result.Lines)
	}
}

func TestInspectSpecDetectsUnknownCoverageReferences(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specDir := filepath.Join(root, ".speckeep", "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(specDir) returned error: %v", err)
	}
	specPath := filepath.Join(specDir, "spec.md")
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 First\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	tasksPath := filepath.Join(root, ".speckeep", "specs", "demo", "plan", "tasks.md")
	if err := os.MkdirAll(filepath.Dir(tasksPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	tasksContent := "# Tasks\n\n## Acceptance Coverage\n- AC-999 -> T1.9\n\n- [ ] T1.1 real task `Touches: foo.go`\n"
	if err := os.WriteFile(tasksPath, []byte(tasksContent), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}

	result, err := InspectSpec(root, ".speckeep/specs/demo/spec.md", ".speckeep/specs/demo/plan/tasks.md")
	if err != nil {
		t.Fatalf("InspectSpec returned error: %v", err)
	}

	if !containsFinding(result.Findings, "unknown_acceptance_reference") {
		t.Fatalf("expected unknown_acceptance_reference finding, got %+v", result.Findings)
	}
	if !containsFinding(result.Findings, "unknown_task_reference") {
		t.Fatalf("expected unknown_task_reference finding, got %+v", result.Findings)
	}
}

func TestCheckImplementReadyDetectsPlanTaskSurfaceMismatch(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specDir := filepath.Join(root, ".speckeep", "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(specDir) returned error: %v", err)
	}
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 First\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	planDir := filepath.Join(root, ".speckeep", "specs", "demo", "plan")
	if err := os.MkdirAll(filepath.Join(planDir, "contracts"), 0o755); err != nil {
		t.Fatalf("MkdirAll(contracts) returned error: %v", err)
	}
	planContent := "# Demo Plan\n\n## Implementation Surfaces\n- src/handlers/demo.ts\n- src/models/demo.ts\n\n## Acceptance Approach\n- AC-001 -> handler path\n"
	if err := os.WriteFile(filepath.Join(planDir, "plan.md"), []byte(planContent), 0o644); err != nil {
		t.Fatalf("WriteFile(plan) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "data-model.md"), []byte("# Data Model\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(data-model) returned error: %v", err)
	}
	tasksContent := "# Tasks\n\n## Surface Map\n| Surface | Tasks |\n|---------|-------|\n| src/services/demo.ts | T1.1 |\n\n## Phase 1: Foundation\n- [ ] T1.1 Build demo path. Touches: src/services/demo.ts\n\n## Acceptance Coverage\n- AC-001 -> T1.1\n"
	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte(tasksContent), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}

	result, err := CheckImplementReady(root, "demo")
	if err != nil {
		t.Fatalf("CheckImplementReady returned error: %v", err)
	}
	if !containsFinding(result.Findings, "plan_surface_missing_from_tasks") {
		t.Fatalf("expected plan_surface_missing_from_tasks finding, got %+v", result.Findings)
	}
	if !containsFinding(result.Findings, "task_surface_missing_from_plan") {
		t.Fatalf("expected task_surface_missing_from_plan finding, got %+v", result.Findings)
	}
}

func TestVerifyTaskStateReportsOpenTasksWithoutFailing(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	tasksPath := filepath.Join(root, ".speckeep", "specs", "demo", "plan", "tasks.md")
	if err := os.MkdirAll(filepath.Dir(tasksPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	tasksContent := "- [x] T1.1 done\n- [ ] T1.2 open\n\n## Acceptance Coverage\n- AC-001 -> T1.1\n"
	if err := os.WriteFile(tasksPath, []byte(tasksContent), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}

	result, summary, err := VerifyTaskState(root, "demo")
	if err != nil {
		t.Fatalf("VerifyTaskState returned error: %v", err)
	}
	if result.Failed {
		t.Fatalf("expected open tasks to warn but not fail, got %+v", result)
	}
	if summary.Open != 1 || summary.Total != 2 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	joined := strings.Join(result.Lines, "\n")
	if !strings.Contains(joined, "TASKS_OPEN=1") || !strings.Contains(joined, "WARN: open tasks remain") {
		t.Fatalf("unexpected verify-task-state output: %s", joined)
	}
}

func TestCheckArchiveReadyBlocksCompletedArchiveWhenTasksRemainOpen(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specDir := filepath.Join(root, ".speckeep", "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(specDir) returned error: %v", err)
	}
	specPath := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specPath, []byte("# Demo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}
	tasksPath := filepath.Join(root, ".speckeep", "specs", "demo", "plan", "tasks.md")
	if err := os.MkdirAll(filepath.Dir(tasksPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(tasksPath, []byte("- [ ] T1.1 open\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}

	result, err := CheckArchiveReady(root, "demo", "completed", "done")
	if err != nil {
		t.Fatalf("CheckArchiveReady returned error: %v", err)
	}
	if !result.Failed {
		t.Fatalf("expected archive readiness to fail, got %+v", result)
	}
	if !strings.Contains(strings.Join(result.Lines, "\n"), "completed archive requested while open tasks remain") {
		t.Fatalf("unexpected archive output: %+v", result.Lines)
	}
}

func TestCheckArchiveReadyAllowsCompletedWithoutReason(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specDir := filepath.Join(root, ".speckeep", "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(specDir) returned error: %v", err)
	}
	specPath := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specPath, []byte("# Demo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	result, err := CheckArchiveReady(root, "demo", "completed", "")
	if err != nil {
		t.Fatalf("CheckArchiveReady returned error: %v", err)
	}
	if result.Failed {
		t.Fatalf("expected archive readiness to pass, got %+v", result)
	}
}

func TestCheckArchiveReadyAllowsHotfixWithoutSpec(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specDir := filepath.Join(root, ".speckeep", "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(specDir) returned error: %v", err)
	}
	hotfixPath := filepath.Join(specDir, "hotfix.md")
	if err := os.WriteFile(hotfixPath, []byte("# Hotfix\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(hotfix) returned error: %v", err)
	}

	result, err := CheckArchiveReady(root, "demo", "completed", "")
	if err != nil {
		t.Fatalf("CheckArchiveReady returned error: %v", err)
	}
	if result.Failed {
		t.Fatalf("expected archive readiness to pass for hotfix, got %+v", result)
	}
}

func TestCheckArchiveReadyRequiresReasonForDeferred(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	specDir := filepath.Join(root, ".speckeep", "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(specDir) returned error: %v", err)
	}
	specPath := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specPath, []byte("# Demo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	result, err := CheckArchiveReady(root, "demo", "deferred", "")
	if err != nil {
		t.Fatalf("CheckArchiveReady returned error: %v", err)
	}
	if !result.Failed {
		t.Fatalf("expected archive readiness to fail, got %+v", result)
	}
	if !strings.Contains(strings.Join(result.Lines, "\n"), "archive reason is required for non-completed statuses") {
		t.Fatalf("unexpected archive output: %+v", result.Lines)
	}
}

func containsFinding(findings []CheckFinding, code string) bool {
	for _, finding := range findings {
		if finding.Code == code {
			return true
		}
	}
	return false
}

func TestInspectSpecDetectsNeedsClarificationMarker(t *testing.T) {
	root := t.TempDir()
	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	specDir := filepath.Join(root, ".speckeep", "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 [NEEDS CLARIFICATION: define scope]\n\n## Acceptance Criteria\n### AC-001 First\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec): %v", err)
	}

	result, err := InspectSpec(root, ".speckeep/specs/demo/spec.md", "")
	if err != nil {
		t.Fatalf("InspectSpec: %v", err)
	}
	if !containsFinding(result.Findings, "needs_clarification_marker") {
		t.Fatalf("expected needs_clarification_marker finding, got %+v", result.Findings)
	}
	if !result.Failed {
		t.Fatalf("expected result to fail due to needs_clarification_marker")
	}
}

func TestInspectSpecWarnsOnMissingAssumptions(t *testing.T) {
	root := t.TempDir()
	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	specDir := filepath.Join(root, ".speckeep", "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Spec with no Assumptions section
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 First\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec): %v", err)
	}

	result, err := InspectSpec(root, ".speckeep/specs/demo/spec.md", "")
	if err != nil {
		t.Fatalf("InspectSpec: %v", err)
	}
	if !containsFinding(result.Findings, "optional_section_missing") {
		t.Fatalf("expected optional_section_missing warning for Assumptions, got %+v", result.Findings)
	}
	found := false
	for _, f := range result.Findings {
		if f.Code == "optional_section_missing" && strings.Contains(strings.Join(f.Refs, ","), "Assumptions") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected optional_section_missing finding referencing Assumptions, got %+v", result.Findings)
	}
}

func TestInspectSpecWarnsOnMissingRQIDs(t *testing.T) {
	root := t.TempDir()
	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	specDir := filepath.Join(root, ".speckeep", "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Requirements section present but no RQ-* IDs
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- The system must do X.\n\n## Acceptance Criteria\n### AC-001 First\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec): %v", err)
	}

	result, err := InspectSpec(root, ".speckeep/specs/demo/spec.md", "")
	if err != nil {
		t.Fatalf("InspectSpec: %v", err)
	}
	if !containsFinding(result.Findings, "requirement_ids_missing") {
		t.Fatalf("expected requirement_ids_missing finding, got %+v", result.Findings)
	}
}

func TestInspectSpecDetectsRQIDs(t *testing.T) {
	root := t.TempDir()
	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	specDir := filepath.Join(root, ".speckeep", "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 The system must do X.\n- RQ-002 The system must do Y.\n\n## Acceptance Criteria\n### AC-001 First\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec): %v", err)
	}

	result, err := InspectSpec(root, ".speckeep/specs/demo/spec.md", "")
	if err != nil {
		t.Fatalf("InspectSpec: %v", err)
	}
	if !containsFinding(result.Findings, "requirement_ids_present") {
		t.Fatalf("expected requirement_ids_present finding, got %+v", result.Findings)
	}
}

func TestCheckConstitutionLanguagePolicyMismatch(t *testing.T) {
	var result CheckResult
	// Constitution says docs: en, but project configured for ru
	constitutionContent := "# Constitution\n\n## Language Policy\n\n- docs: en\n- code: en\n"
	checkConstitutionLanguagePolicy(&result, constitutionContent, "ru")
	if !containsFinding(result.Findings, "constitution_language_mismatch") {
		t.Fatalf("expected constitution_language_mismatch finding, got %+v", result.Findings)
	}
	if result.Warnings != 1 {
		t.Fatalf("expected 1 warning, got %d", result.Warnings)
	}
}

func TestCheckConstitutionLanguagePolicyMatch(t *testing.T) {
	var result CheckResult
	constitutionContent := "# Constitution\n\n## Language Policy\n\n- docs: ru\n- code: en\n"
	checkConstitutionLanguagePolicy(&result, constitutionContent, "ru")
	if !containsFinding(result.Findings, "constitution_language_consistent") {
		t.Fatalf("expected constitution_language_consistent finding, got %+v", result.Findings)
	}
	if result.Warnings != 0 {
		t.Fatalf("expected 0 warnings, got %d", result.Warnings)
	}
}

func TestCheckTouchesFilesExistWarnsOnMissingFile(t *testing.T) {
	root := t.TempDir()
	var result CheckResult
	tasksContent := "- [x] T1.1 done. Touches: src/handler.go\n- [x] T1.2 done. Touches: src/missing.go\n"
	// Create only src/handler.go
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "handler.go"), []byte("package main"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	checkTouchesFilesExist(&result, root, "tasks.md", tasksContent)

	if !containsFinding(result.Findings, "touches_file_missing") {
		t.Fatalf("expected touches_file_missing finding, got %+v", result.Findings)
	}
	if result.Warnings != 1 {
		t.Fatalf("expected exactly 1 warning (only missing.go), got %d", result.Warnings)
	}
}

func TestCheckTouchesFilesExistPassesWhenAllPresent(t *testing.T) {
	root := t.TempDir()
	var result CheckResult
	tasksContent := "- [x] T1.1 done. Touches: src/handler.go\n"
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "handler.go"), []byte("package main"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	checkTouchesFilesExist(&result, root, "tasks.md", tasksContent)

	if containsFinding(result.Findings, "touches_file_missing") {
		t.Fatalf("expected no touches_file_missing finding, got %+v", result.Findings)
	}
}

func TestCheckPlanContentDetectsMissingDecisionIDs(t *testing.T) {
	root := t.TempDir()
	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	specDir := filepath.Join(root, ".speckeep", "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	specContent := "# Demo\n\n## Goal\nx\n\n## Acceptance Criteria\n### AC-001 First\n- Given x\n- When y\n- Then z\n"
	specPath := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec): %v", err)
	}

	var result CheckResult
	// Plan with no DEC-* IDs and no Acceptance Approach or Constitution Compliance
	planContent := "# Plan\n\n## Overview\nSome plan without required sections.\n"
	checkPlanContent(&result, "demo", specPath, "plan.md", planContent)

	if !containsFinding(result.Findings, "plan_no_decision_ids") {
		t.Fatalf("expected plan_no_decision_ids finding, got %+v", result.Findings)
	}
	if !containsFinding(result.Findings, "plan_missing_acceptance_approach") {
		t.Fatalf("expected plan_missing_acceptance_approach finding, got %+v", result.Findings)
	}
	if !containsFinding(result.Findings, "plan_missing_constitution_compliance") {
		t.Fatalf("expected plan_missing_constitution_compliance finding, got %+v", result.Findings)
	}
}

func TestCheckPlanContentDetectsACMismatch(t *testing.T) {
	root := t.TempDir()
	_, err := project.Initialize(root, project.InitOptions{InitGit: false, DefaultLang: "en", Shell: "sh"})
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	specDir := filepath.Join(root, ".speckeep", "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	specContent := "# Demo\n\n## Goal\nx\n\n## Acceptance Criteria\n### AC-001 First\n- Given x\n- When y\n- Then z\n\n### AC-002 Second\n- Given a\n- When b\n- Then c\n"
	specPath := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec): %v", err)
	}

	var result CheckResult
	// Plan references AC-001 in Acceptance Approach but misses AC-002; also references unknown AC-999
	planContent := "# Plan\n\n## DEC-001 Decision\nSome decision.\n\n## Acceptance Approach\n- AC-001 covered by handler layer\n- AC-999 unknown criterion\n\n## Constitution Compliance\n- compliant\n"
	checkPlanContent(&result, "demo", specPath, "plan.md", planContent)

	if !containsFinding(result.Findings, "plan_missing_ac_reference") {
		t.Fatalf("expected plan_missing_ac_reference for AC-002, got %+v", result.Findings)
	}
	if !containsFinding(result.Findings, "plan_unknown_ac_reference") {
		t.Fatalf("expected plan_unknown_ac_reference for AC-999, got %+v", result.Findings)
	}
}
