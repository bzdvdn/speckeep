# Prompt плана SpecKeep

Создаёте или обновляете implementation plan package одной фичи.

## Goal

Соберите planning-артефакты в `<specs_dir>/<slug>/plan/` (по умолчанию: `.speckeep/specs/<slug>/plan/`).

## Phase Contract

Inputs: `.speckeep/constitution.md`, `.speckeep/specs/<slug>/spec.md`, `.speckeep/specs/<slug>/inspect.md`, узкий контекст репозитория.
Outputs: `plan/plan.md`, `plan/data-model.md`; опционально `plan/contracts/`, `plan/research.md`.
Stop if: spec или inspect отсутствуют, spec слишком расплывчата, или конфликт с конституцией.

Если `.speckeep/speckeep.yaml` переопределяет пути — следуйте конфигу. Читайте его ≤1 раза за сессию.

## Флаги

`--update`: точечное изменение — одна секция / один `DEC-*` / одна surface / один контракт без переписывания.

- Читайте существующие артефакты; меняйте только запрошенное.
- Меняется `DEC-*` — обновите его `Affects`/`Validation`, если они устарели.
- Держите `plan.md`, `data-model.md`, `contracts/` согласованными.
- `tasks.md` не инвалидируйте, если правка не затрагивает декомпозицию или AC coverage.

`--research`: сначала research, потом план.

- Определите 1–5 конкретных unknowns, блокирующих планирование → в `plan/research.md`.
- Остановитесь и спросите: «Исследование завершено — перейти к плану?».
- `plan.md`/`data-model.md` — только после явного подтверждения. Отказ — сессия завершается только с `research.md`.

`--update` и `--research` одновременно не принимаются — остановитесь и спросите.

`--greenfield`: режим greenfield-first planning.

- Используйте, когда репозиторий ещё не даёт meaningful implementation surfaces для этой фичи.
- В этом режиме план держится вокруг первого поставляемого среза, bootstrapping surfaces и кратчайшего validation path к MVP.

## Load and Scope

Загружайте: constitution, spec, inspect, и только тот код, что нужен для фиксации surfaces, границ и ограничений.
Не читайте по умолчанию: нерелевантные области репо, артефакты других фич, optional `research.md`.
Не читайте `/.speckeep/scripts/*` по умолчанию — используйте readiness wrapper (кроме отладки скриптов / работы над самим SpecKeep / явной просьбы).

Если доступен `/.speckeep/scripts/check-plan-ready.*` — запускайте вместо чтения исходников. Slug — первый аргумент: `bash ./.speckeep/scripts/check-plan-ready.sh <slug>` (PowerShell: `.\.speckeep\scripts\check-plan-ready.ps1 <slug>`).

Не создавайте и не переключайте ветки — feature branch должна существовать с фазы spec. Не на ней — остановитесь, не создавайте.

Не создавайте `tasks.md` — это отдельная фаза `/speckeep.tasks`. Останавливайтесь после plan-артефактов с next-command.

## Stop Conditions

Остановитесь и попросите refinement, если:

- `spec.md`/`inspect.md` отсутствуют — не создавайте их во время plan; попросите пользователя запустить `/speckeep.spec`, `/speckeep.inspect`, затем повторить `/speckeep.plan`
- spec слишком расплывчата для архитектуры/contracts/data-model
- конституция конфликтует с планом
- план ведёт через неясную integration/architectural boundary без подтверждения spec и focused repo-свидетельствами
- работа имеет смысл только при одновременном планировании нескольких feature packages

Не компенсируйте чтением широкого нерелевантного контекста.

## Required Outputs

Всегда:

- `plan/plan.md`
- `plan/data-model.md` — всегда. Если фича не вводит и не меняет persisted state, форму сущностей, lifecycle или contract-relevant payload shape, записывайте компактный no-change stub вместо пропуска файла.

Создавайте только когда оправдано:

- `plan/contracts/api.md` — фича трогает API boundary
- `plan/contracts/events.md` — фича публикует/потребляет события
- `plan/quickstart.md` — режим `--greenfield` или раннее планирование фичи выигрывает от короткого MVP validation flow

`plan/research.md` — только если хотя бы одно:

- внешняя система/API/зависимость с неясным поведением
- несколько реалистичных implementation options со значимыми trade-off'ами
- неочевидный performance/security/reliability/integration-риск
- нужно исследовать repo-constraint или архитектурную границу до конкретизации плана

Перед `research.md` зафиксируйте unknowns: 1–5 конкретных, привязанных к решению/риску/границе этой фичи; не исследуйте технологию «вообще» — только узкий вопрос, меняющий план. Нет unknowns — нет `research.md`.

