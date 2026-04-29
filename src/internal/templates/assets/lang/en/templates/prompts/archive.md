# SpecKeep Archive Prompt (compact)

You archive one feature by running two scripts. Do not read feature files, do not inspect diffs, do not validate artifacts manually — the scripts do all of that.

## Phase Contract

Inputs: user arguments (`<slug>`, optional `--status`, optional `--reason`).
Outputs: archive snapshot produced by `archive-feature` script.
Stop if: `check-archive-ready` exits non-zero — report its stdout and stop.

## Rules

- Do not read any feature files (`spec.md`, `plan.md`, `tasks.md`, `verify.md`, etc.). The scripts handle all validation.
- Do not run `--help` on any command to discover usage. Use the script paths below exactly.
- Default status is `completed`. Use `--status deferred --reason "..."` only when the user explicitly says so.
- Trust script output completely: if `check-archive-ready` passes, proceed; if it fails, report and stop.

## Steps (always in this order)

1. Run readiness check:
   - `completed`: `./.speckeep/scripts/check-archive-ready.sh <slug> completed`
   - other status: `./.speckeep/scripts/check-archive-ready.sh <slug> <status> --reason "<reason>"`
2. If exit code 0 — run archive:
   - `completed`: `./.speckeep/scripts/archive-feature.sh <slug> . --status completed`
   - other status: `./.speckeep/scripts/archive-feature.sh <slug> . --status <status> --reason "<reason>"`

## Output expectations

- Report the script output (stdout) and final status.
- Do not update `REPOSITORY_MAP.md` during archive by default. Only update it if archive scripts or explicit user request changed repository structure/navigation.
- This is the terminal workflow step for this feature (after verify).
- Final line: `Ready for: /speckeep.recap`
