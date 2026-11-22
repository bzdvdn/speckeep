# Workflow Model

## Strict Phase Chain

```text
constitution -> spec -> [inspect, optional] -> [plan, optional] -> tasks -> implement -> archive
```

`verify` is an optional on-demand audit: it is always available but skipped by default in the normal workflow.

Two lighter lanes keep features closing fast without artifact bloat:

- **Express propose** (`/spk-propose`): a small or medium, low-risk idea → `spec.md` + `tasks.md` in one pass, straight to `implement`, plus `data-model.md` or `contracts/api.md` when the change really touches them. It never writes `plan.md`. It escalates only on hard signals — the request spans several features, a breaking change or schema migration with backfill, cross-service coordination or a security boundary, ACs that cannot be written without inventing decisions, or spec overflow. Uncertainty is recorded in `## Open Questions` and stays in propose.
- **Converge loop** (`/spk-converge` / `speckeep converge <slug>`): after `implement`, re-checks task/proof/surface/AC coverage cheaply and reports gaps. The agent appends follow-up tasks and repeats until `converged` (hard stop after 2 fix rounds). Lighter than `verify` — no report file.

`speckeep guard .` is the machine gate variant: exit 0 only when every active feature is archive-ready. Features living in an OpenSpec or Spec Kit workspace can be brought in with `speckeep import openspec|speckit .`.

For new projects (Greenfield), work begins with an extended **Constitution** phase (using the `--foundation` flag) that codifies both the process rules and the project's technical foundation.

SpecKeep assumes branch-based delivery: each active feature should be developed in its own git branch, with the feature spec and artifact set acting as the shared source of truth instead of a mutable global memory file. The default branch naming convention is `feature/<slug>`.

### Starting the next feature

After a feature is archived and merged, start the next one in a fresh session. Use `/spk-propose` for a small, low-risk idea (`spec.md` + `tasks.md` in one pass, then `/spk-implement`) or `/spk-spec` when the idea is ambiguous, multi-feature, or design-heavy. A natural-language request such as "build X with SpecKeep" routes to `/spk-propose`; name `/spk-spec` explicitly to override. `plan` is opt-in and always comes after spec: `/spk-plan` only for design-heavy work (more than one realistic option with a real trade-off, data-model/contract changes, migration/rollout, or cross-boundary impact), and `/spk-propose` never writes it. A new feature legitimately re-loads its phase skills; within one session, never re-run a phase you are already in.

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

For agent-facing `/spk-spec`, SpecKeep should support optional arguments:

- `--name <feature name>`
- `--slug <feature-slug>`
- `--branch <branch-name>`

Argument semantics:

- `--name` sets the canonical feature name for the current spec request
- `--slug` overrides the spec slug
- `--branch` overrides only the working branch and does not change the spec slug

`/spk-spec` should support two input modes:

- inline mode: the feature name and description are provided in the same message
- staged mode: the user first sends `/spk-spec --name ...` and then sends the feature description in the next message

When `/spk-spec` starts from a prompt file, SpecKeep should prefer top-of-file metadata such as:

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

If `/spk-spec` is invoked with `--name` but the feature description is still not detailed enough for a valid spec, SpecKeep should not lose the request context: it should ask for the missing description or treat the next user message as the continuation of the same spec request.

By default, the feature branch should be `feature/<slug>`. If the user explicitly provides `--branch <name>`, SpecKeep should use that branch name instead without changing the spec slug.

The spec itself should remain branch-agnostic: the working branch belongs to execution context, not to the spec document.

If the request is ambiguous, combines multiple features, or asks to derive one spec from multiple constitutional changes, SpecKeep should stop and ask for one concrete feature before creating the branch or spec.

### `inspect`

Checks consistency and quality for a single feature. It can flag missing scenarios, weak acceptance criteria, constitutional conflicts, plan drift, or missing task coverage.

`inspect` is optional before `tasks` (or before `plan` when the feature needs design). Tasks may proceed without an inspect report when the spec is clear and low-risk.
If an inspect report is present, it must have a valid non-blocking status (`pass` or `concerns`) to proceed.

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

- `specs/active/<slug>/inspect.md`

Use `.speckeep/templates/inspect-report.md` as the canonical template when the report is written to disk.
Persisted inspect and verify reports should start with a machine-readable metadata block containing `report_type`, `slug`, `status`, `docs_language`, and `generated_at`.

Stable acceptance IDs such as `AC-001` make traceability lighter and easier to validate.

SpecKeep should prefer cheap helper findings as the first evidence layer for inspect:

- `check-inspect-ready` and `inspect-spec` should establish structural findings before the agent widens scope
- helper findings may carry categories such as `structure`, `traceability`, `ambiguity`, `consistency`, and `readiness`
- the inspect agent should preserve those findings in the report and only add reasoning where the cheap checks cannot prove the claim directly
- helper findings should not be silently ignored just because the broader narrative sounds acceptable

For cheap `spec <-> plan` consistency checks, SpecKeep should prefer this scope:

- always load: `CONSTITUTION.md`, `spec.md`
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

