package workflow

// TestFullWorkflowCycle exercises the complete spec → inspect → plan → tasks →
// implement → verify lifecycle. It uses the Check* readiness functions and
// verifies State transitions at every phase boundary. The goal is to catch
// regressions that span multiple packages but cannot be detected by unit tests
// that touch only one phase at a time.
//
// Workflow path tested:
//
//	init → spec (ready for plan; inspect optional) → inspect (pass) → plan → tasks (1 open) → tasks (all done)
//	→ verify (pass) → archive ready
//
// Each step:
//  1. Writes the minimum valid artifact for that phase.
//  2. Asserts State.ReadyFor transitions to the expected next phase.
//  3. Calls the corresponding Check*Ready function and asserts it does not fail.

import (
	"os"
	"path/filepath"
	"testing"

	"speckeep/src/internal/project"
)

func TestFullWorkflowCycle(t *testing.T) {
	root := t.TempDir()
	slug := "demo"

	// ── init ──────────────────────────────────────────────────────────────
	_, err := project.Initialize(root, project.InitOptions{
		InitGit:     false,
		DefaultLang: "en",
		Shell:       "sh",
	})
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	assertState(t, root, slug, "spec", "constitution")

	specDir := filepath.Join(root, "specs", slug)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(specDir): %v", err)
	}
	planDir := filepath.Join(specDir, "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(planDir): %v", err)
	}
	// Create source files referenced in tasks Touches: fields so that
	// checkTouchesFilesExist passes during CheckVerifyReady.
	for _, rel := range []string{"src/handlers/export.go", "src/middleware/auth.go"} {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatalf("MkdirAll(%s): %v", filepath.Dir(abs), err)
		}
		if err := os.WriteFile(abs, []byte("package stub\n"), 0o644); err != nil {
			t.Fatalf("WriteFile(%s): %v", rel, err)
		}
	}

	// ── spec ──────────────────────────────────────────────────────────────
	writeFile(t, filepath.Join(specDir, "spec.md"), specMD)

	assertState(t, root, slug, "plan", "spec")
	assertCheckPasses(t, "CheckInspectReady", func() (CheckResult, error) {
		return CheckInspectReady(root, slug)
	})

	// ── inspect ───────────────────────────────────────────────────────────
	writeFile(t, filepath.Join(specDir, "summary.md"), summaryMD)
	writeFile(t, filepath.Join(specDir, "inspect.md"), inspectMD)

	assertState(t, root, slug, "plan", "inspect")
	assertCheckPasses(t, "CheckPlanReady", func() (CheckResult, error) {
		return CheckPlanReady(root, slug)
	})

	// ── plan ──────────────────────────────────────────────────────────────
	writeFile(t, filepath.Join(planDir, "plan.md"), planMD)
	writeFile(t, filepath.Join(planDir, "data-model.md"), "# Data Model\nNo new entities required.\n")

	assertState(t, root, slug, "tasks", "plan")
	assertCheckPasses(t, "CheckTasksReady", func() (CheckResult, error) {
		return CheckTasksReady(root, slug)
	})

	// ── tasks (one open — implement phase) ────────────────────────────────
	writeFile(t, filepath.Join(planDir, "tasks.md"), tasksMDOpen)

	assertState(t, root, slug, "implement", "implement")
	assertCheckPasses(t, "CheckImplementReady", func() (CheckResult, error) {
		return CheckImplementReady(root, slug)
	})

	// ── tasks (all done — verify phase) ───────────────────────────────────
	writeFile(t, filepath.Join(planDir, "tasks.md"), tasksMDDone)

	assertState(t, root, slug, "verify", "verify")
	assertCheckPasses(t, "CheckVerifyReady", func() (CheckResult, error) {
		return CheckVerifyReady(root, slug)
	})

	// ── verify ────────────────────────────────────────────────────────────
	writeFile(t, filepath.Join(planDir, "verify.md"), verifyMD)

	assertState(t, root, slug, "archive", "verify")
	archiveCheck, err := CheckArchiveReady(root, slug, "completed", "")
	if err != nil {
		t.Fatalf("CheckArchiveReady: %v", err)
	}
	if archiveCheck.Failed {
		t.Fatalf("CheckArchiveReady failed:\n%v", archiveCheck.Lines)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

// assertState verifies that State returns the expected ReadyFor and Phase values.
func assertState(t *testing.T, root, slug, wantReadyFor, wantPhase string) {
	t.Helper()
	state, err := State(root, slug)
	if err != nil {
		t.Fatalf("State(%s): %v", slug, err)
	}
	if state.ReadyFor != wantReadyFor {
		t.Fatalf("State.ReadyFor = %q, want %q (phase=%s)", state.ReadyFor, wantReadyFor, state.Phase)
	}
	if state.Phase != wantPhase {
		t.Fatalf("State.Phase = %q, want %q (readyFor=%s)", state.Phase, wantPhase, state.ReadyFor)
	}
}

// assertCheckPasses runs a Check* function and fails the test if it returns a
// blocking error. Non-blocking warnings are allowed and expected.
func assertCheckPasses(t *testing.T, name string, fn func() (CheckResult, error)) {
	t.Helper()
	result, err := fn()
	if err != nil {
		t.Fatalf("%s returned error: %v", name, err)
	}
	if result.Failed {
		t.Fatalf("%s has blocking errors:\n%v", name, result.Lines)
	}
}

// ── fixture content ───────────────────────────────────────────────────────────

const specMD = `# Demo Feature

## Goal
Enable PDF export of reports.

## Requirements
- RQ-001 The system must export reports as PDF.
- RQ-002 The export must be accessible to all authenticated users.

## Acceptance Criteria

### AC-001 PDF Export Available
- Given a user is on the report page
- When they click "Export as PDF"
- Then a PDF file is downloaded to the browser

### AC-002 Export Access for Authenticated Users
- Given any authenticated user
- When they attempt to export any report
- Then the export completes without errors

## Assumptions

- Browser-based PDF generation is sufficient for v1.
`

const summaryMD = `---
slug: demo
generated_at: 2026-04-16
---

## Goal
Enable PDF export of reports.

## Acceptance Criteria
| ID | Summary | Proof Signal |
|---|---|---|
| AC-001 | PDF export available | PDF file downloaded |
| AC-002 | Export access for all | Export completes |

## Out of Scope
- Batch export
- Email delivery
- Non-PDF formats
`

const inspectMD = `---
report_type: inspect
slug: demo
status: pass
docs_language: en
generated_at: 2026-04-16
---
# Inspect Report: demo

## Scope
- snapshot: spec verified for completeness and constitutional consistency
- artifacts:
  - CONSTITUTION.md
  - specs/demo/spec.md

## Verdict
- status: pass

## Errors
- none

## Warnings
- none

## Questions
- none

## Suggestions
- none

## Traceability
- AC-001 and AC-002 both have Given/When/Then structure; no tasks yet

## Next Step
- safe to continue to plan
`

const planMD = `# Demo Plan

## DEC-001 Browser PDF API
Use the browser's built-in print-to-PDF for initial implementation.
Chosen for simplicity and zero additional dependencies.

## Acceptance Approach
- AC-001 -> implement export button and download handler
- AC-002 -> ensure auth middleware applies to the export route

## Constitution Compliance
- No architectural constraints violated.
- Follows repository language policy.

## Implementation Surfaces
- src/handlers/export.go
- src/middleware/auth.go
`

// tasksMDOpen has one open task so that State.ReadyFor becomes "implement".
const tasksMDOpen = `# Demo Tasks

## Surface Map
| Surface | Tasks |
|---------|-------|
| src/handlers/export.go | T1.1 |
| src/middleware/auth.go | T1.2 |

## Phase 1: Implementation

- [ ] T1.1 Add export endpoint. Touches: src/handlers/export.go
- [x] T1.2 Verify auth middleware coverage. Touches: src/middleware/auth.go

## Acceptance Coverage
- AC-001 -> T1.1
- AC-002 -> T1.2
`

// tasksMDDone has all tasks completed so that State.ReadyFor becomes "verify".
const tasksMDDone = `# Demo Tasks

## Surface Map
| Surface | Tasks |
|---------|-------|
| src/handlers/export.go | T1.1 |
| src/middleware/auth.go | T1.2 |

## Phase 1: Implementation

- [x] T1.1 Add export endpoint. Touches: src/handlers/export.go
- [x] T1.2 Verify auth middleware coverage. Touches: src/middleware/auth.go

## Acceptance Coverage
- AC-001 -> T1.1
- AC-002 -> T1.2
`

const verifyMD = `---
report_type: verify
slug: demo
status: pass
docs_language: en
generated_at: 2026-04-16
---
# Verify Report: demo

## Scope
- snapshot: all tasks verified against implementation
- verification_mode: default
- artifacts:
  - CONSTITUTION.md
  - specs/demo/plan/tasks.md
- inspected_surfaces:
  - src/handlers/export.go
  - src/middleware/auth.go

## Verdict
- status: pass
- archive_readiness: safe
- summary: all tasks complete, AC-001 and AC-002 confirmed via Touches inspection

## Checks
- task_state: completed=2, open=0
- acceptance_evidence:
  - AC-001 -> confirmed via T1.1 export handler in src/handlers/export.go
  - AC-002 -> confirmed via T1.2 auth middleware in src/middleware/auth.go
- implementation_alignment:
  - src/handlers/export.go implements the export endpoint described in T1.1

## Errors
- none

## Warnings
- none

## Questions
- none

## Not Verified
- none

## Next Step
- safe to archive
`
