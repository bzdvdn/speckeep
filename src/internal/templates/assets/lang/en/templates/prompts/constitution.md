# SpecKeep Constitution Prompt (compact)

You create or update the project constitution.

## Phase Contract

Inputs: user request + minimum repo context needed to define constraints/architecture.
Outputs: `project.constitution_file` (default: `CONSTITUTION.md`).
Stop if: rules remain `TBD`/placeholder or contradict repo reality without an explicit decision.

## Rules

- Constitution is authoritative: short, concrete, testable rules (no philosophy).
- Include: Purpose, principles, constraints, tech stack, architecture, language policy, workflow.
- Always use `.speckeep/templates/constitution.md` as the skeleton and output format. Do not look for “examples” in other constitutions/projects for shape: it’s wasted tokens and drift.
- If `./.speckeep/scripts/check-constitution.*` exists, run it before finishing.

## Output expectations

- Write/patch the constitution.
- Generate `.speckeep/constitution.summary.md` using this strict compact format (rules only, no prose paragraphs):
  - `Purpose:` one line
  - `Non-negotiables:` 3-6 bullets (`MUST` / `MUST NOT`)
  - `Stack/Architecture:` 2-5 bullets
  - `Workflow/DoD:` 3-6 bullets (include traceability and proof requirements)
  - `Repo Map Policy:` 2-4 bullets
  - `Languages:` one line (`docs=...`, `agent=...`, `comments=...`)
  - hard limit: ≤200 words total
- Summarize the key rules and what changed.
- Final line: `Ready for: /speckeep.spec <slug>`
