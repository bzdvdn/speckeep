# speckeep

Русская версия: [README.ru.md](README.ru.md)

`speckeep` is a lightweight project context kit for development agents and humans.

It keeps project intent, specifications, plan artifacts, and task breakdowns in simple files without introducing a rigid process engine.

SpecKeep is the successor to DraftSpec (archived). If you are migrating an existing DraftSpec workspace, start with `speckeep migrate`.

The first release is intentionally optimized for low overhead and real-world usage: narrow default context, minimal required artifacts, strict workflow discipline without heavyweight orchestration, and branch-based collaboration that works cleanly for both solo and team development.

## Positioning

SpecKeep is a strict lightweight SDD kit for real codebases.

It is designed for teams that want more discipline than a loose planning layer, but do not want the workflow surface, artifact overhead, or orchestration weight of a heavier SDD system.

- stricter than OpenSpec in phase discipline and artifact alignment
- lighter than Spec Kit in default context, workflow surface, and artifact overhead
- optimized for agent-first workflows with narrow context loading
- designed to keep strictness in templates, entrypoints, and readiness checks rather than heavyweight orchestration
- built for brownfield repositories where context must stay narrow, local, and reviewable

In short: SpecKeep aims to maximize discipline per token: strong phase boundaries, low artifact drag, and enough structure to keep agents and humans aligned in everyday work.

## SpecKeep vs OpenSpec vs Spec Kit

| Dimension                | SpecKeep                              | OpenSpec                             | Spec Kit                         |
| ------------------------ | -------------------------------------- | ------------------------------------ | -------------------------------- |
| Workflow style           | Strict phase chain with narrow context | Fluid artifact-guided workflow       | Thorough multi-step SDD workflow |
| Default context size     | Smallest by default                    | Moderate                             | Largest                          |
| Artifact overhead        | Low                                    | Medium                               | High                             |
| Phase discipline         | High                                   | Medium                               | Highest                          |
| Brownfield ergonomics    | High                                   | High                                 | Medium                           |
| Team collaboration model | Branch-first, feature-local artifacts  | Change-folder oriented               | Branch and workflow heavy        |
| Shared mutable state     | Avoided by design                      | Low                                  | Varies by setup                  |
| Best fit                 | Lean strict SDD on real codebases      | Flexible SDD-lite for fast iteration | Full-featured rigorous SDD       |

In short, SpecKeep aims to sit between OpenSpec and Spec Kit: stricter than OpenSpec, lighter than Spec Kit, and optimized for branch-based collaboration with minimal default context.

## Where SpecKeep Stands Out

- Narrow context by default. Each phase is designed to load the smallest useful scope.
- Code reading should stay phase-local and targeted: enough to remove guesswork, not enough to recreate full-repository context.
- Strict workflow chain. Constitution, spec, inspect, plan, tasks, and implementation stay aligned.
- `inspect` is a real quality gate, not a loose suggestion before planning.
- Lightweight traceability. Stable IDs and cheap readiness checks reduce prompt bloat.
- Brownfield-friendly workflow. SpecKeep works well in existing repositories without forcing a heavyweight process layer.
- Branch-first collaboration. Active feature state stays local to the feature instead of spreading through shared mutable memory.
- Inspect is mandatory before planning. Each feature should persist an inspect report that confirms the spec is ready for plan work.
- Optional workflow commands available at any phase: `/speckeep.challenge` (adversarial review — finds weak assumptions and untestable criteria), `/speckeep.handoff` (compact session handoff document so a new session can resume without re-reading all artifacts), `/speckeep.hotfix` (emergency fix outside the standard phase chain — for well-understood fixes touching ≤ 3 files), `/speckeep.scope` (quick scope boundary check, inline only, no file written).

OpenSpec is more flexible by design and works well when teams want a looser artifact-guided workflow.

Spec Kit provides a broader and more thorough workflow surface, but usually at the cost of more artifacts, more context, and more process overhead.

SpecKeep is optimized for discipline per token: strong workflow boundaries, minimal default context, explicit quality gates, and enough structure to keep agents aligned without making the workflow heavy.

## Documentation

Extended documentation lives in [`docs/`](docs/README.md):

- [English docs](docs/en/index.md)
- [Русская документация](docs/ru/index.md)

## Public CLI

```text
speckeep init [path]
speckeep refresh [path]
speckeep add-agent [path]
speckeep list-agents [path]
speckeep remove-agent [path]
speckeep cleanup-agents [path]
speckeep doctor [path]
speckeep doctor [path] --json
speckeep dashboard [path]
speckeep feature <slug> [path]
speckeep feature repair <slug> [path]
speckeep features [path]
speckeep migrate [path]
speckeep list-specs [path]
speckeep show-spec <name> [path]
speckeep check <slug> [path]
speckeep check <slug> [path] --json
speckeep check [path] --all
speckeep check [path] --all --json
speckeep trace [slug] [path]
speckeep demo [path]
speckeep export <slug> [path]
speckeep export <slug> [path] --output <file>
speckeep list-archive [path]
speckeep list-archive [path] --status <status>
speckeep list-archive [path] --since <YYYY-MM-DD>
speckeep list-archive [path] --json
```

