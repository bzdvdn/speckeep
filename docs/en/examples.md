# Examples

This page shows realistic end-to-end SpecKeep scenarios for one feature package.

## Quick Usage Patterns

### New Project

When starting a greenfield project, SpecKeep works best as a minimal project-context layer from day one.

Example:

```bash
speckeep init my-project --lang en --shell sh --agents codex
cd my-project
speckeep doctor .
```

What to do next:

- establish the `constitution` for project rules
- describe the first feature through `spec`
- prepare `plan` and `tasks`
- use `implement` only from the current task list

Why this helps:

- humans and agents start from the same rules
- project context stays explicit and editable from the beginning
- the workflow stays lightweight because SpecKeep does not require a heavy process engine

### Existing Project

For a brownfield codebase, SpecKeep should be adopted incrementally instead of trying to document the whole repository at once.

Example:

```bash
cd existing-project
speckeep init . --lang en --shell sh --agents codex
speckeep doctor .
```

Recommended starting point:

- establish the `constitution` around the project's current reality
- pick one active feature or change request
- create a spec only for that scope
- move to plan, tasks, and implement only within that feature package

What not to do:

- do not try to spec the whole project at once
- do not pull broad repository context unless the active feature really needs it

Why this helps:

- SpecKeep adds a lightweight layer of discipline on top of an existing codebase
- adoption happens one feature at a time
- this keeps token usage down and avoids process bloat

