# SpecKeep Plan Prompt

You are creating or updating the implementation plan package for one feature.

## Goal

Produce plan artifacts under `<specs_dir>/<slug>/plan/` (default: `.speckeep/specs/<slug>/plan/`).

## Phase Contract

Inputs: `.speckeep/constitution.md`, `.speckeep/specs/<slug>/spec.md`, `.speckeep/specs/<slug>/inspect.md`, narrow repo code.
Outputs: `.speckeep/specs/<slug>/plan/plan.md`, `.speckeep/specs/<slug>/plan/data-model.md`; optional `contracts/`, `research.md`.
Stop if: spec or inspect missing, spec too vague for architecture decisions, or constitutional conflict.

If `.speckeep/speckeep.yaml` overrides `paths.specs_dir` or `project.constitution_file`, follow the configured paths. Read it at most once per session.

## Flags

`--update`: targeted edit mode — change one section, one `DEC-*`, one surface, or add one contract without rewriting the plan package.

- Read existing plan artifacts first; change only what was requested.
- If a `DEC-*` changes, refresh `Affects` and `Validation` if they go stale.
- Keep `plan.md`, `data-model.md`, and `contracts/` internally consistent.
- Do not invalidate `tasks.md` unless the change materially affects task decomposition or acceptance coverage.

`--research`: research-first mode.

- Identify the 1–5 concrete unknowns that block planning, write them to `.speckeep/specs/<slug>/plan/research.md`.
- Stop after `research.md` and ask: "Research complete — proceed to full plan?"
- Produce `plan.md` / `data-model.md` only after explicit confirmation. If the user declines, end the session with `research.md` as the only new artifact.

Do not accept `--update` and `--research` together in the same run. If both are present, stop and ask which mode to use.

`--greenfield`: greenfield-first planning mode.

- Prefer this when the repository does not yet provide meaningful implementation surfaces for the feature.
- In this mode, keep the plan anchored around the first deployable slice, bootstrapping surfaces, and the shortest validation path to MVP.

## Load and Scope

Load: `.speckeep/constitution.md`, `.speckeep/specs/<slug>/spec.md`, `.speckeep/specs/<slug>/inspect.md`, and only the repo code needed to identify concrete implementation surfaces, boundaries, and constraints.
Do not load by default: large unrelated repository areas, other features' artifacts, or optional `research.md` unless uncertainty already exists.
Do not read `/.speckeep/scripts/*` by default — use the readiness wrapper unless you are debugging scripts, working on SpecKeep itself, or the user asked.

When `/.speckeep/scripts/check-plan-ready.*` is available, prefer running it as the phase readiness check instead of reading script source. The readiness wrapper runs with the slug as the first argument. Example: `bash ./.speckeep/scripts/check-plan-ready.sh <slug>` (or PowerShell: `.\.speckeep\scripts\check-plan-ready.ps1 <slug>`).

Do not create or switch branches. The feature branch must already exist from the spec phase. If you are not on the expected feature branch, stop and report — do not create it.

Do not create `tasks.md`. Task decomposition is a separate phase triggered by `/speckeep.tasks`. Stop after writing plan artifacts and output the next-command line.

## Stop Conditions

Stop and ask for clarification or refinement if:

- `.speckeep/specs/<slug>/spec.md` or `.speckeep/specs/<slug>/inspect.md` is missing — do not attempt to create them during `plan`; instead instruct the user to run `/speckeep.spec <slug>`, then `/speckeep.inspect <slug>`, then re-run `/speckeep.plan <slug>`
- the spec is too vague to produce architecture, contracts, or data model decisions
- constitutional constraints conflict with the intended plan
- the plan would need to cross an unclear integration or architectural boundary that is not justified by the spec or focused repository evidence
- the work would only make sense if multiple feature packages were planned together

Do not compensate by reading broad unrelated repository context.

## Required Outputs

Always create or update:

- `.speckeep/specs/<slug>/plan/plan.md`
- `.speckeep/specs/<slug>/plan/data-model.md` — always. When the feature does not introduce or modify persisted state, entity shape, lifecycle, or contract-relevant payload shape, write a compact no-change stub instead of omitting the file.

Create only when justified:

- `.speckeep/specs/<slug>/plan/contracts/api.md` — feature touches an API boundary
- `.speckeep/specs/<slug>/plan/contracts/events.md` — feature produces or consumes events
- `.speckeep/specs/<slug>/plan/quickstart.md` — `--greenfield` mode or early-feature planning benefits from a short MVP validation flow

Create `.speckeep/specs/<slug>/plan/research.md` only when at least one of these is true:

- the feature depends on an external system, API, or dependency with behavior that is still unclear
- there are multiple realistic implementation options with meaningful trade-offs that must be preserved
- there is a non-obvious performance, security, reliability, or integration risk that affects planning
- a repository constraint or architectural boundary must be investigated before the plan can be made concrete

Before creating `research.md`, write down the concrete unknowns first:

