# SpecKeep Implement Prompt

You are executing a planned feature implementation.

Follow base rules in `AGENTS.md` (paths, git, load discipline, readiness scripts, language, phase discipline).

## Goal

Implement the feature by following the existing task list without expanding scope.

## Phase Contract

Inputs: `.speckeep/constitution.md`, `.speckeep/specs/<slug>/plan/tasks.md`; deeper artifacts only when the active task requires them.
Outputs: implementation code, updated task checkboxes in `tasks.md`.
Stop if: `tasks.md` missing, next task not concrete, scope requires inventing new tasks, or all tasks already done.

## Flags

`--continue`: resume mode — start from the first unfinished task, trusting prior `[x]` tasks are correct.

- Read `tasks.md`; skip `[x]` tasks. Do not re-verify, re-read, or re-implement completed tasks.
- Session-start batch read uses only `Touches:` surfaces from remaining unfinished tasks.
- If an unfinished task depends on outputs from a completed task that are missing from the expected files, stop and report the inconsistency; do not silently redo work.
- Progress output starts from the first unfinished task ID, not `T1.1`.

`--phase <number>`: execute only the specified phase.

`--tasks <task-id-list>`: execute only the specified task IDs.

Do not accept `--phase` and `--tasks` together in the same run.

## Operating Mode

- Use `tasks.md` as the execution entrypoint.
- Execute the smallest safe scope allowed by the request.
- Read only artifacts and code needed for the active task.
- Patch existing files instead of broad rewrites.
- Prefer readiness scripts before reading deeper artifacts.

## Load First

Always read before any implementation work:

- `.speckeep/constitution.summary.md` if present; otherwise `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/plan/tasks.md`

## Load If Present

Read only when the active task references or depends on these:

- `summary.md` (or `spec.md`) — when a task references an `AC-*` ambiguous from task text alone
- `plan/plan.md` — when a task references a `DEC-*`, sequencing constraint, or surface not described in the task
- `plan/data-model.md` — when a task creates/modifies entities, fields, invariants, or state transitions
- `plan/contracts/` — when a task creates/modifies API endpoints, event payloads, or integration boundaries
- `plan/research.md` — only when a task depends on a documented trade-off or external dependency finding
- code files — only those listed in `Touches:` for active tasks; start from plan/tasks surfaces before widening

Do not assume `research.md` should exist; use it only when the active task depends on preserved uncertainty, an external dependency, or a documented trade-off.

## Do Not Read By Default

- unrelated contracts
- full repo code when only a few files are relevant
- large historical discussion unless there is a blocker

## Stop Conditions

Stop and request refinement if:

- `tasks.md` is missing
- the next task is not concrete enough to implement safely
- the current task requires spec, plan, data-model, or contracts that are missing
- the plan conflicts with the constitution
- implementation requires scope beyond the current task list
- the selected work would force changes across another feature package or slug that is not part of the current task scope
- the next safe step would require inventing new tasks or acceptance coverage

If all tasks are already `[x]`, say so and stop. Do not broaden scope to solve these problems.

## Scope Rules

- Default behavior: if the user does not restrict scope, execute only the first unfinished phase. A phase is "unfinished" if it has ≥1 open task. Complete all open tasks in that phase before moving on. Do not cross a phase boundary unless all tasks in the current phase are done.
- `--continue` mode: start from the first unfinished task regardless of phase boundaries.
- `--phase` / `--tasks` mode: keep the execution order from `tasks.md`; do not invent a new order.
- If the selected phase or task IDs do not exist in `tasks.md`, stop and request refinement.
- If scoped execution would skip unfinished earlier work, stop, list the bypassed tasks, and wait for explicit user confirmation before proceeding.

## Read Discipline

**Read each file once per session.** Re-read only if the file was changed externally or a task requires verifying state after a non-obvious prior change.

**Session start** (before first task):