For a full worked example — real constitution text, a narrowly-scoped spec
on one existing endpoint, and a concrete before/after — see
[Recipe: Introducing speckeep Into an Existing Codebase](#recipe-introducing-speckeep-into-an-existing-codebase).

### Prompt File Input

When `/spk-spec` starts from a local prompt file, prefer explicit metadata instead of relying on a generic filename such as `spec_prompt.md`.

Example prompt file:

```text
name: Add dark mode
slug: add-dark-mode

Add a user-selectable dark theme for the dashboard and settings pages.
```

This lets SpecKeep:

- derive a safe spec path such as `specs/active/add-dark-mode/spec.md`
- create or switch to `feature/add-dark-mode`
- avoid ambiguous slugs from generic filenames

### Staged Input Via `--name`

When the feature name is already clear but the detailed description is easier to send in the next message, `/spk-spec` can start in staged mode.

Example:

```text
/spk-spec --name "Dependency Dashboard"
```

Next message:

```text
Build a dashboard for monitoring microservice dependencies with a dark theme, filters, a dependency graph, summary cards, and auto-refresh.
```

This allows SpecKeep to:

- lock in the canonical feature name up front
- derive a safe slug such as `dependency-dashboard`
- preserve the spec request context across messages

If you need an explicit slug:

```text
/spk-spec --name "Dependency Dashboard" --slug frontend-layout-rework
```

If you need a repository-specific branch override:

```text
/spk-spec --name "Dependency Dashboard" --slug frontend-layout-rework --branch FEAT-142
```

## 1. Create a Constitution for a Brownfield Project

User request:

```text
/spk-constitution Python project, DDD style, split into API and workers, Kafka for asynchronous integration, ClickHouse as the analytical sink.
```

Expected agent behavior:

- read the constitution prompt in `.speckeep/templates/prompts/constitution.md`
- inspect only the minimum repository evidence needed
- create or patch `CONSTITUTION.md`
- run `check-constitution.sh` when appropriate

Expected outcome:

- architecture rules are formalized
- development workflow rules become explicit
- the constitution becomes authoritative for later phases

## 2. Create a Spec

User request:

```text
/spk-spec Add partner-specific ingestion scheduling with retry policy overrides.
```

Expected agent behavior:

- read constitution first
- create `specs/active/partner-scheduling/spec.md`
- write acceptance criteria using canonical `Given / When / Then`
- keep surrounding text in the configured documentation language

Example acceptance criterion:

```md
### Acceptance Criterion 1

- ID: AC-001
- **Given** a partner with a custom retry policy
- **When** the ingestion schedule is evaluated
- **Then** the worker uses the partner-specific retry window instead of the default policy
```

Example with an explicit branch override:

```text
/spk-spec Add partner-specific ingestion scheduling with retry policy overrides --branch NRD-11
```

In that case, the spec slug can still stay `partner-scheduling` while the working branch follows the repository's branch convention, for example `NRD-11`.

## 3. Inspect the Spec

Use this step when the feature is ambiguous, high-risk, or you want a formal quality gate.  
If the spec is already clear and low-risk, you may proceed directly to `/spk-plan <slug>`.

User request:

```text
/spk-inspect partner-scheduling
```

Expected agent behavior:

- read constitution and `specs/active/partner-scheduling/spec.md`
- keep the default inspect scope cheap: prefer `CONSTITUTION.md` and `spec.md`, then pull `plan.md` or `tasks.md` only when they exist and materially affect the finding
- check completeness, constitutional consistency, and scenario quality
- create a focused inspection report
- use `.speckeep/scripts/inspect-spec.sh` or `.speckeep/scripts/inspect-spec.ps1` as a cheap first-pass helper when structural spec or coverage issues need quick confirmation
- persist the inspect report at `specs/active/partner-scheduling/inspect.md`
- use `.speckeep/templates/inspect-report.md` as the canonical report template

Typical findings:

- missing failure-path scenario
- unclear acceptance coverage for manual retry overrides
- open question about scheduler ownership

## 4. Create a Plan Package

User request:

```text
/spk-plan partner-scheduling
```

Expected agent behavior:

- read constitution and the spec
- if `specs/active/partner-scheduling/inspect.md` exists, validate that status is non-blocking (`pass` or `concerns`)
- create `specs/active/partner-scheduling/plan.md`
- create `specs/active/partner-scheduling/data-model.md`
- create `specs/active/partner-scheduling/contracts/`
- create `research.md` only if uncertainty is real

Typical outputs:

- plan for scheduler integration points
- data model for partner overrides and retry windows
- event or API contracts for configuration updates

## 5. Create Tasks

User request:

```text
/spk-tasks partner-scheduling
```

Expected agent behavior:

- use `plan.md` as the decomposition entrypoint
- pull in spec, contracts, or data model only when needed
- produce `specs/active/partner-scheduling/tasks.md`
- include acceptance-to-task coverage

Example task structure:

```md
## Phase 1: Data Model

- [ ] T1.1 Add partner scheduling override model — override fields are persisted
- [ ] T1.2 Persist retry window fields — retry windows are available to scheduling logic

## Acceptance Coverage

- AC-001 -> T1.1, T1.2
```

## 6. Implement the Feature

User request:

```text
/spk-implement partner-scheduling
```

Expected agent behavior:

- read `tasks.md` and use it as the execution manifest
- perform **In-place Decomposition** if a task is too complex, adding indented sub-tasks (e.g., `T1.1.1`)
- record a `Proof:` line in `tasks.md` under each completed task
- mark completed tasks in `tasks.md`
- stay within the `Touches:` list defined for each task

Example `tasks.md` proof entries after implementing `T1.1`:

```md
## Phase 1: Data Model

- [x] T1.1 Add partner scheduling override model — override fields are persisted
      Proof: code src/models/partner_schedule.go PartnerSchedule
      Proof: test src/tests/partner_schedule_test.go TestSavePartnerSchedule
- [ ] T1.2 Persist retry window fields — retry windows are available to scheduling logic

## Acceptance Coverage

- AC-001 -> T1.1, T1.2
```

## 7. Verify the Implementation

Verify is an optional on-demand audit: it is always available but skipped by default in the normal workflow.

User request:

```text
/spk-verify partner-scheduling
```

Expected agent behavior:

- use `.speckeep/scripts/trace.sh partner-scheduling` to collect implementation evidence
- read the `Proof:` lines from `tasks.md` and confirm the referenced files exist
- confirm that implementation matches task descriptions and acceptance criteria
- provide a clear verdict (`pass`, `concerns`, or `blocked`)
- include concrete evidence in the `## Checks` section

The trace helper reads `Proof:` entries under each completed task, for example:

```md
## Phase 1: Data Model

- [x] T1.1 Add partner scheduling override model — override fields are persisted
      Proof: code src/models/partner_schedule.go PartnerSchedule
      Proof: test src/tests/partner_schedule_test.go TestSavePartnerSchedule
```

A `[x]` task without at least one `Proof:` entry is not considered done.

Expected agent behavior:

- start from `tasks.md`
- load spec, plan, data model, or contracts only for the active task
- implement unfinished tasks in order
- report phase progress as it moves through the selected work
- update `tasks.md`

This phase should avoid broad repository reads unless the active task actually requires them.

Example scoped requests:

```text
/spk-implement partner-scheduling --phase 2
/spk-implement partner-scheduling --tasks T1.1,T2.1
```

Expected scoped behavior:

- keep the default full-run behavior only when no scope flag is provided
- execute only the selected phase or task IDs when scope is explicitly narrowed
- preserve task order from `tasks.md`
- warn if selected work skips unfinished earlier phases or tasks

Typical runtime updates:

- `Starting Phase 1: Data Model`
- `Phase 1 complete: T1.1, T1.2`
- `Next: Phase 2: Scheduler Logic`

## 7. Verify the Feature

User request:

```text
/spk-verify partner-scheduling
```

Expected agent behavior:

- read constitution and tasks first
- confirm that completed tasks match the current implementation state closely enough
- produce a lightweight verification report
- start with `.speckeep/scripts/verify-task-state.sh partner-scheduling` when task-state confirmation is enough
- use `.speckeep/templates/verify-report.md` when the report should be persisted
- default to `specs/active/partner-scheduling/verify.md` when no explicit path is provided

## 8. Archive the Feature

CLI follow-up after `verify: pass`:

```bash
speckeep archive partner-scheduling .
```

Expected CLI behavior:

- for `completed` status, validate verify/task prerequisites and stop with a clear error if open tasks remain
- copy the feature package into `specs/archived/partner-scheduling/<YYYY-MM-DD>/`
- write `summary.md`

Expected archive result:

```text
specs/
  archived/
    partner-scheduling/
      2026-03-28/
        summary.md
        spec.md
        plan.md
        tasks.md
        data-model.md
        contracts/
```

## 9. Agent Maintenance Scenario

A practical maintenance flow for agent targets:

```bash
speckeep add-agent my-project --agents claude --agents cursor
speckeep list-agents my-project
speckeep remove-agent my-project --agents cursor
speckeep cleanup-agents my-project
speckeep doctor my-project
```

Use this when a project changes its preferred agent mix over time.

---

## Recipes

The core lifecycle (`constitution → spec → plan → tasks → implement → archive`)
is covered above. These recipes cover the secondary commands: when each one
earns its cost, a concrete before/after, and what not to expect from it.

### Recipe: Introducing speckeep Into an Existing Codebase

Problem this solves: a brownfield repo already has running code, an
established (if informal) architecture, and a team shipping against it. The
temptation is to "do it properly" and spec the whole thing before touching
anything — that's exactly the process-bloat failure mode speckeep is built
to avoid. The right unit of adoption is **one feature**, not the repository.

Starting point — a real, already-messy repo, not an empty directory:

```bash
cd billing-api   # existing Express + Postgres API, ~40k LOC, 4 engineers,
                 # no prior spec process, feature requests currently live in
                 # Jira tickets and Slack threads
speckeep init . --lang en --shell sh --agents claude
```

```text
initialized .speckeep/ workspace
wrote CONSTITUTION.md (template — edit or run /spk-constitution)
wrote AGENTS.md (managed SpecKeep guidance block)
wrote .claude/skills/sdd/ (8 phase skills)
next: speckeep doctor .
```

```bash
speckeep doctor .
```

```text
ok:      .speckeep/ layout present
ok:      AGENTS.md contains managed SpecKeep block
warning: current branch "main" — feature work should happen on feature/<slug>
ok:      claude agent target: skill pack present
```

Nothing here required reading the codebase — `init` only lays down the
speckeep-managed files. The next step is the one that matters: don't spec
the repository, spec **one pain point** the team already agrees on.

**Step 1 — constitution reflects current reality, not an aspiration.**
Write down what's actually true today, including the parts that aren't
pretty — a constitution that describes an imagined future architecture is
worse than none, because agents will trust it.

```text
/spk-constitution Node.js + Express monolith, PostgreSQL via raw pg queries
(no ORM), REST-only (no GraphQL), synchronous request/response (no queue
yet — that's a known gap, not a target state). Tests use Jest + supertest.
No feature may introduce a new HTTP framework or database without an
explicit constitution update first.
```

Resulting `CONSTITUTION.md` excerpt:

```md
## Architecture
- Node.js + Express, PostgreSQL via raw `pg` queries (no ORM by design).
- REST only; no GraphQL. Synchronous request/response — no message queue
  exists yet (known gap, not a target to silently "fix" inside a feature).

## Non-Negotiable Rules
- No new HTTP framework or database without a constitution update first.
- Tests: Jest + supertest; no new test framework introduced ad hoc.
```

**Step 2 — pick the one feature the team is already arguing about,** not a
survey of the whole API surface:

```text
/spk-spec Support idempotency keys on POST /invoices so retried client
requests don't create duplicate invoices.
```

The agent reads the constitution, greps for the existing `/invoices` route
handler and its test file — nothing else — and drafts `spec.md` scoped to
exactly that endpoint. It does **not** open every route file "to understand
the codebase" first; that's the narrow-read discipline the constitution and
`AGENTS.md` guidance exist to enforce.

```text
Slug: invoice-idempotency-keys
Status: spec
Artifacts: specs/active/invoice-idempotency-keys/spec.md
Blockers: none
Ready for: /spk-plan invoice-idempotency-keys
```

From here the feature runs through the normal chain
(`plan → tasks → implement → archive`, see the walkthrough above) exactly
as it would in a greenfield project — the only difference brownfield
introduces is at the constitution and spec steps, where the agent is
constrained to the codebase's actual shape instead of a blank slate.

**What not to do:**

- Don't ask `/spk-spec` (or any phase) to "document the whole billing-api
  service" — that's a repo audit, not a feature spec, and it will produce a
  huge, unreviewable artifact that nobody keeps up to date.
- Don't let the constitution describe where the architecture is *headed* —
  describe where it *is*. Migrations (e.g. "we're adding a queue") belong in
  a feature's own `plan.md`, not baked into the constitution as if already
  true.
- Don't run `/spk-repo-map` on day one "just to have it" — it earns its
  keep once a feature's `Touches:` surfaces stop being obvious from the
  existing directory names, not before.

**Before / after:**

| | Before speckeep | After speckeep |
| --- | --- | --- |
| Where feature intent lives | Jira description + Slack thread | `specs/active/<slug>/spec.md`, versioned with the code |
| How much of the repo an agent reads | Whatever it guesses it needs | Constitution + the one feature's `Touches:` surfaces |
| What "done" means | "Looks right in review" | Every `[x]` task carries a `Proof:` line `speckeep guard` can check |
| Onboarding a new engineer to a feature | Re-read the PR and ask around | `/spk-handoff <slug>` or `/spk-recap` |

If the first feature already feels too big to spec cleanly (touches several
services, unclear ownership, real design risk), that's a legitimate signal
to run `/spk-inspect` before `/spk-plan` — not a reason to widen the spec to
cover more ground defensively.

### Recipe: The Alignment Pass Asks Before It Guesses

Problem this solves: an agent that drafts acceptance criteria from a vague
request either invents requirements silently or produces a spec nobody
agreed to. `/spk-spec` runs a short **alignment pass** first — it only asks
when the request is genuinely ambiguous, never as a formality.

Vague request:

```text
/spk-spec Add a way to export data.
```

Expected agent behavior — it recognizes the actor and the scope boundary are
both missing, and asks once, in a single batch, instead of guessing:

```text
Before drafting, two things I can't infer from the request:
1. Who exports — an end user from the UI, or an automated job/API caller?
2. Format(s): CSV only, or CSV + JSON/Excel?

I'll assume "last 30 days" is the default range unless you say otherwise.
```

Precise request — same command, zero clarifying questions, straight to the
spec:

```text
/spk-spec Let a dashboard user export the visible table as CSV, respecting
the active filters, from a button in the top-right toolbar.
```

Why this matters: a spec with zero clarify rounds is the *good* outcome, not
a shortcut skipped — the alignment pass exists for the ambiguous 30% of
requests, not to interrogate every one. If you don't answer and move straight
to another command, the agent drafts best-effort and puts the open points
under `## Open Questions` instead of inventing an `AC-*` to fill the gap.

### Recipe: One-Shot Fast Lane With `/spk-propose`

Problem this solves: for a small, low-risk change, running
`spec → plan → tasks` as three separate round-trips is process overhead the
change doesn't need. `/spk-propose` collapses idea → `spec.md` + `tasks.md`
(skipping `plan.md` by default — the **express lane**) in one pass, then
hands off straight to `/spk-implement`.

```text
/spk-propose Add a "copy as JSON" button next to the existing "copy as CSV"
button on the report detail page.
```

Expected outcome:

```text
Slug: copy-as-json
Status: propose
Artifacts: specs/active/copy-as-json/spec.md, specs/active/copy-as-json/tasks.md
Blockers: none
Ready for: /spk-implement copy-as-json
```

If the idea turns out to be bigger than it looked — multiple realistic
implementation options, cross-boundary impact, migration risk — propose
stops and falls back to `/spk-spec` instead of forcing a shallow plan into
`tasks.md`. That's a `Blockers:` line, not a bug: propose is deliberately
narrow so it never race-drafts a feature that actually needed design.

### Recipe: Shared Domain Glossary With `/spk-glossary`

Problem this solves: without a shared vocabulary, `spec.md` calls it a
"workspace", `plan.md` calls it a "project", and the code calls it a
"tenant" — three names for one concept, and every reader pays the
translation tax. `.speckeep/glossary.md` is optional and project-level (not
per-feature); it exists only when a term is worth pinning down.

```text
/spk-glossary Define "workspace" vs "project" — we keep using them
interchangeably and it's starting to cause confusion in specs.
```

Resulting `.speckeep/glossary.md`:

```md
# Glossary

| Term | Definition | Aliases (do not use) | Notes |
| --- | --- | --- | --- |
| `Workspace` | A billing-scoped container that holds one or more projects. | `Account`, `Org` | Created at signup; 1:1 with a Stripe customer. |
| `Project` | A named collection of specs/features inside a workspace. | `Workspace` (rejected) | Formerly called "workspace" in old docs — do not reuse that name for this concept. |
```

Once this file exists, `/spk-spec`, `/spk-plan`, and `/spk-tasks` read it
once per session and reuse its terms — they will not introduce a new
synonym for something it already defines. Nothing else changes: no gate, no
required re-read, no mandatory update on every feature.

### Recipe: Finding Blind Spots With `/spk-challenge`

Problem this solves: `/spk-spec` and `/spk-inspect` are written to converge
on a shippable artifact; `/spk-challenge` is deliberately adversarial — it
looks for the gaps a cooperative pass tends to miss (untestable claims,
silent scope expansion, contradictions).

```text
/spk-challenge partner-scheduling
```

Typical findings:

```text
- AC-002 says "the retry window updates promptly" — "promptly" isn't
  observable. Fix: state a concrete bound (e.g. "within 60s") or drop the
  claim to a non-functional note.
- spec.md doesn't say what happens when a partner has no override —
  default policy is implied but never stated. Fix: add an explicit
  AC or an Assumptions line.
- plan.md DEC-002 introduces a new caching layer not mentioned in spec.md
  Out of Scope — possible hidden scope expansion. Fix: confirm with spec
  or move it out of this feature.
```

`/spk-challenge` only reports findings + minimal fixes — it does not emit a
`pass|concerns|blocked` verdict (that's `/spk-inspect`) and does not replace
a scope inventory (that's `/spk-scope`). Run it when you want a second,
skeptical pair of eyes on a spec or plan before committing to it.

### Recipe: A Quick Boundary Check With `/spk-scope`

Problem this solves: "is X in scope for this feature?" is a question that
otherwise gets answered by re-reading the whole spec. `/spk-scope` gives a
fast in/out inventory without a verdict or a deep review.

```text
/spk-scope partner-scheduling
```

```text
In scope:
- Per-partner retry window override (AC-001)
- Falling back to the default policy when no override exists

Out of scope:
- Editing retry policies through the admin UI (tracked separately)
- Historical backfill of past scheduling decisions

Risks:
- "Partner" isn't defined — could mean the billing entity or the API
  integration key holder; these can differ.

Clarify questions:
- Which "partner" identity does the override key on?
```

Use this before `/spk-plan` when scope creep risk feels high, or any time a
teammate asks "wait, is this feature also going to do X?"

### Recipe: Closing Out a "Mostly Done" Feature With `/spk-converge`

Problem this solves: a feature where most tasks are checked off but a few
`Proof:` entries are thin or missing shouldn't need a full `/spk-verify`
audit — `/spk-converge` is the cheap loop that turns exactly those gaps into
follow-up tasks and re-checks until clean.

```bash
speckeep converge partner-scheduling
```

```text
gaps-found: 2 findings
- T2.3 is [x] but has no Proof: line
- AC-003 has no task mapped in Acceptance Coverage
exit code: 1
```

The agent then appends a `## Converge Follow-ups` section to `tasks.md`:

```md
## Converge Follow-ups

- [ ] T3.1 Add Proof for retry-window persistence — outcome: T2.3 has a
      valid Proof: line; Touches: specs/active/partner-scheduling/tasks.md
- [ ] T3.2 Cover AC-003 with a task — outcome: manual override UI hint is
      implemented and traceable; Touches: src/ui/partner_override.go

## Acceptance Coverage

- AC-003 -> T3.2
```

...implements them, and re-runs `speckeep converge partner-scheduling` until
it reports `converged`. This is lighter than `/spk-verify`: no persisted
report, just tasks/coverage brought back in sync — the right tool when the
feature is 90% there, not when you need a full audit trail.

### Recipe: An Emergency Fix With `/spk-hotfix`

Problem this solves: a production bug doesn't wait for
`spec → plan → tasks`. `/spk-hotfix` is a narrow escape hatch: at most 3
files, no re-planning, no scope expansion — if the fix needs more than that,
it isn't a hotfix anymore.

```text
/spk-hotfix The export endpoint 500s when a report has zero rows — NPE on
an empty result set.
```

```text
Changed files: src/export/handler.go, src/export/handler_test.go
Fixed: guard against a nil/empty result set before building the CSV writer;
added a regression test for the zero-row case.
Verify: `go test ./src/export/...` — TestExportEmptyResultSet passes.

Slug: hotfix-export-empty-result
Status: hotfix
Artifacts: src/export/handler.go, src/export/handler_test.go
Blockers: none
Ready for: speckeep archive hotfix-export-empty-result .
```

If the real fix needs a 4th file or a design change, the agent stops and
says so instead of quietly growing the hotfix — that's the signal to go
through `/spk-spec` instead.

### Recipe: Resuming Cleanly With `/spk-handoff`

Problem this solves: a session ends mid-feature and the next session (yours
tomorrow, or a teammate's, or a different agent) shouldn't have to
re-derive state from scratch by reading every artifact end to end.

```text
/spk-handoff partner-scheduling
```

```text
Slug: partner-scheduling
Phase: implement
What changed: T1.1, T1.2 done (data model + retry window persistence).
T2.1 (scheduler integration) is in progress — override lookup wired, not
yet covering the "no override" fallback path.
Open tasks: T2.1 (partial), T2.2, T3.1
Blockers: none
Next command: /spk-implement partner-scheduling --tasks T2.1
```

Paste that into a fresh session (or hand it to a teammate) and the next
agent has everything it needs without re-reading `spec.md`/`plan.md` end to
end.

### Recipe: A Project-Wide Status Read With `/spk-recap`

Problem this solves: "what's the state of everything?" across several
active features, without opening each one.

```text
/spk-recap
```

```text
| Slug                | Phase     | Status (blockers?)        | Next                         |
| -------------------- | --------- | -------------------------- | ----------------------------- |
| partner-scheduling   | implement | in progress                | /spk-implement --tasks T2.1  |
| copy-as-json         | implement | done, unverified           | speckeep archive copy-as-json . |
| dark-mode            | plan      | blocked: inspect concerns  | /spk-spec (resolve open Qs)  |
```

No end block, no per-feature deep dive — `/spk-recap` is a status board, not
a phase output. Reach for it at the start of a session to decide what to
pick up.

### Recipe: Undoing Completed Work With `/spk-rollback`

Problem this solves: a task got marked `[x]` too early, or the approach
turned out wrong and needs re-doing — `/spk-rollback` reverts the *declared*
state (checkboxes, optionally code) with full transparency about what moved.

```text
/spk-rollback partner-scheduling
```

```text
Completed tasks:
Phase T1: T1.1 (data model), T1.2 (retry window persistence)
Phase T2: T2.1 (scheduler integration)

Which tasks should I roll back? (IDs like T1.1,T1.2, a phase like "T1", or
"all")
```

```text
/spk-rollback partner-scheduling T2.1
```

```text
Rolled back: T2.1 -> [ ] (checkbox only; code changes kept)
completed=2, open=2

Slug: partner-scheduling
Status: implement
Artifacts: specs/active/partner-scheduling/tasks.md
Blockers: none
Ready for: /spk-implement partner-scheduling
```

Code is **not** reverted automatically — only say so explicitly if you also
want `git checkout -- <file>` run on the task's `Touches:` files.

### Recipe: Keeping `REPOSITORY_MAP.md` Honest With `/spk-repo-map`

Problem this solves: agents that navigate with `ls`/`find`/glob burn tokens
re-discovering a repo's shape every session. `REPOSITORY_MAP.md` is a
compact, code-only index meant to be read once per session instead.

Run it when the trigger checklist says so — not on every change:

```text
Added or removed a top-level code directory/module.
Moved/renamed key source paths that change navigation.
Added/removed runtime/service/CLI entrypoints.
Reshaped subsystem boundaries (where-to-edit paths changed materially).
```

```text
/spk-repo-map
```

```text
Changed entries:
+ Added `src/internal/importer/` (openspec/speckit migration) to Top-Level Code
+ Added "Add a CLI subcommand" -> src/internal/cli/ to Where To Edit

Map is up to date, 142/180 lines.

Slug: n/a
Status: repo-map
Artifacts: REPOSITORY_MAP.md
Blockers: none
```

A feature that only touches existing files inside existing modules should
*not* trigger a map update — that's the common case, and the agent should
recognize it as such rather than refreshing the map defensively.

### Recipe: Migrating From OpenSpec or Spec Kit

Problem this solves: you have an existing OpenSpec `openspec/changes/<slug>/`
or Spec Kit `specs/<slug>/` package and don't want to hand-copy it.

```bash
speckeep import openspec ./my-project
```

```text
imported: partner-scheduling
  spec.md    <- rebuilt from openspec Requirement:/Scenario: blocks (3 AC-*)
  plan.md    <- rebuilt from design.md (2 DEC-*)
  tasks.md   <- copied best-effort; run /spk-tasks to regenerate
             Touches:/Surface Map/Acceptance Coverage
skipped: dark-mode (already exists in specs/active/dark-mode/)
```

```bash
speckeep import speckit ./my-project
```

```text
imported: export-report
  spec.md    <- copied
  plan.md    <- copied
  tasks.md   <- copied
```

Existing speckeep feature directories are **never** overwritten — a
name collision is reported as `skipped`, not silently clobbered. For an
imported `tasks.md`, run `/spk-tasks <slug>` once to regenerate it with the
`Touches:`/`Surface Map`/`Acceptance Coverage` sections speckeep's
`implement`/`verify` phases rely on.

### Recipe: Wiring `speckeep guard` Into CI

Problem this solves: "is every feature touched by this PR actually
archive-ready?" as a deterministic check instead of a reviewer opinion.

```bash
speckeep guard . --slug partner-scheduling --json
```

```json
{
  "slug": "partner-scheduling",
  "ready": false,
  "findings": [
    { "kind": "open_task", "task": "T2.2" },
    { "kind": "missing_proof", "task": "T2.3" }
  ]
}
```

Exit code 1 on any finding — that's what fails the PR. The bundled GitHub
Action (`contrib/ci/speckeep-guard.yml`) detects changed
`specs/active/<slug>/` directories against the PR base and runs this per
changed slug automatically; you don't have to enumerate slugs by hand in
most repos.

### Recipe: Keeping the Binary Current

Problem this solves: "am I running the latest speckeep, and can I update
without reaching for brew/scoop/curl again?"

```bash
speckeep self check
```

```text
installed: v1.0.0
latest:    v1.1.0
update available — run `speckeep self upgrade`
```

```bash
speckeep self upgrade
```

```text
downloading speckeep_v1.1.0_linux_amd64.tar.gz...
sha256 verified
replaced: /home/you/.local/bin/speckeep (v1.0.0 -> v1.1.0)
```

On Windows, or when the install directory isn't writable, `self upgrade`
prints a manual-install hint instead of failing silently — it never leaves
you with a half-replaced binary.

### Recipe: The Express Lane in Practice

Problem this solves: not every feature needs `data-model.md` and a formal
`plan.md` — for small, low-risk changes, that's process weight the feature
doesn't carry.

```text
/spk-propose Add a "last synced at" timestamp to the integration status
badge in the settings page.
```

The agent goes straight from `spec.md` to `tasks.md` — no `plan.md`, no
`data-model.md`:

```text
Slug: last-synced-badge
Status: propose
Artifacts: specs/active/last-synced-badge/spec.md, specs/active/last-synced-badge/tasks.md
Blockers: none
Ready for: /spk-implement last-synced-badge
```

`speckeep check last-synced-badge` prints an express-mode hint instead of an
error for the missing `plan.md`:

```text
express mode: plan.md not present — closing from spec.md + tasks.md is fine
for this feature size.
```

Any plan-level decision that *is* worth recording (e.g. "reuse the existing
polling interval, don't add a new one") goes into `tasks.md`'s
`## Implementation Context` section instead of a separate `plan.md` — one
file carries the load a three-file package would otherwise need for a
change this small.
