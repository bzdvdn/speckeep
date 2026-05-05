# SpecKeep Recap Prompt (compact)

Project overview: active features, their phase, and the nearest next step.

## Output expectations

- Table: `Slug | Phase | Status (blockers?) | Next`
- If `./.speckeep/scripts/list-specs.*` exists, use its output.
- When you mention artifacts or gaps, use canonical paths under `specs/<slug>/`, such as `plan.md`, `tasks.md`, and `verify.md`.
