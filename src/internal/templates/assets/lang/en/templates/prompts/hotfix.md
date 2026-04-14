# SpecKeep Hotfix Prompt

You are implementing an emergency fix outside the standard planning phase chain.

## Goal

Write a minimal hotfix spec, implement the fix, verify it inline, and prepare for archive — without running inspect, plan, or tasks phases.

## Path Resolution

Paths in this prompt use the default workspace layout. If `.speckeep/speckeep.yaml` overrides `project.constitution_file`, `paths.specs_dir`, or `paths.archive_dir`, always follow the configured paths instead of the defaults shown here.
Read `.speckeep/speckeep.yaml` at most once per session to resolve these paths; do not re-read it unless it changed or a path is ambiguous.

## When to Use

Use only when:
- The fix is well-understood and touches ≤ 3 files
- The root cause is already identified
- A full phase cycle would add friction without safety benefit

If scope is unclear, root cause unknown, or fix is cross-cutting — stop and use the standard workflow instead.

## Phase Contract

Inputs: `.speckeep/constitution.summary.md` (or `.speckeep/constitution.md`), user description of the fix.
Outputs: `.speckeep/specs/<slug>/hotfix.md`, implementation code.
Stop if: root cause unclear, fix exceeds 3 files, or constitutional conflict detected.

## Branch

Before writing any file, create or switch to a `hotfix/<slug>` branch. If the repo is under git:
- If you are not on `hotfix/<slug>`: switch to it; if it does not exist, create and switch (`git switch -c hotfix/<slug>`).
- If branch creation is not possible (no git, detached HEAD, environment constraints), stop and report before changing files.

## Load First

Always read these before writing any code:

- `.speckeep/constitution.summary.md` if present; otherwise `.speckeep/constitution.md`
- only the files directly involved in the fix

## Do Not Read By Default

- full spec history or plan packages
- implementation files not listed in `Touches`
- script source files

## Stop Conditions

Stop and switch to the standard workflow if:

- root cause is unclear — use `/speckeep.spec <slug>` to document the investigation, then `/speckeep.plan <slug>` to design the fix properly
- fix requires changing more than 3 files
- fix touches an API contract, data migration, or auth boundary without a clear rollback
- constitutional conflict is present
- scope requires inventing tasks beyond the stated fix

## Hotfix Spec

Write `.speckeep/specs/<slug>/hotfix.md` before touching any code:

```
---
slug: <slug>
type: hotfix
created_at: <date>
---

## Fix
<what is broken and what the change does — one or two sentences>

## Root Cause
<why it broke — one sentence>

## Risk
<what could break from this fix — one sentence; "none" only if genuinely safe>

## Verification
<concrete observable check — command output, HTTP response, or UI behavior>

## Touches
<file, file>
```

## Invariants

- Write the hotfix spec before any code change.
- Touch only files listed in `Touches`.
- Do not run git commit/git push/git tag or open a PR unless the user explicitly asks.
- Keep the fix minimal — no refactoring, no scope beyond the stated fix.
- If a file outside `Touches` must change, stop and explain before continuing.
- Log non-obvious assumptions as `[ASSUMPTION: ...]` before acting on them.
- Do not re-plan or re-design the fix silently; if the stated fix turns out to be wrong, stop and ask.

## Verification Failure

If the observable check from `## Verification` is not met after implementing the fix:
- Do not mark the hotfix as done.
- Stop and report: what was expected, what was observed, and which file is most likely the cause.
- Do not expand scope to compensate — stop and ask the user whether to adjust the fix or abandon the hotfix path in favor of the standard workflow.

## Archive Note

The `archive` prompt reads `spec.md` as its primary input. For a hotfix, `spec.md` does not exist — only `hotfix.md`. When archiving a hotfix, tell the agent to use `hotfix.md` in place of `spec.md` for the summary and to set `type: hotfix` in the archive metadata.

## Output expectations

- Create or switch to `hotfix/<slug>` branch first
- Write `.speckeep/specs/<slug>/hotfix.md`
- Implement the fix; confirm the observable proof from the `Verification` section is met
- End with a summary block: `Slug`, `Status`, `Fix`, `Verified`, `Ready for`
- When done: `Ready for: /speckeep.archive <slug>` (note: use `hotfix.md` as the spec source when archiving)

## Self-Check

- Did I create or switch to `hotfix/<slug>` branch before writing any file?
- Did I write the hotfix spec before touching any code?
- Is the fix limited to files listed in `Touches`?
- Is the observable proof from `Verification` confirmed?
