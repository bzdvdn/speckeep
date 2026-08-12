# SpecKeep Hotfix Prompt (compact)

You act as a **senior engineer in incident mode**. Find the smallest diff that removes the concrete bug/blocker safely — no scope expansion, no re-planning.

Emergency fix outside the full phase chain.

## Phase Contract

Inputs: user request describing the bug or blocker.
Outputs: repo changes ≤ 3 files.
Stop if: changes exceed 3 files, or require a design change — return to standard phases.

## Rules

- Minimal diff to remove a concrete bug/blocker.
- No scope expansion and no re-planning.
- Follow base rules in AGENTS.md (paths, git, load discipline, language, scripts).

## Output expectations

- List changed files, what was fixed, and how to verify.
- Include a short summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for` (set by the final line below).
- Resolve `workflow.verify` per the **Verify gate policy** in AGENTS.md (`.speckeep/speckeep.yaml`, ≤1 read per session): if `required`, the fix must pass verify before archive.
- Final line:
  - if `workflow.verify: required`: `Ready for: /spk.verify <slug>`
  - if `workflow.verify` is `optional`/absent: `Ready for: /spk.implement <slug>` (known scope, no audit gate) — or `Ready for: speckeep archive <slug> .` when the hotfix is already proven.
