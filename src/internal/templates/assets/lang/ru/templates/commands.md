# Справочник команд SpecKeep

Загружайте этот файл только когда текущая фаза не назвала нужную команду. Промпты фаз лежат в `.speckeep/templates/prompts/`; каждый skill `/spk-<phase>` самодостаточен.

Дефолт для новой фичи: `/spk-propose` → `/spk-implement`. Запрос вида «сделай X по SpecKeep» ведёт в `/spk-propose`; для неоднозначной/мульти-фичевой идеи назови `/spk-spec`, для дизайна — `/spk-plan`. План — opt-in.

## Фазы (с readiness-гейтом)

- `/spk-constitution` — создать/обновить конституцию проекта
- `/spk-propose` — рекомендуемый дефолт: идея → spec + tasks (plan опционально), сразу к implement
- `/spk-spec` — написать spec (branch-first)
- `/spk-inspect` — опциональная глубокая проверка качества
- `/spk-plan` — opt-in plan artifacts (только когда фиче нужен дизайн)
- `/spk-tasks` — написать tasks
- `/spk-implement` — реализовать задачи
- `/spk-verify` — проверить tasks/AC (по умолчанию опциональный аудит)
- `/spk-converge` — быстрый цикл закрытия: повторная проверка задач/proof, добавление follow-up задач, повтор до сходимости

Перед гейтируемой фазой запустите `./.speckeep/scripts/check-ready[SCRIPT_EXT] <phase> <slug>`. У вспомогательных команд readiness-гейта нет.

## Вспомогательные команды (без readiness-гейта)

- `/spk-repo-map` — обновить `REPOSITORY_MAP.md` (триггеры: см. `.speckeep/templates/prompts/repo-map.md`)
- `/spk-module-context <path>` — записать/обновить `<path>/CONTEXT.md`
- `/spk-glossary` — создать/обновить `.speckeep/glossary.md`
- `/spk-challenge` — адверсариальная проверка spec/plan
- `/spk-scope` — быстрая проверка границ фичи
- `/spk-handoff` — handoff-документ сессии по одной фиче
- `/spk-recap` — обзор проекта: активные фичи, фаза, следующий шаг
- `/spk-rollback` — откат выполненных задач
- `/spk-hotfix` — экстренное исправление вне цепочки фаз (≤ 3 файлов, без перепланирования)

## Архив (только CLI)

- `speckeep archive <slug> .` — только когда фича детерминированно доказана (все задачи `[x]` имеют `Proof:`) или после `verify: pass`; при `workflow.verify: required` архив требует `verify: pass`.
