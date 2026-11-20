## SpecKeep

Primary context: `.speckeep/`. Languages: docs=[DOCS_LANGUAGE], agent=[AGENT_LANGUAGE], comments=[COMMENTS_LANGUAGE]

Workflow chain: `constitution → spec → [inspect, optional] → [plan, optional] → tasks → implement → archive`; `verify` is an optional on-demand audit (skipped by default). `/spk-propose` is the default for a new feature (spec + tasks in one pass): use it directly and skip the `sdd` router; `/spk-spec` only for ambiguous or multi-feature ideas; plan is opt-in for design-heavy work. Respect `workflow.verify` in `.speckeep/speckeep.yaml`: `required` restores verify as a mandatory pre-archive gate.

Core rules:
- ⚠️ **Repository map first**: do not use `ls`/`find`/glob for primary navigation — read `REPOSITORY_MAP.md` first (complete map, saves tokens). Read it once per session; re-read only after you updated the map yourself.
- Paths/config: read `.speckeep/speckeep.yaml` ≤ 1 time per session; it defines `<specs_dir>`, `<archive_dir>`, and `project.constitution_file`. If the config is missing, defaults are `specs/active`, `specs/archived`, `CONSTITUTION.md`.
- Constitution: load `.speckeep/constitution.summary.md` first when present; otherwise use `project.constitution_file` from the config.
- Domain language: read `.speckeep/glossary.md` once per session if present; reuse its terms and never add synonyms. Optional; only `/spk-glossary` creates it.
- Branching: only `/spk-spec` and `/spk-propose` may switch/create `feature/<slug>` (or `--branch`). Other phases must already be on the correct branch.
- Scripts: run `./.speckeep/scripts/check-ready[SCRIPT_EXT] <phase> <slug>` only for gated phases (`constitution`, `spec`, `propose`, `inspect`, `plan`, `tasks`, `implement`, `verify`, `converge`); auxiliary commands have no gate. Trust the exit code; never read `.speckeep/scripts/*` source.
- Skills: the workflow is a skill pack (`sdd` router + one self-contained `spk-<phase>/SKILL.md` per phase). Load only the matching phase skill; load `sdd` only when the phase is unclear.
- Scope/load: current slug only; no broad scans or other slugs' artifacts; prefer `Touches:` surfaces. For shapes/examples use `.speckeep/templates/*`, never other slugs.
- Git safety: no `git commit/push/tag` and no PRs unless explicitly asked.
- Done: never mark a task done without observable proof (file path, test output, or command result). Every artifact must be reviewable by a peer without extra explanation.
- Proof: every `[x]` task needs `Proof: <kind> <path> [<anchor>]` (`code|test|docs|chore`) directly under it in `tasks.md`; otherwise it is not done. `trace`, `doctor`, and archive read only these lines.
- Trace: after `feature/<slug>`, run `./.speckeep/scripts/run-speckeep[SCRIPT_EXT] trace init <slug>`. `Touches:` is the plan; `trace <slug> --files` shows actual changes (`Unmapped` = drift); `trace --file <path>` finds who changed a file.
- Token discipline: same phase → continue in-session (`continue`); do not re-invoke the same skill (full prompt reload). A phase transition loads the next prompt once. `Ready for` names the next step, not a re-run.
- End block: every phase ends with the canonical block below (`Return to` when blocked).
  - Canonical end block (single source; prompts reference this):
    ```
    Slug: <slug>
    Status: <phase label>
    Artifacts: <paths>
    Blockers: <none | reason>
    Ready for: <next command>   (or "Return to: /spk-<phase> <slug>" when blocked)
    ```
  - `Ready for` may include `(alt: ...)`; the first command is the primary/safe default.
- Verify gate policy (single source; prompts reference by name):
  - `required` → verify is mandatory before archive: after implement/hotfix/handoff `Ready for` = `/spk-verify <slug>`; archive requires `verify.md` status `pass`.
  - `optional`/absent → on-demand audit; archive allowed when all `[x]` have `Proof:`. A `verify.md` ≠ `pass` still vetoes archive.
- Discovery: do not run `speckeep ... --help` for discovery; use prompt files and readiness scripts instead.
- CLI: use `./.speckeep/scripts/run-speckeep[SCRIPT_EXT]` only for real CLI commands (`doctor`, `check`, `trace`, `export`, `refresh`); never `run-speckeep[SCRIPT_EXT] <phase>` — phases are slash commands.
- Close: `archive` is default (PR/manual merge, deferred/abandoned); use `finish` only when the user explicitly asks to merge/close a local branch. Named target branch → `--into <branch>` (must exist locally). Never push or delete remote branches. Do not use `deferred`/`abandoned` to bypass `Proof:`; non-completed statuses are only for genuinely parked work.
- Output: no large diffs/full files/logs; concise summary + touched files, small snippets only on request.
- Don't invent: do not introduce requirements, dependencies, scope, or passing criteria absent from current phase inputs.
- Escalation: if a phase cannot be completed honestly (missing artifact, ambiguous intent, unsafe change, blocked gate), STOP and report `Return to: /spk-<phase> <slug>` or ask one targeted question. Never invent a pass, hide a gap, or guess.

Command reference: load `.speckeep/templates/commands.md` only when the phase did not name the command; repo-map triggers are in `.speckeep/templates/prompts/repo-map.md` (`/spk-repo-map`).
