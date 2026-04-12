# SpecKeep Tasks Prompt

You are creating or updating `<specs_dir>/<slug>/plan/tasks.md` (default: `.speckeep/specs/<slug>/plan/tasks.md`).

## Goal

Break an approved plan into executable implementation tasks.

## Path Resolution

Paths in this prompt use the default workspace layout. If `.speckeep/speckeep.yaml` overrides `paths.specs_dir` or `project.constitution_file`, always follow the configured paths instead of the defaults shown here.
Read `.speckeep/speckeep.yaml` at most once per session to resolve these paths; do not re-read it unless it changed or a path is ambiguous.

## Phase Contract

Inputs: `.speckeep/constitution.md`, `.speckeep/specs/<slug>/plan/plan.md`; optionally `.speckeep/specs/<slug>/summary.md`, `.speckeep/specs/<slug>/plan/data-model.md`, `.speckeep/specs/<slug>/plan/contracts/` when decomposition requires them.
Outputs: `.speckeep/specs/<slug>/plan/tasks.md` with phased task list and Acceptance Coverage section.
Stop if: plan.md missing, plan underspecified, or any AC cannot be mapped to executable work.

## Operating Mode

- Decompose one approved plan package only.
- Use `plan.md` as the entrypoint and go deeper only when required.
- Produce the smallest task list that still covers the feature safely.
- Prefer explicit sequencing over umbrella tasks.

## Load First

Always read these before decomposing the work:

- `.speckeep/constitution.summary.md` if present; otherwise `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/plan/plan.md`

## Load If Present

Read when a task being decomposed explicitly references or depends on content in these files:

- `.speckeep/specs/<slug>/summary.md` (or `spec.md`) — when an `AC-*` referenced by a task has ambiguous scope or acceptance boundary
- `.speckeep/specs/<slug>/plan/data-model.md` — when tasks must create or modify entities, fields, invariants, or state transitions
- `.speckeep/specs/<slug>/plan/contracts/` — when tasks must create or modify API endpoints, event payloads, or integration boundaries
- `.speckeep/specs/<slug>/plan/research.md` — only when a documented finding changes task sequencing or introduces a risk that affects decomposition
- code files — only the files needed to confirm where implementation surfaces begin and end (e.g., existing function signatures, module boundaries)

Do not assume `research.md` should exist; use it only when the plan clearly depends on preserved uncertainty, an external dependency, or a documented trade-off.

## Do Not Read By Default

- implementation files that are not needed to decompose the work
- broad repository history

## Flags

`--repair <task-id-list>`: targeted repair mode — fix specific tasks identified by verify or review without rewriting the full task list.

When `--repair` is present in the user arguments:
- Read the existing `tasks.md` first.
- Locate only the tasks named in the argument (e.g., `--repair T2.3,T3.1`).
- For each named task: update the description, outcome, `Touches:`, or `AC-*` mapping as needed.
- Do not restructure phases, renumber other tasks, or rewrite sections that are not being repaired.
- If the repair reveals that the plan itself is flawed, stop and suggest `/speckeep.plan <slug> --update` instead of silently expanding scope.
- Update the `## Acceptance Coverage` section only if the repaired tasks change coverage mapping.
- Update the `## Surface Map` section only if `Touches:` values changed.

## Stop Conditions

Stop and ask for refinement if:

- `.speckeep/specs/<slug>/plan/plan.md` is missing
- tasks would be vague because the plan is underspecified
- the current decomposition requires spec, data model, contracts, or research that are missing
- the constitution blocks the proposed decomposition
- the decomposition would span multiple feature slugs or unrelated change sets
- one or more acceptance criteria cannot be mapped to executable work without guessing

Do not jump ahead into implementation.

## Invariants

- Tasks MUST align with the plan and constitution.
- Use `plan.md` as the decomposition entrypoint.
- Never read unrelated feature artifacts to compensate for underspecified planning.
- Read implementation code only when the task list would otherwise stay vague; prefer a narrow file slice over broad exploration.
- When `/.speckeep/scripts/check-tasks-ready.*` is available, prefer running it as the phase readiness check instead of reading script source.
- Do not read `/.speckeep/scripts/*` by default unless you are debugging the scripts, working on SpecKeep itself, or the user explicitly asks to inspect script logic.
- The task list must be executable in order.
- Every acceptance criterion must be covered by at least one task.
- Prefer concrete, testable, implementation-oriented tasks.
- Include validation and documentation alignment work only when needed.
- Do not generate vague umbrella tasks.
- The task list should be readable to both an implementation agent and a human reviewer without extra interpretation.
- Targeted code reading during task decomposition is useful when it reduces re-reading during implementation.
- Do not start implementation work, edit source code, or claim tasks are already done during the tasks phase.

