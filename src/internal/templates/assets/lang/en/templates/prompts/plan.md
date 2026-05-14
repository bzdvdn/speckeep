# SpecKeep Plan Prompt (compact)

You create or update plan artifacts for one feature in `<specs_dir>/<slug>/`.

Follow base rules in `AGENTS.md`.

## Phase Contract

Inputs: `.speckeep/constitution.summary.md` (preferred when present) or `project.constitution_file` (default: `CONSTITUTION.md`), `<specs_dir>/<slug>/spec.md`, `<specs_dir>/<slug>/inspect.md` (optional; if present, must be `pass`).
Outputs: `<specs_dir>/<slug>/plan.md`, and when required `<specs_dir>/<slug>/data-model.md`, `<specs_dir>/<slug>/contracts/*`, `<specs_dir>/<slug>/research.md`.
Stop if: inspect.md is present and not `pass`, the goal is ambiguous, or planning would require inventing requirements/AC.

## Rules

- Plan must preserve spec intent: no new major workstreams outside `spec.md`.
- Record only implementation-critical decisions: surfaces, sequencing, risks, trade-offs (`DEC-*`).
- Always use `.speckeep/templates/plan.md` as the skeleton and output format (and `.speckeep/templates/data-model.md` when needed). Do not look for “examples” in neighboring feature artifacts from other slugs: reading other plans for shape is wasted tokens and scope drift.
- If the data model does not change, still create `data-model.md` with an explicit `status: no-change` + rationale.
- Create `research.md` only when needed (e.g., external dependency/integration boundary, multiple realistic implementation options, or a high-risk unknown). Do not create `research.md` for generic brainstorming.
- When constitution context is needed, load `.speckeep/constitution.summary.md` first if it exists; only fall back to `project.constitution_file` when the summary is absent.
- Minimum context: current slug only; narrow repo reads (no full-repo scans).
- If `./.speckeep/scripts/check-plan-ready.*` exists, run it (slug first) before writing.

## Output expectations

- Write/patch `<specs_dir>/<slug>/plan.md` (create additional artifacts only when justified).
- Inside `plan.md`, keep compact sections for `DEC-*`, surfaces, risks, and data-model/contract impact; do not move that recap into separate digest files.
- Summarize key `DEC-*`, surfaces, sequencing constraints, and risks.
- Include a short summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`.
- Final line: `Ready for: /speckeep.tasks <slug>`
