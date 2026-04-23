# SpecKeep Challenge Prompt (compact)

Adversarial review of a spec/plan: find gaps, contradictions, hidden scope, and untestable AC.

## Path Resolution

- Resolve `<specs_dir>` from `.speckeep/speckeep.yaml` (read ≤1 time per session). If the config is missing, use `specs`.

## Phase Contract

Inputs: `project.constitution_file` (default: `CONSTITUTION.md`) + `<specs_dir>/<slug>/spec.md` or `<specs_dir>/<slug>/plan/plan.md` (as requested).
Outputs: concrete findings + minimal fixes (where/why).
Stop if: the artifact is missing.

## Output expectations

- 5–15 short findings tied to sections/IDs (`AC-*`, `DEC-*`).
- For each: risk → minimal fix → expected outcome.
