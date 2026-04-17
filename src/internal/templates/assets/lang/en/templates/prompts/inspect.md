# SpecKeep Inspect Prompt

You are inspecting one feature package for consistency and quality.

Follow base rules in `AGENTS.md` (paths, git, load discipline, readiness scripts, language, phase discipline).

## Goal

Produce a focused inspection report for one feature without expanding scope.

## Phase Contract

Inputs: `.speckeep/constitution.md`, `.speckeep/specs/<slug>/spec.md`; optionally `plan/plan.md`, `plan/tasks.md` when present.
Outputs: `.speckeep/specs/<slug>/inspect.md` with verdict `pass`, `concerns`, or `blocked`; plus `.speckeep/specs/<slug>/summary.md`.
Stop if: slug ambiguous, spec missing, or report would require inventing product intent.

## Flags

`--delta`: incremental re-check — re-verify only sections changed since the last inspect report.

- Read existing `inspect.md` as baseline; compare current `spec.md` to identify changed sections (AC, scope, assumptions).
- Re-check changed sections and their cross-artifact implications. Preserve still-valid prior findings; do not re-derive.
- Update verdict only if the delta changes it. Resolved `blocked` with no new errors → upgrade.
- Update `generated_at`; add `delta_from: <previous_generated_at>` to the metadata block.
- If delta touches `## Goal`, `## Scope`, or >50% of `AC-*`, fall back to full inspection and note: "Delta mode fell back to full inspection due to broad changes."
- If any `AC-*` changed, regenerate `summary.md`; otherwise leave it.

## Load First

Always read these first:

- `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/spec.md`

## Load If Present

Read these only when they exist and the inspection requires cross-artifact consistency checks (spec↔plan alignment, acceptance↔task coverage):

- `plan/plan.md` — for goal alignment, scope, or plan-level acceptance coverage
- `plan/tasks.md` — to verify every `AC-*` is covered by ≥1 task

## Do Not Read By Default

- `plan/data-model.md`, `plan/contracts/`, `plan/research.md`
- broad repo history
- implementation files — unless a finding names a specific file whose claim cannot be confirmed from spec/plan/tasks

## Stop Conditions

Ask one minimal question only if: slug ambiguous, spec missing, or inspection would invent missing product intent.

## Rules

- Check constitutional consistency first.
- If `/.speckeep/scripts/check-inspect-ready.*` exists, run it as the cheap first pass (slug is first arg: `bash ./.speckeep/scripts/check-inspect-ready.sh <slug>` or PowerShell `.\.speckeep\scripts\check-inspect-ready.ps1 <slug>`). Fallback: `/.speckeep/scripts/inspect-spec.*`. Neither available → proceed with manual inspection from `constitution.md` and `spec.md`.
- Prefer helper output over helper source. Treat helper `ERROR`/`WARN` findings as the primary structural layer; do not re-derive them. Preserve finding categories (structure, traceability, ambiguity, consistency, readiness) when surfaced.
- Do not ignore a concrete helper finding because intuition is optimistic — resolve or explain it.
- Use agent reasoning for what cheap checks cannot prove: constitutional conflicts, invented product intent, unjustified scope expansion, contradictory assumptions, subtle spec↔plan drift.

### Spec checks

- Inspect spec completeness and clarity.
- Verify `constitution <-> spec`: spec must not conflict with constitutional constraints, workflow rules, or language policy.
- Treat technology names, framework choices, library lists, or version pins in the spec as a `Warning` unless they clearly represent a user requirement, repository constraint, or external compatibility contract.
- Every AC MUST use Given/When/Then (markers canonical across languages). Missing G/W/T is an `Error`.
- Any remaining `[NEEDS CLARIFICATION: ...]` is an `Error` — must be resolved before planning.
- Missing `## Assumptions` → `Warning`. An assumption contradicting repo reality → `Error`.
- `## Success Criteria` present → each `SC-*` must have a measurable metric + method. Vague SC → `Warning`.

### Cross-artifact checks

