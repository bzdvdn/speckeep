# SpecKeep Handoff Prompt

You are generating a session handoff document for one feature.

## Goal

Produce a compact handoff document at `.speckeep/handoff/<slug>.md` that allows a new agent session to resume work on the feature without loss of context.

## Phase Contract

Inputs: current feature artifacts for `<slug>` (phase-appropriate subset).
Outputs: `.speckeep/handoff/<slug>.md` with current phase, completed work, open items, key decisions, and next command.
Stop if: slug is ambiguous or no feature artifacts exist for the slug.

## Load First

Always read these first if they exist:

- `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/spec.md`

## Load If Present

Read based on the inferred phase — only the artifacts that exist and contribute to the handoff:

- `.speckeep/specs/<slug>/inspect.md` — read the verdict line to populate `## Current Phase` and detect blockers
- `.speckeep/specs/<slug>/plan/plan.md` — read `DEC-*` entries to populate `## Key Decisions` when the feature is in plan phase or later
- `.speckeep/specs/<slug>/plan/tasks.md` — read task checkboxes to populate `## Open Work` and `## Completed` when the feature is in tasks/implement phase or later
- `.speckeep/specs/<slug>/plan/verify.md` — read the metadata block and verdict line to populate `## Current Phase` and detect verification concerns or blockers

## Do Not Read By Default

- `.speckeep/specs/<slug>/plan/data-model.md`
- `.speckeep/specs/<slug>/plan/contracts/`
- `.speckeep/specs/<slug>/plan/research.md`
- unrelated specs or plan packages
- implementation files
- script source files

## Stop Conditions

Stop and ask only if:

- the slug is ambiguous and cannot be derived from context or arguments
- no feature artifacts exist for the slug
- called without a slug and no active features exist at all

## All-Features Mode

When called without a slug (no slug in the user arguments):
- Run `.speckeep/scripts/list-specs.*` to enumerate active features; do not read its source.
- For each active feature, generate its handoff document at `.speckeep/handoff/<slug>.md` using the same rules as single-feature mode.
- Output a brief inline summary table: one row per feature with slug, phase, and ready for.
- Always overwrite existing handoff files — each is a current-state snapshot.

## Rules

- Infer the current phase from the set of artifacts present: no spec → pre-spec; spec exists, no inspect → spec; inspect exists, no plan → inspect; plan exists, no tasks → plan; tasks exist with open items → tasks or implement; tasks all closed and no `verify.md` → verify; `verify.md` exists with verdict `pass` or `concerns` → archive; `verify.md` exists with verdict `blocked` → return to the phase named in its `Return to:` line (do not suggest archive).
- When available, run `.speckeep/scripts/list-open-tasks.*` to enumerate incomplete tasks; rely on its output rather than reading the script source.
- Keep each handoff document compact: it must be loadable in a single cheap read at the start of a new session. Note that generating handoffs in all-features mode is inherently more expensive than single-slug mode — this is expected and acceptable; the "cheap read" principle applies to each produced document, not to the generation pass itself.
- Do not reproduce full artifact content in the handoff — reference file paths, not contents.
- Reference open items by their stable IDs (T*, AC-*, DEC-*, RQ-*) when available.
- Every section must add signal; omit sections that would be empty or redundant.
- Use the project's configured documentation language when writing to disk.
- Include a machine-readable metadata block at the top.

## Output Structure

Write `.speckeep/handoff/<slug>.md` using this structure:

- YAML metadata block: `report_type`, `slug`, `phase`, `docs_language`, `generated_at`
- `# Handoff: <slug>`
- `## Current Phase` — which phase the feature is in and the evidence for that inference
- `## Completed` — artifacts present and closed tasks (by ID when available)
- `## Open Work` — remaining tasks or missing required artifacts, with stable IDs where available
- `## Key Decisions` — decisions in `plan.md` that materially affect remaining work (DEC-* IDs)
- `## Open Questions` — blockers or unresolved questions that must be addressed before the next phase
- `## Artifacts` — exact project-relative paths to all relevant files for this slug
- `## Next Command` — the exact slash command and slug to resume work immediately

## Output Expectations

- Always overwrite `.speckeep/handoff/<slug>.md` if it already exists — the handoff is a current-state snapshot and the previous version is immediately stale.
- Write the file to `.speckeep/handoff/<slug>.md`.
- Also output a brief inline summary: current phase, number of open items, and ready for.
- End the conversation with the exact `Ready for` line so a new session can start without re-reading.
- If open questions block the next phase, state that in `## Ready for` instead of naming the phase command.

## Self-Check

- Is the current phase correctly inferred from which artifacts are present?
- Are open items referenced by stable IDs wherever possible?
- Is the document short enough to load cheaply at the start of a new session?
- Does `## Next Command` contain the exact slash command and slug to resume?
- Did I avoid reproducing full artifact content in the handoff?
