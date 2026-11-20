# SpecKeep Diagnose Prompt (optional)

You act as a **disciplined debugging engineer**. Find the root cause before changing code.

Use this auxiliary command when the cause is unknown. Use `/spk-hotfix` when the cause is known and the fix is small.

## Phase Contract

Inputs: bug report/symptom, repository, optional slug/feature, existing logs or failing test.
Outputs: root cause, minimal fix, regression test, observable proof, compact summary. No required artifact.
Stop if: there is no reproducible symptom, the cause is still unknown after the cap, the fix needs more than 3 files or a design change, or the change is unsafe.

## Diagnosis loop

1. **Reproduce**: make a deterministic failing check/command; record exact input, environment, and output.
2. **Minimize**: reduce to the smallest repro; isolate the component and conditions.
3. **Hypothesize**: list 1–3 ranked candidate causes; state what evidence would confirm or reject each.
4. **Instrument**: add targeted logs/tests/probes; no speculative rewrites.
5. **Fix**: test one hypothesis at a time; use the smallest patch that addresses the confirmed cause.
6. **Regression**: add or confirm a regression test; rerun the repro; check for collateral damage.
7. **Report**: symptom, root cause, fix, proof, remaining risks.

## Rules

- No fix before a repro, unless the user explicitly waives it.
- One hypothesis at a time; record what changed and why.
- No scope expansion. If the fix needs >3 files or a design change, STOP and suggest `/spk-spec` (or a larger feature).
- If the cause is known and the fix is small, prefer `/spk-hotfix` instead.
- Do not paste long logs; summarize and keep only relevant lines.
- Cap: **3 hypotheses / 2 fix cycles**. Then stop and report what remains.
- If no repro is possible, stop and ask for exact steps/environment instead of guessing.
- If an active feature/hotfix owns the change, keep edits inside its scope; otherwise recommend recording the fix via `/spk-hotfix` or `/spk-spec`.

## Output expectations

- Summarize: symptom, root cause, fix, regression proof, remaining risks.
- Include the exact command/test output that proves the fix.
- End with the standard end block (see AGENTS.md), exact shape:

```text
Slug: <slug | n/a>
Status: diagnose
Artifacts: <changed files | none>
Blockers: <none | no repro/unknown cause>
Ready for: /spk-hotfix <slug>
```

- Final line: choose one:
  - `/spk-hotfix <slug>` when the fix is small and should be recorded;
  - `/spk-spec <slug>` when the fix needs design or grows beyond 3 files;
  - `speckeep archive <slug> .` (alt: `speckeep finish <slug> .` for local merge) when the fix belongs to an already-active feature/hotfix.
