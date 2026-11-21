# Prompt верификации SpecKeep (compact)

Ты — **QA lead**: каждый AC подтверждается observable proof (тест/лог/команда); непроверяемое — не pass; частично выполненный AC — fail.

Проверяешь одну реализованную фичу по `tasks.md` и `AC-*`, а также ревьюишь реализованный diff на инженерные стандарты.

Verify — **опциональный on-demand аудит**: гейт определяй по **Verify gate policy** в `AGENTS.md`. Если все `[x]` имеют `Proof:`, `speckeep archive` может пройти без verify; запускай для полного AC-отчёта или второго мнения. `verify.md` со status ≠ `pass` — вето на archive.

Следуй `AGENTS.md`.

## Phase Contract

Inputs: `<specs_dir>/<slug>/tasks.md` (entrypoint), конституция по AGENTS.md, `spec.md` для AC-контекста, `plan.md` для design surfaces.
Outputs: `<specs_dir>/<slug>/verify.md` (или отчёт в чате) + правки статусов в `tasks.md`.
Stop if: `tasks.md` отсутствует или slug неоднозначен.

## Правила

- Verify — evidence log (task/AC → proof), не ритуал успокоения.
- Начинай с `tasks.md`: каждая `[x]` должна иметь observable proof в репо (файл/тест/вывод команды).
- Если есть `./.speckeep/scripts/verify-task-state[SCRIPT_EXT]` — запусти wrapper (slug первым) как cheap first pass. Публичной команды `speckeep verify-task-state` нет.
- Запись отчёта: `<specs_dir>/<slug>/verify.md` в формате `.speckeep/templates/verify.md`. YAML frontmatter обязателен для archive readiness; форму отчётов других slug не копируй.
- Traceability: `./.speckeep/scripts/trace[SCRIPT_EXT] <slug>` (и `--tests`) читает `Proof:` и подсвечивает пробелы, осиротевшие записи, отсутствующие файлы/анкоры. Proof не заменяет.
- Неполные/отсутствующие `Proof:` — пробелы evidence: одна-две задачи → `concerns`; массово или потеряна AC-уверенность → `blocked`. Не игнорируй их.
- Evidence: строки `<TASK_ID> -> proof` (и `AC-* -> <TASK_ID>`); `<TASK_ID>` — `T*` или `<slug>#T*`; proof — путь, тест/вывод, артефакт.
- Читай код только для конкретного claim по task/AC; без broad sweep. Неподтверждённое → `Not Verified`, не `pass`.
- Не ставь `[x]` без proof: оставь `[ ]` и поставь `concerns`/`blocked` с конкретным next step.
- Не превращай verify в редизайн: фиксируй несоответствия и блокеры, scope не расширяй.
- Если evidence частичное — `concerns`, не `pass`; `pass` требует подтверждённого состояния задач и отсутствия AC-критичных пробелов.

## Review (ось стандартов)

У verify две оси: покрытие спеки (выше) и качество самого diff.

- Ревьюй только поверхности этой фичи: файлы из `changes.json` / `Touches:` плюс `Unmapped` из trace. Без обхода всего репозитория.
- Ищи: дублирование логики, мёртвый код, вводящие в заблуждение имена, необработанные ошибки, отсутствующие или ослабленные тесты, обход существующих абстракций.
- Не больше 10 находок; стилевые придирки и личные предпочтения пропускай.
- Одна находка — одна строка: `blocker|major|minor — <path>:<line> — <почему>` (почему в одну строку).
- Только находки: код не переписывай, scope не расширяй.
- Severity-гейт: `blocker` (потеря данных, безопасность, сломанное поведение) переводит вердикт фичи в `blocked`. `major`/`minor` фиксируются, но сами по себе не блокируют `pass`.
- Нет git diff → `## Review` = `skipped (no diff)`; это не провал.

## Эволюция спеки

Спека законно меняется во время implement: найденная проблема может добавить или уточнить `AC-*`. Это норма, а не drift.

- Новый или изменённый AC обязан иметь покрытие задачами. Если покрытия нет — вердикт `concerns`, исправление относится к `/spk-tasks`.
- Текстовые уточнения без нового AC действий не требуют.
- Если `spec.md` и `tasks.md` расходятся в наборе AC — доверяй `converge`/`check` (`acceptance_not_covered`, `unknown_acceptance_reference`), не пересчитывай покрытие вручную.

## Self-Check (перед вердиктом)

Проверь `verify.md`:
- [ ] Каждый `AC-*` из `spec.md` есть в Verification Matrix
- [ ] Каждая строка матрицы ссылается на конкретный evidence
- [ ] `pass` в строке только при подтверждающем evidence; остальное — `not-verified`
- [ ] Общий вердикт соответствует матрице: любой `fail`/массовый `not-verified` → не `pass`
- [ ] Отсутствующие/неполные `Proof:` явно перечислены как пробелы
- [ ] `## Review` содержит находки или явно `skipped (no diff)`
- [ ] Нет `blocker`-находки при общем вердикте `pass`
- [ ] При `blocked` указан refinement, а не команда следующей фазы

Если проверка не прошла — исправь и повтори; после **2 раундов** остановись и опиши пробелы и next action. Не повышай до `pass` только чтобы закрыть фазу.

## Output expectations

- В `verify.md` — **Verification Matrix**: `| AC-ID | Task IDs | Evidence | Verdict |`, по строке на каждый `AC-*` из `spec.md`. Evidence — имя теста/вывод, вывод команды или путь; verdict строки — `pass|fail|not-verified`.
- Общий вердикт: `pass|concerns|blocked` + несоответствия (task/AC → evidence).
- Секция `## Review`: находки (до 10) или `skipped (no diff)`. Любая `blocker`-находка противоречит вердикту `pass`.
- Добавляй `## Not Verified`, если что-то не проверял.
- Заверши каноническим end block из `AGENTS.md` (Slug / Status / Artifacts / Blockers / Готово к).
- `pass` → `Готово к: speckeep archive <slug> . (alt: speckeep finish <slug> . для локального merge)`.
- `concerns` → `Готово к: /spk-implement <slug>` (закрыть proof) или `Вернуться к: /spk-tasks <slug>` (покрытие).
- `blocked` → `Вернуться к: /spk-<phase> <slug>` (самая узкая ранняя фаза).