- list only the 1–5 specific unknowns that block planning
- tie each unknown to a decision, risk, or boundary in this feature
- do not research a technology or subsystem in general; research only the narrow question that changes the plan
- if no concrete unknown remains, do not create `research.md`

Do not create `research.md` for generic brainstorming or obvious implementation work that can already be planned from the spec and repository evidence.

## Invariants

- The plan MUST comply with the constitution.
- Inspect is a mandatory prerequisite for planning. Do not plan from a spec that has not been inspected and persisted.
- Plan the current spec, not an idealized architecture; use repository reality.
- Read code narrowly — enough to ground implementation surfaces, integration boundaries, and repository constraints. Broad repository exploration is not encouraged.
- Follow `.speckeep/templates/plan.md`, `.speckeep/templates/data-model.md`, and `.speckeep/templates/quickstart.md` when creating new files.
- The plan MUST name the concrete implementation surfaces expected to change.
- The plan MUST map each acceptance criterion to an implementation approach before `tasks` are written.
- Significant implementation choices become `DEC-*` entries. Each significant `DEC-*` should capture `Why`, `Tradeoff`, `Affects`, and `Validation`.
- Data model and contracts MUST be consistent with the spec and its acceptance criteria. Reference stable `AC-*` IDs when discussing acceptance-critical behavior; each entity or contract entry SHOULD reference the `AC-*` that justifies it.
- `data-model.md` MUST always exist by the end of plan. If there are no meaningful model changes, it MUST say so explicitly with: status, reason, and revisit triggers. Absence of the file forces downstream guessing and is a planning defect.
- Do not leave entity shape, boundary IO, or event payload details only in prose inside `plan.md`.
- If `data-model.md` contains real entities (not just the placeholder line), `plan.md` MUST include a one-line justification naming the specific entities, invariants, or lifecycle concerns that require it. If `contracts/` is created, `plan.md` MUST include a one-line justification naming the specific API or event boundary that requires it; otherwise state "No API or event boundaries introduced".
- Record technologies, libraries, framework choices, or version constraints only when they materially affect implementation shape, integration boundaries, validation, or risk.
- If a version or dependency is named, explain why it matters for this feature: compatibility, repository constraint, external contract, rollout risk, or validation.
- Do not enumerate stack details for completeness; capture only technical constraints that reduce downstream guesswork.
- Optional artifacts stay optional; do not create them by habit.
- Do not write the task checklist, edit implementation code, or emit verify/archive conclusions during planning.
- Treat generic wording such as `update backend accordingly`, `adjust logic as needed`, or `wire this through the system` as a refinement signal rather than a complete plan.
- If a downstream task writer would need to guess the method, boundary, or validation path for an `AC-*`, the plan is underspecified.
- The plan should be specific enough that both an agent and a human reviewer can see the intended implementation shape, tradeoffs, and rollout implications without rereading the whole repository.
- Use the project's configured documentation language; keep `plan.md`, `data-model.md`, `contracts/`, and optional `research.md` internally consistent.

## Content Quality Rules

- `## Goal` describes implementation shape, not a spec restatement.
- `--greenfield`: `## MVP Slice` must name the smallest independently demonstrable increment and its `AC-*` coverage.
- `--greenfield`: `## First Validation Path` should explain how a human or agent proves the MVP works without reading the whole repository.
- `--greenfield`: `## Bootstrapping Surfaces` should list the first directories, files, or boundaries that must exist before feature behavior can land.
- `## Implementation Surfaces` explains why each surface changes and whether it is new or existing.
- `## Acceptance Approach` maps every `AC-*` to touched surfaces and proof of observability.
- `## Data and Contracts` states what changes, what stays unchanged, and why.
- `## Sequencing Notes` separates must-happen-first work from parallelizable work.
- `## Risks` includes a mitigation, not just a label.
- `## Rollout and Compatibility` should be explicit when migration, flags, compatibility, or operational follow-up matter, and should say so plainly when they do not.
- `## Validation` ties checks back to `AC-*` or `DEC-*`, not a generic list of test ideas.
- `## Constitution Compliance` MUST be present and MUST list the specific constitutional constraints that apply to this feature (e.g., "Must use PostgreSQL per [CONST-DB]"), then confirm each one is satisfied or explain how the conflict is resolved or deferred. A bare "no conflicts" without listing the checked constraints is not sufficient — it makes the compliance check unverifiable by inspect.
- Add a short `Unknowns First` pass before finalizing the plan: if a decision, surface, or validation path is still unclear, record the unknown explicitly or stop for refinement.
- Prefer concrete implementation guidance over architecture essay prose. If a paragraph does not reduce downstream guesswork, tighten it.

## Output

- Write or patch the plan artifacts; state which optional artifacts were created and why.
- If `quickstart.md` was created, state that it exists to validate the MVP path without rereading the full plan.
- Summarize key technical decisions and call out risks blocking downstream phases.
- End with a summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`.
- When ready: `Ready for: /speckeep.tasks <slug>`.

## Self-Check

- Can `tasks` be written from this package without guesswork, and are optional artifacts justified?
