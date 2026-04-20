# SpecKeep Implement Prompt (compact)

You are implementing a feature strictly from the existing `tasks.md` without expanding scope.

Follow base rules in `AGENTS.md` (paths, git, load discipline, readiness scripts, language, phase discipline).

## Path Resolution

- Resolve `<specs_dir>` from `.speckeep/speckeep.yaml` (read ≤1 time per session). If the config is missing, use `.speckeep/specs`.

## Phase Contract

Inputs: `.speckeep/constitution.md` (or `.speckeep/constitution.summary.md`), `<specs_dir>/<slug>/plan/tasks.md`.
Outputs: repo changes limited to the active task `Touches:` + updated checkboxes in `tasks.md`.
Stop if: `tasks.md` is missing, the next task is not concrete, or execution requires inventing new tasks/AC.

## Execution Rules

- Entrypoint: `tasks.md`. Execute **only** unfinished tasks (`[ ]`) in list order.
- Default scope: only the **first unfinished phase** (unless the user restricts otherwise).
- Read discipline: at session start, batch-read surfaces from `Touches:` for in-scope tasks; read each file ≤ 1 time per session.
- Editing a file outside the active task `Touches:` is a **scope violation** → stop and explain.
- If `/.speckeep/scripts/check-implement-ready.*` exists, run it (slug first) and trust stdout/exit code. Do not read `/.speckeep/scripts/*` source.
- Do not assume `research.md` should exist; only read it if a task explicitly depends on it.
- No redesign / re-planning. If the task cannot be implemented safely from current artifacts → stop and request refinement.
- Annotate every non-trivial change:
  - `// @sk-task <TASK_ID>: <short> (<AC_ID>)`
  - tests: `// @sk-test <TASK_ID>: <TestName> (<AC_ID>)`

## Modes

- `--continue`: start from the first unfinished task, trusting `[x]` tasks without re-verifying.
- `--phase <N>` / `--tasks <list>`: execute only the specified scope, keeping `tasks.md` order. Missing IDs → stop.
- Do not use `--phase` and `--tasks` together.

## Output expectations

- Update code/files and mark completed tasks `[x]` in `tasks.md`.
- Report: closed task IDs, changed files, and the observable proof.
- Include a short summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`.
- Final line (mandatory): `Ready for: /speckeep.verify <slug>`
