# SpecKeep Inspect Prompt

You are inspecting one feature package for consistency and quality.

## Goal

Produce a focused inspection report for one feature without expanding scope.

## Path Resolution

Paths in this prompt use the default workspace layout. If `.speckeep/speckeep.yaml` overrides `paths.specs_dir` or `project.constitution_file`, always follow the configured paths instead of the defaults shown here.
Read `.speckeep/speckeep.yaml` at most once per session to resolve these paths; do not re-read it unless it changed or a path is ambiguous.

## Phase Contract

Inputs: `.speckeep/constitution.md`, `.speckeep/specs/<slug>/spec.md`; optionally `.speckeep/specs/<slug>/plan/plan.md`, `.speckeep/specs/<slug>/plan/tasks.md` when they exist.
Outputs: `.speckeep/specs/<slug>/inspect.md` with verdict `pass`, `concerns`, or `blocked`.
Stop if: slug ambiguous, spec missing, or report would require inventing product intent.

## Flags

`--delta`: incremental re-check mode — verify only the sections that changed since the last inspect report instead of running a full inspection.

When `--delta` is present in the user arguments:
- Read the existing `.speckeep/specs/<slug>/inspect.md` first to establish the baseline.
- Compare the current `spec.md` against the previous inspect report to identify changed sections (new or modified AC, scope changes, assumption changes).
- Re-check only the changed sections and their cross-artifact implications.
- Preserve findings from the previous report that are still valid; do not re-derive them.
- Update the verdict only if the delta changes it. If a previous `blocked` finding is resolved and no new errors appear, upgrade to `pass` or `concerns`.
- Update the `generated_at` timestamp and add a `delta_from: <previous_generated_at>` field to the metadata block.
- If the delta touches the spec's `## Goal`, `## Scope`, or more than half of the `AC-*` entries, treat the change as broad and fall back to a full inspection. Note: "Delta mode fell back to full inspection due to broad changes."
- If the delta changes any `AC-*` entry, re-generate `summary.md` after updating `inspect.md`. If no ACs changed, leave `summary.md` unchanged.

## Load First

Always read these first:

- `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/spec.md`

## Load If Present

Read these only when they exist and the inspection requires cross-artifact consistency checks (spec↔plan alignment, acceptance↔task coverage):

- `.speckeep/specs/<slug>/plan/plan.md` — read when checking goal alignment, scope expansion, or acceptance coverage at plan level
- `.speckeep/specs/<slug>/plan/tasks.md` — read when verifying that every `AC-*` is covered by at least one task

## Do Not Read By Default

- `.speckeep/specs/<slug>/plan/data-model.md`
- `.speckeep/specs/<slug>/plan/contracts/`
- `.speckeep/specs/<slug>/plan/research.md`
- broad repository history
- implementation files unless a finding names a specific file and the claim cannot be confirmed from spec/plan/tasks alone

## Stop Conditions

Stop and ask a minimal follow-up question only if:

- the target slug is ambiguous
- the spec is missing entirely
- the inspection would otherwise invent missing product intent

## Rules

- Check constitutional consistency first.
- If `/.speckeep/scripts/check-inspect-ready.*` is available, prefer it as the cheap first pass before deepening into artifacts.
- Important: the readiness wrapper runs with the slug as the first argument. Example: `bash ./.speckeep/scripts/check-inspect-ready.sh <slug>` (or PowerShell: `.\.speckeep\scripts\check-inspect-ready.ps1 <slug>`).
- Use `/.speckeep/scripts/inspect-spec.*` only as a fallback when the phase-readiness wrapper is unavailable.
- If neither script is available, proceed with full manual inspection using `constitution.md` and `spec.md` directly — do not stop or wait for the scripts.
- Prefer helper script output over reading helper script source.
- Treat helper script output as the primary structural evidence layer for inspect. When the scripts report concrete `ERROR` / `WARN` findings, use those findings as the starting point for your report instead of re-deriving the same points from scratch.
- If helper output exposes finding categories such as structure, traceability, ambiguity, consistency, or readiness, preserve that signal in your reasoning. Expand on it only when extra context is genuinely needed.
- Do not ignore a concrete helper finding just because your broader intuition is optimistic. Resolve or explicitly explain it.
- Use your own reasoning mainly for what the cheap checks cannot prove directly: constitutional conflicts, invented product intent, unjustified scope expansion, contradictory assumptions, or subtle spec↔plan drift.
- Do not read `/.speckeep/scripts/*` by default unless you are debugging the script, working on SpecKeep itself, or the user explicitly asks to inspect script logic.
- Inspect spec completeness and clarity.
- Verify `constitution <-> spec`: the spec must not conflict with explicit constitutional constraints, workflow rules, or language policy.
- Treat technology names, framework choices, library lists, or version pins in the spec as a `Warning` unless they clearly represent a user requirement, repository constraint, or external compatibility contract.
- Every acceptance criterion in the spec MUST have an explicit Given/When/Then format. The `Given`, `When`, and `Then` markers remain canonical regardless of the documentation language. Missing G/W/T is an `Error`, not a `Suggestion`.
- Any `[NEEDS CLARIFICATION: ...]` marker remaining in the spec is an `Error`. These must be resolved before planning can begin.
- If `## Assumptions` is missing, flag it as a `Warning`. If present, check each assumption for plausibility against the constitution and known repository state — an assumption that contradicts repository reality is an `Error`.
- If `## Success Criteria` is present, each `SC-*` must have a measurable metric and measurement method. Vague SC entries (e.g., "system should be fast") are a `Warning`.
- If `tasks.md` exists, verify that every acceptance criterion from the spec is covered by at least one task. An uncovered criterion is an `Error`.
- If `tasks.md` exists and has task IDs but is missing `## Surface Map`, flag it as a `Warning` — the implement agent needs this section as a batch-read manifest.
- If `tasks.md` exists and any task line with a task ID is missing a `Touches:` field, flag it as a `Warning` — tasks without `Touches:` force the implement agent into exploratory reads.
- If `tasks.md` uses task IDs such as `T1.1`, prefer traceability statements that reference those task IDs directly.
- Prefer the cheapest inspection scope first: `constitution.md` and `spec.md`, then `plan.md`, then `tasks.md`, and only then deeper plan artifacts when a concrete claim requires them.
- If no `plan.md` exists, do not widen the inspection into optional plan artifacts or implementation code.
- If plan artifacts exist, check alignment between spec, plan, data model, contracts, and tasks.
- When `plan.md` exists, check `spec <-> plan` consistency before reading deeper plan artifacts.
- Verify `spec <-> plan`: the plan should preserve the feature goal, reflect major acceptance-critical behavior, and avoid unjustified new workstreams.
- If `tasks.md` exists, verify `plan <-> tasks`: task phases and task IDs should reflect the plan intent without obvious missing work for acceptance-critical behavior.
- Treat `spec.md` and `plan.md` as the required inputs for cheap plan consistency checks.
- Only read `data-model.md` or `contracts/` when `plan.md` explicitly depends on them or when they are required to confirm a concrete consistency claim.
- Check `Goal Alignment`: the plan must not change the core feature goal expressed in the spec.
- Check `Scope Expansion`: the plan must not introduce major new workstreams, components, or integration surfaces that are outside the spec.
- Check `Acceptance Coverage at Plan Level`: major acceptance-critical behavior from the spec should be reflected in the plan intent, even before tasks exist.
- Check `Constitution Consistency`: the plan must not violate constitutional rules or architectural constraints.
- If `plan.md` exists and is missing `## Constitution Compliance`, flag it as a `Warning` — this section makes constitution adherence explicit and reviewable.
- Check `Artifact Justification`: if the plan introduces `data-model.md` or `contracts/`, the need for those artifacts should be justified by the spec.
- Do not turn this into a broad design review. Prefer catching obvious drift over scoring architecture quality.
- Keep the inspection report in the project's configured documentation language when writing it to disk.
- Prefer concrete findings over generic advice.
- Prefer this reporting order:
  - 1. structural findings from helper output
  - 2. cross-artifact consistency findings confirmed from the loaded artifacts
  - 3. narrow judgment calls that require agent reasoning
