# SpecKeep Tasks Prompt

You are creating or updating `<specs_dir>/<slug>/plan/tasks.md` (default: `.speckeep/specs/<slug>/plan/tasks.md`).

Follow base rules in `AGENTS.md` (paths, git, load discipline, readiness scripts, language, phase discipline).

## Goal

Break an approved plan into executable implementation tasks.

## Phase Contract

Inputs: `.speckeep/constitution.md`, `.speckeep/specs/<slug>/plan/plan.md`; optionally `summary.md`, `plan/data-model.md`, `plan/contracts/` when decomposition requires them.
Outputs: `.speckeep/specs/<slug>/plan/tasks.md` with phased task list and `## Acceptance Coverage`.
Stop if: `plan.md` missing, plan underspecified, or any AC cannot be mapped to executable work.

## Operating Mode

- Decompose one approved plan package only.
- Use `plan.md` as the entrypoint; go deeper only when required.
- Produce the smallest task list that still covers the feature safely.
- Prefer explicit sequencing over umbrella tasks.

## Load First

Always read before decomposing:

- `.speckeep/constitution.summary.md` if present; otherwise `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/plan/plan.md`

## Load If Present

Read only when a task being decomposed references or depends on these:

- `summary.md` (or `spec.md`) — when an `AC-*` has ambiguous scope or acceptance boundary
- `plan/data-model.md` — when tasks must create/modify entities, fields, invariants, or state transitions
- `plan/contracts/` — when tasks must create/modify API endpoints, event payloads, or integration boundaries
- `plan/research.md` — only when a documented finding changes task sequencing or introduces a risk affecting decomposition
- code files — only files needed to confirm where implementation surfaces begin and end

Do not assume `research.md` should exist; use it only when the plan clearly depends on preserved uncertainty, an external dependency, or a documented trade-off.

## Do Not Read By Default

- implementation files not needed to decompose the work
- broad repository history

## Flags

`--repair <task-id-list>`: targeted repair — fix specific tasks from verify/review without rewriting the full list.

- Read existing `tasks.md` first. Locate only named tasks (e.g., `--repair T2.3,T3.1`).
- For each: update description, outcome, `Touches:`, or `AC-*` mapping as needed.
- Do not restructure phases, renumber others, or rewrite sections not being repaired.
- If the repair reveals the plan is flawed, stop and suggest `/speckeep.plan <slug> --update`.
- Update `## Acceptance Coverage` only if mapping changed; update `## Surface Map` only if `Touches:` changed.

`--greenfield-story`: story-first decomposition for greenfield or early-product work.

- Keep `Touches:` and `## Surface Map` mandatory.
- Group delivery around foundation, MVP story, next story, and hardening instead of only technical layers when that makes the plan easier to execute.
- Prefer phases that map to independently demonstrable product slices.

## Stop Conditions

Stop and ask for refinement if:

- `plan.md` is missing
- tasks would be vague because the plan is underspecified
- decomposition requires spec, data-model, contracts, or research that are missing
- the constitution blocks the decomposition
- the decomposition would span multiple feature slugs or unrelated change sets
- one or more acceptance criteria cannot be mapped to executable work without guessing

Do not jump into implementation.

## Invariants

- Tasks align with plan and constitution.
- `plan.md` is the decomposition entrypoint. Never read unrelated feature artifacts to compensate for underspecified planning.
- Read code only when tasks would otherwise stay vague; prefer a narrow slice over broad exploration.
- If `/.speckeep/scripts/check-tasks-ready.*` exists, prefer running it as the readiness check (slug as first arg: `bash ./.speckeep/scripts/check-tasks-ready.sh <slug>` or PowerShell `.\.speckeep\scripts\check-tasks-ready.ps1 <slug>`).
- The task list must be executable in order. Every AC covered by ≥1 task.
- Prefer concrete, testable, implementation-oriented tasks. Include validation/docs work only when needed. No vague umbrella tasks.
- The task list should be readable to both an implementation agent and a human reviewer without extra interpretation.
- Targeted code reading during decomposition is useful when it reduces re-reading during implementation.
- Do not start implementation work, edit source code, or claim tasks are already done during the tasks phase.

## Task Format Rules

- Follow `.speckeep/templates/tasks.md`: group tasks into ordered phases (`## Phase N: Name`).
- Each task MUST have a phase-scoped task ID `T<phase>.<index>`.
- Format: `- [ ] T<phase>.<index> <action verb> — <concrete measurable outcome>`
- Reference 1–2 stable IDs when possible (`AC-*`, `RQ-*`, `DEC-*`).
- Each task MUST include a `Touches:` field naming the concrete files or modules it will create or modify. This is the primary mechanism preventing re-reads during implement — the implement agent batch-reads these in one pass. Keep it compact (`Touches: src/auth/handler.ts, src/session/store.ts`). Use module-level names only when exact file is undetermined (`Touches: src/auth/`). A task without `Touches:` forces exploratory reads and wastes tokens.
- `## Surface Map` MUST appear before the first phase: two-column table (`Surface | Tasks`) mapping each unique implementation surface to the task IDs touching it. This is the implement agent's batch-read manifest — without it the agent scans every task line to build the read list.
- Tasks together MUST cover all AC. Uncovered criterion = blocker.
- `## Acceptance Coverage` MUST include ≥1 explicit coverage line per AC (`AC-001 -> T1.1, T2.1`).
- An AC is covered only when ALL tasks mapped to it are complete. Any open mapped task → verify must treat the AC as incomplete.
- New task lists require IDs. Meaningful updates to an ID-less list → normalize to ID-based format.

## Content Quality Rules

- Each phase should have a short goal that explains why the phase exists.
- `--greenfield-story`: after foundation, prefer one phase per MVP or prioritized user story only when the plan already defines those slices clearly.
- `--greenfield-story`: keep story phases small; if a story cannot be demonstrated independently, decompose by MVP slice first.
- **Lazy Decomposition**: Prefer a few concrete tasks (5–10 per feature) with measurable outcomes over many tiny bookkeeping items. Do not create "micro-tasks" (1–5 lines of code); the implement agent refines them in-place if needed.
- Focus on "milestone" tasks tied to specific files or functional boundaries.
- Keep each task's outcome ≤ 12 words. More words → split or tighten the verb.
- When acceptance proof is simple, embed it directly in the outcome: prefer `add POST /auth/login — returns 200 with JWT token field — AC-001` over `add login handler — endpoint works — AC-001`.
- Use action verbs tied to observable work: implement, add, migrate, validate, remove, backfill, document.
- Keep foundational setup separate from core behavior; separate proof/validation from broad implementation.
- Keep `Touches:` values as concrete paths (`src/auth/handler.ts`), not abstract concepts (`auth flow`).
- When a task exists only to prove behavior, make it explicit instead of hiding it in a larger task.
- If a phase is unnecessary, omit it or explicitly state it is not needed rather than filling with generic tasks.
- Task text makes the intended outcome obvious without reopening the plan.
- Avoid: `misc`, `cleanup as needed`, `wire everything up`, `final polish`, or verb-hidden outcomes.

## Output

- Write or patch `.speckeep/specs/<slug>/plan/tasks.md`; call out blockers if decomposition is not yet possible.
- End with a summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Ready for`.
- When ready: `Ready for: /speckeep.implement <slug>`.

## Self-Check

- Could another developer execute these tasks in order without guessing what `done` means for each line?
