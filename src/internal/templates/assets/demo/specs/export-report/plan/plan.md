# Export Report (CSV) Plan

## Phase Contract

Inputs: spec and minimal repo context for this feature.
Outputs: plan, data model, and contracts when required.
Stop if: the spec is too vague to plan safely.

## Goal

Add an authenticated export endpoint that produces a CSV matching the current filtered report table view, with stable column header + order, and a consistent filename.

## Scope

- Server-side endpoint to export the filtered view as CSV
- Client-side action to trigger the download from the current view
- Tests proving filter parity, header order, empty-result behavior, and authentication

## Implementation Surfaces

- `server/report_export` (new) — export handler: parse filters, query rows, render CSV
- `server/auth` (existing) — enforce authentication on the export endpoint
- `server/report_filters` (existing) — reuse filter parsing and validation shared with table view
- `client/report_table_toolbar` (existing) — add "Export as CSV" button that calls the endpoint with the current filter state
- `tests/report_export` (new) — integration + unit coverage for CSV output and auth

## Architecture Impact

- No new persistent state.
- Adds one new endpoint and a small CSV rendering component.

## Acceptance Approach

- AC-001 Implement export handler that reuses the same filter parsing path as the table view; integration test asserts parity between table view and export for the same filter state.
- AC-002 Set `Content-Disposition` with `report-YYYY-MM-DD.csv` filename derived from server-side date; unit or integration test asserts header.
- AC-003 Ensure CSV renderer outputs a header-only CSV for empty result sets; unit test for renderer + integration smoke test.
- RQ-003 Validate unauthenticated requests return 401; integration test.

## Data and Contracts

- Endpoint: `GET /reports/export?<filter_params>` → `text/csv`
- Response headers:
  - `Content-Type: text/csv`
  - `Content-Disposition: attachment; filename="report-YYYY-MM-DD.csv"`

## Implementation Strategy

- DEC-001 Synchronous export (no background jobs)
  Why: the scope caps datasets at 50k rows and does not require history or scheduling.
  Tradeoff: response time scales with row count.
  Affects: AC-001, AC-003
  Validation: load test and timeout budget in CI.

- DEC-002 Reuse filter parsing from the table view
  Why: prevents divergence between "what you see" and "what you export".
  Tradeoff: export handler becomes coupled to shared filter semantics.
  Affects: AC-001, RQ-001
  Validation: parity integration test.

- DEC-003 Column order is server-owned
  Why: avoids client-driven inconsistency and keeps CSV stable.
  Tradeoff: client cannot request arbitrary column reorder.
  Affects: RQ-002
  Validation: unit test for column registry → CSV header order.

## Incremental Delivery

### MVP (First Value)

- Backend export endpoint with auth + header-only empty export
- Minimal UI button to trigger download
- Integration tests for auth + filter parity

### Iterative Expansion

- Add performance budget checks and improvements for large exports
- Improve client UX (spinner, error toast) without changing export contract

## Sequencing Notes

- Build backend + CSV renderer first, then wire UI.
- Add tests before polishing UX so correctness stays locked in.

## Risks

- Filter edge cases (date ranges, invalid values) could diverge between table and export.
  Mitigation: reuse the same filter path and add parity tests.

## Rollout and Compatibility

- No migration needed.
- Additive endpoint and UI control; can be shipped behind a feature flag if desired.

## Validation

- Unit: CSV renderer header order + empty result handling
- Integration: authenticated export equals table view results for the same filter set
- Integration: unauthenticated export returns 401
- Manual: filename pattern in browser download dialog

## Constitution Compliance

- no conflicts
