# SpecKeep Spec Prompt (compact)

You create or update one feature spec: `<specs_dir>/<slug>/spec.md`.

## Phase Contract

Inputs: `project.constitution_file` (default: `CONSTITUTION.md`), user request, minimum required repo context.
Outputs: `<specs_dir>/<slug>/spec.md`, `<specs_dir>/<slug>/spec.digest.md`, `<specs_dir>/<slug>/summary.md`.
Stop if: the request is ambiguous/multi-feature or would force inventing `AC-*`.

## Mandatory Rules

- **Branch-first**: before writing any file, switch/create `feature/<slug>` (or `--branch`). If not possible → stop and report why.
- Do not try to “generate the spec via CLI”: as the agent, you must write/update `<specs_dir>/<slug>/spec.md` directly.
  - There is no `speckeep spec` subcommand. Do not run `./.speckeep/scripts/run-speckeep.* spec <slug>`.
  - `./.speckeep/scripts/check-*.{sh,ps1}` are validation gates only, not artifact generators.
- Do not read any `<specs_dir>/*/spec.md` from other slugs for any reason — not for style, not for format, not for examples. Do not list or scan `<specs_dir>/` to survey existing slugs. The template `.speckeep/templates/spec.md` is the sole structure reference; reading it once is sufficient.
- Spec captures intent, not plan/tasks. No implementation steps or decomposition.
- Every `AC-*` is Given/When/Then with observable proof in Then.
- Required sections: Out of Scope, Assumptions, Open Questions (or `none`).
- Clarify with 1–3 targeted questions only if otherwise you must guess AC or scope boundaries.
- If invoked with `--name` but without enough description, ask for it and treat the next non-command user message as the continuation. If the next message starts with `/speckeep.`, staged mode is canceled.
- Do not pin technologies/versions unless required by the user or a hard repo/contract constraint. If a technology choice is an implementation preference, record it in `plan`, not in `spec`.
- Refine instead of guessing: if the request implies multiple feature slugs or multiple independent specs, stop and ask for one concrete feature.
- If `./.speckeep/scripts/check-spec-ready.*` exists, run it (slug first) before finishing.

## Output expectations

- **Write `<specs_dir>/<slug>/spec.digest.md` first** (mandatory, always, even on patch). One line per `AC-*`, format `AC-NNN: <≤10-word summary>`. Example:
  ```
  AC-001: user submits login form with valid credentials
  AC-002: invalid password shows inline error message
  ```
- Write/patch `spec.md` (patch > rewrite).
- Write `summary.md` (≤25 lines): Goal (1–2 sentences), AC table (one row per `AC-*`), Out of Scope list.
- Self-check before finishing: no `TODO`/`???`/`<placeholder>` in spec.md; every AC has Given/When/Then with observable proof; Out of Scope + Assumptions + Open Questions sections exist.
- Summarize: goal, scope, AC list, open questions/blockers.
- Include a short summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`.
- Next steps (offer both):
  - Deep quality review: `/speckeep.inspect <slug>` — checks constitution alignment, AC completeness, ambiguity
  - Skip to planning if spec looks solid: `/speckeep.plan <slug>`
