# speckeep

[![ci](https://github.com/bzdvdn/speckeep/actions/workflows/ci.yml/badge.svg)](https://github.com/bzdvdn/speckeep/actions/workflows/ci.yml)
[![release-build](https://github.com/bzdvdn/speckeep/actions/workflows/release-build.yml/badge.svg)](https://github.com/bzdvdn/speckeep/actions/workflows/release-build.yml)

Русская версия: [English README](README.md)

`speckeep` — это легкий файловый каркас проектного контекста для людей и агентов разработки.

Он хранит намерение проекта, спецификации, плановые артефакты и декомпозицию задач в простых файлах, не превращаясь в жесткий механизм управления процессом.

SpecKeep — преемник DraftSpec (архивирован). Если мигрируете существующий DraftSpec workspace, начните с `speckeep migrate`.

Первый релиз намеренно оптимизирован под низкие накладные расходы и реальную работу: узкий контекст по умолчанию, минимально обязательные артефакты, строгую дисциплину workflow без тяжеловесной оркестрации и branch-first модель collaboration, которая одинаково хорошо подходит и для одиночной, и для командной работы.

## Позиционирование

SpecKeep — это строгий и легкий SDD-kit для реальных кодовых баз.

Он рассчитан на команды, которым нужна более жесткая дисциплина, чем у loose planning layer, но не нужны ширина workflow, артефактные накладные расходы и orchestration-вес более тяжелой SDD-системы.

- строже, чем OpenSpec, в дисциплине фаз и согласованности артефактов
- легче, чем Spec Kit, по контексту по умолчанию, ширине workflow и артефактным накладным расходам
- оптимизирован для agent-first workflow с узкой загрузкой контекста
- сохраняет strictness через templates, entrypoints и readiness checks, а не через разрастание процесса
- рассчитан на brownfield-репозитории, где контекст должен оставаться узким, локальным и reviewable

Коротко: SpecKeep оптимизируется под discipline per token: сильные границы фаз, низкий artifact drag и достаточно структуры, чтобы люди и агенты оставались согласованы в повседневной работе.

## SpecKeep vs OpenSpec vs Spec Kit

| Dimension                | SpecKeep                                   | OpenSpec                                | Spec Kit                                  |
| ------------------------ | ------------------------------------------- | --------------------------------------- | ----------------------------------------- |
| Workflow style           | Строгая цепочка фаз с узким контекстом      | Более гибкий workflow вокруг артефактов | Более подробный многошаговый SDD-workflow |
| Default context size     | Самый маленький по умолчанию                | Средний                                 | Самый большой                             |
| Artifact overhead        | Низкий                                      | Средний                                 | Высокий                                   |
| Phase discipline         | Высокая                                     | Средняя                                 | Максимальная                              |
| Brownfield ergonomics    | Высокая                                     | Высокая                                 | Средняя                                   |
| Team collaboration model | Branch-first, feature-local artifacts       | Модель вокруг change-folders            | Тяжелее по веткам и workflow              |
| Shared mutable state     | Избегается по дизайну                       | Низкий                                  | Зависит от конфигурации                   |
| Best fit                 | Легкий строгий SDD для реальных кодовых баз | Гибкий SDD-lite для быстрой итерации    | Полноценный строгий SDD-toolkit           |

Коротко: SpecKeep занимает место между OpenSpec и Spec Kit: строже OpenSpec, легче Spec Kit и лучше приспособлен для branch-based collaboration с минимальным контекстом по умолчанию.

## Где SpecKeep Сильнее Всего

- Узкий контекст по умолчанию. Каждая фаза должна загружать только минимально полезный объем.
- Чтение кода должно оставаться фазово-локальным и точечным: достаточно, чтобы убрать догадки, но не настолько широко, чтобы пересобирать контекст всего репозитория.
- Строгая цепочка workflow. Конституция, spec, опциональный inspect, plan, tasks и implement остаются согласованными.
- **Greenfield-friendly**. Хотя SpecKeep оптимизирован для brownfield, он отлично подходит для проектов с нуля через подход «Foundation-first».

## Быстрый старт (Greenfield)

Если вы начинаете проект с нуля:

1.  **Init**: `speckeep init . --lang ru --shell sh`
2.  **Establishment**: Опишите стек, архитектуру и правила в `/speckeep.constitution --foundation`. Это создаст единый документ правил и технического фундамента проекта.
3.  **First Feature**: После того как база зафиксирована, переходите к первой функциональной спецификации через `/speckeep.spec`.

## Быстрый старт Skills

Для git-источников `--ref` обязателен: это фиксирует версию skill и предотвращает drift на плавающих ветках.

```bash
# добавить локальный skill
speckeep add-skill my-project --id architecture --from-local skills/architecture

# добавить skill из git (пин через --ref)
speckeep add-skill my-project --id openai-docs --from-git https://example.com/skills.git --ref v1.2.3 --path skills/openai-docs

# посмотреть настроенные skills
speckeep list-skills my-project

# установить skills в агентские папки
speckeep install-skills my-project

# синхронизировать только skills-managed артефакты
speckeep sync-skills my-project
```

## Пример использования (Brownfield)

- `inspect` — это полноценный quality gate, а не необязательная рекомендация перед planning.
- Легкая трассировка. Стабильные ID и дешевые readiness checks уменьшают перегрузку prompt-контекста.
- Удобство для brownfield-репозиториев. SpecKeep хорошо работает в существующих кодовых базах без навязывания тяжелого процессного слоя.
- Branch-first collaboration. Активное состояние фичи остается локальным для самой фичи, а не размазывается по общей изменяемой памяти.
- `inspect` перед `plan` опционален: запускайте его при неоднозначности, высоком риске или когда нужен формальный quality gate. Если inspect-отчет существует, он должен быть валидным и не иметь статуса `blocked`.
- Опциональные workflow-команды, доступные на любой фазе: `/speckeep.challenge` (adversarial review — находит слабые допущения и непроверяемые критерии), `/speckeep.handoff` (компактный документ передачи сессии, чтобы новая сессия могла продолжить без перечитывания всех артефактов), `/speckeep.hotfix` (экстренное исправление вне стандартной цепочки фаз — для понятных исправлений, затрагивающих ≤ 3 файла), `/speckeep.scope` (быстрая проверка границ scope, только inline, файл не создается).

OpenSpec по дизайну более гибкий и хорошо подходит командам, которым нужен более свободный workflow вокруг артефактов.

Spec Kit дает более широкую и подробную workflow surface, но обычно ценой большего числа артефактов, более широкого контекста и более тяжелого процесса.

SpecKeep оптимизируется под discipline per token: сильные границы workflow, минимальный контекст по умолчанию, явные quality gates и достаточную структуру для согласованной работы агентов без превращения процесса в тяжеловесную систему.

## Документация

Расширенная документация находится в [`docs/`](docs/README.md):

- [English docs](docs/en/index.md)
- [Русская документация](docs/ru/index.md)

## Публичный CLI

```text
speckeep init [path]
speckeep refresh [path]
speckeep add-agent [path]
speckeep list-agents [path]
speckeep remove-agent [path]
speckeep cleanup-agents [path]
speckeep add-skill [path]
speckeep list-skills [path]
speckeep remove-skill [path]
speckeep install-skills [path]
speckeep sync-skills [path]
speckeep skills install [path]
speckeep skills sync [path]
speckeep doctor [path]
speckeep doctor [path] --json
speckeep dashboard [path]
speckeep feature <slug> [path]
speckeep feature repair <slug> [path]
speckeep features [path]
speckeep migrate [path]
speckeep list-specs [path]
speckeep show-spec <name> [path]
speckeep check <slug> [path]
speckeep check <slug> [path] --json
speckeep check [path] --all
speckeep check [path] --all --json
speckeep trace [slug] [path]
speckeep demo [path]
speckeep export <slug> [path]
speckeep export <slug> [path] --output <file>
speckeep list-archive [path]
speckeep list-archive [path] --status <status>
speckeep list-archive [path] --since <YYYY-MM-DD>
speckeep list-archive [path] --json
```

## Workflow

```text
constitution -> spec -> [inspect, опционально] -> plan -> tasks -> implement -> verify -> archive
```

## Ключевые Идеи

- Конституция — главный документ проекта.
- Plan package хранит вместе `plan.md`, `tasks.md`, `data-model.md`, `contracts/` и optional `research.md`.
- `data-model.md` и `contracts/` должны оставаться компактными, но структурированными: сущности должны описывать поля, инварианты и жизненный цикл, а контракты — входы и выходы на границах системы, ошибки и предположения о доставке.
- Specs используют канонические маркеры `Given / When / Then` независимо от языка документации.
- SpecKeep предпочитает стабильные ID и явные ссылки вместо повторяющихся narrative summaries: `RQ-*` для требований, `AC-*` для критериев приемки, `DEC-*` для решений плана и phase-scoped `T*` для task IDs.
- Agent workflows должны читать только минимально нужный контекст.
- Strictness обеспечивается phase entrypoints, templates, стабильной структурой артефактов и readiness checks, а не большими prompt contexts.
- `inspect` теперь использует вывод helper scripts как основной слой структурных доказательств: readiness checks могут выдавать категоризированные findings вроде `structure`, `traceability`, `ambiguity`, `consistency` и `readiness`, которые агент должен сохранять и углублять только при необходимости.
- Agent-facing `/speckeep.spec` работает branch-first: от `feature/<slug>`, поддерживает `--name` с optional `--slug` / `--branch` и сохраняет приоритет явных `name:` / `slug:` в prompt-файлах.
- `speckeep init` требует явный `--shell` и генерирует только одно семейство scripts: `sh` или `powershell`. Поддерживаемые agent targets: `claude`, `codex`, `copilot`, `cursor`, `kilocode`, `trae`, `windsurf`, `roocode`, `aider`.
- Сгенерированный workspace включает `.speckeep/scripts/run-speckeep.*` как стабильный CLI launcher для агентов; он сначала использует `DRAFTSPEC_BIN`, а потом пытается вызвать `speckeep` из `PATH`.
- Сгенерированные обёртки `.speckeep/scripts/*` вычисляют корень проекта из расположения скрипта и передают его через `--root`, поэтому их можно запускать из любого текущего каталога.
- `speckeep feature repair` и `speckeep migrate` дают безопасную каноникализацию legacy-артефактов, например старых путей к inspect reports.
- `speckeep check <slug>` показывает наличие артефактов, вердикт inspect и verify, прогресс задач, точную следующую slash-команду и компактную сводку readiness checks; выходит с кодом 1, если заблокировано; поддерживает `--json` для CI. `--all` выводит таблицу готовности по всем фичам.
- `speckeep demo [path]` создает демо-workspace с заполненным примером фичи на фазе implement — spec, inspect-отчет, plan, tasks и data model уже заполнены.
- `speckeep export <slug>` упаковывает все артефакты фичи в один markdown-документ для передачи ревьюеру или новой агентской сессии; поддерживает `--output` для записи в файл.
- `speckeep list-archive` выводит архивированные фичи из `archive/`; показывает одну запись на slug (последний снимок) со статусом, датой и причиной; поддерживает `--status` для фильтрации по статусу, `--since <YYYY-MM-DD>` для фильтрации по дате и `--json` для автоматизации.
- `/speckeep.plan` поддерживает флаг `--research`: входит в режим research-first — агент фиксирует конкретные неизвестные, пишет `research.md`, затем спрашивает "Research complete — proceed to full plan?" перед созданием `plan.md`.
- `/speckeep.plan` включает секцию `## Incremental Delivery`: направляет агентов на определение MVP (наименьшего тестируемого инкремента) и планирование шагов итеративного расширения с трассировкой AC — предотвращает монолитные реализации и позволяет проводить раннюю валидацию.
- `/speckeep.spec` поддерживает флаг `--amend`: режим точечного редактирования — добавить критерий или исправить секцию без переписывания спека и без инвалидации inspect-отчёта.
- `/speckeep.handoff` без slug генерирует handoff-документы для всех активных фич одновременно.
- `/speckeep.hotfix`: экстренный workflow исправления — пишет минимальный hotfix-спек (fix, root cause, risk, verification, touches) до любого изменения кода, реализует, проверяет inline и архивирует; пропускает фазы inspect, plan и tasks; используйте только когда причина известна и исправление затрагивает ≤ 3 файлов.
- `doctor` предупреждает, когда один и тот же стабильный ID (`AC-*`, `RQ-*`) встречается сразу в нескольких спеках.
- **Прослеживаемость (Traceability) по дизайну**. Агенты инструктированы аннотировать код метками `// @sk-task <ID> (<AC_ID>)` во время реализации. Используйте `speckeep trace <slug>` для сканирования и проверки связи между кодом и требованиями.
- Генерируемые docs и prompts поддерживают английский и русский.

## Быстрый Пример

```bash
# попробовать демо сразу — без настройки проекта
speckeep demo ./my-demo

# инициализировать реальный проект
speckeep init my-project --lang ru --shell sh --agents claude --agents codex
speckeep refresh my-project --shell powershell --dry-run
speckeep doctor my-project
speckeep check export-report my-project
```

## Полный пример цикла фичи

Весь workflow на реальной задаче — «Добавить экспорт отчётов в CSV».

<details>
<summary>Посмотреть полный цикл →</summary>

### 1. Инициализация

```bash
speckeep init . --lang ru --shell sh --agents claude
# → scaffold .speckeep/, AGENTS.md и файлы slash-команд созданы
```

Опциональные advanced флаги:

- `--specs-dir` переопределяет директорию для спецификаций (по умолчанию: `specs`)
- `--archive-dir` переопределяет директорию для архива (по умолчанию: `archive`)
- `--constitution-file` переопределяет путь к конституции проекта (по умолчанию: `CONSTITUTION.md`)

### 2. Написать спек

Вызовите `/speckeep.spec --name "Экспорт отчётов в CSV"` в агенте.

`specs/eksport-otchetov-v-csv/spec.md`:

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

Вызовите `/speckeep.inspect eksport-otchetov-v-csv`.

- `specs/eksport-otchetov-v-csv/inspect.md` — вердикт `pass`, все AC в формате G/W/T
- `specs/eksport-otchetov-v-csv/summary.md` — компактная таблица AC, которую используют implement и verify вместо полного spec

### 4. Plan

Вызовите `/speckeep.plan eksport-otchetov-v-csv`.

`plan.md` называет surfaces: `ReportsPage.tsx` (кнопка), `useReportExport.ts` (новый хук, CSV-логика), `reports.test.ts` (тесты).

### 5. Tasks

Вызовите `/speckeep.tasks eksport-otchetov-v-csv`.

`specs/eksport-otchetov-v-csv/plan/tasks.md`:

```markdown
## Surface Map

| Surface                    | Задачи |
| -------------------------- | ------ |
| hooks/useReportExport.ts   | T1.1   |
| components/ReportsPage.tsx | T1.2   |
| tests/reports.test.ts      | T2.1   |

## Фаза 1: Хук и кнопка

- [ ] T1.1 добавить хук `useReportExport` — конвертирует `rows[]` в CSV blob и инициирует скачивание — AC-001 `Touches: hooks/useReportExport.ts`
- [ ] T1.2 добавить кнопку «Экспортировать CSV» в ReportsPage — вызывает хук, неактивна при пустой таблице — AC-001, AC-002 `Touches: components/ReportsPage.tsx`

## Фаза 2: Тесты

- [ ] T2.1 добавить тесты для useReportExport — непустые строки, пустая таблица, формат заголовков — AC-001, AC-002 `Touches: tests/reports.test.ts`

## Покрытие критериев приемки

AC-001 → T1.1, T1.2, T2.1
AC-002 → T1.2, T2.1
```

### 6. Implement, verify, archive

```
/speckeep.implement eksport-otchetov-v-csv   # Фаза 1 выполнена, стоп
/speckeep.implement eksport-otchetov-v-csv   # Фаза 2 выполнена, стоп
/speckeep.verify    eksport-otchetov-v-csv   # вердикт: pass
speckeep archive    eksport-otchetov-v-csv .
```

### Проверить готовность в любой момент

```bash
speckeep check eksport-otchetov-v-csv
# Фаза:   tasks → implement
# Задачи: 0 / 3 выполнено
# Далее:  /speckeep.implement eksport-otchetov-v-csv
```

</details>

## Установка

SpecKeep распространяется как один бинарник через GitHub Releases.

Linux / macOS:

```bash
VERSION=v0.1.0
curl -fsSL "https://raw.githubusercontent.com/bzdvdn/speckeep/${VERSION}/scripts/install.sh" | bash -s -- --version "${VERSION}"
```

Windows (PowerShell):

```powershell
$version="v0.1.0"
$env:DRAFTSPEC_VERSION=$version
powershell -ExecutionPolicy Bypass -c "iwr -useb https://raw.githubusercontent.com/bzdvdn/speckeep/$version/scripts/install.ps1 | iex"
```

Чтобы также добавить папку установки в `PATH`:

- Linux: добавь `--add-to-path` или установи `DRAFTSPEC_ADD_TO_PATH=1`
- Windows: установи `$env:DRAFTSPEC_ADD_TO_PATH=1` или запускай скрипт с `-AddToPath`

Для более подробного знакомства смотри:

- [Обзор](docs/ru/overview.md)
- [CLI](docs/ru/cli.md)
- [Модель workflow](docs/ru/workflow.md)
- [Примеры](docs/ru/examples.md)
- [Roadmap](docs/ru/roadmap.md)

## Разработка

```bash
go test ./...
go build -o bin/speckeep ./src/cmd/speckeep

# с указанием версии
go build -ldflags "-X speckeep/src/internal/cli.Version=v0.1.0" -o bin/speckeep ./src/cmd/speckeep
```

Репозиторий содержит unit tests для config, project lifecycle, doctor checks, specs, templates, agents и CLI-level behavior.

## Лицензия

Проект распространяется по лицензии [MIT](LICENSE).
