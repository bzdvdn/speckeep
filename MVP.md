# speckeep MVP

## Product statement

`speckeep` provides a minimal, file-based context system for software projects.

It helps development agents and humans work from the same project context:

- what the project is trying to do
- what a feature should do
- how the feature will be implemented
- what contracts and data models are required
- what rules are non-negotiable
- how development must be conducted
- what decisions are currently in effect
- which language generated docs and agent prompts should use

The product should feel lightweight, editable, and resilient.

SpecKeep should preserve traceability through compact, stable IDs instead of broader default context or extra summary layers.

## Non-goals

- no mandatory checkpoint flow
- no quickstart wizard
- no built-in approval process
- no AI orchestration embedded into the CLI

Managed generated artifacts should remain refreshable without touching authored feature state. `refresh` should update templates, scripts, config-derived values, project-local agent files, and the managed SpecKeep block in `AGENTS.md` while leaving `constitution`, `specs`, `plans`, and `archive` untouched.

## Workspace layout

```text
.speckeep/
  speckeep.yaml
  constitution.md
  specs/
    <slug>/
      spec.md
      inspect.md
      summary.md
      hotfix.md
      plan/
        plan.md
        tasks.md
        data-model.md
        research.md
        contracts/
          api.md
          events.md
  archive/
    <slug>/
      <YYYY-MM-DD>/
      summary.md
      spec.md
      plan.md
      tasks.md
      data-model.md
      research.md
      contracts/
  templates/
    constitution.md
    spec.md
    plan.md
    tasks.md
    data-model.md
    inspect-report.md
    verify-report.md
    contracts/
      api.md
      events.md
    archive/
      summary.md
    agents-snippet.md
    prompts/
      constitution.md
      spec.md
      inspect.md
      plan.md
      tasks.md
      implement.md
      verify.md
      archive.md
      challenge.md
      handoff.md
      hotfix.md
      recap.md
      scope.md
  scripts/
    run-speckeep.sh
    inspect-spec.sh
    trace.sh
    check-*.sh
AGENTS.md
.kilocode/
  workflows/
    speckeep-*.md
```

## Phase model

The intended agent workflow is strict:

1. `constitution`
2. `spec`
3. `inspect`
4. `plan`
5. `tasks`
6. `implement`
7. `verify`
8. `archive`

Dependency rules:

- `constitution` can be created first
- `spec` depends on the constitution
- `inspect` depends on the constitution and one spec
- `plan` depends on the constitution, one spec, and the persisted inspect report
- `tasks` depends on the constitution and one plan package
- `implement` depends on the constitution and one task list
- `verify` depends on the constitution and one task list
- `archive` depends on one existing spec and archives the related plan package

## Optional workflow commands

Available at any phase:

- `/speckeep.challenge`: adversarial review of spec or plan — finds weak assumptions, untestable AC, scope drift
- `/speckeep.handoff`: compact session handoff document for new sessions
- `/speckeep.hotfix`: emergency fix outside standard chain (≤3 files, known root cause)
- `/speckeep.scope`: quick scope boundary check (inline only, no file)
- `/speckeep.recap`: project-level overview of all active features

## Language model

`speckeep init` supports compact language configuration:

Defaults:

- `default language`: `en`
- supported values: `en`, `ru`

Controls:

- `docs language`: generated project docs and templates
- `agent language`: generated prompts and inserted `AGENTS.md` guidance
- `comments language`: preferred code comment language
- `shell`: generated workflow script family (`sh` or `powershell`)

Language settings stored in `.speckeep/speckeep.yaml`.

## Token efficiency goals

SpecKeep should stay meaningfully lighter than SpecKit by default.

Design constraints for low token usage:

- phase prompts should be short and explicit
- each phase should read only one feature package at a time
- targeted code reading allowed when it removes downstream guesswork
- optional artifacts should stay optional
- readiness scripts enforce prerequisites instead of pushing work into model context
- patch existing files instead of regenerating large documents
- avoid loading unrelated feature plans, tasks, or contracts
- use `plan.md` as tasks entrypoint
- use `tasks.md` as implement entrypoint

Lightweight guardrails:

