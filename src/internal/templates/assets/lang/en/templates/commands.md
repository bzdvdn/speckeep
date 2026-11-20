# SpecKeep Command Reference

Load this file only when the current phase did not name the command you need. Phase prompts live in `.speckeep/templates/prompts/`; each `/spk-<phase>` skill is self-contained.

Default for a new feature: `/spk-propose` → `/spk-implement`. A request like "build X with SpecKeep" routes to `/spk-propose`; name `/spk-spec` for an ambiguous or multi-feature idea, or `/spk-plan` for design work. Plan is opt-in.

`Ready for` may include an `(alt: ...)` parenthetical; the first command is the primary/safe default.

## Phases (readiness-gated)

- `/spk-constitution` — create/update the project constitution
- `/spk-propose` — recommended default: idea → spec + tasks (plan optional), straight to implement
- `/spk-spec` — write spec (branch-first)
- `/spk-inspect` — optional deep quality review
- `/spk-plan` — opt-in design artifacts (only when the feature needs design)
- `/spk-tasks` — write tasks
- `/spk-implement` — implement tasks
- `/spk-verify` — verify tasks/AC (optional audit by default)
- `/spk-converge` — fast closing loop: re-check tasks/proofs, append follow-ups, repeat until converged

Run `./.speckeep/scripts/check-ready[SCRIPT_EXT] <phase> <slug>` before a gated phase. Auxiliary commands have no readiness gate.

## Auxiliary commands (no readiness gate)

- `/spk-repo-map` — update `REPOSITORY_MAP.md` (triggers: see `.speckeep/templates/prompts/repo-map.md`)
- `/spk-module-context <path>` — write/refresh `<path>/CONTEXT.md`
- `/spk-glossary` — create/update `.speckeep/glossary.md`
- `/spk-grill` — optional pre-spec interview: resolve the design tree before a spec
- `/spk-research` — optional technical research: write a cited `research.md`
- `/spk-diagnose` — optional bug diagnosis loop: reproduce → minimize → hypothesize → fix → regression
- `/spk-challenge` — adversarial review of spec/plan
- `/spk-scope` — quick scope boundary check
- `/spk-handoff` — session handoff doc for one feature
- `/spk-resume` — minimal in-session resume: state packet + next task (no artifact)
- `/spk-recap` — project overview: active features, phase, next step
- `/spk-rollback` — roll back completed tasks
- `/spk-hotfix` — emergency fix outside the phase chain (≤ 3 files, no re-planning)

## Close (CLI-only)

- `speckeep archive <slug> .` — default: archive artifacts only (PR/manual merge, deferred/abandoned). Completed archives require closed tasks, `Proof:` on every `[x]` task, and `verify: pass` when `workflow.verify: required`; non-completed statuses may park incomplete work. Legacy completed features without `Proof:` must add evidence or be archived with a non-completed status and a reason.
- `speckeep finish <slug> .` — local close: archive + commit the snapshot + merge `--no-ff` into the base branch + delete the local feature branch. Use only when the user explicitly asks to merge/close. If the user names a target branch, pass `--into <branch>` (e.g. `--into dev`); it must exist locally. Never pushes or deletes remote branches.
