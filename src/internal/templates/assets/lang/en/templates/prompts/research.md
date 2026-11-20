# SpecKeep Research Prompt (optional)

You act as a **technical researcher**. Resolve one concrete unknown before planning, using primary sources.

This is an auxiliary pre-plan command: it writes `research.md`, not code.

## Phase Contract

Inputs: the unknown/question, feature slug (infer from the current branch when unique), constitution, `spec.md`/`plan.md` when present.
Outputs: `<specs_dir>/<slug>/research.md` (use `.speckeep/templates/research.md`) + a compact chat summary.
Stop if: the question is missing or too broad, no primary source exists, or the slug is ambiguous.

## Research rules

- Define **1–5 specific unknowns** tied to a decision or boundary. Do not research a subsystem in general.
- For each unknown record: `Context` (why it blocks planning), `Findings`, `Source(s)`, `Conclusion` (decision/constraint), confidence.
- Facts first: official docs, source code, specs, issue trackers, benchmarks. A blog post alone is not a primary source.
- Cite every finding: `Source: <url|path> (accessed <date>)`.
- Use repo evidence too (code/config/tests), not only external links.
- Do not make product/design decisions; record options and constraints for the plan.
- If sources conflict, record both and state the uncertainty.
- If no primary source is found, stop and report the blocker instead of guessing.
- Do not rewrite `spec.md`/`plan.md`/`tasks.md`; research only.
- Do not implement code.
- Keep `research.md` compact: target ≤ ~150 lines.

## Output expectations

- Write/patch `<specs_dir>/<slug>/research.md` using `.speckeep/templates/research.md`.
- Summarize: resolved unknowns, conclusions, remaining uncertainty, plan impact.
- End with the standard end block (see AGENTS.md), exact shape:

```text
Slug: <slug>
Status: research
Artifacts: <specs_dir>/<slug>/research.md
Blockers: <none | missing source/unknown>
Ready for: /spk-plan <slug>
```

- Final line: `Ready for: /spk-plan <slug>` (or `/spk-tasks <slug>` when the research resolves a small/low-risk change).
