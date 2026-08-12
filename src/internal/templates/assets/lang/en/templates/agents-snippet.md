## SpecKeep

Primary context: `.speckeep/`. Languages: docs=[DOCS_LANGUAGE], agent=[AGENT_LANGUAGE], comments=[COMMENTS_LANGUAGE]

Workflow chain: `constitution → spec → [inspect, optional] → plan → tasks → implement → archive`; `verify` is an optional on-demand audit (always available, skipped by default). Respect `workflow.verify` in `.speckeep/speckeep.yaml`: `required` restores verify as a mandatory pre-archive gate.

Core rules:
- ⚠️ **CRITICAL — Repository map first**: **DO NOT** use `ls`, `find`, or glob for primary navigation. Read `REPOSITORY_MAP.md` first — it contains the complete repo map. This saves tokens and maintains workflow discipline. Read it once per session and reuse notes; re-read only if you updated the map in the same session.
- Paths/config: read `.speckeep/speckeep.yaml` ≤ 1 time per session; if missing, defaults: `<specs_dir>=specs/active`, `<archive_dir>=specs/archived`, constitution=`CONSTITUTION.md`.
- Constitution: load `.speckeep/constitution.summary.md` first if it exists; fall back to `project.constitution_file` (default: `CONSTITUTION.md`) only when the summary is absent.
- Branching: only `/spk.spec` may switch/create `feature/<slug>` (or `--branch`). Other phases must already be on the correct branch.
- Scripts: before each phase, run `check-ready.* <phase> <slug>` (and any extras from Commands); trust stdout/exit code; never read `.speckeep/scripts/*` source.
- Scope/load: default to the current slug only; avoid broad repo scans; prefer `Touches:` surfaces.
- Git safety: no `git commit/push/tag` and no PRs unless explicitly asked.
- Done: never mark a task done without observable proof (file path, test output, or command result). Every artifact must be reviewable by a peer without extra explanation.
- Proof: evidence for every completed task is a `Proof:` line in `tasks.md` directly under the checked task: `Proof: <kind> <path> [<anchor>]` (`kind` = `code|test|docs|chore`). A `[x]` task without a `Proof:` entry is not done yet. `speckeep trace`, `speckeep doctor`, and archive gates read evidence only from `tasks.md` `Proof:` lines.
- End block: every phase output ends with compact summary: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`. `Ready for` is `speckeep archive` once all `[x]` tasks have `Proof:` entries (or after `verify: pass`) UNLESS `workflow.verify: required`, in which case it is `/spk.verify` first. Use `Return to` when blocked.
  - Canonical end block (exact shape, single source — phase prompts reference this, do not re-derive a local variant):
    ```
    Slug: <slug>
    Status: <phase label>
    Artifacts: <paths>
    Blockers: <none | reason>
    Ready for: <next command>   (or "Return to: /spk.<phase> <slug>" when blocked)
    ```
- Verify gate policy (single source — prompts reference this by name; do not fork private copies):
  - `workflow.verify: required` → verify is a **mandatory pre-archive gate**: after implement/hotfix/handoff the `Ready for` line must be `/spk.verify <slug>`, and `speckeep archive` requires `verify.md` with `status: pass`.
  - `workflow.verify: optional` (or absent, default) → verify is an on-demand audit; archive is allowed once all `[x]` tasks carry `Proof:` entries. A `verify.md` present with status ≠ `pass` is still a veto on archive.
- Discovery: do not run `speckeep ... --help` for discovery; use prompt files and readiness scripts instead.
- CLI: use `./.speckeep/scripts/run-speckeep.sh` (PowerShell: `./.speckeep/scripts/run-speckeep.ps1`) only for actual CLI commands (e.g. `doctor`, `check`, `trace`, `export`, `refresh`). Do not run `run-speckeep.* <phase>` like `spec`/`plan`/`tasks` — phases are slash-commands that write artifacts directly.
- Chat output: do not paste large `git diff`s/full files/long logs. Provide a concise change summary + the list of touched files; if details are needed, show only a small snippet around the edit.
- Scope: do not read or modify artifacts from other slugs/specs unless the current task explicitly requires it (otherwise it’s a scope violation).
- Don't invent: do not introduce requirements, dependencies, scope, or passing criteria absent from current phase inputs.
- Escalation: if you cannot honestly complete the current phase (missing artifact, ambiguous intent, unsafe change, blocked gate), STOP and report a `Return to: /spk.<phase> <slug>` or ask the user one targeted question. Never invent a passing criterion, hide a gap, or guess a next step to avoid stopping — stopping with a precise reason is the correct outcome.

Commands (prefix: `/spk.`):
- `/spk.constitution` → update constitution
- `/spk.spec` → write spec (branch-first)
- `/spk.inspect` → optional deep quality review
- `/spk.plan` → write plan artifacts
- `/spk.tasks` → write tasks
- `/spk.implement` → implement tasks
- `/spk.verify` → verify tasks/AC
- `/spk.challenge` → adversarial review of spec/plan (blind spots, untestable AC)
- `/spk.scope` → quick scope boundary check of a feature (in/out, risks)
- `/spk.rollback` → roll back completed tasks for a feature, returning them to unfinished state
- `/spk.recap` → project overview: active features, phase, next step
- `/spk.handoff` → session handoff doc for one feature (resume with zero guesswork)
- `/spk.hotfix` → emergency fix outside the phase chain (≤ 3 files, no re-planning)
- `speckeep archive <slug> .` → CLI-only archive once the feature is deterministically proven (all `[x]` tasks have `Proof:` entries) or after `verify: pass`; with `workflow.verify: required`, archive requires `verify: pass`
- `/spk.repo-map` → update `REPOSITORY_MAP.md` (see dedicated prompt for policy + template)

Trigger checklist (run `/spk.repo-map` if at least one is true):
- Added or removed a top-level code directory/module.
- Moved/renamed key source paths that change navigation.
- Added/removed runtime/service/CLI entrypoints.
- Reshaped subsystem boundaries (where-to-edit paths changed materially).
- User explicitly requested repo map refresh.
