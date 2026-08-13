# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versions follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.8.1] - 2026-08-13

### Fixed

- **`speckeep archive` now respects the `workflow.verify` gate**: `verify.md` is no longer hard-required for archiving — when verify is `optional` (the default), a feature with all tasks complete can be archived without a verify report; a persisted `verify.md` with status ≠ `pass` still vetoes the archive, and `workflow.verify: required` keeps the mandatory pre-archive gate. The generated `summary.md` no longer references a missing `verify.md` (it reports `verify: skipped (optional)`).
- **Archive tests**: added coverage for the optional-verify path (`archive` succeeds without `verify.md`), the `required` mode (`verify.md` must exist), and the non-pass veto (`verify.md` with status `concerns` blocks archiving).

## [v0.8.0] - 2026-08-12

### Added

- **`workflow.verify` config (`optional|required`, default `optional`)**: the `verify` phase is now an optional on-demand audit. Archive is CLI-only and allowed once the feature is deterministically proven (every `[x]` task in `tasks.md` has a `Proof:` entry) or after `verify: pass`. A persisted `verify.md` with status ≠ `pass` vetoes archive. Only `required` mode enforces a passing verify report before archive.
- **Prompt↔CLI consistency tests**: template references from `templates/prompts/*.md` are resolved against the embedded templates (no dead pointers); the `Ready for:`/`Return to:` final lines of the phase prompts are verified against the CLI state machine; and the single-source contract between the repo `AGENTS.md` and `agents-snippet.md` (canonical end block, verify gate policy, phase chain, branch-first, `workflow.verify`) is enforced.

### Changed

- **Traceability moved from code markers to `Proof:` entries in `tasks.md`**: inline `@sk-task` / `@sk-test` / `@ds-*` annotations are deprecated and no longer read. Completed tasks (`[x]`) must carry at least one `Proof:` line (`Proof: <kind> <path> [<anchor>]`, `kind` = `code|test|docs|chore`) directly below the checkbox. New enforcement: a checked task without `Proof:` fails `speckeep check`, `speckeep archive`, and `speckeep doctor` traceability checks.
- **`speckeep trace` rewritten**: now parses `Proof:` entries from `tasks.md` instead of scanning source files. Reports orphaned/duplicate/missing proof, missing files, and warns on missing anchors. `--tests` filters to `kind = test` entries. Legacy marker scanning removed.
- **`speckeep doctor`**: added deprecation warning for stray `@sk-task`/`@sk-test`/`@ds-*` markers found in source code (they are ignored, not read); orphaned/missing `Proof:` entries are hardened errors.
- **`speckeep refresh`**: removed `--rewrite-trace` and the `@ds-* → @sk-*` annotation rewrite logic.
- **Templates (EN/RU)**: `constitution.md` Definition of Done, `tasks.md`, `agents-snippet.md`, and `implement`/`verify` prompts updated to the `Proof:` model; workflow chain in prompts/snippets is now `constitution → spec → [inspect] → plan → tasks → implement → archive` with `verify` as an optional audit.
- **Prompt engineering pass (EN/RU)**: agent prompts now respect the `workflow.verify` mode when choosing the final `Ready for:` line (`required` → `/spk.verify`, `optional` → `speckeep archive`); mandatory Self-Check checklists added to `plan`, `tasks`, and `inspect` prompts (previously only `spec` had one); compact Given/When/Then example added to the `spec` prompt; persona roles added to `constitution`, `hotfix`, `rollback`, `scope`, `recap`, `repo-map`, and `handoff` prompts; a global escalation rule added to `agents-snippet.md` (stop with a precise reason instead of inventing a pass or a next step).
- **Legacy trace markers fully removed from generated guidance (EN/RU)**: templates (`agents-snippet.md`, `implement`, `tasks`, `constitution` prompts) and docs no longer mention `@sk-task`/`@sk-test`/`@ds-*` — fresh workspaces contain zero marker references, so `speckeep doctor` no longer flags its own generated AGENTS.md. `trace`/`doctor` continue to detect stray markers in pre-existing user code.
- **Docs (EN/RU)**: glossary, workflow, cli, examples, architecture, README, and MVP updated to the `Proof:` model and optional-verify semantics.
- **Prompt polish round 2 (EN/RU)**: `workflowChainHint` in `agents/files.go` aligned with the template chain (archive CLI-only, `[inspect, optional]`, verify on-demand); repo-map-first rule moved to the top of `agents-snippet.md` core rules (primacy); phase prompts (`spec`, `plan`, `tasks`, `inspect`, `implement`, `verify`) now carry an inline end-block format sample instead of referencing AGENTS.md; `hotfix` and `handoff` final lines resolve against `workflow.verify`; `repo-map` final line is now concrete (`Ready for: /spk.implement <slug>`) instead of the `<next phase>` placeholder; duplicate map-decision bullet removed from `implement` (RU).
- **Prompt engineering round 3 (EN/RU)**: the `verify` prompt gained a mandatory Self-Check tying the overall verdict to the acceptance-criteria matrix (evidence per row, `pass` only on confirmed evidence, `blocked` states the required refinement); cross-phase rules (canonical end block shape, verify gate policy) centralized once in `agents-snippet.md` and referenced by name from phase prompts; all self-checks (`spec`, `inspect`, `plan`, `tasks`, `verify`) gained a 2-round fix limit before stop; `spec`/`plan`/`tasks` gained size caps; inline `DEC-*`, task-row, and Surface-Map examples added; `challenge` and `scope` explicitly separated from `inspect` (findings/inventory only, no `pass|concerns|blocked` verdict); duplicated verify-policy and constitution-resolution pointers deduplicated; `/spk.handoff`, `/spk.scope`, and `/spk.hotfix` added to the snippet command list; repo `AGENTS.md` now documents `agents-snippet.md` as the single source of truth.

