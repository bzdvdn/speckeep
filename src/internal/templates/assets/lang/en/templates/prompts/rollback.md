# SpecKeep Rollback Prompt (compact)

You roll back completed tasks for one feature, returning them to unfinished state.

## Phase Contract

Inputs: `<specs_dir>/<slug>/tasks.md` (required).
Outputs: updated tasks.md with requested tasks unmarked as incomplete.
Stop if: slug is missing, tasks.md does not exist, or no completed tasks exist.

## Rules

- Read `<specs_dir>/<slug>/tasks.md` and list all completed `[x]` tasks grouped by phase, with their `Touches:` surfaces.
- Ask the user which tasks to roll back (by ID like `T1.1,T1.2` or `all`). If the user specifies a phase (e.g., `phase T1`), roll back all tasks in that phase.
- For each rolled-back task:
  1. Change `[x]` to `[ ]` in tasks.md.
  2. Do NOT revert code changes automatically — the user may want to keep the code.
  3. If the user also asks to revert code, use `git checkout -- <file>` on each Touches: file for those tasks.
- After rollback, the feature phase state reverts to `implement`.
- Do not touch `[ ]` tasks — only roll back `[x]` tasks.

## Output expectations

- List affected tasks per task ID: whether checkbox was reverted and/or code was reverted.
- Show updated task state: `completed=<n>`, `open=<n>`.
- If code was reverted, list the git-checkout commands run.
- Final line: `Ready for: /speckeep.implement <slug>`