- Prefer the cheapest inspection scope first: constitution + spec, then plan, then tasks, then deeper plan artifacts only when a concrete claim requires them.
- No `plan.md` → do not widen into optional plan artifacts or code.
- When `plan.md` exists, check `spec <-> plan` before reading deeper plan artifacts. Read `data-model.md` or `contracts/` only when `plan.md` depends on them or a concrete consistency claim requires it.
- Verify `spec <-> plan`: plan preserves the feature goal, reflects major acceptance-critical behavior, avoids unjustified new workstreams. Check:
  - `Goal Alignment` — core feature goal unchanged
  - `Scope Expansion` — no major new workstreams/components/surfaces outside the spec
  - `Acceptance Coverage at Plan Level` — major acceptance-critical behavior reflected in plan intent
  - `Constitution Consistency` — plan obeys constitutional rules
  - `Artifact Justification` — `data-model.md`/`contracts/` justified by spec
- `plan.md` missing `## Constitution Compliance` → `Warning`.
- `plan.md` exists but `data-model.md` is missing → `Error`. The plan phase MUST either define model changes or persist an explicit no-change stub.
- `data-model.md` exists but only implies "no changes" vaguely → `Warning`; prefer an explicit status/reason/revisit-triggers stub so downstream phases do not guess.
- If `tasks.md` exists, verify `plan <-> tasks`: task phases and IDs reflect plan intent without obvious missing work for acceptance-critical behavior.
- If `tasks.md` exists, verify every AC is covered by ≥1 task; uncovered AC → `Error`. Missing `## Surface Map` → `Warning`. Task-ID line missing `Touches:` → `Warning`. Prefer traceability statements referencing task IDs like `T1.1` directly.
- Do not turn this into a broad design review.

## Report

- Keep the report in the configured documentation language.
- Prefer concrete findings over generic advice. Reporting order: (1) helper-output structural findings, (2) cross-artifact consistency findings from loaded artifacts, (3) narrow judgment calls.
- Default to a compact report in conversation output: always include `Verdict`; include `Errors`/`Warnings`/`Next Step` when non-empty; include `Questions`/`Suggestions`/`Traceability` only when they add signal.
- Produce the full sectioned report only when the user explicitly asks for a full report or when persisting to disk.
- When writing to disk, include a machine-readable metadata block at the top with `report_type`, `slug`, `status`, `docs_language`, `generated_at`.
- Structure: YAML metadata → `# Inspect Report: <slug>` → `## Scope` → `## Verdict` → `## Errors` → `## Warnings` → `## Questions` → `## Suggestions` → `## Traceability` → `## Next Step`.
- The `## Verdict` section MUST use one of: `pass`, `concerns`, `blocked`.
  - `pass`: no errors, only minor or no warnings.
  - `concerns`: can move forward, but warnings / traceability gaps / open questions should be resolved soon.
  - `blocked`: constitutional conflicts, missing spec intent, missing Given/When/Then, uncovered AC, or major `spec <-> plan` contradictions block safe progression.
- `## Traceability` summarizes AC → tasks when `tasks.md` exists. Prefer stable IDs like `AC-001 -> T1.1, T2.1`.
- `## Next Step`:
  - For `pass`, name the exact next slash command.
  - For `concerns`, say whether workflow may continue; if yes, name the exact next slash command.
  - For `blocked`, do not suggest the next phase command; state which refinement is required first.
- Avoid duplicating findings across sections.

## Spec Summary Artifact

After the inspect report, also write `.speckeep/specs/<slug>/summary.md`:

- YAML frontmatter: `slug`, `generated_at`
- `## Goal` — one sentence
- `## Acceptance Criteria` — table `ID | Summary | Proof Signal`; summary ≤ 8 words; proof signal = observable check from `Then`
- `## Out of Scope` — 3–5 bullets

Keep it under 25 lines. It is loaded by `tasks`/`implement`/`verify` to reduce context. It does not substitute the full spec in phases that require complete AC inspection (`inspect`, `plan`).

## Output

- Persist `inspect.md` and write `summary.md`.
- Summarize verdict in the conversation (compact report, non-empty sections only).
- End with a summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`.
- When ready: `Ready for: /speckeep.plan <slug>` (or `/speckeep.tasks <slug>` if plan already exists; after archive, `/speckeep.recap` is optional — don't advertise as required).
