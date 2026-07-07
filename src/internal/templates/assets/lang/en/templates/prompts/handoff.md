# SpecKeep Handoff Prompt (compact)

Create a short handoff for one feature.

## Phase Contract

Inputs: current phase (state), `<specs_dir>/<slug>/tasks.md`, recent changes (files/commands if known).
Outputs: handoff summary.
Stop if: tasks.md is missing.

## Path Resolution

- Resolve `<specs_dir>` from `.speckeep/speckeep.yaml` (read ≤1 time per session). If the config is missing, use `specs/active`.

## Output expectations

- `Slug`, `Phase`, `What changed`, `Open tasks`, `Blockers`, `Next command`.
- Final line (detect phase from state):
  - If blocked: `Return to: /spk.<phase> <slug>`
  - If ready for next phase: `Ready for: /spk.<next> <slug>`
  - If all done: `Ready for: speckeep archive <slug> .`