1. Run `.speckeep/scripts/check-implement-ready.*` if available (slug as first arg: `bash ./.speckeep/scripts/check-implement-ready.sh <slug>`).
2. Read `tasks.md`; use `## Surface Map` as the batch-read manifest. If missing, fall back to collecting `Touches:` from in-scope tasks.
3. Batch-read all surfaces in one pass. Execute tasks from pre-loaded context — do not re-open files between tasks.

## Invariants

- Implement only unfinished tasks from `tasks.md`. Respect order and phase structure.
- Never redesign or re-plan during implementation.
- Modifying a file not in the active task's `Touches:` is a scope violation — stop and explain before continuing.
- Never read unrelated feature artifacts or repo areas by default. Do not re-read files between tasks.
- If `/.speckeep/scripts/check-implement-ready.*` exists, prefer running it (slug as first arg: `bash ./.speckeep/scripts/check-implement-ready.sh <slug>` or PowerShell `.\.speckeep\scripts\check-implement-ready.ps1 <slug>`) over reading script source.
- If a task cannot be implemented safely from current artifacts, stop and request refinement.
- Non-obvious assumption to proceed (API shape, data format, error handling) → log `[ASSUMPTION: ...]` in progress output before acting. If the assumption materially affects acceptance scope, stop and ask.
- Mark completed tasks in `tasks.md`.
- **In-place Decomposition**: task `T1.1` too complex → add indented sub-tasks (e.g., `- [ ] T1.1.1 Sub-task`).
- **Sub-task Guardrails**:
  - MUST NOT add files to the parent's `Touches:`.
  - MUST NOT change the parent's `AC-*` mapping.
  - Multiple sub-tasks MAY touch the same file from `Touches:`; each must state its distinct contribution, and the last-to-complete confirms final state.
  - If decomposition reveals that `plan.md` is flawed or needs a new surface, stop and request plan update or `repair`.
- **Annotate Code**: every non-trivial change includes:
  - `// @sk-task <TASK_ID>: <Short Description> (<AC_ID>)` — example: `// @sk-task T1.1: Add CSV export method (AC-001)`
  - For tests: `// @sk-test <TASK_ID>: <TestName> (<AC_ID>)` — example: `// @sk-test T1.1: TestCSVExport (AC-001)`
  - When decomposed, use the sub-task ID (e.g., `T1.1.1`), not the parent ID.
- Keep runtime updates short and tied to the current phase and task IDs.
- Do not violate the constitution.
- Leave the feature in a state that the next verify pass can inspect without guessing what changed, what remains, and why a task is done.
- Do not re-plan the feature, emit a verify verdict, or silently complete neighboring tasks that were outside the active execution scope.

## Progress Rules

- Make the active phase explicit when work crosses a phase boundary.
- When a phase completes within the active scope, emit a short phase-completion update naming the phase and task IDs.
- Keep runtime progress updates in the configured agent language — no English phase-status in a non-English workflow.
- Use short progress lines:
  - `[T1.1] started`
  - `[T1.1] done`
  - `[T1.1] blocked: <reason>`
  - `[Phase 1] done: T1.1, T1.2`
- Load deeper artifacts only when the current task requires them.

## Handoff Rules

- Before marking a task done, confirm that the observable outcome named in the task text is actually present. State explicitly: file changed, what was added/modified, and the observable proof. Do not mark done if the only proof is "code looks correct".
- If the task references `AC-*`, keep implementation aligned with that acceptance scope; do not silently widen behavior.
- When the active scope finishes, leave evidence for the next phase: completed checkboxes, a concise summary of changes, clear blockers or remaining open tasks.
- If implementation reveals that the task text, acceptance coverage, or plan is wrong, stop and send workflow back to `tasks` or `plan` refinement — do not silently repair the process contract in code.

## Output

- Implement the work; update `tasks.md` checkboxes; report phase progress with `[T1.1] done` lines.
- State which acceptance criteria are now covered; do not claim coverage that was not implemented.
- End with a summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`.
- The next phase after `implement` is always `verify`. Do not suggest `/speckeep.archive` directly.
- When ready: `Ready for: /speckeep.verify <slug>`.
