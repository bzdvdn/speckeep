# <Spec Title> Plan

## Phase Contract

Inputs: spec and minimal repo context for this feature.
Outputs: plan, data model, and contracts when required.
Stop if: the spec is too vague to plan safely.

## Goal

Describe the implementation shape for this feature without restating the full spec. Make it clear what changes, where the work lives, and why the approach is safe.

## Scope

- In-scope implementation area 1
- In-scope implementation area 2
- Explicitly call out the important boundary that remains untouched

## Implementation Surfaces

- Package, module, file, or boundary expected to change, plus why it is involved
- Package, module, file, or boundary expected to change, plus whether it is new or existing
- State explicitly when a new surface must be introduced and why the existing ones are insufficient

## Architecture Impact

- Local component or package impact
- Cross-boundary or integration impact when relevant
- Migration, compatibility, or rollout implication when relevant

## Acceptance Approach

- AC-001 -> implementation approach, touched surfaces, and proof that the result will be observable
- AC-002 -> implementation approach, touched surfaces, and proof that the result will be observable
- State explicitly when an AC depends on contracts, data-model expansion, rollout steps, or migration work

## Data and Contracts

- Reference `AC-*`, entities, and boundaries that drive implementation design
- Call out which data-model updates are required and which are intentionally unnecessary
- Call out which API or event contracts are affected and how compatibility is preserved
- State explicitly when no extra contract or data-model expansion is needed

## Implementation Strategy

- DEC-001 Decision title
  Why: why this approach is chosen over the obvious alternative
  Tradeoff: what this decision costs or constrains
  Affects: packages, modules, files, or boundaries touched
  Validation: test, check, or observable condition that proves the decision works
- DEC-002 Decision title
  Why: why this approach is chosen over the obvious alternative
  Tradeoff: what this decision costs or constrains
  Affects: packages, modules, files, or boundaries touched
  Validation: test, check, or observable condition that proves the decision works

## Incremental Delivery

### MVP (First Value)

- Minimal set of tasks that can be implemented and tested independently
- MVP readiness criteria: which ACs are covered and how to quickly verify

### Iterative Expansion

- What to add after MVP for the next value increment
- Which ACs are covered at each step and how to validate them independently

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
- Monitoring, auditability, or operational follow-up if the feature changes behavior in production
- State explicitly when no special rollout handling is needed

## Validation

- Automated tests to add or update
- Targeted manual checks, review evidence, or operational checks to run
- Acceptance IDs and decision IDs that each validation step proves

## Constitution Compliance

- no conflicts | list specific conflicts with constitution sections and how the plan resolves or defers them
