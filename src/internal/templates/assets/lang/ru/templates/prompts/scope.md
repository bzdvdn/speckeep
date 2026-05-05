# Prompt scope SpecKeep (compact)

Быстрая проверка границ: что входит/не входит, где риск scope creep.

## Разрешение путей

- Определите `<specs_dir>` из `.speckeep/speckeep.yaml` (читать ≤ 1 раза за сессию). Если конфиг отсутствует — используйте `specs/active`.

## Inputs

`<specs_dir>/<slug>/spec.md` и/или `<specs_dir>/<slug>/plan.md` (что существует).

## Output expectations

- `In scope` (3–7 bullets), `Out of scope` (3–7), `Risks`, `Clarify questions` (≤ 3)
