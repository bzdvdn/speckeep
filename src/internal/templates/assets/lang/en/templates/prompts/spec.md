# SpecKeep Spec Prompt

You are creating or updating one feature spec.

## Goal

Produce a feature spec at `<specs_dir>/<slug>/spec.md` (default: `.speckeep/specs/<slug>/spec.md`) that complies with the constitution.

## Phase Contract

Inputs: `.speckeep/constitution.md`, user request, minimal repo context.
Outputs: `.speckeep/specs/<slug>/spec.md` (created or patched).
Stop if: goal ambiguous, multiple features in one request, or AC would be invented rather than derived.

If `.speckeep/speckeep.yaml` overrides `paths.specs_dir` or `project.constitution_file`, follow the configured paths. Read it at most once per session.

Before writing any file: create or switch to `feature/<slug>` (or the explicit `--branch`). Use `git switch <branch>` or `git switch -c <branch>`. If switching is impossible (no git, detached HEAD, environment constraints), stop and report the reason before changing files.

## Flags

`--name <feature name>`, optional `--slug`, optional `--branch`.

- `--name` is the canonical feature name. Prompt-file metadata (`name:`, `slug:`) is accepted but command-line flags take priority.
- If `--slug` is missing, derive it from `--name` (or from `name:` in the prompt file).
- An explicit `--branch` overrides `feature/<slug>` but does not change the slug.
- Fall back to the prompt filename only when it is specific enough to produce a safe slug.
- If only `--name` is given without description, ask for the description and keep staged mode active for the next non-command user message.
- If the next user message begins with `/speckeep.`, staged mode is canceled and the new slash-command takes priority.
- If the next user message does not begin with `/speckeep.`, treat it as the continuation of the staged spec request.

`--amend`: targeted edit mode — update one section or add a single requirement / `AC-*` without rewriting the spec.

- Read the existing spec first; change only what was requested, do not restructure untouched sections.
- Any new or amended AC MUST use Given/When/Then with a stable `AC-*` ID.
- If the amendment materially changes AC, scope, or goal, the existing inspect report is invalidated: delete `.speckeep/specs/<slug>/inspect.md` and `.speckeep/specs/<slug>/summary.md`, and end with `Ready for: /speckeep.inspect <slug>`.

`--greenfield`: greenfield-first mode.

- Use when the repository is empty, nearly empty, or the feature defines one of the first deliverable product slices.
- In this mode, prefer user-story and MVP-slice clarity over repository-surface detail that does not exist yet.
- Add `## User Stories`, `## MVP Slice`, and `## First Deployable Outcome` when they materially help downstream planning.

## Load and Scope

Load: `.speckeep/constitution.md`, the user request, and the smallest repo context needed to remove ambiguity.
Do not load by default: implementation-heavy code, or other features' contracts / data models.
When `/.speckeep/scripts/check-spec-ready.*` is available, run it as the readiness check before writing the spec; pass the slug when known.

## Stop Conditions

Stop and ask a minimal follow-up if:

- the goal is ambiguous or combines multiple features into one spec
- the request would require multiple feature slugs or multiple independent specs to stay honest about scope
- the input is a prompt file with a generic filename and no `name:` / `slug:` metadata, and the user did not provide `--name` / `--slug`
- the input looks like a URL rather than a concrete feature title
- acceptance criteria would be invented rather than derived
- the feature conflicts with the constitution
- any required section would be left as TBD or placeholder text

When clarification is needed, prefer a tiny structured clarify pass instead of a broad open-ended interview: at most 1–3 targeted questions about gaps that would otherwise force invented AC, unclear scope, or ambiguous success. In `--greenfield` mode, spend that budget first on the MVP slice, first deployable outcome, and core user-flow ambiguity. Patch the spec as soon as answers are sufficient.

If the spec already exists and is current, say so and do not modify the file.

## Invariants

