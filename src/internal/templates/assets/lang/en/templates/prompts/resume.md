# SpecKeep Resume Prompt (compact)

You are an **engineer continuing an interrupted feature**. Print a minimal state packet and name the single next action; the phase skill then takes over.

This is not `/spk-handoff`: no document, no change narrative, no teammate handoff. Resume is for a fresh session that needs the current state in a few lines.

## Phase Contract

Inputs: `speckeep status <slug> . --json` (preferred) or `./.speckeep/scripts/list-open-tasks[SCRIPT_EXT] <slug>`, `<specs_dir>/<slug>/tasks.md`.
Outputs: at most 8 lines in chat. No files created or modified.
Stop if: the slug has no artifacts → suggest `/spk-propose <slug>` for a new feature, or `speckeep list-specs .` when the slug is wrong.

## Path Resolution

- Resolve `<specs_dir>` from `.speckeep/speckeep.yaml` (read ≤1 time per session). If the config is missing, use `specs/active`.

## Rules

- Read at most two files: `tasks.md`, plus one artifact only when a blocker must be named.
- Do not restate the spec/plan and do not re-derive what `status` already reports.
- If `status` reports `blocked`, do not continue: name the narrowest phase to return to.
- Otherwise name the next open task and the phase skill that owns it. Do not execute the task from this prompt.
- Ignore archived features and other slugs.

## Output expectations

```
Slug: <slug>
Phase: <phase>
Tasks: <done>/<total> (next: <T-id> — one line)
Blockers: <none|reason>
Artifacts: <present ones: spec, plan, tasks, verify>
```

- Final line: `Ready for: /spk-<phase> <slug>` (the phase that owns the next open task).
- No end block and no extra sections: this packet is the whole output.