### Fixed

- **Doctor: false positives and uninitialized-project noise** — `speckeep doctor` on an uninitialized directory now fails with a single actionable message (run `speckeep init`) instead of 20+ misleading findings; markdown files are excluded from stray-marker scanning (docs legitimately describe the deprecated markers, so AGENTS.md was self-flagging); the deprecated-command check no longer matches the `.speckeep/speckeep.yaml` config path; per-slug "no safe migrations were needed" notices no longer surface as `connect` warnings.

## [v0.7.1] - 2026-07-08

### Fixed

- **Install scripts: stale versions, fragile `--version` check, legacy `draftspec` references**:
  - `scripts/install.sh`: example versions updated from `v0.5.1` to `v0.7.0`
  - `scripts/install.ps1`: `--version` call wrapped in try/catch — no longer crashes on success if the binary fails to run after install; `-AddToPath` flag inverted to default-on (`-NoPath` to opt out) — speckeep now automatically added to User PATH
  - `run-speckeep.ps1` / `run-speckeep.sh`: removed legacy `DRAFTSPEC_BIN` env var and `draftspec` binary fallbacks
  - `doctor`: removed `DRAFTSPEC_BIN` / `draftspec` checks from `speckeepEntrypointWarning`

## [v0.7.0] - 2026-07-07

### Changed

- **Agent slash commands shortened: `/speckeep.*` → `/spk.*`**:
  - All agent file paths renamed (e.g. `.claude/commands/speckeep.inspect.md` → `.claude/commands/spk.inspect.md`)
  - All adapter implementations updated (Claude, Codex, Copilot, Cursor, Kilocode, OpenCode, Roocode, Trae, Windsurf, Aider)
  - All prompt templates (EN/RU) for every phase: constitution, spec, inspect, plan, tasks, implement, verify, handoff, hotfix, rollback
  - CLI output: root help, `check`, `dashboard`, `demo`, `explore`, `init`, `status`, `archive`
  - README (EN/RU) and CLI docs (EN/RU) updated
  - `agents-snippet.md` simplified — removed the "⚠️ prefix is speckeep with a p" warning since `/spk` is unambiguous
  - `LegacyPrefixPaths()` function in `agents/files.go` enables discovery of stale `speckeep.*` artifacts
  - `doctor` warns about remaining old-prefix agent files
  - `refresh` auto-removes old-prefix files and regenerates with `/spk.*` naming
  - `CleanupAgents` also removes orphaned old-prefix artifacts

- **`speckeep init` now syncs skills manifest and `.gitignore`** — previously only `refresh` handled this; init was missing the skills setup step

- **Doctor: stronger deprecated command detection** — `/speckeep.*` detection now catches any remaining old-style references (was limited to `/speckeep.archive`)

### Fixed

