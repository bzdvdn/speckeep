# Workflow Model

## Strict Phase Chain

```text
constitution -> spec -> inspect -> plan -> tasks -> implement -> verify -> archive
```

For new projects (Greenfield), work begins with an extended **Constitution** phase (using the `--foundation` flag) that codifies both the process rules and the project's technical foundation.

SpecKeep assumes branch-based delivery: each active feature should be developed in its own git branch, with the feature spec and plan package acting as the shared source of truth instead of a mutable global memory file. The default branch naming convention is `feature/<slug>`.

## Phase Responsibilities

### `constitution`

Defines the non-negotiable rules of the project.

For from-scratch projects (Greenfield), the constitution is extended with **Tech Stack** and **Core Architecture** sections, replacing the need for a separate "zero" feature design. This creates a unified, non-archived source of truth for the entire project.

Mandatory sections:

- `Purpose`
- `Core Principles`
- `Constraints`
- `Tech Stack`
- `Core Architecture`
- `Language Policy`
- `Development Workflow`
- `Governance`
- `Last Updated`

After updating the constitution, the agent checks whether any active specs conflict with the changed rules and flags them as `NEEDS RE-INSPECT` without modifying the specs themselves.

### `spec`

Captures one feature request as a concrete spec. Acceptance criteria should use canonical `Given / When / Then` markers even when the surrounding document language is Russian.

For agent-facing `/speckeep.spec`, SpecKeep should support optional arguments:

- `--name <feature name>`
- `--slug <feature-slug>`
- `--branch <branch-name>`

Argument semantics:

- `--name` sets the canonical feature name for the current spec request
- `--slug` overrides the spec slug
- `--branch` overrides only the working branch and does not change the spec slug

`/speckeep.spec` should support two input modes:

- inline mode: the feature name and description are provided in the same message
- staged mode: the user first sends `/speckeep.spec --name ...` and then sends the feature description in the next message

When `/speckeep.spec` starts from a prompt file, SpecKeep should prefer top-of-file metadata such as:

```text
name: Add dark mode
slug: add-dark-mode
```

Priority rules for the slug:

1. `--slug`
2. `slug:`
3. a slug derived from `--name`
4. a slug derived from `name:`
5. a safe fallback from the filename or short request text only when sufficiently specific

Priority rules for the feature name:

1. `--name`
2. `name:`
3. a concise feature name safely derived from the user request

If `/speckeep.spec` is invoked with `--name` but the feature description is still not detailed enough for a valid spec, SpecKeep should not lose the request context: it should ask for the missing description or treat the next user message as the continuation of the same spec request.

By default, the feature branch should be `feature/<slug>`. If the user explicitly provides `--branch <name>`, SpecKeep should use that branch name instead without changing the spec slug.

The spec itself should remain branch-agnostic: the working branch belongs to execution context, not to the spec document.

If the request is ambiguous, combines multiple features, or asks to derive one spec from multiple constitutional changes, SpecKeep should stop and ask for one concrete feature before creating the branch or spec.

### `inspect`

Checks consistency and quality for a single feature. It can flag missing scenarios, weak acceptance criteria, constitutional conflicts, plan drift, or missing task coverage.

`inspect` is mandatory before `plan`. Planning should not proceed until the feature has a persisted inspect report at the canonical path.

`--delta`: incremental re-check mode. After a `spec --amend`, re-checks only the changed sections instead of running a full inspection. Preserves valid findings from the previous report. Falls back to full inspection when changes are too broad (>50% rewritten).

A full inspection report should use a stable structure:

- `# Inspect Report: <slug>`
- `## Scope`
- `## Verdict`
- `## Errors`
- `## Warnings`
- `## Questions`
- `## Suggestions`
- `## Traceability`
- `## Next Step`

`Verdict` should be one of:

- `pass`
- `concerns`
- `blocked`

Suggested semantics:

