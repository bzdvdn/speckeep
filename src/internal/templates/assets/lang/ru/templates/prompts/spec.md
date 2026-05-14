# Prompt спецификации SpecKeep (compact)

Вы создаёте или обновляете одну feature specification: `<specs_dir>/<slug>/spec.md`.

## Phase Contract

Inputs: `.speckeep/constitution.summary.md` (предпочтительно, если файл есть) или `project.constitution_file` (по умолчанию `CONSTITUTION.md`), запрос пользователя, минимально нужный контекст репозитория.
Outputs: `<specs_dir>/<slug>/spec.md`.
Stop if: запрос неоднозначен/мульти-фича или пришлось бы выдумывать `AC-*`.

## Обязательные правила

- **Branch-first**: до записи любого файла переключитесь/создайте ветку `feature/<slug>` (или `--branch`). Если это невозможно — стоп и объяснить.
- Не пытайтесь «сгенерировать spec через CLI»: вы как агент должны сами записать/обновить файл `<specs_dir>/<slug>/spec.md`.
  - Команды `speckeep spec` не существует. Не запускайте `./.speckeep/scripts/run-speckeep.* spec <slug>`.
  - Скрипты `./.speckeep/scripts/check-*.{sh,ps1}` — только для проверок (gate), а не для генерации артефактов.
- Не читайте `<specs_dir>/*/spec.md` других slug ни по какой причине — ни ради стиля, ни ради формата, ни ради примеров. Не листайте и не сканируйте `<specs_dir>/` чтобы посмотреть существующие slug. Шаблон `.speckeep/templates/spec.md` — единственный структурный ориентир; одного прочтения достаточно.
- Spec описывает intent, а не план/задачи. Никаких implementation steps и декомпозиции.
- Каждый `AC-*` — Given/When/Then, наблюдаемый outcome (proof signal в Then).
- Обязательны: `Вне scope`, `Допущения`, `Открытые вопросы` (или `none`).
- Минимальный clarify-pass: 1–3 точечных вопроса только если иначе придётся выдумывать AC или размоется scope.
- Если вызвано с `--name` без достаточного описания — запросите его и считайте следующее сообщение (не начинающееся с `/speckeep.`) продолжением. Сообщение, начинающееся с `/speckeep.`, отменяет staged mode.
- Если нужен контекст конституции, сначала загрузите `.speckeep/constitution.summary.md`, если файл существует; только при его отсутствии переходите к `project.constitution_file`.
- Не фиксируйте технологии/версии, если это не требование пользователя или жёсткий repo/contract constraint. Если это лишь implementation preference — фиксируйте в `plan`, не в `spec`.
- Вместо догадок — refinement: если запрос подразумевает несколько feature slug или несколько независимых спецификаций, остановитесь и попросите выбрать одну фичу.
- Если есть `./.speckeep/scripts/check-spec-ready.*` — запустите (slug первым аргументом) перед завершением.

## Output expectations

- Запишите/patch `spec.md` (patch > переписывание).
- Самопроверка перед завершением: в spec.md нет `TODO`/`???`/`<placeholder>`; каждый AC содержит Given/When/Then с observable proof; секции Out of Scope, Assumptions, Open Questions существуют.
- Коротко суммируйте цель, scope, список AC, открытые вопросы/блокеры в ответе, не создавая отдельные derived-файлы только ради recap.
- Добавьте короткий summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Готово к`.
- Следующие шаги (предложите оба варианта):
  - Глубокая проверка качества: `/speckeep.inspect <slug>` — проверяет соответствие конституции, полноту AC, неоднозначности
  - Перейти к плану если спека выглядит хорошо: `/speckeep.plan <slug>`
- Финальная строка (обязательно): `Готово к: /speckeep.inspect <slug>` или `Готово к: /speckeep.plan <slug>`. Предпочитайте `/speckeep.inspect`, если остались неоднозначности, риски или открытые вопросы.
