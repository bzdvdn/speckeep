# <Spec Title>

## Scope Snapshot

- In scope: one-line summary of the user-visible change this feature must deliver.
- Out of scope: one-line summary of adjacent work this feature does not take on.

## Goal

One concise paragraph covering who benefits, what changes for them, and how success becomes visible.

## Primary User Flow

1. Starting point: where the user or system begins.
2. Main interaction: what they do or what event happens.
3. Outcome: what becomes true when the feature works.
4. Failure/fallback path when it materially changes the experience.

## User Stories

- Optional for brownfield; recommended for greenfield when it reduces ambiguity.
- P1 Story: the smallest independently valuable user outcome.
- P2 Story: the next valuable extension after the MVP.
- State `none` when the feature is better described without story grouping.

## MVP Slice

- The smallest slice that delivers independently demonstrable value.
- Name which `AC-*` this slice must satisfy first.

## First Deployable Outcome

- What can be demonstrated, reviewed, or manually validated after the first implementation pass.
- State explicitly when the feature is not deployable independently and why.

## Scope

- In-scope behavior or boundary 1
- In-scope behavior or boundary 2
- Repository or product surface intentionally included

## Context

- Existing repository constraint, dependency, or operational reality shaping the solution
- Existing user workflow or system behavior this feature must preserve or extend
- Assumption that must remain true for the feature to be valid

## Requirements

- RQ-001 Clear, testable requirement written as expected behavior or capability
- RQ-002 Clear, testable requirement written as expected behavior or capability
- RQ-003 System MUST [capability] [NEEDS CLARIFICATION: detail — option A or option B?]
- Scope each requirement tightly enough that a reviewer can confirm it is satisfied
- Mark unclear requirements inline with `[NEEDS CLARIFICATION: what is unknown and why]`

## Non-Goals

- Out-of-scope behavior or adjacent enhancement 1
- Out-of-scope behavior or adjacent enhancement 2
- Deferred refinement that should not be silently pulled into implementation

## Acceptance Criteria

### AC-001 Criterion title

- Why this matters: one line of user or business value
- **Given** the initial state or precondition
- **When** the action or event
- **Then** the expected observable outcome
- Evidence: what a developer, reviewer, or user can directly observe when this criterion passes

### AC-002 Criterion title

- Why this matters / Given / When / Then / Evidence — same structure as AC-001

## Assumptions

- Assumption about environment, users, or system state that must hold
- Reasonable default chosen when the feature description did not specify a detail
- Dependency on an existing system, service, or behavior assumed stable

## Success Criteria

- SC-001 Measurable outcome defining quality beyond behavioral correctness (e.g., "Export completes in under 5s for 10k rows")
- SC-002 Measurable outcome (e.g., "Error rate stays below 0.1% after rollout")
- Include only when the feature has meaningful performance/reliability/UX targets; omit for purely behavioral features

## Edge Cases

- Empty, first-run, or missing-data condition
- Failure, retry, or timeout behavior
- Permission, role, or conflicting-state condition when relevant

## Open Questions

- Question 1
- State `none` when the feature is clear enough to proceed
