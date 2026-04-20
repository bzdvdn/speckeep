# Prompt задач SpecKeep (compact)

Вы создаёте или обновляете `<specs_dir>/<slug>/plan/tasks.md`.

Следуйте базовым правилам в `AGENTS.md`.

## Разрешение путей

- Определите `<specs_dir>` из `.speckeep/speckeep.yaml` (читать ≤ 1 раза за сессию). Если конфиг отсутствует — используйте `.speckeep/specs`.

## Phase Contract

Inputs: `.speckeep/constitution.md` (или `.speckeep/constitution.summary.md`), `<specs_dir>/<slug>/plan/plan.md`, опционально `summary.md`/`spec.md` если нужно уточнить `AC-*`.
Outputs: `tasks.md` с фазами, `Touches:` для каждой задачи, `## Surface Map`, и `## Покрытие критериев приемки` (AC → tasks).
Stop if: `plan.md` отсутствует/расплывчат или хотя бы один `AC-*` нельзя привязать к исполнимым задачам без догадок.

## Правила

- Делайте **минимальный** список задач, достаточный для выполнения фичи по плану.
- Каждая задача: измеримый outcome + явные `Touches:` (файлы/модули). Без `Touches:` — дефект.
- Перед фазами обязателен `## Surface Map` (Surface | Tasks) для batch-reads на implement.
- Каждый `AC-*` должен быть покрыт ≥ 1 задачей: `AC-001 -> T1.1, T2.1`.
- Не начинайте implementation и не редактируйте исходный код на фазе tasks.
- Не считайте, что `research.md` обязан существовать; ссылайтесь на него только если план явно от него зависит.
- Если есть `/.speckeep/scripts/check-tasks-ready.*` — запустите (slug первым аргументом) и используйте вывод как cheap gate.

## Output expectations

- Запишите/patch `tasks.md` (не переписывайте лишнее при небольших правках).
- Коротко суммируйте: фазы, основные surfaces, покрытие AC, blockers.
- Добавьте короткий summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Готово к`.
- Финальная строка: `Готово к: /speckeep.implement <slug>`