Optional/opt-in: use plan only when the feature needs design work (multiple realistic options, data-model/contracts changes, migration/rollout, or cross-boundary impact). For small/low-risk changes, skip to `tasks`.

Produces technical design artifacts for one feature package:

- `plan.md`
- `data-model.md`
- `contracts/`
- `research.md` (optional) — used to identify and resolve technical unknowns, architecture trade-offs, or integration constraints before finalizing the implementation plan.

### `tasks`

Turns the plan (when present) or the spec into executable tasks. `tasks.md` lives next to other feature artifacts inside `specs/active/<slug>/`.

SpecKeep uses **Lazy Decomposition** to keep context narrow:

- **Phase `tasks`**: Produces a high-level map (5-10 tasks) tied to functional boundaries. Micro-tasks (1-5 lines of code) are discouraged at this stage to save tokens.
- **Phase `implement`**: The agent can perform **In-place Decomposition** by adding indented sub-tasks (e.g., `T1.1.1`) to the active task only.

Tasks should be grouped by phase and use phase-scoped task IDs such as `T1.1`, `T1.2`, and `T2.1`.

Acceptance coverage should reference those task IDs directly:

```text
AC-001 -> T1.1, T2.1
```

`--repair <task-id-list>`: targeted repair mode. Fixes specific tasks identified by verify or review (e.g. `--repair T2.3,T3.1`) without rewriting the full task list. If the repair reveals a plan-level flaw, suggests `/spk-plan --update` instead.

### `implement`

Executes unfinished tasks and updates `tasks.md`.

**In-place Decomposition rules**:

- Sub-tasks MUST NOT add new files to the `Touches:` list of the parent task.
- Sub-tasks MUST NOT change the `AC-*` mapping of the parent task.
- If decomposition reveals a plan-level flaw, the agent MUST stop and request a plan update.

**Spec evolution** is allowed and expected: implementation often reveals a missing requirement or a wrong assumption. In that case update `spec.md` and `tasks.md` together — new or amended `AC-*` plus the task that covers it — and continue. Never leave the spec ahead of the tasks. Anything larger (new scope, design change) is still a stop: return to `/spk-tasks` or `/spk-plan`. Structural drift stays deterministic: `converge` and `check` report `acceptance_not_covered` and `unknown_acceptance_reference` when the AC set and task coverage disagree.

Default behavior should remain full-run: without explicit scope flags, SpecKeep continues through all unfinished tasks in task-list order.

Selective execution is allowed when the user explicitly narrows scope:

- `--phase <number>` for one implementation phase
- `--tasks <task-id-list>` for one or more specific task IDs such as `T1.1,T2.1`

`--continue`: resume mode. Starts from the first unfinished task, trusting that all previously checked-off tasks are correctly completed. Batch-reads only the surfaces from remaining unfinished tasks. Useful after session interruptions (timeout, context overflow).

`--phase` and `--tasks` should not be combined in the same run.

When selective execution skips unfinished earlier work, SpecKeep should warn about the sequencing risk without silently broadening scope.

During implementation, SpecKeep should emit short runtime progress updates whenever it starts or completes a phase in the active execution scope.

Evidence for completed work is recorded as `Proof:` entries in `tasks.md` — that is what `speckeep trace`, `speckeep doctor`, and archive gates read; there are no trace markers in source code.

For every completed task (`[x]`) in `tasks.md`, write a `Proof:` line directly under the checked task:

- format: `Proof: <kind> <path> [<anchor>]`, where `kind` is `code|test|docs|chore`, `path` is a repo-root-relative path, and `anchor` is the owning function/test/type name (optional but recommended)
- examples: `Proof: code src/handlers/export.go ExportHandler`, `Proof: test src/tests/export_test.go TestExportFlow`, `Proof: docs docs/export.md`
- a `[x]` task without at least one `Proof:` entry is NOT done; `speckeep check` and a `completed` archive block it
- `Proof:` entries MUST point to existing files (missing file is a hardened error); anchors SHOULD resolve to owning symbols (missing anchor is a warning)

Change attribution is separate from proof: `Touches:` records what a task plans to change, while `Proof:` records the evidence. Right after creating `feature/<slug>`, pin the base commit with `speckeep trace init <slug>`; `speckeep trace <slug> --files` then shows the actual diff mapped to tasks, and `speckeep trace --file <path>` answers "which spec changed this file" across active and archived features. Archive stores the mapping in `changes.json`.

Those phase-status updates should follow the project's configured agent language rather than defaulting to English.

### `verify`

Optional on-demand audit: verify runs a lightweight post-implementation check to confirm that completed work is aligned enough with tasks and project rules to move forward safely. It is always available but skipped by default — the normal workflow proceeds from `implement` straight to `archive`.

Verification uses **traceability data** collected via `./.speckeep/scripts/trace.* <slug>`, which reads `Proof:` lines from `tasks.md` to confirm that implementation matches task claims and acceptance criteria. The verify phase never scans source code for annotations; it reads the touched diff only for the standards review.

A full verification report should use a stable structure:

