## SpecKeep

Primary context: `.speckeep/`. Languages: docs=[DOCS_LANGUAGE], agent=[AGENT_LANGUAGE], comments=[COMMENTS_LANGUAGE]

Workflow chain: `constitution → spec → inspect → plan → tasks → implement → verify → archive`

Core rules:
- Paths/config: read `.speckeep/speckeep.yaml` ≤ 1 time per session; if missing, defaults: `<specs_dir>=specs`, `<archive_dir>=archive`, constitution=`CONSTITUTION.md`.
- Constitution: prefer `.speckeep/constitution.summary.md` over `CONSTITUTION.md` when loading constitution in any phase.
- Branching: only `/speckeep.spec` may switch/create `feature/<slug>` (or `--branch`). Other phases must already be on the correct branch.
- Scripts: run readiness scripts; trust stdout/exit code; never read `.speckeep/scripts/*` source.
- Scope/load: default to the current slug only; avoid broad repo scans; prefer `Touches:` surfaces.
- Git safety: no `git commit/push/tag` and no PRs unless explicitly asked.
- Done: never mark a task done without observable proof (file path, test output, or command result).
- Discovery: do not run `speckeep ... --help` for discovery; use prompt files and readiness scripts instead.
- CLI: use `./.speckeep/scripts/run-speckeep.sh` (PowerShell: `./.speckeep/scripts/run-speckeep.ps1`) only for actual CLI commands (e.g. `doctor`, `check`, `trace`, `export`, `refresh`). Do not run `run-speckeep.* <phase>` like `spec`/`plan`/`tasks` — phases are slash-commands that write artifacts directly.
- Chat output: do not paste large `git diff`s/full files/long logs. Provide a concise change summary + the list of touched files; if details are needed, show only a small snippet around the edit.
- Scope: do not read or modify artifacts from other slugs/specs unless the current task explicitly requires it (otherwise it’s a scope violation).

Commands:
- `/speckeep.constitution` → update constitution
- `/speckeep.spec` → write spec (branch-first)
- `/speckeep.inspect` → inspect spec
- `/speckeep.plan` → write plan package
- `/speckeep.tasks` → write tasks
- `/speckeep.implement` → implement tasks
- `/speckeep.verify` → verify tasks/AC
- `/speckeep.archive` → archive
