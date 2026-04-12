# Export Report (CSV) Data Model

## ExportRequest

Value object passed from the export handler to the CSV renderer. Not persisted.

| Field | Type | Notes |
|------|------|-------|
| Filters | FilterSet | Parsed from query params using the same filter parsing path as the report table view |
| Columns | []Column | Ordered list from the server-owned column registry (DEC-003) |
| ExportDate | Date | Server-side date used for filename formatting (AC-002) |

Invariants:

- `Columns` must be non-nil; an empty slice is valid and produces a header-only CSV (AC-003).
- `Filters` must be valid; invalid filters are rejected before creating this value.

## Column

| Field | Type | Notes |
|------|------|-------|
| Key | string | Internal column key |
| Header | string | Display label used as CSV header |

Justification: AC-001 (headers), RQ-002 (order matches table view).

## No New Persistent Entities

This feature adds no new database tables, migrations, or persisted state.
