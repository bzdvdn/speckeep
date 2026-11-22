# Glossary

## Constitution

The highest-priority project document. It defines non-negotiable rules for architecture, workflow, language policy, and governance. For new projects, it also serves as the technical foundation by codifying the tech stack and core architecture.

## Spec

A feature-level document that describes what should be built and why.

## Inspect

A review phase that checks a spec and related artifacts for completeness, consistency, and constitutional compliance.

## Plan

An opt-in technical design phase that translates one spec into implementation-oriented artifacts. Used only when the feature needs design work; small/low-risk features go straight to tasks.

## Propose

The default new-feature lane: `spec.md` + `tasks.md` in one pass, plus `data-model.md` or `contracts/*` when the change really touches them. It never writes `plan.md`; uncertainty stays in `## Open Questions`, and only hard signals (several features, a breaking change or schema migration, cross-service work) escalate to `/spk-spec`.

## Feature Artifact Set

The group of files stored under `specs/active/<slug>/`, typically including `spec.md`, `plan.md`, `tasks.md`, `data-model.md`, `verify.md`, `contracts/`, and optional `research.md`.

## Tasks

The executable breakdown of a feature's spec (and plan when present). In SpecKeep, `tasks.md` lives inside the feature artifact set, next to `plan.md` when that artifact exists.

## Implement

The execution phase that works from unfinished tasks and updates task state.

## Verify

A lightweight post-implementation check that confirms whether a feature is aligned enough with tasks, specs, plan artifacts, and project rules to proceed safely toward archive or further completion claims.

## Archive

The phase and storage area used to preserve a historical snapshot of a finished or inactive feature package.

## Finish

The local-merge close: archive + commit the snapshot + `merge --no-ff` into the base branch + delete the local feature branch. Never pushes or deletes remote branches.

## Archive Status

The reason recorded on an archived package: `completed`, `superseded`, `abandoned`, `rejected`, or `deferred`. Only `completed` enforces closed tasks, `Proof:`, and the verify policy; the others park incomplete work and require `--reason`.

## Proof

The evidence line under a completed task in `tasks.md`: `Proof: <kind> <path> [<anchor>]`. `trace`, `check`, `doctor`, and the archive gate read only these lines.

## Converge

A cheap closing loop that re-checks completed tasks, `Proof:`, declared surfaces, and AC coverage, then reports gaps so follow-up tasks can be appended and the check re-run. Lighter than `verify` — it produces no report file.

## Grill

An optional pre-spec interview that resolves a design tree in rounds with a recommended answer per question. Writes no artifacts by default.

## Diagnose

An optional bounded bug loop: reproduce → minimize → hypothesize → instrument → fix → regression test. Use hotfix instead when the cause is already known.

## Resume

A thin in-session handoff: a short state packet (phase, tasks, next open task, blockers) that creates no artifact. Use handoff when the state must be shared with a person.

## Handoff

A shareable summary of the current feature state so another session or teammate can continue without re-reading the artifact set.

## Hotfix

An emergency fix lane outside the phase chain for a known cause and a small, bounded diff.

## Data Model

A plan artifact that describes entities, structures, relationships, and important invariants needed for implementation.

## Contracts

Plan artifacts that define interfaces such as APIs, events, or other external interaction boundaries.

## Research

An optional plan artifact used only when uncertainty, external investigation, or architecture tradeoffs need to be preserved.

## Agent Target

A supported agent ecosystem for which SpecKeep can generate project-local command or prompt files, such as `claude`, `codex`, `copilot`, `cursor`, `kilocode`, `opencode`, or `trae`.

## Orphaned Agent Artifact

A generated agent file that still exists on disk even though its target is no longer enabled in `.speckeep/speckeep.yaml`.

## Docs Language

The configured language used for generated project documentation such as `constitution.md`, specs, plans, and tasks.

## Agent Language

The configured language used for generated prompts and `AGENTS.md` guidance.

## Comments Language

The configured preferred language for new or edited code comments during implementation.

## Given / When / Then

The canonical BDD markers used for acceptance criteria regardless of documentation language. The surrounding text may still follow the configured docs language.

## Acceptance ID

A stable identifier for an acceptance criterion, such as `AC-001`. It helps keep traceability explicit across specs, tasks, and inspection reports.

## Traceability

The practice of linking implementation evidence (files and tests) back to tasks and acceptance criteria using stable IDs. In SpecKeep, this is achieved via `Proof:` entries recorded in `tasks.md` under each completed task — evidence lives there, not in source-code markers.

## Lazy Decomposition

A strategy to minimize context bloat by keeping initial tasks high-level and allowing the implementation agent to perform **In-place Decomposition** only for the active task.

## In-place Decomposition

The act of refining a complex task by adding indented sub-tasks (e.g., `T1.1.1`) directly in `tasks.md` during the implementation phase.

## Smart Branching

A feature that ensures Git development alignment by checking if the current branch matches the expected `feature/<slug>` pattern.
