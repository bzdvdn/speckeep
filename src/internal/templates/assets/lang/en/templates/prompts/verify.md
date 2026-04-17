# SpecKeep Verify Prompt

You are verifying one implemented feature package after task execution.

Follow base rules in `AGENTS.md` (paths, git, load discipline, readiness scripts, language, phase discipline).

## Goal

Confirm whether the implemented work is aligned enough with tasks and project rules to proceed safely.

## Flags

`--deep`: full implementation validation — read all plan artifacts and inspect actual code for every completed task and AC. Produces per-AC evidence. Without `--deep`, verification stays structural and cheap.

`--persist`: kept for backward compatibility. By default you MUST persist the report to `.speckeep/specs/<slug>/plan/verify.md` (in addition to conversation output), using `.speckeep/templates/verify.md` with the machine-readable metadata block. Skip writing only if the user explicitly asks for chat-only output.

## Phase Contract

Inputs: `.speckeep/constitution.md`, `.speckeep/specs/<slug>/plan/tasks.md`; spec, plan, code only to confirm concrete claims (all artifacts in `--deep`).
Outputs: verdict report (`pass`, `concerns`, or `blocked`) in conversation AND persisted to `plan/verify.md` by default.
Stop if: slug ambiguous, `tasks.md` missing, or verdict would require inventing implementation facts.

## Load First

Always read first:

- `.speckeep/constitution.summary.md` if present; it always lives at the fixed technical path in `.speckeep/`
- Otherwise `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/plan/tasks.md`

## Load If Present

Read only when a specific check references content there:

- `summary.md` (or `spec.md`) — for acceptance coverage or task-to-AC alignment
- `plan/plan.md` — when a task references a `DEC-*` or architectural decision to confirm
- `plan/data-model.md` — when a task touches persisted state or entity shape
- `plan/contracts/` — when a task touches API or event boundaries
- `plan/research.md` — only when a check depends on a documented trade-off or external dependency finding
- code files — only files named in an active task's `Touches:` needed to confirm implementation

## Do Not Read By Default

- unrelated code areas
- broad repo history
- archives unless the current verification explicitly depends on them

## Stop Conditions

Stop and ask only if:

- slug ambiguous
- `tasks.md` missing
- verification would invent implementation facts
- conclusion would require a broad repository sweep instead of focused evidence
- the implementation claim cannot be confirmed from the current tasks, plan artifacts, and targeted code inspection

## Rules

- Start from `tasks.md` as the entrypoint.
- If `/.speckeep/scripts/check-verify-ready.*` exists, run it as the cheap first pass (slug as first arg: `bash ./.speckeep/scripts/check-verify-ready.sh <slug>` or PowerShell `.\.speckeep\scripts\check-verify-ready.ps1 <slug>`). Fallback: `/.speckeep/scripts/verify-task-state.*`. Prefer helper output over source.
- Treat verify as an evidence log, not a reassurance ritual.
- Verify that completed tasks are consistent with the current state of the feature package.
- Verify that open tasks do not contradict any claim that the feature is fully complete.
- Verify acceptance-to-task coverage consistency when `tasks.md` has `## Acceptance Coverage`.
- Reference task IDs (`T1.1`) directly in checks, findings, and conclusions.
- Prefer confirming concrete implementation claims over broad subjective review.
- Prefer `concerns` over `pass` when the evidence is partial but no contradiction has been found.

### Traceability

- Run `/.speckeep/scripts/trace.* <slug>` to scan for `@sk-task` and `@sk-test` annotations. Include findings in `## Checks` as concrete implementation evidence.
- **Legacy fallback**: if `trace` returns no findings (older features without annotations), inspect manually: for each completed task, read files listed in `Touches:` and confirm the described change is present; confirm the `AC-*` observable behavior is reachable. Do not invent evidence — unconfirmed claims go to `## Not Verified`. Add to `## Warnings`: "No `@sk-task` / `@sk-test` annotations found; traceability verified through `Touches:` inspection only."

### Mode rules

- Keep default verification structural and cheap by default.
- `--deep` mode:
  - Read all plan artifacts (`plan.md`, `data-model.md`, `contracts/`, `research.md`).
  - For every completed task, read `Touches:` files and confirm the work matches. Do not widen beyond `Touches:` unless a concrete contradiction requires it.
  - For every `AC-*`, confirm it is satisfied from `Touches:` code evidence of mapped tasks — ≥1 concrete proof per AC. Do not require exhaustive archaeology.
  - `## Scope` must state `mode: deep` and list inspected surfaces.
  - `## Not Verified` should be minimal or `none`.
- Without `--deep`, deepen only when a concrete contradiction cannot be resolved from tasks, plan, and focused evidence.

### Verdict rules

- Verdicts: `pass`, `concerns`, `blocked`.
  - `pass`: no blocking problems; only minor or no warnings.
  - `concerns`: can move forward, but warnings/open questions should be resolved soon.
  - `blocked`: missing task completion or contradictory implementation state would make archive/completion claims unsafe.
- Do not use `pass` unless the completed task state is confirmed, no blocking contradiction remains, and every claim you mention is backed by inspected evidence.

### Report

- Keep in the configured documentation language when persisting. Use `.speckeep/templates/verify.md` as the canonical template. Include a machine-readable metadata block at the top with `report_type`, `slug`, `status`, `docs_language`, `generated_at`.
- Structure: YAML metadata → `# Verify Report: <slug>` → `## Scope` → `## Verdict` → `## Checks` → `## Errors` → `## Warnings` → `## Questions` → `## Not Verified` → `## Next Step`.
- `## Scope`: actual verification mode and surfaces really inspected.
- `## Verdict`: include `archive_readiness` and a one-line summary justifying the verdict.
- `## Checks` must explicitly cover:
  - `task_state` with completed/open counts
  - `acceptance_evidence` for the `AC-*` items you actually confirmed
  - `implementation_alignment` with the concrete surface inspected
- `## Not Verified`: list material claims or surfaces intentionally not checked. `none` only when no material gaps remain in the chosen scope.
- Keep claims scoped. If you only checked task state + one endpoint or file, say so — don't imply full feature validation.

### Recovery

If verification reveals a workflow gap, send the feature back to the narrowest earlier phase that can honestly fix it:
- `implement` — missing or contradictory implementation
- `tasks` — missing/misleading/incomplete decomposition
- `plan` — implementation cannot be judged because design intent is underspecified

### Next Step

- For `pass`, name the exact archive command.
- For `concerns`, say whether workflow may continue; if not, use an explicit return command for the earlier phase.
- For `blocked`, do not suggest archive; end with `Return to: /speckeep.<phase> <slug>` for the narrowest honest recovery phase.

## Output

- Output the report to the conversation AND persist to `plan/verify.md` by default (skip only if user explicitly asks).
- Summarize verdict, completed checks, remaining concerns, and archive safety.
- End with a summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, and either `Ready for` or `Return to`.
- When safe to archive: `Ready for: /speckeep.archive <slug>`; when returning to an earlier phase, name it explicitly with its slash command.
