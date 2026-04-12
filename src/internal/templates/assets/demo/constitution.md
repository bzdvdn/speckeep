# My Project Constitution (Demo)

This constitution is a filled example used by `speckeep demo`. Adjust it for your real project after running `/speckeep.constitution`.

## Purpose

Provide a clear, stable foundation for specification-driven development so agents and humans can collaborate safely and predictably.

## Core Principles

- Prefer clarity over cleverness.
- Keep interfaces small and testable.
- Preserve traceability from spec (`AC-*`, `RQ-*`) to tasks (`T*`) to code.

## Constraints

- No breaking changes without explicit decision records.
- No hidden shared mutable state across subsystems.
- Keep defaults safe and usable out of the box.

## Decision Priorities

1. Correctness
2. Safety
3. Maintainability
4. Performance (when measured and justified)

## Key Quality Dimensions

- Deterministic behavior and clear error handling
- Test coverage for acceptance paths and critical edge cases
- Simple operational story (logs, metrics, predictable rollout)

## Language Policy

- Documentation language: English
- Agent interaction language: English
- Code comments language: English

## Development Workflow

- Use `feature/<slug>` branches for feature work.
- Keep spec → inspect → plan → tasks aligned; do not implement from a stale spec.
- Prefer small, reviewable slices and ship incrementally.

## Governance

The constitution is the highest-priority project document. If a plan conflicts with this constitution, record an exception explicitly and justify it.

## Exceptions Protocol

When deviating from this constitution:

1. Document the exception and the reason.
2. Define the boundaries and the rollback plan.
3. Add validation steps to detect regressions.

## Last Updated

2026-04-06

