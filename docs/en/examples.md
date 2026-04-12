# Examples

This page shows realistic end-to-end SpecKeep scenarios for one feature package.

## Quick Usage Patterns

### New Project

When starting a greenfield project, SpecKeep works best as a minimal project-context layer from day one.

Example:

```bash
speckeep init my-project --lang en --shell sh --agents codex
cd my-project
speckeep doctor .
```

What to do next:

- establish the `constitution` for project rules
- describe the first feature through `spec`
- prepare `plan` and `tasks`
- use `implement` only from the current task list

Why this helps:

- humans and agents start from the same rules
- project context stays explicit and editable from the beginning
- the workflow stays lightweight because SpecKeep does not require a heavy process engine

### Existing Project

For a brownfield codebase, SpecKeep should be adopted incrementally instead of trying to document the whole repository at once.

Example:

```bash
cd existing-project
speckeep init . --lang en --shell sh --agents codex
speckeep doctor .
```

Recommended starting point:

- establish the `constitution` around the project's current reality
- pick one active feature or change request
- create a spec only for that scope
- move to plan, tasks, and implement only within that feature package

What not to do:

- do not try to spec the whole project at once
- do not pull broad repository context unless the active feature really needs it

Why this helps:

- SpecKeep adds a lightweight layer of discipline on top of an existing codebase
- adoption happens one feature at a time
- this keeps token usage down and avoids process bloat

### Prompt File Input

When `/speckeep.spec` starts from a local prompt file, prefer explicit metadata instead of relying on a generic filename such as `spec_prompt.md`.

Example prompt file:

```text
name: Add dark mode
slug: add-dark-mode

Add a user-selectable dark theme for the dashboard and settings pages.
```

This lets SpecKeep:

- derive a safe spec path such as `.speckeep/specs/add-dark-mode/spec.md`
- create or switch to `feature/add-dark-mode`
- avoid ambiguous slugs from generic filenames

### Staged Input Via `--name`

When the feature name is already clear but the detailed description is easier to send in the next message, `/speckeep.spec` can start in staged mode.

Example:

```text
/speckeep.spec --name "Dependency Dashboard"
```

Next message:

```text
Build a dashboard for monitoring microservice dependencies with a dark theme, filters, a dependency graph, summary cards, and auto-refresh.
```

This allows SpecKeep to:

- lock in the canonical feature name up front
- derive a safe slug such as `dependency-dashboard`
- preserve the spec request context across messages

If you need an explicit slug:

```text
/speckeep.spec --name "Dependency Dashboard" --slug frontend-layout-rework
```

If you need a repository-specific branch override:

```text
/speckeep.spec --name "Dependency Dashboard" --slug frontend-layout-rework --branch FEAT-142
```

## 1. Create a Constitution for a Brownfield Project

User request:

```text
/speckeep.constitution Python project, DDD style, split into API and workers, Kafka for asynchronous integration, ClickHouse as the analytical sink.
```

Expected agent behavior:

- read the constitution prompt in `.speckeep/templates/prompts/constitution.md`
- inspect only the minimum repository evidence needed
- create or patch `.speckeep/constitution.md`
- run `check-constitution.sh` when appropriate

Expected outcome:

- architecture rules are formalized
- development workflow rules become explicit
- the constitution becomes authoritative for later phases

## 2. Create a Spec

User request:

```text
/speckeep.spec Add partner-specific ingestion scheduling with retry policy overrides.
```

Expected agent behavior:

- read constitution first
- create `.speckeep/specs/partner-scheduling/spec.md`
- write acceptance criteria using canonical `Given / When / Then`
- keep surrounding text in the configured documentation language

Example acceptance criterion:

```md
### Acceptance Criterion 1

- ID: AC-001
- **Given** a partner with a custom retry policy
- **When** the ingestion schedule is evaluated
- **Then** the worker uses the partner-specific retry window instead of the default policy
```

Example with an explicit branch override:

```text
/speckeep.spec Add partner-specific ingestion scheduling with retry policy overrides --branch NRD-11
```

In that case, the spec slug can still stay `partner-scheduling` while the working branch follows the repository's branch convention, for example `NRD-11`.

## 3. Inspect the Spec

User request:

```text
/speckeep.inspect partner-scheduling
```

Expected agent behavior:

- read constitution and `.speckeep/specs/partner-scheduling/spec.md`
- keep the default inspect scope cheap: prefer `constitution.md` and `spec.md`, then pull `plan.md` or `tasks.md` only when they exist and materially affect the finding
- check completeness, constitutional consistency, and scenario quality
- create a focused inspection report
- use `.speckeep/scripts/inspect-spec.sh` or `.speckeep/scripts/inspect-spec.ps1` as a cheap first-pass helper when structural spec or coverage issues need quick confirmation
- persist the inspect report at `.speckeep/specs/partner-scheduling/inspect.md`
- use `.speckeep/templates/inspect-report.md` as the canonical report template