- **Init: missing skills bootstrap** — `Initialize` now calls `syncSkillsManifest` and `syncSkillsGitignore`, matching `refresh` behavior

## [v0.6.0] - 2026-06-28

### Added

- **Templates: new spec/plan sections**:
  - `spec.md` — `## Dependencies` section for cross-spec and external reference linking
  - `plan.md` — `## Performance Budget` section for latency/memory/allocation limits
  - `agents-snippet.md` — `recap` and `challenge` commands added to command list
  - quality bar merged into `## Done` checklist
  - prompts optimized: redundant "don't invent" removed from per-phase prompts (now lives in agents-snippet only)
  - spec prompt requires explicit AC section before writing acceptance criteria
- **Base roles**: added base role guidance to system prompt templates (en/ru)

### Changed

- **Checks package refactored**: `checks.go` (1375 lines) split into 8 focused files:
  - `check_constitution.go`, `check_spec.go`, `check_inspect.go`, `check_plan.go`, `check_tasks.go`, `check_implement.go`, `check_verify.go`, `check_archive.go`
  - `check_helpers.go` — common path resolution, heading checks, pattern matching
  - `checks.go` — type definitions, constants, regex patterns, `CheckResult` methods, `checkConstitutionLanguagePolicy`
  - all methods and types preserved; `go build`, `go test`, `go vet` clean
- **Lifecycle: archive only after `StatusPass`**:
  - `inferLifecycle` now separates `concerns` from `pass` — only `StatusPass` leads to `ReadyFor = "archive"`
  - `CheckArchiveReady` only accepts `StatusPass` (was `StatusPass` or `StatusConcerns`)
  - unused imports cleaned up across all check files

### Fixed

- **Agent command typo**: agents-snippet now spells `speckeep` (with **p**), not `speckeek` (with **k**)
- **Install scripts**: PowerShell detection, bash compatibility, v0.5.1 download URLs

### Documentation

- README.md / README.ru.md: updated prompts, commands, and workflow guidance

## [v0.5.1] - 2026-06-19

### Changed

- **Agent templates (`agents-snippet.md`) — repository map rule strengthened**:
  - added `⚠️ CRITICAL` prefix and explicit `DO NOT use ls/find/glob` prohibition
  - rule is now an unconditional imperative (was conditional "if REPOSITORY_MAP.md exists")
  - rationale included: saves tokens and maintains workflow discipline
  - applied to both EN and RU templates

## [v0.5.0] - 2026-06-01

### Added

- **`context.Context` + Service interfaces across all internal packages**:
  - all I/O-bound functions now accept `context.Context` as the first parameter
  - `Service` interfaces introduced for `config`, `gitutil`, `trace`, `skills`, `specs`, `workflow`, `doctor`, `status` — enables clean test mocking
  - `exec.Command` replaced with `exec.CommandContext` in `gitutil` for cancellation support
  - all top-level Cobra handlers pass `context.Background()`
- **Sentinel errors** across the entire internal package surface:
  - defined `ErrUnsupportedTarget`, `ErrUnsupportedShell`, `ErrSpecNotFound`, `ErrNotInitialized`, `ErrSlugEmpty`, `ErrVerifyMissing`, `ErrInputEmpty`, `ErrSlugInvalid`, `ErrSpecExists`, `ErrGitRefRequired`, `ErrSkillExists`, `ErrUnsupportedSrc`, `ErrManifestVersion`, `ErrCheckoutNotFound`
  - all support `errors.Is` / `errors.As` wrapping
- **New `/speckeep.repo-map` command**:
  - dedicated command definition and prompt template (en/ru)
  - agents-snippet now references repo-map with its own trigger checklist

### Changed

- **Go 1.26 migration** (`go 1.23.5` → `go 1.26`):
  - removed `GOROOT`-probe fallbacks used with older Go install layouts
  - removed deprecated `sort` import aliases
  - removed unused `service.go` shim (replaced by proper Service interfaces)
  - cleaned up stale `go.sum` entries
  - `go build`, `go vet`, `go test ./...` pass cleanly on Go 1.26
  - CI uses `go-version-file: go.mod` — auto-picks Go 1.26
