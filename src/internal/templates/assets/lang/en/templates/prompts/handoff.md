# SpecKeep Handoff Prompt (compact)

Create a short handoff for one feature.

## Path Resolution

- Resolve `<specs_dir>` from `.speckeep/speckeep.yaml` (read ≤1 time per session). If the config is missing, use `.speckeep/specs`.

## Inputs

- current phase (state)
- `<specs_dir>/<slug>/plan/tasks.md`
- recent changes (files/commands), if known

## Output expectations

- `Slug`, `Phase`, `What changed`, `Open tasks`, `Blockers`, `Next command`
- Final line: `Ready for: /speckeep.<next> <slug>`
