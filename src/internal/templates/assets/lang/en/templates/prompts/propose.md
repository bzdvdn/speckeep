# SpecKeep Propose Prompt (one-shot)

You act as a **product + tech lead in one** turning a raw idea into a ready-to-implement feature in a single pass: `spec.md` + `tasks.md` (and `plan.md` only when the change is genuinely non-trivial).

This is the **fast lane** — it replaces the `spec → plan → tasks` sequence when the change is small/low-risk. When the intent is ambiguous, splits into multiple features, or needs serious design, STOP and fall back to `/spk.spec` (and `/spk.plan`).

Follow base rules in `AGENTS.md`.

## Phase Contract

Inputs: the user idea, `.speckeep/constitution.summary.md` (preferred) or `project.constitution_file`, minimal repo context.
Outputs: `<specs_dir>/<slug>/spec.md` + `<specs_dir>/<slug>/tasks.md` (plan.md optional).
Stop if: the idea is ambiguous, covers more than one feature, or would require inventing requirements/AC.

## Workflow

1. Run the pre-phase readiness check: `./.speckeep/scripts/check-ready.sh propose [slug]`.
2. Derive the slug (or accept `--slug`). Work from branch `feature/<slug>` — create/switch it if absent (same rule as `/spk.spec`).
3. Write `spec.md` from the spec template: `## Goal`, `## Requirements` (`RQ-*`), `## Acceptance Criteria` (`AC-*` with **Given / When / Then**), `## Assumptions`.
4. Decide the lane:
   - **express** (default): no `plan.md` — go straight to tasks.
   - **planned**: only when multiple realistic options, cross-boundary impact, or migration/rollout risk — then also write `plan.md` (with `DEC-*`) and `data-model.md` only if the data model really changes.
5. Write `tasks.md` from the tasks template, keeping it implement-self-contained: `## Surface Map`, `Touches:` on every task, `## Implementation Context`, `## Acceptance Coverage` (`AC-* -> T*`).
6. Do NOT implement code in this phase.

## Rules

- Minimum context: current slug only; narrow, targeted repo reads (no full-repo scans).
- Use `.speckeep/templates/spec.md` and `.speckeep/templates/tasks.md` as the skeletons; never search other slugs for shape.
- Size discipline: `spec.md` ≤ ~80 lines, `tasks.md` ≤ ~150 lines; overflow usually means the idea is too big for propose — stop and return to `/spk.spec`.
- Every `AC-*` maps to ≥ 1 task; every task has `Touches:` and an observable outcome.
- If the data model changes, create `data-model.md`; otherwise a `Data model: no change` line belongs in `tasks.md` → `Implementation Context` or `plan.md`.
- Constitution: AGENTS.md (`.speckeep/constitution.summary.md` preferred).
- Branch: `feature/<slug>` unless the user passed `--branch`.

## Self-Check (mandatory before finishing)

- [ ] Idea is single-feature and unambiguous; no invented requirements/AC
- [ ] `spec.md` has `RQ-*` + `AC-*` (Given/When/Then) + `## Assumptions`
- [ ] `tasks.md` has `## Surface Map`, `Touches:` everywhere, `## Implementation Context`, `## Acceptance Coverage`
- [ ] Every `AC-*` is covered by ≥ 1 task
- [ ] `plan.md` exists only when the change is non-trivial (express lane is the default)

If any check fails: fix and re-run. After **2 fix rounds** that still fail, stop and return to `/spk.spec <slug>` (or ask one targeted question) instead of forcing a propose.

## Output expectations

- Write `spec.md` + `tasks.md` (patch-in-place for `--amend`-style reopen).
- Summarize: slug, lane (express/planned), surfaces, AC coverage.
- End with standard end block (see AGENTS.md):
  ```
  Slug: <slug>
  Status: propose
  Artifacts: <paths>
  Blockers: <none | reason>
  Ready for: /spk.implement <slug>
  ```
- Final line: `Ready for: /spk.implement <slug>`