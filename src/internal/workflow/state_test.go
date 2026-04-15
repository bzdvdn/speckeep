package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"speckeep/src/internal/project"
)

func TestStateInfersLifecycleAndReportStatuses(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	state, err := State(root, "demo")
	if err != nil {
		t.Fatalf("State returned error: %v", err)
	}
	if state.Phase != "constitution" || state.ReadyFor != "spec" || !state.Blocked {
		t.Fatalf("unexpected initial state: %+v", state)
	}

	specDir := filepath.Join(root, ".speckeep", "specs", "demo")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(specDir) returned error: %v", err)
	}
	specPath := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specPath, []byte("# Demo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	state, err = State(root, "demo")
	if err != nil {
		t.Fatalf("State returned error: %v", err)
	}
	if state.Phase != "spec" || state.ReadyFor != "inspect" {
		t.Fatalf("unexpected state after creating spec: %+v", state)
	}

	inspectPath := filepath.Join(specDir, "inspect.md")
	inspectInvalid := "# Inspect Report: demo\n\n## Verdict\n\n(no status)\n"
	if err := os.WriteFile(inspectPath, []byte(inspectInvalid), 0o644); err != nil {
		t.Fatalf("WriteFile(inspect invalid) returned error: %v", err)
	}

	state, err = State(root, "demo")
	if err != nil {
		t.Fatalf("State returned error: %v", err)
	}
	if state.Phase != "spec" || state.ReadyFor != "inspect" {
		t.Fatalf("expected invalid inspect report to require inspect, got %+v", state)
	}

	inspectContent := "---\nreport_type: inspect\nslug: demo\nstatus: concerns\ndocs_language: en\ngenerated_at: 2026-03-31\n---\n# Inspect Report: demo\n\n## Verdict\n\n- status: concerns\n"
	if err := os.WriteFile(inspectPath, []byte(inspectContent), 0o644); err != nil {
		t.Fatalf("WriteFile(inspect) returned error: %v", err)
	}

	state, err = State(root, "demo")
	if err != nil {
		t.Fatalf("State returned error: %v", err)
	}
	if state.Phase != "inspect" || state.ReadyFor != "plan" {
		t.Fatalf("unexpected inspect state: %+v", state)
	}
	if state.InspectStatus != StatusConcerns {
		t.Fatalf("InspectStatus = %q, want %q", state.InspectStatus, StatusConcerns)
	}
}

