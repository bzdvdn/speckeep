# Prompt задач SpecKeep (compact)

Вы создаёте или обновляете `<specs_dir>/<slug>/tasks.md`.

Следуйте базовым правилам в `AGENTS.md`.

## Phase Contract

Inputs: `.speckeep/constitution.summary.md` (предпочтительно, если файл есть) или `project.constitution_file` (по умолчанию `CONSTITUTION.md`), `<specs_dir>/<slug>/plan.digest.md` (предпочтительно) или `plan.md`, опционально `spec.digest.md` (предпочтительно) или `summary.md`/`spec.md` если нужно уточнить `AC-*`.
Outputs: `tasks.md` с фазами, `Touches:` для каждой задачи, `## Surface Map`, и `## Покрытие критериев приемки` (AC → tasks).
Stop if: `plan.md` отсутствует/расплывчат или хотя бы один `AC-*` нельзя привязать к исполнимым задачам без догадок.

## Правила

- Делайте **минимальный** список задач, достаточный для выполнения фичи по плану.
- Каждая задача: измеримый outcome + явные `Touches:` (файлы/модули). Без `Touches:` — дефект.
- Перед фазами обязателен `## Surface Map` (Surface | Tasks) для batch-reads на implement.
- Не ищите «примеры» в соседних спеках/тасках других slug: это почти всегда лишний токен‑расход и scope drift. Форму/структуру берите из шаблона `.speckeep/templates/tasks.md` и текущего `<specs_dir>/<slug>/plan.md`.
- Делайте `tasks.md` самодостаточным для implement: implement-агент должен выполнять задачи, читая только `tasks.md` + файлы из `Touches:` активной задачи (без обязательного reread `plan.md`/`spec.md`/`data-model.md`).
- Если для выполнения нужны ключевые решения/инварианты из plan/data-model, вынесите их в короткий раздел `## Implementation Context` (≤ ~20 строк) и ссылайтесь на них из задач (например `DEC-*` / `DM`), чтобы implement не перечитывал исходные артефакты целиком.
- Рекомендуемый шаблон `## Implementation Context` (держать коротким, без воды):
  - `Цель MVP:` (1 строка)
  - `Инварианты/семантика:` (2–5 bullets)
  - `Ошибки/коды:` (1–3 bullets)
  - `Контракты/протокол:` (1–3 bullets: пути/форматы)
  - `Границы scope:` (2 bullets “не делаем …”)
  - `Proof signals:` (1–3 bullets “что считаем доказательством”)
  - `References (опц.):` `DEC-*`, `DM`, `RQ-*` (без обязательного перечтения исходников)
- Каждый `AC-*` должен быть покрыт ≥ 1 задачей: `AC-001 -> T1.1, T2.1`.
- Не начинайте implementation и не редактируйте исходный код на фазе tasks.
- Не считайте, что `research.md` обязан существовать; ссылайтесь на него только если план явно от него зависит.
- Если нужен контекст конституции, сначала загрузите `.speckeep/constitution.summary.md`, если файл существует; только при его отсутствии переходите к `project.constitution_file`.
- Если есть `./.speckeep/scripts/check-tasks-ready.*` — запустите (slug первым аргументом) и используйте вывод как cheap gate.

## Output expectations

- Запишите/patch `tasks.md` (не переписывайте лишнее при небольших правках).
- Коротко суммируйте: фазы, основные surfaces, покрытие AC, blockers.
- Добавьте короткий summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Готово к`.
- Финальная строка: `Готово к: /speckeep.implement <slug>`
