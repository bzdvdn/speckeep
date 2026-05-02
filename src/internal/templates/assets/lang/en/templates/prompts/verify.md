# SpecKeep Verify Prompt (compact)

You verify one implemented feature against `tasks.md` and related `AC-*`.

Follow base rules in `AGENTS.md`.

## Phase Contract

Inputs: `<specs_dir>/<slug>/plan/tasks.md` (entrypoint), optionally `spec.digest.md` (preferred) or `summary.md`/`spec.md` when AC context is needed, `plan.md` when design surfaces must be confirmed.
Outputs: `<specs_dir>/<slug>/plan/verify.md` (or a chat report) + `tasks.md` status fixes when a checkbox is wrong.
Stop if: `tasks.md` missing or slug ambiguous.

## Rules

- Treat verify as an evidence log (task/AC → proof), not a reassurance ritual.
- Start from `tasks.md`: every `[x]` task must have observable proof in the repo (file/test/command output).
- If `./.speckeep/scripts/verify-task-state.*` exists, run it (slug first) as a cheap first pass.
- If you persist the report to a file, use `.speckeep/templates/verify.md` as the format and write to `<specs_dir>/<slug>/plan/verify.md`. Do not look for “examples” in other slugs’ verify reports for shape.
- Collect traceability as a cheap integrity check: use `./.speckeep/scripts/trace.* <slug>` (and `./.speckeep/scripts/trace.* <slug> --tests` when needed). This does not replace proof, but quickly highlights gaps/orphaned annotations.
- Deep code reads only to confirm a specific claim for a specific task/AC.
- Avoid re-reading the same files repeatedly: keep focused evidence notes and use targeted slices + `git diff` when you need to confirm what changed.
- Verify is not redesign: report mismatches and blockers, don’t expand scope.
- Evidence format: list `<TASK_ID> -> proof` lines (and `AC-* -> <TASK_ID>` when relevant), where `<TASK_ID>` is either `T*` or `<slug>#T*`. Proof must be something observable: changed file path, test name/output, command output, or a documented artifact.
- Task status hygiene: do not flip a checkbox to `[x]` unless proof is present. If evidence is missing or ambiguous, keep it `[ ]` and mark the feature `concerns`/`blocked` with a concrete next step.
- Keep claims scoped: avoid a broad repository sweep instead of focused evidence. If a claim cannot be confirmed from tasks + plan artifacts + targeted code inspection, mark it as `Not Verified` and avoid upgrading to `pass`.
- Send the feature back to the narrowest earlier phase that can honestly fix it (usually `tasks` for coverage gaps, `plan` for missing design surfaces, `spec` for intent/AC issues).
- Prefer `concerns` over `pass` when evidence is partial but no contradiction is found.
- Do not use `pass` unless the completed task state is confirmed and no AC-critical gaps remain.

## Output expectations

- Verdict: `pass|concerns|blocked` + concrete mismatches (task/AC → evidence).
- Include `## Not Verified` items when you did not confirm something (explicitly list what you did not check).
- If `blocked`, end with `Return to: /speckeep.<phase> <slug>`.
- Include a short summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, and either `Ready for` or `Return to`.
- If `pass`, final line: `Ready for: speckeep archive <slug> .`