## Workflow

```text
constitution -> spec -> inspect -> plan -> tasks -> implement -> verify -> archive
```

## Key Points

- The constitution is the highest-priority project document.
- Plan packages keep `plan.md`, `tasks.md`, `data-model.md`, `contracts/`, and optional `research.md` together.
- `data-model.md` and `contracts/` are intentionally compact but structured: entities should capture fields, invariants, and lifecycle; contracts should capture boundary IO, failures, and delivery assumptions.
- Specs use canonical `Given / When / Then` markers across documentation languages.
- SpecKeep prefers stable IDs and explicit references over repeated narrative summaries: `RQ-*` for requirements, `AC-*` for acceptance criteria, `DEC-*` for plan decisions, and phase-scoped `T*` task IDs.
- Agent workflows are designed to load only the minimum context required.
- Strictness comes from phase entrypoints, templates, stable artifact structure, and readiness checks rather than large default prompts.
- `inspect` now treats helper-script output as the primary structural evidence layer: readiness checks can emit categorized findings such as `structure`, `traceability`, `ambiguity`, `consistency`, and `readiness`, which the agent should preserve and only deepen when necessary.
- Agent-facing `/speckeep.spec` is branch-first: it should work from `feature/<slug>`, support `--name` with optional `--slug` / `--branch`, and still prefer explicit `name:` / `slug:` metadata for prompt files.
- `speckeep init` requires an explicit `--shell` and generates one script family: `sh` or `powershell`. Supported agent targets: `claude`, `codex`, `copilot`, `cursor`, `kilocode`, `trae`, `windsurf`, `roocode`, `aider`.
- Generated workspaces include `.speckeep/scripts/run-speckeep.*` as the stable CLI launcher for agents; it resolves `DRAFTSPEC_BIN` first and falls back to `speckeep` from `PATH`.
- Generated `.speckeep/scripts/*` wrappers compute the project root from the script location and pass it via `--root`, so they can be executed from any working directory.
- `speckeep feature repair` and `speckeep migrate` provide safe canonicalization for legacy artifacts such as old inspect report paths.
- `speckeep check <slug>` shows artifact presence, inspect and verify verdict, task progress, the exact next slash command, and a compact readiness summary from structured checks; exits with code 1 when blocked; supports `--json` for CI use. `--all` shows a readiness table across all features.
- `speckeep demo [path]` creates a demo workspace pre-populated with an example feature at the implement phase — spec, inspect report, plan, tasks, and data model are all populated.
- `speckeep export <slug>` bundles all feature artifacts into one markdown document for sharing with a reviewer or new agent session; supports `--output` to write to a file.
- `speckeep list-archive` lists archived features from `.speckeep/archive/`; shows one entry per slug (most recent snapshot) with status, date, and reason; supports `--status` to filter by archive status, `--since <YYYY-MM-DD>` to filter by date, and `--json` for automation.
- `/speckeep.plan` supports `--research`: enters research-first mode — agent identifies concrete unknowns, writes `research.md`, then asks "Research complete — proceed to full plan?" before producing `plan.md`.
- `/speckeep.plan` includes `## Incremental Delivery`: guides agents to define MVP (smallest testable increment) and plan iterative expansion steps with AC traceability — avoids monolithic implementations and enables early validation.
- `/speckeep.spec` supports `--amend`: targeted edit mode — update one section or add one criterion without rewriting the spec or invalidating an existing inspect report.
- `/speckeep.handoff` without a slug generates handoff documents for all active features at once.
- `/speckeep.hotfix`: emergency fix workflow — writes a minimal hotfix spec (fix, root cause, risk, verification, touches) before any code change, implements, verifies inline, then archives; skips inspect, plan, and tasks phases; use only when the root cause is known and the fix touches ≤ 3 files.
- `doctor` warns when the same stable ID (`AC-*`, `RQ-*`) appears across multiple specs.
- **Traceability by design**. Agents are instructed to annotate code with `// @sk-task <ID> (<AC_ID>)` during implementation. Use `speckeep trace <slug>` to scan and verify these links between code and requirements.
- Generated docs and prompts support English and Russian.

- **Greenfield-friendly**. While SpecKeep is optimized for brownfield, it works great for from-scratch projects using a "Foundation-first" approach.

## Quick Start (Greenfield)

If you are starting a project from scratch:

1.  **Init**: `speckeep init . --lang en --shell sh`
2.  **Establishment**: Define the tech stack, architecture, and rules via `/speckeep.constitution --foundation`. This creates a unified document for project rules and technical foundation.
3.  **First Feature**: Once the baseline is established, move to the first functional specification via `/speckeep.spec`.

## Usage Example (Brownfield)

