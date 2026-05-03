# CLI Reference

## Install

SpecKeep is distributed as a single binary via GitHub Releases.

Linux:

```bash
VERSION=v0.1.0
curl -fsSL "https://raw.githubusercontent.com/bzdvdn/speckeep/${VERSION}/scripts/install.sh" | bash -s -- --version "${VERSION}"
```

Windows (PowerShell):

```powershell
$version="v0.1.0"
$env:DRAFTSPEC_VERSION=$version
powershell -ExecutionPolicy Bypass -c "iwr -useb https://raw.githubusercontent.com/bzdvdn/speckeep/$version/scripts/install.ps1 | iex"
```

To also add the install directory to `PATH`:

- Linux: add `--add-to-path` or set `DRAFTSPEC_ADD_TO_PATH=1`
- Windows: set `$env:DRAFTSPEC_ADD_TO_PATH=1` or run the script with `-AddToPath`

## Commands

### `speckeep init [path]`

Initializes a SpecKeep workspace in the target project.

Examples:

```bash
speckeep init
speckeep init my-project --lang en --shell sh
speckeep init my-project --lang en --shell sh --specs-dir .speckeep/specifications --archive-dir .speckeep/artifacts/archive --constitution-file docs/constitution.md
speckeep init my-project --docs-lang ru --agent-lang en --comments-lang en --shell powershell --agents claude --agents cursor
```

Important flags:

- `--git` initializes a Git repository when true; default is enabled
- `--lang` sets the base language; default is `en`
- `--shell` selects the generated workflow script family; required: `sh` or `powershell`
- `--specs-dir` overrides the specs directory (advanced)
- `--archive-dir` overrides the archive directory (advanced)
- `--constitution-file` overrides the constitution file path (advanced)
- `--docs-lang` sets the generated documentation language
- `--agent-lang` sets the generated prompt and agent guidance language
- `--comments-lang` records the preferred code comment language
- `--agents` generates project-local agent command files

### `speckeep refresh [path]`

Refreshes only SpecKeep-managed generated artifacts in an existing project.

This command updates:

- `.speckeep/speckeep.yaml`
- `.speckeep/skills/manifest.yaml`
- `.speckeep/templates/**`
- `.speckeep/scripts/**`
- project-local agent command files
- the managed SpecKeep guidance block inside `AGENTS.md`

This command does not update:

- the constitution file (`project.constitution_file`, default: `CONSTITUTION.md`)
- contents under `specs_dir/**` (but it can safely move the directory with `--specs-dir`)
- contents under `specs_dir/<slug>/plan/**`
- contents under `archive_dir/**` (but it can safely move the directory with `--archive-dir`)

Examples:

```bash
speckeep refresh my-project
speckeep refresh my-project --shell powershell --agents claude --dry-run
speckeep refresh my-project --agent-lang ru --json
```

Important flags:

- `--lang`, `--docs-lang`, `--agent-lang`, `--comments-lang` override the existing configured languages
- `--shell` overrides the generated workflow script family
- `--constitution-file` overrides the configured constitution file path (and safely moves the existing file when possible)
- `--specs-dir` overrides `paths.specs_dir` (and safely moves the existing specs directory when possible)
- `--archive-dir` overrides `paths.archive_dir` (and safely moves the existing archive directory when possible)
- `--agents` overrides enabled project-local agent targets
- `--dry-run` reports pending managed changes without writing them
- `--json` outputs the refresh result as JSON

### `speckeep add-agent [path]`

Adds one or more agent targets to an existing SpecKeep project.

```bash
speckeep add-agent my-project --agents claude --agents codex
```

### `speckeep list-agents [path]`

Lists enabled agent targets from `.speckeep/speckeep.yaml`.

### `speckeep remove-agent [path]`

Disables one or more agent targets and removes their generated files.

### `speckeep cleanup-agents [path]`

Removes orphaned agent artifacts that no longer match enabled targets in config.

### `speckeep add-skill [path]`

Adds or updates one skill in `.speckeep/skills/manifest.yaml`.

For git sources, `--ref` is required to keep installs reproducible and avoid floating branch drift.

Use `--no-install` to update only manifest/AGENTS and skip immediate install into agent skill folders.

Examples:

```bash
speckeep add-skill my-project --id architecture --from-local skills/architecture
speckeep add-skill my-project --id openai-docs --from-git https://example.com/skills.git --ref v1.2.3 --path skills/openai-docs
```

### `speckeep list-skills [path]`

Lists configured skills from `.speckeep/skills/manifest.yaml`.

Use `--json` for machine-readable output.

### `speckeep remove-skill [path]`

Removes one skill from `.speckeep/skills/manifest.yaml`.

Use `--no-install` to skip immediate reconciliation of installed skills in agent folders.

### `speckeep install-skills [path]`

Installs enabled skills from `.speckeep/skills/manifest.yaml` into target agent skill folders.

By default, uses enabled targets from `.speckeep/speckeep.yaml`. Override with `--targets codex,claude`.

Important flags:

- `--dry-run` reports pending changes without writing them
- `--json` outputs install results as JSON
- `--include-disabled` installs disabled skills too

Equivalent subcommand:

