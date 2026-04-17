# SpecKeep Constitution Prompt

You are creating or updating the project's constitution file (configured via `project.constitution_file`; default: `.speckeep/constitution.md`).

Follow base rules in `AGENTS.md` (paths, git, load discipline, readiness scripts, language, phase discipline).

## Goal

Produce a strict project constitution that is authoritative for both humans and development agents.

For a **greenfield project** (from scratch): establish immutable architectural boundaries, select the tech stack, and define quality standards serving as the basis for downstream "Foundation" design.

For an **existing codebase**: formalize the project's observable reality first, then separately codify any new mandatory rules explicitly requested by the user.

## Greenfield mode

- codify the selected tech stack (language, frameworks, database)
- define core architectural patterns (Clean Architecture, Hexagonal, etc.)
- establish naming conventions and directory structure
- define dependency management and external integration rules
- fill `## Tech Stack` and `## Core Architecture` with enough detail to anchor downstream specs/plans

## Brownfield mode

When the project exists, work in two layers:

- `Observed Reality` — record only what is supported by repo structure, config, key entrypoints, dependencies, existing docs
- `Declared Law` — record new mandatory rules only when explicitly requested by the user or already strongly grounded

Do not redesign an existing project into an idealized architecture. Describe current reality first, then formalize how future changes must be governed.

## Load First

- current user request and conversation
- `.speckeep/constitution.md`
- top-level directory structure
- only the smallest repo context needed to make the constitution concrete

## Load If Present

- `README.md`, `AGENTS.md`
- project manifests / config files explaining language, runtime, architectural boundaries, or integrations

## Repository Evidence

**Strong signals** (safe to derive rules from): directory boundaries (`api/`, `cmd/`, `internal/`, `migrations/`), dependencies/config revealing transports/storage/runtime, existing workflow docs, entrypoint files showing composition root or role separation.

**Weak signals** (context only, not rule sources): isolated files without structural support, naming not confirmed by config, generic best-practice expectations not grounded in the repo.

## Do Not Read By Default

- large code areas not affecting the constitution
- old feature artifacts unless needed to resolve a constitutional conflict
- the whole repository by default

## Stop Conditions

Stop and ask one minimal follow-up if:

- the project purpose cannot be stated concretely
- workflow rules would be guessed rather than grounded
- a constitutional conflict is visible but cannot be resolved from available context
- architecture boundaries/ownership/workflow would have to be declared as mandatory without evidence

Do not broaden reading unless one of these conditions is met.

If the constitution is already current and does not conflict with the request, say so and do not modify the file.

## Required behavior

- Patch the existing `.speckeep/constitution.md`.
- Preserve these mandatory sections exactly:
  - `## Purpose`, `## Core Principles`, `## Constraints`, `## Tech Stack`, `## Core Architecture`, `## Decision Priorities`, `## Key Quality Dimensions`, `## Language Policy`, `## Development Workflow`, `## Governance`, `## Exceptions Protocol`, `## Last Updated`
- `## Core Principles` MUST have ≥5 subsections using `### Principle Name` headings.
- Extra sections allowed when they materially improve governance.
- Replace placeholder tokens like `[PROJECT_NAME]`, `[TECH_STACK]`, `[ARCHITECTURE]` with concrete text.
- `--foundation` flag: treat the constitution as the definitive record of what the project is built on. Deeply fill `## Tech Stack` and `## Core Architecture` so they are actionable downstream:
  - languages, frameworks, databases, infrastructure — with version constraints where relevant
  - structural patterns, data flows, module boundaries, directory organization
  - non-negotiable constraints for all future specs/plans/implementation
- Brownfield: codify what the codebase already demonstrates before introducing new norms.
- User-requested new rules → encode in `## Development Workflow` and `## Governance` as mandatory.
- `## Development Workflow` MUST define how feature branches, specs, inspect, plans, tasks, and implementation relate to constitutional compliance.
- `## Development Workflow` MUST state the conditions under which a spec, inspect, plan, tasks, or implementation violates the constitution and cannot proceed.
- `## Decision Priorities` MUST capture 3–5 short rule-like priorities for resolving trade-offs (simplicity vs extensibility, correctness vs delivery speed, maintainability vs cleverness).
- `## Key Quality Dimensions` MUST include only project-relevant dimensions — 3–5 short testable bullets, not a generic essay.
- `## Exceptions Protocol` MUST explain how acceptable deviations are recorded and when downstream phases treat a conflict as a blocker.
- Do not declare DDD boundaries, event-contract ownership, release policy, or branch strategy as mandatory unless repository-grounded or explicitly requested.
- Use strict, testable language. Each principle must make it possible to answer: "does this decision conform to the constitution?"
- Do not turn `## Decision Priorities`, `## Key Quality Dimensions`, or `## Exceptions Protocol` into a long handbook. Prefer compact bullets useful for downstream phase checks.
- The constitution is the highest-priority project document. Specs, inspect reports, plans, tasks, and implementation must conform.

## Update rules

- Keep existing good principles unless they conflict with new requirements.
- Prefer patching over rewrites. When refining a principle, preserve testability; do not replace concrete rules with vague generalizations.
- Repository-inferred rule → phrase as observed stable norm. User-intent rule → phrase as law for downstream phases.
- Update `## Last Updated` with today's date in `YYYY-MM-DD` whenever the constitution changes.

## Summary artifact

After writing/patching `constitution.md`, also write/update `.speckeep/constitution.summary.md` with only:

- `## Purpose` — 1–2 sentences
- `## Key Constraints` — 3–5 bullets, hard non-negotiable limits only
- `## Language Policy` — 3 lines: docs, agent, comments
- `## Development Workflow` — 3–5 key rules most relevant to spec/plan/implement
- `## Decision Priorities` — 3–5 bullets

Keep the summary under 60 lines. It is loaded by `implement`/`tasks`/`verify` instead of the full constitution. It does not substitute the full constitution in phases requiring constitutional consistency checks (`spec`, `inspect`, `plan`).

## Post-Update Impact Check

After writing/patching `constitution.md`, check whether active specs may now conflict:

- Run `.speckeep/scripts/list-specs.*` to enumerate active specs; fallback: list files in `.speckeep/specs/`.
- For each active spec, compare `## Goal`, `## Scope`, `## Acceptance Criteria` against changed sections.
- Flag conflicts as `NEEDS RE-INSPECT: <slug> — <reason>`. Do not modify specs — only report.
- If no active specs or none conflict: "No active specs affected by this update."

## Output

- Write updated `.speckeep/constitution.md` and `.speckeep/constitution.summary.md`.
- Summarize changes and unresolved questions; note what was inferred from the codebase vs. added as new law.
- Mark unresolved constitutional questions as **BLOCKER** for downstream phases.
- When ready: `Ready for: /speckeep.spec`.
