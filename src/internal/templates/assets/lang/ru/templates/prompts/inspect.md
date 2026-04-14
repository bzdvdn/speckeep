# Prompt проверки SpecKeep

Вы проверяете один пакет фичи на согласованность и качество.

## Goal

Сформируйте сфокусированный отчет проверки для одной фичи, не расширяя scope.

## Примечание о путях

Пути в этом промпте показаны для layout по умолчанию. Если в `.speckeep/speckeep.yaml` переопределены `paths.specs_dir` или `project.constitution_file`, всегда следуйте путям из конфигурации, а не примерам по умолчанию.
Читайте `.speckeep/speckeep.yaml` максимум один раз за сессию для резолва путей; не перечитывайте его без необходимости (только если конфиг изменился или путь неоднозначен).

## Phase Contract

Inputs: `.speckeep/constitution.md`, `.speckeep/specs/<slug>/spec.md`; опционально `.speckeep/specs/<slug>/plan/plan.md`, `.speckeep/specs/<slug>/plan/tasks.md`, если они существуют.
Outputs: `.speckeep/specs/<slug>/inspect.md` с verdict `pass`, `concerns` или `blocked`.
Stop if: slug неоднозначен, spec отсутствует, или отчет потребовал бы выдумывать продуктовый intent.

## Flags

`--delta`: режим инкрементальной перепроверки — проверяйте только секции, которые изменились с последнего inspect report, вместо полной проверки.

Когда `--delta` присутствует в аргументах пользователя:
- Сначала прочитайте существующий `.speckeep/specs/<slug>/inspect.md` для установления baseline.
- Сравните текущий `spec.md` с предыдущим inspect report, чтобы определить изменённые секции (новые или модифицированные AC, изменения scope, изменения assumptions).
- Перепроверьте только изменённые секции и их cross-artifact implications.
- Сохраните findings из предыдущего отчёта, которые всё ещё валидны; не пере-derive их.
- Обновите verdict только если delta его меняет. Если предыдущий `blocked` finding разрешён и нет новых ошибок, повысьте до `pass` или `concerns`.
- Обновите `generated_at` timestamp и добавьте поле `delta_from: <previous_generated_at>` в metadata block.
- Если delta затрагивает `## Goal`, `## Scope` или более половины `AC-*` записей в спеке, считайте изменение обширным и откатитесь к полной проверке с пометкой: «Delta mode откатился к полной проверке из-за масштаба изменений.»
- Если delta меняет хотя бы один `AC-*`, перегенерируйте `summary.md` после обновления `inspect.md`. Если AC не изменились — оставьте `summary.md` без изменений.

## Load First

Всегда сначала прочитайте:

- `.speckeep/constitution.md`
- `.speckeep/specs/<slug>/spec.md`

## Load If Present

Читайте только если файл существует и inspect требует cross-artifact consistency checks (выравнивание spec↔plan, покрытие acceptance↔tasks):

- `.speckeep/specs/<slug>/plan/plan.md` — читать при проверке goal alignment, scope expansion или acceptance coverage на уровне плана
- `.speckeep/specs/<slug>/plan/tasks.md` — читать при проверке, что каждый `AC-*` покрыт хотя бы одной задачей

## Do Not Read By Default

- `.speckeep/specs/<slug>/plan/data-model.md`
- `.speckeep/specs/<slug>/plan/contracts/`
- `.speckeep/specs/<slug>/plan/research.md`
- широкую историю репозитория
- implementation-файлы, если только finding не ссылается на конкретный файл и вывод нельзя подтвердить из spec/plan/tasks

## Stop Conditions

Остановитесь и задайте минимальный уточняющий вопрос только если:

- целевой slug неоднозначен
- сама спецификация отсутствует
- без этого пришлось бы выдумывать продуктовый intent

## Rules

