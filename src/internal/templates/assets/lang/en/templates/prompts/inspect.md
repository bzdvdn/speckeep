# SpecKeep Inspect Prompt (compact)

You inspect one feature for consistency and readiness to plan.

Follow base rules in `AGENTS.md`.

## Path Resolution

- Resolve `<specs_dir>` from `.speckeep/speckeep.yaml` (read ≤1 time per session). If the config is missing, use `.speckeep/specs`.

## Phase Contract

Inputs: `.speckeep/constitution.md`, `<specs_dir>/<slug>/spec.md`.
Outputs: `<specs_dir>/<slug>/inspect.md` with `pass|concerns|blocked` and `<specs_dir>/<slug>/summary.md`.
Stop if: spec missing, slug ambiguous, or the verdict would require inventing product intent.

## Checks (strict but cheap)

- Always start with the cheapest scope: constitution + spec, then plan, then tasks. Do not jump to code unless a concrete claim cannot be confirmed from artifacts.
- Avoid repetitive full-file reads “for reassurance”: keep brief notes and re-open only targeted sections when needed.
- Take the report format from `.speckeep/templates/inspect.md`. Do not look for “examples” in other slugs’ inspect reports for shape: it’s wasted tokens and scope drift.
- Constitution ↔ spec: no conflicts with constraints, workflow rules, and language policy.
- `AC-*`: every AC uses Given/When/Then; no placeholders; no open `[NEEDS CLARIFICATION: ...]`.
- Scope: exactly one feature; explicit Out of Scope + Assumptions + Open Questions (or `none`).
- Technology mentions: treat technology names, frameworks, library lists, or version pins in the spec as a Warning unless they are a user requirement, repository constraint, or external contract.
- Ambiguity: flag vague adjectives (fast, scalable, secure, intuitive, robust) without measurable criteria as Warnings; if it blocks planning, treat as blocked.
- Placeholders: any `TODO`, `TKTK`, `???`, `<placeholder>` or similar unresolved marker is an Error.
- If `<specs_dir>/<slug>/plan/plan.md` exists: verify `spec <-> plan` (goal/scope preserved; no new major workstreams).
- If `<specs_dir>/<slug>/plan/tasks.md` exists: verify `plan <-> tasks` and AC coverage (each `AC-*` covered by ≥ 1 task).
- If `<specs_dir>/<slug>/plan/tasks.md` exists: treat missing `Touches:` as a Warning (token-discipline defect) because it forces broad reads during implement.

If `/.speckeep/scripts/check-inspect-ready.*` exists, run it (slug first) and use its output as a baseline. Do not read `/.speckeep/scripts/*` source.

## Output expectations

- Write `inspect.md` and `summary.md` (summary ~≤25 lines: Goal, AC table, Out of Scope).
- `inspect.md` MUST include: verdict, Errors, Warnings, and Next step (when not blocked).
- For `blocked`, do not suggest the next phase command; state which refinement is required first.
- In chat: compact verdict + non-empty Errors/Warnings + Next step.
- Final line: `Ready for: /speckeep.plan <slug>`
