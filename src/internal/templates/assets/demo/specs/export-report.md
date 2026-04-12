# Export Report (CSV)

## Scope Snapshot

- In scope: Export the current filtered report table view as a CSV download.
- Out of scope: PDF/XLSX exports, background jobs, and export history.

## Goal

Allow authenticated users to download the current filtered table view as a CSV file so they can analyze the data in external tools without manual copy-paste.

## Primary User Flow

1. User opens a report table view with filters applied.
2. User clicks "Export as CSV".
3. System generates a CSV matching the current filtered view.
4. Browser downloads the file named `report-YYYY-MM-DD.csv`.
5. User opens the file and sees all visible columns and matching rows.

## Scope

- CSV download for the current filtered report table view
- Authentication gate on the export endpoint
- Empty-result handling (header-only CSV)
- Stable column header + order that matches the table view

## Context

- Report tables are server-rendered; filter state is held in query parameters.
- The export endpoint can reuse the existing filter parsing logic already used by the table view.

## Requirements

- RQ-001 Export MUST reflect active filters (not the full unfiltered dataset).
- RQ-002 Column order in CSV MUST match the visible column order in the table view.
- RQ-003 The export endpoint MUST require authentication; unauthenticated requests return 401.

## Non-Goals

- PDF or Excel export formats
- Scheduled/email-delivered exports
- Export history or audit log
- Client-side "custom column selection" UI
- Streaming for datasets larger than 50k rows

## Acceptance Criteria

### AC-001 Filtered export

- Why this matters: users must get exactly what they are looking at, without manual cleanup.
- **Given** an authenticated user viewing a report table with an active filter
- **When** they click "Export as CSV"
- **Then** a CSV file is downloaded containing all rows matching the active filter, with column headers in the first row
- Evidence: downloaded CSV rows match the table view for the same filter state.

### AC-002 File naming

- Why this matters: users can file and share exports consistently.
- **Given** a successful export response
- **When** the browser begins the download
- **Then** the filename follows `report-YYYY-MM-DD.csv` using the server-side date
- Evidence: response includes `Content-Disposition` with the expected filename.

### AC-003 Empty result

- Why this matters: exports should remain useful even when a filter matches nothing.
- **Given** an authenticated user viewing a report table with zero matching rows
- **When** they click "Export as CSV"
- **Then** a CSV file is downloaded containing only the header row and no data rows
- Evidence: the CSV has exactly 1 row (headers) and valid column names.

## Assumptions

- The report dataset is capped at 50k rows for this feature scope.
- Column definitions (header labels and order) are managed server-side.

## Success Criteria

- SC-001 Export completes in under 5 seconds for 10k rows on typical production hardware.

## Edge Cases

- Session expires mid-export → 401; client shows the standard auth error.
- Filter values are invalid → request is rejected consistently with table view behavior.

## Open Questions

none
