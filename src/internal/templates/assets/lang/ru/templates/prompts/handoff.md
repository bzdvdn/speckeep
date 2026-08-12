# Prompt передачи SpecKeep (compact)

Вы действуете как **senior engineer, передающий работу новой сессии**. Пишите настолько точно, чтобы следующая сессия продолжила без догадок.

Сформируйте короткий handoff по одной фиче.

## Phase Contract

Inputs: текущая фаза (state), `<specs_dir>/<slug>/tasks.md`, последние изменения (файлы/команды, если известны).
Outputs: handoff-summary.
Stop if: tasks.md отсутствует.

## Разрешение путей

- Определите `<specs_dir>` из `.speckeep/speckeep.yaml` (читать ≤ 1 раза за сессию). Если конфиг отсутствует — используйте `specs/active`.

## Output expectations

- `Slug`, `Phase`, `What changed`, `Open tasks`, `Blockers`, `Next command`.
- Финальная строка (определите фазу по состоянию; `workflow.verify` — по **Verify gate policy** в AGENTS.md):
  - Если blocked: `Вернуться к: /spk.<phase> <slug>`
  - Если готово к следующей фазе: `Готово к: /spk.<next> <slug>`
  - Если всё готово и `workflow.verify: required`: `Готово к: /spk.verify <slug>`
  - Если всё готово и `workflow.verify` — `optional`/отсутствует: `Готово к: speckeep archive <slug> .`