- Сначала проверяйте соответствие конституции.
- Если доступен `/.speckeep/scripts/check-inspect-ready.*`, предпочитайте его как cheap first pass перед углублением в артефакты.
- Важно: readiness wrapper запускается со slug как первым аргументом. Пример: `bash ./.speckeep/scripts/check-inspect-ready.sh <slug>` (или PowerShell: `.\.speckeep\scripts\check-inspect-ready.ps1 <slug>`).
- Используйте `/.speckeep/scripts/inspect-spec.*` только как fallback, когда phase readiness wrapper недоступен.
- Если оба скрипта недоступны, выполняйте полную ручную проверку напрямую по `constitution.md` и `spec.md` — не останавливайтесь и не ждите скриптов.
- Предпочитайте вывод helper scripts чтению их исходников.
- Считайте вывод helper scripts основным слоем структурных доказательств для inspect. Если скрипты уже сообщили конкретные `ERROR` / `WARN` findings, используйте их как стартовую основу отчёта, а не переизобретайте те же выводы заново.
- Если helper output показывает категории findings вроде structure, traceability, ambiguity, consistency или readiness, сохраняйте этот сигнал в reasoning. Расширяйте его только когда действительно нужен дополнительный контекст.
- Не игнорируйте конкретный helper finding только потому, что у вас есть более оптимистичное общее впечатление. Его нужно либо закрыть, либо явно объяснить.
- Собственное reasoning используйте в первую очередь для того, что cheap checks не могут доказать напрямую: конфликтов с конституцией, выдуманного product intent, необоснованного scope expansion, противоречивых assumptions или тонкого drift между `spec` и `plan`.
- Не читайте `/.speckeep/scripts/*` по умолчанию, если только не отлаживаете сам script, не работаете над самим SpecKeep или пользователь явно не просит проанализировать script logic.
- Проверяйте полноту и ясность спецификации.
- Проверяйте `constitution <-> spec`: спецификация не должна противоречить явным ограничениям конституции, workflow-правилам или language policy.
- Считайте technology names, framework choices, library lists или version pins в спецификации `Warning`, если они явно не выглядят как user requirement, repository constraint или внешний compatibility contract.
- Каждый критерий приемки в спецификации ДОЛЖЕН иметь явный формат Given/When/Then. Маркеры `Given`, `When`, `Then` остаются каноническими независимо от языка документации. Отсутствие G/W/T — `Error`, а не `Suggestion`.
- Любой оставшийся маркер `[NEEDS CLARIFICATION: ...]` в spec — это `Error`. Они должны быть закрыты до начала planning.
- Если `## Assumptions` / `## Допущения` отсутствует, это `Warning`. Если присутствует, проверяйте каждое допущение на правдоподобность по конституции и известному состоянию репозитория — допущение, противоречащее реальности репозитория, это `Error`.
- Если `## Success Criteria` / `## Критерии успеха` присутствует, каждый `SC-*` должен иметь измеримую метрику и метод измерения. Размытые SC (напр., «система должна быть быстрой») — `Warning`.
- Если `tasks.md` существует, проверяйте, что каждый критерий приемки из spec покрыт хотя бы одной задачей. Непокрытый критерий — `Error`.
- Если `tasks.md` существует, содержит task IDs, но не имеет `## Surface Map` — это `Warning`: implement-агенту нужна эта секция как манифест batch-чтения.
- Если `tasks.md` существует и любая строка задачи с task ID не содержит поле `Touches:` — это `Warning`: задачи без `Touches:` вынуждают implement-агента к exploratory reads.
- Если `tasks.md` использует task IDs вроде `T1.1`, предпочитайте traceability-формулировки с прямыми ссылками на эти task IDs.
- Предпочитайте самый дешевый inspection scope: `constitution.md` и `spec.md`, затем `plan.md`, затем `tasks.md`, и только после этого более глубокие plan artifacts, если они нужны для подтверждения конкретного вывода.
- Если `plan.md` отсутствует, не расширяйте проверку на optional plan artifacts или implementation code.
- Если артефакты планирования уже существуют, проверяйте согласованность между spec, plan, data model, contracts и tasks.
- Когда существует `plan.md`, сначала проверяйте `spec <-> plan` consistency, не читая более глубокие plan artifacts без необходимости.
- Проверяйте `spec <-> plan`: план должен сохранять цель фичи, отражать major acceptance-critical behavior и не добавлять необоснованные новые workstreams.
- Если `tasks.md` существует, проверяйте `plan <-> tasks`: фазы и task IDs должны отражать intent плана без явных пропусков по acceptance-critical behavior.
- Считайте `spec.md` и `plan.md` обязательными входами для дешевой проверки согласованности плана.
- Читайте `data-model.md` или `contracts/` только если `plan.md` явно на них опирается или без них нельзя подтвердить конкретный consistency claim.
- Проверяйте `Goal Alignment`: plan не должен менять основную цель фичи, выраженную в spec.
- Проверяйте `Scope Expansion`: plan не должен вводить крупные новые workstreams, компоненты или integration surfaces, которых нет в spec.
- Проверяйте `Acceptance Coverage at Plan Level`: major acceptance-critical behavior из spec должно быть отражено в намерении плана, даже до появления tasks.
- Проверяйте `Constitution Consistency`: plan не должен нарушать правила конституции или архитектурные ограничения.
- Если `plan.md` существует и не содержит `## Соответствие конституции` / `## Constitution Compliance` — это `Warning`: эта секция делает соответствие конституции явным и проверяемым.
- Проверяйте `Artifact Justification`: если plan вводит `data-model.md` или `contracts/`, необходимость этих артефактов должна быть оправдана spec.
- Не превращайте это в широкий design review. Предпочитайте ловить явный drift, а не оценивать качество архитектуры целиком.
- Если записываете отчет в файл, держите его на настроенном языке документации проекта.
- Предпочитайте конкретные находки вместо общих советов.
- Предпочитайте такой порядок формирования отчёта:
  - 1. структурные findings из helper output
  - 2. cross-artifact consistency findings, подтверждённые загруженными артефактами
  - 3. узкие выводы, которые действительно требуют agent reasoning
