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
- **Traceability** (записи `Proof:` в `tasks.md`) — доказательство, что каждое требование реализовано и протестировано
- **19 адаптеров агентов** — Claude Code, Codex, Cursor, Copilot, OpenCode, aider, Amazon Q, Gemini, Jules, Cline, Devin, Goose, Refact, Windsurf и другие ([уровни поддержки](docs/ru/agents.md#уровни-поддержки)) — поддержка Continue.dev снята после сообщений о свёртывании продукта
- **Быстрый цикл закрытия** (`/spk.converge`) — превращает разрывы реализации в follow-up задачи и повторяет до сходимости
- **CI-гейт** (`speckeep guard`) — машино-проверяемый ответ «можно ли прямо сейчас закрыть каждую фичу?»
- **Express-полоса** — крошечные/низкорисковые фичи пропускают `plan.md`/`data-model.md` и закрываются через `spec.md` + `tasks.md`
- **One-shot propose** (`/spk.propose`) — идея → `spec.md` + `tasks.md` за один проход, сразу к implement
- **Миграция на лету** (`speckeep import openspec|speckit`) — конвертация feature packages из OpenSpec/Spec Kit в наш layout за секунды
- **Компактный архив** (`--compact`) — хранит только `summary.md` + git-указатель вместо копий всех артефактов

Результат на практике: агенты реже ошибаются с первого раза, хендоффы между сессиями требуют меньше контекста, а требования остаются читаемыми для человека.

---

## Workflow

```
constitution → spec → [inspect] → plan → tasks → implement → archive
```
`verify` — опциональный on-demand аудит; archive разрешён, когда каждая задача `[x]` имеет `Proof:` (или после `verify: pass`).

Каждая фаза загружает только минимум контекста. Опциональные команды на любой фазе: `/spk.challenge`, `/spk.handoff`, `/spk.hotfix`, `/spk.scope`, `/spk.recap`.

---

## Позиционирование

| Dimension             | speckeep                            | OpenSpec                  | Spec Kit                |
| --------------------- | ----------------------------------- | ------------------------- | ----------------------- |
| Стиль workflow        | Строгая цепочка фаз, узкий контекст | Гибкий, вокруг артефактов | Полный многошаговый SDD |
| Контекст по умолчанию | Минимальный                         | Средний                   | Максимальный            |
| Накладные расходы     | Низкие                              | Средние                   | Высокие                 |
| Brownfield            | Высокие                             | Высокие                   | Средние                 |
| Коллаборация          | Branch-first, feature-local         | Change-folder oriented    | Тяжёлые ветки           |
| Подходит для          | Лёгкий строгий SDD                  | Гибкий SDD-lite           | Полноценный строгий SDD |

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
speckeep converge <slug> [path] [--json]
speckeep guard [path] [--slug <slug>] [--json]
speckeep import <openspec|speckit> [path] [--json]
speckeep demo [path]
speckeep archive <slug> [path] [--compact]
speckeep list-archive [path] [--status <status>] [--since <YYYY-MM-DD>] [--json]
speckeep self check | self upgrade
speckeep migrate [path]
speckeep add-agent | list-agents | remove-agent | cleanup-agents [path]
```

Агентская поддержка **skills-first**: `speckeep init --agents opencode,claude` раскладывает композитный SpecKeep `sdd` skill-pack (корневой `SKILL.md` плюс самодостаточные скиллы по фазам, каждый со своими полными инструкциями) в стандартную skills-директорию таргета (`.opencode/skills/`, `.claude/skills/`, …). CLI остаётся детерминированным остовом, который скиллы зовут как гейты (`check`, `converge`, `guard`).

---

## Установка

**Linux / macOS:**

```bash
curl -fsSL "https://raw.githubusercontent.com/bzdvdn/speckeep/main/scripts/install.sh" | bash
# --version v1.0.0 для пина версии; --add-to-path для регистрации PATH
```

**Windows (PowerShell):**

```powershell
powershell -ExecutionPolicy Bypass -c "iwr -useb https://raw.githubusercontent.com/bzdvdn/speckeep/main/scripts/install.ps1 | iex"
```

На Windows бинарник добавляется в PATH автоматически. На Linux используйте `--add-to-path`, если директория установки ещё не в вашем PATH.

**Пакетные менеджеры:**

- Homebrew: `brew install bzdvdn/speckeep/speckeep` (см. `contrib/packaging/brew/`)
- Scoop: `scoop bucket add speckeep https://github.com/bzdvdn/speckeep && scoop install speckeep` (см. `contrib/packaging/scoop/`)
- npm: `npx speckeep init` или `npm install -g speckeep` — тонкий launcher, который при установке скачивает подходящий нативный бинарник (см. `contrib/packaging/npm/`)
- Go: `go install speckeep@latest`

**Обновление установленного бинарника:**

```bash
speckeep self check     # текущая vs последняя версия (read-only)
speckeep self upgrade   # скачивание, sha256-проверка и замена на месте
```

**CI:**

```yaml
- uses: bzdvdn/speckeep/.github/actions/speckeep@main
  with:
    root: .
```

Экшен ставит speckeep, находит изменившиеся `specs/active/<slug>/` фичи относительно базы PR и прогоняет `speckeep guard` по каждой — PR падает, если затронутая фича не готова к закрытию. Готовый workflow — в [`contrib/ci/speckeep-guard.yml`](contrib/ci/speckeep-guard.yml).

**Сборка из исходников:**

```bash
go build -ldflags "-X speckeep/src/internal/cli.Version=v0.8.0" -o bin/speckeep ./src/cmd/speckeep
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

| Surface                    | Задачи |
| -------------------------- | ------ |
| hooks/useReportExport.ts   | T1.1   |
| components/ReportsPage.tsx | T1.2   |
| tests/reports.test.ts      | T2.1   |

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

Во время реализации фиксируйте доказательство каждой завершённой задачи строкой `Proof:` сразу под чекбоксом `[x]` в `tasks.md`:

```
- [x] T1.1 Добавить обработчик экспорта
  Proof: code src/handlers/export.go ExportHandler
  Proof: test src/tests/export_test.go TestExportFlow
```

`Proof: <kind> <path> [<anchor>]`, где `kind` — `code|test|docs|chore`. Отмеченная задача без записи `Proof:` не завершена — `speckeep check` и `speckeep archive` блокируют её. Evidence читается только из `tasks.md`; trace-маркеров в исходном коде нет.

Проверьте:

```bash
speckeep trace <slug> .
```

### Адаптеры агентов (skills-first)

Поддерживаются из коробки: `claude`, `codex`, `copilot`, `cursor`, `kilocode`, `opencode`, `trae`, `windsurf`, `roocode`, `aider`, `amazonq`, `gemini`, `jules`, `cline`, `devin`, `goose`, `refact`, `codiumate`, `qwen-code`.

```bash
speckeep init my-project --agents opencode,claude    # раскладка sdd skill-pack
```

Скиллы лежат в skills-директории таргета: лёгкий обзорный скилл `sdd` плюс по одному независимому, напрямую слэш-вызываемому скиллу на фазу:

```text
.<target>/skills/
  sdd/
    SKILL.md          # обзор: цепочка workflow, гейты — сюда, когда фаза не очевидна
  spk-spec/
    SKILL.md           # каждая фаза — свой top-level скилл: /spk-spec, /spk-plan, /spk-implement, ...
  spk-plan/
    SKILL.md
  ...
```

Каждый фазовый скилл — своя директория (не вложенный файл-ресурс), поэтому он напрямую слэш-вызываемый — например, набрать `/spk-spec` в Claude Code — а не доступен только через решение модели открыть связанный файл. Каждый инлайнит канонический промпт из `.speckeep/templates/prompts/` (синхронизируется автоматически), так что агент получает полные инструкции фазы из одного файла, и гейтится через `speckeep check`; закрытие — через `speckeep converge` / `speckeep guard`. Для `aider` дополнительно генерируется `.aider/CONVENTIONS.md`-указатель, т.к. у него нет загрузчика скиллов.

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
