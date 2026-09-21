# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versions follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Release gate** (`scripts/release-check.sh`, wired into CI): a single entry point that runs static analysis (`gofmt`, `go vet`, `go test [-race]`), the agent-artifact golden matrix + upgrade-from-legacy Go tests, a `GOOS/GOARCH` build matrix, a black-box end-to-end smoke against the built binary (`speckeep demo --agents all` → per-target artifacts, OpenCode skills-only, `check-ready` no-op for auxiliary commands, `refresh` idempotency, `doctor`), and packaging syntax smoke. `--quick` skips `-race` and the cross-builds; CI uses `--quick` on PRs and the full gate on `main`/`master`. Documented in `CONTRIBUTING.md`.
- **`TestAgentArtifactMatrix`** (`src/internal/agents/matrix_test.go`): golden path-set + invariant test over all 19 targets × {en,ru} × {sh,powershell} — exact generated file set, no legacy paths, OpenCode skills-only, and readiness reminders only on gated phases.
- **`TestUpgradeFromLegacyLayout`** (`src/internal/cli/upgrade_test.go`): emulates a pre-skills-first workspace (per-command wrappers, nested `sdd/phases/`, OpenCode command files), asserts `doctor` flags every stale artifact, `refresh` heals it, OpenCode ends skills-only, user data survives, and `doctor` is clean afterwards.

### Changed

- **Unified slash-command form on `/spk-<phase>` everywhere**: prompts, `agents-snippet.md`, `AGENTS.md` output, CLI help/hints (`init`, `check`, `status`, `dashboard`, `archive`, `import`, `demo`, `explore`), the importer, and `docs/{en,ru}` still used the pre-skills-first `/spk.<phase>` dot form, which no longer matches any generated skill or command. All user- and agent-facing references now use `/spk-<phase>`. `doctor` flags an `AGENTS.md` that still carries the old `/spk.*` form and tells you to `speckeep refresh .`.
- **OpenCode is now skills-only**: the generated `.opencode/commands/spk-<phase>.md` set was removed as a redundant second entry point. OpenCode's `/` palette is built from `.opencode/commands/*` and explicitly skips commands whose `source === "skill"`, so the command files duplicated the skill pack without adding reachability; phase skills stay available through OpenCode's `/skills` picker. `targetFlatCommandDirs` no longer lists `opencode`, so `doctor`/`PathsForTarget`/`refresh` stop expecting or generating those files. `agents.LegacyOpenCodeCommandPaths` lets `doctor` flag and `refresh`/`cleanup-agents` remove leftovers from earlier versions; `docs/{en,ru}/agents.md` updated.

### Fixed

- **Readiness reminder no longer advertised for phases without a readiness check**: `check-ready.sh <phase>` has a real gate only for the core chain (`constitution`, `spec`, `propose`, `inspect`, `plan`, `tasks`, `implement`, `verify`, `converge`). Auxiliary phases (`handoff`, `challenge`, `scope`, `glossary`, `recap`, `hotfix`, `repo-map`, `rollback`) were still told to run it, which failed with `unknown phase` / `unknown flag: --root`. Generated phase skills and flat/Gemini commands now include the readiness line only when `agents.hasReadyCheck` reports a real gate. Removed the dangling `check-hotfix-ready` reference from the `hotfix` command definition.
- **`check-ready` is a no-op for non-gated commands**: `check-ready.*` now short-circuits with `OK: no readiness gate for <phase>` (exit 0) for auxiliary commands instead of forwarding an unknown `check-<phase>-ready` to the CLI, which failed with `unknown flag: --root`. The managed `AGENTS.md` block and the `sdd` overview skill now list auxiliary commands (`repo-map`, `glossary`, `challenge`, `handoff`, `recap`, `scope`, `rollback`, `hotfix`) separately as "not phases, no readiness gate" instead of implying every `/spk-*` is a gated phase.

## [v1.0.0] - 2026-09-11

### Added

