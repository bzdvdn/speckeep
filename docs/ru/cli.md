# CLI

## Установка

SpecKeep распространяется как один бинарник через GitHub Releases.

Linux:

```bash
VERSION=v0.3.1
curl -fsSL "https://raw.githubusercontent.com/bzdvdn/speckeep/${VERSION}/scripts/install.sh" | bash -s -- --version "${VERSION}"
```

Windows (PowerShell):

```powershell
$version="v0.3.1"
$env:DRAFTSPEC_VERSION=$version
powershell -ExecutionPolicy Bypass -c "iwr -useb https://raw.githubusercontent.com/bzdvdn/speckeep/$version/scripts/install.ps1 | iex"
```

Чтобы также добавить папку установки в `PATH`:

- Linux: добавь `--add-to-path` или установи `DRAFTSPEC_ADD_TO_PATH=1`
- Windows: установи `$env:DRAFTSPEC_ADD_TO_PATH=1` или запускай скрипт с `-AddToPath`

## Команды

### `speckeep init [path]`

Инициализирует SpecKeep workspace в целевом проекте.

Примеры:

```bash
speckeep init
speckeep init my-project --lang ru --shell sh
speckeep init my-project --lang ru --shell sh --specs-dir .speckeep/specifications --archive-dir .speckeep/artifacts/archive --constitution-file docs/constitution.md
speckeep init my-project --docs-lang ru --agent-lang en --comments-lang en --shell powershell --agents claude --agents cursor
```

Важные флаги:

- `--git` инициализирует Git-репозиторий; по умолчанию включен
- `--lang` задает базовый язык; по умолчанию `en`
- `--shell` выбирает семейство генерируемых workflow scripts; обязателен: `sh` или `powershell`
- `--specs-dir` переопределяет директорию спецификаций (advanced)
- `--archive-dir` переопределяет директорию архива (advanced)
- `--constitution-file` переопределяет путь к конституции проекта (advanced)
- `--docs-lang` задает язык генерируемой документации
- `--agent-lang` задает язык генерируемых промтов и guidance для агентов
- `--comments-lang` фиксирует предпочитаемый язык комментариев в коде
- `--agents` генерирует project-local agent files

### `speckeep refresh [path]`

Обновляет только SpecKeep-managed generated artifacts в уже существующем проекте.

Эта команда обновляет:

- `.speckeep/speckeep.yaml`
- `.speckeep/skills/manifest.yaml`
- `.speckeep/templates/**`
- `.speckeep/scripts/**`
- project-local agent command files
- managed SpecKeep block внутри `AGENTS.md`

	Эта команда не обновляет:

	- файл конституции (`project.constitution_file`, по умолчанию `CONSTITUTION.md`)
	- содержимое `specs_dir/**` (но может безопасно перенести директорию при `--specs-dir`)
	- содержимое `specs_dir/<slug>/plan/**`
	- содержимое `archive_dir/**` (но может безопасно перенести директорию при `--archive-dir`)

Lean-модель артефактов теперь является дефолтной для generated guidance и readiness checks. В существующих feature package ещё могут лежать legacy `summary.md`, `spec.digest.md` или `plan.digest.md`, но refresh их больше не требует, а новые workspace на них не опираются.

Примеры:

```bash
speckeep refresh my-project
speckeep refresh my-project --shell powershell --agents claude --dry-run
speckeep refresh my-project --agent-lang ru --json
```

	Важные флаги:

	- `--lang`, `--docs-lang`, `--agent-lang`, `--comments-lang` переопределяют существующие language settings из config
	- `--shell` переопределяет семейство генерируемых workflow scripts
	- `--constitution-file` переопределяет путь к конституции в config (и безопасно переносит существующий файл, когда это возможно)
	- `--specs-dir` переопределяет `paths.specs_dir` (и безопасно переносит существующую директорию спецификаций, когда это возможно)
	- `--archive-dir` переопределяет `paths.archive_dir` (и безопасно переносит существующую директорию архива, когда это возможно)
	- `--agents` переопределяет набор включенных project-local agent targets
	- `--dry-run` показывает pending changes без записи на диск
	- `--json` выводит результат refresh в JSON

### `speckeep add-agent [path]`

Добавляет один или несколько agent targets в уже инициализированный проект.

```bash
speckeep add-agent my-project --agents claude --agents codex
```

### `speckeep list-agents [path]`

Показывает включенные agent targets из `.speckeep/speckeep.yaml`.

### `speckeep remove-agent [path]`

Отключает один или несколько agent targets и удаляет их generated files.

### `speckeep cleanup-agents [path]`

Удаляет осиротевшие agent artifacts, которые больше не соответствуют включенным targets в config.

### `speckeep add-skill [path]`

Добавляет или обновляет один skill в `.speckeep/skills/manifest.yaml`.

Для git-источников `--ref` обязателен: это фиксирует версию и предотвращает drift на плавающих ветках.

