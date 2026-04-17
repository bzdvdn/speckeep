# Prompt проверки реализации SpecKeep

Проверяете один feature package после выполнения задач.

Следуйте базовым правилам в `AGENTS.md` (пути, git, load discipline, readiness scripts, язык, phase discipline).

## Цель

Подтвердить, что реализация достаточно согласована с задачами и правилами проекта для безопасного перехода.

## Flags

`--deep`: полная валидация — читает все plan artifacts и реальный код для каждой завершённой задачи и AC. Per-AC evidence. Без — структурная и cheap.

`--persist`: для обратной совместимости. По умолчанию ОБЯЗАНЫ сохранять отчёт в `.speckeep/specs/<slug>/plan/verify.md` (в дополнение к чату) по шаблону `.speckeep/templates/verify.md` с machine-readable metadata block. Не пишите файл только если пользователь явно просит «только в чат».

## Phase Contract

Inputs: `.speckeep/constitution.md`, `.speckeep/specs/<slug>/plan/tasks.md`; spec/plan/код — только для подтверждения конкретных выводов (все артефакты в `--deep`).
Outputs: отчёт с verdict (`pass`, `concerns` или `blocked`) в чате И сохранённый в `plan/verify.md` по умолчанию.
Stop if: slug неоднозначен, `tasks.md` нет, или verdict потребовал бы выдумывать факты о реализации.

## Load First

Всегда сначала:

- `.speckeep/constitution.summary.md` если есть; иначе `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/plan/tasks.md`

## Load If Present

Только когда конкретная проверка ссылается на содержимое:

- `summary.md` (или `spec.md`) — acceptance coverage или task↔AC alignment
- `plan/plan.md` — задача ссылается на `DEC-*` или архитектурное решение, требующее подтверждения
- `plan/data-model.md` — задача трогает persisted state или форму сущности
- `plan/contracts/` — задача трогает API или event boundaries
- `plan/research.md` — только если проверка зависит от зафиксированного trade-off или внешней зависимости
- code files — только файлы из `Touches:` активной задачи, нужные для подтверждения

## Do Not Read By Default

- нерелевантные области кода
- широкая история репо
- архивы, если проверка от них явно не зависит

## Stop Conditions

Остановитесь и уточните только если:

- slug неоднозначен
- `tasks.md` отсутствует
- проверка пришлось бы выдумывать факты о реализации
- вывод требует broad repository sweep instead of focused evidence
- the implementation claim cannot be confirmed from the current tasks, plan artifacts, and targeted code inspection

## Rules

- Начинайте с `tasks.md` как entrypoint.
- Если `/.speckeep/scripts/check-verify-ready.*` доступен — запускайте как cheap first pass (slug первый аргумент: `bash ./.speckeep/scripts/check-verify-ready.sh <slug>` или PowerShell `.\.speckeep\scripts\check-verify-ready.ps1 <slug>`). Fallback: `/.speckeep/scripts/verify-task-state.*`. Предпочитайте вывод helper scripts исходникам.
- Treat verify as an evidence log, not a reassurance ritual.
- Проверяйте, что завершённые задачи согласованы с текущим состоянием feature package.
- Проверяйте, что незавершённые задачи не противоречат заявлению о полной готовности.
- Verify acceptance-to-task coverage consistency если в `tasks.md` есть `## Покрытие критериев приемки`.
- Ссылайтесь на task IDs (`T1.1`) напрямую в checks, findings, выводах.
- Предпочитайте подтверждение конкретных claims широкому субъективному review.
- Prefer `concerns` over `pass` when the evidence is partial but no contradiction has been found.

### Traceability

- Запускайте `/.speckeep/scripts/trace.* <slug>` для поиска `@sk-task` и `@sk-test` в коде. Включайте findings в `## Checks`.
- **Legacy fallback**: `trace` ничего не нашёл (старые фичи без аннотаций) — ручная проверка: для каждой завершённой задачи читайте файлы из `Touches:`, подтверждайте описанное изменение; подтверждайте достижимость observable behavior из `AC-*`. Не изобретайте evidence — неподтверждённое идёт в `## Not Verified`. В `## Warnings`: «No `@sk-task` / `@sk-test` annotations found; traceability verified through `Touches:` inspection only.»

### Режимы

- Держите default verification структурным и cheap.
- `--deep` режим:
  - Читайте все plan artifacts (`plan.md`, `data-model.md`, `contracts/`, `research.md`).
  - Для каждой завершённой задачи читайте `Touches:` и подтверждайте соответствие. Не расширяйте scope за пределы `Touches:`, если противоречие этого не требует.
  - Для каждого `AC-*` — ≥ 1 конкретное доказательство по code evidence из `Touches:` привязанных задач. Без исчерпывающей archaeology.
  - `## Scope` должна указывать `mode: deep` и перечислять проверенные surfaces.
  - `## Not Verified` — минимальна или `none`.
- Без `--deep` углубляйтесь только если противоречие нельзя разрешить по tasks + plan + focused evidence.

### Verdict

- Verdicts: `pass`, `concerns`, `blocked`.
  - `pass`: нет блокирующих проблем; только минорные warnings или нет.
  - `concerns`: можно двигаться, но warnings / open questions закрыть в ближайшее время.
  - `blocked`: отсутствие завершения задач или противоречие состояния делают archive/claim небезопасными.
- Do not use `pass` unless the completed task state is confirmed, нет blocking contradiction, и каждый упомянутый claim подкреплён проверенными evidence.

### Отчёт

- Пишите на настроенном языке документации. Используйте `.speckeep/templates/verify.md` как канонический шаблон. Сверху machine-readable metadata block с `report_type`, `slug`, `status`, `docs_language`, `generated_at`.
- Структура: YAML metadata → `# Verify Report: <slug>` → `## Scope` → `## Verdict` → `## Checks` → `## Errors` → `## Warnings` → `## Questions` → `## Not Verified` → `## Next Step`.
- `## Scope`: реальный verification mode и surfaces.
- `## Verdict`: `archive_readiness` и однострочное обоснование.
- `## Checks` явно отражает:
  - `task_state` с completed/open counts
  - `acceptance_evidence` for the `AC-*` items you actually confirmed
  - `implementation_alignment` с конкретной проверенной surface
- `## Not Verified`: material claims/surfaces, которые сознательно не проверяли. `none` только если material gaps нет.
- Keep claims scoped. Проверили только task state + один endpoint/file — так и напишите.

### Recovery

Verify нашёл workflow-gap → верните фичу на самую узкую раннюю фазу:
- `implement` — отсутствующая/противоречивая реализация
- `tasks` — неполная/вводящая в заблуждение/отсутствующая декомпозиция
- `plan` — intent реализации нельзя оценить из-за недостаточно конкретного дизайна

### Next Step

- `pass`: точная archive-команда.
- `concerns`: можно ли дальше; если нет — явная return-команда.
- `blocked`: не предлагайте archive; `Return to: /speckeep.<phase> <slug>` для самой узкой честной recovery-фазы.

## Output

- Отчёт в чат И в `plan/verify.md` по умолчанию (skip только если пользователь просит «только в чат»).
- Суммируйте verdict, проверки, concerns, archive safety.
- Завершайте summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Готово к` / `Return to`.
- Archive-safe: `Готово к: /speckeep.archive <slug>`; возврат — name it explicitly with its slash command.