```bash
# try the demo instantly — no project setup required
speckeep demo ./my-demo

# init a real project
speckeep init my-project --lang en --shell sh --agents claude --agents codex
speckeep refresh my-project --shell powershell --dry-run
speckeep doctor my-project
speckeep check export-report my-project
```

## Example Feature Cycle

A full workflow for a real feature — "Add CSV export to the reports page".

<details>
<summary>See the full cycle →</summary>

### 1. Init

```bash
speckeep init . --lang en --shell sh --agents claude
# → .speckeep/ scaffold, AGENTS.md, agent slash-command files written
```

Optional advanced flags:

- `--specs-dir` overrides where specs are stored (default: `.speckeep/specs`)
- `--archive-dir` overrides where archived artifacts are stored (default: `.speckeep/archive`)

### 2. Write the spec

Call `/speckeep.spec --name "CSV export for reports"` in your agent.

`.speckeep/specs/csv-export-for-reports/spec.md`:

```markdown
## Goal

Allow users to download the reports table as a CSV file.

## Acceptance Criteria

**AC-001** Export produces a file
Given the Reports page has at least one row
When the user clicks "Export CSV"
Then a .csv file downloads with column headers and all visible rows

**AC-002** Empty state is handled
Given the reports table is empty
When the user clicks "Export CSV"
Then a .csv with headers only downloads — no error shown
```

### 3. Inspect

Call `/speckeep.inspect csv-export-for-reports`.

- `.speckeep/specs/csv-export-for-reports/inspect.md` — verdict `pass`, all AC have G/W/T
- `.speckeep/specs/csv-export-for-reports/summary.md` — compact AC table used by implement and verify instead of the full spec

### 4. Plan

Call `/speckeep.plan csv-export-for-reports`.

`.speckeep/specs/csv-export-for-reports/plan/plan.md` names the implementation surfaces: `ReportsPage.tsx` (add button), `useReportExport.ts` (new hook, CSV logic), `reports.test.ts` (new tests).

### 5. Tasks

Call `/speckeep.tasks csv-export-for-reports`.

`.speckeep/specs/csv-export-for-reports/plan/tasks.md`:

```markdown
## Surface Map

| Surface                    | Tasks |
| -------------------------- | ----- |
| hooks/useReportExport.ts   | T1.1  |
| components/ReportsPage.tsx | T1.2  |
| tests/reports.test.ts      | T2.1  |

## Phase 1: Hook and button

- [ ] T1.1 add `useReportExport` hook — converts `rows[]` to CSV blob and triggers browser download — AC-001 `Touches: hooks/useReportExport.ts`
- [ ] T1.2 add Export CSV button to ReportsPage — calls hook on click, disabled when rows empty — AC-001, AC-002 `Touches: components/ReportsPage.tsx`

## Phase 2: Tests

- [ ] T2.1 add tests for useReportExport — covers non-empty rows, empty rows, header-only output — AC-001, AC-002 `Touches: tests/reports.test.ts`

## Acceptance Coverage

AC-001 → T1.1, T1.2, T2.1
AC-002 → T1.2, T2.1
```

### 6. Implement, verify, archive

```
/speckeep.implement csv-export-for-reports   # Phase 1 done, stops
/speckeep.implement csv-export-for-reports   # Phase 2 done, stops
/speckeep.verify    csv-export-for-reports   # verdict: pass
/speckeep.archive   csv-export-for-reports
```

### Check readiness at any point

```bash
speckeep check csv-export-for-reports
# Phase:  tasks → implement
# Tasks:  0 / 3 done
# Next:   /speckeep.implement csv-export-for-reports
```

</details>

## Demo

A reproducible terminal demo kit lives under [`demo/`](demo/README.md).

Build the local binary and render the quick terminal demo:

```bash
go build -o bin/speckeep ./src/cmd/speckeep
vhs demo/quick.tape
```

Demo assets:

- [Quick terminal demo](demo/README.md)
- [Brownfield walkthrough](demo/brownfield.md)
- [Self-hosting walkthrough](demo/self-hosting.md)

The quick tape produces `demo/speckeep-demo.gif` and demonstrates `init`, generated agent files, `AGENTS.md`, and launcher-based `doctor` / `refresh --dry-run`.

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

For deeper guidance, use:

- [Overview](docs/en/overview.md)
- [CLI Reference](docs/en/cli.md)
- [Workflow Model](docs/en/workflow.md)
- [Examples](docs/en/examples.md)
- [Roadmap](docs/en/roadmap.md)

Project contribution and trust docs:

- [Contributing](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Security Policy](SECURITY.md)

## Development

```bash
go test ./...
go build -o bin/speckeep ./src/cmd/speckeep

# with version stamp
go build -ldflags "-X speckeep/src/internal/cli.Version=v0.1.0" -o bin/speckeep ./src/cmd/speckeep
```

The repository includes unit tests for config, project lifecycle, doctor checks, specs, templates, agents, and CLI-level behavior.

## License

Released under the [MIT License](LICENSE).