Typical findings:

- missing failure-path scenario
- unclear acceptance coverage for manual retry overrides
- open question about scheduler ownership

## 4. Create a Plan Package

User request:

```text
/speckeep.plan partner-scheduling
```

Expected agent behavior:

- read constitution and the spec
- create `.speckeep/specs/partner-scheduling/plan/plan.md`
- create `.speckeep/specs/partner-scheduling/plan/data-model.md`
- create `.speckeep/specs/partner-scheduling/plan/contracts/`
- create `research.md` only if uncertainty is real

Typical outputs:

- plan for scheduler integration points
- data model for partner overrides and retry windows
- event or API contracts for configuration updates

## 5. Create Tasks

User request:

```text
/speckeep.tasks partner-scheduling
```

Expected agent behavior:

- use `plan.md` as the decomposition entrypoint
- pull in spec, contracts, or data model only when needed
- produce `.speckeep/specs/partner-scheduling/plan/tasks.md`
- include acceptance-to-task coverage

Example task structure:

```md
## Phase 1: Data Model

- [ ] T1.1 Add partner scheduling override model — override fields are persisted
- [ ] T1.2 Persist retry window fields — retry windows are available to scheduling logic

## Acceptance Coverage

- AC-001 -> T1.1, T1.2
```

## 6. Implement the Feature

User request:

```text
/speckeep.implement partner-scheduling
```

Expected agent behavior:

- read `tasks.md` and use it as the execution manifest
- perform **In-place Decomposition** if a task is too complex, adding indented sub-tasks (e.g., `T1.1.1`)
- annotate every non-trivial code change with `// @sk-task <ID> (<AC_ID>)`
- mark completed tasks in `tasks.md`
- stay within the `Touches:` list defined for each task

Example code annotation:

```go
// @sk-task T1.1: Add partner scheduling override model (AC-001)
func SavePartnerSchedule(p Partner) {
    // ...
}
```

## 7. Verify the Implementation

User request:

```text
/speckeep.verify partner-scheduling
```

Expected agent behavior:

- use `/.speckeep/scripts/trace.sh partner-scheduling` to collect implementation evidence
- look for `// @sk-task` and `// @sk-test` annotations in the code
- confirm that implementation matches task descriptions and acceptance criteria
- provide a clear verdict (`pass`, `concerns`, or `blocked`)
- include concrete evidence in the `## Checks` section

Example test annotation:

```go
// @sk-test T1.1: TestSavePartnerSchedule (AC-001)
func TestSavePartnerSchedule(t *testing.T) {
    // ...
}
```

If the feature is older and lacks annotations, the agent falls back to manual inspection of the files listed in `Touches:` and running tests manually.

Expected agent behavior:

- start from `tasks.md`
- load spec, plan, data model, or contracts only for the active task
- implement unfinished tasks in order
- report phase progress as it moves through the selected work
- update `tasks.md`

This phase should avoid broad repository reads unless the active task actually requires them.

Example scoped requests:

```text
/speckeep.implement partner-scheduling --phase 2
/speckeep.implement partner-scheduling --tasks T1.1,T2.1
```

Expected scoped behavior:

- keep the default full-run behavior only when no scope flag is provided
- execute only the selected phase or task IDs when scope is explicitly narrowed
- preserve task order from `tasks.md`
- warn if selected work skips unfinished earlier phases or tasks

Typical runtime updates:

- `Starting Phase 1: Data Model`
- `Phase 1 complete: T1.1, T1.2`
- `Next: Phase 2: Scheduler Logic`

## 7. Verify the Feature

User request:

```text
/speckeep.verify partner-scheduling
```

Expected agent behavior:

- read constitution and tasks first
- confirm that completed tasks match the current implementation state closely enough
- produce a lightweight verification report
- start with `.speckeep/scripts/verify-task-state.sh partner-scheduling` when task-state confirmation is enough
- use `.speckeep/templates/verify-report.md` when the report should be persisted
- default to `.speckeep/specs/partner-scheduling/plan/verify.md` when no explicit path is provided

## 8. Archive the Feature

User request:

```text
/speckeep.archive partner-scheduling
```

Expected agent behavior:

- for `completed` status, start with `.speckeep/scripts/verify-task-state.sh partner-scheduling` and stop if open tasks remain
- copy the feature package into `.speckeep/archive/partner-scheduling/<YYYY-MM-DD>/`
- write `summary.md`

Expected archive result:

```text
.speckeep/archive/
  partner-scheduling/
    2026-03-28/
      summary.md
      spec.md
      plan.md
      tasks.md
      data-model.md
      contracts/
```

## 9. Agent Maintenance Scenario

A practical maintenance flow for agent targets:

```bash
speckeep add-agent my-project --agents claude --agents cursor
speckeep list-agents my-project
speckeep remove-agent my-project --agents cursor
speckeep cleanup-agents my-project
speckeep doctor my-project
```

Use this when a project changes its preferred agent mix over time.
