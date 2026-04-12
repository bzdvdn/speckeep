# SpecKeep Constitution Prompt

You are creating or updating the project's constitution file (configured via `project.constitution_file`; default: `.speckeep/constitution.md`).

## Goal

Produce a strict project constitution that is authoritative for both humans and development agents.

## Path Resolution

Paths in this prompt use the default workspace layout. If `.speckeep/speckeep.yaml` overrides `project.constitution_file`, `paths.specs_dir`, or `paths.archive_dir`, always follow the configured paths instead of the defaults shown here.
Read `.speckeep/speckeep.yaml` at most once per session to resolve these paths; do not re-read it unless it changed or a path is ambiguous.

For a **greenfield project** (from scratch), focus on establishing immutable architectural boundaries, selecting the technology stack, and defining quality standards that will serve as the basis for the subsequent "Foundation" design.

For an existing codebase, formalize the project's observable reality first, then separately codify any new mandatory rules explicitly requested by the user.

## Greenfield mode

When starting a project from scratch:

- codify the selected technology stack (language, frameworks, database)
- define core architectural patterns (e.g., Clean Architecture, Hexagonal)
- establish naming conventions and directory structure
- define dependency management and external integration rules
- fill `## Tech Stack` and `## Core Architecture` sections with enough detail to serve as the foundation for downstream spec and plan phases

## Brownfield mode

When the project already exists, work in two layers:

- `Observed Reality`
  - record only what is supported by repository structure, configuration, key entrypoints, dependencies, and existing documentation
- `Declared Law`
  - record new mandatory development rules only when they are explicitly requested by the user or already strongly grounded in the project

Do not redesign an existing project into an idealized architecture. Describe current reality first, then formalize how future changes must be governed.

## Load First

- current user request and conversation
- `.speckeep/constitution.md`
- top-level directory structure
- only the smallest amount of repository context needed to make the constitution concrete

## Load If Present

- `README.md`
- `AGENTS.md`
- project manifests and configuration files that quickly explain language, runtime, architectural boundaries, or integrations

## Repository Evidence

**Strong signals** (safe to derive rules from): directory boundaries (`api/`, `cmd/`, `internal/`, `migrations/`), dependencies and config revealing transports/storage/runtime, existing workflow docs, entrypoint files showing composition root or role separation.

**Weak signals** (context only, not rule sources): isolated files without structural support, naming not confirmed by config, general best-practice expectations not grounded in the repo.

## Do Not Read By Default

- large code areas that do not affect the constitution
- old feature artifacts unless they are required to resolve a constitutional conflict
- the whole repository by default

## Stop Conditions

Stop and ask a minimal follow-up question if:

- the project purpose cannot be stated concretely
- the development workflow rules would be guessed rather than grounded
- a constitutional conflict is visible but cannot be resolved from available context
- architecture boundaries, ownership, or workflow would have to be declared as mandatory without enough evidence

Do not broaden repository reading unless one of these conditions is met.

If the constitution is already current and does not conflict with the request, say so and do not modify the file.

## Required behavior

- Work by patching the existing `.speckeep/constitution.md` file.
- Preserve these mandatory sections exactly:
  - `## Purpose`
  - `## Core Principles`
  - `## Constraints`
  - `## Tech Stack`
  - `## Core Architecture`
  - `## Decision Priorities`
  - `## Key Quality Dimensions`
  - `## Language Policy`
  - `## Development Workflow`
  - `## Governance`
  - `## Exceptions Protocol`
  - `## Last Updated`
- Ensure there are at least 5 principle subsections under `## Core Principles` using `### Principle Name` headings.
- You may add extra sections when they materially improve project governance.
- Replace placeholder tokens like `[PROJECT_NAME]`, `[TECH_STACK]`, or `[ARCHITECTURE]` with concrete text.
- When the `--foundation` flag is used, treat the constitution as the definitive record of what the project is built on and how it is structured. Deeply fill the `## Tech Stack` and `## Core Architecture` sections so they are actionable for all downstream phases. These sections must describe:
  - Selected languages, frameworks, databases, and infrastructure — with version constraints where relevant.
  - Structural patterns, data flows, module boundaries, and directory organization.
  - Any constraints that are non-negotiable for all future specs, plans, and implementation decisions.
