# Self-Hosting Walkthrough

This repository builds the `speckeep` CLI itself, so it is a natural dogfooding example.

## Why This Matters

Self-hosting is one of the strongest trust signals for SpecKeep.

It shows that:

- the workflow is practical enough to use on a real CLI codebase
- the generated agent guidance is not only for toy examples
- the product can help evolve its own prompts, scripts, checks, and templates

## Source Of Truth Reminder

In this repository:

- `README.md`
- `MVP.md`
- Go source under `src/`
- embedded assets under `src/internal/templates/assets/`

are the product source of truth.

Generated self-hosting files such as `/.speckeep/` and repository-local `AGENTS.md` are test artifacts and should not be committed as canonical design inputs.

## Good Self-Hosting Demo Themes

Strong examples in this repository include:

- refining agent integration behavior
- improving `refresh` or `doctor`
- tightening inspect or verify checks
- aligning generated templates with CLI behavior
- reducing drift between `init` and `refresh`

## Example Story

One practical self-hosting storyline:

1. capture a product issue in the current repo
2. write or refine the feature spec
3. inspect it before planning
4. write the plan and tasks
5. implement the change in Go and templates
6. verify the generated output still matches the product contract

This is especially compelling when the change affects both:

- CLI behavior
- generated assets or prompts

because that is where SpecKeep's strictness is easiest to demonstrate.

## What To Show Publicly

For a public self-hosting demo, focus on:

- one real bug or refinement
- the feature-local artifacts involved
- the code changes
- the final generated output or behavior check

Avoid trying to show the whole workflow history for the whole project.

## Suggested Output Format

The self-hosting story works best as:

- a short markdown walkthrough
- links to the changed product files
- a before/after summary
- optionally a short terminal clip for one narrow step such as `doctor` or `refresh`