Для git-источников SpecKeep materialize'ит checkout в `.speckeep/skills/checkouts/<id>`. Это runtime-cache для skill source, и SpecKeep поддерживает для него managed block в корневом `.gitignore`.

Используйте `--no-install`, чтобы обновить только manifest/AGENTS без немедленной установки в agent skill folders.

Примеры:

```bash
speckeep add-skill my-project --id architecture --from-local skills/architecture
speckeep add-skill my-project --id openai-docs --from-git https://example.com/skills.git --ref v1.2.3 --path skills/openai-docs
```

### `speckeep list-skills [path]`

Показывает настроенные skills из `.speckeep/skills/manifest.yaml`.

Для machine-readable output используйте `--json`.

### `speckeep remove-skill [path]`

Удаляет один skill из `.speckeep/skills/manifest.yaml`.

Используйте `--no-install`, чтобы пропустить немедленную синхронизацию установленных skills в agent folders.

### `speckeep install-skills [path]`

Устанавливает включенные skills из `.speckeep/skills/manifest.yaml` в skill-папки выбранных агентов.

По умолчанию используются targets из `.speckeep/speckeep.yaml`. Можно переопределить через `--targets codex,opencode`.

Для git-backed skills команда умеет rehydrate отсутствующие `.speckeep/skills/checkouts/<id>` из данных manifest (`location` + `ref`) перед установкой. Это помогает, если checkout был удален локально. Если исходный git source недоступен, одного manifest недостаточно для восстановления содержимого.

Важные флаги:

- `--dry-run` показывает pending changes без записи на диск
- `--json` выводит результат установки в JSON
- `--include-disabled` ставит и disabled skills

Эквивалентная subcommand:

```bash
speckeep skills install my-project
```

### `speckeep skills-restore [path]`

Восстанавливает git-backed checkouts в `.speckeep/skills/checkouts/` по данным из `.speckeep/skills/manifest.yaml`, не устанавливая skill'ы в agent folders.

Полезно, если checkout'ы были удалены локально, но `manifest.yaml` сохранил `location` и pinned `ref`.

Если upstream git source недоступен, команда не сможет восстановить содержимое только по manifest.

Для machine-readable output используйте `--json`.

Эквивалентная subcommand:

```bash
speckeep skills restore my-project
```

### `speckeep sync-skills [path]`

Синхронизирует только skill-managed артефакты:

- `.speckeep/skills/manifest.yaml`
- managed block в корневом `.gitignore` для `.speckeep/skills/checkouts/`
- managed SpecKeep block в `AGENTS.md` (включая секцию skills)

Важные флаги:

- `--dry-run` показывает pending changes без записи на диск
- `--json` выводит результат синхронизации в JSON

Эквивалентная subcommand:

```bash
speckeep skills sync my-project
```

## Миграция Существующих Feature Packages

Если в активных feature folders у вас ещё лежат `summary.md`, `spec.digest.md` или `plan.digest.md`:

- их можно временно оставить; текущие prompts и checks по умолчанию их игнорируют
- считайте `tasks.md` главным operational entrypoint для `implement` и `verify`
- важный recap-контекст лучше перенести в `tasks.md` `## Implementation Context`
- `.speckeep/constitution.summary.md` стоит сохранить как компактный policy layer
- перед нормализацией старых feature packages сначала выполните `speckeep refresh . --dry-run`, чтобы увидеть managed changes без записи

### `speckeep doctor [path]`

Проверяет здоровье workspace.

`doctor` выводит:

- `error` для отсутствующих обязательных файлов и невалидных значений config
- `warning` для orphaned agent artifacts, которые все еще лежат на диске
- `warning` для нестандартных имен веток Git
- `ok`, когда workspace выглядит здоровым

Используй `--json`, если нужен machine-readable output для automation и CI.

### `speckeep dashboard [path]`

Отображает визуальный дашборд всех активных фич в проекте.

Дашборд включает:

- Слаг фичи
- Текущую фазу workflow
- Процент прогресса реализации
- Статус (READY/BLOCKED)
- Текущую ветку Git (с пометкой `!!` при несоответствии слагу фичи)

```bash
speckeep dashboard
```

### `speckeep feature <slug> [path]`

Показывает подробную workflow-карточку одной фичи.

Текстовый вывод включает:

- текущую фазу и `ready_for`
- статус inspect и verify, если отчеты существуют
- прогресс задач, если существует `tasks.md`
- сгруппированные workflow-findings
- короткую подсказку `focus` о наиболее вероятном следующем действии

Используй `--json`, чтобы получить структурированное состояние и feature-local findings.

### `speckeep feature repair <slug> [path]`

Исправляет безопасные feature-local проблемы SpecKeep.

Сейчас repair умеет:

