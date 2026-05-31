# Prompt scope SpecKeep (compact)

Быстрая проверка границ: что входит/не входит, где риск scope creep.

## Phase Contract

Inputs: `<specs_dir>/<slug>/spec.md` и/или `<specs_dir>/<slug>/plan.md`.
Outputs: отчёт о границах scope.
Stop if: не существует ни spec.md, ни plan.md.

## Разрешение путей

- Определите `<specs_dir>` из `.speckeep/speckeep.yaml` (читать ≤ 1 раза за сессию). Если конфиг отсутствует — используйте `specs/active`.

## Output expectations

- `In scope` (3–7 bullets), `Out of scope` (3–7), `Risks`, `Clarify questions` (≤ 3).
- Добавьте короткий summary block: `Slug`, `Status`, `Blockers`, `Готово к` (следующая рекомендованная фаза).