- Default to a compact report in conversation output: always include `Verdict`, include `Errors`, `Warnings`, and `Next Step` when non-empty, and include `Questions`, `Suggestions`, or `Traceability` only when they add real signal.
- Produce the full sectioned report only when the user explicitly asks for a full report or when the report is being persisted to a file.
- When writing the report to disk, include a machine-readable metadata block at the top with `report_type`, `slug`, `status`, `docs_language`, and `generated_at`.
- Use this report structure:
  - YAML-style metadata block at the top
  - `# Inspect Report: <slug>`
  - `## Scope`
  - `## Verdict`
  - `## Errors`
  - `## Warnings`
  - `## Questions`
  - `## Suggestions`
  - `## Traceability`
  - `## Next Step`
- The `## Verdict` section MUST use one of: `pass`, `concerns`, `blocked`.
- Use `pass` when no errors are present and only minor or no warnings remain.
- Use `concerns` when the feature can still move forward, but warnings, traceability gaps, or open questions should be resolved soon.
- Use `blocked` when constitutional conflicts, missing spec intent, missing Given/When/Then acceptance criteria, uncovered acceptance criteria, or major `spec <-> plan` contradictions prevent the next workflow step from proceeding safely.
- `## Traceability` should summarize how acceptance criteria map to tasks when `tasks.md` exists.
- Prefer traceability statements that reference stable acceptance IDs and task IDs such as `AC-001 -> T1.1, T2.1`.
- `## Next Step` should say whether it is safe to continue to `plan`, `tasks`, or whether refinement is required first.
- For `pass`, name the exact next slash command.
- For `concerns`, say whether the workflow may continue; if it may, include the exact next slash command.
- For `blocked`, do not suggest the next phase command; state which refinement is required first.
- Avoid duplicating the same issue in multiple sections. If helper output already established the concrete problem, keep your wording concise and move on to its consequence or required refinement.

## Spec Summary Artifact

After writing the inspect report, also write `.speckeep/specs/<slug>/summary.md`.

The summary MUST contain only:

- A YAML frontmatter block with `slug` and `generated_at`
- `## Goal` — one sentence
- `## Acceptance Criteria` — a table: `ID | Summary | Proof Signal`; summary ≤ 8 words; proof signal = the observable check from the `Then` clause
- `## Out of Scope` — 3-5 bullets

Keep the summary under 25 lines. It is loaded by `tasks`, `implement`, and `verify` instead of the full spec to reduce context overhead. The summary is not a substitute for the full spec in phases that require complete acceptance-criterion inspection (inspect, plan).

## Output expectations

- Persist to `.speckeep/specs/<slug>/inspect.md` and write `.speckeep/specs/<slug>/summary.md`
- Summarize verdict in the conversation; prefer compact report with only non-empty sections.
- End with a summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`
- When ready: `Ready for: /speckeep.plan <slug>` (or `/speckeep.tasks <slug>` when plan already exists; after archive you MAY mention `/speckeep.recap` as an optional summary but don’t advertise it as required)

## Self-Check

- Did I check every AC for Given/When/Then format?
- Is the verdict (`pass`, `concerns`, `blocked`) supported by concrete findings, not general impressions?
- If `tasks.md` exists, did I verify every AC is covered by at least one task?
