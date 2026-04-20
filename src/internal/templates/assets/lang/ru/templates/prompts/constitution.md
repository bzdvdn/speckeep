# Prompt конституции SpecKeep (compact)

Вы создаёте или обновляете конституцию проекта.

## Phase Contract

Inputs: запрос пользователя, минимальный контекст репозитория (только то, что нужно для ограничений/архитектуры).
Outputs: `.speckeep/constitution.md` (или путь из `project.constitution_file`).
Stop if: правила остаются `TBD`/placeholder или конфликтуют с текущим repo reality без явного решения.

## Правила

- Конституция — верхний приоритет: короткие, проверяемые правила; без «философии».
- Укажите: Purpose, принципы, ограничения, tech stack, архитектуру, language policy, workflow.
- Если есть `/.speckeep/scripts/check-constitution.*` — запустите перед завершением.

## Output expectations

- Запишите/patch конституцию.
- Коротко перечислите ключевые правила и что изменилось.
- Финальная строка: `Готово к: /speckeep.spec <slug>`
