# SpecKeep Spec Prompt (compact)

You create or update one feature spec: `<specs_dir>/<slug>/spec.md`.

## Path Resolution

- Resolve `<specs_dir>` from `.speckeep/speckeep.yaml` (read ≤1 time per session). If the config is missing, use `.speckeep/specs`.

## Phase Contract

Inputs: `.speckeep/constitution.md`, user request, minimum required repo context.
Outputs: `<specs_dir>/<slug>/spec.md`.
Stop if: the request is ambiguous/multi-feature or would force inventing `AC-*`.

## Mandatory Rules

- **Branch-first**: before writing any file, switch/create `feature/<slug>` (or `--branch`). If not possible → stop and report why.
- Do not try to “generate the spec via CLI”: as the agent, you must write/update `<specs_dir>/<slug>/spec.md` directly. `/.speckeep/scripts/check-*.{sh,ps1}` are validation gates only, not artifact generators.
- Do not look for “examples” in neighboring specs from other slugs: take structure from `.speckeep/templates/spec.md` and the user’s requirements. Reading other specs for style is wasted tokens and scope drift.
- Spec captures intent, not plan/tasks. No implementation steps or decomposition.
- Every `AC-*` is Given/When/Then with observable proof in Then.
- Required sections: Out of Scope, Assumptions, Open Questions (or `none`).
- Clarify with 1–3 targeted questions only if otherwise you must guess AC or scope boundaries.
- If invoked with `--name` but without enough description, ask for it and treat the next non-command user message as the continuation. If the next message starts with `/speckeep.`, staged mode is canceled.
- Do not pin technologies/versions unless required by the user or a hard repo/contract constraint. If a technology choice is an implementation preference, record it in `plan`, not in `spec`.
- Refine instead of guessing: if the request implies multiple feature slugs or multiple independent specs, stop and ask for one concrete feature.
- If `/.speckeep/scripts/check-spec-ready.*` exists, run it (slug first) before finishing.

## Output expectations

- Write/patch `spec.md` (patch > rewrite).
- Summarize: goal, scope, AC list, open questions/blockers.
- Include a short summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`.
- Final line: `Ready for: /speckeep.inspect <slug>`
