# Примеры

На этой странице собраны реалистичные end-to-end сценарии SpecKeep для одного feature package.

## Быстрые Сценарии Использования

### Новый проект

Когда проект создается с нуля, SpecKeep лучше вводить как минимальный каркас проектного контекста с самого начала.

Пример:

```bash
speckeep init my-project --lang ru --shell sh --agents codex
cd my-project
speckeep doctor .
```

Что делать дальше:

- сформировать `constitution` для правил проекта
- описать первую фичу через `spec`
- подготовить `plan` и `tasks`
- выполнять `implement` только от текущего task list

Практический смысл такого старта:

- команда и агент сразу работают от одного набора правил
- контекст проекта с самого начала остается явным и редактируемым
- workflow остается легким, потому что SpecKeep не требует тяжелого process engine

### Уже существующий проект

Для brownfield-проекта SpecKeep лучше вводить постепенно, а не пытаться сразу описать всю кодовую базу.

Пример:

```bash
cd existing-project
speckeep init . --lang ru --shell sh --agents codex
speckeep doctor .
```

Рекомендуемый старт:

- сначала зафиксировать `constitution` под текущую реальность проекта
- выбрать одну активную фичу или change request
- создать spec только для нее
- переходить к plan, tasks и implement только внутри этой feature scope

Чего не стоит делать:

- не пытаться сразу документировать весь проект
- не тянуть широкий repository context, если текущая фича этого не требует

Практический смысл такого входа:

- SpecKeep добавляет легкий слой дисциплины поверх уже существующей кодовой базы
- adoption идет по одной фиче за раз
- это снижает токеноемкость и уменьшает риск бюрократии

### Вход Через Prompt-Файл

Когда `/spk.spec` запускается от локального prompt-файла, лучше использовать явные метаданные, а не полагаться на generic filename вроде `spec_prompt.md`.

Пример prompt-файла:

```text
name: Add dark mode
slug: add-dark-mode

Add a user-selectable dark theme for the dashboard and settings pages.
```

Это позволяет SpecKeep:

- вывести безопасный путь спецификации вроде `specs/active/add-dark-mode/spec.md`
- создать или переключить `feature/add-dark-mode`
- избежать неоднозначных slug из generic filename

### Поэтапный Ввод Через `--name`

Когда имя фичи уже понятно, а подробное описание удобнее прислать следующим сообщением, `/spk.spec` может стартовать в staged mode.

Пример:

```text
/spk.spec --name "Dependency Dashboard"
```

Следующее сообщение:

```text
Нужен dashboard для мониторинга зависимостей микросервисов с тёмной темой, фильтрами, dependency graph, summary cards и auto-refresh.
```

Это позволяет SpecKeep:

- зафиксировать каноническое имя фичи заранее
- безопасно вывести slug вроде `dependency-dashboard`
- не терять контекст spec-запроса между сообщениями

Если нужен явный slug:

```text
/spk.spec --name "Dependency Dashboard" --slug frontend-layout-rework
```

Если нужен repository-specific branch override:

```text
/spk.spec --name "Dependency Dashboard" --slug frontend-layout-rework --branch FEAT-142
```

## 1. Создание Конституции для Brownfield-проекта

Пример запроса:

```text
/spk.constitution Python-проект в стиле DDD, разделен на API и workers, Kafka для асинхронной интеграции, ClickHouse как аналитический sink.
```

Ожидаемое поведение агента:

- прочитать prompt `.speckeep/templates/prompts/constitution.md`
- собрать только минимально нужные evidence из репозитория
- создать или обновить `CONSTITUTION.md`
- при необходимости запустить `check-constitution.sh`

Ожидаемый результат:

- архитектурные правила формализованы
- правила разработки зафиксированы явно
- конституция становится главным документом для следующих фаз

## 2. Создание Spec

Пример запроса:

```text
/spk.spec Добавить partner-specific расписание ingestion с override для retry policy.
```

Ожидаемое поведение агента:

- сначала прочитать constitution
- создать `specs/active/partner-scheduling/spec.md`
- записать acceptance criteria в каноническом формате `Given / When / Then`
- остальной текст держать на configured documentation language

