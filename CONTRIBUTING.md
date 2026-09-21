# Contributing

Thanks for your interest in improving `speckeep`.

This project aims to stay lightweight, predictable, and practical for real codebases. Contributions are welcome, especially when they improve:

- agent-first workflow quality
- brownfield ergonomics
- docs and prompt clarity
- token discipline
- reliability of generated assets, scripts, and checks

## Before You Start

Please read these first:

- `README.md`
- `MVP.md`
- relevant Go files under `src/cmd/` and `src/internal/`

If you are working inside this repository itself, keep in mind:

- `/.speckeep/` is generated test output
- `/AGENTS.md` is generated test output
- neither should be treated as canonical product source
- the product source of truth lives in `README.md`, `MVP.md`, Go source, and embedded template assets under `src/internal/templates/assets/`

## Development Setup

Build and test locally with:

```bash
go test ./...
go build -o bin/speckeep ./src/cmd/speckeep
```

If your change affects generated templates, scripts, prompts, or agent files, also verify the generated output path that `speckeep init` or `speckeep refresh` would produce.

## Release Gate

Before tagging a release, run the full gate:

```bash
sh scripts/release-check.sh          # full: static + -race + cross-builds + e2e
sh scripts/release-check.sh --quick  # fast: static + e2e, no -race / matrix
```

It runs, in order:

- **L0** — `gofmt` (tracked files), `go vet`, `go test [-race]`
- **L1/L3** — the agent-artifact golden matrix (`TestAgentArtifactMatrix`) and the upgrade-from-legacy check (`TestUpgradeFromLegacyLayout`)
- **build matrix** — `GOOS/GOARCH` builds for linux/darwin/windows × amd64/arm64 (skipped with `--quick`)
- **L2 e2e** — builds the binary and black-box tests `speckeep demo --agents all`: expected artifacts per target, OpenCode skills-only, `check-ready` no-op for auxiliary commands, `refresh` idempotency, `doctor`
- **L4** — `install.sh` and npm-launcher syntax smoke

The same gate runs in CI: `--quick` on pull requests, full on pushes to `main`/`master`.

## Workflow Expectations

- Keep changes focused and reviewable.
- Prefer small pull requests over broad mixed refactors.
- Preserve SpecKeep's lightweight design goals unless there is a strong reason to widen the workflow.
- Keep generated assets, docs, prompts, config, and implementation aligned when behavior changes.
- Do not commit generated `/.speckeep/` or repository-local `AGENTS.md` test artifacts from self-hosting runs.

When changing CLI behavior, please update the relevant combination of:

- Go implementation
- embedded template assets
- `README.md`
- `MVP.md`
- tests

## Pull Requests

For pull requests, please:

- explain the user-visible problem being solved
- describe the chosen approach and any tradeoffs
- mention docs or template updates when behavior changes
- include tests when the change affects behavior
- call out anything intentionally left for follow-up

Helpful PR scopes include:

- bug fixes
- prompt or template consistency fixes
- CLI ergonomics improvements
- doctor, refresh, or migration improvements
- brownfield workflow improvements
- docs and examples

## Design Guardrails

Please try to preserve these project constraints:

- narrow default context
- strict phase discipline
- cheap readiness checks over heavy orchestration
- compact artifacts over artifact sprawl
- explicit traceability over repeated summary layers

If a proposal adds substantial process surface, wider default reads, or new mandatory artifacts, explain why the added weight is worth it.

## Questions And Feedback

If you are unsure whether a change fits the project direction, opening an issue or discussion before a larger PR is appreciated.
