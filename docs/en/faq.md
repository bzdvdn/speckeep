# FAQ

## What is the difference between `spec` and `inspect`?

`spec` creates or updates the feature specification.

`inspect` reviews that specification and related artifacts for completeness, consistency, and constitutional compliance. It is an optional quality gate, not an authoring phase.

## When should I create `research.md`?

Create `research.md` only when there is real uncertainty that should be preserved:

- external protocol or integration details need investigation
- multiple architecture options are being evaluated
- a design decision needs supporting evidence

Do not create it by default for every feature.

## When should I archive a feature?

Archive a feature when it is no longer active in the main workflow, for example when it is:

- completed
- superseded
- rejected
- abandoned
- deferred

The archive keeps historical context without bloating the active workspace.

## Should I use `archive` or `finish`?

- `speckeep archive <slug> .` is the default and the safe choice: it closes the artifact package only, and you merge through a PR or by hand afterwards. Use it for `completed`, `deferred`, `superseded`, `abandoned`, and `rejected` features.
- `speckeep finish <slug> .` is the local-merge shortcut: archive + commit the snapshot + `merge --no-ff` into `main`/`master` (or `--into <branch>`) + delete the local feature branch. Use it only when you explicitly want the branch closed locally. It never pushes or deletes remote branches.

## When should I use `/spk-propose` instead of `/spk-spec`?

`/spk-propose` is the express lane for a small or medium, low-risk change: it writes `spec.md` + `tasks.md` and may add `data-model.md` or `contracts/api.md`; it never creates `plan.md`. The default is to stay in propose. It escalates only on hard signals: the request spans several features, a breaking change or schema migration with backfill, cross-service coordination or a security boundary, or ACs that cannot be written without inventing decisions. If you are unsure, stay in propose and record the uncertainty under `## Open Questions`; `/spk-plan` remains opt-in after the spec.

## I upgraded but `/spk-propose` still writes `plan.md` — why?

Prompts are project-local. Upgrading the binary does not touch `.speckeep/templates/prompts/` or the generated skills, so run `speckeep refresh .` (preview it with `--dry-run`). In 1.4.1 propose no longer writes `plan.md` at all: it stops at a hard signal and hands off to `/spk-spec` or `/spk-plan`.

## Does `verify` review code quality?

Yes. `verify` has two axes: AC/evidence coverage and a standards review of the implemented diff (duplicated logic, dead code, misleading names, unhandled errors, weakened tests). Findings are capped at 10, recorded in `## Review`, and only `blocker` forces `blocked`. Review never rewrites code — it reports.

## Implementation revealed a gap in the spec — what now?

That is normal, not drift. Update `spec.md` and `tasks.md` together: add or amend the `AC-*` and add the task that covers it, then continue. Never leave the spec ahead of the tasks. If the gap is a design or scope change, stop and return to `/spk-plan` or `/spk-tasks`. `converge` and `check` also report an AC with no task coverage.

## What is the difference between `remove-agent` and `cleanup-agents`?

`remove-agent` updates `.speckeep/speckeep.yaml` and removes generated files for the selected enabled targets.

`cleanup-agents` removes leftover orphaned files that no longer match the enabled target set in config.

Use `doctor` after either one if you want to verify workspace health.

## How do I update prompts, scripts, and agent files in an existing project?

Use `speckeep refresh [path]`.

`refresh` updates only SpecKeep-managed generated artifacts such as `.speckeep/templates/`, `.speckeep/scripts/`, `.speckeep/speckeep.yaml`, project-local agent files, and the managed SpecKeep block in `AGENTS.md`.

It does not modify authored feature state contents such as `constitution.md`, `specs/active/`, or `specs/archived/`, but it can safely move the specs/archive directories via `--specs-dir` / `--archive-dir`.

## Why does SpecKeep keep `Given / When / Then` in English even in Russian docs?

Those markers are intentionally canonical. They are easier for agents to recognize consistently and easier for validation rules to enforce.

The surrounding document text can still be written in Russian.

## Does `implement` always need to read the whole feature package?

No. `implement` should start from `tasks.md` and load deeper artifacts only when the active task requires them.

Typical minimal read order:

- `tasks.md`
- `.speckeep/constitution.summary.md` when policy context is needed
- then `CONSTITUTION.md` only if the summary is absent or insufficient
- then `spec.md`, `plan.md`, `data-model.md`, `contracts/`, or `research.md` only if needed

## Can `plan` run without `spec`?

No. The intended chain is strict:

```text
constitution -> spec -> [inspect, optional] -> [plan, optional] -> tasks -> implement -> [verify, optional] -> archive
```

`plan` depends on an existing spec.

## Why use stable acceptance IDs?

Identifiers such as `AC-001` make traceability cheaper and clearer across specs, tasks, inspect reports, and later verification work.

## What is `verify` for?

`verify` is a lightweight confirmation phase after `implement`. It helps answer whether the feature is aligned enough with tasks, specs, plan artifacts, and project rules to move toward archive or completion claims.

## How much context does SpecKeep add?

Less than the code you read. Typical numbers in v1.4.1:

- always-on guidance (`AGENTS.md`): ~5.3 KB EN / ~6.9 KB RU
- one phase prompt: ~1–8 KB, plus ~0.4 KB of skill scaffolding
- a full feature lifecycle (spec → tasks → implement → verify): ~22 KB EN / ~27 KB RU
- the `sdd` router (only when the phase is unclear): ~1.2 KB

That is roughly 1.3–1.5k tokens per session and 3–7k tokens per feature. The dominant cost is repository reading, not prompts, so the levers that matter are `REPOSITORY_MAP.md`, `Touches:`, `trace --files`, review scoped to `changes.json`, and `/spk-resume`. Run `speckeep context <slug>` to see what each artifact costs in a specific feature.
