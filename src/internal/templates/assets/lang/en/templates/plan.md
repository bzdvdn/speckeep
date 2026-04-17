# <Spec Title> Plan

## Phase Contract

Inputs: spec and minimal repo context.
Outputs: plan, data model, and contracts when required.
Stop if: the spec is too vague to plan safely.

## Goal

Describe the implementation shape without restating the full spec. Make clear what changes, where the work lives, and why the approach is safe.

## MVP Slice

- Name the smallest independently demonstrable increment.
- List the `AC-*` that must be satisfied before expanding scope.

## First Validation Path

- Explain how a human or agent can prove the MVP works quickly.
- Prefer a short manual or scripted validation path over a broad checklist.

## Scope

- In-scope implementation area 1
- In-scope implementation area 2
- Call out the important boundary that remains untouched

## Implementation Surfaces

- Package/module/file/boundary expected to change, plus why it is involved and whether new or existing
- Say explicitly when a new surface must be introduced and why existing ones are insufficient

## Bootstrapping Surfaces

- First directories/files/boundaries that must exist before feature behavior can land
- State `none` when the repository already has the needed structure

## Architecture Impact

- Local component or package impact
- Cross-boundary or integration impact when relevant
- Migration, compatibility, or rollout implication when relevant

## Acceptance Approach

- AC-001 -> implementation approach, touched surfaces, and proof that the result will be observable
- AC-002 -> implementation approach, touched surfaces, and proof that the result will be observable
- State explicitly when an AC depends on contracts, data-model expansion, rollout, or migration

## Data and Contracts

- Reference `AC-*`, entities, and boundaries driving implementation design
- Call out which data-model updates are required and which are intentionally unnecessary
- Call out which API/event contracts are affected and how compatibility is preserved
- State explicitly when no extra contract expansion is needed
- `data-model.md` is always required: either document the model changes or point to the explicit no-change stub

## Implementation Strategy

- DEC-001 Decision title
  Why: why this approach over the obvious alternative
  Tradeoff: what this decision costs or constrains
  Affects: packages/modules/files/boundaries touched
  Validation: test, check, or observable condition proving it works
- DEC-002 Decision title — same structure (Why / Tradeoff / Affects / Validation)

## Incremental Delivery

### MVP (First Value)

- Minimal set of tasks implementable and testable independently
- MVP readiness criteria: which ACs covered and how to quickly verify

### Iterative Expansion

- What to add after MVP for the next value increment
- Which ACs covered at each step and how to validate them independently

## Sequencing Notes

- What must happen first and why
- What can be parallelized safely
- What must stay behind a flag, migration, or guarded rollout step when relevant

## Risks

- Risk 1
  Mitigation: how the plan reduces or contains it
- Risk 2
  Mitigation: how the plan reduces or contains it

## Rollout and Compatibility

- Backfill, migration, feature-flag, or compatibility consideration
- Monitoring, auditability, or operational follow-up if behavior changes in production
- State explicitly when no special rollout handling is needed

## Validation

- Automated tests to add or update
- Targeted manual checks, review evidence, or operational checks
- Acceptance IDs and decision IDs each validation step proves

## Constitution Compliance

- no conflicts | list specific conflicts with constitution sections and how the plan resolves or defers them
