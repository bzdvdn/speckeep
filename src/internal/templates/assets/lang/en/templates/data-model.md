# <Spec Title> Data Model

## Scope

- Related acceptance IDs: `AC-001`
- Related decision IDs: `DEC-001`
- Status: `changed` or `no-change`
- State explicitly when the feature does not require a meaningful data-model change

## Entities

### DM-001 Entity 1

- Purpose:
- Source of truth:
- Invariants:
- Related acceptance IDs:
- Related decision IDs:
- Fields:
  - `field_name` - type/shape, required or optional, meaning, default or validation rule when important
- Lifecycle:
  - created when and by what action
  - updated when and by what action
  - removed, expired, or archived when
- Failure or consistency notes:
  - stale, duplicated, out-of-order, or partially-written state that must not happen

### DM-002 Entity 2 — same structure (Purpose / Source of truth / Invariants / Fields / Lifecycle / Failure or consistency notes)

## Relationships

- `DM-001 -> DM-002`: relationship, ownership, or cardinality
- State explicitly when there are no meaningful cross-entity relationships

## Derived Rules

- Computed value, normalization rule, or state transition rule
- State explicitly when no derived rules are needed

## State Transitions

- Trigger/event -> previous state -> next state
- Validation or guard that prevents an invalid transition
- State explicitly when lifecycle is simple enough that a separate transition list is unnecessary

## Out of Scope

- Entity, field, or state intentionally not modeled here

## No-Change Stub

Use this section instead of deleting the file when the feature does not change the model:

- Status: `no-change`
- Reason: this feature does not add or modify persisted entities, value objects, state transitions, or contract-relevant payload shapes
- Revisit triggers:
  - new stored data appears
  - new invariants or lifecycle states appear
  - API or event payload shape must be tracked here
