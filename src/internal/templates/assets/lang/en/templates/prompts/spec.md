# SpecKeep Spec Prompt

You are creating or updating one feature spec.

## Goal

Produce a clear feature specification at `.speckeep/specs/<slug>/spec.md` that is compliant with the constitution.

Before writing or updating the spec, ensure work is happening on the feature branch for `<slug>`. The default branch naming convention is `feature/<slug>`.
If the repo is under git, do this explicitly and reproducibly:
- Determine the target branch: `--branch <name>` (when provided) else `feature/<slug>`
- If you are not on the target branch: switch to it, and if it does not exist, create and switch (`git switch <branch>` or `git switch -c <branch>`)
- If switching/creating the branch is not possible (no git, detached HEAD, environment constraints), stop and report the reason briefly before changing files

## Phase Contract

Inputs: `.speckeep/constitution.md`, user request, minimal repo context.
Outputs: `.speckeep/specs/<slug>/spec.md` (created or patched).
Stop if: goal ambiguous, multiple features in one request, or AC would be invented rather than derived.

## Flags

`--amend`: targeted edit mode — update one section, add a single requirement or acceptance criterion, or adjust scope without rewriting the entire spec.
When `--amend` is present in the user arguments:
- Read the existing spec first
- Apply only the change described in the remaining arguments
- Do not restructure or rewrite sections that are not being changed
- If the amendment materially changes acceptance criteria, scope, or the feature goal, the existing inspect report is invalidated: delete `.speckeep/specs/<slug>/inspect.md` and `.speckeep/specs/<slug>/summary.md` after writing the spec, and end with `Ready for: /speckeep.inspect <slug>`. Do not preserve a stale inspect report.
- Any new or amended acceptance criterion MUST use Given/When/Then format with a stable `AC-*` ID — the same rule as initial spec creation. Do not add ACs without GWT structure even in amend mode.

## Operating Mode

- Work on exactly one feature.
- Prefer patching an existing spec over rewriting it.
- Load only the minimum context needed to remove ambiguity.
- Do not drift into planning or implementation.
- When `/.speckeep/scripts/check-spec-ready.*` is available, run it as a readiness check before writing the spec. If the slug is already known, pass it as an argument to enable the branch check.

## Load First

- `.speckeep/constitution.md`
- the current user request and conversation
- the smallest amount of repository context needed to remove ambiguity

## Do Not Read By Default

- implementation-heavy code areas unless they are needed to define scope correctly
- contracts or data models for other features

## Stop Conditions

Stop and ask a minimal follow-up question if:

- the feature goal is ambiguous
- the request asks to derive one spec from multiple constitutional changes without naming a single concrete feature or change
- the request combines multiple features or unrelated changes into one spec
- the request would require multiple feature slugs or multiple independent specs to stay honest about scope
- the input is a prompt file with a generic filename and no explicit `name:` or `slug:` metadata, and the user did not provide `--name` or `--slug`
- the input looks like a URL rather than a concrete feature title
- acceptance criteria would be invented rather than derived
- the requested feature appears to conflict with the constitution
- any required section would be left as TBD or placeholder text

If the user provided `--name` but has not yet given enough feature detail, do not lose the request context: ask for the missing description and treat the next non-command user message as the continuation of the same spec request.

If clarification is needed, prefer a tiny structured clarify pass instead of a broad open-ended interview:

- ask at most 1-3 questions
- ask only about gaps that would otherwise force invented acceptance criteria, unclear scope boundaries, or ambiguous success conditions
- prefer coverage-based questions such as missing scenario, constraint, actor, or edge-condition clarification
- once the answers are sufficient, patch the spec immediately instead of starting a separate clarification workflow

Do not continue into planning or implementation thinking when the spec itself is still unclear.

If the spec already exists and is current, say so and do not modify the file.

## Invariants

- The spec MUST comply with the constitution.
- Keep the spec focused on one feature or change.
- Never load unrelated feature artifacts to compensate for unclear requirements.
- Derive acceptance criteria from the request and repository reality; do not invent them.
- Use explicit scope boundaries. The out-of-scope section is mandatory.
- The spec should be detailed enough that both an agent and a human reviewer can understand the user flow and scope boundaries without reading planning artifacts.
- Do not write planning decisions, task decomposition, or implementation steps in the spec itself.
- Do not lock in technologies, libraries, framework choices, or version details by default.
- Mention stack or version constraints only when they are explicitly required by the user, needed to reflect an existing repository constraint, or required by an external or compatibility contract that changes acceptance scope.
- If a technology choice matters only as an implementation preference, record it in `plan`, not in `spec`.
- Do not mix languages inside the same spec without a strong project reason.
- Follow `.speckeep/templates/spec.md` when creating a new file.
- Every acceptance criterion MUST use Given/When/Then format.
- Every acceptance criterion MUST have a stable ID such as `AC-001`.
- Ask follow-up questions only when the missing information is critical.
- When a requirement or AC detail is unclear but the spec can still proceed, mark it inline with `[NEEDS CLARIFICATION: what is unknown and why it matters]` instead of blocking the entire spec. Inspect will flag these as errors that must be resolved before planning.
- `## Assumptions` is mandatory. Record reasonable defaults chosen when the feature description did not specify a detail, environmental assumptions, and dependencies on existing systems. Making assumptions explicit lets inspect catch wrong ones early.
- `## Success Criteria` with `SC-*` IDs is optional. Include it only when the feature has measurable performance, reliability, or user-experience targets that go beyond behavioral correctness (e.g., latency, throughput, error rate, task completion time). Omit for purely behavioral features.

