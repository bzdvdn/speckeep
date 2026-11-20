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
- При записи `verify.md` обязателен YAML frontmatter (формат — `.speckeep/templates/verify.md`; нужен для archive readiness).
- Начинай с `tasks.md`: каждая `[x]` должна иметь observable proof в репо.
- Если есть `./.speckeep/scripts/verify-task-state[SCRIPT_EXT]` — запусти wrapper (slug первым) как cheap first pass. Публичной команды `speckeep verify-task-state` нет.
- Формат отчёта — `.speckeep/templates/verify.md`; отчёты других slug не читай.
- Traceability: `./.speckeep/scripts/trace[SCRIPT_EXT] <slug>` (и `--tests` при необходимости) читает `Proof:` и подсвечивает пробелы/осиротевшие записи; proof не заменяет.
- Неполные/отсутствующие `Proof:` — пробелы evidence, не косметика.
- Глубокое чтение кода — только для конкретного claim по task/AC.
- Не перечитывай файлы многократно; веди evidence-заметки, используй точечные выборки и `git diff`.
- Не превращай verify в редизайн: фиксируй несоответствия и статус.
- Evidence: строки `<TASK_ID> -> proof` (и `AC-* -> <TASK_ID>`); proof — путь, тест/вывод, артефакт.
- Не ставь `[x]` без proof. Нет/неоднозначно evidence → `[ ]` + `concerns`/`blocked` с next step.
- Одна-две задачи без `Proof:` при остальном evidence → обычно `concerns`; массовые пробелы → `blocked`.
- Claims держи узкими: без broad sweep; неподтверждённое — `## Not Verified`, не `pass`.
- Возвращай фичу в самую узкую раннюю фазу: `tasks` (coverage), `plan` (surfaces), `spec` (intent/AC).
- Если evidence частичное без противоречий — `concerns`, не `pass`.
- `pass` — только при подтверждённом состоянии задач и без AC-критичных пробелов.

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
- [ ] `pass` только при подтверждённом состоянии задач и без AC-критичных пробелов
- [ ] Отсутствующие/неполные `Proof:` явно перечислены как пробелы
- [ ] `## Review` содержит находки или явно `skipped (no diff)`
- [ ] Нет `blocker`-находки при общем вердикте `pass`
- [ ] При `blocked` указан refinement, а не команда следующей фазы

Если проверка не прошла — исправь и повтори; после **2 раундов** остановись и опиши пробелы и next action. Не повышай до `pass` только чтобы закрыть фазу.

## Output expectations

- В `verify.md` — таблица **Verification Matrix**:
  ```
  | AC-ID | Task IDs | Evidence | Verdict |
  |-------|----------|----------|---------|
  | AC-001 | T1.1, T2.1 | test_export_downloads_file: pass | pass |
  | AC-002 | T1.2 | test_export_empty_state: pass | pass |
  ```
  - Каждый AC-* из spec.md — строка матрицы.
  - Evidence: имена тестов, вывод команд или пути.
  - Verdict строки: `pass` | `fail` | `not-verified`.
- Общий вердикт: `pass|concerns|blocked` + несоответствия (task/AC → evidence).
- Секция `## Review`: список находок (до 10) или `skipped (no diff)`. Любая `blocker`-находка противоречит вердикту `pass`.
- Добавляй `## Not Verified`, если что-то не проверял.
- Явно перечисляй пробелы `Proof:` у завершённых задач.
- Заверши каноническим end block из `AGENTS.md` (Slug / Status / Artifacts / Blockers / Готово к).
- `pass` → `Готово к: speckeep archive <slug> . (alt: speckeep finish <slug> . для локального merge)`.
- `concerns` → `Готово к: /spk-implement <slug>` (закрыть proof) или `Вернуться к: /spk-tasks <slug>` (покрытие).
- `blocked` → `Вернуться к: /spk-<phase> <slug>` (самая узкая ранняя фаза).
