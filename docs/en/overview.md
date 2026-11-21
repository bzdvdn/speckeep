# Overview

## What SpecKeep Is

`speckeep` keeps project intent, specifications, feature artifacts, and tasks in plain files. It is designed to help humans and development agents share the same project context without introducing a rigid process engine.

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
  templates/
  scripts/
CONSTITUTION.md
specs/
  active/
    <slug>/
      spec.md
      inspect.md
      hotfix.md
      plan.md
      tasks.md
      data-model.md
      verify.md
      research.md
      contracts/
        api.md
        events.md
  archived/
    <slug>/
      <YYYY-MM-DD>/
        summary.md
        spec.md
        plan.md
        tasks.md
        data-model.md
        research.md
        contracts/
AGENTS.md
```

`tasks.md` is the main operational entrypoint for `implement` and `verify`; `spec.md` and `plan.md` remain source artifacts, while `.speckeep/constitution.summary.md` is the preferred compact policy context when constitution rules are needed.

## Public CLI Surface

The CLI is intentionally operational: workspace management, inspection, traceability, and closing. Day-to-day spec/plan/tasks/implement work stays agent-facing.

- **Workspace**: `init`, `refresh`, `add-agent`, `list-agents`, `remove-agent`, `cleanup-agents`, `doctor`, `migrate`, `import`
- **Features**: `list-specs`, `show-spec`, `features`, `feature`, `feature repair`, `status`, `check`, `risk`, `export`, `context`
- **Traceability**: `trace`, `trace init`, `trace --files`, `trace --file`, `trace rebuild`
- **Close**: `archive`, `finish`, `converge`, `guard`, `list-archive`
- **Scaffolding**: `demo`, `explore`, `schema`
- **Self**: `self check`, `self upgrade`

Each command, flag, and exit-code contract is documented in the [CLI Reference](/en/cli).