- `pass`: no blocking problems; only minor or no warnings remain
- `concerns`: the workflow can continue, but warnings or open questions should be resolved soon
- `blocked`: the next workflow step would otherwise proceed on missing or contradictory information

When an inspection report is persisted to disk, SpecKeep should use this canonical path:

- `.speckeep/specs/<slug>/inspect.md`

Use `.speckeep/templates/inspect-report.md` as the canonical template when the report is written to disk.
Persisted inspect and verify reports should start with a machine-readable metadata block containing `report_type`, `slug`, `status`, `docs_language`, and `generated_at`.

Stable acceptance IDs such as `AC-001` make traceability lighter and easier to validate.

SpecKeep should prefer cheap helper findings as the first evidence layer for inspect:

- `check-inspect-ready` and `inspect-spec` should establish structural findings before the agent widens scope
- helper findings may carry categories such as `structure`, `traceability`, `ambiguity`, `consistency`, and `readiness`
- the inspect agent should preserve those findings in the report and only add reasoning where the cheap checks cannot prove the claim directly
- helper findings should not be silently ignored just because the broader narrative sounds acceptable

For cheap `spec <-> plan` consistency checks, SpecKeep should prefer this scope:

- always load: `constitution.md`, `spec.md`
- load if needed: `plan.md`, `tasks.md`
- conditional deeper reads only when a concrete claim requires them: `data-model.md`, `contracts/`, `research.md`
- do not read implementation code by default

The goal is to catch obvious drift, not to run a full architectural review. Useful checks include:

- constitution-to-spec alignment
- goal alignment
- unjustified scope expansion
- acceptance-critical behavior reflected at the plan level
- plan-to-task alignment when `tasks.md` exists
- planned implementation surfaces compared with `tasks.md` `Surface Map` and `Touches:` references when both plan and tasks exist
- constitutional consistency
- justification for richer plan artifacts such as `data-model.md` and `contracts/`

### `plan`

Produces technical design artifacts for one feature package:

- `plan.md`
- `data-model.md`
- `contracts/`
- `research.md` (optional) — used to identify and resolve technical unknowns, architecture trade-offs, or integration constraints before finalizing the implementation plan.

### `tasks`

Turns the plan package into executable tasks. `tasks.md` lives next to other plan artifacts inside `.speckeep/specs/<slug>/plan/`.

SpecKeep uses **Lazy Decomposition** to keep context narrow:

- **Phase `tasks`**: Produces a high-level map (5-10 tasks) tied to functional boundaries. Micro-tasks (1-5 lines of code) are discouraged at this stage to save tokens.
- **Phase `implement`**: The agent can perform **In-place Decomposition** by adding indented sub-tasks (e.g., `T1.1.1`) to the active task only.

Tasks should be grouped by phase and use phase-scoped task IDs such as `T1.1`, `T1.2`, and `T2.1`.

Acceptance coverage should reference those task IDs directly:

```text
AC-001 -> T1.1, T2.1
```

`--repair <task-id-list>`: targeted repair mode. Fixes specific tasks identified by verify or review (e.g. `--repair T2.3,T3.1`) without rewriting the full task list. If the repair reveals a plan-level flaw, suggests `/speckeep.plan --update` instead.

### `implement`

Executes unfinished tasks and updates `tasks.md`.

**In-place Decomposition rules**:

- Sub-tasks MUST NOT add new files to the `Touches:` list of the parent task.
- Sub-tasks MUST NOT change the `AC-*` mapping of the parent task.
- If decomposition reveals a plan-level flaw, the agent MUST stop and request a plan update.

Default behavior should remain full-run: without explicit scope flags, SpecKeep continues through all unfinished tasks in task-list order.

Selective execution is allowed when the user explicitly narrows scope:

- `--phase <number>` for one implementation phase
- `--tasks <task-id-list>` for one or more specific task IDs such as `T1.1,T2.1`

`--continue`: resume mode. Starts from the first unfinished task, trusting that all previously checked-off tasks are correctly completed. Batch-reads only the surfaces from remaining unfinished tasks. Useful after session interruptions (timeout, context overflow).

