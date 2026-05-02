package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"speckeep/src/internal/config"
)

func executeRoot(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	cmd := NewRootCmd()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func ensureSpecDir(t *testing.T, root, slug string) string {
	t.Helper()

	dir := filepath.Join(root, "specs", slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) returned error: %v", dir, err)
	}
	return dir
}

func createGitSkillRepo(t *testing.T, root string) string {
	t.Helper()

	repoDir := filepath.Join(root, "git-skill-repo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "SKILL.md"), []byte("# skill\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(SKILL.md) returned error: %v", err)
	}
	runGitTest(t, repoDir, "init")
	runGitTest(t, repoDir, "config", "user.email", "test@example.com")
	runGitTest(t, repoDir, "config", "user.name", "Test User")
	runGitTest(t, repoDir, "add", "SKILL.md")
	runGitTest(t, repoDir, "commit", "-m", "init skill")
	runGitTest(t, repoDir, "tag", "v1.2.3")
	return repoDir
}

func runGitTest(t *testing.T, dir string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v, output=%s", args, err, string(output))
	}
}

func TestInitCommandCreatesWorkspace(t *testing.T) {
	root := t.TempDir()

	stdout, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh", "--agents", "claude")
	if err != nil {
		t.Fatalf("init command returned error: %v", err)
	}
	if !strings.Contains(stdout, "enabled agent targets: claude") {
		t.Fatalf("unexpected init output: %s", stdout)
	}

	required := []string{
		filepath.Join(root, ".speckeep", "speckeep.yaml"),
		filepath.Join(root, "CONSTITUTION.md"),
		filepath.Join(root, ".claude", "commands", "speckeep.inspect.md"),
	}
	for _, path := range required {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}
}

