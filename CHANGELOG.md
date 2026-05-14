# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versions follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- New `opencode` agent target support:
  - generates SpecKeep workflow wrappers under `.opencode/commands/`
  - included in agent target normalization, refresh/cleanup flows, CLI help text, and EN/RU docs

### Changed

- Lean feature artifact layout is now the default for generated guidance and readiness checks:
  - canonical active artifacts now center on `spec.md`, optional `inspect.md`, `plan.md`, `tasks.md`, `data-model.md`, `contracts/`, and `verify.md`
  - generated prompts no longer require `summary.md`, `spec.digest.md`, or `plan.digest.md`
  - `tasks.md` now carries an explicit `Implementation Context` section as the main operational bridge for `implement` and `verify`
  - `refresh` and new workspaces treat old summary/digest files as legacy optional artifacts rather than canonical defaults

## [v0.3.1] - 2026-05-12

### Changed

- Agent phase prompts now resolve constitution context more reliably:
  - when `.speckeep/constitution.summary.md` exists, agents are instructed to prefer it over the full constitution file
  - this guidance is now reinforced both in generated phase prompts and in agent-target wrappers
- Next-step guidance after each phase is now stricter and more explicit:
  - `spec` now requires a final `Ready for:` / `Готово к:` line pointing to either `inspect` or `plan`
  - `inspect` now distinguishes pass/concerns vs blocked outcomes, returning to `spec` when refinement is required
  - generated wrappers for Codex, Windsurf, Trae, and other targets now emphasize preserving the exact final next-command line from the prompt
- Trae adapter generation now matches the per-command wrapper model used by other agent targets:
  - generated workflow files now live under `.trae/rules/`
  - Trae no longer relies on a single aggregated `.trae/project_rules.md`

### Fixed

- Reduced cases where agents missed `.speckeep/constitution.summary.md` and read only the adjacent constitution file instead
- Reduced cases where agents completed a phase response without surfacing the next command for the user

## [v0.3.0] - 2026-05-07

### Added

- Repository map guidance tightened in generated agent instructions:
  - agents should read `REPOSITORY_MAP.md` before broad file discovery when it exists
  - implement now carries an explicit `Map update: yes|no` decision
  - `/speckeep.repo-map` remains the canonical way to refresh repository navigation notes

### Changed

- Archive is now CLI-first after verification:
  - agent-facing workflow ends at `verify`
  - successful verify now points to `speckeep archive <slug> .` instead of `/speckeep.archive <slug>`
  - generated agent wrappers no longer include an `archive` phase prompt
- Default feature storage layout is now nested under `specs/`:
  - active feature packages default to `specs/active/<slug>/`
  - archived snapshots default to `specs/archived/<slug>/<YYYY-MM-DD>/`
  - docs, examples, demo assets, and generated agent guidance now consistently reflect the new defaults
- `refresh` now auto-migrates the legacy default layout `specs/` + `archive/` to `specs/active/` + `specs/archived/` when paths were not explicitly customized
- `refresh` now removes deprecated archive-managed artifacts automatically:
  - legacy `.speckeep/templates/prompts/archive.md`
  - legacy agent wrapper files such as `speckeep.archive.md` / `speckeep-archive.mdc`
- `doctor` now treats archive as CLI-only operational follow-up rather than an agent prompt dependency

### Fixed

- Workspace health/reporting consistency around deprecated archive artifacts:
  - `doctor` warns when old `/speckeep.archive` guidance or legacy archive wrappers remain in the workspace
  - `doctor` warns when a workspace still uses the legacy default layout or a mixed old/new layout
  - `refresh` and generated assets stay aligned after the archive prompt removal

## [v0.2.0] - 2026-04-28

### Added

- Skills subsystem under `.speckeep/skills/manifest.yaml` with CLI commands:
  - `speckeep add-skill`, `speckeep list-skills`, `speckeep remove-skill`
  - `speckeep install-skills`, `speckeep sync-skills`
  - grouped subcommands: `speckeep skills install`, `speckeep skills sync`
- Skill sources:
  - local directories via `--from-local`
  - git sources via `--from-git` with required pinned `--ref` (tag/commit)
- Git skill materialization:
  - clone + checkout into `.speckeep/skills/checkouts/<id>`
  - stored `resolved_commit` and `checkout_dir` in manifest entries
- Skills validation in `doctor` (manifest consistency, refs, local paths, checkout state)
- Skills section in managed SpecKeep block in `AGENTS.md`
- Sync path for skill artifacts (`refresh` + dedicated `sync-skills`)
- Skill installation/reconciliation into target agent folders:
  - `.codex/skills/<id>`
  - `.claude/skills/<id>`
  - `.kilocode/skills/<id>`
  - `.windsurf/skills/<id>`
  - `.trae/skills/<id>`