`--phase` and `--tasks` should not be combined in the same run.

When selective execution skips unfinished earlier work, SpecKeep should warn about the sequencing risk without silently broadening scope.

During implementation, SpecKeep should emit short runtime progress updates whenever it starts or completes a phase in the active execution scope.

Every non-trivial code change should include a **traceability annotation** linking it back to the task ID and the primary acceptance criterion (AC).
Format: `// @sk-task <TASK_ID>: <Description> (<AC_ID>)`

Those phase-status updates should follow the project's configured agent language rather than defaulting to English.

### `verify`

Runs a lightweight post-implementation check to confirm that completed work is aligned enough with tasks and project rules to move forward safely.

Verification uses **traceability data** collected via `/.speckeep/scripts/trace.* <slug>` to confirm that implementation matches task claims and acceptance criteria (including `@sk-test` annotations).

**Legacy Fallback**: For features without annotations, the agent falls back to manual inspection of the files listed in `Touches:` and running relevant tests. This ensures SpecKeep remains compatible with older features while encouraging token-efficient verification for new ones.

A full verification report should use a stable structure:

- `# Verify Report: <slug>`
- `## Scope`
- `## Verdict`
- `## Checks`
- `## Errors`
- `## Warnings`
- `## Questions`
- `## Not Verified`
- `## Next Step`

Recommended report details:

- `## Scope` should record the actual verification mode such as `default` or `deep`
- `## Scope` should list the concrete surfaces that were inspected
- `## Verdict` should include `archive_readiness`
- `## Verdict` should include a one-line summary of why the verdict is justified
- `## Checks` should include `task_state`
- `## Checks` should include `acceptance_evidence` for the `AC-*` items actually confirmed
- `## Checks` should include `implementation_alignment` tied to the concrete surface inspected
- `## Not Verified` should list material claims or surfaces that were intentionally not checked

`Verdict` should be one of:

- `pass`
- `concerns`
- `blocked`

Suggested semantics:

- `pass`: no blocking problems are present; only minor or no warnings remain
- `concerns`: the feature can move forward, but warnings or open questions should be resolved soon
- `blocked`: archive or completion claims would otherwise proceed on contradictory implementation state or unfinished required work

Use `concerns` rather than `pass` when the evidence is partial but no concrete contradiction has been found.

`--persist`: write the report to `.speckeep/specs/<slug>/plan/verify.md` in addition to conversation output. Without this flag, the report stays in the conversation only.

Use `.speckeep/templates/verify-report.md` as the canonical template when the report is written to disk.

Persisted verify reports should start with the same machine-readable metadata block used by inspect reports: `report_type`, `slug`, `status`, `docs_language`, and `generated_at`.

When available, SpecKeep should prefer `.speckeep/scripts/check-verify-ready.sh <slug>` as the cheap readiness pass before deeper verification.

Use `.speckeep/scripts/verify-task-state.sh <slug>` as the cheapest first-pass helper when you only need task-state confirmation.

Note: generated `.speckeep/scripts/*` wrappers compute the project root from the script location and pass it via `--root`, so they can be executed from any working directory.

### `archive`

Copies a completed, superseded, rejected, abandoned, or deferred feature package into `.speckeep/archive/<slug>/<YYYY-MM-DD>/`.

The archive script validates verify status and open tasks internally — it returns a clear error if prerequisites are not met. Default archive status is `completed`; other statuses (`superseded`, `abandoned`, `rejected`, `deferred`) require an explicit `--reason`.

## Why This Chain Exists

The chain keeps the product strict without becoming bureaucratic:

- architecture and workflow rules come first
- user intent becomes a spec before technical planning starts
- technical planning happens before task breakdown
- implementation follows tasks instead of improvisation
- lightweight verification closes the gap between implementation and archive
- completed feature packages can be archived without bloating the active workspace
