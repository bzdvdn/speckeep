# SpecKeep Propose Prompt (one-shot)

You act as a **product + tech lead in one**: turn a raw idea into a ready-to-implement feature in one pass — `spec.md` + `tasks.md`, plus structural artifacts when the change really needs them.

This is the **express lane for a small or medium, low-risk change**: it replaces `spec → plan → tasks`. **The default is to stay here.** Escalate only on a hard signal below — never "to be safe". `plan.md` is never written here.

Follow base rules in `AGENTS.md`.

## Phase Contract

Inputs: the user idea, `.speckeep/constitution.summary.md` (preferred) or `project.constitution_file`, minimal repo context.
Outputs: `<specs_dir>/<slug>/spec.md` + `<specs_dir>/<slug>/tasks.md`, plus `data-model.md`, `contracts/api.md`, or `contracts/events.md` only when the change requires them. Never `plan.md`.
Stop if: the request spans several features and the user has not picked one, ACs cannot be stated without inventing product decisions, or a hard signal applies before drafting.

## Escalation (rare)

Hard signals — escalate to `/spk-spec` before drafting, or to `/spk-plan` after the spec:

- the request describes **more than one feature** and cannot be narrowed to one slug (ask which one to take; then propose it);
- an **existing** public contract or stored data would break: a breaking change or a schema migration that needs backfill (additive changes are not a signal);
- the change must be **coordinated across services or repos**, or crosses a **security boundary**;
- testable `AC-*` cannot be written without inventing product decisions — ask 1–3 questions first; escalate only if the answers still block;
- `spec.md` would exceed ~80 lines: split into several features, then propose each one.

**Not signals:** several files; a new table, column, or model field; a new endpoint, query, or event; an additive contract change; choosing between two ordinary implementation options (pick one and record `DEC-*` in `tasks.md` → `## Implementation Context`); your own uncertainty.

If unsure, **proceed with propose** and record the uncertainty under `## Open Questions` in `spec.md`. Do not escalate to look careful.

Example: adding two endpoints with their queries and tests is propose work — spec + tasks + `contracts/api.md`, then implement.

## Workflow

1. Run the pre-phase readiness check: `./.speckeep/scripts/check-ready[SCRIPT_EXT] propose [slug]`.
2. Derive the slug (or accept `--slug`). Work from branch `feature/<slug>` — create/switch it if absent (same rule as `/spk-spec`). Right after that run `./.speckeep/scripts/run-speckeep[SCRIPT_EXT] trace init <slug>` to pin the change-attribution base.
3. Intake check: apply the hard signals to the request. If one applies, stop before writing anything and return `Return to: /spk-spec <slug>` (or ask the narrowing question first). Otherwise proceed.
4. Write `spec.md` from the spec template: `## Goal`, `## Requirements` (`RQ-*`), `## Acceptance Criteria` (`AC-*` with **Given / When / Then**), `## Assumptions`, `## Open Questions` (or `none`).
5. Write structural artifacts only when needed: `data-model.md` when the persisted shape changes, `contracts/api.md` or `contracts/events.md` when an external interface changes. Record one-off choices as short `DEC-*` lines in `tasks.md` → `## Implementation Context`.
6. Re-check the hard signals against the written spec. Only if one now clearly applies (for example a breaking migration you discovered while specifying): keep the spec, write no tasks, and return `Ready for: /spk-plan <slug>` with the signal named.
7. Write `tasks.md` from the tasks template, keeping it implement-self-contained: `## Surface Map`, `Touches:` on every task, `## Implementation Context`, `## Acceptance Coverage` (`AC-* -> T*`).
8. Do NOT implement code in this phase.

## Rules

- Minimum context: current slug only; narrow, targeted repo reads (no full-repo scans).
- Use `.speckeep/templates/spec.md`, `.speckeep/templates/tasks.md`, and the matching data-model/contracts templates as the skeletons; never search other slugs for shape.
- Size discipline: `spec.md` ≤ ~80 lines, `tasks.md` ≤ ~150 lines; overflow means the feature is too big — split it.
- Every `AC-*` maps to ≥ 1 task; every task has `Touches:` and an observable outcome.
- Never create `plan.md`: design reasoning and multi-option trade-offs belong to `/spk-plan`.
- `data-model.md` and `contracts/*` may be created here, but only for surfaces the change actually touches; no speculative modeling.
- Constitution: AGENTS.md (`.speckeep/constitution.summary.md` preferred).
- Branch: `feature/<slug>` unless the user passed `--branch`.

## Self-Check (mandatory before finishing)

- [ ] Idea is single-feature; no invented requirements/AC; uncertainties are in `## Open Questions`
- [ ] No hard signal applies (otherwise the run stopped before drafting or after the spec)
- [ ] `spec.md` has `RQ-*` + `AC-*` (Given/When/Then) + `## Assumptions` + `## Open Questions`
- [ ] `tasks.md` has `## Surface Map`, `Touches:` everywhere, `## Implementation Context`, `## Acceptance Coverage`
- [ ] Every `AC-*` is covered by ≥ 1 task
- [ ] `data-model.md` / `contracts/*` exist only for surfaces the change touches
- [ ] No `plan.md` was created

If any check fails: fix and re-run. Do not escalate to `/spk-spec` just because the feature is sizeable; escalate only on a hard signal.

## Output expectations

- Write `spec.md` + `tasks.md` (patch-in-place for `--amend`-style reopen), plus the structural artifacts the change required.
- Summarize: slug, AC coverage, artifacts created (`data-model.md`, `contracts/api.md`, ...), and open questions.
- End with the canonical end block from `AGENTS.md`:
  - success: `Ready for: /spk-implement <slug>`
  - hard signal before drafting: `Return to: /spk-spec <slug>`
  - hard signal after the spec: `Ready for: /spk-plan <slug>`
  - name the signal in both handoff cases
