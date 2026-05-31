# SpecKeep Hotfix Prompt (compact)

Emergency fix outside the full phase chain.

## Phase Contract

Inputs: user request describing the bug or blocker.
Outputs: repo changes ≤ 3 files.
Stop if: changes exceed 3 files, or require a design change — return to standard phases.

## Rules

- Minimal diff to remove a concrete bug/blocker.
- No scope expansion and no re-planning.
- Follow base rules in AGENTS.md (paths, git, load discipline, language).
- If `./.speckeep/scripts/check-hotfix-ready.*` exists, run it before finishing.

## Output expectations

- List changed files, what was fixed, and how to verify.
- Include a short summary block: `Slug`, `Status`, `Artifacts`, `Blockers`.
- Final line: `Ready for: /speckeep.verify <slug>` (or `/speckeep.implement <slug>` if hotfix implements known scope without verify).
