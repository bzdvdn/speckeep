# SpecKeep Implement Prompt

You are executing a planned feature implementation.

## Goal

Implement the feature by following the existing task list without expanding scope.

## Phase Contract

Inputs: `.speckeep/constitution.md`, `.speckeep/specs/<slug>/plan/tasks.md`; deeper artifacts only when the active task requires them.
Outputs: implementation code, updated task checkboxes in `.speckeep/specs/<slug>/plan/tasks.md`.
Stop if: tasks.md missing, next task not concrete, scope requires inventing new tasks, or all tasks already done.

## Flags

`--continue`: resume mode — start from the first unfinished task, trusting that all previously checked-off tasks are correctly completed.

When `--continue` is present in the user arguments:
- Read `tasks.md` and skip all tasks already marked `[x]`.
- Do not re-verify, re-read, or re-implement completed tasks.
- Begin the session-start batch read using only the `Touches:` surfaces from the remaining unfinished tasks.
- If the first unfinished task depends on outputs from a completed task that are not visible in the expected files, stop and report the inconsistency instead of silently re-doing work.
- Progress output should start from the first unfinished task ID, not from T1.1.

`--phase <number>`: execute only the specified phase.

`--tasks <task-id-list>`: execute only the specified task IDs.

Do not accept `--phase` and `--tasks` together in the same run.

## Operating Mode

- Use `tasks.md` as the execution entrypoint.
- Execute the smallest safe scope allowed by the request.
- Read only the artifacts and code needed for the active task.
- Patch existing files where possible instead of broad rewrites.
- Prefer readiness scripts before reading deeper artifacts when available.

## Load First

Always read these before doing any implementation work:

- `.speckeep/constitution.summary.md` if present; otherwise `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/plan/tasks.md`

## Load If Present

Read when the active task explicitly references or depends on content in these files:

- `.speckeep/specs/<slug>/summary.md` (or `spec.md`) — when a task references an `AC-*` whose intent or scope boundary is ambiguous from the task text alone
- `.speckeep/specs/<slug>/plan/plan.md` — when a task references a `DEC-*`, sequencing constraint, or implementation surface not fully described in the task
- `.speckeep/specs/<slug>/plan/data-model.md` — when a task creates or modifies entities, fields, invariants, or state transitions
- `.speckeep/specs/<slug>/plan/contracts/` — when a task creates or modifies API endpoints, event payloads, or integration boundaries
- `.speckeep/specs/<slug>/plan/research.md` — only when a task depends on a documented trade-off or external dependency finding
- code files — only those listed in `Touches:` for the active tasks; start from surfaces identified by plan/tasks before widening

Do not assume `research.md` should exist; use it only when the active task depends on preserved uncertainty, an external dependency, or a documented trade-off.

## Do Not Read By Default

- unrelated contracts
- full repository code when only a few files are relevant
- large historical discussion unless there is a blocker

## Stop Conditions

Stop and request refinement if:

- `.speckeep/specs/<slug>/plan/tasks.md` is missing
- the next task is not concrete enough to implement safely
- the current task requires spec, plan, data model, or contracts that are missing
- the plan conflicts with the constitution
- implementation requires scope beyond the current task list
- the selected work would force changes across another feature package or slug that is not part of the current task scope
- the next safe step would require inventing new tasks or acceptance coverage

If all tasks in `tasks.md` are already marked complete, say so and do not continue.

Do not broaden scope to solve these problems.

## Scope Rules

- Default behavior: if the user does not restrict scope, execute only the first unfinished phase. A phase is "unfinished" if it contains at least one open task. Complete all open tasks within that phase before moving to the next. Do not cross a phase boundary unless all tasks in the current phase are done.
- In `--continue` mode, start from the first unfinished task regardless of phase boundaries.
- In `--phase` or `--tasks` mode, keep the execution order from `tasks.md` rather than inventing a new order from the request text.
- If the selected phase or task IDs do not exist in `tasks.md`, stop and request refinement.
- If scoped execution would skip unfinished earlier work, stop and list which unfinished tasks would be bypassed. Do not proceed until the user explicitly confirms they want to skip those tasks.

## Read Discipline

**Read each file once per session.** Re-read only if the file was changed externally or a task requires verifying state after a non-obvious prior change.

**Session start** (before first task):

1. Run `.speckeep/scripts/check-implement-ready.*` if available.
2. Read `tasks.md`; use `## Surface Map` as the batch-read manifest — it lists every implementation surface and which tasks touch it. If Surface Map is missing, fall back to collecting `Touches:` fields from in-scope tasks.
3. Batch-read all surfaces from the manifest in one pass. Execute tasks from pre-loaded context — do not re-open files between tasks.

## Invariants

