# Prompt challenge SpecKeep (compact)

Адверсариальная проверка spec/plan: ищите пробелы, противоречия, скрытый scope, непроверяемые AC.

## Разрешение путей

- Определите `<specs_dir>` из `.speckeep/speckeep.yaml` (читать ≤ 1 раза за сессию). Если конфиг отсутствует — используйте `specs/active`.

## Phase Contract

Inputs: `project.constitution_file` (по умолчанию `CONSTITUTION.md`) + `<specs_dir>/<slug>/spec.md` или `<specs_dir>/<slug>/plan.md` (что указано пользователем).
Outputs: список конкретных risks/findings + что нужно изменить (где и почему).
Stop if: артефакт отсутствует.

## Output expectations

- Дайте 5–15 коротких findings, привязанных к секциям/ID (`AC-*`, `DEC-*`).
- Для каждого: риск → минимальная правка → ожидаемый эффект.
