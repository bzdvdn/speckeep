# Agents

## Supported Agent Targets

SpecKeep can generate project-local command or prompt files for:

- `claude`
- `codex`
- `copilot`
- `cursor`
- `kilocode`
- `opencode`
- `trae`
- `windsurf`
- `roocode`
- `aider`
- `amazonq`
- `gemini`
- `jules`
- `cline`
- `devin`
- `goose`
- `refact`
- `codiumate`
- `qwen-code`
- `all`

### Maintenance Tiers

All targets share the same generator (`agents/skill_pack.go`) and get the identical composite `sdd` skill pack, so every target is functionally complete. The tiers below only describe how actively each target's ecosystem-specific integration (skills directory conventions, loader quirks) is verified against real tool behavior:

- **Tier 1 — maintained-first**: `claude`, `cursor`, `codex`, `copilot`, `windsurf`, `opencode`. These get priority attention when an agent-ecosystem change (skills loader behavior, directory conventions) needs adapting.
- **Tier 2 — best-effort / community**: every other target (`kilocode`, `trae`, `roocode`, `aider`, `amazonq`, `gemini`, `jules`, `cline`, `devin`, `goose`, `refact`, `codiumate`, `qwen-code`). These are generated the same way and expected to work, but ecosystem-specific quirks are fixed on report rather than proactively tracked.

If you rely on a Tier 2 target and hit a mismatch with how its tool actually loads skills, please open an issue — that's the signal that moves a target up a tier.

### Slash-Invocation: What's Actually Verified

Every target gets an identical `sdd/` overview skill plus one `spk-<phase>/` skill per phase, in that tool's own `skills/`-equivalent directory. Whether typing `/spk-spec` (or `@spk-spec`) actually works, though, depends on how that specific tool's runtime treats a Skills directory — that's runtime behavior, not something speckeep controls, and it differs per tool. All 19 targets have now been individually checked against official docs:

