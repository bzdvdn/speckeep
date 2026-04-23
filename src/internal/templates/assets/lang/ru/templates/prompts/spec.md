# Prompt спецификации SpecKeep (compact)

Вы создаёте или обновляете одну feature specification: `<specs_dir>/<slug>/spec.md`.

## Разрешение путей

- Определите `<specs_dir>` из `.speckeep/speckeep.yaml` (читать ≤ 1 раза за сессию). Если конфиг отсутствует — используйте `specs`.

## Phase Contract

Inputs: `project.constitution_file` (по умолчанию `CONSTITUTION.md`, или `.speckeep/constitution.summary.md` если есть), запрос пользователя, минимально нужный контекст репозитория.
Outputs: `<specs_dir>/<slug>/spec.md`, `<specs_dir>/<slug>/spec.digest.md`.
Stop if: запрос неоднозначен/мульти-фича или пришлось бы выдумывать `AC-*`.

## Обязательные правила

- **Branch-first**: до записи любого файла переключитесь/создайте ветку `feature/<slug>` (или `--branch`). Если это невозможно — стоп и объяснить.
- Не пытайтесь «сгенерировать spec через CLI»: вы как агент должны сами записать/обновить файл `<specs_dir>/<slug>/spec.md`.
  - Команды `speckeep spec` не существует. Не запускайте `./.speckeep/scripts/run-speckeep.* spec <slug>`.
  - Скрипты `./.speckeep/scripts/check-*.{sh,ps1}` — только для проверок (gate), а не для генерации артефактов.
- Не ищите «примеры» в соседних спеках других slug: форма/структура берётся из шаблона `.speckeep/templates/spec.md` и требований пользователя. Чтение чужих спек ради стиля — лишний токен‑расход и scope drift.
- Spec описывает intent, а не план/задачи. Никаких implementation steps и декомпозиции.
- Каждый `AC-*` — Given/When/Then, наблюдаемый outcome (proof signal в Then).
- Обязательны: `Вне scope`, `Допущения`, `Открытые вопросы` (или `none`).
- Минимальный clarify-pass: 1–3 точечных вопроса только если иначе придётся выдумывать AC или размоется scope.
- Если вызвано с `--name` без достаточного описания — запросите его и считайте следующее сообщение (не начинающееся с `/speckeep.`) продолжением. Сообщение, начинающееся с `/speckeep.`, отменяет staged mode.
- Не фиксируйте технологии/версии, если это не требование пользователя или жёсткий repo/contract constraint. Если это лишь implementation preference — фиксируйте в `plan`, не в `spec`.
- Вместо догадок — refinement: если запрос подразумевает несколько feature slug или несколько независимых спецификаций, остановитесь и попросите выбрать одну фичу.
- Если есть `./.speckeep/scripts/check-spec-ready.*` — запустите (slug первым аргументом) перед завершением.

## Output expectations

- Запишите/patch `spec.md` (patch > переписывание).
- Запишите `<specs_dir>/<slug>/spec.digest.md`: одна строка на каждый `AC-*`, формат `AC-NNN: <краткое описание ≤10 слов>`. Никакой детализации.
- Коротко суммируйте цель, scope, список AC, открытые вопросы/блокеры.
- Добавьте короткий summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Готово к`.
- Финальная строка: `Готово к: /speckeep.inspect <slug>`