- **Project package decomposed**: `AddAgents`, `RemoveAgents`, `ListAgents`, `CleanupAgents` extracted from `init.go` into `project/agents.go`
- **Agent prompt templates refined** (all phases):
  - handoff, hotfix, scope, inspect, and verify prompts tightened
  - hotfix now requires a short summary block (`Slug`, `Status`, `Artifacts`, `Blockers`)
  - implement prompt streamlined (removed redundant per-task bullet)
  - agents-snippet end block clarified: `speckeep archive` only after `verify: pass` (was ambiguous "when done")

### Fixed

- Agents no longer suggest `speckeep archive` immediately after implement — the end block rule now explicitly says `only after verify: pass`

### Documentation

- README, README.ru: Go 1.26+ requirement noted in Development section
- docs/en/index.md, docs/ru/index.md: Go 1.26+ noted

## [v0.4.0] - 2026-05-14

### Added

- New `opencode` agent target support:
  - generates SpecKeep workflow wrappers under `.opencode/commands/`
  - included in agent target normalization, refresh/cleanup flows, skill installation paths, CLI help text, and EN/RU docs
  - OpenCode now participates in project-local skills installation under `.opencode/skills/<id>`
- New skill checkout recovery command:
  - `speckeep skills-restore [path]`
  - grouped subcommand: `speckeep skills restore [path]`
  - restores missing git-backed `.speckeep/skills/checkouts/<id>` from skills manifest metadata (`location` + pinned `ref`)

### Changed

- Lean feature artifact layout is now the default for generated guidance and readiness checks:
  - canonical active artifacts now center on `spec.md`, optional `inspect.md`, `plan.md`, `tasks.md`, `data-model.md`, `contracts/`, and `verify.md`
  - generated prompts no longer require legacy digest artifacts such as `summary.md`, `spec.digest.md`, or `plan.digest.md`
  - `tasks.md` now carries an explicit `Implementation Context` section as the main operational bridge for `implement` and `verify`
  - `refresh` and new workspaces treat old summary/digest files as legacy optional artifacts rather than canonical defaults
- Legacy digest artifacts are now effectively retired from the default workflow:
  - `summary.md`, `spec.digest.md`, and `plan.digest.md` are no longer part of the canonical generated artifact set
  - readiness checks and generated guidance no longer depend on `*.digest` files
  - existing repositories may keep old digest files temporarily, but SpecKeep no longer treats them as required operational inputs
- Skills lifecycle is now more self-healing and explicit:
  - `install-skills` auto-rehydrates missing git-backed checkouts from `.speckeep/skills/manifest.yaml` before installing into agent folders
  - `add-skill`, `sync-skills`, and `refresh` now maintain a managed root `.gitignore` block for `.speckeep/skills/checkouts/`
  - README and CLI docs now document checkout caching, restore flow, and managed `.gitignore` behavior
- Traceability guidance is now stricter and more language-aware across constitutions, prompts, workflow docs, CLI docs, and examples:
  - namespaced `@sk-task <slug>#<TASK_ID>` / `@sk-test <slug>#<TASK_ID>` guidance is now reinforced throughout the docs and generated guidance
  - placement rules now explicitly forbid file-level/package-level markers and include language-specific examples (Go, Python, JS/TS, Java, C#/.NET, C/C++, Shell, SQL)
  - if multiple tests verify the same task, trace markers are now required on each such test/case
- `opencode` and `windsurf` wrappers now emphasize trace-marker placement more strongly for implement flows

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

[0.1.0]: https://github.com/bzdvdn/speckeep/releases/tag/v0.1.0
[0.2.0]: https://github.com/bzdvdn/speckeep/releases/tag/v0.2.0
[0.3.0]: https://github.com/bzdvdn/speckeep/releases/tag/v0.3.0
[0.3.1]: https://github.com/bzdvdn/speckeep/releases/tag/v0.3.1
[0.4.0]: https://github.com/bzdvdn/speckeep/releases/tag/v0.4.0
[0.5.0]: https://github.com/bzdvdn/speckeep/releases/tag/v0.5.0
[0.5.1]: https://github.com/bzdvdn/speckeep/releases/tag/v0.5.1
[0.6.0]: https://github.com/bzdvdn/speckeep/releases/tag/v0.6.0
[0.7.0]: https://github.com/bzdvdn/speckeep/releases/tag/v0.7.0
[0.7.1]: https://github.com/bzdvdn/speckeep/releases/tag/v0.7.1
[unreleased]: https://github.com/bzdvdn/speckeep/compare/v0.7.1...HEAD
