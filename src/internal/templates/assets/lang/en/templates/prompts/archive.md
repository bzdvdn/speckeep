# SpecKeep Archive Prompt (compact)

You archive one feature package into `.speckeep/archive/`.

## Path Resolution

- Resolve `<specs_dir>` from `.speckeep/speckeep.yaml` (read ≤1 time per session). If the config is missing, use `.speckeep/specs`.

## Phase Contract

Inputs: `<specs_dir>/<slug>/` (spec/inspect/plan/tasks/verify as present).
Outputs: snapshot under `.speckeep/archive/<slug>/...` (move-based by default).
Stop if: verify is not complete or the archive status cannot be justified.

## Rules

- Prefer running `/.speckeep/scripts/check-archive-ready.*` (slug first) before archiving.
- Default archive status is `completed`. Non-`completed` requires an explicit reason.
- Archive is a snapshot step, not a place for new implementation changes.

## Output expectations

- Create the snapshot; list moved artifacts and final status.
- This is the terminal workflow step for this feature (after verify).
- Final line: `Ready for: /speckeep.recap`
