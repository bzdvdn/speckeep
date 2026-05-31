# Prompt handoff SpecKeep (compact)

Сформируйте короткий handoff по одной фиче.

## Phase Contract

Inputs: текущая фаза (state), `<specs_dir>/<slug>/tasks.md`, последние изменения (файлы/команды, если известны).
Outputs: handoff-summary.
Stop if: tasks.md отсутствует.

## Разрешение путей

- Определите `<specs_dir>` из `.speckeep/speckeep.yaml` (читать ≤ 1 раза за сессию). Если конфиг отсутствует — используйте `specs/active`.

## Output expectations

- `Slug`, `Phase`, `What changed`, `Open tasks`, `Blockers`, `Next command`.
- Финальная строка (определите по состоянию фазы):
  - Если blocked: `Вернуться к: /speckeep.<phase> <slug>`
  - Если готово к следующей фазе: `Готово к: /speckeep.<next> <slug>`
  - Если всё готово: `Готово к: speckeep archive <slug> .`
