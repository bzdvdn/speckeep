# SpecKeep Archive Prompt

You are archiving one feature package.

## Goal

Create a durable archive snapshot for one feature, or restore a previously archived feature back to active development.

## Path Resolution

Paths in this prompt use the default workspace layout. If `.speckeep/speckeep.yaml` overrides `paths.specs_dir` or `paths.archive_dir`, always follow the configured paths instead of the defaults shown here.
Read `.speckeep/speckeep.yaml` at most once per session to resolve these paths; do not re-read it unless it changed or a path is ambiguous.

## Flags

`--copy`: keep originals in place after archiving (copy-only mode). By default, originals are deleted after archiving; `--copy` preserves them. Useful for `deferred` features that may return to active development.

`--restore`: reverse a previous archive — copy the latest snapshot back into active `specs/<slug>/`, then remove the archive entry. See Restore Rules below.

## Script Call

**Do NOT read or copy files manually. Run the script directly — it validates verify status internally and returns an error if prerequisites are not met.**

Default `--status` to `completed` unless the user explicitly provided a different one. Valid statuses: `completed`, `superseded`, `abandoned`, `rejected`, `deferred`.

If status is not `completed` and no `--reason` was provided, ask the user for a reason before running.

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

## Output expectations

### Default mode

- Say: `Ready for: ./.speckeep/scripts/archive-feature.sh <slug> --status <status>` (add `--reason "..."`
  when status is not `completed`, or when the user explicitly wants to preserve a reason even for `completed`)
- After execution: confirm success, summarize status and archived files
- State that archive is the terminal workflow step for this feature

### Restore mode

- After execution: confirm restored files, suggest `/speckeep.inspect <slug>`
