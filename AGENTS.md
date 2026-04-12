# speckeep — Agent Guide

## Project Map

This is the **speckeep** repository — a lightweight SDD (Specification-Driven Development) kit for development agents.

### Key Directories

- `src/` — Go source code (CLI implementation)
  - `src/cmd/speckeep/` — main CLI entrypoint
  - `src/internal/` — internal packages (templates, cli, config, etc.)
- `.speckeep/` — speckeep's own project context (dogfooding)
  - `specs/` — feature specifications for speckeep itself
  - `specs/<slug>/plan/` — plan package per feature
  - `templates/` — template files that speckeep generates for other projects
  - `scripts/` — helper scripts for workflow phases
- `docs/` — extended documentation (en/ru)
- `demo/` — demo assets and terminal tapes

### Important Files

- `README.md` / `README.ru.md` — project overview and usage
- `CHANGELOG.md` — release history
- `MVP.md` — product definition and roadmap
- `go.mod` — Go module definition
- `.speckeep/constitution.md` — project constitution (highest-priority rules)

## Workflow

speckeep follows a strict phase chain:

```
constitution → spec → inspect → plan → tasks → implement → verify → archive
```

When working on speckeep features:

1. **Check existing specs**: `speckeep list-specs`
2. **Read the spec**: `.speckeep/specs/<slug>/spec.md`
3. **Check inspect report**: `.speckeep/specs/<slug>/inspect.md` (must be `pass`)
4. **Read the plan**: `.speckeep/specs/<slug>/plan/plan.md`
5. **Read tasks**: `.speckeep/specs/<slug>/plan/tasks.md`
6. **Implement tasks in order**, marking them complete
7. **Verify**: ensure all ACs are covered

## Development Rules

### Go Standards

- Run `go fmt ./...` before committing
- Run `go vet ./...` — must pass with no issues
- Run `go test ./...` — all tests must pass
- Follow Clean Architecture: domain → application → infrastructure

### Constitution Compliance

Always read `.speckeep/constitution.md` before starting work. Key principles:

- **Interface abstraction**: external dependencies through Go interfaces
- **Clean Architecture**: domain layer imports only standard library
- **Context safety**: all operations accept `context.Context` as first parameter
- **Testability**: every public interface must have mock implementations
- **Minimal configuration**: package works out of the box with sensible defaults

### Language Policy

- Documentation: Russian (docs/README.ru.md, constitution.md)
- Agent communication: Russian
- Code comments: Russian (godoc)
- Variable/function names: English (Go standard)

## CLI Commands

```bash
# Check workspace health
speckeep doctor .

# List active features
speckeep list-specs .

# Show feature readiness
speckeep check <slug> .

# Show all features status
speckeep check . --all

# Visual dashboard
speckeep dashboard .

# Trace requirements in code
speckeep trace <slug> .

# Export feature artifacts
speckeep export <slug> . --output feature.md
```

## Template Updates

When modifying templates in `src/internal/templates/assets/`:

- Update both `lang/ru/` and `lang/en/` versions
- Test with `speckeep demo ./test-demo`
- Regenerate with `speckeep refresh .`

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Build with version stamp
go build -ldflags "-X speckeep/src/internal/cli.Version=v0.2.0" -o bin/speckeep ./src/cmd/speckeep
```

## Key Design Principles

1. **Discipline per token**: maximize workflow strictness while minimizing context size
2. **Branch-first**: each feature lives in `feature/<slug>` branch
3. **Stable IDs**: `AC-*`, `RQ-*`, `DEC-*`, `T*` for traceability
4. **Narrow context**: each phase loads only the minimum required artifacts
5. **No shared mutable state**: feature state stays local to the feature branch
