# speckeep

[![ci](https://github.com/bzdvdn/speckeep/actions/workflows/ci.yml/badge.svg)](https://github.com/bzdvdn/speckeep/actions/workflows/ci.yml)
[![release-build](https://github.com/bzdvdn/speckeep/actions/workflows/release-build.yml/badge.svg)](https://github.com/bzdvdn/speckeep/actions/workflows/release-build.yml)

[English README](README.md)

`speckeep` — лёгкий Spec-Driven Development kit для агентов разработки и людей. Хранит specs, планы, задачи и traceability в простых файлах, чтобы агенты и люди работали согласованно без тяжеловесного процесса.

SpecKeep — преемник DraftSpec (архивирован). Миграция: `speckeep migrate`.

---

## Быстрый старт — 30 секунд

```bash
# 1. Попробовать сразу — без настройки проекта
speckeep demo ./my-demo

# 2. Посмотреть что создалось
speckeep dashboard ./my-demo

# 3. Инициализировать реальный проект
speckeep init my-project --lang ru --shell sh --agents claude
```

Готово. Проект содержит конституцию, spec, plan, tasks, inspect-отчёт и data model — плюс файлы промптов для агента и `AGENTS.md`.

---

## Зачем speckeep?

Команды, использующие AI-агентов, сталкиваются с общей проблемой: агенты теряют контекст между сессиями, отходят от требований и пишут непроверяемый код.

speckeep решает это через **discipline per token** — минимальную файловую структуру, которая держит агентов и людей в одном контексте:

- **Specs** со стабильными ID (`AC-*`, `RQ-*`) — агенты точно знают, что строить и проверять
- **Tasks** с surface map и группировкой по фазам — агенты выполняют по порядку, одну фазу за раз
- **Traceability** (`@sk-task` аннотации) — проверка, что каждое требование реализовано и протестировано
- **10 адаптеров агентов** — Claude Code, Cursor, Copilot, OpenCode, aider, Windsurf и другие

Результат на практике: агенты реже ошибаются с первого раза, хендоффы между сессиями требуют меньше контекста, а требования остаются читаемыми для человека.

---

## Workflow

```
constitution → spec → [inspect] → plan → tasks → implement → verify → archive
```

Каждая фаза загружает только минимум контекста. Опциональные команды на любой фазе: `/spk.challenge`, `/spk.handoff`, `/spk.hotfix`, `/spk.scope`, `/spk.recap`.

---

## Позиционирование

| Dimension | speckeep | OpenSpec | Spec Kit |
|---|---|---|---|
| Стиль workflow | Строгая цепочка фаз, узкий контекст | Гибкий, вокруг артефактов | Полный многошаговый SDD |
| Контекст по умолчанию | Минимальный | Средний | Максимальный |
| Накладные расходы | Низкие | Средние | Высокие |
| Brownfield | Высокие | Высокие | Средние |
| Коллаборация | Branch-first, feature-local | Change-folder oriented | Тяжёлые ветки |
| Подходит для | Лёгкий строгий SDD | Гибкий SDD-lite | Полноценный строгий SDD |

---

## CLI

```text
speckeep init [path]
speckeep refresh [path]
speckeep doctor [path] [--json]
speckeep dashboard [path]
speckeep check <slug> [path] [--json]
speckeep check [path] --all [--json]
speckeep feature <slug> [path]
speckeep features [path]
speckeep list-specs [path]
speckeep show-spec <name> [path]
speckeep trace <slug> [path]
speckeep export <slug> [path] [--output <file>]
speckeep demo [path]
speckeep archive <slug> [path]
speckeep list-archive [path] [--status <status>] [--since <YYYY-MM-DD>] [--json]
speckeep migrate [path]
speckeep add-agent | list-agents | remove-agent | cleanup-agents [path]
speckeep add-skill | list-skills | remove-skill | install-skills | skills-restore [path]
```

---

## Установка

**Linux / macOS:**

```bash
VERSION=v0.5.1
curl -fsSL "https://raw.githubusercontent.com/bzdvdn/speckeep/${VERSION}/scripts/install.sh" | bash -s -- --version "${VERSION}"
```

**Windows (PowerShell):**

```powershell
$version="v0.5.1"
$env:SPECKEEP_VERSION=$version
powershell -ExecutionPolicy Bypass -c "iwr -useb https://raw.githubusercontent.com/bzdvdn/speckeep/$version/scripts/install.ps1 | iex"
```

Добавьте `--add-to-path` (Linux) или `-AddToPath` (Windows) чтобы добавить папку установки в `PATH`.

**Go:**

```bash
go install speckeep@latest
```

**Сборка из исходников:**

```bash
go build -ldflags "-X speckeep/src/internal/cli.Version=v0.5.1" -o bin/speckeep ./src/cmd/speckeep
```

---

## Пример полного цикла фичи

<details>
<summary>Полный workflow: «Добавить экспорт отчётов в CSV» →</summary>

### 1. Инициализация

```bash
speckeep init . --lang ru --shell sh --agents claude
```

### 2. Spec

Вызовите `/spk.spec --name "Экспорт отчётов в CSV"` в агенте.

`specs/active/eksport-otchetov-v-csv/spec.md`:

```markdown
## Цель
Позволить пользователю скачать таблицу отчётов в виде CSV-файла.

## Критерии приемки

**AC-001** Экспорт создаёт файл
Given на странице Отчёты есть хотя бы одна строка
When пользователь нажимает «Экспортировать CSV»
Then скачивается .csv-файл с заголовками и всеми видимыми строками

**AC-002** Пустое состояние обработано
Given таблица отчётов пуста
When пользователь нажимает «Экспортировать CSV»
Then скачивается .csv только с заголовками — без ошибок
```

### 3. Inspect

Вызовите `/spk.inspect eksport-otchetov-v-csv`.

### 4. Plan

Вызовите `/spk.plan eksport-otchetov-v-csv`. Surfaces: `ReportsPage.tsx`, `useReportExport.ts`, `reports.test.ts`.

### 5. Tasks

Вызовите `/spk.tasks eksport-otchetov-v-csv`. Результат:

| Surface | Задачи |
|---|---|
| hooks/useReportExport.ts | T1.1 |
| components/ReportsPage.tsx | T1.2 |
| tests/reports.test.ts | T2.1 |

### 6. Implement, verify, archive

```
/spk.implement eksport-otchetov-v-csv
/spk.verify    eksport-otchetov-v-csv   # вердикт: pass
speckeep archive    eksport-otchetov-v-csv .
```

### Проверить готовность в любой момент

```bash
speckeep check eksport-otchetov-v-csv
# Фаза:   tasks → implement
# Задачи: 0 / 3 выполнено
# Далее:  /spk.implement eksport-otchetov-v-csv
```

</details>

---

## Ключевые концепции

### Артефакты

Каждая фича живёт в `specs/<slug>/`:

- `spec.md` — требования (`RQ-*`) и критерии приемки (`AC-*` с Given/When/Then)
- `inspect.md` (опционально) — quality gate перед planning
- `plan.md` — дизайн-решения (`DEC-*`) и incremental delivery
- `tasks.md` — задачи с surface map и группировкой по фазам
- `data-model.md` — сущности, поля, инварианты
- `contracts/api.md`, `contracts/events.md` (опционально)
- `verify.md` — результаты верификации

### Traceability

Аннотируйте код во время реализации:

```go
// @sk-task T1.1 (AC-001)
```

Проверьте:

```bash
speckeep trace <slug> .
```

### Адаптеры агентов

Поддерживаются из коробки: `claude`, `codex`, `copilot`, `cursor`, `kilocode`, `opencode`, `trae`, `windsurf`, `roocode`, `aider`.

### Skills

Переиспользуемые пакеты guidance из локальных путей или git-репозиториев:

```bash
speckeep add-skill my-project --id architecture --from-local skills/architecture
speckeep install-skills my-project
```

---

## Документация

Расширенная документация в [`docs/`](docs/README.md):

- [Обзор](docs/ru/overview.md)
- [CLI](docs/ru/cli.md)
- [Модель workflow](docs/ru/workflow.md)
- [Архитектура](docs/ru/architecture.md)
- [Агенты](docs/ru/agents.md)
- [Примеры](docs/ru/examples.md)
- [FAQ](docs/ru/faq.md)
- [Глоссарий](docs/ru/glossary.md)
- [Roadmap](docs/ru/roadmap.md)

Проектные документы:

- [Contributing](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Security Policy](SECURITY.md)
- [MVP](MVP.md)
- [Changelog](CHANGELOG.md)

---

## Разработка

Требуется **Go 1.26+**.

```bash
go test ./...
go vet ./...
go build -o bin/speckeep ./src/cmd/speckeep
```

## Лицензия

Проект распространяется по лицензии [MIT](LICENSE).