## Language Rules

- Use the project's configured documentation language for new or updated task content.
- Preserve an established local task-document convention only when needed to keep an existing file coherent.
- Do not mix task languages inside the same task list without a strong project reason.
- Load deeper artifacts only when the current decomposition needs them.

## Task Format Rules

- Follow the structure of `.speckeep/templates/tasks.md`: group tasks into ordered phases (`## Phase N: Name`).
- Each task MUST include a phase-scoped task ID in the form `T<phase>.<index>`.
- Each task MUST follow the format: `- [ ] T<phase>.<index> <action verb> — <concrete measurable outcome>`
- Each task SHOULD reference 1-2 stable IDs when possible (`AC-*`, `RQ-*`, `DEC-*`).
- Each task MUST include a `Touches:` field naming the concrete files or modules the task will create or modify. This is the primary mechanism for preventing re-reads during implement — the implement agent uses these to batch-read all needed files in one pass. Keep it compact (`Touches: src/auth/handler.ts, src/session/store.ts`). Use module-level names only when the exact file cannot be determined yet (`Touches: src/auth/`). A task without `Touches:` forces the implement agent into exploratory reads, which wastes tokens.
- The `## Surface Map` section MUST appear before the first phase: a two-column table mapping each unique implementation surface to the task IDs that touch it (`Surface | Tasks`). This is the implement agent's batch-read manifest — without it, the agent must scan every task line to build the read list.
- The tasks taken together MUST cover all acceptance criteria from the spec. Any uncovered criterion is a blocker.
- The `## Acceptance Coverage` section MUST include at least one explicit coverage line for each acceptance criterion.
- Coverage lines SHOULD reference stable acceptance IDs and task IDs such as `AC-001 -> T1.1, T2.1`.
- An acceptance criterion is considered covered only when ALL tasks mapped to it are complete. If any mapped task is open, the criterion is not yet satisfied — verify must treat it as incomplete regardless of other mapped tasks being done.
- For newly created task lists, task IDs are required.
- When meaningfully updating an existing task list without task IDs, normalize it to the ID-based format.

## Content Quality Rules

- Each phase should have a short goal that explains why the phase exists.
- **Lazy Decomposition**: Prefer a few concrete tasks (5-10 per feature) with measurable outcomes over many tiny bookkeeping items. Do not create "micro-tasks" (1-5 lines of code) during the tasks phase; the implementation agent will refine them in-place if needed.
- Focus on "milestone" tasks tied to specific files or functional boundaries.
- Keep the outcome part of each task to ≤ 12 words. If more words are needed, the task is not concrete enough — split it or tighten the verb.
- When the acceptance proof is simple, embed it directly in the outcome instead of requiring a spec lookup: prefer `add POST /auth/login — returns 200 with JWT token field — AC-001` over `add login handler — endpoint works — AC-001`.
- Use action verbs tied to observable work: implement, add, migrate, validate, remove, backfill, document.
- Keep foundational setup separate from core behavior and separate proof/validation from broad implementation.
- Keep `Touches:` values as concrete file paths when possible (`src/auth/handler.ts`), not abstract concepts (`auth flow`). The implement agent needs paths, not descriptions.
- When a task exists only to prove behavior, make that explicit instead of hiding it inside a larger implementation task.
- If a phase is unnecessary for this feature, omit it or state that it is intentionally not needed instead of filling it with generic tasks.
- Task text should make the intended outcome obvious to a reviewer without needing to reopen the plan for every line.
- Negative examples: avoid `misc`, `cleanup as needed`, `wire everything up`, `final polish`, or task text that hides the outcome behind a generic verb.

## Output expectations

- Write or patch `.speckeep/specs/<slug>/plan/tasks.md`; call out blockers if decomposition is not yet possible
- End with a summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`
- When ready: `Ready for: /speckeep.implement <slug>`

## Self-Check

- Does every task have a stable task ID and measurable outcome?
- Is every acceptance criterion covered explicitly?
- Could another developer execute these tasks in order without guessing what `done` means for each line?