| Target | Invocable by name? | Notes |
| --- | --- | --- |
| `claude` | Yes, `/name` | Confirmed against official Claude Code docs. |
| `cursor` | Yes, `/name` | Confirmed against official Cursor docs. |
| `qwen-code` | Yes, `/name` | Confirmed — Skills are genuinely slash-invocable, matching speckeep's `.qwen/skills/` output exactly. |
| `codex` | Likely, `/name` | Skills are slash-invocable per OpenAI's docs, but the *canonical* documented path is `.agents/skills/`, not `.codex/skills/` — `.codex/skills/` is community-documented and not independently confirmed as read by Codex CLI. |
| `copilot` | Yes (CLI), `/name` | Confirmed for Copilot **CLI** against official docs. VS Code Copilot Chat has its own separate "Agent Skills" documentation not yet cross-checked for parity. |
| `windsurf` | No via Skills — **yes via generated Workflow** | Windsurf Skills load automatically or via `@mention`, never `/`. speckeep also generates `.windsurf/workflows/spk-<phase>.md` (Windsurf's real slash-command mechanism), so `/spk-spec` genuinely works there. |
| `opencode` | No via Skills — **yes via generated Command** | OpenCode Skills are loaded by the agent calling a `skill()` tool, never `/`. speckeep also generates `.opencode/commands/spk-<phase>.md` (OpenCode's real slash-command mechanism), so `/spk-spec` genuinely works there. |
| `cline` | No via Skills (`.clinerules/` is context, not commands) — **yes via generated Workflow** | speckeep also generates `.clinerules/workflows/spk-<phase>.md` (Cline's real slash-command mechanism, invoked as `/spk-<phase>.md`). |
| `amazonq` | No via `/`, but **yes via generated Prompt with `@name`** | Amazon Q Developer has no confirmed native Skills reader at all (the only official "Skills" doc found belongs to an unrelated "Agent Toolkit for AWS" plugin). speckeep now writes to the real dotfile root `.amazonq/` (not `.q/` — the old path was likely never read) and generates `.amazonq/prompts/spk-<phase>.md`, invoked with `@spk-<phase>`, not a slash. |
| `gemini` | No via Skills — **yes via generated TOML command** | Gemini CLI Skills are model-invoked only. speckeep also generates `.gemini/commands/spk-<phase>.toml` (Gemini's real command format — TOML, not markdown), so `/spk-spec` genuinely works there. |
| `devin` | No, `@skills:name` only | Skills load automatically or via explicit `@skills:name` mention, never `/`. `.agents/skills/` is the documented canonical path; speckeep's `.devin/skills/` is a confirmed-valid fallback path, just not the primary one. |
| `goose` | No, `/skills <name>` meta-command only | Not directly `/name`-invocable — reach it via `/skills <name>` or auto-trigger. Goose's own real slash commands are a separate "recipes" system (user-aliased parameterized YAML tasks), which speckeep does not generate. `.agents/skills/` is canonical; `.goose/skills/` is a confirmed-valid fallback. |
| `kilocode` | Unverified / likely no via Skills | Community docs describe Skills as model-invoked only, with real slash commands at a **`.kilo/commands/`** path — but those same docs also suggest the product's dotfile root may have moved from `.kilocode/` to `.kilo/`, which is not independently confirmed. speckeep has **not** changed `.kilocode/skills/` on the strength of a single source; treat this row as open until verified. |
| `roocode` | No, and no flat-command fallback exists | Skills are model-invoked only. Unlike Windsurf/OpenCode/Cline, Roo Code has no documented one-file-per-command mechanism — its only project-local convention (`.roo/rules/*.md`) is concatenated free-form context, not individually invocable. There is currently no way for speckeep to give Roo Code a real `/spk-spec`. |
| `trae` | Unverified, likely not | Trae's own docs describe natural-language/auto-selection triggering, not `/name`. An open upstream issue also suggests drop-in `SKILL.md` files may not be auto-discovered without registering through Trae's own UI/CLI. |
| `aider` | N/A — no Skills loader | Confirmed: Aider has no Skills or custom-slash-command mechanism. It only reads `CONVENTIONS.md` when explicitly told to (`/read`, `--read`, or `.aider.conf.yml`) — the existing `.aider/CONVENTIONS.md` pointer is the correct approach. |
| `jules` | N/A — no mechanism at all | Jules is an async/non-interactive agent; it only reads `AGENTS.md` (falling back to `README.md`) for context. Neither Skills nor slash commands apply — speckeep's `.jules/skills/` output is currently unread but harmless. |
| `refact` | Unverified, likely no | No Skills, SKILL.md, or project-local command convention found anywhere in Refact.ai's docs or GitHub. Its customization is UI-driven ("AI Toolbox", `Alt+T`), not a repo-local file. speckeep's `.refact/skills/` output is unconfirmed to be read at all. |
| `codiumate` | Unverified | **Product renamed**: CodiumAI → Qodo → "Codiumate" is now "Qodo Gen". Qodo does document an Agent-Skills-compatible system, but no official source confirms the installed project path or slash-invocability — treat as unconfirmed rather than assume parity. |

If you hit a target where the documented command doesn't appear, please open an issue with what you see — that's exactly the signal that moves a row from "unverified" to confirmed (or, like windsurf/opencode/cline/amazonq/gemini, a target-specific fix).

**`continue` was removed as a supported target.** Multiple secondary sources reported Continue.dev was acquired by Cursor and discontinued around mid-2026 (its GitHub repo made read-only); not worth maintaining generation for. `speckeep doctor` flags any leftover `.continue/skills/` output from before the removal, and `speckeep refresh`/`cleanup-agents` remove it; a `continue` entry lingering in an existing `speckeep.yaml` is dropped silently on the next refresh rather than erroring.

## Generated Locations

Every target receives the same skill set at `<target-skills-dir>/`: a lightweight `sdd/` overview skill plus one independent, self-contained `spk-<phase>/` skill per phase (e.g. `spk-spec/`, `spk-plan/`, `spk-implement/`, ...), designed to be directly slash-invocable — see the verification table above for which targets that's actually confirmed on:

- Claude: `.claude/skills/{sdd,spk-*}/`
- Codex: `.codex/skills/{sdd,spk-*}/`
- Copilot: `.github/skills/{sdd,spk-*}/`
- Cursor: `.cursor/skills/{sdd,spk-*}/`
- Kilo Code: `.kilocode/skills/{sdd,spk-*}/`
- OpenCode: `.opencode/skills/{sdd,spk-*}/` + `.opencode/commands/spk-*.md` (real slash mechanism)
- Trae: `.trae/skills/{sdd,spk-*}/`
- Windsurf: `.windsurf/skills/{sdd,spk-*}/` + `.windsurf/workflows/spk-*.md` (real slash mechanism)
- Roo Code: `.roo/skills/{sdd,spk-*}/`
- Aider: `.aider/skills/{sdd,spk-*}/` (plus `.aider/CONVENTIONS.md` pointer)
- Amazon Q: `.amazonq/skills/{sdd,spk-*}/` + `.amazonq/prompts/spk-*.md` (real mechanism, invoked with `@name`)
- Gemini: `.gemini/skills/{sdd,spk-*}/` + `.gemini/commands/spk-*.toml` (real slash mechanism)
- Jules: `.jules/skills/{sdd,spk-*}/`
- Cline: `.clinerules/skills/{sdd,spk-*}/` + `.clinerules/workflows/spk-*.md` (real slash mechanism)
- Devin: `.devin/skills/{sdd,spk-*}/`
- Goose: `.goose/skills/{sdd,spk-*}/`
- Refact: `.refact/skills/{sdd,spk-*}/`
- Codiumate: `.codiumate/skills/{sdd,spk-*}/`
- Qwen Code: `.qwen/skills/{sdd,spk-*}/`

`<target>/sdd/SKILL.md` is an overview (workflow chain, gates) for when the phase isn't obvious from the request; each `<target>/spk-<phase>/SKILL.md` is its own top-level skill — directly slash-invocable (e.g. `/spk-spec`) rather than only reachable through the model opening a linked file — inlining the canonical prompt body from `.speckeep/templates/prompts/`. The CLI remains the deterministic verification spine — a skill alone never replaces `check`/`guard` verdicts. Tools that do not auto-index a skills directory (e.g. Aider) get a single loader pointer; the managed `AGENTS.md` block always points at the skills as well.

## Agent Discipline

The agent-facing workflows are:

- `constitution`
- `spec`
- `inspect`
- `plan`
- `tasks`
- `implement`
- `verify`

Each prompt is designed to:

- read only the minimum required context
- stop when prerequisites are missing
- respect the configured documentation and agent languages
- preserve constitutional authority over specs, plans, tasks, and implementation

Each generated agent wrapper includes:

- **Workflow chain hint**: `constitution → spec → [inspect, optional] → plan → tasks → implement → verify → archive` — prevents agents from skipping required phases or jumping ahead, while making archive an explicit CLI follow-up after agent verification
- **Script execution discipline**: explicit instruction to execute scripts as shell commands, trust stdout/exit code, and never read or inspect script source
- **Anti-pattern block**: common mistakes to avoid — skipping readiness scripts, re-planning during implement, marking tasks done without observable proof, reading the full repository when minimal context is required

`spec` should stay branch-first:

- it should create or switch to `feature/<slug>` before writing `specs/active/<slug>/spec.md` when the environment allows it
- it should support `--name`, optional `--slug`, and optional `--branch` for chat-oriented input
- if `/spk.spec` is invoked with `--name` but without enough description, it should preserve context and ask for or accept the next message as the continuation of the spec request
- when the input comes from a local prompt file, it should prefer top-of-file `name:` and optional `slug:` metadata over a generic filename
- if the request is ambiguous, multi-feature, URL-like, or tries to derive one spec from multiple constitutional changes, it should stop and ask for one concrete feature

`verify` is intentionally lightweight:

- it starts from `tasks.md`
- it can use `.speckeep/scripts/verify-task-state.sh <slug>` as a cheap first-pass helper
- `.speckeep/scripts/*` wrappers compute the project root and pass it via `--root`, so they can be executed from any working directory
- it reads deeper artifacts only when needed to confirm a concrete claim
- it is meant to confirm readiness for archive or follow-up refinement, not to become a heavy review engine

After `verify: pass`, prefer the explicit CLI follow-up `speckeep archive <slug> .` so archiving stays outside the agent reasoning loop. Default archive status is `completed`; non-`completed` statuses require an explicit `--reason`. For `completed`, it is fine (and cheap) to reuse `verify-task-state.sh` before creating the snapshot.

## Maintenance Commands

Use the public CLI to manage agent targets safely:

```bash
speckeep add-agent my-project --agents claude --agents cursor
speckeep list-agents my-project
speckeep remove-agent my-project --agents cursor
speckeep cleanup-agents my-project
speckeep doctor my-project
```
