# Roadmap

This roadmap focuses on the next practical iterations for SpecKeep rather than a long speculative backlog.

## Direction

SpecKeep should continue to position itself between heavier spec-driven systems and looser change-driven systems:

- stricter than OpenSpec
- lighter than spec-kit
- optimized for agent-first workflows on real codebases

## First Release Focus

The immediate goal for SpecKeep is not to match heavier SDD systems in phase count, artifact count, or automation depth.

The immediate goal is to:

- ship a lightweight first release
- test the workflow in real codebases and real agent sessions
- validate that a strict-by-structure approach works without large default context

Before those field tests, SpecKeep should prefer:

- narrow default context over broad repository reads
- cheap checks and readiness scripts over heavy orchestration
- a minimal required artifact set over growing every feature package
- stronger traceability and consistency without increasing prompt mass

Lightweight guardrails for the next iterations:

- each phase should define `always load`, `load if needed`, and `never load by default` inputs
- `implement` should remain task-scoped by default and open deeper artifacts only when the active task requires them
- `verify` should stay cheap-by-default and only deepen into code or wider review when explicitly requested
- prerequisite checks should move into helper scripts and readiness checks wherever practical instead of repeated prompt-time reasoning
- traceability should improve through stable IDs and explicit references instead of new shared summary artifacts
- `archive` should remain a compact historical record rather than becoming a new mutable working-memory layer

Before those field tests, SpecKeep should avoid rushing into:

- new mandatory phases
- wider default inspect/verify context
- automation that makes the workflow heavier before its value is proven

## Iteration 1

### Primary goal

Strengthen `inspect` as the central quality layer.

Status: the mandatory persisted inspect report, shared report parsing, and core semantic validation are now in place.

Release filter: strengthen `inspect` only in ways that keep it cheap in context terms and prevent it from becoming a mandatory heavy review engine.

### Planned work

- deepen semantic checks for `constitution <-> spec`
- deepen semantic checks for `spec <-> plan`
- deepen semantic checks for `plan <-> tasks`
- improve how inspect findings are presented in CLI and docs
- keep strengthening acceptance-to-task traceability checks

### Anti-Bloat Notes

Safe direction:

- stronger structural checks
- clearer verdict semantics
- better traceability through stable acceptance IDs
- cheap `spec <-> plan` consistency checks based on `spec.md` and `plan.md` only

Use caution with:

- reading implementation code by default during inspect
- turning inspect into a broad review engine
- pulling `data-model`, `contracts`, and code into every inspect run by default

### Why this matters

If `inspect` is strong, every downstream phase gets better with less wasted implementation effort.

## Iteration 2

### Primary goal

Add a lightweight post-implementation verification layer.

Status: lightweight contract, prompt, readiness script, report template, evidence-oriented report shape, and token-safe validator checks are now in place. The remaining work is to deepen checks and improve presentation without expanding default context.

Release filter: `verify` should remain a lightweight safety layer, not a new heavy review or QA engine.

### Planned work

- deepen checks for whether completed tasks match implementation state
- deepen checks for whether implementation still matches spec and plan intent
- improve verify report evidence quality and archive-readiness consistency
- improve verify findings and status presentation in CLI
- ensure archived feature state and task state remain consistent where verification depends on them

### Anti-Bloat Notes

Safe direction:

- task-state verification helpers
- archive/task consistency checks
- evidence-oriented verify reports without broadening default reads
  Status: coarse helper-based sync checks and basic report semantics are now in place for `verify`.

Use caution with:

- reading code by default during verify
- turning verify into a heavy review or QA engine

### Why this matters

This closes the gap between "tasks were executed" and "the feature is actually aligned with its intended design".

## Iteration 3

### Primary goal

Strengthen brownfield ergonomics and automation outputs.

Release filter: add automation outputs only where they reuse existing checks and do not pull in new mandatory context.

### Planned work

- improve archive summaries and archive linkage
- keep completed-archive checks cheap by reusing task-state verification
- add machine-readable outputs such as `doctor --json`
  Status: implemented for `doctor`; extend this pattern only when outputs stay cheap and reuse existing checks.
- improve config-aware helper behavior for scripts and future tooling
- continue polishing multilingual consistency in docs and prompts

### Anti-Bloat Notes

Safe direction:

- machine-readable outputs for existing checks
- better archive indexing and summaries
- config-aware helpers that reduce repeated reasoning

Use caution with:

- archive flows that require reading broad repository history
- new automation outputs that introduce mandatory new artifacts
- brownfield helpers that silently widen the default context

### Why this matters

This makes SpecKeep easier to automate, easier to operate at scale, and stronger for long-lived existing codebases.

## Always-On Quality Work

Alongside feature work, SpecKeep should keep improving:

- documentation consistency
- unit test coverage
- CLI ergonomics
- prompt clarity and token discipline
- brownfield workflow quality

## Not Planned Right Now

SpecKeep should avoid these unless there is a very strong reason:

- a heavy orchestration engine
- mandatory checkpoint systems
- approval-gate bureaucracy
- large default prompt contexts
- required artifact sprawl for every feature
- trying to become a full process operating system before the lightweight core is proven in practice