- The spec MUST comply with the constitution. Work on exactly one feature. Prefer patching over rewriting.
- Do not write planning decisions, task decomposition, or implementation steps in the spec itself.
- Derive AC from the request and repo reality; do not invent them. Every AC MUST use Given/When/Then and have a stable `AC-*` ID.
- `## Out of Scope` is mandatory. `## Assumptions` is mandatory — record env, user behavior, system state, and dependency assumptions.
- `## Success Criteria` (`SC-*`) is optional: include only for measurable targets (latency, throughput, error rate, task completion time).
- When a detail is unclear but the spec can still proceed, mark it inline with `[NEEDS CLARIFICATION: what is unknown and why]`. Inspect will flag these as errors to resolve before planning.
- Do not lock in technologies, libraries, framework choices, or version details by default.
- Mention stack or version constraints only when explicitly required by the user, by an existing repository constraint, or by an external / compatibility contract that changes acceptance scope.
- If a technology choice matters only as an implementation preference, record it in `plan`, not in `spec`.
- The spec should be detailed enough that both an agent and a human reviewer can understand the user flow and scope boundaries without reading planning artifacts.
- Do not run git commit/git push/git tag or open a PR unless the user explicitly asks.
- Follow `.speckeep/templates/spec.md` when creating a new file. Use the project's configured documentation language; do not mix languages inside the same spec.

## Acceptance Rules

`Given`, `When`, `Then` markers stay canonical regardless of the documentation language:

- **Given** — precondition
- **When** — action or event
- **Then** — expected observable outcome

AC must be observable and testable. Prefer a small set of strong criteria over a long redundant list. Each `AC-*` should include an observable proof signal, and when it helps downstream planning, one short line on why it matters.

## Content Quality Rules

- `## Goal` should explain who benefits, what changes, and what success looks like.
- `--greenfield`: add `## User Stories` only when it reduces ambiguity; keep it to 1–3 prioritized stories (`P1/P2/P3`) tied to the MVP, not a full backlog.
- `--greenfield`: `## MVP Slice` should say what the first independently useful slice is and which `AC-*` it must satisfy.
- `--greenfield`: `## First Deployable Outcome` should state what can be demonstrated or manually validated after the first implementation pass.
- `## Why Now` captures timing, pain, or business pressure when it materially affects prioritization.
- `## Primary User Flow` should describe the main path in 3-5 concrete steps, not generic prose.
- `## Change Delta` should make it obvious what becomes newly possible, what changes, and what stays unchanged.
- `## Affected Surfaces` names only the user-visible or repo-visible surfaces that define the feature boundary.
- `## Scope Snapshot`, `## Scope`, and `## Non-Goals` should make the feature boundary obvious to a reviewer.
- `## Context` captures repository constraints, preserved behavior, and integration points that materially affect the feature.
- `## Assumptions` lists every non-obvious assumption — environment, user behavior, system state, service stability. Omitting it is a defect.
- `## Success Criteria` (when present) gives each `SC-*` a number and a measurement method (e.g., `SC-001 Dashboard loads in <1s at p95 for 100 concurrent users`).
- `## Requirements` should be clear and testable (e.g., `RQ-001 Export returns a valid CSV with UTF-8 BOM; rows match the current filter`), not vague wording like "support X properly".
- `## Edge Cases` includes only behavior that materially changes implementation or validation.
- `## Open Questions` says `none` when no real question remains.
- Do not merge multiple features into one spec; do not hide scope expansion inside edge cases; do not leave AC as `TBD`; do not add library lists, framework choices, SDK names, or version pins to the spec unless they are product or repository constraints.
- Prefer density over length: every section should help planning or review; filler text is a defect.

## Output

- Write or patch `.speckeep/specs/<slug>/spec.md` after confirming the feature branch is checked out.
- Summarize goal, scope, acceptance criteria, and open questions.
- End with a summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`.
- When ready: `Ready for: /speckeep.inspect <slug>`.

## Self-Check

- Am I on the feature branch, inside one feature, and does every AC have a stable ID + Given/When/Then?
- Would a reviewer understand the primary user flow and scope boundary without extra explanation?
