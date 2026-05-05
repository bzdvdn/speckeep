# Export Report (CSV) Tasks

## Phase Contract

Inputs: plan and minimal supporting artifacts for this feature.
Outputs: ordered executable tasks with coverage mapping.
Stop if: tasks would be vague or acceptance coverage cannot be mapped.

## Surface Map

| Surface | Tasks |
|---------|-------|
| server/report_export | T1.1, T1.2, T2.1 |
| server/report_filters | T1.2 |
| server/auth | T2.1 |
| client/report_table_toolbar | T2.2 |
| tests/report_export | T3.1 |

## Phase 1: Foundation

Goal: establish the minimum structure so later work stays predictable.

- [x] T1.1 Establish export endpoint scaffold — route exists and is wired to an export handler. Touches: server/report_export
- [x] T1.2 Reuse the table-view filter parsing path inside the export handler; add a parity test harness. Touches: server/report_export, server/report_filters

## Phase 2: Core Implementation

Goal: deliver the primary behavior and the key edge cases.

- [ ] T2.1 Enforce authentication on the export endpoint and return 401 for unauthenticated requests. Touches: server/auth, server/report_export (RQ-003)
- [ ] T2.2 Implement CSV rendering: stable column header + order, plus `Content-Disposition` filename behavior. Touches: server/report_export (AC-001, AC-002, RQ-002)
- [ ] T2.3 Implement empty-result export (header-only CSV) and ensure it matches table column registry. Touches: server/report_export (AC-003)
- [ ] T2.4 Add "Export as CSV" button in the report table toolbar and pass current filter state via query params. Touches: client/report_table_toolbar (AC-001)

## Phase 3: Validation

Goal: prove the feature works and guard regressions.

- [ ] T3.1 Add automated coverage: auth 401, filter parity, header order, and empty-result behavior. Touches: tests/report_export (AC-001, AC-002, AC-003, RQ-003)
- [ ] T3.2 Run verification checklist and update docs if needed (e.g., README note on export limits). Touches: docs/

## Acceptance Coverage

- AC-001 -> T1.2, T2.2, T2.4, T3.1
- AC-002 -> T2.2, T3.1
- AC-003 -> T2.3, T3.1
- RQ-003 -> T2.1, T3.1

## Notes

- Keep task ordering aligned with the plan and avoid mixing validation into core implementation tasks.
