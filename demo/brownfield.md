# Brownfield Walkthrough

This walkthrough shows how to introduce `speckeep` into an existing repository without trying to document the whole codebase at once.

## Goal

Use SpecKeep as a strict, lightweight coordination layer for one active feature in a real repository.

## When To Use This

Use this flow when:

- the repository already exists
- the team is already shipping code
- you want better agent discipline without adding a heavy process layer
- you want feature-local artifacts instead of shared mutable planning state

## Suggested Demo Setup

Start in an existing repository root:

```bash
speckeep init . --lang en --shell sh --agents codex
speckeep doctor .
```

Expected outcome:

- `.speckeep/` exists
- `AGENTS.md` contains SpecKeep guidance
- project-local agent command files exist for the chosen targets
- `doctor` reports a healthy workspace

## Feature-First Adoption

Do not try to spec the whole repository.

Pick one active feature and drive only that scope through the workflow:

1. `/speckeep.constitution`
2. `/speckeep.spec`
3. `/speckeep.inspect`
4. `/speckeep.plan`
5. `/speckeep.tasks`
6. `/speckeep.implement`

Example feature request:

```text
/speckeep.spec Add partner-specific ingestion scheduling with retry policy overrides.
```

Expected artifact growth:

- `specs/active/<slug>/spec.md`
- `specs/active/<slug>/inspect.md`
- `specs/active/<slug>/plan/plan.md`
- `specs/active/<slug>/plan/tasks.md`
- optional compact plan artifacts only when needed

## What To Highlight In A Demo

For a brownfield demo, the most important points are:

- the repository does not need a full rewrite or full-repo spec effort
- the workflow stays feature-local
- `inspect` is required before planning
- code reading stays narrow and task-driven
- generated scripts help the agent validate readiness before widening context

## Good Before / After Story

Before SpecKeep:

- feature intent lives in chat history or scattered notes
- agents reread too much repository context
- plans drift from specs
- branch work is hard to audit

After SpecKeep:

- feature intent is persisted in small canonical files
- each phase has a clear entrypoint
- the agent is told what to read first and what not to read by default
- branch-local artifacts keep active work reviewable

## Recommended Capture Format

For public promotion, this is usually best shown as:

- one short terminal GIF for setup
- one markdown walkthrough for the workflow
- one screenshot or excerpt of the generated artifacts

That keeps the demo maintainable while still showing real-world value.