- **Recipe: Introducing speckeep Into an Existing Codebase** (`docs/{en,ru}/examples.md`): a full brownfield-adoption walkthrough (real `init`/`doctor` transcript, a constitution written from the codebase's actual current state, a spec narrowly scoped to one existing endpoint, an explicit "what not to do" list, and a before/after table). Cross-linked from the existing "Existing Project" quick pattern.
- **Examples & Recipes expanded** (`docs/{en,ru}/examples.md`): added 15 problem-first recipes covering every secondary command — the alignment pass, `/spk.propose` fast lane, `/spk.glossary`, `/spk.challenge`, `/spk.scope`, `/spk.converge`, `/spk.hotfix`, `/spk.handoff`, `/spk.recap`, `/spk.rollback`, `/spk.repo-map`, `speckeep import` (openspec/speckit), `speckeep guard` in CI, `self check`/`self upgrade`, and the express lane — each with a concrete before/after transcript, mirrored EN/RU. Docs index labels now describe the recipe coverage.
- **npm launcher package** (`contrib/packaging/npm/`): `npx speckeep` / `npm install -g speckeep` now work without any rewrite — the package ships no compiled code, just a `postinstall` script that downloads the release archive matching the package version + host OS/arch from GitHub Releases, verifies its sha256 against `sha256sum.txt`, and a thin `bin/speckeep.js` launcher that execs the resulting native binary with forwarded argv/stdio/exit-code. Closes the "try without installing anything special" gap versus `npx`/`uvx`-based competitors while keeping the Go binary as the single source of truth.
- **`npm-publish` CI job** (`.github/workflows/manual-release.yml`): after a manual release creates the GitHub release + assets, this job syncs `contrib/packaging/npm/package.json`'s version to the released tag and runs `npm publish --access public` — gated on the `NPM_TOKEN` repo secret being set (logs a no-op skip notice otherwise, never fails the release).
- **`/spk.glossary`**: new optional workflow command + prompt (EN/RU) that creates/updates a project-level `.speckeep/glossary.md` — a small shared domain-vocabulary table (Term / Definition / Aliases / Notes). `spec`/`plan`/`tasks` prompts now read it once per session (if present) and reuse its terms instead of introducing synonyms; it is never generated implicitly and carries no mandatory gate.
- **Alignment pass in `/spk.spec`**: before drafting `AC-*`, the prompt now decides whether the request is precise enough to draft directly or needs one batched round of 1–3 clarifying questions (actor/consumer, observable success, split-feature risk, scope boundary). Capped at one round; unresolved points fall through to `Open Questions` instead of guessed `AC-*`. Added a matching self-check item (EN/RU).
- **`speckeep self check` / `speckeep self upgrade`**: first-class binary management. `self check` reports installed vs latest GitHub release (`--json`); `self upgrade` downloads the latest release archive, verifies its sha256 checksum against the published `sha256sum.txt`, and replaces the running binary in place (falls back to a manual hint on Windows or when the install directory is not writable).
- **GitHub Action `speckeep`** (`.github/actions/speckeep/action.yml`): installs speckeep, detects changed `specs/active/<slug>/` features against the PR base, and runs `speckeep guard` per changed slug, failing the PR when a touched feature is not closeable. Ready-to-copy workflow in `contrib/ci/speckeep-guard.yml`.
- **Package-manager packaging** under `contrib/packaging/` (Homebrew formula + Scoop manifest) and README install/update docs for curl, PowerShell, brew, scoop, `go install`, and `self upgrade`.
- **`/spk.propose` one-shot lane**: new optional workflow command + prompt (EN/RU). Turns an idea into `spec.md` + `tasks.md` in a single pass (plan.md only when non-trivial — express lane by default) and goes straight to `implement`. Falls back to `/spk.spec` on ambiguity. Backed by deterministic `CheckProposeReady` wired into `check-ready.propose` for all agent adapters.
- **`speckeep import <openspec|speckit> [path]`**: migrates feature packages from OpenSpec (`openspec/changes/<slug>/` → rebuilds `spec.md` RQ-/AC-\* from `### Requirement:`/`#### Scenario:` blocks, `plan.md` from `design.md`, best-effort `tasks.md`) and Spec Kit (`specs/<slug>/` → copies spec/plan/tasks). Never overwrites existing speckeep features; reports skipped packages; imported `tasks.md` notes that a `/spk.tasks` regeneration adds `Touches:`/Surface Map/Acceptance Coverage.
- **`express` lane for small features**: `plan.md` and `data-model.md` are now optional-on-demand. Tiny/low-risk changes close from `spec.md` + `tasks.md` — the lifecycle lets such features jump straight from tasks to implement/archive, and `check` prints an express hint when `plan.md` is absent.
- **`/spk.converge` fast closing loop**: new optional workflow command + prompt (EN/RU). `speckeep converge <slug>` cheaply re-checks tasks, `Proof:` entries, touched surfaces, and `AC-*` coverage; reports gaps (exit 1) and the agent appends `## Converge Follow-ups` tasks until `converged`. Lighter than `verify` — no report file. Wired into `check-ready.converge` for all agent adapters.
- **`speckeep guard` CI gate**: `speckeep guard [path] [--slug <slug>] [--json]` exits 0 only when every active feature is archive-ready; fails on open tasks, missing Proof, blocked inspect/verify, or branch mismatch.
- **`speckeep archive --compact`**: lean archive that stores only `summary.md` + `snapshot.sha` (git branch/commit/file list) instead of copying every artifact; restore re-creates files via `git show`. Requires a git repo with committed artifacts.
- **Optional `data-model.md` and `plan.md` checks**: `check-tasks-ready`, `check-implement-ready`, and the new converge check treat missing `plan.md`/`data-model.md` as warnings (express mode) instead of errors.
- **Skills-first agent support (20 targets)**: `speckeep init --agents <target>` now lays a composite `sdd` skill pack (root `SKILL.md` + one self-contained per-phase skill under `phases/`) into each target's standard skills directory (`.claude/skills/`, `.opencode/skills/`, `.codex/skills/`, `.gemini/skills/`, … plus `aider`'s `CONVENTIONS.md` pointer). All targets share one renderer (`agents/skill_pack.go`); per-command wrapper generation was removed.
- **Skills-management subsystem removed**: `add-skill`, `list-skills`, `remove-skill`, `install-skills`, `skills-restore`, `sync-skills` (and the grouped `speckeep skills`), the `.speckeep/skills/manifest.yaml` flow, and the git/local skill installer are gone. SpecKeep now ships only its own workflow as skills.

