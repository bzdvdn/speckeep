# Prompt конституции SpecKeep (compact)

Вы создаёте или обновляете конституцию проекта.

## Phase Contract

Inputs: запрос пользователя, минимальный контекст репозитория (только то, что нужно для ограничений/архитектуры).
Outputs: `project.constitution_file` (по умолчанию `CONSTITUTION.md`).
Stop if: правила остаются `TBD`/placeholder или конфликтуют с текущим repo reality без явного решения.

## Правила

- Конституция — верхний приоритет: короткие, проверяемые правила; без «философии».
- Укажите: Purpose, принципы, ограничения, tech stack, архитектуру, language policy, workflow.
- Всегда используйте шаблон `.speckeep/templates/constitution.md` как каркас и формат результата. Не ищите «примеры» в чужих конституциях/проектах ради формы: это лишний токен‑расход и дрейф.
- Если есть `./.speckeep/scripts/check-constitution.*` — запустите перед завершением.

## Output expectations

- Запишите/patch конституцию.
- Сгенерируйте `.speckeep/constitution.summary.md` в строгом компактном формате (только правила, без абзацев рассуждений):
  - `Purpose:` одна строка
  - `Non-negotiables:` 3-6 bullets (`MUST` / `MUST NOT`)
  - `Stack/Architecture:` 2-5 bullets
  - `Workflow/DoD:` 3-6 bullets (обязательно traceability и proof-требования)
  - `Repo Map Policy:` 2-4 bullets
  - `Languages:` одна строка (`docs=...`, `agent=...`, `comments=...`)
  - жесткий лимит: ≤200 слов суммарно
- Коротко перечислите ключевые правила и что изменилось.
- Финальная строка: `Готово к: /speckeep.spec <slug>`
