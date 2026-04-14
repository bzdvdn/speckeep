# FAQ

## What is the difference between `spec` and `inspect`?

`spec` creates or updates the feature specification.

`inspect` reviews that specification and related artifacts for completeness, consistency, and constitutional compliance. It is a quality gate, not an authoring phase.

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

## What is the difference between `remove-agent` and `cleanup-agents`?

`remove-agent` updates `.speckeep/speckeep.yaml` and removes generated files for the selected enabled targets.

`cleanup-agents` removes leftover orphaned files that no longer match the enabled target set in config.

Use `doctor` after either one if you want to verify workspace health.

## How do I update prompts, scripts, and agent files in an existing project?

Use `speckeep refresh [path]`.

`refresh` updates only SpecKeep-managed generated artifacts such as `.speckeep/templates/`, `.speckeep/scripts/`, `.speckeep/speckeep.yaml`, project-local agent files, and the managed SpecKeep block in `AGENTS.md`.

It does not modify authored feature state contents such as `constitution.md`, `specs/`, or `archive/`, but it can safely move the specs/archive directories via `--specs-dir` / `--archive-dir`.

## Why does SpecKeep keep `Given / When / Then` in English even in Russian docs?

Those markers are intentionally canonical. They are easier for agents to recognize consistently and easier for validation rules to enforce.

The surrounding document text can still be written in Russian.

## Does `implement` always need to read the whole feature package?

No. `implement` should start from `tasks.md` and load deeper artifacts only when the active task requires them.

Typical minimal read order:

- `constitution.md`
- `tasks.md`
- then `spec.md`, `plan.md`, `data-model.md`, `contracts/`, or `research.md` only if needed

## Can `plan` run without `spec`?

No. The intended chain is strict:

```text
constitution -> spec -> inspect -> plan -> tasks -> implement -> verify -> archive
```

`plan` depends on an existing spec.

## Why use stable acceptance IDs?

Identifiers such as `AC-001` make traceability cheaper and clearer across specs, tasks, inspect reports, and later verification work.

## What is `verify` for?

`verify` is a lightweight confirmation phase after `implement`. It helps answer whether the feature is aligned enough with tasks, specs, plan artifacts, and project rules to move toward archive or completion claims.
