## SpecKeep

Primary context: `.speckeep/`. Languages: docs=[DOCS_LANGUAGE], agent=[AGENT_LANGUAGE], comments=[COMMENTS_LANGUAGE]

Workflow chain: `constitution → spec → [inspect, optional] → [plan, optional] → tasks → implement → archive`; `verify` is an optional on-demand audit (always available, skipped by default). `/spk-propose` is the default for a new feature (spec + tasks in one pass): use it directly and skip the `sdd` router; `/spk-spec` only for ambiguous or multi-feature ideas; plan is opt-in for design-heavy work. Respect `workflow.verify` in `.speckeep/speckeep.yaml`: `required` restores verify as a mandatory pre-archive gate.

Core rules:
- ⚠️ **Repository map first**: do not use `ls`/`find`/glob for primary navigation — read `REPOSITORY_MAP.md` first (complete map, saves tokens). Read it once per session; re-read only after you updated the map yourself.
- Paths/config: read `.speckeep/speckeep.yaml` ≤ 1 time per session; it defines `<specs_dir>`, `<archive_dir>`, and `project.constitution_file`. If the config is missing, defaults are `specs/active`, `specs/archived`, `CONSTITUTION.md`.
- Constitution: load `.speckeep/constitution.summary.md` first when present; otherwise use `project.constitution_file` from the config.
- Domain language: if `.speckeep/glossary.md` exists, read it once per session and reuse its terms in spec/plan/tasks; never introduce a synonym for a term it already defines. It is optional and created/updated only via `/spk-glossary` — do not generate it implicitly from another phase.
- Branching: only `/spk-spec` and `/spk-propose` may switch/create `feature/<slug>` (or `--branch`). Other phases must already be on the correct branch.
- Scripts: run `./.speckeep/scripts/check-ready[SCRIPT_EXT] <phase> <slug>` only before gated phases (`constitution`, `spec`, `propose`, `inspect`, `plan`, `tasks`, `implement`, `verify`, `converge`; `archive` is CLI). Auxiliary commands (`repo-map`, `module-context`, `glossary`, `challenge`, `handoff`, `recap`, `scope`, `rollback`, `hotfix`) have no gate. Trust the exit code; never read `.speckeep/scripts/*` source.
- Skills: the workflow is a skill pack (`sdd` router + one self-contained `spk-<phase>/SKILL.md` per phase). Load only the matching phase skill; load `sdd` only when the phase is unclear.
- Scope/load: default to the current slug only; avoid broad repo scans; prefer `Touches:` surfaces.
- Git safety: no `git commit/push/tag` and no PRs unless explicitly asked.
- Done: never mark a task done without observable proof (file path, test output, or command result). Every artifact must be reviewable by a peer without extra explanation.
- Proof: evidence for every completed task is a `Proof:` line in `tasks.md` directly under the checked task: `Proof: <kind> <path> [<anchor>]` (`kind` = `code|test|docs|chore`). A `[x]` task without a `Proof:` entry is not done yet. `speckeep trace`, `speckeep doctor`, and archive gates read evidence only from `tasks.md` `Proof:` lines.
- Trace: right after `feature/<slug>`, pin the base via `./.speckeep/scripts/run-speckeep[SCRIPT_EXT] trace init <slug>`. `Touches:` is the plan; `speckeep trace <slug> --files` shows actual changes (`Unmapped` = drift); `speckeep trace --file <path>` finds specs/tasks that changed a file.
- Token discipline: staying in the **same phase** — continue in-session (user replies `continue`); do NOT re-invoke the same slash command/skill, it reloads the whole prompt. A phase transition loads the next prompt exactly once; invoke a new command only when moving to a **different phase** or after a context reset. The `Ready for` line names the next phase + a `continue?` question, not a re-run of the current command.
- End block: every phase output ends with the canonical block below (use `Return to` when blocked).
  - Canonical end block (exact shape, single source — phase prompts reference this, do not re-derive a local variant):
    ```
    Slug: <slug>
    Status: <phase label>
    Artifacts: <paths>
    Blockers: <none | reason>
    Ready for: <next command>   (or "Return to: /spk-<phase> <slug>" when blocked)
    ```
- Verify gate policy (single source — prompts reference this by name; do not fork private copies):
  - `workflow.verify: required` → verify is a **mandatory pre-archive gate**: after implement/hotfix/handoff the `Ready for` line must be `/spk-verify <slug>`, and `speckeep archive` requires `verify.md` with `status: pass`.
  - `workflow.verify: optional` (or absent, default) → verify is an on-demand audit; archive is allowed once all `[x]` tasks carry `Proof:` entries. A `verify.md` present with status ≠ `pass` is still a veto on archive.
- Discovery: do not run `speckeep ... --help` for discovery; use prompt files and readiness scripts instead.
- CLI: use `./.speckeep/scripts/run-speckeep[SCRIPT_EXT]` only for actual CLI commands (e.g. `doctor`, `check`, `trace`, `export`, `refresh`). Do not run `run-speckeep[SCRIPT_EXT] <phase>` like `spec`/`plan`/`tasks` — phases are slash-commands that write artifacts directly.
- Chat output: do not paste large `git diff`s/full files/long logs. Provide a concise change summary + the list of touched files; if details are needed, show only a small snippet around the edit.
- Scope: do not read or modify artifacts from other slugs/specs unless the current task explicitly requires it (otherwise it’s a scope violation).
- Don't invent: do not introduce requirements, dependencies, scope, or passing criteria absent from current phase inputs.
- Escalation: if you cannot honestly complete the current phase (missing artifact, ambiguous intent, unsafe change, blocked gate), STOP and report a `Return to: /spk-<phase> <slug>` or ask the user one targeted question. Never invent a passing criterion, hide a gap, or guess a next step to avoid stopping — stopping with a precise reason is the correct outcome.

Command reference: load `.speckeep/templates/commands.md` only when you need a command the current phase did not name. Repo-map triggers live in `.speckeep/templates/prompts/repo-map.md` (`/spk-repo-map`).
