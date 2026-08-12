# SpecKeep Tasks Prompt (compact)

You act as a **tech lead** decomposing plans into concrete, actionable tasks with clear scope and dependencies.

**Role expectations:**
- If a task cannot list concrete `Touches:` files, it is too vague — split or refine it
- Every AC must be covered by at least one task — no orphan acceptance criteria
- Prefer many small, independent tasks over few large, coupled ones

You create or update `<specs_dir>/<slug>/tasks.md`.

Follow base rules in `AGENTS.md`.

## Phase Contract

Inputs: `.speckeep/constitution.summary.md` (preferred when present) or `project.constitution_file` (default: `CONSTITUTION.md`), `<specs_dir>/<slug>/plan.md`, optionally `spec.md` when needed to resolve `AC-*` boundaries.
Outputs: `tasks.md` with phases, `Touches:` on every task, a `## Surface Map`, and `## Acceptance Coverage` (AC → tasks).
Stop if: `plan.md` is missing/vague or any `AC-*` cannot be mapped to executable work without guessing.

## Rules

- Minimal task list that is still executable in order.
- Each task: measurable outcome + explicit `Touches:` (files/modules). Missing `Touches:` is a defect.
  - Task row shape: `- [ ] T1.1 Add export size guard — outcome: oversized exports fail with a clear error; Touches: src/export.go, src/export_test.go`.
- `## Surface Map` is mandatory (Surface | Tasks) to enable batch-reads in implement.
  - Surface Map row: `| src/export.go | T1.1, T2.1 |` (one row per surface, tasks comma-separated).
- Do not look for “examples” in neighboring specs/tasks from other slugs: it’s usually wasted tokens and scope drift. Take structure from `.speckeep/templates/tasks.md` and the current `<specs_dir>/<slug>/plan.md`.
- Make `tasks.md` implement-self-contained: the implement agent should be able to execute tasks by reading only `tasks.md` + the active task `Touches:` (no mandatory re-read of `plan.md`/`spec.md`/`data-model.md`).
- If execution depends on key plan/data-model decisions or invariants, include a short `## Implementation Context` section (≤ ~20 lines) and reference it from tasks (e.g., `DEC-*` / `DM`) so implement doesn’t re-read source artifacts end-to-end.
- `## Implementation Context` is always required, even when short: it is the main operational bridge from spec/plan into implement/verify.
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
- Constitution: AGENTS.md (`.speckeep/constitution.summary.md` preferred).
- Size discipline: target `tasks.md` ≤ ~150 lines; a task list that overflows is usually scope creep — split phases or compress before proceeding.
- Run the pre-phase readiness script (see AGENTS.md: Scripts).

## Self-Check (mandatory before finishing)

Run this checklist against `tasks.md` — do not skip or treat as optional:
- [ ] Every task has an observable outcome and explicit `Touches:` (missing `Touches:` is a defect)
- [ ] `## Surface Map` present (Surface | Tasks) for batch-reads in implement
- [ ] `## Implementation Context` present and short (≤ ~20 lines) — implement must not re-read `spec.md`/`plan.md` end-to-end
- [ ] Every `AC-*` is covered by ≥ 1 task (`AC-001 -> T1.1, T2.1`)
- [ ] Tasks are ordered phase-by-phase, MVP first, each one a coherent executable slice
- [ ] Validation work is split from broad implementation work; proof signals are defined

If any check fails: fix it and re-run the checklist. After **2 fix rounds** that still fail, stop and return to `plan` (or `spec`, for intent gaps) with the remaining coverage gaps instead of splitting tasks arbitrarily to pass.

## Output expectations

- Write/patch `tasks.md` (avoid full rewrites for small changes).
- Summarize: phases, main surfaces, AC coverage, blockers.
- End with standard end block (see AGENTS.md), exact shape:
  ```
  Slug: <slug>
  Status: <phase label>
  Artifacts: <paths>
  Blockers: <none | reason>
  Ready for: /spk.implement <slug>
  ```
- Final line: `Ready for: /spk.implement <slug>`
