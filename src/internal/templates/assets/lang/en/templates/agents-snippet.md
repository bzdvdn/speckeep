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

## Base Rules (apply to every slash command unless a prompt overrides)

- **Paths**: use `.speckeep/` defaults unless `.speckeep/speckeep.yaml` overrides `paths.specs_dir`, `paths.archive_dir`, or `project.constitution_file`. Read the config at most once per session.
- **Git**: only `/speckeep.spec` creates/switches branches (to `feature/<slug>` or the explicit `--branch`). Other phases must already be on the correct branch — stop and report if not; do not create branches. Never run `git commit`/`push`/`tag` or open PRs unless the user explicitly asks. For CLI use `./.speckeep/scripts/run-speckeep.sh`.
- **Load discipline**: load only the current feature slug by default. Never read unrelated specs/plans, broad repo scans, `/.speckeep/scripts/*` source, or files already read this session (unless you edited them). Prefer readiness-script output over reading deeper artifacts or script sources.
- **Readiness scripts**: when `/.speckeep/scripts/check-<phase>-ready.*` exists, run it with the slug as the first argument (e.g. `bash ./.speckeep/scripts/check-plan-ready.sh <slug>` or `.\.speckeep\scripts\check-plan-ready.ps1 <slug>`). Use its findings as the primary structural evidence layer; do not re-derive them.
- **Language**: use the configured docs language for new/edited artifacts, configured comment language for new/edited code comments. Preserve existing file conventions; do not mix languages inside one artifact without a strong reason.
- **Phase discipline**: do not drift into other phases' work — each command writes only its own artifacts.

Before meaningful changes: review `constitution.md`, the relevant `specs/<slug>/spec.md`, and `specs/<slug>/plan/` if present. After changes: keep specs, plans, tasks, and implementation aligned.