func TestInitCommandSupportsCustomSpecsAndArchiveDirs(t *testing.T) {
	root := t.TempDir()

	specsDir := ".speckeep/specifications"
	archiveDir := ".speckeep/artifacts/archive"
	constitutionFile := "docs/constitution.md"

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh", "--specs-dir", specsDir, "--archive-dir", archiveDir, "--constitution-file", constitutionFile); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load config returned error: %v", err)
	}

	if cfg.Paths.SpecsDir != specsDir {
		t.Fatalf("expected specs_dir=%q, got %q", specsDir, cfg.Paths.SpecsDir)
	}
	if cfg.Paths.ArchiveDir != archiveDir {
		t.Fatalf("expected archive_dir=%q, got %q", archiveDir, cfg.Paths.ArchiveDir)
	}
	if cfg.Project.ConstitutionFile != constitutionFile {
		t.Fatalf("expected constitution_file=%q, got %q", constitutionFile, cfg.Project.ConstitutionFile)
	}

	absSpecsDir, err := cfg.SpecsDir(root)
	if err != nil {
		t.Fatalf("SpecsDir returned error: %v", err)
	}
	if _, err := os.Stat(absSpecsDir); err != nil {
		t.Fatalf("expected specs directory to exist: %v", err)
	}

	absArchiveDir, err := cfg.ArchiveDir(root)
	if err != nil {
		t.Fatalf("ArchiveDir returned error: %v", err)
	}
	if _, err := os.Stat(absArchiveDir); err != nil {
		t.Fatalf("expected archive directory to exist: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, constitutionFile)); err != nil {
		t.Fatalf("expected constitution file to exist: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "CONSTITUTION.md")); !os.IsNotExist(err) {
		t.Fatalf("expected default CONSTITUTION.md to be absent when constitution-file is overridden, got err=%v", err)
	}
}

func TestInternalListSpecsRespectsConfiguredSpecsDir(t *testing.T) {
	root := t.TempDir()

	specsDir := ".speckeep/specifications"
	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh", "--specs-dir", specsDir); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	slug := "alpha"
	specPath := filepath.Join(root, specsDir, slug, "spec.md")
	if err := os.MkdirAll(filepath.Dir(specPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(specPath, []byte("# Alpha\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "__internal", "list-specs", "--root", root)
	if err != nil {
		t.Fatalf("internal list-specs returned error: %v", err)
	}
	if strings.TrimSpace(stdout) != "alpha" {
		t.Fatalf("unexpected list-specs output: %q", stdout)
	}

	stdout, _, err = executeRoot(t, "__internal", "show-spec", "--root", root, slug)
	if err != nil {
		t.Fatalf("internal show-spec returned error: %v", err)
	}
	if strings.TrimSpace(stdout) != "# Alpha" {
		t.Fatalf("unexpected show-spec output: %q", stdout)
	}
}

func TestListSpecsAndShowSpecCommands(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	specsDir := filepath.Join(root, "specs")
	if err := os.MkdirAll(filepath.Join(specsDir, "alpha"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "alpha", "spec.md"), []byte("# Alpha\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(specsDir, "beta"), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "beta", "spec.md"), []byte("# Beta\nBody\n"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "list-specs", root)
	if err != nil {
		t.Fatalf("list-specs command returned error: %v", err)
	}
	if strings.TrimSpace(stdout) != "alpha\nbeta" {
		t.Fatalf("unexpected list-specs output: %q", stdout)
	}

	stdout, _, err = executeRoot(t, "show-spec", "beta", root)
	if err != nil {
		t.Fatalf("show-spec command returned error: %v", err)
	}
	if stdout != "# Beta\nBody\n" {
		t.Fatalf("unexpected show-spec output: %q", stdout)
	}
}

func TestAddAgentAndDoctorCommands(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "add-agent", root, "--agents", "cursor")
	if err != nil {
		t.Fatalf("add-agent command returned error: %v", err)
	}
	if !strings.Contains(stdout, "enabled agent targets: cursor") {
		t.Fatalf("unexpected add-agent output: %s", stdout)
	}

	stdout, _, err = executeRoot(t, "doctor", root)
	if err != nil {
		t.Fatalf("doctor command returned error: %v", err)
	}
	if !strings.Contains(stdout, "summary:") || !strings.Contains(stdout, "oks:") || !strings.Contains(stdout, "speckeep workspace looks healthy") {
		t.Fatalf("unexpected doctor output: %s", stdout)
	}
}

func TestDoctorCommandJSONOutput(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "doctor", root, "--json")
	if err != nil {
		t.Fatalf("doctor --json returned error: %v", err)
	}

	var payload struct {
		Findings []struct {
			Level   string `json:"Level"`
			Message string `json:"Message"`
		} `json:"Findings"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse doctor json output %q: %v", stdout, err)
	}
	if len(payload.Findings) == 0 {
		t.Fatalf("expected findings in doctor json output, got %q", stdout)
	}
	if payload.Findings[len(payload.Findings)-1].Level != "ok" {
		t.Fatalf("expected trailing ok finding in json output, got %+v", payload.Findings)
	}
}

func TestStatusCommandJSONOutput(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	specPath := filepath.Join(ensureSpecDir(t, root, "demo"), "spec.md")
	if err := os.WriteFile(specPath, []byte("# Demo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	inspectPath := filepath.Join(ensureSpecDir(t, root, "demo"), "inspect.md")
	if err := os.WriteFile(inspectPath, []byte("---\nreport_type: inspect\nslug: demo\nstatus: pass\ndocs_language: en\ngenerated_at: 2026-03-30\n---\n# Inspect Report: demo\n\n## Verdict\n\n- status: pass\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(inspect) returned error: %v", err)
	}

	planDir := filepath.Join(root, "specs", "demo", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(planDir) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "plan.md"), []byte("# Demo Plan\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(plan) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte("- [x] T1.1 Done\n- [ ] T1.2 Open\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "status", "demo", root, "--json")
	if err != nil {
		t.Fatalf("status --json returned error: %v", err)
	}

	var payload struct {
		Slug           string `json:"slug"`
		Phase          string `json:"phase"`
		SpecExists     bool   `json:"spec_exists"`
		InspectExists  bool   `json:"inspect_exists"`
		PlanExists     bool   `json:"plan_exists"`
		TasksExists    bool   `json:"tasks_exists"`
		VerifyExists   bool   `json:"verify_exists"`
		TasksTotal     int    `json:"tasks_total"`
		TasksCompleted int    `json:"tasks_completed"`
		TasksOpen      int    `json:"tasks_open"`
		ReadyFor       string `json:"ready_for"`
		Blocked        bool   `json:"blocked"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse status json output %q: %v", stdout, err)
	}
	if payload.Slug != "demo" || payload.Phase != "implement" {
		t.Fatalf("unexpected status payload: %+v", payload)
	}
	if !payload.SpecExists || !payload.InspectExists || !payload.PlanExists || !payload.TasksExists {
		t.Fatalf("expected spec/plan/tasks to exist, got %+v", payload)
	}
	if payload.TasksTotal != 2 || payload.TasksCompleted != 1 || payload.TasksOpen != 1 {
		t.Fatalf("unexpected task counts: %+v", payload)
	}
	if payload.ReadyFor != "implement" || payload.Blocked {
		t.Fatalf("unexpected ready/block state: %+v", payload)
	}
}

func TestInitAndStatusCommandsFollowFeatureLifecycle(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	type statusPayload struct {
		Slug           string `json:"slug"`
		Phase          string `json:"phase"`
		SpecExists     bool   `json:"spec_exists"`
		InspectExists  bool   `json:"inspect_exists"`
		PlanExists     bool   `json:"plan_exists"`
		TasksExists    bool   `json:"tasks_exists"`
		VerifyExists   bool   `json:"verify_exists"`
		TasksTotal     int    `json:"tasks_total"`
		TasksCompleted int    `json:"tasks_completed"`
		TasksOpen      int    `json:"tasks_open"`
		ReadyFor       string `json:"ready_for"`
		Blocked        bool   `json:"blocked"`
	}

	checkStatus := func(want statusPayload) {
		t.Helper()

		stdout, _, err := executeRoot(t, "status", "demo", root, "--json")
		if err != nil {
			t.Fatalf("status --json returned error: %v", err)
		}

		var payload statusPayload
		if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
			t.Fatalf("failed to parse status json output %q: %v", stdout, err)
		}

		if payload.Slug != "demo" {
			t.Fatalf("slug = %q, want demo", payload.Slug)
		}
		if payload.Phase != want.Phase || payload.ReadyFor != want.ReadyFor || payload.Blocked != want.Blocked {
			t.Fatalf("unexpected phase payload: %+v, want %+v", payload, want)
		}
		if payload.SpecExists != want.SpecExists || payload.InspectExists != want.InspectExists || payload.PlanExists != want.PlanExists || payload.TasksExists != want.TasksExists || payload.VerifyExists != want.VerifyExists {
			t.Fatalf("unexpected artifact flags: %+v, want %+v", payload, want)
		}
		if payload.TasksTotal != want.TasksTotal || payload.TasksCompleted != want.TasksCompleted || payload.TasksOpen != want.TasksOpen {
			t.Fatalf("unexpected task counts: %+v, want %+v", payload, want)
		}
	}

	checkStatus(statusPayload{
		Phase:         "constitution",
		ReadyFor:      "spec",
		Blocked:       true,
		SpecExists:    false,
		InspectExists: false,
		PlanExists:    false,
		TasksExists:   false,
		VerifyExists:  false,
	})

	specPath := filepath.Join(ensureSpecDir(t, root, "demo"), "spec.md")
	specContent := "# Feature Specification: Demo\n\n## Requirements\n- RQ-001 Support a minimal demo flow.\n\n## Acceptance Criteria\n- AC-001\n  - Given a prepared workspace\n  - When the feature lifecycle is checked\n  - Then the status should advance predictably.\n"
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	checkStatus(statusPayload{
		Phase:         "spec",
		ReadyFor:      "plan",
		Blocked:       false,
		SpecExists:    true,
		InspectExists: false,
		PlanExists:    false,
		TasksExists:   false,
		VerifyExists:  false,
	})

	inspectPath := filepath.Join(ensureSpecDir(t, root, "demo"), "inspect.md")
	inspectContent := "---\nreport_type: inspect\nslug: demo\nstatus: pass\ndocs_language: en\ngenerated_at: 2026-03-30\n---\n# Inspect Report: demo\n\n## Verdict\n\n- status: pass\n"
	if err := os.WriteFile(inspectPath, []byte(inspectContent), 0o644); err != nil {
		t.Fatalf("WriteFile(inspect) returned error: %v", err)
	}

	checkStatus(statusPayload{
		Phase:         "inspect",
		ReadyFor:      "plan",
		Blocked:       false,
		SpecExists:    true,
		InspectExists: true,
		PlanExists:    false,
		TasksExists:   false,
		VerifyExists:  false,
	})

	planDir := filepath.Join(root, "specs", "demo", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(planDir) returned error: %v", err)
	}
	planContent := "# Implementation Plan: Demo\n\n## Decisions\n- DEC-001 Keep the integration test minimal and deterministic.\n"
	if err := os.WriteFile(filepath.Join(planDir, "plan.md"), []byte(planContent), 0o644); err != nil {
		t.Fatalf("WriteFile(plan) returned error: %v", err)
	}

	checkStatus(statusPayload{
		Phase:         "plan",
		ReadyFor:      "tasks",
		Blocked:       false,
		SpecExists:    true,
		InspectExists: true,
		PlanExists:    true,
		TasksExists:   false,
		VerifyExists:  false,
	})

	tasksContent := "# Tasks: Demo\n\n## Phase 1: Implementation\n- [x] T1.1 Create the first slice\n- [ ] T1.2 Finish the second slice\n\n## Acceptance Coverage\n- AC-001 -> T1.1, T1.2\n"
	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte(tasksContent), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}

	checkStatus(statusPayload{
		Phase:          "implement",
		ReadyFor:       "implement",
		Blocked:        false,
		SpecExists:     true,
		InspectExists:  true,
		PlanExists:     true,
		TasksExists:    true,
		VerifyExists:   false,
		TasksTotal:     2,
		TasksCompleted: 1,
		TasksOpen:      1,
	})

	completeTasks := "# Tasks: Demo\n\n## Phase 1: Implementation\n- [x] T1.1 Create the first slice\n- [x] T1.2 Finish the second slice\n\n## Acceptance Coverage\n- AC-001 -> T1.1, T1.2\n"
	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte(completeTasks), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks complete) returned error: %v", err)
	}

	checkStatus(statusPayload{
		Phase:          "verify",
		ReadyFor:       "verify",
		Blocked:        false,
		SpecExists:     true,
		InspectExists:  true,
		PlanExists:     true,
		TasksExists:    true,
		VerifyExists:   false,
		TasksTotal:     2,
		TasksCompleted: 2,
		TasksOpen:      0,
	})

	verifyContent := "---\nreport_type: verify\nslug: demo\nstatus: pass\ndocs_language: en\ngenerated_at: 2026-03-30\n---\n# Verify Report: demo\n\n## Verdict\n\n- status: pass\n"
	if err := os.WriteFile(filepath.Join(planDir, "verify.md"), []byte(verifyContent), 0o644); err != nil {
		t.Fatalf("WriteFile(verify) returned error: %v", err)
	}

	checkStatus(statusPayload{
		Phase:          "verify",
		ReadyFor:       "archive",
		Blocked:        false,
		SpecExists:     true,
		InspectExists:  true,
		PlanExists:     true,
		TasksExists:    true,
		VerifyExists:   true,
		TasksTotal:     2,
		TasksCompleted: 2,
		TasksOpen:      0,
	})
}

func TestDashboardCommandJSONOutput(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	specPath := filepath.Join(ensureSpecDir(t, root, "demo"), "spec.md")
	if err := os.WriteFile(specPath, []byte("# Demo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "dashboard", root, "--json")
	if err != nil {
		t.Fatalf("dashboard --json returned error: %v", err)
	}

	var payload struct {
		Showing        string `json:"showing"`
		DisplayedCount int    `json:"displayed_count"`
		Features       []struct {
			Slug     string `json:"slug"`
			Archived bool   `json:"archived"`
		} `json:"features"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse dashboard json output %q: %v", stdout, err)
	}
	if payload.Showing != "active" {
		t.Fatalf("showing = %q, want active", payload.Showing)
	}
	if payload.DisplayedCount != 1 || len(payload.Features) != 1 || payload.Features[0].Slug != "demo" || payload.Features[0].Archived {
		t.Fatalf("unexpected dashboard payload: %+v", payload)
	}
}

func TestDashboardCommandJSONIncludesArchivedWithAllFlag(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	archiveSlugDir := filepath.Join(root, "archive", "old")
	if err := os.MkdirAll(archiveSlugDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(archiveSlugDir) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(archiveSlugDir, "summary.md"), []byte("archived"), 0o644); err != nil {
		t.Fatalf("WriteFile(archive summary) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "dashboard", root, "--all", "--json")
	if err != nil {
		t.Fatalf("dashboard --all --json returned error: %v", err)
	}

	var payload struct {
		Showing        string `json:"showing"`
		DisplayedCount int    `json:"displayed_count"`
		Features       []struct {
			Slug     string `json:"slug"`
			Archived bool   `json:"archived"`
		} `json:"features"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse dashboard json output %q: %v", stdout, err)
	}
	if payload.Showing != "active+archived" {
		t.Fatalf("showing = %q, want active+archived", payload.Showing)
	}
	var foundOld bool
	for _, f := range payload.Features {
		if f.Slug == "old" && f.Archived {
			foundOld = true
			break
		}
	}
	if !foundOld || payload.DisplayedCount != len(payload.Features) {
		t.Fatalf("expected archived feature in payload, got %+v", payload)
	}
}

func TestFeaturesCommandSummarizesProjectWorkflow(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ensureSpecDir(t, root, "alpha"), "spec.md"), []byte("# Alpha\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "features", root)
	if err != nil {
		t.Fatalf("features command returned error: %v", err)
	}
	if !strings.Contains(stdout, "slug") || !strings.Contains(stdout, "issues") || !strings.Contains(stdout, "alpha") || !strings.Contains(stdout, "inspect") || !strings.Contains(stdout, "verify") || !strings.Contains(stdout, "tasks") {
		t.Fatalf("unexpected features output: %s", stdout)
	}
}

func TestFeatureCommandShowsDetailedWorkflowView(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 Demo\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(filepath.Join(ensureSpecDir(t, root, "demo"), "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ensureSpecDir(t, root, "demo"), "inspect.md"), []byte("---\nreport_type: inspect\nslug: demo\nstatus: concerns\ndocs_language: en\ngenerated_at: 2026-03-31\n---\n# Inspect Report: demo\n\n## Verdict\n\n- status: concerns\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(inspect) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "feature", "demo", root)
	if err != nil {
		t.Fatalf("feature command returned error: %v", err)
	}
	if !strings.Contains(stdout, "inspect_status: concerns") || !strings.Contains(stdout, "ready_for: plan") || !strings.Contains(stdout, "issues:") || !strings.Contains(stdout, "focus: write the plan package") || strings.Contains(stdout, "verify_path:") {
		t.Fatalf("unexpected feature output: %s", stdout)
	}
}

func TestFeatureCommandShowsSemanticFindings(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 Demo\n- Given x\n- When y\n- Then z\n\n### AC-002 Demo\n- Given a\n- When b\n- Then c\n"
	if err := os.WriteFile(filepath.Join(ensureSpecDir(t, root, "demo"), "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}
	inspectContent := "---\nreport_type: inspect\nslug: demo\nstatus: pass\ndocs_language: en\ngenerated_at: 2026-03-31\n---\n# Inspect Report: demo\n\n## Verdict\n\n- status: pass\n"
	if err := os.WriteFile(filepath.Join(ensureSpecDir(t, root, "demo"), "inspect.md"), []byte(inspectContent), 0o644); err != nil {
		t.Fatalf("WriteFile(inspect) returned error: %v", err)
	}
	planDir := filepath.Join(root, "specs", "demo", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "plan.md"), []byte("# Demo Plan\n\n## Acceptance Approach\n- AC-001 -> path\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(plan) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte("# Tasks\n\n## Phase 1: Foundation\n- [ ] T1.1 do\n\n## Acceptance Coverage\n- AC-001 -> T1.1\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "feature", "demo", root)
	if err != nil {
		t.Fatalf("feature command returned error: %v", err)
	}
	if !strings.Contains(stdout, "warnings:") || !strings.Contains(stdout, "plan does not reference acceptance criterion AC-002") || strings.Contains(stdout, "for slug demo") {
		t.Fatalf("expected feature output to include semantic findings, got %s", stdout)
	}
}

func TestFeatureCommandShowsStructuredCheckDetails(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	specPath := filepath.Join(ensureSpecDir(t, root, "demo"), "spec.md")
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 The flow should feel fast.\n\n## Acceptance Criteria\n### AC-001 Demo\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "feature", "demo", root)
	if err != nil {
		t.Fatalf("feature command returned error: %v", err)
	}
	if !strings.Contains(stdout, "check_issues: warnings=6") {
		t.Fatalf("expected structured check summary in feature output, got %s", stdout)
	}
	if !strings.Contains(stdout, "check_detail: warning [structure] missing section:") {
		t.Fatalf("expected structured check detail in feature output, got %s", stdout)
	}
}

func TestFeatureCommandJSONIncludesStructuredCheckDetails(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	specPath := filepath.Join(ensureSpecDir(t, root, "demo"), "spec.md")
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 The flow should feel fast.\n\n## Acceptance Criteria\n### AC-001 Demo\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "feature", "demo", root, "--json")
	if err != nil {
		t.Fatalf("feature --json command returned error: %v", err)
	}

	var payload struct {
		CheckSummary struct {
			Warnings int `json:"warnings"`
		} `json:"check_summary"`
		CheckFindings []struct {
			Code string `json:"code"`
		} `json:"check_findings"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse feature json output %q: %v", stdout, err)
	}
	if payload.CheckSummary.Warnings != 6 {
		t.Fatalf("expected structured warnings in feature json, got %q", stdout)
	}
	found := false
	for _, finding := range payload.CheckFindings {
		if finding.Code == "ambiguous_wording" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected ambiguous_wording in feature json output, got %q", stdout)
	}
}

func TestCheckCommandShowsStructuredCheckSummary(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	specPath := filepath.Join(ensureSpecDir(t, root, "demo"), "spec.md")
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 The flow should feel fast.\n\n## Acceptance Criteria\n### AC-001 Demo\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "check", "demo", root)
	if err != nil {
		t.Fatalf("check command returned error: %v", err)
	}
	if !strings.Contains(stdout, "checks:   warnings=6") {
		t.Fatalf("expected check summary in output, got %s", stdout)
	}
	if !strings.Contains(stdout, "warning_categories=structure=3,ambiguity=2") {
		t.Fatalf("expected categorized warning summary, got %s", stdout)
	}
	if !strings.Contains(stdout, "detail:   warning [structure] missing section:") {
		t.Fatalf("expected finding detail in output, got %s", stdout)
	}
}

func TestCheckCommandJSONIncludesStructuredFindings(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	specPath := filepath.Join(ensureSpecDir(t, root, "demo"), "spec.md")
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 The flow should feel fast.\n\n## Acceptance Criteria\n### AC-001 Demo\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "check", "demo", root, "--json")
	if err != nil {
		t.Fatalf("check --json command returned error: %v", err)
	}

	var payload struct {
		CheckSummary struct {
			Warnings          int            `json:"warnings"`
			WarningCategories map[string]int `json:"warning_categories"`
		} `json:"check_summary"`
		CheckFindings []struct {
			Code     string `json:"code"`
			Category string `json:"category"`
		} `json:"check_findings"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse check json output %q: %v", stdout, err)
	}
	if payload.CheckSummary.Warnings != 6 {
		t.Fatalf("expected warnings in check summary, got %q", stdout)
	}
	if payload.CheckSummary.WarningCategories["ambiguity"] == 0 {
		t.Fatalf("expected ambiguity category in check summary, got %q", stdout)
	}
	found := false
	for _, finding := range payload.CheckFindings {
		if finding.Code == "ambiguous_wording" && finding.Category == "ambiguity" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected ambiguous_wording finding in json output, got %q", stdout)
	}
}

func TestCheckCommandBlocksWhenReadinessErrorsPresent(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	specPath := filepath.Join(ensureSpecDir(t, root, "demo"), "spec.md")
	// Missing required sections (Goal/Acceptance Criteria) → readiness errors.
	specContent := "# Demo\n\n## Requirements\n- RQ-001 x\n"
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "check", "demo", root, "--json")
	if err == nil {
		t.Fatalf("expected check --json to return an error when blocked")
	}

	var payload struct {
		Verdict      string `json:"verdict"`
		Blocked      bool   `json:"blocked"`
		NextCommand  string `json:"next_command"`
		CheckSummary struct {
			Errors int `json:"errors"`
		} `json:"check_summary"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse check json output %q: %v", stdout, err)
	}
	if !payload.Blocked || payload.Verdict != "blocked" {
		t.Fatalf("expected blocked verdict in json output, got %q", stdout)
	}
	if payload.NextCommand != "/speckeep.plan demo" {
		t.Fatalf("expected next_command plan, got %q", payload.NextCommand)
	}
	if payload.CheckSummary.Errors == 0 {
		t.Fatalf("expected errors > 0, got %q", stdout)
	}
}

func TestCheckCommandResolvesCustomSpecsDir(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh", "--specs-dir", "docs/specs", "--archive-dir", "docs/specs-archive"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	specPath := filepath.Join(root, "docs", "specs", "demo", "spec.md")
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 Demo\n- Given x\n- When y\n- Then z\n"
	if err := os.MkdirAll(filepath.Dir(specPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "check", "demo", root, "--json")
	if err != nil {
		t.Fatalf("check --json command returned error: %v", err)
	}

	var payload struct {
		Artifacts struct {
			Spec struct {
				Present bool `json:"present"`
			} `json:"spec"`
			Inspect struct {
				Present bool `json:"present"`
			} `json:"inspect"`
		} `json:"artifacts"`
		NextCommand string `json:"next_command"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse check json output %q: %v", stdout, err)
	}
	if !payload.Artifacts.Spec.Present {
		t.Fatalf("expected spec present=true in json output, got %q", stdout)
	}
	if payload.Artifacts.Inspect.Present {
		t.Fatalf("expected inspect present=false in json output, got %q", stdout)
	}
	if payload.NextCommand != "/speckeep.plan demo" {
		t.Fatalf("expected next_command plan, got %q", payload.NextCommand)
	}
}

func TestDoctorCommandPrefixesWorkspaceAndFeatureFindings(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 Demo\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(filepath.Join(ensureSpecDir(t, root, "demo"), "spec.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ensureSpecDir(t, root, "demo"), "inspect.md"), []byte("---\nreport_type: inspect\nslug: demo\nstatus: pass\ndocs_language: en\ngenerated_at: 2026-03-31\n---\n# Inspect Report: demo\n\n## Verdict\n\n- status: pass\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(inspect) returned error: %v", err)
	}
	planDir := filepath.Join(root, "specs", "demo", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "plan.md"), []byte("# Demo Plan\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(plan) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "doctor", root)
	if err != nil {
		t.Fatalf("doctor command returned error: %v", err)
	}
	if !strings.Contains(stdout, "[workspace]") || !strings.Contains(stdout, "[demo]") {
		t.Fatalf("expected doctor output to prefix workspace and feature findings, got %s", stdout)
	}
}

func TestAddListRemoveSkillCommands(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	localSkillDir := filepath.Join(root, "skills", "architecture")
	if err := os.MkdirAll(localSkillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "add-skill", root, "--id", "architecture", "--from-local", "skills/architecture")
	if err != nil {
		t.Fatalf("add-skill command returned error: %v", err)
	}
	if !strings.Contains(stdout, "added skill") {
		t.Fatalf("unexpected add-skill output: %s", stdout)
	}
	if !strings.Contains(stdout, "installed skills into agent folders") {
		t.Fatalf("expected add-skill to install skills into agent folders, got %s", stdout)
	}

	stdout, _, err = executeRoot(t, "list-skills", root)
	if err != nil {
		t.Fatalf("list-skills command returned error: %v", err)
	}
	if !strings.Contains(stdout, "architecture\tenabled\tskills/architecture") {
		t.Fatalf("unexpected list-skills output: %s", stdout)
	}

	stdout, _, err = executeRoot(t, "remove-skill", root, "--id", "architecture")
	if err != nil {
		t.Fatalf("remove-skill command returned error: %v", err)
	}
	if !strings.Contains(stdout, "removed skill \"architecture\"") {
		t.Fatalf("unexpected remove-skill output: %s", stdout)
	}
	if !strings.Contains(stdout, "installed skills into agent folders") {
		t.Fatalf("expected remove-skill to reconcile installed skills, got %s", stdout)
	}
}

func TestAddRemoveSkillNoInstallFlag(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh", "--agents", "codex"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	localSkillDir := filepath.Join(root, "skills", "architecture")
	if err := os.MkdirAll(localSkillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localSkillDir, "SKILL.md"), []byte("# Architecture\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(SKILL.md) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "add-skill", root, "--id", "architecture", "--from-local", "skills/architecture", "--no-install")
	if err != nil {
		t.Fatalf("add-skill --no-install command returned error: %v", err)
	}
	if !strings.Contains(stdout, "skipped skill installation into agent folders (--no-install)") {
		t.Fatalf("expected add-skill --no-install to report skipped install, got %s", stdout)
	}
	if _, err := os.Stat(filepath.Join(root, ".codex", "skills", "architecture")); !os.IsNotExist(err) {
		t.Fatalf("expected add-skill --no-install to not create .codex/skills/architecture, got err=%v", err)
	}

	installedDir := filepath.Join(root, ".codex", "skills", "architecture")
	if err := os.MkdirAll(installedDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(installedDir) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(installedDir, "SKILL.md"), []byte("# stale\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(stale SKILL.md) returned error: %v", err)
	}

	stdout, _, err = executeRoot(t, "remove-skill", root, "--id", "architecture", "--no-install")
	if err != nil {
		t.Fatalf("remove-skill --no-install command returned error: %v", err)
	}
	if !strings.Contains(stdout, "skipped skill installation into agent folders (--no-install)") {
		t.Fatalf("expected remove-skill --no-install to report skipped install, got %s", stdout)
	}
	if _, err := os.Stat(installedDir); err != nil {
		t.Fatalf("expected remove-skill --no-install to keep installed dir untouched, got err=%v", err)
	}
}

func TestListSkillsCommandJSON(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	gitSkillRepo := createGitSkillRepo(t, root)
	stdout, _, err := executeRoot(t, "add-skill", root, "--id", "openai-docs", "--from-git", gitSkillRepo, "--ref", "v1.2.3")
	if err != nil {
		t.Fatalf("add-skill command returned error: %v", err)
	}
	if !strings.Contains(stdout, "openai-docs") {
		t.Fatalf("unexpected add-skill output: %s", stdout)
	}

	stdout, _, err = executeRoot(t, "list-skills", root, "--json")
	if err != nil {
		t.Fatalf("list-skills --json command returned error: %v", err)
	}

	var payload struct {
		Skills []struct {
			ID   string `json:"id"`
			Ref  string `json:"ref"`
			From string `json:"source"`
		} `json:"skills"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse list-skills json output %q: %v", stdout, err)
	}
	if len(payload.Skills) != 1 || payload.Skills[0].ID != "openai-docs" || payload.Skills[0].Ref != "v1.2.3" {
		t.Fatalf("unexpected list-skills json payload: %s", stdout)
	}
}

func TestSkillsSyncCommandUpdatesAgentsBlock(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	localSkillDir := filepath.Join(root, "skills", "architecture")
	if err := os.MkdirAll(localSkillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	if _, _, err := executeRoot(t, "add-skill", root, "--id", "architecture", "--from-local", "skills/architecture"); err != nil {
		t.Fatalf("add-skill command returned error: %v", err)
	}

	agentsPath := filepath.Join(root, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("manual header\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(AGENTS.md) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "skills", "sync", root)
	if err != nil {
		t.Fatalf("skills sync command returned error: %v", err)
	}
	if !strings.Contains(stdout, "update AGENTS.md") {
		t.Fatalf("expected skills sync to update AGENTS.md, got %s", stdout)
	}

	content, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("ReadFile(AGENTS.md) returned error: %v", err)
	}
	if !strings.Contains(string(content), "architecture") {
		t.Fatalf("expected AGENTS.md to contain skill listing, got %q", string(content))
	}
}

func TestSkillsSyncCommandDryRunJSON(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	localSkillDir := filepath.Join(root, "skills", "architecture")
	if err := os.MkdirAll(localSkillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if _, _, err := executeRoot(t, "add-skill", root, "--id", "architecture", "--from-local", "skills/architecture"); err != nil {
		t.Fatalf("add-skill command returned error: %v", err)
	}

	agentsPath := filepath.Join(root, "AGENTS.md")
	original := "manual header\n"
	if err := os.WriteFile(agentsPath, []byte(original), 0o644); err != nil {
		t.Fatalf("WriteFile(AGENTS.md) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "skills", "sync", root, "--dry-run", "--json")
	if err != nil {
		t.Fatalf("skills sync --dry-run --json command returned error: %v", err)
	}

	var payload struct {
		DryRun  bool     `json:"dry_run"`
		Updated []string `json:"updated"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse skills sync json output %q: %v", stdout, err)
	}
	if !payload.DryRun {
		t.Fatalf("expected dry_run=true, got %s", stdout)
	}
	foundAgents := false
	for _, path := range payload.Updated {
		if path == "AGENTS.md" {
			foundAgents = true
			break
		}
	}
	if !foundAgents {
		t.Fatalf("expected AGENTS.md in updated list, got %s", stdout)
	}

	content, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("ReadFile(AGENTS.md) returned error: %v", err)
	}
	if string(content) != original {
		t.Fatalf("expected dry-run to keep AGENTS.md unchanged, got %q", string(content))
	}
}

func TestSyncSkillsAliasCommand(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	localSkillDir := filepath.Join(root, "skills", "architecture")
	if err := os.MkdirAll(localSkillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if _, _, err := executeRoot(t, "add-skill", root, "--id", "architecture", "--from-local", "skills/architecture"); err != nil {
		t.Fatalf("add-skill command returned error: %v", err)
	}

	agentsPath := filepath.Join(root, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("manual header\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(AGENTS.md) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "sync-skills", root)
	if err != nil {
		t.Fatalf("sync-skills command returned error: %v", err)
	}
	if !strings.Contains(stdout, "update AGENTS.md") {
		t.Fatalf("expected sync-skills to update AGENTS.md, got %s", stdout)
	}
}

func TestInstallSkillsCommandCopiesToAgentDirs(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh", "--agents", "codex,claude,windsurf,kilocode,trae"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	localSkillDir := filepath.Join(root, "skills", "architecture")
	if err := os.MkdirAll(localSkillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localSkillDir, "SKILL.md"), []byte("# Architecture\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(SKILL.md) returned error: %v", err)
	}

	if _, _, err := executeRoot(t, "add-skill", root, "--id", "architecture", "--from-local", "skills/architecture"); err != nil {
		t.Fatalf("add-skill command returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "install-skills", root)
	if err != nil {
		t.Fatalf("install-skills command returned error: %v", err)
	}
	if !strings.Contains(stdout, ".codex/skills/architecture") {
		t.Fatalf("expected install output to mention codex skills dir, got %s", stdout)
	}

	for _, rel := range []string{
		".codex/skills/architecture/SKILL.md",
		".claude/skills/architecture/SKILL.md",
		".windsurf/skills/architecture/SKILL.md",
		".kilocode/skills/architecture/SKILL.md",
		".trae/skills/architecture/SKILL.md",
	} {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected installed skill file %s to exist: %v", rel, err)
		}
	}
}

func TestSkillsInstallSubcommandDryRunJSON(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh", "--agents", "codex"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	localSkillDir := filepath.Join(root, "skills", "architecture")
	if err := os.MkdirAll(localSkillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localSkillDir, "SKILL.md"), []byte("# Architecture\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(SKILL.md) returned error: %v", err)
	}

	if _, _, err := executeRoot(t, "add-skill", root, "--id", "architecture", "--from-local", "skills/architecture"); err != nil {
		t.Fatalf("add-skill command returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "skills", "install", root, "--dry-run", "--json")
	if err != nil {
		t.Fatalf("skills install --dry-run --json command returned error: %v", err)
	}

	var payload struct {
		DryRun    bool     `json:"dry_run"`
		Created   []string `json:"created"`
		Unchanged []string `json:"unchanged"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse skills install json output %q: %v", stdout, err)
	}
	if !payload.DryRun {
		t.Fatalf("expected dry_run=true, got %s", stdout)
	}
	found := false
	for _, path := range payload.Created {
		if path == ".codex/skills/architecture" {
			found = true
			break
		}
	}
	if !found {
		for _, path := range payload.Unchanged {
			if path == ".codex/skills/architecture" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("expected .codex/skills/architecture in created or unchanged list, got %s", stdout)
	}
	if _, err := os.Stat(filepath.Join(root, ".codex", "skills", "architecture")); err != nil {
		t.Fatalf("expected add-skill to auto-install .codex/skills/architecture before dry-run, got err=%v", err)
	}
}

func TestFeatureCommandShowsLegacyInspectHint(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	// legacy flat inspect: specs/demo.inspect.md instead of specs/demo/inspect.md
	legacyInspectPath := filepath.Join(root, "specs", "demo.inspect.md")
	content := "---\nreport_type: inspect\nslug: demo\nstatus: concerns\ndocs_language: en\ngenerated_at: 2026-03-31\n---\n# Inspect Report: demo\n\n## Verdict\n\n- status: concerns\n"
	if err := os.WriteFile(legacyInspectPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(inspect) returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ensureSpecDir(t, root, "demo"), "spec.md"), []byte("# Demo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "feature", "demo", root)
	if err != nil {
		t.Fatalf("feature command returned error: %v", err)
	}
	if !strings.Contains(stdout, "inspect_legacy: true") {
		t.Fatalf("expected feature output to show legacy inspect hint, got %s", stdout)
	}
}

func TestFeatureRepairCommandMigratesLegacyFlatSpec(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	// legacy flat spec: specs/demo.md instead of specs/demo/spec.md
	legacySpecPath := filepath.Join(root, "specs", "demo.md")
	content := "# Demo Spec\n\n## Goal\nTest.\n\n## Requirements\n- RQ-001 test\n\n## Acceptance Criteria\n### AC-001\n- **Given** x\n- **When** y\n- **Then** z\n"
	if err := os.WriteFile(legacySpecPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "feature", "repair", "demo", root)
	if err != nil {
		t.Fatalf("feature repair command returned error: %v", err)
	}
	if !strings.Contains(stdout, "changed: true") || !strings.Contains(stdout, "move legacy") {
		t.Fatalf("unexpected feature repair output: %s", stdout)
	}
	if _, err := os.Stat(filepath.Join(root, "specs", "demo", "spec.md")); err != nil {
		t.Fatalf("expected canonical spec after repair: %v", err)
	}
}

func TestMigrateCommandRepairsLegacyFlatSpecsAcrossProject(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	// legacy flat specs: specs/alpha.md, specs/beta.md
	for _, slug := range []string{"alpha", "beta"} {
		legacySpecPath := filepath.Join(root, "specs", slug+".md")
		content := "# " + slug + " Spec\n\n## Goal\nTest.\n\n## Requirements\n- RQ-001 test\n\n## Acceptance Criteria\n### AC-001\n- **Given** x\n- **When** y\n- **Then** z\n"
		if err := os.WriteFile(legacySpecPath, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile returned error: %v", err)
		}
	}

	stdout, _, err := executeRoot(t, "migrate", root)
	if err != nil {
		t.Fatalf("migrate command returned error: %v", err)
	}
	if !strings.Contains(stdout, "slug: alpha") || !strings.Contains(stdout, "slug: beta") {
		t.Fatalf("unexpected migrate output: %s", stdout)
	}
	for _, slug := range []string{"alpha", "beta"} {
		if _, err := os.Stat(filepath.Join(root, "specs", slug, "spec.md")); err != nil {
			t.Fatalf("expected canonical spec for %s after migrate: %v", slug, err)
		}
	}
}

func TestCleanupAgentsCommandRemovesOrphanedArtifacts(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh", "--agents", "cursor"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}
	if _, _, err := executeRoot(t, "remove-agent", root, "--agents", "cursor"); err != nil {
		t.Fatalf("remove-agent command returned error: %v", err)
	}

	orphanPath := filepath.Join(root, ".cursor", "rules", "speckeep-inspect.mdc")
	if err := os.MkdirAll(filepath.Dir(orphanPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(orphanPath, []byte("orphan"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "cleanup-agents", root)
	if err != nil {
		t.Fatalf("cleanup-agents command returned error: %v", err)
	}
	if !strings.Contains(stdout, "removed orphaned agent artifact") {
		t.Fatalf("unexpected cleanup-agents output: %s", stdout)
	}
	if _, err := os.Stat(orphanPath); !os.IsNotExist(err) {
		t.Fatalf("expected orphaned file to be removed, got err=%v", err)
	}
}

func TestInitCommandRequiresShell(t *testing.T) {
	root := t.TempDir()

	_, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en")
	if err == nil {
		t.Fatal("expected init without --shell to fail")
	}
}

func TestRefreshCommandUpdatesManagedArtifacts(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh", "--agents", "claude"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	promptPath := filepath.Join(root, ".speckeep", "templates", "prompts", "inspect.md")
	if err := os.WriteFile(promptPath, []byte("stale prompt"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "refresh", root, "--shell", "powershell")
	if err != nil {
		t.Fatalf("refresh command returned error: %v", err)
	}
	if !strings.Contains(stdout, "update .speckeep/templates/prompts/inspect.md") {
		t.Fatalf("unexpected refresh output: %s", stdout)
	}

	if _, err := os.Stat(filepath.Join(root, ".speckeep", "scripts", "check-spec-ready.ps1")); err != nil {
		t.Fatalf("expected refreshed powershell script to exist: %v", err)
	}
}

func TestRefreshCommandRemovesLegacyArchiveArtifacts(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh", "--agents", "claude"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	legacyPrompt := filepath.Join(root, ".speckeep", "templates", "prompts", "archive.md")
	if err := os.MkdirAll(filepath.Dir(legacyPrompt), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(legacyPrompt, []byte("legacy prompt"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	legacyAgent := filepath.Join(root, ".claude", "commands", "speckeep.archive.md")
	if err := os.WriteFile(legacyAgent, []byte("legacy agent"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "refresh", root)
	if err != nil {
		t.Fatalf("refresh command returned error: %v", err)
	}
	if !strings.Contains(stdout, "remove .speckeep/templates/prompts/archive.md") || !strings.Contains(stdout, "remove .claude/commands/speckeep.archive.md") {
		t.Fatalf("unexpected refresh output: %s", stdout)
	}
}

func TestRefreshCommandCanMoveConstitutionFile(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	oldPath := filepath.Join(root, "CONSTITUTION.md")
	oldContent, err := os.ReadFile(oldPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	newRel := "docs/constitution.md"
	newAbs := filepath.Join(root, newRel)

	if _, _, err := executeRoot(t, "refresh", root, "--constitution-file", newRel); err != nil {
		t.Fatalf("refresh command returned error: %v", err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load config returned error: %v", err)
	}
	if cfg.Project.ConstitutionFile != newRel {
		t.Fatalf("expected constitution_file=%q, got %q", newRel, cfg.Project.ConstitutionFile)
	}

	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("expected old constitution file to be moved away, got err=%v", err)
	}

	newContent, err := os.ReadFile(newAbs)
	if err != nil {
		t.Fatalf("expected new constitution file to exist: %v", err)
	}
	if string(newContent) != string(oldContent) {
		t.Fatalf("expected moved constitution content to match")
	}
}

func TestRefreshCommandJSONDryRunOutput(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "refresh", root, "--shell", "powershell", "--dry-run", "--json")
	if err != nil {
		t.Fatalf("refresh command returned error: %v", err)
	}

	var payload struct {
		DryRun  bool     `json:"dry_run"`
		Updated []string `json:"updated"`
		Created []string `json:"created"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse refresh json output %q: %v", stdout, err)
	}
	if !payload.DryRun {
		t.Fatalf("expected dry_run true in refresh json output, got %q", stdout)
	}
	if len(payload.Updated) == 0 && len(payload.Created) == 0 {
		t.Fatalf("expected refresh json to report pending changes, got %q", stdout)
	}
}

func TestRefreshCommandCanMoveSpecsAndArchiveDirs(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh", "--specs-dir", ".speckeep/specs", "--archive-dir", ".speckeep/archive"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	specPath := filepath.Join(root, ".speckeep", "specs", "demo", "spec.md")
	if err := os.MkdirAll(filepath.Dir(specPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(specPath, []byte("# Demo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	archiveMarker := filepath.Join(root, ".speckeep", "archive", ".keep")
	if err := os.MkdirAll(filepath.Dir(archiveMarker), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(archiveMarker, []byte("keep"), 0o644); err != nil {
		t.Fatalf("WriteFile(archive marker) returned error: %v", err)
	}

	if _, _, err := executeRoot(t, "refresh", root, "--specs-dir", "specs", "--archive-dir", "archive"); err != nil {
		t.Fatalf("refresh command returned error: %v", err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load config returned error: %v", err)
	}
	if got, want := cfg.Paths.SpecsDir, "specs"; got != want {
		t.Fatalf("expected specs_dir=%q, got %q", want, got)
	}
	if got, want := cfg.Paths.ArchiveDir, "archive"; got != want {
		t.Fatalf("expected archive_dir=%q, got %q", want, got)
	}

	if _, err := os.Stat(filepath.Join(root, ".speckeep", "specs")); !os.IsNotExist(err) {
		t.Fatalf("expected old specs dir to be moved away, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".speckeep", "archive")); !os.IsNotExist(err) {
		t.Fatalf("expected old archive dir to be moved away, got err=%v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "specs", "demo", "spec.md")); err != nil {
		t.Fatalf("expected moved spec to exist, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "archive", ".keep")); err != nil {
		t.Fatalf("expected moved archive marker to exist, got err=%v", err)
	}
}

func TestInternalInspectSpecCommandUsesWorkflowBackend(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	specPath := filepath.Join(ensureSpecDir(t, root, "demo"), "spec.md")
	specContent := "# Demo\n\n## Goal\nx\n\n## Requirements\n- RQ-001 x\n\n## Acceptance Criteria\n### AC-001 Demo\n- Given x\n- When y\n- Then z\n"
	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("WriteFile(spec) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "__internal", "inspect-spec", "--root", root, "specs/demo/spec.md")
	if err != nil {
		t.Fatalf("internal inspect-spec command returned error: %v", err)
	}
	if !strings.Contains(stdout, "SUMMARY: errors=0") {
		t.Fatalf("unexpected internal inspect-spec output: %s", stdout)
	}
}

func TestInternalVerifyTaskStateCommandReturnsNonFatalWarnings(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	tasksPath := filepath.Join(root, "specs", "demo", "plan", "tasks.md")
	if err := os.MkdirAll(filepath.Dir(tasksPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(tasksPath, []byte("- [ ] T1.1 open\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "__internal", "verify-task-state", "--root", root, "demo")
	if err != nil {
		t.Fatalf("internal verify-task-state command returned error: %v", err)
	}
	if !strings.Contains(stdout, "TASKS_OPEN=1") || !strings.Contains(stdout, "WARN: open tasks remain") {
		t.Fatalf("unexpected internal verify-task-state output: %s", stdout)
	}
}

func TestInternalListOpenTasksCommandUsesCLIBackend(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	tasksPath := filepath.Join(root, "specs", "demo", "plan", "tasks.md")
	if err := os.MkdirAll(filepath.Dir(tasksPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(tasksPath, []byte("- [x] T1.1 done\n- [ ] T1.2 open\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks) returned error: %v", err)
	}

	stdout, _, err := executeRoot(t, "__internal", "list-open-tasks", "--root", root, "demo")
	if err != nil {
		t.Fatalf("internal list-open-tasks command returned error: %v", err)
	}
	if strings.TrimSpace(stdout) != "- [ ] T1.2 open" {
		t.Fatalf("unexpected internal list-open-tasks output: %q", stdout)
	}
}

func TestInternalLinkAgentsCommandUsesCLIBackend(t *testing.T) {
	root := t.TempDir()

	if _, _, err := executeRoot(t, "init", root, "--git=false", "--lang", "en", "--shell", "sh"); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	agentsPath := filepath.Join(root, "CUSTOM_AGENTS.md")
	snippetPath := filepath.Join(root, ".speckeep", "templates", "agents-snippet.md")

	stdout, _, err := executeRoot(t, "__internal", "link-agents", "--root", root, "CUSTOM_AGENTS.md", ".speckeep/templates/agents-snippet.md")
	if err != nil {
		t.Fatalf("internal link-agents command returned error: %v", err)
	}
	if !strings.Contains(stdout, "SpecKeep block added to CUSTOM_AGENTS.md") {
		t.Fatalf("unexpected internal link-agents output: %s", stdout)
	}

	content, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	snippet, err := os.ReadFile(snippetPath)
	if err != nil {
		t.Fatalf("ReadFile(snippet) returned error: %v", err)
	}
	if !strings.Contains(string(content), strings.TrimSpace(string(snippet))) {
		t.Fatalf("expected linked agents file to contain snippet, got %q", string(content))
	}
}
