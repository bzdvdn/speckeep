# SpecKeep Tasks Prompt (compact)

You create or update `<specs_dir>/<slug>/plan/tasks.md`.

Follow base rules in `AGENTS.md`.

## Path Resolution

- Resolve `<specs_dir>` from `.speckeep/speckeep.yaml` (read ≤1 time per session). If the config is missing, use `.speckeep/specs`.

## Phase Contract

Inputs: `.speckeep/constitution.md` (or `.speckeep/constitution.summary.md`), `<specs_dir>/<slug>/plan/plan.md`, optionally `summary.md`/`spec.md` to resolve `AC-*` boundaries.
Outputs: `tasks.md` with phases, `Touches:` on every task, a `## Surface Map`, and `## Acceptance Coverage` (AC → tasks).
Stop if: `plan.md` is missing/vague or any `AC-*` cannot be mapped to executable work without guessing.

## Rules

- Minimal task list that is still executable in order.
- Each task: measurable outcome + explicit `Touches:` (files/modules). Missing `Touches:` is a defect.
- `## Surface Map` is mandatory (Surface | Tasks) to enable batch-reads in implement.
- Every `AC-*` must be covered by ≥ 1 task: `AC-001 -> T1.1, T2.1`.
- Do not implement or edit source code in the tasks phase.
- Do not assume `research.md` should exist; only reference it when the plan explicitly depends on it.
- If `/.speckeep/scripts/check-tasks-ready.*` exists, run it (slug first) as a cheap gate.

## Output expectations

- Write/patch `tasks.md` (avoid full rewrites for small changes).
- Summarize: phases, main surfaces, AC coverage, blockers.
- Include a short summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`.
- Final line: `Ready for: /speckeep.implement <slug>`
