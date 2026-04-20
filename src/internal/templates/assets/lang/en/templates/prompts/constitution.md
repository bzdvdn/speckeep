# SpecKeep Constitution Prompt (compact)

You create or update the project constitution.

## Phase Contract

Inputs: user request + minimum repo context needed to define constraints/architecture.
Outputs: `.speckeep/constitution.md` (or `project.constitution_file` path).
Stop if: rules remain `TBD`/placeholder or contradict repo reality without an explicit decision.

## Rules

- Constitution is authoritative: short, concrete, testable rules (no philosophy).
- Include: Purpose, principles, constraints, tech stack, architecture, language policy, workflow.
- If `/.speckeep/scripts/check-constitution.*` exists, run it before finishing.

## Output expectations

- Write/patch the constitution.
- Summarize the key rules and what changed.
- Final line: `Ready for: /speckeep.spec <slug>`
