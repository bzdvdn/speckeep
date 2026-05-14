# Prompt планирования SpecKeep (compact)

Вы создаёте или обновляете plan artifacts одной фичи в `<specs_dir>/<slug>/`.

Следуйте базовым правилам в `AGENTS.md`.

## Phase Contract

Inputs: `.speckeep/constitution.summary.md` (предпочтительно, если файл есть) или `project.constitution_file` (по умолчанию `CONSTITUTION.md`), `<specs_dir>/<slug>/spec.md`, `<specs_dir>/<slug>/inspect.md` (опционален; если присутствует — должен быть `pass`).
Outputs: `<specs_dir>/<slug>/plan.md`, и при необходимости `<specs_dir>/<slug>/data-model.md`, `<specs_dir>/<slug>/contracts/*`, `<specs_dir>/<slug>/research.md`.
Stop if: inspect.md присутствует и не `pass`, slug/цель неоднозначны, или план потребовал бы выдумывать требования/AC.

## Правила

- План **не меняет** intent spec: не добавляйте новые большие workstreams вне `spec.md`.
- Фиксируйте только решения, нужные для реализации: surfaces, sequencing, риски/decisions (`DEC-*`).
- Всегда используйте шаблон `.speckeep/templates/plan.md` как каркас и формат результата (и при необходимости `.speckeep/templates/data-model.md`). Не ищите «примеры» в соседних feature artifacts других slug: чтение чужих планов ради формы — лишний токен‑расход и scope drift.
- Если data model не меняется — всё равно создайте `data-model.md` со stub: `status: no-change` + причина.
- `research.md` создавайте только по необходимости (напр., внешняя зависимость/граница интеграции, несколько реалистичных вариантов, или high-risk unknown). Не создавайте `research.md` для общего брейншторма.
- Если нужен контекст конституции, сначала загрузите `.speckeep/constitution.summary.md`, если файл существует; только при его отсутствии переходите к `project.constitution_file`.
- Минимальный контекст: только текущий slug и нужные repo surfaces; не читайте весь репозиторий.
- Если есть `./.speckeep/scripts/check-plan-ready.*` — запустите (slug первым аргументом) перед записью.

## Output expectations

- Запишите/patch `<specs_dir>/<slug>/plan.md` (и связанные артефакты только если реально нужны).
- Внутри `plan.md` обязательно держите компактные sections для `DEC-*`, surfaces, рисков, влияния на data model и contracts; не выносите эти recap-элементы в отдельные digest-файлы.
- Укажите ключевые `DEC-*`, surfaces, sequencing constraints и риски.
- Добавьте короткий summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Готово к`.
- Финальная строка: `Готово к: /speckeep.tasks <slug>`
