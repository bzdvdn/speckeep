# SpecKeep Verify Prompt

You are verifying one implemented feature package after task execution.

## Goal

Confirm whether the implemented work is aligned enough with tasks and project rules to proceed safely.

## Path Resolution

Paths in this prompt use the default workspace layout. If `.speckeep/speckeep.yaml` overrides `paths.specs_dir`, `paths.archive_dir`, or `project.constitution_file`, always follow the configured paths instead of the defaults shown here.
Read `.speckeep/speckeep.yaml` at most once per session to resolve these paths; do not re-read it unless it changed or a path is ambiguous.

## Flags

`--deep`: full implementation validation mode — read all plan artifacts and inspect actual code for every completed task and acceptance criterion, not just structural checks. Produces a comprehensive report with per-AC evidence. Without this flag, verification stays structural and cheap by default.

`--persist`: kept for backward compatibility. By default, you MUST persist the verification report to `.speckeep/specs/<slug>/plan/verify.md` (in addition to conversation output). When persisting, use `.speckeep/templates/verify-report.md` as the canonical template and include the machine-readable metadata block. Only skip writing the file if the user explicitly asks for chat-only output.

## Phase Contract

Inputs: `.speckeep/constitution.md`, `.speckeep/specs/<slug>/plan/tasks.md`; spec, plan, code only to confirm concrete claims (or all artifacts in `--deep` mode).
Outputs: verdict report (`pass`, `concerns`, or `blocked`) in conversation AND persisted to `.speckeep/specs/<slug>/plan/verify.md` by default.
Stop if: slug ambiguous, tasks.md missing, or verdict would require inventing implementation facts.

## Load First

Always read these first:

- `.speckeep/constitution.summary.md` if present; otherwise `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/plan/tasks.md`

## Load If Present

Read when a specific check references content in these files (e.g., a task claims to satisfy an `AC-*`, or a `DEC-*` constrains implementation shape):

- `.speckeep/specs/<slug>/summary.md` (or `spec.md`) — when verifying acceptance coverage or task-to-AC alignment
- `.speckeep/specs/<slug>/plan/plan.md` — when a task references a `DEC-*` or architectural decision that must be confirmed
- `.speckeep/specs/<slug>/plan/data-model.md` — when a task touches persisted state or entity shape
- `.speckeep/specs/<slug>/plan/contracts/` — when a task touches API or event boundaries
- `.speckeep/specs/<slug>/plan/research.md` — only when the check depends on a documented trade-off or external dependency finding
- code files — only the specific files named in a task's `Touches:` that are needed to confirm the task was actually implemented

## Do Not Read By Default

- unrelated code areas
- broad repository history
- archives unless the current verification explicitly depends on them

## Stop Conditions

Stop and ask for clarification only if:

- the slug is ambiguous
- the tasks file is missing
- the verification would otherwise invent implementation facts
- the requested conclusion would require a broad repository sweep instead of focused evidence for this feature package
- the implementation claim cannot be confirmed from the current tasks, plan artifacts, and targeted code inspection

## Rules

- Start from `tasks.md` as the verification entrypoint.
- If `/.speckeep/scripts/check-verify-ready.*` is available, prefer it as the cheap first pass before reading deeper artifacts.
- Important: the readiness wrapper runs with the slug as the first argument. Example: `bash ./.speckeep/scripts/check-verify-ready.sh <slug>` (or PowerShell: `.\.speckeep\scripts\check-verify-ready.ps1 <slug>`).
- Use `/.speckeep/scripts/verify-task-state.*` only as a fallback when the phase-readiness wrapper is unavailable.
- Prefer helper script output over reading helper script source.
- Do not read `/.speckeep/scripts/*` by default unless you are debugging the script, working on SpecKeep itself, or the user explicitly asks to inspect script logic.
- Prefer confirming concrete implementation claims over broad subjective review.
- Treat verify as an evidence log, not a reassurance ritual.
- Verify that completed tasks are consistent with the current state of the feature package.
- **Traceability Evidence**: Use `/.speckeep/scripts/trace.* <slug>` to scan for `@sk-task` and `@sk-test` annotations in the code. Include these findings in the `## Checks` section as concrete implementation evidence.
- **Legacy Fallback**: If `trace` returns no findings (e.g., for older features without annotations), proceed with manual inspection: for each completed task, read the files listed in its `Touches:` field and confirm the specific change described in the task outcome is present. Check that the observable behavior named in each `AC-*` is reachable from those files. Do not invent evidence — if a claim cannot be confirmed from `Touches:` files, record it as unverified in `## Not Verified`. Add an entry to `## Warnings` stating: "No `@sk-task` / `@sk-test` annotations found; traceability verified through `Touches:` inspection only."
- Do not run git commit/git push/git tag or open a PR unless the user explicitly asks.
- Verify that open tasks do not contradict any claim that the feature is fully complete.
- Verify acceptance-to-task coverage consistency when `tasks.md` includes an `Acceptance Coverage` section.
- When `tasks.md` uses task IDs such as `T1.1`, reference those IDs directly in checks, findings, and conclusions.
- Prefer `concerns` over `pass` when the evidence is partial but no contradiction has been found.
- Keep default verification structural and cheap by default.
- When `--deep` is present in the user arguments, switch to full validation mode:
  - Read all plan artifacts (`plan.md`, `data-model.md`, `contracts/`, `research.md`).
  - For every completed task, read the actual implementation files listed in `Touches:` and confirm the work matches the task description. Do not widen beyond `Touches:` unless a concrete contradiction requires it.
  - For every `AC-*`, confirm it is satisfied using the code evidence found in the `Touches:` files of its mapped tasks — at least one concrete proof per AC. Do not require exhaustive code archaeology beyond those boundaries.
  - The `## Scope` section must state `mode: deep` and list all surfaces inspected.
  - The `## Not Verified` section should be minimal or `none` — deep mode is expected to be thorough within the `Touches:` boundaries.
