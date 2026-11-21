# Language and Configuration

## Supported Languages

SpecKeep currently supports:

- `en`
- `ru`

## Language Settings

The generated config stores four language values:

```yaml
language:
  default: en
  docs: en
  agent: en
  comments: en
```

### `default`

Base language used when the more specific values are not overridden.

### `docs`

Controls generated project documentation such as:

- `constitution.md`
- spec templates
- plan templates
- tasks templates

### `agent`

Controls generated prompts and `AGENTS.md` guidance.

### `comments`

Records the preferred language for new or edited code comments during `implement`.

## Canonical BDD Markers

Even when `docs` is set to Russian, SpecKeep keeps `Given / When / Then` as canonical acceptance markers. The surrounding text should still follow the configured documentation language.

## Config File

The main config lives in `.speckeep/speckeep.yaml`. `speckeep init` generates it and `speckeep refresh` keeps managed parts aligned; edit it only for intentional layout or policy changes.

```yaml
version: 1
project:
  name: my-project
  constitution_file: CONSTITUTION.md
runtime:
  shell: sh                # sh | powershell
paths:
  specs_dir: specs/active
  archive_dir: specs/archived
  templates_dir: .speckeep/templates
  scripts_dir: .speckeep/scripts
language:
  default: en
  docs: en
  agent: en
  comments: en
agents:
  update_agents_md: true
  agents_file: AGENTS.md
  targets: [claude]
workflow:
  verify: optional          # optional | required
trace:
  ignore:
    - vendor/
    - node_modules/
    - dist/
```

Key settings:

| Key | Default | Meaning |
|---|---|---|
| `paths.specs_dir` | `specs/active` | active feature packages |
| `paths.archive_dir` | `specs/archived` | archive snapshots |
| `paths.templates_dir` | `.speckeep/templates` | generated templates |
| `paths.scripts_dir` | `.speckeep/scripts` | generated helper scripts |
| `runtime.shell` | `sh` | generated script family (`sh` or `powershell`) |
| `language.*` | `en` | docs / agent / code-comment languages |
| `agents.targets` | `[]` | generated agent targets; manage with `add-agent` / `remove-agent` |
| `agents.update_agents_md` | `true` | maintain the managed block in `AGENTS.md` |
| `workflow.verify` | `optional` | `required` makes `verify: pass` a hard pre-archive gate |
| `trace.ignore` | `vendor/`, `node_modules/`, `dist/` | repo-relative prefixes excluded from change attribution |

Notes:

- `init --specs-root <dir>` writes `paths.specs_dir=<dir>/active` and `paths.archive_dir=<dir>/archived`. Explicit `--specs-dir` / `--archive-dir` set exact paths instead.
- Changing `paths` after init does not move existing files: move them yourself, then run `speckeep doctor .` and `speckeep refresh .`.
- `templates.*` and `scripts.*` rename generated files; keep the defaults unless you have a concrete reason.
- `workflow.schema` selects the artifact dependency preset; see `speckeep schema`.
- `speckeep doctor .` reports config/workspace drift.
