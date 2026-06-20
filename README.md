# speckeep

[![ci](https://github.com/bzdvdn/speckeep/actions/workflows/ci.yml/badge.svg)](https://github.com/bzdvdn/speckeep/actions/workflows/ci.yml)
[![release-build](https://github.com/bzdvdn/speckeep/actions/workflows/release-build.yml/badge.svg)](https://github.com/bzdvdn/speckeep/actions/workflows/release-build.yml)

Русская версия: [README.ru.md](README.ru.md)

`speckeep` — a lightweight Spec-Driven Development kit for dev agents and humans. It keeps specs, plans, tasks, and traceability in simple files so agents and people stay aligned without heavy process.

SpecKeep is the successor to DraftSpec (archived). Migrate with `speckeep migrate`.

---

## Quick Start — 30 seconds

```bash
# 1. Try it instantly — no project setup needed
speckeep demo ./my-demo

# 2. See what was created
speckeep dashboard ./my-demo

# 3. Init a real project
speckeep init my-project --lang en --shell sh --agents claude
```

That's it. You now have a project with constitution, spec, plan, tasks, inspect report, and data model — plus agent prompt files and `AGENTS.md`.

---

## Why speckeep?

Teams using AI coding agents face a common problem: agents lose context between sessions, drift from requirements, and produce untestable code.

speckeep solves this with **discipline per token** — minimal file-based structure that keeps agents and humans on the same page:

- **Specs** with stable IDs (`AC-*`, `RQ-*`) — agents know exactly what to build and verify
- **Tasks** with surface maps and phase grouping — agents execute in order, one phase at a time
- **Traceability** (`@sk-task` annotations) — verify that every requirement is implemented and tested
- **10 agent adapters** — Claude Code, Cursor, Copilot, OpenCode, aider, Windsurf, and more

Results in practice: agents produce correct code on first try more often, handoffs between sessions cost less context, and requirements stay reviewable by humans.

---

## Workflow

```
constitution → spec → [inspect] → plan → tasks → implement → verify → archive
```

Each phase loads only the minimum context. Optional workflow commands available at any phase: `/speckeep.challenge`, `/speckeep.handoff`, `/speckeep.hotfix`, `/speckeep.scope`, `/speckeep.recap`.

---

## Positioning

| Dimension | speckeep | OpenSpec | Spec Kit |
|---|---|---|---|
| Workflow style | Strict phase chain, narrow context | Fluid artifact-guided | Thorough multi-step SDD |
| Default context | Smallest | Moderate | Largest |
| Artifact overhead | Low | Medium | High |
| Brownfield | High | High | Medium |
| Collaboration | Branch-first, feature-local | Change-folder oriented | Branch-heavy |
| Best fit | Lean strict SDD on real codebases | Flexible SDD-lite | Full-featured rigorous SDD |

---

## CLI

```text
speckeep init [path]
speckeep refresh [path]
speckeep doctor [path] [--json]
speckeep dashboard [path]
speckeep check <slug> [path] [--json]
speckeep check [path] --all [--json]
speckeep feature <slug> [path]
speckeep features [path]
speckeep list-specs [path]
speckeep show-spec <name> [path]
speckeep trace <slug> [path]
speckeep export <slug> [path] [--output <file>]
speckeep demo [path]
speckeep archive <slug> [path]
speckeep list-archive [path] [--status <status>] [--since <YYYY-MM-DD>] [--json]
speckeep migrate [path]
speckeep add-agent | list-agents | remove-agent | cleanup-agents [path]
speckeep add-skill | list-skills | remove-skill | install-skills | skills-restore [path]
```

---

## Install

**Linux / macOS:**

```bash
VERSION=v0.5.1
curl -fsSL "https://raw.githubusercontent.com/bzdvdn/speckeep/${VERSION}/scripts/install.sh" | bash -s -- --version "${VERSION}"
```

**Windows (PowerShell):**

```powershell
$version="v0.5.1"
$env:SPECKEEP_VERSION=$version
powershell -ExecutionPolicy Bypass -c "iwr -useb https://raw.githubusercontent.com/bzdvdn/speckeep/$version/scripts/install.ps1 | iex"
```

Add `--add-to-path` (Linux) or `-AddToPath` (Windows) to add the install directory to `PATH`.

**Go users:**

```bash
go install speckeep@latest
```

**Build from source:**

```bash
go build -ldflags "-X speckeep/src/internal/cli.Version=v0.5.1" -o bin/speckeep ./src/cmd/speckeep
```

---

## Example Feature Cycle

<details>
<summary>Full workflow: "Add CSV export to reports" →</summary>

### 1. Init

```bash
speckeep init . --lang en --shell sh --agents claude
```

### 2. Spec

Call `/speckeep.spec --name "CSV export for reports"` in your agent.

`specs/active/csv-export-for-reports/spec.md`:

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

Call `/speckeep.inspect csv-export-for-reports`. Produces `inspect.md` with verdict.

### 4. Plan

Call `/speckeep.plan csv-export-for-reports`. Surfaces: `ReportsPage.tsx`, `useReportExport.ts`, `reports.test.ts`.

### 5. Tasks

Call `/speckeep.tasks csv-export-for-reports`. Produces `tasks.md`:

| Surface | Tasks |
|---|---|
| hooks/useReportExport.ts | T1.1 |
| components/ReportsPage.tsx | T1.2 |
| tests/reports.test.ts | T2.1 |

### 6. Implement, verify, archive

```
/speckeep.implement csv-export-for-reports
/speckeep.verify    csv-export-for-reports   # verdict: pass
speckeep archive    csv-export-for-reports .
```

### Check readiness at any point

```bash
speckeep check csv-export-for-reports
# Phase:  tasks → implement
# Tasks:  0 / 3 done
# Next:   /speckeep.implement csv-export-for-reports
```

</details>

---

## Key Concepts

### Artifacts

Each feature lives under `specs/<slug>/` with:

- `spec.md` — requirements (`RQ-*`) and acceptance criteria (`AC-*` with Given/When/Then)
- `inspect.md` (optional) — quality gate before planning
- `plan.md` — design decisions (`DEC-*`) and incremental delivery
- `tasks.md` — executable tasks with surface map and phase grouping
- `data-model.md` — entities, fields, invariants
- `contracts/api.md`, `contracts/events.md` (optional)
- `verify.md` — verification evidence

### Traceability

Annotate code during implementation:

```go
// @sk-task T1.1 (AC-001)
```

Then verify with:

```bash
speckeep trace <slug> .
```

### Agent adapters

Supported out of the box: `claude`, `codex`, `copilot`, `cursor`, `kilocode`, `opencode`, `trae`, `windsurf`, `roocode`, `aider`.

### Skills

Reusable guidance packages from local paths or git repos:

```bash
speckeep add-skill my-project --id architecture --from-local skills/architecture
speckeep install-skills my-project
```

---

## Documentation

Extended docs in [`docs/`](docs/README.md):

- [Overview](docs/en/overview.md)
- [CLI Reference](docs/en/cli.md)
- [Workflow Model](docs/en/workflow.md)
- [Architecture](docs/en/architecture.md)
- [Agents](docs/en/agents.md)
- [Examples](docs/en/examples.md)
- [FAQ](docs/en/faq.md)
- [Glossary](docs/en/glossary.md)
- [Roadmap](docs/en/roadmap.md)

Project:

- [Contributing](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Security Policy](SECURITY.md)
- [MVP Definition](MVP.md)
- [Changelog](CHANGELOG.md)

## Demo

A reproducible terminal demo is available under [`demo/`](demo/README.md):

```bash
go build -o bin/speckeep ./src/cmd/speckeep
vhs demo/quick.tape
```

## Development

Requires **Go 1.26+**.

```bash
go test ./...
go vet ./...
go build -o bin/speckeep ./src/cmd/speckeep
```

## License

Released under the [MIT License](LICENSE).
