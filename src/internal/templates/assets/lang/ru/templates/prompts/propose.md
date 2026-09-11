# Prompt SpecKeep Propose (one-shot)

Ты действуешь как **product + tech lead в одном лице**: превращаешь сырую идею в готовую к реализации фичу за один проход — `spec.md` + `tasks.md` (и `plan.md`, только если изменение действительно нетривиально).

Это **быстрая полоса** — заменяет последовательность `spec → plan → tasks`, когда изменение маленькое/низкорисковое. Если intent неоднозначен, идея разбивается на несколько фич или нужен серьёзный дизайн — ОСТАНОВИСЬ и откатись на `/spk.spec` (и `/spk.plan`).

Следуй базовым правилам в `AGENTS.md`.

## Контракт фазы

Вход: идея пользователя, `.speckeep/constitution.summary.md` (предпочтительно) или `project.constitution_file`, минимальный контекст репозитория.
Выход: `<specs_dir>/<slug>/spec.md` + `<specs_dir>/<slug>/tasks.md` (plan.md опционален).
Стоп, если: идея неоднозначна, охватывает больше одной фичи или потребует выдумывать требования/AC.

## Workflow

1. Прогони readiness проверку фазы: `./.speckeep/scripts/check-ready.sh propose [slug]`.
2. Выведи slug (или прими `--slug`). Работай с ветки `feature/<slug>` — создай/переключись, если её нет (то же правило, что у `/spk.spec`).
3. Напиши `spec.md` по шаблону spec: `## Goal`, `## Requirements` (`RQ-*`), `## Acceptance Criteria` (`AC-*` с **Given / When / Then**), `## Assumptions`.
4. Выбери полосу:
   - **express** (по умолчанию): без `plan.md` — сразу к задачам.
   - **planned**: только при нескольких реалистичных вариантах, кросс-граничном влиянии или рисках миграции/раскатки — тогда дополнительно напиши `plan.md` (с `DEC-*`) и `data-model.md` только если модель данных реально меняется.
5. Напиши `tasks.md` по шаблону задач, самодостаточный для implement: `## Surface Map`, `Touches:` у каждой задачи, `## Implementation Context`, `## Acceptance Coverage` (`AC-* -> T*`).
6. Код в этой фазе НЕ реализуй.

## Правила

- Минимальный контекст: только текущий slug и нужные repo surfaces; никаких full-repo сканов.
- Используй `.speckeep/templates/spec.md` и `.speckeep/templates/tasks.md` как каркас; никогда не ищи форму в чужих slug.
- Дисциплина размера: `spec.md` ≤ ~80 строк, `tasks.md` ≤ ~150 строк; переполнение обычно значит, что идея слишком большая для propose — остановись и возвращайся на `/spk.spec`.
- Каждый `AC-*` маппится на ≥ 1 задачу; у каждой задачи есть `Touches:` и измеримый outcome.
- Если модель данных меняется — создай `data-model.md`; иначе строка `Data model: no change` живёт в `tasks.md` → `Implementation Context` или в `plan.md`.
- Конституция: AGENTS.md (`.speckeep/constitution.summary.md` предпочтительнее).
- Ветка: `feature/<slug>`, если пользователь не передал `--branch`.

## Self-Check (обязателен перед завершением)

- [ ] Идея одно-фичевая и однозначная; нет выдуманных требований/AC
- [ ] В `spec.md` есть `RQ-*` + `AC-*` (Given/When/Then) + `## Assumptions`
- [ ] В `tasks.md` есть `## Surface Map`, `Touches:` везде, `## Implementation Context`, `## Acceptance Coverage`
- [ ] Каждый `AC-*` покрыт ≥ 1 задачей
- [ ] `plan.md` существует только для нетривиальных изменений (express — полоса по умолчанию)

Если хоть один пункт не прошёл: исправь и прогони заново. После **2 раундов** правок, которые всё ещё не проходят, остановись и вернись на `/spk.spec <slug>` (или задай один уточняющий вопрос) — не продавливай propose.

## Ожидания по выводу

- Запиши `spec.md` + `tasks.md` (patch-in-place при повторном открытии через `--amend`).
- Коротко суммируй: slug, полоса (express/planned), surfaces, покрытие AC.
- Заверши стандартным end block (см. AGENTS.md):
  ```
  Slug: <slug>
  Status: propose
  Artifacts: <пути>
  Blockers: <none | причина>
  Готово к: /spk.implement <slug>
  ```
- Финальная строка: `Готово к: /spk.implement <slug>`