Пример acceptance criterion:

```md
### Acceptance Criterion 1

- ID: AC-001
- **Given** у партнера задана собственная retry policy
- **When** рассчитывается расписание ingestion
- **Then** worker использует partner-specific retry window вместо default policy
```

Пример с явным branch override:

```text
/spk.spec Добавить partner-specific расписание ingestion с override для retry policy --branch NRD-11
```

В этом случае slug спецификации может по-прежнему оставаться `partner-scheduling`, а рабочая ветка будет следовать branch convention репозитория, например `NRD-11`.

## 3. Проверка Spec через Inspect

Используйте этот шаг, когда фича неоднозначная, рискованная или нужен формальный quality gate.  
Если спецификация уже ясная и низкорисковая, можно сразу переходить к `/spk.plan <slug>`.

Пример запроса:

```text
/spk.inspect partner-scheduling
```

Ожидаемое поведение агента:

- прочитать constitution и `specs/active/partner-scheduling/spec.md`
- держать default inspect scope дешевым: сначала `CONSTITUTION.md` и `spec.md`, а `plan.md` или `tasks.md` подтягивать только если они существуют и реально влияют на вывод
- проверить полноту, соответствие конституции и качество сценариев
- выпустить focused inspection report
- использовать `.speckeep/scripts/inspect-spec.sh` или `.speckeep/scripts/inspect-spec.ps1` как дешевый helper первого прохода, когда нужно быстро подтвердить структурные проблемы spec или coverage
- сохранять inspect report в `specs/active/partner-scheduling/inspect.md`
- использовать `.speckeep/templates/inspect-report.md` как канонический шаблон отчета

Типовые находки:

- отсутствует failure-path сценарий
- непонятно покрытие для manual retry overrides
- есть открытый вопрос про ownership scheduler logic

## 4. Создание Plan Package

Пример запроса:

```text
/spk.plan partner-scheduling
```

Ожидаемое поведение агента:

- прочитать constitution и spec
- если есть `specs/active/partner-scheduling/inspect.md`, проверить, что статус неблокирующий (`pass` или `concerns`)
- создать `specs/active/partner-scheduling/plan.md`
- создать `specs/active/partner-scheduling/data-model.md`
- создать `specs/active/partner-scheduling/contracts/`
- создавать `research.md` только если действительно есть неопределенность

Типовые выходы:

- plan по integration points для scheduler
- data model для partner overrides и retry windows
- event или API contracts для обновления конфигурации

## 5. Создание Tasks

Пример запроса:

```text
/spk.tasks partner-scheduling
```

Ожидаемое поведение агента:

- использовать `plan.md` как decomposition entrypoint
- подтягивать spec, contracts или data model только при необходимости
- создать `specs/active/partner-scheduling/tasks.md`
- включить acceptance-to-task coverage

Пример структуры задач:

```md
## Phase 1: Data Model

- [ ] T1.1 Add partner scheduling override model — override fields are persisted
- [ ] T1.2 Persist retry window fields — retry windows are available to scheduling logic

## Acceptance Coverage

- AC-001 -> T1.1, T1.2
```

## 6. Реализация Фичи

Пример запроса:

```text
/spk.implement partner-scheduling
```

Ожидаемое поведение агента:

- прочитать `tasks.md` и использовать его как манифест выполнения
- выполнять **In-place Декомпозицию**, если задача слишком сложная, добавляя вложенные подзадачи (напр., `T1.1.1`)
- фиксировать доказательство выполненной задачи записью `Proof:` в `tasks.md`
- отмечать завершенные задачи в `tasks.md`
- оставаться в рамках списка `Touches:`, определенного для каждой задачи

Пример записей `Proof:` в `tasks.md` под завершенными задачами:

```text
- [x] T1.1 Add partner scheduling override model — override fields are persisted
    Proof: code src/models/partner.go SavePartnerSchedule
    Proof: test src/tests/partner_test.go TestSaveOverride

- [x] T1.2 Persist retry window fields — retry windows are available to scheduling logic
    Proof: code src/storage/retry.go PersistRetryWindow
    Proof: docs docs/scheduling.md
```

