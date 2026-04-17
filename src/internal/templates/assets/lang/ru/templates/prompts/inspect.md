# Prompt проверки SpecKeep

Проверяете один пакет фичи на согласованность и качество.

Следуйте базовым правилам в `AGENTS.md` (пути, git, load discipline, readiness scripts, язык, phase discipline).

## Цель

Сфокусированный отчёт проверки для одной фичи без расширения scope.

## Phase Contract

Inputs: `.speckeep/constitution.md`, `.speckeep/specs/<slug>/spec.md`; опционально `plan/plan.md`, `plan/tasks.md` если есть.
Outputs: `.speckeep/specs/<slug>/inspect.md` с verdict `pass`, `concerns` или `blocked`; и `.speckeep/specs/<slug>/summary.md`.
Stop if: slug неоднозначен, spec отсутствует, или отчёт потребовал бы выдумывать product intent.

## Flags

`--delta`: инкрементальная перепроверка — только секции, изменившиеся с последнего inspect report.

- Прочитайте существующий `inspect.md` как baseline; сравните текущий `spec.md` для выявления изменённых секций (AC, scope, assumptions).
- Перепроверьте изменённые секции и их cross-artifact implications. Сохраните валидные findings; не пере-derive.
- Verdict меняйте, только если delta его меняет. Разрешённый `blocked` без новых ошибок → upgrade.
- Обновите `generated_at`; добавьте `delta_from: <previous_generated_at>` в metadata block.
- Delta затронула `## Goal`, `## Scope` или > 50% `AC-*` → откат к полной проверке с пометкой: «Delta mode откатился к полной проверке из-за масштаба изменений.»
- Изменился любой `AC-*` → регенерируйте `summary.md`. Иначе — оставьте.

## Load First

Всегда сначала прочитайте:

- `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/spec.md`

## Load If Present

Читайте только если файл существует и inspect требует cross-artifact consistency checks (spec↔plan alignment, acceptance↔tasks coverage):

- `plan/plan.md` — для goal alignment, scope, plan-level acceptance coverage
- `plan/tasks.md` — чтобы проверить покрытие каждого `AC-*` ≥ 1 задачей

## Do Not Read By Default

- `plan/data-model.md`, `plan/contracts/`, `plan/research.md`
- широкая история репо
- implementation-файлы — кроме случая, когда finding называет конкретный файл, и claim нельзя подтвердить из spec/plan/tasks

## Stop Conditions

Минимальный уточняющий вопрос только если: slug неоднозначен, spec отсутствует, или inspect пришлось бы выдумывать product intent.

## Rules

- Сначала проверяйте соответствие конституции.
- Если `/.speckeep/scripts/check-inspect-ready.*` доступен — запускайте как cheap first pass (slug первый аргумент: `bash ./.speckeep/scripts/check-inspect-ready.sh <slug>` или PowerShell `.\.speckeep\scripts\check-inspect-ready.ps1 <slug>`). Fallback: `/.speckeep/scripts/inspect-spec.*`. Нет ни того, ни другого — делайте ручную проверку по `constitution.md` + `spec.md`.
- Предпочитайте вывод helper scripts чтению исходников. Используйте `ERROR`/`WARN` findings как основной структурный слой; не переизобретайте. Сохраняйте категории finding'ов (structure, traceability, ambiguity, consistency, readiness).
- Не игнорируйте конкретный helper finding из-за общего оптимизма — либо закрывайте, либо объясняйте.
- Собственное reasoning — для того, что cheap checks не доказывают: конфликты с конституцией, выдуманный product intent, необоснованный scope expansion, противоречивые assumptions, тонкий spec↔plan drift.

### Проверки spec

- Проверяйте полноту и ясность спецификации.
- Verify `constitution <-> spec`: spec не должна противоречить constitutional constraints, workflow-правилам или language policy.
- Treat technology names, framework choices, library lists, or version pins in the spec as a `Warning` unless they clearly represent a user requirement, repository constraint, or external compatibility contract.
- Каждый AC ДОЛЖЕН использовать Given/When/Then (маркеры канонические). Отсутствие G/W/T — `Error`.
- Оставшийся `[NEEDS CLARIFICATION: ...]` — `Error` (закрыть до plan).
- Нет `## Допущения` → `Warning`. Допущение, противоречащее реальности репо, — `Error`.
- `## Критерии успеха` присутствует → каждый `SC-*` с метрикой + методом. Размытый SC — `Warning`.

### Cross-artifact проверки

