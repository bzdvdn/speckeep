# CLI

## Установка

SpecKeep распространяется как один бинарник через GitHub Releases.

Linux:

```bash
VERSION=v0.5.1
curl -fsSL "https://raw.githubusercontent.com/bzdvdn/speckeep/${VERSION}/scripts/install.sh" | bash -s -- --version "${VERSION}"
```

Windows (PowerShell):

```powershell
$version="v0.5.1"
$env:SPECKEEP_VERSION=$version
powershell -ExecutionPolicy Bypass -c "iwr -useb https://raw.githubusercontent.com/bzdvdn/speckeep/$version/scripts/install.ps1 | iex"
```

Чтобы также добавить папку установки в `PATH`:

- Linux: добавь `--add-to-path` или установи `SPECKEEP_ADD_TO_PATH=1`
- Windows: установи `$env:SPECKEEP_ADD_TO_PATH=1` или запускай скрипт с `-AddToPath`

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

- `.speckeep/spk.yaml`
- `.speckeep/templates/**`
- `.speckeep/scripts/**`
- project-local agent skill packs
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

Показывает включенные agent targets из `.speckeep/spk.yaml`.

### `speckeep remove-agent [path]`

Отключает один или несколько agent targets и удаляет их generated files.

### `speckeep cleanup-agents [path]`

Удаляет осиротевшие agent artifacts, которые больше не соответствуют включенным targets в config.

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

Читает записи `Proof:` из `tasks.md` и связывает их с ID задач и критериями приемки из `spec.md`. Исходный код больше не сканируется.

Формат записи `Proof:` для каждой завершенной задачи (`[x]`) в `tasks.md`:

- `Proof: <kind> <path> [<anchor>]`, где `kind` — одно из `code|test|docs|chore`, `path` — путь относительно корня репозитория, `anchor` — owning-функция/тест/тип (опционален, но рекомендуется).

Примеры:

```text
Proof: code src/handlers/export.go ExportHandler
Proof: test src/tests/export_test.go TestExportFlow
Proof: docs docs/export.md
```

Команда отчитывается об осиротевших/дублирующихся/отсутствующих записях `Proof:`, об отсутствующих файлах и предупреждает о нерезолвящихся anchor'ах.

Используй `slug`, чтобы отфильтровать трассировку для конкретной фичи.
Используй `--tests`, чтобы показать только записи `Proof:` с `kind = test`.
Используй `--json` для машинно-читаемого вывода.

```bash
speckeep trace
speckeep trace export-report
speckeep trace export-report --tests
speckeep trace export-report my-project --json
```

### `speckeep converge <slug> [path]`

Дешёвый цикл закрытия для «почти готовых» фич: подтверждает, что каждая завершённая задача несёт валидный `Proof:` (существующий файл, резолвящийся anchor), каждый touched surface существует, а покрытие `AC-*` держится.

Это **легче verify** — файл отчёта не создаётся. Когда остаются разрывы, команда выдаёт список структурированных находок и выходит с кодом 1; агент добавляет follow-up задачи (`## Converge Follow-ups` в `tasks.md`) и повторяет до `converged`, с жёсткой остановкой после 2 раундов правок.

Используй `--json` для машинно-читаемого вывода. Код выхода 1 — фича не сошлась.

```bash
speckeep converge export-report
speckeep converge export-report my-project --json
```

### `speckeep guard [path]`

Детерминированный CI-гейт для закрытия фич: падает (exit 1), когда хотя бы одна активная фича не готова к архивации — открытые задачи, отсутствующие `Proof:`, заблокированный inspect/verify или mismatch ветки.

Передавай `--slug <slug>`, чтобы ограничить проверку одной фичей. Используй `--json` для логов CI.

```bash
speckeep guard my-project
speckeep guard my-project --slug export-report
speckeep guard my-project --json
```

### `speckeep import <openspec|speckit> [path]`

Мигрирует feature packages из другой spec-системы в текущий speckeep workspace.

- **openspec**: читает `openspec/changes/<slug>/` и пересобирает `spec.md` (`RQ-*`/`AC-*` из блоков `### Requirement:` и `#### Scenario:`), `plan.md` из `design.md`, копирует `tasks.md` best-effort (с пометкой перегенерировать через `/spk.tasks`).
- **speckit**: читает `specs/<slug>/` и копирует `spec.md` / `plan.md` / `tasks.md`.

Существующие фичи speckeep никогда не перезаписываются — они сообщаются как пропущенные. Используй `--json` для машинно-читаемого вывода.

```bash
speckeep import openspec ./
speckeep import speckit ./spec-kit-project
speckeep import openspec . --json
```

### `speckeep self check` / `speckeep self upgrade`

Управление установленным бинарником.

- `speckeep self check`: сообщает установленную версию против последнего GitHub-релиза (`--json` для машинного вывода). Read-only — завершается с 0.
- `speckeep self upgrade`: скачивает архив последнего релиза, проверяет sha256 по `sha256sum.txt` и заменяет запущенный бинарник на месте. На Windows или при недоступной записи в директорию установки выдаёт подсказку по ручной установке.

```bash
speckeep self check
speckeep self check --json
speckeep self upgrade
```

### `speckeep demo [path]`

Создаёт демо-workspace по указанному пути (по умолчанию: `./speckeep-demo`).

Workspace заполнен примером фичи (`export-report`) на фазе implement — spec, inspect report, plan, tasks и data model уже присутствуют. После создания предлагает попробовать `/spk.scope`, `/spk.challenge` и `/spk.handoff`.

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
