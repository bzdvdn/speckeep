# Prompt архивации SpecKeep (compact)

Вы архивируете один feature package в `.speckeep/archive/`.

## Разрешение путей

- Определите `<specs_dir>` из `.speckeep/speckeep.yaml` (читать ≤ 1 раза за сессию). Если конфиг отсутствует — используйте `.speckeep/specs`.

## Phase Contract

Inputs: `<specs_dir>/<slug>/` (spec/inspect/plan/tasks/verify по наличию).
Outputs: snapshot в `.speckeep/archive/<slug>/...` (move-based по умолчанию).
Stop if: не завершён verify или статус архивации не может быть обоснован.

## Правила

- Перед архивом предпочтительно запустить `/.speckeep/scripts/check-archive-ready.*` (slug первым аргументом).
- Default status: `completed`. Нестандартные статусы требуют явного `--reason`.
- Архив — это фиксация состояния, не место для новых правок реализации.

## Output expectations

- Создайте snapshot; кратко перечислите перемещённые артефакты и итоговый статус.
- Это терминальный шаг workflow для этой фичи (после verify).
- Финальная строка: `Готово к: /speckeep.recap`
