# Agents

## Supported Agent Targets

SpecKeep can generate project-local command or prompt files for:

- `claude`
- `codex`
- `copilot`
- `cursor`
- `kilocode`
- `trae`
- `all`

## Generated Locations

- Claude: `.claude/commands/`
- Codex: `.codex/prompts/`
- Copilot: `.github/prompts/`
- Cursor: `.cursor/rules/`
- Kilo Code: `.kilocode/rules/`
- Trae: `.trae/project_rules.md`

These generated files are thin wrappers around the canonical SpecKeep prompts in `.speckeep/templates/prompts/`.

## Agent Discipline

The agent-facing workflows are:

- `constitution`
- `spec`
- `inspect`
- `plan`
- `tasks`
- `implement`
- `verify`

Each prompt is designed to:

- read only the minimum required context
- stop when prerequisites are missing
- respect the configured documentation and agent languages
- preserve constitutional authority over specs, plans, tasks, and implementation

Each generated agent wrapper includes:

- **Workflow chain hint**: `constitution → spec → [inspect, optional] → plan → tasks → implement → verify → archive` — prevents agents from skipping required phases or jumping ahead, while making archive an explicit CLI follow-up after agent verification
- **Script execution discipline**: explicit instruction to execute scripts as shell commands, trust stdout/exit code, and never read or inspect script source
- **Anti-pattern block**: common mistakes to avoid — skipping readiness scripts, re-planning during implement, marking tasks done without observable proof, reading the full repository when minimal context is required

`spec` should stay branch-first:

- it should create or switch to `feature/<slug>` before writing `specs/active/<slug>/spec.md` when the environment allows it
- it should support `--name`, optional `--slug`, and optional `--branch` for chat-oriented input
- if `/speckeep.spec` is invoked with `--name` but without enough description, it should preserve context and ask for or accept the next message as the continuation of the spec request
- when the input comes from a local prompt file, it should prefer top-of-file `name:` and optional `slug:` metadata over a generic filename
- if the request is ambiguous, multi-feature, URL-like, or tries to derive one spec from multiple constitutional changes, it should stop and ask for one concrete feature

`verify` is intentionally lightweight:

- it starts from `tasks.md`
- it can use `.speckeep/scripts/verify-task-state.sh <slug>` as a cheap first-pass helper
- `.speckeep/scripts/*` wrappers compute the project root and pass it via `--root`, so they can be executed from any working directory
- it reads deeper artifacts only when needed to confirm a concrete claim
- it is meant to confirm readiness for archive or follow-up refinement, not to become a heavy review engine

After `verify: pass`, prefer the explicit CLI follow-up `speckeep archive <slug> .` so archiving stays outside the agent reasoning loop. Default archive status is `completed`; non-`completed` statuses require an explicit `--reason`. For `completed`, it is fine (and cheap) to reuse `verify-task-state.sh` before creating the snapshot.

## Maintenance Commands

Use the public CLI to manage agent targets safely:

```bash
speckeep add-agent my-project --agents claude --agents cursor
speckeep list-agents my-project
speckeep remove-agent my-project --agents cursor
speckeep cleanup-agents my-project
speckeep doctor my-project
```
