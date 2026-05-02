# Architecture

This page explains how SpecKeep is structured internally.

## High-Level Layers

SpecKeep is split into a few practical layers:

- `src/cmd/speckeep/` for the CLI entrypoint
- `src/internal/cli/` for Cobra commands and user-facing command wiring
- `src/internal/project/` for workspace lifecycle operations such as `init`, agent maintenance, and cleanup
- `src/internal/config/` for loading, defaulting, and saving `.speckeep/speckeep.yaml`
- `src/internal/templates/` for localized generated assets and file generation
- `src/internal/agents/` for project-local command or prompt file generation
- `src/internal/specs/` for reading spec files used by the public CLI
- `src/internal/trace/` for scanning traceability annotations in code
- `src/internal/doctor/` for workspace health checks

## CLI Layer

The public CLI is intentionally small. `src/internal/cli/` wires commands such as:

- `init`
- `add-agent`
- `list-agents`
- `remove-agent`
- `cleanup-agents`
- `doctor`
- `dashboard`
- `trace`
- `list-specs`
- `show-spec`

The CLI layer should stay thin. Most behavior belongs in project, config, templates, agents, and doctor packages.

## Config Layer

The config source of truth lives in `.speckeep/speckeep.yaml`.

`src/internal/config/config.go` is responsible for:

- defining the config schema
- applying defaults
- resolving important workspace paths
- loading config from disk
- saving updated config back to disk

This is what allows the rest of the codebase to avoid hardcoding too many `.speckeep/...` paths.

## Template and Asset Layer

SpecKeep generates files from localized assets stored under:

- `src/internal/templates/assets/lang/en/`
- `src/internal/templates/assets/lang/ru/`

These include:

- `constitution.md`
- spec, plan, tasks, and archive templates
- prompts for `constitution`, `spec`, `inspect`, `plan`, `tasks`, `implement`, and `verify`
- localized `agents-snippet.md`

Shared shell scripts live in:

- `src/internal/templates/assets/scripts/`

The `templates` package assembles these assets into the generated `.speckeep/` workspace.

## Agent Generation Layer

`src/internal/agents/files.go` generates project-local files for supported targets:

- `claude`
- `codex`
- `copilot`
- `cursor`
- `kilocode`
- `trae`

These generated files are wrappers that point back to canonical SpecKeep prompts inside `.speckeep/templates/prompts/`.

This keeps one main source of truth for workflow prompts while still supporting multiple agent ecosystems.

## Project Lifecycle Layer

`src/internal/project/init.go` handles the workspace lifecycle:

- initialize a project
- add agent targets
- list agent targets
- remove agent targets
- clean up orphaned agent artifacts
- create helper scripts in `.speckeep/scripts/`

This layer is responsible for:

- creating the `.speckeep/` layout
- writing generated files
- updating `AGENTS.md`
- syncing enabled agent targets into config

## Traceability and Verification Layer

`src/internal/trace/trace.go` handles the core logic for linking code to requirements:

- scans files for `@sk-task` and `@sk-test` annotations
- filters findings by feature slug
- supports JSON output for integration with verification workflows

This layer allows the `verify` phase to remain token-efficient by replacing manual code review with deterministic metadata scanning.

## Health and Maintenance Layer

`src/internal/doctor/doctor.go` and `src/internal/gitutil/` check workspace health and alignment.

It verifies:

- required directories and files exist
- configured languages are valid
- enabled agent targets have their generated files
- disabled targets do not leave stale artifacts behind unnoticed
- **Smart Branching**: checks if the current Git branch matches the feature slug (expected `feature/<slug>`)
- **Traceability Integrity**: warns about orphaned `@sk-task` annotations or invalid `AC-*` references

Related maintenance commands:

- `remove-agent` disables a target and removes its generated files
- `cleanup-agents` removes orphaned leftovers for disabled targets
- `doctor` reports `error`, `warning`, and `ok`
- `dashboard` provides a visual summary of project progress, status, and branch health

## Design Principles

The internal architecture follows a few important principles:

- keep the public CLI small
- keep generated assets language-aware but structurally consistent
- push readiness checks into shell scripts when possible
- prefer one canonical prompt source over duplicated prompt logic
- keep workflow phases strict without embedding a heavy orchestration engine
- optimize for low token usage by controlling read sets and artifact scope

## Anti-Bloat Checklist

Use this checklist before adding a new feature, prompt rule, script, or artifact:

- Does it increase the default read set? If yes, treat it as risky by default.
- Can it be solved with a cheap deterministic helper instead of more reasoning?
- Does it make a new artifact mandatory for every feature? If yes, reconsider.
- Does it require reading implementation code by default? If yes, it is probably too heavy.
- Can the workflow still begin from constitution, spec, plan, or tasks before touching code?
- Does it expand the public CLI without clear value for everyday users?
- Does it add a brand-new gate, or can it strengthen an existing phase instead?
- Can its value be explained in one short sentence? If not, it may be process complexity without enough return.
- Does it push SpecKeep closer to spec-kit-style bureaucracy without matching value?
- Does it improve brownfield ergonomics in a real way?

A change is usually a good fit when it does at least one of these:

- improves quality without expanding the default context
- replaces expensive reasoning with a cheap structural check
