# SpecKeep Plan Prompt (compact)

You create or update the plan package for one feature: `<specs_dir>/<slug>/plan/`.

Follow base rules in `AGENTS.md`.

## Phase Contract

Inputs: `project.constitution_file` (default: `CONSTITUTION.md`), `<specs_dir>/<slug>/spec.md`, `<specs_dir>/<slug>/inspect.md` (optional; if present, must be `pass`).
Outputs: `<specs_dir>/<slug>/plan/plan.md`, `<specs_dir>/<slug>/plan/plan.digest.md`, and when required `<specs_dir>/<slug>/plan/data-model.md`, `<specs_dir>/<slug>/plan/contracts/*`, `<specs_dir>/<slug>/plan/research.md`.
Stop if: inspect.md is present and not `pass`, the goal is ambiguous, or planning would require inventing requirements/AC.

## Rules

- Plan must preserve spec intent: no new major workstreams outside `spec.md`.
- Record only implementation-critical decisions: surfaces, sequencing, risks, trade-offs (`DEC-*`).
- Always use `.speckeep/templates/plan.md` as the skeleton and output format (and `.speckeep/templates/data-model.md` when needed). Do not look for “examples” in neighboring plan packages from other slugs: reading other plans for shape is wasted tokens and scope drift.
- If the data model does not change, still create `plan/data-model.md` with an explicit `status: no-change` + rationale.
- Create `plan/research.md` only when needed (e.g., external dependency/integration boundary, multiple realistic implementation options, or a high-risk unknown). Do not create `research.md` for generic brainstorming.
- Minimum context: current slug only; narrow repo reads (no full-repo scans).
- If `./.speckeep/scripts/check-plan-ready.*` exists, run it (slug first) before writing.

## Output expectations

- **Write `<specs_dir>/<slug>/plan/plan.digest.md` first** (mandatory, always, even on patch). One line per `DEC-*`, then a `Surfaces:` block (surface name + key files, one line each). If no `DEC-*` entries exist, write `Surfaces:` only. Example:
  ```
  DEC-001: use postgres, not sqlite, for persistence
  DEC-002: REST over GraphQL for public API surface
  Surfaces:
  - auth-handler: internal/auth/handler.go, internal/auth/middleware.go
  - user-store: internal/store/user.go
  ```
- Write/patch `<specs_dir>/<slug>/plan/plan.md` (create additional artifacts only when justified).
- Summarize key `DEC-*`, surfaces, sequencing constraints, and risks.
- Include a short summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`.
- Final line: `Ready for: /speckeep.tasks <slug>`
