## SpecKeep

Primary context: `.speckeep/`. Languages: docs=[DOCS_LANGUAGE], agent=[AGENT_LANGUAGE], comments=[COMMENTS_LANGUAGE]

Workflow chain: `constitution → spec → inspect → plan → tasks → implement → verify → archive`

Core rules:
- Paths/config: use `.speckeep/` defaults; read `.speckeep/speckeep.yaml` ≤ 1 time per session.
- Branching: only `/speckeep.spec` may switch/create `feature/<slug>` (or `--branch`). Other phases must already be on the correct branch.
- Scripts: run readiness scripts; trust stdout/exit code; never read `/.speckeep/scripts/*` source.
- Scope/load: default to the current slug only; avoid broad repo scans; prefer `Touches:` surfaces.
- Git safety: no `git commit/push/tag` and no PRs unless explicitly asked.
- CLI: use `./.speckeep/scripts/run-speckeep.sh` (PowerShell: `./.speckeep/scripts/run-speckeep.ps1`).

Commands:
- `/speckeep.constitution` → update constitution
- `/speckeep.spec` → write spec (branch-first)
- `/speckeep.inspect` → inspect spec
- `/speckeep.plan` → write plan package
- `/speckeep.tasks` → write tasks
- `/speckeep.implement` → implement tasks
- `/speckeep.verify` → verify tasks/AC
- `/speckeep.archive` → archive
