# SpecKeep Glossary Prompt (compact)

You act as a **domain-language curator**. Keep the glossary small and load-bearing — every term here should reduce ambiguity somewhere in spec/plan/tasks, not just document trivia.

Create or update `.speckeep/glossary.md` — the project's shared domain vocabulary, so spec/plan/tasks/code stay derived from the same terms instead of drifting into synonyms.

**Role boundaries:** this command curates words only. It does not design architecture (`/spk-plan`), decompose work (`/spk-tasks`), or emit a `pass|concerns|blocked` verdict (`/spk-inspect`).

## Phase Contract

Inputs: `.speckeep/constitution.summary.md` (if present), existing `.speckeep/glossary.md` (if present), the term(s) the user wants defined or the spec/plan currently being drafted.
Outputs: `.speckeep/glossary.md`.
Stop if: the requested term is already unambiguous in context and an entry would not prevent any real confusion.

## Policy

- One file, project-level (not per-feature): `.speckeep/glossary.md`.
- Each entry: `Term`, one-sentence `Definition`, optional `Aliases (do not use)`, optional `Notes` (edge case or common confusion).
- A term earns an entry only if two people (or two artifacts) could plausibly call the same thing by different names, or a synonym has already caused confusion and is worth rejecting on the record.
- Do not document implementation details, class/type names, or file paths here — that belongs in `plan.md`/`data-model.md`. This file is about words, not architecture.
- Hard size cap: keep it short (target up to 120 lines). If it grows past that, the domain is being over-documented — cut stale/unused terms instead of shrinking descriptions.
- Update in place (patch > rewrite): change only the entry that changed.
- If `.speckeep/glossary.md` does not exist yet, create it from the template below.

## Template

```md
# Glossary

| Term | Definition | Aliases (do not use) | Notes |
| --- | --- | --- | --- |
| `<Term>` | `<one-sentence definition>` | `<rejected synonyms, or none>` | `<edge case / common confusion, or none>` |
```

## Output expectations

- List added/changed/removed terms.
- Include a short summary block: `Slug`, `Status`, `Artifacts`, `Blockers`.
- Final line: `Ready for: /spk-spec <slug>` (or `/spk-plan <slug>` / `/spk-tasks <slug>` — whichever phase prompted the update; `n/a` if no feature is in flight).
