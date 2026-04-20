# Prompt планирования SpecKeep (compact)

Вы создаёте или обновляете plan package для одной фичи: `<specs_dir>/<slug>/plan/`.

Следуйте базовым правилам в `AGENTS.md`.

## Разрешение путей

- Определите `<specs_dir>` из `.speckeep/speckeep.yaml` (читать ≤ 1 раза за сессию). Если конфиг отсутствует — используйте `.speckeep/specs`.

## Phase Contract

Inputs: `.speckeep/constitution.md` (или `.speckeep/constitution.summary.md`), `<specs_dir>/<slug>/spec.md`, `<specs_dir>/<slug>/inspect.md` (должен быть `pass`).
Outputs: `<specs_dir>/<slug>/plan/plan.md` и при необходимости `<specs_dir>/<slug>/plan/data-model.md`, `<specs_dir>/<slug>/plan/contracts/*`, `<specs_dir>/<slug>/plan/research.md`.
Stop if: inspect не `pass`, slug/цель неоднозначны, или план потребовал бы выдумывать требования/AC.

## Правила

- План **не меняет** intent spec: не добавляйте новые большие workstreams вне `spec.md`.
- Фиксируйте только решения, нужные для реализации: surfaces, sequencing, риски/decisions (`DEC-*`).
- Если data model не меняется — всё равно создайте `plan/data-model.md` со stub: `status: no-change` + причина.
- `plan/research.md` создавайте только по необходимости (напр., внешняя зависимость/граница интеграции, несколько реалистичных вариантов, или high-risk unknown). Не создавайте `research.md` для общего брейншторма.
- Минимальный контекст: только текущий slug и нужные repo surfaces; не читайте весь репозиторий.
- Если есть `/.speckeep/scripts/check-plan-ready.*` — запустите (slug первым аргументом) перед записью.

## Output expectations

- Запишите/patch `<specs_dir>/<slug>/plan/plan.md` (и связанные артефакты только если реально нужны).
- Укажите ключевые `DEC-*`, surfaces, sequencing constraints и риски.
- Добавьте короткий summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Готово к`.
- Финальная строка: `Готово к: /speckeep.tasks <slug>`
