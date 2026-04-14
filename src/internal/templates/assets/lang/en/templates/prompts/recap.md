# SpecKeep Recap Prompt

You are producing a project-level overview for a new agent session.

## Goal

Give a concise, current-state summary of the project and all active features so a new session can orient itself without re-reading every artifact.

This command is optional, requires no slug, and produces no file — inline response only.

## Phase Contract

Inputs: project constitution file (configured via `project.constitution_file`; default: `.speckeep/constitution.md`); spec files and inspect reports for active features.
Outputs: inline response only — no file is written.
Stop if: the configured constitution file is missing.

## Path Resolution

Paths in this prompt use the default workspace layout. If `.speckeep/speckeep.yaml` overrides `project.constitution_file`, `paths.specs_dir`, or `paths.archive_dir`, always follow the configured paths instead of the defaults shown here. Read `.speckeep/speckeep.yaml` at most once per session to resolve these paths; do not re-read it unless it changed or a path is ambiguous.

## Load First

Always read these first:

- `.speckeep/constitution.summary.md` if present; otherwise the configured constitution file (`project.constitution_file`; default: `.speckeep/constitution.md`)

## Load If Present

Read these only to determine per-feature phase and status:

- `.speckeep/scripts/list-specs.*` — run it to enumerate active specs; do not read its source
- One header block per active spec (goal and phase markers only — do not read full spec content)
- `.speckeep/specs/<slug>/inspect.md` metadata block only (verdict line) when inspect exists
- `.speckeep/archive/` — list subdirectories to identify recently archived features; read only `summary.md` from archives dated within the last 7 days

## Do Not Read By Default

- Full spec content beyond the goal and phase indicators
- Plan packages unless the user explicitly asks for plan-level detail
- tasks.md, data-model.md, contracts/
- archived features older than 7 days
- implementation files
- script source files

## Stop Conditions

Stop and ask only if:

- `.speckeep/constitution.md` is missing

## Rules

- Read the constitution once; extract the project purpose, constraints, and key principles in one pass.
- Run `list-specs` to get the active spec list. If the script is unavailable, list files in `.speckeep/specs/` instead.
- For each active spec, read only enough to extract: slug, feature name, current phase, and inspect verdict if available.
- Determine current phase from artifact presence: no inspect → spec phase; inspect exists, no plan → inspect phase; plan exists, no tasks or open tasks → plan or tasks or implement phase; all tasks closed and no `verify.md` → verify phase; `verify.md` exists with `pass` or `concerns` → archive phase; `verify.md` exists with `blocked` → blocked at verify (return required).
- Do not read the full plan package for any feature unless the user explicitly asks.
- Keep the response compact: it must be readable in under 30 seconds.
- If there are no active features, say so clearly after summarizing the project context.
- Check `.speckeep/archive/` for features archived within the last 7 days. If any exist, include them in the output under **Recently Archived**.

## Output Format

Respond inline only. Use this structure:

**Project** — one or two sentences from the constitution: what the project is and its primary constraint or goal.

**Active Features** — one row per feature:

```
<slug>  <name>  [phase]  [inspect: pass|concerns|blocked — omit if no inspect]
```

**Recently Archived** (only if any archives exist from the last 7 days) — one row per feature:

```
<slug>  <status>  <archived_at>  <one-line reason from summary.md>
```

**Summary** — one line: how many features are active, how many are blocked, what the dominant phase is, how many were recently archived.

**Suggested next command** — the single most useful next action based on current state (e.g. the most-blocked feature's recovery command, or the feature closest to archive).

Keep the entire response under 20 lines unless the project has more than 5 active features.

## Self-Check

- Did I read only the minimum needed to determine each feature's phase?
- Is the response compact enough to orient a new session quickly?
- Did I avoid reading full spec or plan content?
- Is the suggested next command grounded in actual feature state?
