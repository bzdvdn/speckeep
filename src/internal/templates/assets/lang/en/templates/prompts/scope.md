# SpecKeep Scope Prompt (compact)

Quick boundary check: what is in/out, where scope creep risk exists.

## Path Resolution

- Resolve `<specs_dir>` from `.speckeep/speckeep.yaml` (read ≤1 time per session). If the config is missing, use `.speckeep/specs`.

## Inputs

`<specs_dir>/<slug>/spec.md` and/or `<specs_dir>/<slug>/plan/plan.md` (as present).

## Output expectations

- `In scope` (3–7 bullets), `Out of scope` (3–7), `Risks`, `Clarify questions` (≤ 3)