- each phase defines `always load`, `load if needed`, `never load by default`
- `implement` stays task-scoped by default
- `verify` stays cheap by default, deepens only when explicitly requested
- helper scripts preferred over repeated prompt-time reasoning
- traceability through stable IDs instead of shared summary artifacts

Recommended stable ID scheme:

- `RQ-*` for spec requirements
- `AC-*` for acceptance criteria
- `DEC-*` for plan decisions
- `DM-*` for data-model entities
- `API-*` for API contracts
- `EVT-*` for event contracts
- `T<phase>.<index>` for tasks

## Plan package

Each feature plan lives under `.speckeep/specs/<slug>/plan/`.

Required artifacts:

- `plan.md` (includes `## Incremental Delivery` section)
- `tasks.md`
- `data-model.md`

Optional artifacts:

- `contracts/api.md` — when feature touches API boundaries
- `contracts/events.md` — when feature produces/consumes events
- `research.md` — only when genuine uncertainty exists

## Constitution workflow

`constitution` is agent-driven and strict.

Mandatory sections:

- `Purpose`
- `Core Principles`
- `Constraints`
- `Language Policy`
- `Development Workflow`
- `Governance`
- `Last Updated`

The constitution is authoritative over specs, plans, tasks, and implementation.

## Inspect workflow

`inspect` is agent-driven.

Inputs:

- `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/spec.md`
- optional plan artifacts when they exist (cheapest scope first)

Outputs:

- focused inspection report for one feature
- persisted inspect report at `.speckeep/specs/<slug>/inspect.md`
- explicit Given/When/Then acceptance criteria
- cheap checks for constitution↔spec, spec↔plan, plan↔tasks

## Verify workflow

`verify` is agent-driven.

Inputs:

- `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/plan/tasks.md`
- optional spec, plan, data-model, contracts, targeted code

Outputs:

- verdict: `pass`, `concerns`, `blocked`
- compact by default, persisted when explicitly requested
- supports `--deep` mode for full implementation validation

The verify phase must:

- treat task list as entrypoint
- confirm completed task state against implementation evidence
- use `speckeep trace` as primary evidence
- prefer `concerns` over `pass` when evidence is partial

## Archive workflow

`archive` is agent-driven.

Inputs:

- `.speckeep/specs/<slug>/spec.md`
- optional plan artifacts

Outputs:

- `.speckeep/archive/<slug>/<YYYY-MM-DD>/summary.md`
- archived copies of spec and plan artifacts
- move-based by default (deletes originals after copy)
- supports `--restore` to reverse archive

## Traceability

`speckeep trace <slug> [path]` scans for `@sk-task` and `@sk-test` annotations in code.

Provides verifiable traceability between:

- requirements (AC-*)
- tasks (T*)
- implementation (code annotations)
- tests (@sk-test)

## CLI entrypoint contract

Generated workspaces include launcher script under `.speckeep/scripts/`:

- `run-speckeep.sh` for `sh`
- `run-speckeep.ps1` for `powershell`

Resolution order:

1. `DRAFTSPEC_BIN`
2. `speckeep` from `PATH`

## Spec workflow

`spec` is agent-driven.

Inputs:

- `.speckeep/constitution.md`
- user request
- minimal repository context when needed

Supports:

- `--name <feature name>`
- `--slug <feature-slug>`
- `--branch <branch-name>`
- `--amend`: targeted edit mode

Output:

- `.speckeep/specs/<slug>/spec.md`
- work from `feature/<slug>` branch

Spec includes:

- `RQ-*` IDs for requirements
- `AC-*` IDs for acceptance criteria (Given/When/Then)
- `## Assumptions` section (mandatory)
- `## Success Criteria` with `SC-*` IDs (optional)
- `[NEEDS CLARIFICATION: ...]` inline markers

## Plan workflow

`plan` translates spec into technical design.

Inputs:

- `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/spec.md`
- `.speckeep/specs/<slug>/inspect.md` (required)

Outputs:

- `.speckeep/specs/<slug>/plan/plan.md`
- `.speckeep/specs/<slug>/plan/data-model.md`
- optional contracts and research

Plan includes:

