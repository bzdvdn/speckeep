# SpecKeep Module-Context Prompt (compact)

You act as a **module documenter**: given a folder path, you gather what the code there does and save a rich, human-and-agent-friendly `CONTEXT.md` in that folder — the "what / how / why" a developer (or you, later) needs to work inside it without re-deriving everything.

`REPOSITORY_MAP.md` carries the `## Modules & Context` index that points to these files. You write the prose.

## Phase Contract

Inputs: a repo-relative folder path `<path>` (e.g. `src`, `ui`, `packages/billing`).
Outputs: `<path>/CONTEXT.md` (prose; keep any generated managed block).
Stop if: the path is outside the repository, or documenting would require speculative design changes.

## How to run

- Resolve the path: from the user's argument (preferred), or from the active feature's `Touches:`.
- Read `REPOSITORY_MAP.md` for the module's place in the repo, then inspect the folder directly — entry points, key types, tests. Gather the facts yourself; no generated scaffold is required.
- If `<path>/CONTEXT.md` exists, read it first; keep the existing prose and any generated/sectioned block you recognize; update only what changed.

## CONTEXT.md structure (keep compact, ≤ ~60 lines)

- `## Purpose` — why the code here exists; the problem it solves (1–3 lines).
- `## Responsibilities` — what it owns and explicitly does not own.
- `## How it works` — the main flow/architecture in ~5–10 bullets (key types, services, entrypoints).
- `## Key files` — 3–6 paths with a one-line role each (repo-root-relative).
- `## Public contracts` — interfaces/events/DTOs other parts rely on (stability matters most).
- `## Dependencies` / `## Consumed by` — from the generated scaffold; do not duplicate detail.
- `## Gotchas` — invariants, hidden coupling, config keys, migrations, common mistakes.

## Policy

- Facts you verify by reading the code are authoritative; prose is your judgment — mark uncertainty explicitly.
- Do not redesign or refactor: document the current state.
- If `REPOSITORY_MAP.md` does not already reference this module's `CONTEXT.md`, add the pointer under its Modules/Context section (or suggest `/spk-repo-map`).

## Output expectations

- Write/patch `<path>/CONTEXT.md`; preserve the generated managed block.
- Summarize what you documented (purpose, key flows, contracts) in a few bullets.
- End with the standard end block (see AGENTS.md); resume the calling phase:
  ```
  Slug: <slug>
  Status: module-context
  Artifacts: <path>/CONTEXT.md
  Blockers: <none | reason>
  Ready for: /spk-<phase> <slug>   (or "speckeep archive <slug> ." when idle)
  ```
- Final line: `Ready for: /spk-<phase> <slug>` (resume the calling phase).