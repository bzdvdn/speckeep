---
report_type: verify
slug: <slug>
status: pass
docs_language: <en|ru>
generated_at: <YYYY-MM-DD>
---

# Verify Report: <slug>

## Scope

- snapshot: однострочное резюме того, что проверили
- verification_mode: default | deep
- artifacts:
  - .speckeep/constitution.md
  - .speckeep/specs/<slug>/plan/tasks.md
- inspected_surfaces:
  - перечислите только те code paths, endpoints, jobs, docs или migrations, которые реально проверили

## Verdict

- status: pass
- archive_readiness: safe
- summary: однострочная причина, почему этот verdict обоснован

## Checks

- task_state: completed=<n>, open=<n>; укажите still-open или спорные task IDs
- acceptance_evidence:
  - AC-001 -> подтверждено через T1.1 и конкретную проверенную поверхность
- implementation_alignment:
  - назовите конкретное поведение, file, endpoint или flow, который подтвердил task claim

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
