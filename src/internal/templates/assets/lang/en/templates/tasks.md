# <Spec Title> Tasks

## Phase Contract

Inputs: plan and minimal supporting artifacts for this feature.
Outputs: ordered executable tasks with coverage mapping.
Stop if: tasks would be vague or acceptance coverage cannot be mapped.

## Surface Map

| Surface | Tasks |
|---------|-------|
| src/models/feature.ts | T1.1, T1.2, T2.1 |
| src/handlers/feature.ts | T2.1, T2.2 |
| src/tests/feature.test.ts | T3.1 |

## Phase 1: Foundation

Goal: establish the minimum structure, contracts, or data prerequisites so later work stays predictable.

- [ ] T1.1 Establish or align the feature scaffold — the implementation entrypoint exists and matches the planned surface area. Touches: src/models/feature.ts
- [ ] T1.2 Add foundational model, contract, migration, or flag work when the later phases depend on it. Touches: src/models/feature.ts, src/db/migrations/

## Phase 2: MVP Slice

Goal: deliver the smallest independently demonstrable product value before broader expansion.

- [ ] T2.1 Implement the MVP acceptance path — the first usable behavior works end to end. Touches: src/handlers/feature.ts, src/models/feature.ts
- [ ] T2.2 Prove the MVP path — focused checks or tests confirm the slice is reviewable. Touches: src/tests/feature.test.ts

## Phase 3: Core Implementation

Goal: deliver the primary feature behavior and the important edge or failure paths.

- [ ] T3.1 Implement the main acceptance path beyond MVP — the remaining primary behavior works on the intended surface. Touches: src/handlers/feature.ts, src/models/feature.ts
- [ ] T3.2 Implement edge, failure, permission, or conflicting-state behavior when it changes user-visible outcomes. Touches: src/handlers/feature.ts

## Phase 4: Validation

Goal: prove the feature works and leave the package in a reviewable state.

- [ ] T4.1 Add or update automated coverage — tests or checks prove the intended behavior and guard regressions. Touches: src/tests/feature.test.ts
- [ ] T4.2 Run verification, cleanup, or documentation updates required to leave the feature ready for review or verify

## Acceptance Coverage

- AC-001 -> T1.2, T2.1, T4.1
- AC-002 -> T3.2, T4.1, T4.2

## Notes

- Keep task ordering aligned with the plan and use later phases only for work that truly depends on earlier ones
- Use phase-scoped task IDs in the form `T<phase>.<index>`
- Make each task concrete, measurable, and executable as a single coherent slice of work
- Prefer action verbs tied to observable outcomes: implement, add, migrate, validate, remove, backfill
- For greenfield or first-feature work, prefer MVP-first sequencing over broad technical completeness
- Reference 1-2 stable IDs in task text when useful (`AC-*`, `RQ-*`, `DEC-*`)
- Separate validation work from broad implementation work instead of hiding proof inside a large task
- Mark tasks complete as implementation progresses and do not leave acceptance criteria without task coverage
- State explicitly when a phase is intentionally omitted because the feature does not need it
