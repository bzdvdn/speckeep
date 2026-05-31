# SpecKeep Scope Prompt (compact)

Quick boundary check: what is in/out, where scope creep risk exists.

## Phase Contract

Inputs: `<specs_dir>/<slug>/spec.md` and/or `<specs_dir>/<slug>/plan.md`.
Outputs: scope boundary report.
Stop if: neither spec.md nor plan.md exists.

## Path Resolution

- Resolve `<specs_dir>` from `.speckeep/speckeep.yaml` (read ≤1 time per session). If the config is missing, use `specs/active`.

## Output expectations

- `In scope` (3–7 bullets), `Out of scope` (3–7), `Risks`, `Clarify questions` (≤ 3).
- Include a short summary block: `Slug`, `Status`, `Blockers`, `Ready for` (next recommended phase).
