## SpecKeep

Primary project context lives in `.speckeep/`. Languages: docs=[DOCS_LANGUAGE], agent=[AGENT_LANGUAGE], comments=[COMMENTS_LANGUAGE]

Workflow chain: `constitution → spec → inspect → plan → tasks → implement → verify → archive`
- `/speckeep.constitution`: create or patch `.speckeep/constitution.md`
- `/speckeep.spec`: create or refine `.speckeep/specs/<slug>/spec.md`; `--amend` for targeted edits (**mandatory** branch-first: before writing any file, switch/create `feature/<slug>` or the explicit `--branch` value)
- `/speckeep.inspect`: check one feature for consistency and quality
- `/speckeep.plan`: create or patch `.speckeep/specs/<slug>/plan/`; `--update` for targeted edits, `--research` for research-first
- `/speckeep.tasks`: create or patch `.speckeep/specs/<slug>/plan/tasks.md`
- `/speckeep.implement`: execute unfinished tasks from `tasks.md`
- `/speckeep.verify`: verify one feature package; `--deep` for full per-AC code tracing
- `/speckeep.archive`: archive to `.speckeep/archive/` (move-based); `--copy` keeps originals, `--restore` unarchives

Optional (any point): `/speckeep.challenge` (adversarial review; `--spec`/`--plan`), `/speckeep.handoff` (session handoff), `/speckeep.hotfix` (emergency fix ≤ 3 files), `/speckeep.scope` (boundary check; `--plan`/`--tasks`), `/speckeep.recap` (project overview)

Read discipline:
- Do not skip phases; load only the current feature slug by default
- Prefer readiness scripts over reading deeper artifacts; use `./.speckeep/scripts/run-speckeep.sh` for CLI access
- Never load: unrelated specs/plans, broad repo scans, script source, files already read this session (unless you edited them)
- Use the configured comment language for new/edited code comments; preserve existing file conventions

Before meaningful changes: review `constitution.md`, the relevant `specs/<slug>/spec.md`, and `specs/<slug>/plan/` if present. After changes: keep specs, plans, tasks, and implementation aligned.
