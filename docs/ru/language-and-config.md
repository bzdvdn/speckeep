# Языки и Конфигурация

## Поддерживаемые языки

Сейчас SpecKeep поддерживает:

- `en`
- `ru`

## Языковые настройки

Основной config хранит четыре языковых значения:

```yaml
language:
  default: en
  docs: en
  agent: en
  comments: en
```

### `default`

Базовый язык, который используется, когда более точные настройки не переопределены.

### `docs`

Управляет генерируемой проектной документацией, например:

- `constitution.md`
- шаблонами spec
- шаблонами plan
- шаблонами tasks

### `agent`

Управляет генерируемыми prompts и guidance, которое вставляется в `AGENTS.md`.

### `comments`

Фиксирует предпочитаемый язык новых или измененных комментариев в коде во время `implement`.

## Канонические BDD-маркеры

Даже когда `docs` установлен в `ru`, SpecKeep оставляет `Given / When / Then` каноническими acceptance markers. Остальной текст вокруг них должен следовать выбранному языку документации.

## Файл конфигурации

Основной config лежит в `.speckeep/speckeep.yaml`. Его создаёт `speckeep init`, а `speckeep refresh` поддерживает управляемые части в актуальном состоянии; правьте вручную только при осознанном изменении layout или политики.

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

Ключевые настройки:

| Ключ | По умолчанию | Смысл |
|---|---|---|
| `paths.specs_dir` | `specs/active` | активные feature packages |
| `paths.archive_dir` | `specs/archived` | архивные снимки |
| `paths.templates_dir` | `.speckeep/templates` | сгенерированные шаблоны |
| `paths.scripts_dir` | `.speckeep/scripts` | сгенерированные helper-скрипты |
| `runtime.shell` | `sh` | семейство скриптов (`sh` или `powershell`) |
| `language.*` | `en` | языки docs / agent / комментариев в коде |
| `agents.targets` | `[]` | включённые agent targets; управляются через `add-agent` / `remove-agent` |
| `agents.update_agents_md` | `true` | поддерживать управляемый блок в `AGENTS.md` |
| `workflow.verify` | `optional` | `required` делает `verify: pass` жёстким гейтом перед archive |
| `trace.ignore` | `vendor/`, `node_modules/`, `dist/` | префиксы, исключаемые из change attribution |

Примечания:

- `init --specs-root <dir>` записывает `paths.specs_dir=<dir>/active` и `paths.archive_dir=<dir>/archived`. Явные `--specs-dir` / `--archive-dir` задают точные пути.
- Смена `paths` после init не переносит существующие файлы: перенесите их сами, затем выполните `speckeep doctor .` и `speckeep refresh .`.
- `templates.*` и `scripts.*` переименовывают генерируемые файлы; оставляйте значения по умолчанию без конкретной причины.
- `workflow.schema` выбирает пресет графа зависимостей артефактов; см. `speckeep schema`.
- `speckeep doctor .` сообщает о расхождении config и workspace.