- По умолчанию делайте compact report в разговоре: всегда включайте `Verdict`, включайте `Errors`, `Warnings` и `Next Step`, если они не пусты, а `Questions`, `Suggestions` и `Traceability` — только когда они действительно добавляют сигнал.
- Полный sectioned report используйте только если пользователь явно просит полный отчет или если отчет сохраняется в файл.
- Если отчет сохраняется в файл, добавляйте сверху machine-readable metadata block с полями `report_type`, `slug`, `status`, `docs_language` и `generated_at`.
- Используйте такую структуру отчета:
  - YAML-подобный metadata block в начале
  - `# Inspect Report: <slug>`
  - `## Scope`
  - `## Verdict`
  - `## Errors`
  - `## Warnings`
  - `## Questions`
  - `## Suggestions`
  - `## Traceability`
  - `## Next Step`
- В секции `## Verdict` ДОЛЖНО использоваться одно из значений: `pass`, `concerns`, `blocked`.
- Используйте `pass`, когда ошибок нет и остаются только незначительные предупреждения или предупреждений нет совсем.
- Используйте `concerns`, когда по фиче можно двигаться дальше, но warnings, пробелы traceability или открытые вопросы желательно закрыть в ближайшее время.
- Используйте `blocked`, если конфликт с конституцией, отсутствие продуктового intent, отсутствие Given/When/Then в acceptance criteria, непокрытые acceptance criteria или крупные противоречия между `spec` и `plan` не позволяют безопасно продолжать следующую фазу.
- `## Traceability` должна кратко показывать, как acceptance criteria связаны с задачами, если `tasks.md` уже существует.
- Предпочитайте traceability-строки со стабильными acceptance IDs и task IDs, например `AC-001 -> T1.1, T2.1`.
- `## Next Step` должна явно говорить, можно ли безопасно продолжать к `plan`, `tasks`, или сначала нужно уточнение.
- Для `pass` указывайте точную следующую slash-команду.
- Для `concerns` явно говорите, можно ли двигаться дальше; если можно, указывайте точную следующую slash-команду.
- Для `blocked` не подсказывайте следующую фазу; вместо этого указывайте, какой refinement нужен сначала.
- Не дублируйте одну и ту же проблему в нескольких секциях. Если helper output уже зафиксировал конкретную проблему, формулируйте её кратко и переходите к следствию или нужному refinement.

## Артефакт краткого описания спецификации

После записи отчёта проверки также запишите `.speckeep/specs/<slug>/summary.md`.

Summary ДОЛЖЕН содержать только:

- YAML frontmatter с полями `slug` и `generated_at`
- `## Goal` — одно предложение
- `## Acceptance Criteria` — таблица: `ID | Summary | Proof Signal`; summary ≤ 8 слов; proof signal = наблюдаемая проверка из `Then`-клаузы
- `## Out of Scope` — 3-5 bullets

Держите summary не длиннее 25 строк. Он загружается в `tasks`, `implement` и `verify` вместо полного spec.md, чтобы снизить context overhead. Summary не заменяет полный spec на фазах, требующих полной inspect-проверки критериев (inspect, plan).

## Output expectations

- Сохраняйте отчет в `.speckeep/specs/<slug>/inspect.md` и записывайте `.speckeep/specs/<slug>/summary.md`; кратко суммируйте verdict в разговоре compact report с непустыми секциями
- Завершайте разговор summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Готово к`
- Когда можно продолжать: в `## Next Step` указывайте точную slash-команду следующей фазы; после archive можно упомянуть `/speckeep.recap` как опциональный итог, но не рассматривайте его как обязательный
- Когда verdict позволяет продолжать: завершайте также строкой `Готово к: /speckeep.plan <slug>` (или `Готово к: /speckeep.tasks <slug>`, если plan package уже существует)
- Если сначала нужен refinement — говорите об этом прямо

## Self-Check

- Я проверил каждый AC на наличие формата Given/When/Then?
- Verdict (`pass`, `concerns`, `blocked`) опирается на конкретные находки, а не общие впечатления?
- Если `tasks.md` существует, я убедился, что каждый AC покрыт хотя бы одной задачей?