- Without `--deep`, only deepen into broader implementation validation when a concrete contradiction cannot be resolved from tasks, plan artifacts, and focused evidence.
- Use a simple verdict: `pass`, `concerns`, or `blocked`.
- Use `pass` when no blocking problems are present and only minor or no warnings remain.
- Use `concerns` when the feature can move forward, but warnings or open questions should be resolved soon.
- Use `blocked` when missing task completion or contradictory implementation state would make archive or completion claims unsafe.
- Do not use `pass` unless the completed task state is confirmed, no blocking contradiction remains, and every acceptance or implementation claim you mention is backed by inspected evidence.
- Keep the verification output in the project's configured documentation language when writing it to disk.
- Use `.speckeep/templates/verify-report.md` as the canonical template when writing the report to disk.
- When writing the report to disk, include a machine-readable metadata block at the top with `report_type`, `slug`, `status`, `docs_language`, and `generated_at`.
- Use this report structure:
  - YAML-style metadata block at the top
  - `# Verify Report: <slug>`
  - `## Scope`
  - `## Verdict`
  - `## Checks`
  - `## Errors`
  - `## Warnings`
  - `## Questions`
  - `## Not Verified`
  - `## Next Step`
- In `## Scope`, record the actual verification mode and the surfaces you really inspected.
- In `## Verdict`, include `archive_readiness` and a one-line summary that explains why the verdict is justified.
- In `## Checks`, explicitly cover:
  - `task_state` with completed/open counts
  - `acceptance_evidence` for the `AC-*` items you actually confirmed
  - `implementation_alignment` with the concrete surface inspected
- In `## Not Verified`, list any material claims or surfaces you intentionally did not check. Use `none` only when no material gaps remain inside the chosen verification scope.
- Keep claims scoped. If you only checked task state plus one endpoint or file path, say that directly instead of implying full feature validation.
- If verification discovers a workflow gap, send the feature back to the narrowest earlier phase that can honestly fix it:
  - `implement` for missing or contradictory implementation
  - `tasks` for missing, misleading, or incomplete task decomposition
  - `plan` when the implementation cannot be judged honestly because the design intent is underspecified
- For `pass`, name the exact archive command.
- For `concerns`, say whether the workflow may continue; if it may not, use an explicit return command for the earlier phase.
- For `blocked`, do not suggest archive; end with `Return to: /speckeep.<phase> <slug>` for the narrowest honest recovery phase.

## Output expectations

- Output the report to the conversation AND persist to `.speckeep/specs/<slug>/plan/verify.md` by default (skip persistence only if the user explicitly asks)
- Summarize the verdict, completed checks, remaining concerns, and whether the feature is safe to archive
- End with a summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, and either `Ready for` or `Return to`
- When safe to archive: `Ready for: /speckeep.archive <slug>`; when returning to an earlier phase, name it explicitly with its slash command

## Self-Check

- Is every verdict claim backed by inspected evidence, not just checkbox state?
- Is the `Not Verified` section honest about what I did not check?
- Is the next step or return phase appropriate for the verdict?