## Resolution Rules

- `/speckeep.spec` may receive `--name <feature name>`, optional `--slug <feature-slug>`, and optional `--branch <branch-name>`.
- If `--name` is present, use it as the canonical feature name.
- If `--slug` is present, use it for the spec path.
- When `/speckeep.spec` starts from a prompt file, prefer explicit metadata at the top of the file:
  - `name: <feature name>`
  - optional `slug: <feature-slug>`
- Command-style `--name` and `--slug` arguments take priority over `name:` and `slug:` in a prompt file.
- If `slug:` is present, use it for the spec path and feature branch.
- If `--name` is present without `--slug`, derive `<slug>` from `--name`.
- If only `name:` is present, derive `<slug>` from it.
- Fall back to the file basename only when it is specific enough to produce a safe slug.
- If the user explicitly provides `--branch <name>`, use that branch name as-is instead of the default `feature/<slug>`.
- An explicit `--branch` override does not change the spec slug unless the user also requests a different slug.
- If the user provided only `--name` and the detailed description is still missing, ask for the description and keep staged mode active for the next non-command user message.
- If the next user message begins with `/speckeep.`, staged mode is canceled and the new slash-command takes priority.
- If the next user message does not begin with `/speckeep.`, treat it as the continuation of the staged spec request.

## Language Rules

- Use the project's configured documentation language for new or updated spec content.
- Respect an established local document convention only when preserving an existing file would otherwise become inconsistent.
- Do not introduce mixed-language headings or sections in the same spec without a strong project reason.

## Acceptance Rules

- The `Given`, `When`, and `Then` markers remain canonical regardless of the documentation language:
  - **Given** — the initial state or precondition
  - **When** — the action or event
  - **Then** — the expected observable outcome
- Acceptance criteria should be observable and testable.
- Prefer a small set of strong criteria over a long redundant list.
- Each `AC-*` should explain why it matters in one short line when that context helps downstream planning stay grounded.
- Each `AC-*` should include evidence or an observable proof signal, not just a generic desired state.

## Content Quality Rules

- `## Goal` should explain who benefits, what changes, and what success looks like.
- `## Why Now` should capture why this change matters now when timing, pain, or business pressure materially affects prioritization.
- `## Primary User Flow` should describe the main path in 3-5 concrete steps, not generic prose.
- `## Change Delta` should make it obvious what becomes newly possible, what changes, and what stays unchanged.
- `## Affected Surfaces` should stay compact and name only the user-visible or repository-visible surfaces that define the feature boundary.
- `## Scope Snapshot`, `## Scope`, and `## Non-Goals` should make the feature boundary obvious to a reviewer.
- `## Context` should capture repository constraints, preserved behavior, and integration points that materially affect the feature.
- `## Assumptions` should list every non-obvious assumption the spec relies on — environment, user behavior, system state, existing service stability. Bad: omitting assumptions entirely. Good: `Users have stable network; export API is available and returns <2s for 10k records`.
- `## Success Criteria` (when present) should define measurable outcomes separate from behavioral AC. Each `SC-*` must have a number and a measurement method. Bad: `system should be fast`. Good: `SC-001 Dashboard loads in <1s for 95th percentile on 100 concurrent users`.
- `## Requirements` should stay clear and testable. Bad: `support CSV export properly`. Good: `RQ-001 Export returns a valid CSV file with UTF-8 BOM header; rows match the current filter`.
- `## Edge Cases` should include only behavior that materially changes implementation or validation. Bad: `think about what happens with large files`. Good: `Export with >50k rows streams to disk instead of buffering in memory; user sees a progress indicator`.
- `## Open Questions` should say `none` when no real question remains.
- Negative examples: do not merge multiple features into one spec, do not hide scope expansion inside edge cases, and do not use `TBD` acceptance criteria.
- Negative examples: do not add library lists, framework choices, SDK names, or version pins to the spec unless they are product or repository constraints.
- Prefer density over length: every section should help planning or review, and filler text is a defect.

## Output expectations

- **Before writing any file**: create or switch to `feature/<slug>` (or the explicit `--branch` value). This step is mandatory and must happen first — do not skip it even if the spec file already exists.
- Write or patch `.speckeep/specs/<slug>/spec.md`
- Summarize goal, scope, acceptance criteria, and open questions
- End with a summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`
- When ready: `Ready for: /speckeep.inspect <slug>`

## Self-Check

- Did I switch to or create the feature branch before writing any file?
- Did I stay within one feature?
- Did every acceptance criterion get a stable ID and Given/When/Then form?
- Would a human reviewer understand the primary user flow and scope boundary without additional explanation?
