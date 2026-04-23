---
report_type: verify
slug: <slug>
status: pass
docs_language: <en|ru>
generated_at: <YYYY-MM-DD>
---

# Verify Report: <slug>

## Scope

- snapshot: one-line summary of what was verified
- verification_mode: default | deep
- artifacts:
  - CONSTITUTION.md
  - specs/<slug>/plan/tasks.md
- inspected_surfaces:
  - list only the code paths, endpoints, jobs, docs, or migrations you actually checked

## Verdict

- status: pass
- archive_readiness: safe
- summary: one-line reason this verdict is justified

## Checks

- task_state: completed=<n>, open=<n>; name any still-open or disputed task IDs
- acceptance_evidence:
  - AC-001 -> confirmed via T1.1 and the specific surface inspected
- implementation_alignment:
  - name the concrete behavior, file, endpoint, or flow that matched the task claim

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