func TestInferLifecycleEmptyTasksFile(t *testing.T) {
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
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# Demo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}
	inspectContent := "---\nreport_type: inspect\nslug: demo\nstatus: pass\ndocs_language: en\ngenerated_at: 2026-03-31\n---\n## Verdict\n\n- status: pass\n"
	if err := os.WriteFile(filepath.Join(specDir, "inspect.md"), []byte(inspectContent), 0o644); err != nil {
		t.Fatalf("WriteFile(inspect) returned error: %v", err)
	}
	planDir := filepath.Join(specDir, "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(planDir) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "plan.md"), []byte("# Plan\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(plan) returned error: %v", err)
	}
	// tasks.md with no checkboxes (only headings)
	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte("# Tasks\n\n## Phase 1\n\nNo tasks defined yet.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}

	state, err := State(root, "demo")
	if err != nil {
		t.Fatalf("State returned error: %v", err)
	}
	if state.TasksTotal != 0 {
		t.Fatalf("expected TasksTotal=0, got %d", state.TasksTotal)
	}
	if state.Phase != "plan" || state.ReadyFor != "tasks" {
		t.Fatalf("expected phase=plan readyFor=tasks for empty tasks file, got phase=%s readyFor=%s", state.Phase, state.ReadyFor)
	}
}

func TestStateTreatsInvalidVerifyReportAsNotVerified(t *testing.T) {
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
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# Demo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}
	inspectContent := "---\nreport_type: inspect\nslug: demo\nstatus: pass\ndocs_language: en\ngenerated_at: 2026-03-31\n---\n## Verdict\n\n- status: pass\n"
	if err := os.WriteFile(filepath.Join(specDir, "inspect.md"), []byte(inspectContent), 0o644); err != nil {
		t.Fatalf("WriteFile(inspect) returned error: %v", err)
	}

	planDir := filepath.Join(specDir, "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(planDir) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "plan.md"), []byte("# Plan\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(plan) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte("- [x] T1.1 done\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}

	// Create an invalid verify report (missing status).
	if err := os.WriteFile(filepath.Join(planDir, "verify.md"), []byte("# Verify Report: demo\n\n## Verdict\n\n(no status)\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(verify invalid) returned error: %v", err)
	}

	state, err := State(root, "demo")
	if err != nil {
		t.Fatalf("State returned error: %v", err)
	}
	if state.Phase != "verify" || state.ReadyFor != "verify" {
		t.Fatalf("expected invalid verify report to require verify, got %+v", state)
	}
}

func TestInferLifecycleSubTaskCheckboxes(t *testing.T) {
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
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# Demo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}
	inspectContent := "---\nreport_type: inspect\nslug: demo\nstatus: pass\ndocs_language: en\ngenerated_at: 2026-03-31\n---\n## Verdict\n\n- status: pass\n"
	if err := os.WriteFile(filepath.Join(specDir, "inspect.md"), []byte(inspectContent), 0o644); err != nil {
		t.Fatalf("WriteFile(inspect) returned error: %v", err)
	}
	planDir := filepath.Join(specDir, "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(planDir) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "plan.md"), []byte("# Plan\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(plan) returned error: %v", err)
	}
	// T1.1 has two indented sub-tasks; parent is still open
	tasksContent := "# Tasks\n\n## Phase 1\n- [ ] T1.1 main task\n  - [x] T1.1.1 sub-task done\n  - [ ] T1.1.2 sub-task open\n"
	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte(tasksContent), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}

	state, err := State(root, "demo")
	if err != nil {
		t.Fatalf("State returned error: %v", err)
	}
	// total = 3 (parent + 2 sub-tasks), completed = 1, open = 2
	if state.TasksTotal != 3 {
		t.Fatalf("expected TasksTotal=3, got %d", state.TasksTotal)
	}
	if state.TasksCompleted != 1 {
		t.Fatalf("expected TasksCompleted=1, got %d", state.TasksCompleted)
	}
	if state.TasksOpen != 2 {
		t.Fatalf("expected TasksOpen=2, got %d", state.TasksOpen)
	}
	if state.Phase != "implement" {
		t.Fatalf("expected phase=implement, got %s", state.Phase)
	}
}

func TestInferLifecycleBranchMismatchBlocks(t *testing.T) {
	// Test inferLifecycle directly: BranchMismatch must set Blocked=true.
	state := FeatureState{
		Slug:           "demo",
		SpecExists:     true,
		InspectExists:  true,
		InspectStatus:  StatusPass,
		PlanExists:     true,
		TasksExists:    true,
		TasksTotal:     1,
		TasksCompleted: 1,
		TasksOpen:      0,
		VerifyExists:   true,
		VerifyStatus:   StatusPass,
		BranchMismatch: true,
		Archived:       false,
	}
	inferLifecycle(&state)
	if !state.Blocked {
		t.Fatalf("expected Blocked=true when BranchMismatch=true, got Blocked=%v", state)
	}
}

func TestInferLifecycleBranchMismatchDoesNotBlockArchived(t *testing.T) {
	// Archived features must not be blocked even if BranchMismatch is set.
	state := FeatureState{
		Slug:           "demo",
		Archived:       true,
		BranchMismatch: true,
	}
	inferLifecycle(&state)
	if state.Blocked {
		t.Fatalf("expected Blocked=false for archived feature, got Blocked=%v", state)
	}
}

func TestValidateProjectFindsSemanticWorkflowProblems(t *testing.T) {
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
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 Demo\n- **Given** x\n- **When** y\n- **Then** z\n"
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	inspectPath := filepath.Join(specDir, "inspect.md")
	inspectContent := "---\nreport_type: inspect\nslug: wrong-slug\nstatus: blocked\ndocs_language: en\ngenerated_at: 2026-03-31\n---\n# Inspect Report: demo\n\n## Verdict\n\n- status: blocked\n"
	if err := os.WriteFile(inspectPath, []byte(inspectContent), 0o644); err != nil {
		t.Fatalf("WriteFile(inspect) returned error: %v", err)
	}

	planDir := filepath.Join(root, ".speckeep", "specs", "demo", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(planDir) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "plan.md"), []byte("# Demo Plan\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(plan) returned error: %v", err)
	}

	findings, err := ValidateProject(root)
	if err != nil {
		t.Fatalf("ValidateProject returned error: %v", err)
	}

	var foundMismatch bool
	var foundBlocked bool
	for _, finding := range findings {
		if finding.Level == "error" && strings.Contains(finding.Message, "metadata slug mismatch") {
			foundMismatch = true
		}
		if finding.Level == "error" && strings.Contains(finding.Message, "downstream artifacts exist after blocked inspect") {
			foundBlocked = true
		}
	}
	if !foundMismatch || !foundBlocked {
		t.Fatalf("expected semantic findings, got %+v", findings)
	}
}