Не создавайте `research.md` ради generic brainstorming или очевидной реализации, уже выводимой из spec и repo.

## Инварианты

- Соответствие конституции. Inspect — обязательный prerequisite.
- Планируйте текущую spec, а не идеализированную архитектуру; опирайтесь на репо-реальность.
- Код читайте узко — ровно для фиксации surfaces, boundaries, constraints. Широкое исследование не приветствуется.
- Новые файлы — по шаблонам `.speckeep/templates/plan.md`, `.speckeep/templates/data-model.md`, `.speckeep/templates/quickstart.md`.
- План называет concrete implementation surfaces и сопоставляет каждый `AC-*` с подходом до того, как пишутся `tasks`.
- Значимые решения → `DEC-*` с `Why`, `Tradeoff`, `Affects`, `Validation`.
- Data-model и contracts согласованы со spec и AC. Каждая сущность/контракт ссылается на оправдывающий `AC-*`. Форму сущностей, boundary IO, event payload не оставляйте только в prose внутри `plan.md`.
- `data-model.md` ДОЛЖЕН существовать всегда к концу plan. Если meaningful model changes нет, файл всё равно обязан явно содержать: status, reason и revisit triggers. Отсутствие файла заставляет следующие фазы гадать и считается дефектом planning.
- Реальные сущности в `data-model.md` → в `plan.md` однострочное обоснование с именами сущностей/инвариантов/lifecycle. Создан `contracts/` → однострочное обоснование с именем API/event boundary; иначе пишите «No API or event boundaries introduced».
- Технологии/библиотеки/версии фиксируйте только когда они материально влияют на implementation shape, integration boundaries, validation или risk. Названная версия/зависимость — с объяснением почему важна (compatibility, repo-constraint, внешний contract, rollout risk, validation). Не перечисляйте stack ради полноты.
- Опциональные артефакты остаются опциональными — не создавайте по привычке.
- Не пишите task checklist, не редактируйте implementation-код, не выдавайте verify/archive-вердикты в planning.
- Формулировки «обновить backend по необходимости»/«провести через систему» — сигнал refinement, не план.
- Если downstream task-writer должен гадать про метод/границу/validation для `AC-*` — план недостаточно конкретен.
- План достаточно конкретен, чтобы агент и reviewer видели форму реализации, tradeoffs и rollout implications без повторного чтения всего репо.
- Настроенный язык документации; `plan.md`/`data-model.md`/`contracts/`/`research.md` внутренне согласованы.

## Правила качества секций

- `## Цель` — форма реализации, не пересказ spec.
- `--greenfield`: `## MVP Slice` должен назвать минимальный независимо демонстрируемый инкремент и его покрытие по `AC-*`.
- `--greenfield`: `## First Validation Path` должен объяснять, как человек или агент быстро доказывает работоспособность MVP без перечитывания всего репо.
- `--greenfield`: `## Bootstrapping Surfaces` должен перечислить первые директории, файлы или границы, которые должны появиться до приземления feature behavior.
- `## Implementation Surfaces` — почему каждая меняется, новая или существующая.
- `## Acceptance Approach` — каждый `AC-*` → затрагиваемые surfaces + observable proof.
- `## Данные и контракты` — что меняется, что остаётся, почему.
- `## Порядок реализации` — must-happen-first vs параллелизуемое.
- `## Риски` — с mitigation, не только ярлыком.
- `## Rollout and Compatibility` — явно, когда migration/flags/compatibility/ops важны; явно, когда нет.
- `## Проверка` — привязана к `AC-*`/`DEC-*`, не общий список тест-идей.
- `## Соответствие конституции` — обязательна, перечисляет конкретные constraints (напр., «PostgreSQL согласно [CONST-DB]») с подтверждением каждого или объяснением resolve/defer. Голое «нет конфликтов» недостаточно.
- `Unknowns First` pass перед финализацией: неясное решение/surface/validation — фиксируйте явно или останавливайтесь на refinement.
- Конкретные implementation-guidance > архитектурное эссе. Абзац не уменьшает downstream guesswork — ужмите.

## Output

- Запишите/patch plan-артефакты; укажите, какие опциональные созданы и почему.
- Если создан `quickstart.md`, прямо скажите, что он нужен для проверки MVP path без повторного чтения всего плана.
- Опишите ключевые технические решения и риски, блокирующие следующие фазы.
- Завершайте summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Готово к`.
- Готово: `Готово к: /speckeep.tasks <slug>`.

## Self-Check

- Можно писать `tasks` без догадок, и опциональные артефакты оправданы?
