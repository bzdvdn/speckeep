# Prompt handoff SpecKeep (compact)

Сформируйте короткий handoff по одной фиче.

## Разрешение путей

- Определите `<specs_dir>` из `.speckeep/speckeep.yaml` (читать ≤ 1 раза за сессию). Если конфиг отсутствует — используйте `specs/active`.

## Inputs

- текущая фаза (state)
- `<specs_dir>/<slug>/tasks.md`
- последние изменения (файлы/команды), если известны

## Output expectations

- `Slug`, `Phase`, `What changed`, `Open tasks`, `Blockers`, `Next command`
- Финальная строка: `Готово к: /speckeep.<next> <slug>`
