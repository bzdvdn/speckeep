# [PROJECT_NAME] Constitution

## Purpose

[PURPOSE]

## Core Principles

### [PRINCIPLE_1_NAME]
<!-- Example: I. Scope Discipline -->

[PRINCIPLE_1_RULES]

### [PRINCIPLE_2_NAME]
<!-- Example: II. Architecture Boundaries -->

[PRINCIPLE_2_RULES]

### [PRINCIPLE_3_NAME]
<!-- Example: III. Traceability (NON-NEGOTIABLE) -->

[PRINCIPLE_3_RULES]

### [PRINCIPLE_4_NAME]
<!-- Example: IV. Verify Before Archive -->

[PRINCIPLE_4_RULES]

### [PRINCIPLE_5_NAME]
<!-- Example: V. Simplicity and Operability -->

[PRINCIPLE_5_RULES]

[ADDITIONAL_PRINCIPLES]

## Non-Negotiable Rules

- Rules in this section are `MUST` / `MUST NOT` and are enforceable.
- Implementation `MUST` follow active spec/plan/tasks and remain in declared scope.
- Work `MUST NOT` proceed from ambiguous or placeholder requirements.
- Public behavior changes `MUST` be reflected in specs/tasks before merge.
- If implementation conflicts with this constitution, amend constitution first.

## Constraints

[CONSTRAINTS]

## Tech Stack

[TECH_STACK]

## Core Architecture

[ARCHITECTURE]

## Language Policy

- Documentation language: [DOCS_LANGUAGE]
- Agent interaction language: [AGENT_LANGUAGE]
- Code comment language: [COMMENTS_LANGUAGE]

## Development Workflow

- Each feature MUST be developed in a dedicated git branch.
- Feature branches SHOULD follow the project's feature branch naming convention such as `feature/<slug>`.
- Work SHOULD begin from an explicit spec before implementation starts.
- Plans and tasks SHOULD be derived from the active spec and remain aligned with it.
- Implementation, specs, plans, and tasks MUST comply with this constitution.
- If work reveals a conflict with this constitution, the constitution MUST be amended before incompatible implementation proceeds.

## Definition of Done

- A task is done only with observable proof: changed files, targeted test output, or command result.
- Traceability markers are mandatory on non-trivial changes:
  - code: `@sk-task <slug>#<TASK_ID>: <short> (<AC_ID>)`
  - tests: `@sk-test <slug>#<TASK_ID>: <TestName> (<AC_ID>)`
  - if multiple tests/cases verify the same task, `@sk-test <slug>#<TASK_ID>` must appear on each such test/case, not just on one representative test.
- Marker placement rule:
  - place trace markers on function/method/struct/class declarations (or behavior block headers), not on field lines.
  - never place trace markers at `package`, `import`, or file-header comment level; every marker must belong to a concrete owning symbol.
  - if the language has named declarations, place the marker directly above the declaration that owns the behavior or test.
- Placement and style examples by language:
  - Go: `//` directly above `func`, method receiver, `type`, `func Test...`; if multiple `Test...` functions verify one task, each of them needs its own `@sk-test`.
  - Python: `#` as the first line inside the body of `def` / `async def` / `class` / `def test_*`; never in the module docstring and never above `import`. If one task is covered by multiple test functions, each function needs its own marker in its body.
  - JavaScript / TypeScript: `//` above `function`, `async function`, class method, `class`; for `test(...)`/`it(...)`, put it on the first line inside the callback/body. If multiple cases verify one task, each `test/it` needs the marker.
  - Java: `//` or `/* */` directly above `class`, `interface`, `enum`, method, or JUnit test method; never above `package`, `import`, or fields.
  - C# / .NET: `//` or `/* */` directly above `class`, `record`, `interface`, method, or test method; never above `using`, namespace headers, or properties/fields without owning behavior.
  - C / C++: `//` or `/* */` above function, method, class/struct declaration, or test case macro/block; never above `#include`, namespace-only headers, or standalone struct fields.
  - Shell / Bash: `#` above `function name()` or as the first line of a named behavior/test block; never in the file header just to make the file look traced.
  - SQL / migrations: `--` or `/* */` directly above `CREATE FUNCTION|PROCEDURE|TRIGGER|VIEW` or on the first line of an explicitly named migration block; never in a top-of-file comment with no concrete target.
- Existing trace markers MUST be preserved; new task coverage appends markers (do not overwrite).
- If one method/test covers multiple tasks, multiple markers MUST coexist on that method/test.
- Verification MUST confirm acceptance-criteria coverage before archive.

## Repository Map Policy

- `REPOSITORY_MAP.md` is a compact code-navigation index, not a process document.
- Update the map only when code structure/navigation changes materially.
- Map updates MUST be minimal-diff and in-place; do not rewrite unchanged sections.
- Exclude operational/spec artifacts from indexing when configured by project policy.

[ADDITIONAL_REQUIRED_SECTIONS]

## Governance

- This constitution is authoritative for project decisions.
- Changes to architecture, specs, plans, and tasks MUST comply with these principles.
- If implementation conflicts with this constitution, the constitution wins unless it is explicitly amended first.
- Amend by patching this file, preserving mandatory sections and keeping guidance concrete and testable.

## Constitution Metadata

- Version: [CONSTITUTION_VERSION]
- Ratified: [RATIFICATION_DATE]
- Last Amended: [LAST_AMENDED_DATE]

## Last Updated

[YYYY-MM-DD]
