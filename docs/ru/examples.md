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

Полный разобранный пример — реальный текст constitution, узко
ограниченный spec на один существующий эндпоинт и конкретный before/after —
см. [Рецепт: Внедряем speckeep в уже существующий проект](#рецепт-внедряем-speckeep-в-уже-существующий-проект).

### Вход Через Prompt-Файл

Когда `/spk-spec` запускается от локального prompt-файла, лучше использовать явные метаданные, а не полагаться на generic filename вроде `spec_prompt.md`.

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

Когда имя фичи уже понятно, а подробное описание удобнее прислать следующим сообщением, `/spk-spec` может стартовать в staged mode.

Пример:

```text
/spk-spec --name "Dependency Dashboard"
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
/spk-spec --name "Dependency Dashboard" --slug frontend-layout-rework
```

Если нужен repository-specific branch override:

```text
/spk-spec --name "Dependency Dashboard" --slug frontend-layout-rework --branch FEAT-142
```

## 1. Создание Конституции для Brownfield-проекта

Пример запроса:

```text
/spk-constitution Python-проект в стиле DDD, разделен на API и workers, Kafka для асинхронной интеграции, ClickHouse как аналитический sink.
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
/spk-spec Добавить partner-specific расписание ingestion с override для retry policy.
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
/spk-spec Добавить partner-specific расписание ingestion с override для retry policy --branch NRD-11
```

В этом случае slug спецификации может по-прежнему оставаться `partner-scheduling`, а рабочая ветка будет следовать branch convention репозитория, например `NRD-11`.

## 3. Проверка Spec через Inspect

Используйте этот шаг, когда фича неоднозначная, рискованная или нужен формальный quality gate.  
Если спецификация уже ясная и низкорисковая, можно сразу переходить к `/spk-plan <slug>`.

Пример запроса:

```text
/spk-inspect partner-scheduling
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
/spk-plan partner-scheduling
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
/spk-tasks partner-scheduling
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
/spk-implement partner-scheduling
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
/spk-verify partner-scheduling
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
/spk-implement partner-scheduling --phase 2
/spk-implement partner-scheduling --tasks T1.1,T2.1
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
/spk-verify partner-scheduling
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

---

## Рецепты

Основной жизненный цикл (`constitution → spec → plan → tasks → implement → archive`)
описан выше. Эти рецепты — про второстепенные команды: когда каждая из них
окупается, конкретный before/after и чего от неё не стоит ждать.

### Рецепт: Внедряем speckeep в уже существующий проект

Какую проблему решает: в brownfield-репозитории уже есть работающий код,
сложившаяся (пусть и неформальная) архитектура, и команда, которая на нём
что-то отгружает. Соблазн — "сделать по-нормальному" и заспекать всё перед
тем, как что-то трогать — это ровно тот failure mode process-bloat, от
которого speckeep и защищает. Правильная единица внедрения — **одна
фича**, а не репозиторий.

Стартовая точка — реальный, уже не идеальный репозиторий, а не пустая
директория:

```bash
cd billing-api   # существующий Express + Postgres API, ~40k LOC, 4 инженера,
                 # никакого спек-процесса раньше не было, запросы на фичи
                 # живут в тикетах Jira и тредах Slack
speckeep init . --lang ru --shell sh --agents claude
```

```text
initialized .speckeep/ workspace
wrote CONSTITUTION.md (шаблон — отредактируйте или запустите /spk-constitution)
wrote AGENTS.md (managed SpecKeep guidance block)
wrote .claude/skills/sdd/ (8 phase skills)
next: speckeep doctor .
```

```bash
speckeep doctor .
```

```text
ok:      .speckeep/ layout present
ok:      AGENTS.md contains managed SpecKeep block
warning: current branch "main" — feature work should happen on feature/<slug>
ok:      claude agent target: skill pack present
```

Здесь ничего не потребовало чтения кодовой базы — `init` только раскладывает
управляемые speckeep файлы. Следующий шаг — тот, что реально важен: не
спекать репозиторий, а заспекать **одну боль**, с которой команда уже
согласна.

**Шаг 1 — constitution отражает текущую реальность, а не устремление.**
Зафиксируйте то, что реально верно сегодня, включая некрасивые части —
constitution, описывающий воображаемую будущую архитектуру, хуже, чем его
отсутствие, потому что агенты будут ему доверять.

```text
/spk-constitution Node.js + Express монолит, PostgreSQL через сырые pg
запросы (без ORM), только REST (без GraphQL), синхронный request/response
(очереди пока нет — это известный пробел, а не целевое состояние). Тесты —
Jest + supertest. Ни одна фича не может вводить новый HTTP-фреймворк или БД
без явного обновления constitution сначала.
```

Итоговый фрагмент `CONSTITUTION.md`:

```md
## Architecture
- Node.js + Express, PostgreSQL через сырые `pg`-запросы (без ORM — так
  задумано).
- Только REST; без GraphQL. Синхронный request/response — очереди сообщений
  пока не существует (известный пробел, а не то, что стоит тихо "починить"
  внутри одной фичи).

## Non-Negotiable Rules
- Никакого нового HTTP-фреймворка или БД без предварительного обновления
  constitution.
- Тесты: Jest + supertest; новый test framework не вводится по ходу дела.
```

**Шаг 2 — берём ту одну фичу, о которой команда уже спорит,** а не обзор
всей поверхности API:

```text
/spk-spec Поддержать idempotency keys на POST /invoices, чтобы повторные
запросы клиента не создавали дублирующиеся инвойсы.
```

Агент читает constitution, ищет существующий обработчик роута `/invoices`
и его тестовый файл — и больше ничего — и пишет `spec.md`, ограниченный
ровно этим эндпоинтом. Он **не** открывает каждый файл роутов "чтобы
понять кодовую базу" сначала — это и есть дисциплина узкого чтения, ради
которой существуют constitution и `AGENTS.md`.

```text
Slug: invoice-idempotency-keys
Status: spec
Artifacts: specs/active/invoice-idempotency-keys/spec.md
Blockers: none
Ready for: /spk-plan invoice-idempotency-keys
```

Отсюда фича проходит по обычной цепочке (`plan → tasks → implement →
archive`, см. walkthrough выше) точно так же, как в greenfield-проекте —
единственное отличие, которое привносит brownfield, — на шагах constitution
и spec, где агент ограничен реальной формой кодовой базы, а не чистым
листом.

**Чего не делать:**

- Не просите `/spk-spec` (или любую фазу) "задокументировать весь
  billing-api сервис целиком" — это repo-аудит, а не feature spec, и он
  породит огромный неревьюабельный артефакт, который никто не будет
  поддерживать в актуальном состоянии.
- Не позволяйте constitution описывать, куда архитектура *движется* —
  описывайте, где она *есть сейчас*. Миграции (например "добавляем
  очередь") принадлежат `plan.md` конкретной фичи, а не constitution, как
  будто это уже правда.
- Не запускайте `/spk-repo-map` в первый же день "чтобы было" — он
  окупается, когда `Touches:`-поверхности фичи перестают быть очевидны из
  существующих имён директорий, а не раньше.

**Before / after:**

| | До speckeep | После speckeep |
| --- | --- | --- |
| Где живёт intent фичи | Описание в Jira + тред в Slack | `specs/active/<slug>/spec.md`, версионируется вместе с кодом |
| Сколько репозитория читает агент | Сколько сам решит, что нужно | Constitution + `Touches:`-поверхности одной фичи |
| Что значит "готово" | "Выглядит правильно в ревью" | У каждой `[x]` задачи есть строка `Proof:`, которую проверяет `speckeep guard` |
| Онбординг нового инженера на фичу | Перечитать PR и поспрашивать вокруг | `/spk-handoff <slug>` или `/spk-recap` |

Если первая же фича уже кажется слишком большой, чтобы заспекать чисто
(затрагивает несколько сервисов, неясное владение, реальный риск дизайна) —
это законный сигнал запустить `/spk-inspect` перед `/spk-plan`, а не повод
защитно расширять spec, чтобы покрыть побольше.

### Рецепт: Alignment Pass спрашивает прежде, чем гадать

Какую проблему решает: агент, который пишет acceptance criteria по
расплывчатому запросу, либо молча выдумывает требования, либо делает спеку,
с которой никто не соглашался. `/spk-spec` сначала прогоняет короткий
**alignment pass** — вопросы задаются только когда запрос реально
неоднозначен, а не для проформы.

Расплывчатый запрос:

```text
/spk-spec Добавь возможность экспортировать данные.
```

Ожидаемое поведение агента — он видит, что не хватает и actor'а, и границы
scope, и задаёт вопросы один раз, одним пакетом, вместо того чтобы гадать:

```text
Прежде чем писать спеку, два момента, которые не вывести из запроса:
1. Кто экспортирует — пользователь из UI, или автоматическая job/API-вызов?
2. Форматы: только CSV, или CSV + JSON/Excel?

Приму "последние 30 дней" как диапазон по умолчанию, если не скажете иначе.
```

Точный запрос — та же команда, ноль уточняющих вопросов, сразу спека:

```text
/spk-spec Дай пользователю дашборда экспортировать видимую таблицу в CSV,
с учётом активных фильтров, кнопкой в правом верхнем углу тулбара.
```

Почему это важно: spec с нулём clarify-раундов — это *хороший* результат, а
не срезанный угол — alignment pass существует для тех ~30% запросов, что
реально неоднозначны, а не чтобы допрашивать каждый. Если вы не отвечаете и
сразу переходите к другой команде, агент пишет черновик best-effort и кладёт
нерешённые пункты в `## Открытые вопросы` вместо того, чтобы выдумать
`AC-*` для заполнения пробела.

### Рецепт: One-shot fast lane через `/spk-propose`

Какую проблему решает: для маленькой низкорисковой правки прогонять
`spec → plan → tasks` тремя отдельными раундами — это process overhead,
который правке не нужен. `/spk-propose` схлопывает идею → `spec.md` +
`tasks.md` (по умолчанию без `plan.md` — **express lane**) за один проход и
сразу передаёт в `/spk-implement`.

```text
/spk-propose Добавь кнопку "copy as JSON" рядом с существующей кнопкой
"copy as CSV" на странице деталей отчёта.
```

Ожидаемый результат:

```text
Slug: copy-as-json
Status: propose
Artifacts: specs/active/copy-as-json/spec.md, specs/active/copy-as-json/tasks.md
Blockers: none
Ready for: /spk-implement copy-as-json
```

Если идея оказывается крупнее, чем выглядела — несколько реалистичных
вариантов реализации, влияние на границы, миграционный риск — propose
останавливается и падает обратно на `/spk-spec`, вместо того чтобы
проталкивать неглубокий план в `tasks.md`. Это строка `Blockers:`, а не
баг: propose намеренно узкий, чтобы никогда не race-drafted фичу, которая
на самом деле требовала дизайна.

### Рецепт: Общий доменный глоссарий через `/spk-glossary`

Какую проблему решает: без общего словаря `spec.md` называет это
"workspace", `plan.md` — "project", а код — "tenant" — три имени для
одного понятия, и каждый читатель платит налог на перевод.
`.speckeep/glossary.md` опционален и живёт на уровне проекта (не per-feature);
он появляется только когда термин реально стоит зафиксировать.

```text
/spk-glossary Определи "workspace" vs "project" — мы используем их
взаимозаменяемо, и это начинает путать в спеках.
```

Итоговый `.speckeep/glossary.md`:

```md
# Glossary

| Term | Definition | Aliases (do not use) | Notes |
| --- | --- | --- | --- |
| `Workspace` | Billing-scoped контейнер, содержащий один или несколько project. | `Account`, `Org` | Создаётся при регистрации; 1:1 с Stripe-клиентом. |
| `Project` | Именованная коллекция specs/features внутри workspace. | `Workspace` (отклонено) | Раньше называлось "workspace" в старых доках — не переиспользуйте это имя для этого понятия. |
```

После появления этого файла `/spk-spec`, `/spk-plan` и `/spk-tasks` читают
его один раз за сессию и переиспользуют термины — они не введут новый
синоним тому, что он уже определяет. Больше ничего не меняется: нет гейта,
нет обязательного перечитывания, нет обязательного обновления на каждой
фиче.

### Рецепт: Поиск слепых зон через `/spk-challenge`

Какую проблему решает: `/spk-spec` и `/spk-inspect` написаны так, чтобы
сходиться к готовому к отгрузке артефакту; `/spk-challenge` — намеренно
адверсариален: он ищет именно то, что кооперативный проход обычно
пропускает (непроверяемые утверждения, тихое расширение scope,
противоречия).

```text
/spk-challenge partner-scheduling
```

Типичные находки:

```text
- AC-002 говорит "retry window обновляется promptly" — "promptly" не
  наблюдаемо. Fix: зафиксировать конкретную границу (например "в течение
  60с") или убрать claim в non-functional заметку.
- spec.md не говорит, что происходит, если у партнёра нет override —
  default policy подразумевается, но нигде не сформулирована. Fix:
  добавить явный AC или строку Assumptions.
- plan.md DEC-002 вводит новый caching layer, не упомянутый в spec.md
  Out of Scope — возможное скрытое расширение scope. Fix: подтвердить
  через spec или вынести из этой фичи.
```

`/spk-challenge` только выдаёт находки + минимальные fix'ы — он не выносит
вердикт `pass|concerns|blocked` (это `/spk-inspect`) и не заменяет scope
inventory (это `/spk-scope`). Запускайте, когда хотите второй, скептичный
взгляд на spec или plan перед тем, как закоммититься к нему.

### Рецепт: Быстрая проверка границ через `/spk-scope`

Какую проблему решает: "входит ли X в scope этой фичи?" — вопрос, на
который иначе отвечают перечитыванием всей спеки. `/spk-scope` даёт
быструю in/out опись без вердикта и без глубокого ревью.

```text
/spk-scope partner-scheduling
```

```text
In scope:
- Per-partner override retry window (AC-001)
- Fallback на default policy, если override отсутствует

Out of scope:
- Редактирование retry policy через admin UI (трекается отдельно)
- Historical backfill прошлых scheduling-решений

Риски:
- "Partner" не определён — может значить billing entity или держателя
  API-ключа интеграции; это могут быть разные сущности.

Уточняющие вопросы:
- На какой идентификатор "partner" завязан override?
```

Используйте это перед `/spk-plan`, когда риск scope creep кажется высоким,
или в любой момент, когда коллега спрашивает "погоди, а эта фича ещё и X
делает?".

### Рецепт: Закрытие "почти готовой" фичи через `/spk-converge`

Какую проблему решает: фича, где большинство задач отмечено, но пара
`Proof:`-записей тонкие или отсутствуют, не должна требовать полного
аудита `/spk-verify` — `/spk-converge` это дешёвый цикл, который превращает
именно эти пробелы в follow-up задачи и перепроверяет до чистого состояния.

```bash
speckeep converge partner-scheduling
```

```text
gaps-found: 2 findings
- T2.3 is [x] but has no Proof: line
- AC-003 has no task mapped in Acceptance Coverage
exit code: 1
```

Затем агент добавляет секцию `## Converge Follow-ups` в `tasks.md`:

```md
## Converge Follow-ups

- [ ] T3.1 Добавить Proof для сохранения retry-window — outcome: у T2.3
      есть валидная строка Proof:; Touches: specs/active/partner-scheduling/tasks.md
- [ ] T3.2 Покрыть AC-003 задачей — outcome: подсказка manual override UI
      реализована и трассируема; Touches: src/ui/partner_override.go

## Acceptance Coverage

- AC-003 -> T3.2
```

...реализует их и перезапускает `speckeep converge partner-scheduling`, пока
не получит `converged`. Это легче, чем `/spk-verify`: без сохраняемого
отчёта, просто задачи/coverage возвращаются в синхрон — правильный
инструмент, когда фича готова на 90%, а не когда нужен полный аудит.

### Рецепт: Экстренный фикс через `/spk-hotfix`

Какую проблему решает: production-баг не ждёт
`spec → plan → tasks`. `/spk-hotfix` — узкий escape hatch: максимум 3
файла, без replanning, без расширения scope — если фикс требует больше,
это уже не hotfix.

```text
/spk-hotfix Экспорт-эндпоинт падает с 500, когда в отчёте ноль строк —
NPE на пустом result set.
```

```text
Changed files: src/export/handler.go, src/export/handler_test.go
Fixed: guard против nil/empty result set перед созданием CSV writer;
добавлен regression-тест для случая с нулём строк.
Verify: `go test ./src/export/...` — TestExportEmptyResultSet проходит.

Slug: hotfix-export-empty-result
Status: hotfix
Artifacts: src/export/handler.go, src/export/handler_test.go
Blockers: none
Ready for: speckeep archive hotfix-export-empty-result .
```

Если реальный фикс требует 4-й файл или изменение дизайна, агент
останавливается и говорит об этом, вместо того чтобы тихо растить hotfix —
это сигнал перейти на `/spk-spec`.

### Рецепт: Чистое возобновление через `/spk-handoff`

Какую проблему решает: сессия заканчивается посреди фичи, и следующая
сессия (ваша завтра, коллеги, или другой агент) не должна заново выводить
состояние, перечитывая все артефакты целиком.

```text
/spk-handoff partner-scheduling
```

```text
Slug: partner-scheduling
Phase: implement
What changed: T1.1, T1.2 готовы (data model + сохранение retry window).
T2.1 (scheduler integration) в процессе — override lookup подключён, ещё
не покрыт fallback-путь "нет override".
Open tasks: T2.1 (partial), T2.2, T3.1
Blockers: none
Next command: /spk-implement partner-scheduling --tasks T2.1
```

Вставьте это в новую сессию (или передайте коллеге) — и следующий агент
получает всё нужное без перечитывания `spec.md`/`plan.md` целиком.

### Рецепт: Обзор всего проекта через `/spk-recap`

Какую проблему решает: "какое состояние у всего?" по нескольким активным
фичам, не открывая каждую по отдельности.

```text
/spk-recap
```

```text
| Slug                | Phase     | Status (blockers?)        | Next                         |
| -------------------- | --------- | -------------------------- | ----------------------------- |
| partner-scheduling   | implement | in progress                | /spk-implement --tasks T2.1  |
| copy-as-json         | implement | done, unverified           | speckeep archive copy-as-json . |
| dark-mode            | plan      | blocked: inspect concerns  | /spk-spec (resolve open Qs)  |
```

Без end block, без глубокого разбора каждой фичи — `/spk-recap` это
status board, а не вывод фазы. Используйте его в начале сессии, чтобы
решить, за что браться.

### Рецепт: Отмена выполненной работы через `/spk-rollback`

Какую проблему решает: задача была отмечена `[x]` слишком рано, или подход
оказался неверным и требует переделки — `/spk-rollback` откатывает
*декларируемое* состояние (чекбоксы, опционально код) с полной
прозрачностью о том, что именно откатилось.

```text
/spk-rollback partner-scheduling
```

```text
Completed tasks:
Phase T1: T1.1 (data model), T1.2 (retry window persistence)
Phase T2: T2.1 (scheduler integration)

Which tasks should I roll back? (IDs like T1.1,T1.2, a phase like "T1", or
"all")
```

```text
/spk-rollback partner-scheduling T2.1
```

```text
Rolled back: T2.1 -> [ ] (только чекбокс; код изменения сохранены)
completed=2, open=2

Slug: partner-scheduling
Status: implement
Artifacts: specs/active/partner-scheduling/tasks.md
Blockers: none
Ready for: /spk-implement partner-scheduling
```

Код **не** откатывается автоматически — явно попросите об этом отдельно,
если хотите, чтобы для `Touches:`-файлов задачи был выполнен
`git checkout -- <file>`.

### Рецепт: Держим `REPOSITORY_MAP.md` честной через `/spk-repo-map`

Какую проблему решает: агенты, которые навигируют через `ls`/`find`/glob,
тратят токены на переоткрытие формы репозитория каждую сессию.
`REPOSITORY_MAP.md` — компактный, только-код индекс, который задуман для
чтения один раз за сессию вместо этого.

Запускайте, когда trigger checklist это подтверждает — не на каждое
изменение:

```text
Добавлена или удалена top-level директория/модуль кода.
Перемещены/переименованы ключевые пути исходников, меняющие навигацию.
Добавлены/удалены runtime/service/CLI entrypoints.
Изменены границы подсистем (пути "где редактировать" сместились существенно).
```

```text
/spk-repo-map
```

```text
Changed entries:
+ Добавлен `src/internal/importer/` (миграция openspec/speckit) в Top-Level Code
+ Добавлено "Add a CLI subcommand" -> src/internal/cli/ в Where To Edit

Map is up to date, 142/180 lines.

Slug: n/a
Status: repo-map
Artifacts: REPOSITORY_MAP.md
Blockers: none
```

Фича, которая трогает только существующие файлы внутри существующих
модулей, *не должна* триггерить обновление карты — это распространённый
случай, и агент должен распознавать его как таковой, а не обновлять карту
на всякий случай.

### Рецепт: Миграция из OpenSpec или Spec Kit

Какую проблему решает: у вас уже есть OpenSpec-пакет
`openspec/changes/<slug>/` или Spec Kit-пакет `specs/<slug>/`, и вы не
хотите переносить его вручную.

```bash
speckeep import openspec ./my-project
```

```text
imported: partner-scheduling
  spec.md    <- rebuilt from openspec Requirement:/Scenario: blocks (3 AC-*)
  plan.md    <- rebuilt from design.md (2 DEC-*)
  tasks.md   <- copied best-effort; run /spk-tasks to regenerate
             Touches:/Surface Map/Acceptance Coverage
skipped: dark-mode (already exists in specs/active/dark-mode/)
```

```bash
speckeep import speckit ./my-project
```

```text
imported: export-report
  spec.md    <- copied
  plan.md    <- copied
  tasks.md   <- copied
```

Существующие директории speckeep-фич **никогда** не перезаписываются —
коллизия имён репортится как `skipped`, а не тихо затирается. Для
импортированного `tasks.md` один раз запустите `/spk-tasks <slug>`, чтобы
пересобрать его с секциями `Touches:`/`Surface Map`/`Acceptance Coverage`,
на которые опираются фазы `implement`/`verify` speckeep.

### Рецепт: Подключаем `speckeep guard` в CI

Какую проблему решает: "готова ли каждая фича, затронутая этим PR, к
archive?" как детерминированная проверка, а не мнение ревьюера.

```bash
speckeep guard . --slug partner-scheduling --json
```

```json
{
  "slug": "partner-scheduling",
  "ready": false,
  "findings": [
    { "kind": "open_task", "task": "T2.2" },
    { "kind": "missing_proof", "task": "T2.3" }
  ]
}
```

Exit code 1 при любой находке — именно это валит PR. Готовый GitHub Action
(`contrib/ci/speckeep-guard.yml`) сам определяет изменённые директории
`specs/active/<slug>/` относительно base PR и прогоняет эту проверку по
каждому изменённому slug — в большинстве репозиториев не нужно перечислять
slug'и вручную.

### Рецепт: Держим бинарь актуальным

Какую проблему решает: "стоит ли у меня последняя версия speckeep, и могу
ли я обновиться, не прибегая снова к brew/scoop/curl?"

```bash
speckeep self check
```

```text
installed: v1.0.0
latest:    v1.1.0
update available — run `speckeep self upgrade`
```

```bash
speckeep self upgrade
```

```text
downloading speckeep_v1.1.0_linux_amd64.tar.gz...
sha256 verified
replaced: /home/you/.local/bin/speckeep (v1.0.0 -> v1.1.0)
```

На Windows, или когда директория установки недоступна для записи,
`self upgrade` печатает подсказку для ручной установки вместо того, чтобы
тихо провалиться — он никогда не оставляет вас с наполовину заменённым
бинарём.

### Рецепт: Express lane на практике

Какую проблему решает: не каждой фиче нужен `data-model.md` и формальный
`plan.md` — для маленьких низкорисковых правок это process weight, который
фича не несёт.

```text
/spk-propose Добавь метку времени "last synced at" на бейдж статуса
интеграции на странице настроек.
```

Агент идёт прямо от `spec.md` к `tasks.md` — без `plan.md`, без
`data-model.md`:

```text
Slug: last-synced-badge
Status: propose
Artifacts: specs/active/last-synced-badge/spec.md, specs/active/last-synced-badge/tasks.md
Blockers: none
Ready for: /spk-implement last-synced-badge
```

`speckeep check last-synced-badge` печатает express-mode подсказку вместо
ошибки для отсутствующего `plan.md`:

```text
express mode: plan.md not present — closing from spec.md + tasks.md is fine
for this feature size.
```

Любое plan-уровня решение, которое *действительно* стоит зафиксировать
(например "переиспользуем существующий polling interval, не добавляем
новый") идёт в секцию `## Implementation Context` файла `tasks.md` вместо
отдельного `plan.md` — один файл несёт ту нагрузку, для которой пакету из
трёх файлов пришлось бы существовать ради настолько маленькой правки.
