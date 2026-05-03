# Change Spec Archive Roadmap

Переход на новый default layout:

- `specs_dir = specs/active`
- `archive_dir = specs/archived`

## Status

- done: config defaults and core path support
- done: refresh/doctor migration behavior for legacy default layout
- done: prompt/template/doc updates
- done: demo assets and demo workspace paths
- remaining: final wording sweep in ancillary roadmap/architecture notes if needed

## Must Update

- [src/internal/config/config.go](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/config/config.go)
  - defaults: `specs/active`, `specs/archived`
- [src/internal/project/init.go](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/project/init.go)
  - init directory creation
- [src/internal/project/refresh.go](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/project/refresh.go)
  - safe move from old defaults
- [src/internal/featurepaths/featurepaths.go](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/featurepaths/featurepaths.go)
  - active slug discovery assumptions
- [src/internal/workflow/state.go](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/workflow/state.go)
  - active/archive slug collection
- [src/internal/specs/specs.go](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/specs/specs.go)
  - list/show behavior over new root
- [src/internal/cli/archive.go](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/cli/archive.go)
  - archive/restore help text and paths
- [src/internal/cli/list_archive.go](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/cli/list_archive.go)
  - default path wording
- [src/internal/doctor/doctor.go](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/doctor/doctor.go)
  - workspace health around new defaults
- [src/internal/templates/assets/lang/en/templates/agents-snippet.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/templates/assets/lang/en/templates/agents-snippet.md)
- [src/internal/templates/assets/lang/ru/templates/agents-snippet.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/templates/assets/lang/ru/templates/agents-snippet.md)
  - fallback defaults and repo-map indexing rules
- [README.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/README.md)
- [README.ru.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/README.ru.md)
- [docs/en/cli.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/docs/en/cli.md)
- [docs/ru/cli.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/docs/ru/cli.md)
- [docs/en/workflow.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/docs/en/workflow.md)
- [docs/ru/workflow.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/docs/ru/workflow.md)
- [docs/en/examples.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/docs/en/examples.md)
- [docs/ru/examples.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/docs/ru/examples.md)

## Should Update

- [src/internal/templates/assets/lang/en/templates/inspect-report.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/templates/assets/lang/en/templates/inspect-report.md)
- [src/internal/templates/assets/lang/ru/templates/inspect-report.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/templates/assets/lang/ru/templates/inspect-report.md)
- [src/internal/templates/assets/lang/en/templates/verify-report.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/templates/assets/lang/en/templates/verify-report.md)
- [src/internal/templates/assets/lang/ru/templates/verify-report.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/templates/assets/lang/ru/templates/verify-report.md)
- [src/internal/templates/assets/lang/en/templates/inspect.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/templates/assets/lang/en/templates/inspect.md)
- [src/internal/templates/assets/lang/ru/templates/inspect.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/templates/assets/lang/ru/templates/inspect.md)
- [src/internal/templates/assets/lang/en/templates/verify.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/templates/assets/lang/en/templates/verify.md)
- [src/internal/templates/assets/lang/ru/templates/verify.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/templates/assets/lang/ru/templates/verify.md)
- [docs/en/agents.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/docs/en/agents.md)
- [docs/ru/agents.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/docs/ru/agents.md)
- [docs/en/glossary.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/docs/en/glossary.md)
- [docs/ru/glossary.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/docs/ru/glossary.md)
- [docs/en/overview.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/docs/en/overview.md)
- [docs/ru/overview.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/docs/ru/overview.md)
- [docs/en/faq.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/docs/en/faq.md)
- [docs/ru/faq.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/docs/ru/faq.md)
- [CHANGELOG.md](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/CHANGELOG.md)

## Nice To Update

- architecture/roadmap wording where old path examples still appear
- output copy in help/panels where old defaults are implied rather than explicit

## Tests That Will Almost Certainly Need Touching

- [src/internal/config/config_test.go](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/config/config_test.go)
- [src/internal/project/init_test.go](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/project/init_test.go)
- [src/internal/cli/root_test.go](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/cli/root_test.go)
- [src/internal/featurepaths/featurepaths_test.go](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/featurepaths/featurepaths_test.go)
- [src/internal/workflow/state_test.go](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/workflow/state_test.go)
- [src/internal/doctor/doctor_test.go](/home/bzdv/PAT_PROJECTS/DRAFTSPEC/src/internal/doctor/doctor_test.go)
- workflow integration tests
- archive/list-archive tests

## Best Order

1. code support + tests
2. init/refresh migration behavior
3. prompt/snippet defaults
4. docs/examples
5. changelog

## Important Sanity Checks After

- old project with `specs` + `archive` still works
- new `init` produces `specs/active` + `specs/archived`
- `list-specs` only sees active features
- `list-archive` only sees archived snapshots
- `doctor` does not confuse archived slugs with active specs
- prompts/snippets never tell the agent to fall back to plain `specs`/`archive`
