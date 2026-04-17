# Prompt задач SpecKeep

Создаёте или обновляете `<specs_dir>/<slug>/plan/tasks.md` (по умолчанию: `.speckeep/specs/<slug>/plan/tasks.md`).

Следуйте базовым правилам в `AGENTS.md` (пути, git, load discipline, readiness scripts, язык, phase discipline).

## Цель

Разбить согласованный план на исполнимые implementation tasks.

## Phase Contract

Inputs: `.speckeep/constitution.md`, `.speckeep/specs/<slug>/plan/plan.md`; опционально `summary.md`, `plan/data-model.md`, `plan/contracts/` если декомпозиция требует.
Outputs: `.speckeep/specs/<slug>/plan/tasks.md` с фазовым списком и `## Покрытие критериев приемки`.
Stop if: `plan.md` нет, план расплывчат, или хотя бы один AC нельзя привязать к исполнимой работе.

## Режим работы

- Декомпозируйте один согласованный plan package.
- `plan.md` — entrypoint; глубже только по необходимости.
- Минимальный task list, безопасно покрывающий фичу.
- Явная последовательность > umbrella tasks.

## Load First

Всегда сначала:

- `.speckeep/constitution.summary.md` если есть; он всегда живет по фиксированному technical path в `.speckeep/`
- Иначе `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/plan/plan.md`

## Load If Present

Только когда декомпозируемая задача на них ссылается / от них зависит:

- `summary.md` (или `spec.md`) — `AC-*` с неоднозначным scope или границей приёмки
- `plan/data-model.md` — задачи создают/меняют сущности, поля, инварианты, transitions
- `plan/contracts/` — задачи создают/меняют API endpoints, event payloads, integration boundaries
- `plan/research.md` — только если finding меняет порядок задач или вносит риск, влияющий на декомпозицию
- code files — только файлы, нужные для определения начала/конца implementation surfaces

Не считайте, что `research.md` обязан существовать; используйте его только когда план явно зависит от сохранённой неопределённости, внешней зависимости или зафиксированного trade-off.

## Do Not Read By Default

- implementation-файлы, не нужные для декомпозиции
- широкая история репо

## Flags

`--repair <task-id-list>`: точечное исправление — правьте конкретные задачи из verify/review без переписывания всего списка.

- Прочитайте существующий `tasks.md`. Найдите только указанные задачи (`--repair T2.3,T3.1`).
- Для каждой: обновите описание, outcome, `Touches:` или привязку `AC-*`.
- Не перестраивайте фазы, не перенумеровывайте, не переписывайте неправленые секции.
- Repair выявил проблему в плане → стоп и предложите `/speckeep.plan <slug> --update`.
- `## Покрытие критериев приемки` — обновляйте только если изменился маппинг. `## Surface Map` — только если изменились `Touches:`.

`--greenfield-story`: story-first декомпозиция для greenfield или ранней продуктовой работы.

- `Touches:` и `## Surface Map` остаются обязательными.
- Группируйте поставку вокруг foundation, MVP story, следующей story и hardening, а не только по техническим слоям, если так проще исполнять план.
- Предпочитайте фазы, соответствующие независимо демонстрируемым продуктовым срезам.

## Stop Conditions

Остановитесь и запросите refinement, если:

- `plan.md` отсутствует
- задачи вышли бы расплывчатыми из-за слабого плана
- декомпозиции не хватает spec/data-model/contracts/research
- конституция блокирует декомпозицию
- декомпозиция расползается на несколько feature slug или несвязанные change sets
- один или несколько acceptance criteria нельзя сопоставить с исполнимой работой без догадок

Не переходите к реализации.

## Инварианты