- Optional install skip flag for mutation commands:
  - `speckeep add-skill --no-install`
  - `speckeep remove-skill --no-install`
- Digest artifacts support for feature lifecycle (archive/doctor/templates integration)
- Traceability improvements:
  - slug-defined trace handling
  - updated trace scripts/templates and `trace` command behavior
- Expanded agent wrapper generation updates across adapters (Claude, Codex, Copilot, Cursor, Kilocode, Roocode, Windsurf, Aider, Trae)

### Changed

- Workflow chain now treats inspect as optional gate:
  - `constitution → spec → [inspect, optional] → plan → tasks → implement → verify → archive`
  - if inspect report exists, it must remain valid and non-blocking
- `add-skill`/`remove-skill` now auto-install/reconcile skills in target agent folders by default
- Prompt/template system optimized and reworked for both `en` and `ru`:
  - stricter output expectations and readiness behavior
  - updated prompt packs for `constitution/spec/inspect/plan/tasks/implement/verify/archive` and optional commands
  - updated agents snippets and embedded assets
- Workflow guidance tightened to reduce overhead and scope drift during execution
- Documentation updated (EN/RU): README and CLI docs for skills lifecycle, git pinning, install/sync commands, and optional inspect
- CLI/help/schema text updated to reflect optional inspect and new skills commands

### Fixed

- `doctor`: fixed handling of inactive specs and improved workspace findings robustness
- Workflow checks/state edge cases around inspect/implement/task readiness
- Agent command/rendering issues in wrappers and scripts for cross-agent consistency

### Changed

- Canonical feature artifact layout flattened from `specs/<slug>/plan/...` to `specs/<slug>/...` for `plan.md`, `tasks.md`, `data-model.md`, and `verify.md`; `contracts/` remains a dedicated subdirectory
- Added legacy fallback and safe migration support for existing `specs/<slug>/plan/` workspaces
- Documentation, prompts, templates, and examples updated to reflect the new artifact layout

## [v0.1.0] - 2026-04-16

### Added

- Initial release of the Speckeep CLI (specification-driven development kit for agents and humans)
- Canonical workspace under `.speckeep/` with file-based artifacts (specs, feature artifacts, reports, scripts, templates)
- Strict phase chain: `constitution → spec → inspect → plan → tasks → implement → verify → archive`
- Bilingual templates/prompts: English (`en`) and Russian (`ru`)
- Shell support: `sh` and `powershell`
- Core CLI:
  - `speckeep init`, `speckeep refresh`, `speckeep doctor`
  - `speckeep list-specs`, `speckeep show-spec`, `speckeep check`, `speckeep trace`
  - `speckeep feature`, `speckeep feature repair`, `speckeep features`, `speckeep migrate`
  - `speckeep export`, `speckeep demo`, `speckeep archive`, `speckeep list-archive`
- Managed agent integrations (generated wrapper files + prompts) for: Claude, Codex, Copilot, Cursor, Kilocode, Trae, Windsurf, Roocode, Aider
- Phase readiness scripts and internal CLI plumbing (`__internal`) to keep wrappers cheap and deterministic
- Stable IDs for traceability: `RQ-*`, `AC-*`, `DEC-*`, `T*` + acceptance coverage mapping (`AC-* -> T*`)
- Migration support from legacy `.draftspec/` workspace into `.speckeep/` (safe move/copy + path canonicalization)
- Extended `CheckInspectReady`: detects `[NEEDS CLARIFICATION]` markers, counts `RQ-*` IDs, warns on missing `## Assumptions` section, checks constitution language policy consistency
- Extended `CheckVerifyReady`, `CheckImplementReady`, `CheckTasksReady`: optional `summary.md` presence warning, `Touches:` file existence check, plan content validation (`DEC-*` IDs, `## Acceptance Approach`, `## Constitution Compliance`, AC alignment)
- Stricter verify report traceability: requires `## Checks` section with `task_state` and per-AC `acceptance_evidence` entries
- Package-level tests for `featurepaths` (17 tests) and `gitutil` (7 tests)
- Full workflow integration test (`TestFullWorkflowCycle`) covering the complete lifecycle from `init` through archive-readiness in a temporary directory

[0.2.0]: https://github.com/bzdvdn/speckeep/releases/tag/v0.2.0
[0.1.0]: https://github.com/bzdvdn/speckeep/releases/tag/v0.1.0
[0.3.1]: https://github.com/bzdvdn/speckeep/releases/tag/v0.3.1
[0.3.0]: https://github.com/bzdvdn/speckeep/releases/tag/v0.3.0
[unreleased]: https://github.com/bzdvdn/speckeep/compare/v0.3.1...HEAD
