## SpecKeep

Primary context: `.speckeep/`. Languages: docs=[DOCS_LANGUAGE], agent=[AGENT_LANGUAGE], comments=[COMMENTS_LANGUAGE]

Workflow chain: `constitution → spec → [inspect, optional] → plan → tasks → implement → verify → archive (CLI-only after verify)`

Core rules:
- Paths/config: read `.speckeep/speckeep.yaml` ≤ 1 time per session; if missing, defaults: `<specs_dir>=specs/active`, `<archive_dir>=specs/archived`, constitution=`CONSTITUTION.md`.
- Constitution: load `.speckeep/constitution.summary.md` first if it exists; fall back to `project.constitution_file` (default: `CONSTITUTION.md`) only when the summary is absent.
- Branching: only `/speckeep.spec` may switch/create `feature/<slug>` (or `--branch`). Other phases must already be on the correct branch.
- Scripts: before each phase, run `check-ready.* <phase> <slug>` (and any extras from Commands); trust stdout/exit code; never read `.speckeep/scripts/*` source.
- Scope/load: default to the current slug only; avoid broad repo scans; prefer `Touches:` surfaces.
- ⚠️ **CRITICAL — Repository map first**: **DO NOT** use `ls`, `find`, or glob for primary navigation. Read `REPOSITORY_MAP.md` first — it contains the complete repo map. This saves tokens and maintains workflow discipline. Read it once per session and reuse notes; re-read only if you updated the map in the same session.
- Git safety: no `git commit/push/tag` and no PRs unless explicitly asked.
- Done: never mark a task done without observable proof (file path, test output, or command result). Every artifact must be reviewable by a peer without extra explanation.
- Traceability: for every non-trivial completed task, trace markers are mandatory. No `@sk-task` in changed code or no `@sk-test` in changed tests for that task means the task is not done yet.
- Placement: trace markers must not live at `package`, `import`, or file-header comment level; place them above the owning function/method/test/type declaration.
- End block: every phase output ends with compact summary: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for` (or `Return to` when blocked / `speckeep archive` only after `verify: pass`).
- Discovery: do not run `speckeep ... --help` for discovery; use prompt files and readiness scripts instead.
- CLI: use `./.speckeep/scripts/run-speckeep.sh` (PowerShell: `./.speckeep/scripts/run-speckeep.ps1`) only for actual CLI commands (e.g. `doctor`, `check`, `trace`, `export`, `refresh`). Do not run `run-speckeep.* <phase>` like `spec`/`plan`/`tasks` — phases are slash-commands that write artifacts directly.
- Chat output: do not paste large `git diff`s/full files/long logs. Provide a concise change summary + the list of touched files; if details are needed, show only a small snippet around the edit.
- Scope: do not read or modify artifacts from other slugs/specs unless the current task explicitly requires it (otherwise it’s a scope violation).
- Don't invent: do not introduce requirements, dependencies, scope, or passing criteria absent from current phase inputs.

Commands (⚠️ prefix is **speckeep** with a **p**, NOT speckeek with a k):
- `/speckeep.constitution` → update constitution
- `/speckeep.spec` → write spec (branch-first)
- `/speckeep.inspect` → optional deep quality review
- `/speckeep.plan` → write plan artifacts
- `/speckeep.tasks` → write tasks
- `/speckeep.implement` → implement tasks
- `/speckeep.verify` → verify tasks/AC
- `/speckeep.challenge` → adversarial review of spec/plan (blind spots, untestable AC)
- `/speckeep.rollback` → roll back completed tasks for a feature, returning them to unfinished state
- `/speckeep.recap` → project overview: active features, phase, next step
- `speckeep archive <slug> .` → CLI-only archive after `verify: pass`
- `/speckeep.repo-map` → update `REPOSITORY_MAP.md` (see dedicated prompt for policy + template)

Trigger checklist (run `/speckeep.repo-map` if at least one is true):
- Added or removed a top-level code directory/module.
- Moved/renamed key source paths that change navigation.
- Added/removed runtime/service/CLI entrypoints.
- Reshaped subsystem boundaries (where-to-edit paths changed materially).
- User explicitly requested repo map refresh.
