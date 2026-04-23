# SpecKeep Tasks Prompt (compact)

You create or update `<specs_dir>/<slug>/plan/tasks.md`.

Follow base rules in `AGENTS.md`.

## Path Resolution

- Resolve `<specs_dir>` from `.speckeep/speckeep.yaml` (read ≤1 time per session). If the config is missing, use `specs`.

## Phase Contract

Inputs: `project.constitution_file` (default: `CONSTITUTION.md`, or `.speckeep/constitution.summary.md` if present), `<specs_dir>/<slug>/plan/plan.digest.md` (preferred) or `plan.md`, optionally `spec.digest.md` (preferred) or `summary.md`/`spec.md` to resolve `AC-*` boundaries.
Outputs: `tasks.md` with phases, `Touches:` on every task, a `## Surface Map`, and `## Acceptance Coverage` (AC → tasks).
Stop if: `plan.md` is missing/vague or any `AC-*` cannot be mapped to executable work without guessing.

## Rules

- Minimal task list that is still executable in order.
- Each task: measurable outcome + explicit `Touches:` (files/modules). Missing `Touches:` is a defect.
- `## Surface Map` is mandatory (Surface | Tasks) to enable batch-reads in implement.
- Do not look for “examples” in neighboring specs/tasks from other slugs: it’s usually wasted tokens and scope drift. Take structure from `.speckeep/templates/tasks.md` and the current `<specs_dir>/<slug>/plan/plan.md`.
- Make `tasks.md` implement-self-contained: the implement agent should be able to execute tasks by reading only `tasks.md` + the active task `Touches:` (no mandatory re-read of `plan.md`/`spec.md`/`data-model.md`).
- If execution depends on key plan/data-model decisions or invariants, include a short `## Implementation Context` section (≤ ~20 lines) and reference it from tasks (e.g., `DEC-*` / `DM`) so implement doesn’t re-read source artifacts end-to-end.
- Recommended `## Implementation Context` template (keep it short, no fluff):
  - `MVP goal:` (1 line)
  - `Invariants/semantics:` (2–5 bullets)
  - `Errors/codes:` (1–3 bullets)
  - `Contracts/protocol:` (1–3 bullets: paths/formats)
  - `Scope boundaries:` (2 bullets “do not …”)
  - `Proof signals:` (1–3 bullets “what counts as proof”)
  - `References (opt.):` `DEC-*`, `DM`, `RQ-*` (without mandatory re-reading)
- Every `AC-*` must be covered by ≥ 1 task: `AC-001 -> T1.1, T2.1`.
- Do not implement or edit source code in the tasks phase.
- Do not assume `research.md` should exist; only reference it when the plan explicitly depends on it.
- If `./.speckeep/scripts/check-tasks-ready.*` exists, run it (slug first) as a cheap gate.

## Output expectations

- Write/patch `tasks.md` (avoid full rewrites for small changes).
- Summarize: phases, main surfaces, AC coverage, blockers.
- Include a short summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`.
- Final line: `Ready for: /speckeep.implement <slug>`