- Задачи соответствуют плану и конституции.
- `plan.md` — entrypoint. Не читайте нерелевантные артефакты других фич, чтобы компенсировать слабый план.
- Читайте код только если без этого задачи останутся расплывчатыми; узкий срез > широкое исследование.
- Если `/.speckeep/scripts/check-tasks-ready.*` доступен — запускайте (slug первый аргумент: `bash ./.speckeep/scripts/check-tasks-ready.sh <slug>` или PowerShell `.\.speckeep\scripts\check-tasks-ready.ps1 <slug>`).
- Task list исполним по порядку. Каждый AC покрыт ≥ 1 задачей.
- Предпочитайте конкретные, проверяемые, implementation-oriented задачи. Validation/docs — только когда нужно. Никаких umbrella-задач.
- The task list should be readable to both an implementation agent and a human reviewer без дополнительной интерпретации.
- Targeted code reading во время декомпозиции полезен, если уменьшает повторное чтение на implement.
- Не начинайте implementation work, не редактируйте исходный код и не заявляйте, что задачи выполнены, на фазе tasks.

## Правила формата задач

- Следуйте `.speckeep/templates/tasks.md`: фазы `## Фаза N: Название`.
- Каждая задача — phase-scoped `T<phase>.<index>`.
- Формат: `- [ ] T<phase>.<index> <глагол> — <конкретный измеримый результат>`
- Ссылайтесь на 1–2 стабильных ID (`AC-*`, `RQ-*`, `DEC-*`) когда возможно.
- Каждая задача содержит `Touches:` с конкретными файлами/модулями, которые она создаёт или меняет. Это основной механизм предотвращения повторных чтений на implement — агент batch-читает за один проход. Компактно (`Touches: src/auth/handler.ts, src/session/store.ts`). Module-level — только когда точный файл неясен (`Touches: src/auth/`). Без `Touches:` → exploratory reads, трата токенов.
- Секция `## Surface Map` перед первой фазой: таблица (`Surface | Tasks`) с каждой уникальной surface и task IDs, которые её трогают. Без неё агент сканирует каждую строку задачи.
- Задачи в совокупности покрывают все AC. Непокрытый AC = blocker.
- `## Покрытие критериев приемки` — ≥ 1 явная строка покрытия на AC (`AC-001 -> T1.1, T2.1`).
- AC покрыт, только когда ВСЕ привязанные задачи завершены. Любая открытая → verify считает AC незавершённым.
- Новые task lists требуют ID. Существенная правка списка без ID → нормализуйте в ID-формат.

## Правила качества содержимого

- Each phase should have a short goal, объясняющий зачем эта фаза существует.
- `--greenfield-story`: после foundation предпочитайте одну фазу на MVP или приоритезированную user story только когда план уже чётко задаёт эти срезы.
- `--greenfield-story`: держите story-фазы маленькими; если story нельзя продемонстрировать независимо, сначала декомпозируйте по MVP slice.
- **Ленивая декомпозиция**: несколько конкретных задач (5–10 на фичу) с измеримым результатом > много мелких bookkeeping. Не создавайте «микро-задачи» (1–5 строк кода); implement уточнит по месту.
- Фокус на «этапных» задачах, привязанных к файлам или функциональным границам.
- Outcome каждой задачи ≤ 12 слов. Больше → разбейте или уточните глагол.
- Простой acceptance proof → встраивайте в outcome: лучше `add POST /auth/login — возвращает 200 с полем JWT token — AC-001`, чем `add login handler — endpoint работает — AC-001`.
- Action verbs: implement, add, migrate, validate, remove, backfill, document.
- Отделяйте foundational setup от core behavior; proof/validation — отдельно от broad implementation.
- `Touches:` — конкретные пути (`src/auth/handler.ts`), не абстрактные концепции (`auth flow`).
- Задача существует только для доказательства поведения → говорите это явно, не прячьте в большей задаче.
- Фаза не нужна → пропустите или явно укажите, не заполняйте generic-tasks.
- Текст задачи делает результат очевидным без возврата в plan.
- Избегайте: `misc`, `cleanup as needed`, `wire everything up`, `финальная полировка`, прячущих outcome за generic-глаголом.

## Output

- Запишите/patch `tasks.md`; если декомпозиция заблокирована — скажите прямо.
- Завершайте summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Готово к`.
- Готово: `Готово к: /speckeep.implement <slug>`.

## Self-Check

- Could another developer execute these tasks in order without guessing what `done` means для каждой строки?
