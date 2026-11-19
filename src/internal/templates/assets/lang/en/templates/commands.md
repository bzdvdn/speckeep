# SpecKeep Command Reference

Load this file only when the current phase did not name the command you need. Phase prompts live in `.speckeep/templates/prompts/`; each `/spk-<phase>` skill is self-contained.

Default for a new feature: `/spk-propose` → `/spk-implement`. A request like "build X with SpecKeep" routes to `/spk-propose`; name `/spk-spec` for an ambiguous or multi-feature idea, or `/spk-plan` for design work. Plan is opt-in.

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
- `/spk-challenge` — adversarial review of spec/plan
- `/spk-scope` — quick scope boundary check
- `/spk-handoff` — session handoff doc for one feature
- `/spk-recap` — project overview: active features, phase, next step
- `/spk-rollback` — roll back completed tasks
- `/spk-hotfix` — emergency fix outside the phase chain (≤ 3 files, no re-planning)

## Archive (CLI-only)

- `speckeep archive <slug> .` — only when the feature is deterministically proven (all `[x]` tasks have `Proof:`) or after `verify: pass`; with `workflow.verify: required`, archive requires `verify: pass`.