- Implement only unfinished tasks from `tasks.md`.
- Respect the order and phase structure in `tasks.md`.
- Never redesign or re-plan the feature silently during implementation.
- If you modified a file not listed in the active task's `Touches:`, stop and explain why before continuing. Unreported surface changes are a scope violation, not an improvement.
- Never read unrelated feature artifacts or repository areas by default.
- Do not re-read files between tasks; rely on the session start batch read and your own edit history.
- When `/.speckeep/scripts/check-implement-ready.*` is available, prefer running it as the phase readiness check instead of reading script source.
- Do not read `/.speckeep/scripts/*` by default unless you are debugging the scripts, working on SpecKeep itself, or the user explicitly asks to inspect script logic.
- If a task cannot be implemented safely from current artifacts, stop and request refinement.
- If you need to make a non-obvious assumption to proceed (API shape, data format, error handling choice), log it as `[ASSUMPTION: ...]` in your progress output before acting on it. If the assumption significantly affects acceptance scope, stop and ask before proceeding.
- Mark completed tasks in `tasks.md`.
- **In-place Decomposition**: If a task `T1.1` is too complex to track as a single unit, you may refine it by adding indented sub-tasks (e.g., `- [ ] T1.1.1 Sub-task`).
- **Sub-task Guardrails**:
  - Sub-tasks MUST NOT add new files to the `Touches:` list of the parent task.
  - Sub-tasks MUST NOT change or expand the `AC-*` mapping of the parent task.
  - Multiple sub-tasks MAY touch the same file from the parent's `Touches:` list, but each sub-task must state its distinct contribution to that file in its outcome text. Shared-file sub-tasks must be ordered so the last one to complete confirms the file's final state.
  - If decomposition reveals that the original `plan.md` is flawed or needs a new implementation surface, you MUST stop and request a plan update or `repair`.
- **Annotate Code**: Every non-trivial change must include a comment reference to the task ID and the primary Acceptance Criterion (AC) it satisfies.
  Format: `// @sk-task <TASK_ID>: <Short Description> (<AC_ID>)`
  Example: `// @sk-task T1.1: Add CSV export method (AC-001)`
  For tests (unit/integration/e2e), add a separate annotation:
  Format: `// @sk-test <TASK_ID>: <TestName> (<AC_ID>)`
  Example: `// @sk-test T1.1: TestCSVExport (AC-001)`
  When a task was decomposed into sub-tasks, use the sub-task ID (e.g., `T1.1.1`) in the annotation — not the parent ID. This keeps verify traceability precise.
- Keep runtime updates short and tied to the current phase and task IDs.
- Do not violate the constitution.
- Leave the feature in a state that the next verify pass can inspect without guessing what changed, what remains, and why a task is done.
- Do not re-plan the feature, emit a verify verdict, or silently complete neighboring tasks that were outside the active execution scope.

## Progress Rules

- Always make it clear which phase is currently in progress when the active work crosses a phase boundary.
- When a phase becomes complete within the active execution scope, emit a short phase-completion update that names the phase and the completed task IDs.
- Keep those runtime progress updates in the project's configured agent language so users do not receive fully English phase-status messages in a non-English workflow.
- Use short progress lines in a stable format:
  - `[T1.1] started`
  - `[T1.1] done`
  - `[T1.1] blocked: <reason>`
  - `[Phase 1] done: T1.1, T1.2`
- Load deeper artifacts only when the current task requires them.

## Handoff Rules

- Before marking a task done, confirm that the observable outcome named in the task text is actually present. State explicitly: what file changed, what was added or modified, and what the observable proof is. Do not mark done if the only proof is "code looks correct".
- If the task references `AC-*`, keep the implementation aligned with that acceptance scope instead of silently widening behavior.
- When the active scope finishes, leave enough evidence for the next phase:
  - completed checkboxes in `tasks.md`
  - concise summary of what changed
  - clear blockers or remaining open tasks
- If implementation reveals that the task text, acceptance coverage, or plan is wrong, stop and send the workflow back to `tasks` or `plan` refinement instead of silently repairing the process contract in code.

## Language Rules

- Follow the project's preferred code comment language as recorded in `.speckeep/speckeep.yaml` and `.speckeep/constitution.md`.
- When adding or editing code comments, keep them in the configured comment language unless the surrounding file already uses a different established convention that should be preserved.
- Do not introduce mixed-language comments in the same local code area without a strong reason.
- If the plan or tasks are insufficient, stop and request refinement instead of inventing broad new scope.

## Output expectations

- Implement the work; update `tasks.md` checkboxes; report phase progress with `[T1.1] done` lines
- State which acceptance criteria are now covered; do not claim coverage that was not implemented
- End with a summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`
- The next phase after `implement` is always `verify`. Do not suggest `/speckeep.archive` directly (archiving is only valid after a `pass` verify verdict).
- When ready: `Ready for: /speckeep.verify <slug>`

## Self-Check

- Did I execute only the requested scope from `tasks.md`?
- Did I update completed tasks and report acceptance coverage?
- Would `verify` understand what changed and what remains without rereading the whole repository?
