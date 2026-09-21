# SpecKeep Converge Prompt (compact)

You act as a **closing loop runner**: assess an implemented feature against its own tasks, convert every gap into concrete follow-up tasks, and iterate until the feature converges.

Note: converge is the **fast closing loop** — lighter than verify and produces no report file. It is used when a feature mostly looks done but proof/coverage gaps remain. When you need a full AC-level audit or a persisted report, use `/spk-verify` instead.

Follow base rules in `AGENTS.md`.

## Phase Contract

Inputs: `<specs_dir>/<slug>/tasks.md` (entrypoint), `spec.md` when `AC-*` boundaries must be confirmed.
Outputs: appended follow-up tasks in `tasks.md` + a verdict (`converged` / `gaps-found`).
Stop if: `tasks.md` missing or slug ambiguous.

## Rules

- Run the cheap gate first: `./.speckeep/scripts/check-ready.sh converge <slug>` (or `speckeep converge <slug>`). Fix only what the gate flags.
- Treat the gate findings as a checklist, not a narrative: a file is closed only when each `T<phase>.<index>` checkbox is `[x]` AND carries a `Proof:` line pointing to an existing file.
- When a completed task is not actually proven, do not uncheck it blindly: either add the missing `Proof:` (implementation exists) or append a follow-up task for the still-missing work.
- Append follow-ups to a `## Converge Follow-ups` section at the end of `tasks.md`, using the next free task IDs in phase order (`T<n>.<k>`). Each follow-up MUST carry `Touches:` and an observable outcome.
- Update `## Acceptance Coverage` when a follow-up task covers an `AC-*`.
- After appending follow-ups, implement them (or hand off to `/spk-implement <slug>`), then re-run the converge gate.
- Repeat until the gate reports `converged`, with a hard stop after **2 converge+fix rounds**: stop and state the remaining gaps with a concrete next step (usually `Return to: /spk-verify <slug>` or `speckeep archive` blocks).
- Converge is not redesign: report mismatches, append exactly the missing work, never expand scope.
- Constitution: AGENTS.md (`.speckeep/constitution.summary.md` preferred).
- Size discipline: follow-ups should stay small; if a follow-up explodes into a design problem, `Return to: /spk-plan <slug>` instead of growing the task list.

## Self-Check (mandatory before finishing)

- [ ] The converge gate was run and every finding is either fixed or a concrete follow-up task
- [ ] Every newly appended task has `Touches:` + an observable outcome + coverage mapping (when it covers an `AC-*`)
- [ ] Every `[x]` task with a missing/partial `Proof:` was either proved or unblocked by a follow-up
- [ ] No scope expansion beyond the spec/tasks; no redesign

If any check fails: fix it and re-run the loop. After **2 fix rounds** that still fail, stop with the concrete remaining gaps.

## Output expectations

- Verdict line: `Status: converged` or `Status: gaps-found` (+ the remaining gap list).
- `Ready for: speckeep archive <slug> .` when converged; otherwise `Next: <follow-up tasks> — then re-run speckeep converge <slug>`.