- `DEC-*` IDs for implementation decisions
- `## Constitution Compliance` section (mandatory)
- `## Incremental Delivery` with MVP and iterative expansion
- References to `AC-*` for acceptance-critical behavior

Supports `--update` for targeted edits without rewriting entire plan.

## Tasks workflow

`tasks` reads:

- `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/plan/plan.md`
- reads spec/data-model/contracts only as needed

Must:

- produce concrete executable tasks
- group by implementation phase
- assign phase-scoped IDs (T1.1, T1.2, etc.)
- map each `AC-*` to at least one task
- include `## Surface Map` (MUST)
- include `Touches:` in each task (MUST)

## Implement workflow

`implement` reads:

- `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/plan/tasks.md`
- reads spec/plan/data-model/contracts only as needed

Must:

- execute only unfinished tasks
- respect task order and phase structure
- update `tasks.md`
- emit progress updates
- report coverage in terms of completed task IDs and AC-* IDs
- allow in-place decomposition for active task only

## Status and dashboard

`speckeep check <slug>` shows:

- artifact presence
- inspect and verify verdict
- task progress
- next slash command
- readiness summary

`speckeep check --all` shows readiness table across all features.

`speckeep dashboard` provides visual dashboard of all active features.

## Repair and migration

`feature repair` and `migrate` are CLI-driven.

Safe repair scope:

- migrate legacy flat spec artifacts to `.speckeep/specs/<slug>/`
- migrate legacy inspect reports to `.speckeep/specs/<slug>/inspect.md`
- remove duplicate legacy copies when byte-identical
- stop with warning when canonical and legacy copies differ

## Doctor checks

`doctor` reports:

- missing required files (error)
- orphaned agent artifacts (warning)
- cross-spec ID collisions (warning)
- unfilled constitution placeholders (warning)
- branch name mismatches (warning)
- orphaned `@sk-task` annotations (warning)
- invalid `AC-*` references (warning)

## Configuration file

```yaml
version: 1

project:
  name: my-project
  constitution_file: .speckeep/constitution.md

runtime:
  shell: sh

paths:
  specs_dir: .speckeep/specs
  templates_dir: .speckeep/templates
  scripts_dir: .speckeep/scripts

language:
  default: en
  docs: en
  agent: en
  comments: en

agents:
  update_agents_md: true
  agents_file: AGENTS.md
  targets: []

templates:
  spec: spec.md
  plan: plan.md
  tasks: tasks.md
  data_model: data-model.md
  contracts_api: contracts/api.md
  contracts_events: contracts/events.md
  archive_summary: archive/summary.md
  inspect_report: inspect-report.md
  verify_report: verify-report.md
  constitution: constitution.md
  agents_snippet: agents-snippet.md
  constitution_prompt: prompts/constitution.md
  spec_prompt: prompts/spec.md
  inspect_prompt: prompts/inspect.md
  plan_prompt: prompts/plan.md
  tasks_prompt: prompts/tasks.md
  implement_prompt: prompts/implement.md
  verify_prompt: prompts/verify.md
  archive_prompt: prompts/archive.md
  challenge_prompt: prompts/challenge.md
  handoff_prompt: prompts/handoff.md
  hotfix_prompt: prompts/hotfix.md
  recap_prompt: prompts/recap.md
  scope_prompt: prompts/scope.md

scripts:
  run_speckeep: run-speckeep.sh
  inspect_spec: inspect-spec.sh
  trace: trace.sh
  check_constitution: check-constitution.sh
  check_spec_ready: check-spec-ready.sh
  check_inspect_ready: check-inspect-ready.sh
  check_plan_ready: check-plan-ready.sh
  check_tasks_ready: check-tasks-ready.sh
  check_implement_ready: check-implement-ready.sh
  check_verify_ready: check-verify-ready.sh
  check_archive_ready: check-archive-ready.sh
  verify_task_state: verify-task-state.sh
  list_open_tasks: list-open-tasks.sh
  link_agents: link-agents.sh
  list_specs: list-specs.sh
  show_spec: show-spec.sh
```

`cleanup-agents` removes orphaned agent artifacts for disabled targets.

For PowerShell projects, use `.ps1` extensions instead of `.sh`.
