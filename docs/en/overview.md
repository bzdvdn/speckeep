# Overview

## What SpecKeep Is

`speckeep` keeps project intent, specifications, plan artifacts, and tasks in plain files. It is designed to help humans and development agents share the same project context without introducing a rigid process engine.

## Core Ideas

- The constitution is the highest-priority project document.
- Every feature starts as a spec and evolves through a strict workflow.
- Each feature should be developed in its own git branch so teams can collaborate without shared-memory merge churn.
- Generated docs and prompts can use English or Russian.
- Agent workflows should load only the minimum context needed.
- Readiness checks belong in scripts whenever possible.

## Workspace Layout

```text
.speckeep/
  speckeep.yaml
  constitution.md
  specs/
    <slug>/
      spec.md
      inspect.md
      summary.md
      hotfix.md
      plan/
        plan.md
        tasks.md
        data-model.md
        research.md
        contracts/
          api.md
          events.md
  archive/
    <slug>/
      <YYYY-MM-DD>/
        summary.md
        spec.md
        plan.md
        tasks.md
        data-model.md
        research.md
        contracts/
  templates/
  scripts/
AGENTS.md
```

## Public CLI Surface

The public CLI stays intentionally small:

- `speckeep init [path]`
- `speckeep add-agent [path]`
- `speckeep list-agents [path]`
- `speckeep remove-agent [path]`
- `speckeep cleanup-agents [path]`
- `speckeep doctor [path]`
- `speckeep list-specs [path]`
- `speckeep show-spec <name> [path]`

Creation and evolution of specs, plans, tasks, and implementation are agent-facing workflows, not public CLI subcommands.
