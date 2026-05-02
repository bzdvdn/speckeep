## SpecKeep

Primary context: `.speckeep/`. Languages: docs=[DOCS_LANGUAGE], agent=[AGENT_LANGUAGE], comments=[COMMENTS_LANGUAGE]

Workflow chain: `constitution → spec → [inspect, optional] → plan → tasks → implement → verify → archive`

Core rules:
- Paths/config: read `.speckeep/speckeep.yaml` ≤ 1 time per session; if missing, defaults: `<specs_dir>=specs`, `<archive_dir>=archive`, constitution=`CONSTITUTION.md`.
- Constitution: prefer `.speckeep/constitution.summary.md` over `CONSTITUTION.md` when loading constitution in any phase.
- Branching: only `/speckeep.spec` may switch/create `feature/<slug>` (or `--branch`). Other phases must already be on the correct branch.
- Scripts: run readiness scripts; trust stdout/exit code; never read `.speckeep/scripts/*` source.
- Scope/load: default to the current slug only; avoid broad repo scans; prefer `Touches:` surfaces.
- Repository map first: if `REPOSITORY_MAP.md` exists, read it before any broad file discovery. Read it once per session and reuse notes; re-read only if you updated the map in the same session.
- Git safety: no `git commit/push/tag` and no PRs unless explicitly asked.
- Done: never mark a task done without observable proof (file path, test output, or command result).
- Discovery: do not run `speckeep ... --help` for discovery; use prompt files and readiness scripts instead.
- CLI: use `./.speckeep/scripts/run-speckeep.sh` (PowerShell: `./.speckeep/scripts/run-speckeep.ps1`) only for actual CLI commands (e.g. `doctor`, `check`, `trace`, `export`, `refresh`). Do not run `run-speckeep.* <phase>` like `spec`/`plan`/`tasks` — phases are slash-commands that write artifacts directly.
- Chat output: do not paste large `git diff`s/full files/long logs. Provide a concise change summary + the list of touched files; if details are needed, show only a small snippet around the edit.
- Scope: do not read or modify artifacts from other slugs/specs unless the current task explicitly requires it (otherwise it’s a scope violation).

Commands:
- `/speckeep.constitution` → update constitution
- `/speckeep.spec` → write spec (branch-first)
- `/speckeep.inspect` → optional deep quality review
- `/speckeep.plan` → write plan package
- `/speckeep.tasks` → write tasks
- `/speckeep.implement` → implement tasks
- `/speckeep.verify` → verify tasks/AC
- `/speckeep.repo-map` → update `REPOSITORY_MAP.md` using the compact template below

Repository map policy (`/speckeep.repo-map`):
- Keep `REPOSITORY_MAP.md` compact and code-only (paths + short role descriptions).
- Language-agnostic: detect stack from repository markers (e.g. `go.mod`, `package.json`, `pyproject.toml`, `Cargo.toml`, `pom.xml`, `*.csproj`) and adapt sections to the detected stack.
- Do not assume Go-specific layout in non-Go projects.
- Hard size cap: keep the map short (target up to 180 lines); if it grows, compress instead of expanding prose.
- Update in place (minimal diff): preserve unchanged lines/order and edit only impacted entries/sections.
- Do not rewrite the whole file if only a subset changed.
- If `REPOSITORY_MAP.md` does not exist, create it from template; otherwise patch existing content.
- Exclude from indexing: `src/internal/agents/**`, `.speckeep/**`, `specs/**`, `archive/**`, `.git/**`, `bin/**`, `demo/**`, `docs/**`, `TESTS/**`, `node_modules/**`, `vendor/**`, `dist/**`, `build/**`, `coverage/**`.
- Note: project settings are already sourced from `.speckeep/speckeep.yaml`; do not duplicate that config in the map.

Update trigger checklist (run `/speckeep.repo-map` if at least one is true):
- Added or removed a top-level code directory/module.
- Moved/renamed key source paths that change navigation.
- Added/removed runtime/service/CLI entrypoints.
- Reshaped subsystem boundaries (where-to-edit paths changed materially).
- User explicitly requested repo map refresh.

Repository map template:
```md
# Repository Map

## Entry Points
- `<path>` — `<runtime/service/cli entrypoint>`

## Top-Level Code
- `<path>` — `<module role>`

## Key Paths
- `<path>` — `<what is implemented here>`

## Where To Edit
- `<change type>` — `<likely paths>`

## Excluded
- `<glob>` — `excluded from indexing`
```
