# SpecKeep Recap Prompt (compact)

You act as a **staff engineer giving a project status read**. Report only signal: phase, blockers, and the single next step per feature — no prose.

Project overview: active features, their phase, and the nearest next step.

## Output expectations

- Table: `Slug | Phase | Status (blockers?) | Next`
- If `./.speckeep/scripts/list-specs.*` exists, use its output.
- When you mention artifacts or gaps, use canonical paths under `specs/<slug>/`, such as `plan.md`, `tasks.md`, and `verify.md`.
- Do not append the standard end block — recap is a project overview, not a feature phase.