func TestValidateProjectFindsPlanTasksAndVerifyCoherenceProblems(t *testing.T) {
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
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 Demo\n- Given x\n- When y\n- Then z\n\n### AC-002 Demo\n- Given a\n- When b\n- Then c\n"
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	inspectPath := filepath.Join(specDir, "inspect.md")
	inspectContent := "---\nreport_type: inspect\nslug: demo\nstatus: pass\ndocs_language: en\ngenerated_at: 2026-03-31\n---\n# Inspect Report: demo\n\n## Verdict\n\n- status: pass\n\n## Errors\n\n- none\n\n## Warnings\n\n- none\n\n## Questions\n\n- none\n"
	if err := os.WriteFile(inspectPath, []byte(inspectContent), 0o644); err != nil {
		t.Fatalf("WriteFile(inspect) returned error: %v", err)
	}

	planDir := filepath.Join(root, ".speckeep", "specs", "demo", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(planDir) returned error: %v", err)
	}
	planContent := "# Demo Plan\n\n## Acceptance Approach\n- AC-001 -> implementation path\n\n## Implementation Strategy\n- DEC-001 x\n"
	if err := os.WriteFile(filepath.Join(planDir, "plan.md"), []byte(planContent), 0o644); err != nil {
		t.Fatalf("WriteFile(plan) returned error: %v", err)
	}

	tasksContent := "# Demo Tasks\n\n## Phase 1: Foundation\n- [x] T1.1 done\n- [ ] T1.2 open\n\n## Acceptance Coverage\n- AC-001 -> T1.1\n- AC-003 -> T9.9\n"
	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte(tasksContent), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}

	verifyContent := "---\nreport_type: verify\nslug: demo\nstatus: pass\ndocs_language: en\ngenerated_at: 2026-03-31\n---\n# Verify Report: demo\n\n## Verdict\n\n- status: pass\n\n## Errors\n\n- unresolved bug still present\n"
	if err := os.WriteFile(filepath.Join(planDir, "verify.md"), []byte(verifyContent), 0o644); err != nil {
		t.Fatalf("WriteFile(verify) returned error: %v", err)
	}

	findings, err := ValidateProject(root)
	if err != nil {
		t.Fatalf("ValidateProject returned error: %v", err)
	}

	wantSubstrings := []string{
		"plan does not reference acceptance criterion AC-002",
		"acceptance coverage references unknown acceptance criterion AC-003",
		"acceptance coverage references unknown task ID T9.9",
		"acceptance criterion AC-002 is not covered by tasks",
		"verify report says pass but still lists errors",
		"verify report says pass while open tasks remain",
	}
	for _, want := range wantSubstrings {
		var found bool
		for _, finding := range findings {
			if strings.Contains(finding.Message, want) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected finding containing %q, got %+v", want, findings)
		}
	}
}

func TestValidateProjectFindsVerifyTraceabilityAndArchiveReadinessProblems(t *testing.T) {
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
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 Demo\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	inspectPath := filepath.Join(specDir, "inspect.md")
	inspectContent := "---\nreport_type: inspect\nslug: demo\nstatus: pass\ndocs_language: en\ngenerated_at: 2026-03-31\n---\n# Inspect Report: demo\n\n## Verdict\n\n- status: pass\n\n## Errors\n\n- none\n\n## Warnings\n\n- none\n\n## Questions\n\n- none\n"
	if err := os.WriteFile(inspectPath, []byte(inspectContent), 0o644); err != nil {
		t.Fatalf("WriteFile(inspect) returned error: %v", err)
	}

	planDir := filepath.Join(root, ".speckeep", "specs", "demo", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(planDir) returned error: %v", err)
	}

	tasksContent := "# Demo Tasks\n\n## Phase 1: Foundation\n- [x] T1.1 done\n\n## Acceptance Coverage\n- AC-001 -> T1.1\n"
	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte(tasksContent), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}

	verifyContent := "---\nreport_type: verify\nslug: demo\nstatus: concerns\ndocs_language: en\ngenerated_at: 2026-03-31\n---\n# Verify Report: demo\n\n## Verdict\n\n- status: concerns\n- archive_readiness: safe\n\n## Checks\n\n- task_state: completed=1, open=0\n\n## Errors\n\n- none\n\n## Warnings\n\n- none\n\n## Questions\n\n- none\n\n## Next Step\n\n- safe to archive\n"
	if err := os.WriteFile(filepath.Join(planDir, "verify.md"), []byte(verifyContent), 0o644); err != nil {
		t.Fatalf("WriteFile(verify) returned error: %v", err)
	}

	findings, err := ValidateProject(root)
	if err != nil {
		t.Fatalf("ValidateProject returned error: %v", err)
	}

	wantSubstrings := []string{
		"verify report says concerns but still marks archive_readiness safe",
		"verify report is missing Not Verified section",
		"verify report does not reference any task IDs",
		"verify report does not reference any acceptance criteria",
	}
	for _, want := range wantSubstrings {
		var found bool
		for _, finding := range findings {
			if strings.Contains(finding.Message, want) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected finding containing %q, got %+v", want, findings)
		}
	}
}

func TestValidateFeatureReturnsOnlyFindingsForRequestedSlug(t *testing.T) {
	root := t.TempDir()

	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	alphaDir := filepath.Join(root, ".speckeep", "specs", "alpha")
	if err := os.MkdirAll(alphaDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(alphaDir) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(alphaDir, "spec.md"), []byte("# Alpha\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(alpha) returned error: %v", err)
	}
	betaDir := filepath.Join(root, ".speckeep", "specs", "beta")
	if err := os.MkdirAll(betaDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(betaDir) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(betaDir, "spec.md"), []byte("# Beta\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(beta) returned error: %v", err)
	}

	findings, err := ValidateFeature(root, "alpha")
	if err != nil {
		t.Fatalf("ValidateFeature returned error: %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("expected findings for alpha")
	}
	for _, finding := range findings {
		if slug := FindingSlug(finding); slug != "" && slug != "alpha" {
			t.Fatalf("expected only alpha findings, got %+v", findings)
		}
	}
}