- For a brownfield project, codify what the codebase already demonstrates before introducing new mandatory norms.
- If the user explicitly requests new development rules, encode them in `## Development Workflow` and `## Governance` as mandatory rules for future work.
- The `## Development Workflow` section MUST define how feature branches, specs, inspect, plans, tasks, and implementation relate to constitutional compliance.
- The `## Development Workflow` section MUST explicitly state the conditions under which a spec, inspect, plan, tasks, or implementation violates the constitution and cannot proceed.
- The `## Decision Priorities` section MUST capture 3-5 short, rule-like priorities for resolving trade-offs such as simplicity vs extensibility, correctness vs delivery speed, or maintainability vs cleverness.
- The `## Key Quality Dimensions` section MUST include only project-relevant quality dimensions. Do not write a generic quality essay; keep it to 3-5 short, testable bullets.
- The `## Exceptions Protocol` section MUST explain how acceptable deviations from the constitution are recorded and when downstream phases should treat a conflict as a blocker.
- Do not declare DDD boundaries, event-contract ownership, release policy, or branch strategy as mandatory unless they are repository-grounded or explicitly requested by the user.
- If critical information is missing, ask only the minimum necessary follow-up questions.
- Use strict, testable language. Avoid vague wording. Each principle must make it possible to answer concretely: "does this decision conform to the constitution?"
- Do not turn `## Decision Priorities`, `## Key Quality Dimensions`, or `## Exceptions Protocol` into a long handbook. Prefer compact bullets that are useful for downstream phase checks.
- The constitution is the highest-priority project document. Specs, inspection reports, plans, tasks, and implementation must conform to it.

## Update rules

- Keep existing good principles unless they conflict with new requirements.
- Prefer patching and refinement over full rewrites. When refining a principle, preserve its testability; do not replace concrete rules with vague generalizations.
- If a rule is inferred from the repository, phrase it as an observed stable norm of the project rather than an abstract best practice.
- If a rule is introduced by user intent, phrase it as law for downstream phases.
- Update `## Last Updated` with today's date in `YYYY-MM-DD` format whenever the constitution changes.

## Summary artifact

After writing or patching `constitution.md`, also write or update `.speckeep/constitution.summary.md`.

The summary MUST contain only:

- `## Purpose` — 1-2 sentences
- `## Key Constraints` — 3-5 bullets, hard non-negotiable limits only
- `## Language Policy` — 3 lines: docs, agent, comments
- `## Development Workflow` — 3-5 key rules most relevant to spec, plan, and implement phases
- `## Decision Priorities` — 3-5 bullets

Keep the summary under 60 lines. It is loaded by `implement`, `tasks`, and `verify` instead of the full constitution to reduce context overhead. The summary is not a substitute for the full constitution in phases that require constitutional consistency checks (spec, inspect, plan).

## Post-Update Impact Check

After writing or patching `constitution.md`, check whether existing active specs may now conflict with the changed rules:

- Run `.speckeep/scripts/list-specs.*` to enumerate active specs. If the script is unavailable, list files in `.speckeep/specs/` directly to identify active slugs.
- For each active spec, compare its `## Goal`, `## Scope`, and `## Acceptance Criteria` against the changed constitutional sections.
- If a spec may conflict with the updated constitution, flag it as `NEEDS RE-INSPECT: <slug> — <reason>` in the output.
- Do not modify the specs themselves — only report which ones need re-inspection.
- If no active specs exist or none conflict, state: "No active specs affected by this update."

## Output expectations

- Write the updated `.speckeep/constitution.md` and `.speckeep/constitution.summary.md`
- Briefly summarize what changed and what remains unresolved
- Note what was inferred from the codebase and what was added as new mandatory law
- Mark unresolved constitutional questions as **BLOCKER** for downstream phases
- When ready: `Ready for: /speckeep.spec`