- Самый дешёвый scope сначала: constitution + spec, затем plan, затем tasks, затем глубокие plan-артефакты — только если конкретный claim их требует.
- Нет `plan.md` → не расширяйте на optional plan artifacts или код.
- Есть `plan.md` → сначала Verify `spec <-> plan`; `data-model.md` / `contracts/` — только если `plan.md` на них опирается или claim их требует. План сохраняет goal, отражает major acceptance-critical behavior, не добавляет необоснованных новых workstreams. Проверяйте:
  - `Goal Alignment` — цель фичи не изменилась
  - `Scope Expansion` — нет новых крупных workstreams/компонентов/surfaces вне spec
  - `Acceptance Coverage at Plan Level` — major acceptance-critical behavior отражено в намерении плана
  - `Constitution Consistency` — план не нарушает конституции
  - `Artifact Justification` — `data-model.md`/`contracts/` оправданы spec'ом
- `plan.md` без `## Соответствие конституции` / `## Constitution Compliance` — `Warning`.
- `plan.md` есть, но `data-model.md` отсутствует → `Error`. Фаза plan обязана либо описать model changes, либо сохранить явный no-change stub.
- `data-model.md` есть, но лишь расплывчато намекает на отсутствие изменений → `Warning`; предпочитайте явный stub со status/reason/revisit triggers, чтобы downstream phases не гадали.
- `tasks.md` существует → verify `plan <-> tasks`: фазы/ID отражают план без явных пропусков по acceptance-critical behavior.
- `tasks.md` существует → каждый AC покрыт ≥ 1 задачей; непокрытый AC — `Error`. Нет `## Surface Map` — `Warning`. Строка с task ID без `Touches:` — `Warning`. Traceability со ссылкой на task ID (`T1.1`) напрямую.
- Не превращайте в широкий design review.

## Отчёт

- Пишите отчёт на настроенном языке документации.
- Предпочитайте конкретные findings общим советам. Порядок: (1) структурные из helper output, (2) cross-artifact consistency из загруженных артефактов, (3) узкие judgment calls.
- Default to a compact report in conversation output: всегда `Verdict`; `Errors`/`Warnings`/`Next Step` — если не пусты; `Questions`/`Suggestions`/`Traceability` — только когда дают сигнал.
- Produce the full sectioned report only when the user explicitly asks for a full report или при сохранении в файл.
- При записи в файл добавляйте сверху machine-readable metadata block с `report_type`, `slug`, `status`, `docs_language`, `generated_at`.
- Структура: YAML metadata → `# Inspect Report: <slug>` → `## Scope` → `## Verdict` → `## Errors` → `## Warnings` → `## Questions` → `## Suggestions` → `## Traceability` → `## Next Step`.
- The `## Verdict` section MUST use one of: `pass`, `concerns`, `blocked`.
  - `pass`: ошибок нет, только минорные warnings или их нет.
  - `concerns`: можно двигаться, но warnings / traceability gaps / open questions закрыть в ближайшее время.
  - `blocked`: constitutional conflicts, отсутствующий spec intent, отсутствующий Given/When/Then, непокрытые AC или major `spec <-> plan` contradictions блокируют безопасный переход.
- `## Traceability` суммирует AC → tasks когда `tasks.md` есть. Предпочитайте ID вроде `AC-001 -> T1.1, T2.1`.
- `## Next Step`:
  - `pass` — точная следующая slash-команда.
  - `concerns` — можно ли продолжать; если да — точная slash-команда.
  - `blocked` — не предлагайте следующую фазу; укажите, какой refinement нужен.
- Не дублируйте findings между секциями.

## Артефакт summary.md

После inspect-отчёта также запишите `.speckeep/specs/<slug>/summary.md`:

- YAML frontmatter: `slug`, `generated_at`
- `## Goal` — одно предложение
- `## Acceptance Criteria` — таблица `ID | Summary | Proof Signal`; summary ≤ 8 слов; proof signal = наблюдаемая проверка из `Then`
- `## Out of Scope` — 3–5 bullets

Не длиннее 25 строк. Загружается `tasks`/`implement`/`verify` для снижения context. Не заменяет полный spec на фазах с полной AC-проверкой (inspect, plan).

## Output

- Пишите `inspect.md` и `summary.md`.
- Суммируйте verdict в разговоре (compact report, только непустые секции).
- Завершайте summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Готово к`.
- Готово: `Готово к: /speckeep.plan <slug>` (или `/speckeep.tasks <slug>` если план уже есть; после archive `/speckeep.recap` — опционально, не обязательно).
