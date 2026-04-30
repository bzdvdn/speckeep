# SpecKeep Implement Prompt (compact)

You are implementing a feature strictly from the existing `tasks.md` without expanding scope.

Follow base rules in `AGENTS.md` (paths, git, load discipline, readiness scripts, language, phase discipline).

## Phase Contract

Inputs: `project.constitution_file` (default: `CONSTITUTION.md`), `<specs_dir>/<slug>/plan/tasks.md`.
Outputs: repo changes limited to the active task `Touches:` + updated checkboxes in `tasks.md`.
Stop if: `tasks.md` is missing, the next task is not concrete, or execution requires inventing new tasks/AC.

## Execution Rules

- Entrypoint: `tasks.md`. Execute **only** unfinished tasks (`[ ]`) in list order.
- Default scope: only the **first unfinished phase** (unless the user restricts otherwise).
- Before reading any other file, explicitly state `Active phase: T<N>` and list the active task IDs you will execute in this run (only `T<N>.*` from the first unfinished phase). Do not proceed until this is clear.
- Do not read or edit anything before selecting the active tasks, except `tasks.md` itself.
- Do not move to phase `T(N+1).*` until all `T(N).*` tasks are checked `[x]` in `tasks.md` and you list observable proof per task (files/tests/trace/command output).
- Read discipline: at session start, batch-read surfaces from `Touches:` for in-scope tasks; read each file ≤ 1 time per session.
- Do not re-read already opened files end-to-end “for reassurance”: keep short notes and use targeted slices (`rg`, `sed -n`) and `git diff` to verify changes.
- If `tasks.md` lists “Inputs” at the top (e.g., `plan.md`/`spec.md`/`data-model.md`), do not treat them as mandatory re-reads during implement: open them only when a concrete active task explicitly requires it, or when `tasks.md` is missing critical context.
- Editing a file outside the active task `Touches:` is a **scope violation** → stop and explain.
- Tests: run only targeted package/tests. Do not run `go test ./...` unless explicitly requested. Do not paste long logs; summarize and include only the last lines when needed.
- If `./.speckeep/scripts/check-implement-ready.*` exists, run it (slug first) and trust stdout/exit code. Do not read `./.speckeep/scripts/*` source.
- Do not assume `research.md` should exist; only read it if a task explicitly depends on it.
- No redesign / re-planning. If the task cannot be implemented safely from current artifacts → stop and request refinement.
- Prefer minimal patches over full-file rewrites. Do not rewrite a whole file “for simplicity” unless strictly necessary.
- Annotate every non-trivial change:
  - always (avoid cross-spec collisions): `// @sk-task <slug>#<TASK_ID>: <short> (<AC_ID>)`
  - tests: `// @sk-test <slug>#<TASK_ID>: <TestName> (<AC_ID>)`
  - legacy is allowed only if `<slug>` is unknown (not in args/context) or if you're intentionally preserving the existing code style: `// @sk-task <TASK_ID> ...` / `// @sk-test <TASK_ID> ...`
- Annotation placement:
  - place `@sk-test` directly above the target test function declaration (for example above `func Test...`), not on struct fields/local variables/assert blocks.
  - place `@sk-task` near the owning behavior entrypoint (function/method/block header), not on unrelated field lines.
- Traceability is mandatory for implementation work:
  - do not mark a task complete if changed code lacks `@sk-task` or changed tests lack `@sk-test` for that task.
  - before finalizing, run a quick self-check on changed files to confirm trace markers are present.
  - minimum self-check command pattern before closing tasks: inspect changed files and confirm `@sk-task` / `@sk-test` presence via targeted search (for example with `git diff --name-only` + `rg`).
  - never replace existing trace markers when adding coverage for a new task; append new markers.
  - if one method/test covers multiple tasks, keep multiple annotations on that same method/test (for example both `@sk-task <slug>#T1.1 ...` and `@sk-task <slug>#T3.1 ...`).

## Modes

- `--continue`: start from the first unfinished task, trusting `[x]` tasks without re-verifying.
- `--phase <N>` / `--tasks <list>`: execute only the specified scope, keeping `tasks.md` order. Missing IDs → stop.
- Do not use `--phase` and `--tasks` together.

## Output expectations

- Update code/files and mark completed tasks `[x]` in `tasks.md`.
- Before finalizing, make an explicit map decision line: `Map update: yes|no` + reason (based on `/speckeep.repo-map` trigger checklist in `AGENTS.md`).
- If `Map update: yes`, run `/speckeep.repo-map` and include `REPOSITORY_MAP.md` in changed files.
- If repository structure/navigation changed (new/moved modules, new entrypoints, major path reshaping), `Map update` must be `yes`.
- If changes are local and do not affect structure/navigation, do not touch `REPOSITORY_MAP.md`.
- Report: closed task IDs, changed files, and the observable proof.
- Include a short traceability proof line (which files contain `@sk-task` / `@sk-test` for the closed tasks).
- Do not finish without explicit `Trace proof:` lines in the form: `<TASK_ID> -> <file>:<line> (@sk-task|@sk-test)`.
- Include a short summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`.
- Final line (mandatory): `Ready for: /speckeep.verify <slug>`