```bash
speckeep skills install my-project
```

### `speckeep sync-skills [path]`

Synchronizes skill-managed artifacts only:

- `.speckeep/skills/manifest.yaml`
- managed SpecKeep block in `AGENTS.md` (including skills section)

Important flags:

- `--dry-run` reports pending changes without writing them
- `--json` outputs sync results as JSON

Equivalent subcommand:

```bash
speckeep skills sync my-project
```

### `speckeep doctor [path]`

Checks workspace health.

`doctor` reports:

- `error` for missing required files or invalid config values
- `warning` for orphaned agent artifacts still present on disk
- `warning` for non-standard Git branch names
- `ok` when the workspace is healthy

Use `--json` for machine-readable output in automation and CI.

### `speckeep dashboard [path]`

Displays a visual dashboard of all active features in the project.

The dashboard includes:

- Feature slug
- Current workflow phase
- Implementation progress percentage
- Status (READY/BLOCKED)
- Current Git branch (with `!!` warning if there is a mismatch with the feature slug)

```bash
speckeep dashboard
```

### `speckeep feature <slug> [path]`

Shows a detailed workflow view for one feature.

The text view includes:

- current phase and `ready_for`
- inspect and verify status when reports exist
- task progress when `tasks.md` exists
- grouped workflow findings
- a short `focus` hint for the next likely action

Use `--json` to return structured state plus feature-local findings.

### `speckeep feature repair <slug> [path]`

Repairs safe feature-local SpecKeep issues.

Current repair scope includes:

- migrating flat spec artifacts (`specs/<slug>.md`) to the canonical directory layout (`specs/active/<slug>/spec.md`)
- migrating plan artifacts from the old `plans/<slug>/` layout to `specs/active/<slug>/plan/`

Use `--dry-run` to preview changes and `--json` for structured output.

### `speckeep features [path]`

Lists workflow status across all discovered features.

The text view summarizes:

- phase and `ready_for`
- inspect and verify verdicts
- task progress
- grouped issue counts
- artifact presence

Use `--json` for machine-readable output.

### `speckeep migrate [path]`

Runs safe project-wide SpecKeep migrations.

Current migration scope focuses on canonicalizing legacy inspect report paths across the project.

### `speckeep list-specs [path]`

Lists spec slugs from `specs_dir/` (default: `specs/active/`).

### `speckeep show-spec <name> [path]`

Prints one spec file by slug.

### `speckeep check <slug> [path]`

Shows feature readiness and the exact next action for one feature.

Output includes artifact presence, inspect and verify verdict, task progress, the exact next slash command, and a compact structured-check summary when phase-specific readiness checks produce categorized findings.

Use `--all` to check every feature in one table. Exits with code 1 when any feature is blocked.
Use `--json` for machine-readable output suitable for CI, including `check_summary` and `check_findings` when available.

```bash
speckeep check export-report
speckeep check export-report my-project --json
speckeep check my-project --all
speckeep check my-project --all --json
```

### `speckeep trace [slug] [path]`

Scans for traceability annotations in the codebase.

Annotations follow the format:
- `// @sk-task <TASK_ID>: <Description> (<AC_ID>)` for implementation code.
- `// @sk-test <TASK_ID>: <TestName> (<AC_ID>)` for test evidence.

This command identifies links between implementation code, task IDs from `tasks.md`, and acceptance criteria from `spec.md`.

Use `slug` to filter findings for a specific feature.
Use `--tests` to show only test evidence.
Use `--json` for machine-readable output.

```bash
speckeep trace
speckeep trace export-report
speckeep trace export-report --tests
speckeep trace export-report my-project --json
```

### `speckeep demo [path]`

Creates a demo workspace at the given path (default: `./speckeep-demo`).

The workspace is pre-populated with an example feature (`export-report`) at the implement phase — spec, inspect report, plan, tasks, and data model are all present. Suggests `/speckeep.scope`, `/speckeep.challenge`, and `/speckeep.handoff` to try immediately.

```bash
speckeep demo
speckeep demo ./my-demo --agents claude
```

### `speckeep export <slug> [path]`

Bundles all artifacts for one feature into a single markdown document.

Reads and concatenates: spec, inspect report, plan, tasks, data model, research, challenge report, and verify report (skips missing files). Useful for sharing full feature context with a reviewer or a new agent session.

Use `--output <file>` to write to a file instead of stdout.

```bash
speckeep export export-report
speckeep export export-report my-project --output export-report-bundle.md
```

### `speckeep list-archive [path]`

Lists archived features from `archive_dir/` (default: `specs/archived/`).

Shows one entry per slug (most recent snapshot) with status, archived date, and reason. Entries are sorted by date descending. Status values are color-coded: `completed` in green, `deferred` in yellow, `abandoned` and `rejected` in red.

Flags:

- `--status` — filter by archive status: `completed`, `superseded`, `abandoned`, `rejected`, `deferred`
- `--since <YYYY-MM-DD>` — filter to archives on or after this date
- `--json` — output as JSON for automation and CI

```bash
speckeep list-archive
speckeep list-archive my-project --status deferred
speckeep list-archive my-project --since 2026-01-01
speckeep list-archive my-project --json
```
