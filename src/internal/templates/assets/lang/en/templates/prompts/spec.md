# SpecKeep Spec Prompt (compact)

You create or update one feature spec: `<specs_dir>/<slug>/spec.md`.

## Phase Contract

Inputs: `.speckeep/constitution.summary.md` (preferred when present) or `project.constitution_file` (default: `CONSTITUTION.md`), user request, minimum required repo context.
Outputs: `<specs_dir>/<slug>/spec.md`.
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
- Constitution: see AGENTS.md (`.speckeep/constitution.summary.md` preferred over full constitution).
- Do not pin technologies/versions unless required by the user or a hard repo/contract constraint. If a technology choice is an implementation preference, record it in `plan`, not in `spec`.
- Refine instead of guessing: if the request implies multiple feature slugs or multiple independent specs, stop and ask for one concrete feature.
- If `./.speckeep/scripts/check-spec-ready.*` exists, run it (slug first) before finishing.

## Output expectations

- Write/patch `spec.md` (patch > rewrite).
- Self-check before finishing: no `TODO`/`???`/`<placeholder>` in spec.md; every AC has Given/When/Then with observable proof; Out of Scope + Assumptions + Open Questions sections exist.
- Summarize: goal, scope, AC list, open questions/blockers in the response; do not create extra derived recap files just for this summary.
- Include a short summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`.
- Next steps (offer both):
  - Deep quality review: `/speckeep.inspect <slug>` — checks constitution alignment, AC completeness, ambiguity
  - Skip to planning if spec looks solid: `/speckeep.plan <slug>`
- Final line (mandatory): `Ready for: /speckeep.inspect <slug>` or `Ready for: /speckeep.plan <slug>`. Prefer `/speckeep.inspect` when ambiguity, risk, or open questions remain.
