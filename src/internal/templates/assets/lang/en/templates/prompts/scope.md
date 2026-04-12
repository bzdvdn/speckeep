# SpecKeep Scope Prompt

You are performing a quick scope boundary check for one feature.

## Goal

Answer one question: does the current plan, task list, or set of decisions stay within the scope boundaries defined in the spec?

This command is optional, produces no persistent artifact, and can be called at any point in the workflow.

## Phase Contract

Inputs: `.speckeep/specs/<slug>/spec.md` (scope sections only), plus the target artifact for the flag in use.
Outputs: inline conversation response only — no file is written.
Stop if: slug ambiguous or spec missing.

## Flags

`--plan`: check `plan.md` against spec scope.
`--tasks`: check `tasks.md` against spec scope.

Default (no flag): check whatever is present — prefer `tasks.md` if it exists, otherwise `plan.md`, otherwise report scope boundaries from the spec only.

## Load First

Read only the scope-relevant sections of the spec:

- `## Scope`, `## Non-Goals`, `## Out of Scope`, `## Scope Snapshot`, or equivalent sections in `.speckeep/specs/<slug>/spec.md`

Do not read the full spec by default. If the spec has no dedicated scope section, read `## Goal` and `## Acceptance Criteria` to infer the intended boundary.

## Load If Present

- `.speckeep/specs/<slug>/plan/tasks.md` — when `--tasks` or default mode applies
- `.speckeep/specs/<slug>/plan/plan.md` — when `--plan` or default mode applies and tasks do not exist

## Do Not Read By Default

- full spec beyond scope sections
- `.speckeep/specs/<slug>/plan/data-model.md`
- `.speckeep/specs/<slug>/plan/contracts/`
- `.speckeep/specs/<slug>/plan/research.md`
- `.speckeep/specs/<slug>/plan/challenge.md`
- implementation files
- script source files
- unrelated specs or plan packages

## Stop Conditions

Stop and ask only if:

- the slug is ambiguous
- no spec exists for the slug

## Check Rules

Read the spec scope sections first. Then compare the target artifact against those boundaries.

Flag as `drift` when:
- a task, decision, or implementation surface is outside the in-scope list but is a minor or implicit extension
- the plan introduces a component or integration not mentioned in scope but arguably required by an AC
- a DEC-* makes a technology or architectural choice that was explicitly marked out-of-scope or non-goal

Flag as `out-of-scope` when:
- a task or decision directly contradicts the out-of-scope or non-goals section
- the plan introduces a major new workstream, surface, or integration not justified by any AC
- the feature boundary has visibly shifted from what the spec describes

Flag as `in-scope` when:
- all tasks and decisions can be traced to in-scope ACs or requirements
- no out-of-scope or non-goal item is touched

Be specific: name the exact task ID, DEC-*, or plan section that is drifting, and quote the spec boundary it crosses.

Do not flag implementation details as scope drift. Scope drift is about feature boundaries and acceptance coverage, not about how something is built.

## Output Format

Respond inline only — do not write a file.

Use this structure in the conversation:

- **Verdict**: `in-scope`, `drift`, or `out-of-scope`
- **Boundary** (one line): what the spec defines as in-scope and out-of-scope
- **Findings**: specific items that drift or violate scope, each referencing the artifact ID and the spec boundary it crosses — omit if verdict is `in-scope`
- **Action**: one concrete next step — omit if verdict is `in-scope`

Keep the response compact. If the feature is cleanly in scope, a two-line response is sufficient.

## Self-Check

- Did I read only the scope-relevant sections of the spec?
- Is every finding grounded in a specific artifact ID and a specific spec boundary?
- Did I avoid flagging implementation choices as scope violations?
- Is the response compact enough to be useful as a quick check?
