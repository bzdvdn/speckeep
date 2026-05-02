# Prompt верификации SpecKeep (compact)

Вы проверяете реализацию одной фичи по `tasks.md` и связанным `AC-*`.

Следуйте базовым правилам в `AGENTS.md`.

## Phase Contract

Inputs: `<specs_dir>/<slug>/plan/tasks.md` (entrypoint), опционально `spec.digest.md` (предпочтительно) или `summary.md`/`spec.md` если нужен AC-контекст, `plan.md` если нужно подтвердить design surfaces.
Outputs: `<specs_dir>/<slug>/plan/verify.md` (или verify report в разговоре) + обновления в `tasks.md`, если выявлены ошибки статуса.
Stop if: `tasks.md` отсутствует или slug неоднозначен.

## Правила

- Treat verify как evidence log (task/AC → proof), не как «ритуал успокоения».
- Начинайте с `tasks.md`: каждая `[x]` задача должна иметь observable proof в репо (файл/тест/выход команды).
- Если доступен `./.speckeep/scripts/verify-task-state.*` — запустите (slug первым аргументом) как cheap first pass.
- Если сохраняете отчёт в файл, используйте `.speckeep/templates/verify.md` как формат и пишите в `<specs_dir>/<slug>/plan/verify.md`. Не ищите «примеры» verify-отчётов в других slug ради формы.
- Соберите traceability как дешёвую проверку целостности: используйте `./.speckeep/scripts/trace.* <slug>` (и при необходимости `./.speckeep/scripts/trace.* <slug> --tests`). Это не заменяет proof, но помогает быстро найти пробелы/осиротевшие метки.
- Глубокое чтение кода — только когда нужно подтвердить конкретный claim по конкретному AC/task.
- Не перечитывайте одни и те же файлы многократно «для уверенности»: ведите краткие evidence-заметки и используйте точечные выборки + `git diff` для проверки изменений.
- Не превращайте verify в редизайн: находите несоответствия и фиксируйте статус/блокеры.
- Формат evidence: перечисляйте строки `<TASK_ID> -> proof` (и `AC-* -> <TASK_ID>`, когда релевантно), где `<TASK_ID>` это `T*` или `<slug>#T*`. Proof должен быть наблюдаемым: путь к файлу, имя/вывод теста, вывод команды, или артефакт в `<specs_dir>/<slug>/...`.
- Гигиена статусов: не переключайте чекбокс на `[x]` без proof. Если evidence отсутствует/неоднозначно — оставляйте `[ ]` и ставьте `concerns`/`blocked` с конкретным next step.
- Держите claims узкими: не делайте broad repository sweep вместо focused evidence. Если claim нельзя подтвердить из tasks + plan artifacts + targeted code inspection — добавьте в `## Not Verified` и не повышайте до `pass`.
- Возвращайте фичу в самую узкую раннюю фазу, которая честно исправит проблему (обычно `tasks` для coverage gaps, `plan` для отсутствующих surfaces/решений, `spec` для intent/AC).
- Если evidence частичное, но противоречий не найдено — предпочитайте `concerns`, а не `pass`.
- `pass` допустим только если подтверждено состояние завершённых задач и нет AC-критичных пробелов.

## Output expectations

- Дайте verdict: `pass|concerns|blocked` + список конкретных несоответствий (task/AC → evidence).
- Добавляйте `## Not Verified`, если что-то не проверяли (явно перечислите, что не подтверждено).
- Если `blocked` — завершите: `Вернуться к: /speckeep.<phase> <slug>`.
- Добавьте короткий summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, и `Готово к` или `Вернуться к`.
- Если `pass` — финальная строка: `Готово к: speckeep archive <slug> .`
