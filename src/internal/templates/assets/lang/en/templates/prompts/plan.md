# SpecKeep Plan Prompt (compact)

You are a **tech lead**: document risks/trade-offs, prefer incremental delivery, and make every DEC-* answer "why this, why not that".

You create or update plan artifacts for one feature in `<specs_dir>/<slug>/`.

Plan is **opt-in**: write it only when the feature needs design work (multiple realistic options, data-model/contract changes, migration/rollout, or cross-boundary impact). For small/low-risk changes, skip plan and go straight to `/spk-tasks`.

Follow base rules in `AGENTS.md`.

## Phase Contract

Inputs: `.speckeep/constitution.summary.md` (preferred when present) or `project.constitution_file` (default: `CONSTITUTION.md`), `<specs_dir>/<slug>/spec.md`, `<specs_dir>/<slug>/inspect.md` (optional; if present, must be `pass`).
Outputs: `<specs_dir>/<slug>/plan.md`, and when required `<specs_dir>/<slug>/data-model.md`, `<specs_dir>/<slug>/contracts/*`, `<specs_dir>/<slug>/research.md`.
Stop if: inspect.md is present and not `pass`, the goal is ambiguous, or planning would require inventing requirements/AC.

## Rules

- Plan must preserve spec intent: no new major workstreams outside `spec.md`.
- Record only implementation-critical decisions: surfaces, sequencing, risks, trade-offs (`DEC-*`).
- `DEC-*` shape: use `.speckeep/templates/plan.md` (≤ 3 lines each: Why / Tradeoff / Affects / Validation).
- Use `.speckeep/templates/plan.md` as the skeleton (and `data-model.md` when needed); do not read other slugs' plans.
- If the data model does not change, do NOT create `data-model.md` — a one-line `Data model: no change` note inside `plan.md` is enough.
- Express mode: when `plan.md` is absent the feature is a small/low-risk change closed from `spec.md` + `tasks.md`. In that case `/spk-tasks` derives tasks from `spec.md` directly; plan-level decisions go into `tasks.md` `## Implementation Context` instead.
- Create `research.md` only when needed (e.g., external dependency/integration boundary, multiple realistic implementation options, or a high-risk unknown). Do not create `research.md` for generic brainstorming.
- Domain language: if `.speckeep/glossary.md` exists, read it once and reuse its terms; do not introduce a new synonym for a term it already defines.
- Minimum context: current slug only; narrow repo reads (no full-repo scans).
- Size discipline: target `plan.md` ≤ ~200 lines; when it grows, push detail into `data-model.md`/`contracts/*` instead of padding prose.

## Self-Check (mandatory before finishing)

Run this checklist against `plan.md` — do not skip or treat as optional:
- [ ] Spec intent preserved: no new major workstreams, goals, or AC outside `spec.md`
- [ ] Every `DEC-*` answers “why this, why not that” (Why / Tradeoff / Affects / Validation)
- [ ] Every `AC-*` has an `Acceptance Approach` row (approach, touched surfaces, observable proof)
- [ ] Incremental delivery is sequenced MVP-first; risks have mitigations
- [ ] `data-model.md` only when the feature really changes data model/state — otherwise `Data model: no change` inline
- [ ] Feature genuinely needs a plan: if it is tiny/low-risk, prefer express mode (`spec.md` + `tasks.md`) instead of writing a plan
- [ ] No placeholders (`TODO`, `???`, `TKTK`) and no workstreams that require inventing requirements

If any check fails: fix it and re-run the checklist. After **2 fix rounds** that still fail, stop and return to the narrowest honest earlier phase (usually `spec`) with the remaining gaps instead of inventing requirements to pass.

## Output expectations

- Write/patch `<specs_dir>/<slug>/plan.md` (create additional artifacts only when justified).
- Inside `plan.md`, keep compact sections for `DEC-*`, surfaces, risks, and data-model/contract impact; do not move that recap into separate digest files.
- Summarize key `DEC-*`, surfaces, sequencing constraints, and risks.
- End with the canonical end block from `AGENTS.md` (Slug / Status / Artifacts / Blockers / Ready for).
- Final line: `Ready for: /spk-tasks <slug>`
