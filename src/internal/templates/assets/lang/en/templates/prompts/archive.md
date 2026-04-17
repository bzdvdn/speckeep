# SpecKeep Archive Prompt

You are archiving one feature package.

Follow base rules in `AGENTS.md` (paths, git, load discipline, readiness scripts, language, phase discipline).

## Goal

Create a durable archive snapshot for one feature, or restore a previously archived feature back to active development.

## Flags

`--copy`: keep originals in place after archiving (copy-only mode). By default originals are deleted; `--copy` preserves them. Useful for `deferred` features that may return.

`--restore`: reverse a previous archive — copy the latest snapshot back into active `specs/<slug>/`, then remove the archive entry.

## Script Call

**Do NOT read or copy files manually. Run the script directly — it validates verify status internally and returns an error if prerequisites are not met.**

Default `--status` to `completed` unless the user explicitly provided a different one. Valid statuses: `completed`, `superseded`, `abandoned`, `rejected`, `deferred`. If status is not `completed` and no `--reason` was provided, ask the user for a reason before running.

**Unix/macOS:**
```bash
./.speckeep/scripts/archive-feature.sh <slug> --status <status> [--reason "<reason>"]
```

**Windows (PowerShell):**
```powershell
.\.speckeep\scripts\powershell\archive-feature.ps1 <slug> --status <status> [--reason "<reason>"]
```

Examples (Unix):
- `./.speckeep/scripts/archive-feature.sh my-feature`
- `./.speckeep/scripts/archive-feature.sh my-feature --status completed`
- `./.speckeep/scripts/archive-feature.sh my-feature --status deferred --reason "Postponed to Q3" --copy`
- `./.speckeep/scripts/archive-feature.sh my-feature --restore`

Examples (Windows):
- `.\.speckeep\scripts\powershell\archive-feature.ps1 my-feature`
- `.\.speckeep\scripts\powershell\archive-feature.ps1 my-feature --status completed`
- `.\.speckeep\scripts\powershell\archive-feature.ps1 my-feature --status deferred --reason "Postponed to Q3" --copy`
- `.\.speckeep\scripts\powershell\archive-feature.ps1 my-feature --restore`

## Output

### Default mode

- Archive as `completed` (default): `Ready for: ./.speckeep/scripts/archive-feature.sh <slug>`
- Different status: `Ready for: ./.speckeep/scripts/archive-feature.sh <slug> --status <status> --reason "<reason>"` (and `--copy` when needed).
- After execution: confirm success, summarize status and archived files.
- State that archive is the terminal workflow step for this feature.

### Restore mode

- After execution: confirm restored files and list paths.
- Note that the restored spec is unverified — any existing `inspect.md` may be stale.
- End with: `Ready for: /speckeep.inspect <slug>` (re-inspect required before planning resumes).