- переносить flat spec artifacts (`specs/<slug>.md`) в канонический directory layout (`specs/active/<slug>/spec.md`)
- переносить plan artifacts из старого layout `plans/<slug>/` в `specs/active/<slug>/`

Используй `--dry-run`, чтобы посмотреть изменения без применения, и `--json` для структурированного вывода.

### `speckeep features [path]`

Показывает workflow-состояние по всем найденным фичам.

Текстовый вывод суммирует:

- фазу и `ready_for`
- verdict для inspect и verify
- прогресс задач
- сгруппированные issue counts
- наличие артефактов

Используй `--json`, если нужен machine-readable output.

### `speckeep migrate [path]`

Запускает безопасные project-wide миграции SpecKeep.

Сейчас основная область миграции — каноникализация legacy inspect reports по всему проекту.

### `speckeep list-specs [path]`

Показывает список spec slug'ов из `specs_dir/` (по умолчанию: `specs/active/`).

### `speckeep show-spec <name> [path]`

Печатает одну спецификацию по slug.

### `speckeep check <slug> [path]`

Показывает готовность одной фичи и точное следующее действие.

Вывод включает наличие артефактов, вердикт inspect и verify, прогресс задач, точную следующую slash-команду и компактную сводку structured checks, если phase-specific readiness checks уже дали категоризированные findings.

Используй `--all`, чтобы проверить все фичи одной таблицей. Выходит с кодом 1, если хоть одна фича заблокирована.
Используй `--json` для машинно-читаемого вывода в CI, включая `check_summary` и `check_findings`, когда они доступны.

```bash
speckeep check export-report
speckeep check export-report my-project --json
speckeep check my-project --all
speckeep check my-project --all --json
```

### `speckeep trace [slug] [path]`

Сканирует кодовую базу на наличие аннотаций прослеживаемости (traceability).

Форматы аннотаций:
- `// @sk-task <TASK_ID>: <Описание> (<AC_ID>)` для кода реализации.
- `// @sk-test <TASK_ID>: <НазваниеТеста> (<AC_ID>)` для тестовых доказательств.

Правило размещения:
- trace-маркер должен стоять над конкретным owning declaration или behavior block, а не на уровне файла.
- не ставьте его над `package`, `import`, file-header comment или другой строкой, которая не принадлежит конкретной функции/методу/типу/тесту.
- если одну задачу подтверждают несколько тестов/кейсов, `@sk-test <slug>#<TASK_ID>` нужен на каждом таком тесте/кейсе.
- ориентир по стеку: Go `//` над `func/type/Test...`; Python `#` первой строкой внутри `def/class/test_*`; JS/TS `//` над declaration, а для `test()/it()` первой строкой внутри callback; Java/C#/C/C++ comment над method/class/test block.

Эта команда находит связи между кодом реализации, ID задач из `tasks.md` и критериями приемки из `spec.md`.

Используй `slug`, чтобы отфильтровать находки для конкретной фичи.
Используй `--tests`, чтобы показать только тестовые доказательства.
Используй `--json` для машинно-читаемого вывода.

```bash
speckeep trace
speckeep trace export-report
speckeep trace export-report --tests
speckeep trace export-report my-project --json
```

### `speckeep demo [path]`

Создаёт демо-workspace по указанному пути (по умолчанию: `./speckeep-demo`).

Workspace заполнен примером фичи (`export-report`) на фазе implement — spec, inspect report, plan, tasks и data model уже присутствуют. После создания предлагает попробовать `/speckeep.scope`, `/speckeep.challenge` и `/speckeep.handoff`.

```bash
speckeep demo
speckeep demo ./my-demo --agents claude
```

### `speckeep export <slug> [path]`

Упаковывает все артефакты одной фичи в один markdown-документ.

Читает и конкатенирует: spec, inspect report, plan, tasks, data model, research, challenge report и verify report (пропускает отсутствующие файлы). Удобно для передачи полного контекста фичи ревьюеру или новой агентской сессии.

Используй `--output <file>`, чтобы записать в файл вместо stdout.

```bash
speckeep export export-report
speckeep export export-report my-project --output export-report-bundle.md
```

### `speckeep list-archive [path]`

Выводит архивированные фичи из `archive_dir/` (по умолчанию: `specs/archived/`).

Показывает одну запись на slug (последний снимок) со статусом, датой архивации и причиной. Записи отсортированы по дате по убыванию. Статусы выделены цветом: `completed` — зелёный, `deferred` — жёлтый, `abandoned` и `rejected` — красный.

Флаги:

- `--status` — фильтровать по статусу: `completed`, `superseded`, `abandoned`, `rejected`, `deferred`
- `--since <YYYY-MM-DD>` — показать архивы начиная с указанной даты
- `--json` — вывод в JSON для автоматизации и CI

```bash
speckeep list-archive
speckeep list-archive my-project --status deferred
speckeep list-archive my-project --since 2026-01-01
speckeep list-archive my-project --json
```