Формат записи: `Proof: <kind> <path> [<anchor>]`, где `kind` — одно из `code|test|docs|chore`, `path` — путь относительно корня репозитория, `anchor` — owning-функция/тест/тип (опционален, но рекомендуется). Задача `[x]` без хотя бы одной записи `Proof:` не считается выполненной.

## 7. Верификация реализации

`verify` — это **опциональный аудит по требованию**: всегда доступен, но по умолчанию пропускается. Он запускается как audit + отчет на уровне AC (вторая оценка). Готовность фичи к архивации можно подтвердить и детерминированно — через `speckeep check` или `speckeep archive`.

Пример запроса:

```text
/spk.verify partner-scheduling
```

Ожидаемое поведение агента:

- использовать `.speckeep/scripts/trace.sh partner-scheduling` для сбора доказательств реализации — скрипт читает записи `Proof:` из `tasks.md`
- подтверждать соответствие реализации описанию задач и критериям приемки
- предоставить четкий вердикт (`pass`, `concerns` или `blocked`)
- включить конкретные доказательства в секцию `## Checks`
- если одной задаче соответствует несколько файлов/тестов, ожидать отдельную запись `Proof:` для каждого доказательства

Проверка проходит по записям `Proof:` из `tasks.md`, без сканирования исходного кода на аннотации.

Ожидаемое поведение агента:

- стартовать от `tasks.md`
- читать spec, plan, data model или contracts только для активной задачи
- выполнять незавершенные задачи по порядку
- сообщать phase progress по мере движения по выбранному scope
- обновлять `tasks.md`

Эта фаза не должна читать широкий контекст репозитория без реальной необходимости.

Примеры выборочных запросов:

```text
/spk.implement partner-scheduling --phase 2
/spk.implement partner-scheduling --tasks T1.1,T2.1
```

Ожидаемое поведение в scoped mode:

- сохранять full-run behavior только когда scope-флагов нет
- выполнять только выбранную фазу или выбранные task IDs, если scope явно сужен
- сохранять порядок задач из `tasks.md`
- предупреждать, если выбранная работа перескакивает через незавершенные более ранние фазы или задачи

Типичные runtime updates:

- `Начинаю Фазу 1: Модель данных`
- `Фаза 1 завершена: T1.1, T1.2`
- `Дальше: Фаза 2: Логика планировщика`

## 7. Verify Фичи

`verify` — опциональный аудит по требованию; по умолчанию фаза пропускается.

Пример запроса:

```text
/spk.verify partner-scheduling
```

Ожидаемое поведение агента:

- сначала прочитать constitution и tasks
- подтвердить, что завершенные задачи достаточно соответствуют текущему состоянию реализации
- выпустить легкий verification report
- начинать с `.speckeep/scripts/verify-task-state.sh partner-scheduling`, если сначала нужно только подтвердить состояние задач
- использовать `.speckeep/templates/verify-report.md`, если отчет нужно сохранить в файл
- по умолчанию использовать `specs/active/partner-scheduling/verify.md`, если путь явно не указан

## 8. Архивация Фичи

CLI-шаг, разрешенный при детерминированной доказанности фичи (все задачи `[x]` имеют записи `Proof:`) или после `verify: pass`:

```bash
speckeep archive partner-scheduling .
```

Ожидаемое поведение CLI:

- для статуса `completed` провалидировать prerequisites verify/tasks и остановиться с понятной ошибкой, если открытые задачи остались или у задач `[x]` нет записей `Proof:`
- скопировать feature package в `specs/archived/partner-scheduling/<YYYY-MM-DD>/`
- записать `summary.md`

Ожидаемый результат архива:

```text
specs/
  archived/
    partner-scheduling/
      2026-03-28/
        summary.md
        spec.md
        plan.md
        tasks.md
        data-model.md
        contracts/
```

## 9. Сценарий Обслуживания Агентов

Практический maintenance flow для agent targets:

```bash
speckeep add-agent my-project --agents claude --agents cursor
speckeep list-agents my-project
speckeep remove-agent my-project --agents cursor
speckeep cleanup-agents my-project
speckeep doctor my-project
```

Этот сценарий полезен, когда проект со временем меняет предпочитаемый набор агентов.