- `# Verify Report: <slug>`
- `## Scope`
- `## Verdict`
- `## Checks`
- `## Errors`
- `## Warnings`
- `## Review`
- `## Questions`
- `## Not Verified`
- `## Next Step`

`verify` covers two axes:

- **Spec coverage** — every `AC-*` has evidence and every completed task has a valid `Proof:`.
- **Standards** — the implemented diff is reviewed for duplicated logic, dead code, misleading names, unhandled errors, weakened tests, and changes that bypass existing abstractions. Review is scoped to the files this feature touched (`changes.json` / `Touches:` plus `Unmapped` drift).

Review findings are capped at 10 and recorded as `blocker|major|minor — path:line — why`. Findings are reported, not fixed. Only `blocker` forces the verdict to `blocked`; `major`/`minor` do not block `pass`. If no diff is available, the section says `skipped (no diff)`.

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

`--persist`: write the report to `specs/active/<slug>/verify.md` in addition to conversation output. Without this flag, the report stays in the conversation only.

Use `.speckeep/templates/verify.md` as the canonical template when the report is written to disk.

Persisted verify reports should start with the same machine-readable metadata block used by inspect reports: `report_type`, `slug`, `status`, `docs_language`, and `generated_at`.

When available, SpecKeep should prefer `./.speckeep/scripts/check-ready.sh verify <slug>` as the cheap readiness pass before deeper verification.

Use `.speckeep/scripts/verify-task-state.sh <slug>` as the cheapest first-pass helper when you only need task-state confirmation.

Note: generated `.speckeep/scripts/*` wrappers compute the project root from the script location and pass it via `--root`, so they can be executed from any working directory.

### `archive`

Copies a completed, superseded, rejected, abandoned, or deferred feature package into `specs/archived/<slug>/<YYYY-MM-DD>/`.

Archive is CLI-only. For `completed`, the feature must be deterministically proven — no open tasks, every `[x]` task has a valid `Proof:`, and the verify policy is satisfied; a `verify.md` with status other than `pass` vetoes archiving. `superseded`, `abandoned`, `rejected`, and `deferred` skip those gates and may archive incomplete work (they still require an explicit `--reason`). Default archive status is `completed`. Legacy `completed` features without `Proof:` must add evidence or be archived with a non-completed status and a reason.

`--compact` archives lean: it stores only `summary.md` plus a `snapshot.sha` git pointer (branch + commit + file list) instead of copying every artifact — the full history already lives in git. Restore (`--restore`) re-creates the files via `git show`. Compact requires committed artifacts and a git repo.

### `finish`

Local-merge close for a completed feature: `speckeep finish <slug> .` archives the feature, commits the archive snapshot on the feature branch, merges the branch into the base branch with `--no-ff`, and deletes the local feature branch. It requires the feature/hotfix branch checked out and a clean worktree; the target defaults to `main`, then `master` (`--into <branch>` overrides, e.g. `--into dev`; the branch must exist locally). It never pushes and never deletes remote branches. Use `archive` instead when merging via a pull request, or for deferred/abandoned features.

## Optional Commands

These commands sit beside the phase chain. None of them is a gate and none is required for a normal feature.

| Command | Use it when | Artifact |
|---|---|---|
| `/spk-grill` | the idea is raw and the design tree is unresolved | none |
| `/spk-research` | a technical unknown must be resolved from primary sources | `research.md` |
| `/spk-diagnose` | a bug's cause is unknown | none (fix + regression test) |
| `/spk-hotfix` | the cause is known and the fix is small | `hotfix.md` |
| `/spk-challenge` | a spec or plan needs an adversarial pass | none |
| `/spk-scope` | you need an in/out boundary check | none |
| `/spk-resume` | a fresh session must pick up the feature quickly | none |
| `/spk-handoff` | the state must be shared with a person or another session | handoff summary |
| `/spk-recap` | you want a project-wide status board | none |
| `/spk-converge` | you want the cheap implement → check → fix loop instead of full verify | none |
| `/spk-rollback` | completed tasks must return to unfinished | `tasks.md` |

`/spk-grill` and `/spk-research` end by pointing at `/spk-propose` or `/spk-spec`; they never write a spec themselves. `/spk-resume` writes nothing and is deliberately thinner than `/spk-handoff`.

## Closing a Feature

| Situation | Use |
|---|---|
| merge through a pull request | `speckeep archive <slug> .`, then push the branch and open a PR |
| merge locally, no PR | `speckeep finish <slug> .` (optionally `--into <branch>`) |
| park unfinished work | `speckeep archive <slug> . --status deferred --reason "<why>"` |
| supersede or abandon | `speckeep archive <slug> . --status superseded --reason "<why>"` (or `abandoned`) |

## Why This Chain Exists

The chain keeps the product strict without becoming bureaucratic:

- architecture and workflow rules come first
- user intent becomes a spec before technical planning starts
- technical planning happens before task breakdown
- implementation follows tasks instead of improvisation
- deterministic `Proof:` evidence (and optional verification when a full audit is desired) confirms readiness before archive
- completed feature packages can be archived without bloating the active workspace