### Removed

- **`continue` (Continue.dev) dropped as a supported agent target**, confirmed by the maintainer after multiple secondary sources reported the product was acquired by Cursor and discontinued around mid-2026 (GitHub repo made read-only). Down from 20 to 19 supported targets. Handled non-destructively: `agents.NormalizeTargets` now silently drops `continue` (via a new `deprecatedTargets` set) instead of erroring, so a project whose `speckeep.yaml` still lists it doesn't hard-fail — `refresh` self-heals the saved config by dropping it and removes any stale `.continue/skills/*` output (new `agents.LegacyContinueSkillPaths`, flagged by `doctor`, cleaned by `refresh`/`cleanup-agents`). An explicit `speckeep init/add-agent --agents continue` now prints `skipped "continue": no longer a supported agent target` instead of silently doing nothing. Docs (`docs/{en,ru}/agents.md`, both READMEs) updated to 19 targets with an explicit note on the removal.

### Fixed

- **Remaining Tier 2 targets individually verified against official docs; Amazon Q, Gemini, Cline given real invocable commands**: extended the same research pass to `kilocode`, `roocode`, `aider`, `amazonq`, `gemini`, `jules`, `cline`, `continue`, `devin`, `goose`, `refact`, `codiumate`, `qwen-code`. Confirmed fixes shipped:
  - **`amazonq`**: real dotfile root is `.amazonq/`, not `.q/` (the old path had no confirmed reader at all — the only official "Skills" doc under that name belongs to an unrelated AWS "Agent Toolkit" plugin). `targetSkillDirs["amazonq"]` corrected; new `.amazonq/prompts/spk-<phase>.md` generated (Amazon Q's real invocable-prompt mechanism, invoked with `@spk-<phase>`, not a slash). New `agents.LegacyAmazonQSkillPaths` lets `doctor`/`refresh`/`cleanup-agents` flag and remove the old `.q/skills/` output.
  - **`gemini`**: Skills there are model-invoked only; added `.gemini/commands/spk-<phase>.toml` (Gemini's real custom-command format — TOML, not markdown) via new `renderGeminiCommand`.
  - **`cline`**: `.clinerules/` is a context directory, not commands; added `.clinerules/workflows/spk-<phase>.md` (Cline's real Workflows slash mechanism) — `targetFlatCommandDirs` generalized to a struct carrying each target's invoke prefix (`/` vs `@`) to support this and the amazonq fix uniformly.
  - **`qwen-code`**: confirmed correct as-is — Skills are genuinely slash-invocable there, matching speckeep's existing `.qwen/skills/` output exactly.
  - **`aider`**: confirmed correct as-is — no Skills/command loader exists; the `.aider/CONVENTIONS.md` pointer remains the right approach.
    Documented, not code-fixed (either no safe fix exists or evidence was too thin to act on): `kilocode` (community docs suggest a `.kilo/` vs `.kilocode/` root rename and a real `.kilo/commands/` mechanism, but this is single-sourced and was **not** applied to avoid breaking existing installs on unconfirmed grounds); `roocode` (Skills are model-invoked only and Roo Code has no per-command flat-file mechanism at all — nothing to target); `devin`/`goose` (Skills work via `@skills:name` / `/skills <name>` meta-command respectively, not direct `/name`; speckeep's existing paths are confirmed-valid fallback paths, just not each tool's canonical `.agents/skills/`); `jules` (async agent, no Skills/command mechanism applies — only reads `AGENTS.md`); `refact` (no evidence of any Skills or project-local command convention); `codiumate` (renamed to Qodo Gen; a compatible system exists but its exact path/invocability isn't confirmed by official sources). `docs/{en,ru}/agents.md`'s verification table now covers all 20 targets.
  - **`continue`**: flagged, not acted on — multiple secondary sources suggest Continue.dev was acquired by Cursor and discontinued around mid-2026 (GitHub repo made read-only), unconfirmed via a primary announcement. This is a maintainer decision (keep generating for a possibly-discontinued tool, or downgrade/remove the target), not something to silently change.
- **Windsurf and OpenCode get real `/spk-<phase>` slash commands again**: research against each tool's official docs found their Skills mechanism is _not_ slash-invocable at all — Windsurf Skills load only automatically or via `@mention`, OpenCode Skills only via the agent calling a `skill()` tool. Their actual slash-command mechanisms are separate, flat-file directories (`.windsurf/workflows/*.md`, `.opencode/commands/*.md`) that the recent legacy-wrapper cleanup was deleting as "superseded," leaving those two targets with no slash-invocable phases at all. `speckeep init`/`refresh` now additionally generate `.windsurf/workflows/spk-<phase>.md` and `.opencode/commands/spk-<phase>.md` (new `targetFlatCommandDirs`, `renderFlatCommand`) alongside the skill pack for exactly these two targets. `docs/{en,ru}/agents.md` gained a per-target "what's actually verified" table (Claude/Cursor confirmed slash-invocable; Codex's `.codex/skills/` path is community-documented, not confirmed canonical — `.agents/skills/` is; Trae's Skills auto-discovery of drop-in files is unverified per an open upstream issue) instead of asserting uniform slash-invocability across all 20 targets.
- **Phase skills are directly slash-invocable again (`/spk-spec`, `/spk-plan`, ...)**: the composite `sdd` skill pack (one root `SKILL.md` + nested `sdd/phases/*.md` resource files) meant skill loaders like Claude Code — which only ever slash-invoke a top-level `<dir>/SKILL.md`, never a nested file — could only reach the whole workflow via `/sdd`, never an individual phase directly; `/spk.spec` etc. stopped being real commands the moment skills-first replaced the old per-command wrapper files, with no equivalent restored. Each phase is now its own independent top-level skill (`spk-<phase>/SKILL.md`, e.g. `spk-spec`, `spk-plan`, `spk-implement`), matching how Claude Code (and mattpocock/skills' own convention: one top-level directory per user-invoked skill) actually resolves slash commands. The `sdd` overview skill is kept as a lightweight entry point for when the phase isn't obvious from the request. New `agents.LegacySkillPhasePaths` lets `doctor` flag and `refresh`/`cleanup-agents` remove the superseded nested layout.
- **`doctor`/`refresh`/`cleanup-agents` now clean up pre-skills-first `/spk.*` per-command wrapper files** (e.g. `.claude/commands/spk.spec.md`, `.codex/prompts/spk.spec.md`, `.opencode/commands/spk.spec.md`, `.cursor/rules/spk-spec.mdc`), not just the older pre-rename `/speckeep.*` ones. `LegacyPrefixPaths` only ever matched the `speckeep.` prefix; anyone upgrading a workspace generated before the skills-first migration landed (one file per command per target) had those files silently ignored by every cleanup path — `doctor` never flagged them and `refresh`/`cleanup-agents` never removed them, even though the composite `sdd` skill pack fully supersedes them. New `agents.LegacyCommandWrapperPaths` covers the `spk.` prefix era; `doctor` reports it with an accurate "superseded by the sdd skill pack" message instead of the misleading "rename to /spk.\*" text that only applies to the older prefix.

### Changed

- **Agent-target maintenance tiers documented** (`docs/{en,ru}/agents.md`): all 20 targets stay generated the same way (no code/behavior change), but docs now split them into Tier 1 (`claude`, `cursor`, `codex`, `copilot`, `windsurf`, `opencode` — maintained-first) and Tier 2 (everything else — best-effort/community, fixed on report). README links to the tier section next to the adapter list.
- **Skill packs are self-contained**: generated per-phase skill files (`<target>/skills/sdd/phases/<phase>.md`) now inline the canonical prompt body directly instead of pointing at `.speckeep/templates/prompts/<phase>.md`. A skill-driven agent gets full phase instructions from one read; the canonical prompt file remains the single authored source and is still referenced for traceability/sync. New `templates.PromptContent(lang, name)` embeds the prompt body at render time.
- **Lifecycle (`inferLifecycle`)**: a feature with `tasks.md` present no longer requires `plan.md` to reach `implement`/`archive` (express mode); the default chain still suggests `plan` first when both are absent.
- **Agent distribution becomes skills-first**: per-command wrappers for all target tools are replaced by one composable `sdd` skill pack per target; `init`/`add-agent`/`refresh` write skills only.
- **Install docs (EN/RU)**: README defaults to `main` script install, adds brew/scoop/`go install` rows and `self check`/`self upgrade`; a new CI section documents the GitHub Action.
- **Docs (EN/RU)**: README, CLI reference, workflow, architecture, agents, and MVP cover propose, import, express lane, converge, guard, compact archive, skills-first adapters, and the removed skills-management commands. `plan.md` template and prompts no longer require a `data-model.md` no-change stub.
- **Prompts (EN/RU)**: `plan`/`tasks` prompts document the express lane and one-shot propose; `agents-snippet.md` command list includes `/spk.propose` and `/spk.converge`.

### Testing

- Added `go test -race ./...` as the local verification baseline for the release; the full suite runs race-clean. New coverage for: `self check`/`self upgrade` (checksum mismatch keeps the old binary), propose readiness, express-mode lifecycle, converge loop, guard (including missing-Proof detection), compact archive roundtrip, and openspec/speckit import.

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
[0.8.0]: https://github.com/bzdvdn/speckeep/releases/tag/v0.8.0
[0.8.1]: https://github.com/bzdvdn/speckeep/releases/tag/v0.8.1
[1.0.0]: https://github.com/bzdvdn/speckeep/releases/tag/v1.0.0
[unreleased]: https://github.com/bzdvdn/speckeep/compare/v1.0.0...HEAD
