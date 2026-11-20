# Справочник команд SpecKeep

Загружайте этот файл только когда текущая фаза не назвала нужную команду. Промпты фаз лежат в `.speckeep/templates/prompts/`; каждый skill `/spk-<phase>` самодостаточен.

Дефолт для новой фичи: `/spk-propose` → `/spk-implement`. Запрос вида «сделай X по SpecKeep» ведёт в `/spk-propose`; для неоднозначной/мульти-фичевой идеи назови `/spk-spec`, для дизайна — `/spk-plan`. План — opt-in.

`Ready for` может содержать `(alt: ...)`; первая команда — основной/безопасный вариант.

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
- `/spk-grill` — опциональное pre-spec интервью: раскрыть design tree до spec
- `/spk-research` — опциональное техническое research: cited `research.md`
- `/spk-diagnose` — опциональный bug-цикл: reproduce → minimize → hypothesize → fix → regression
- `/spk-challenge` — адверсариальная проверка spec/plan
- `/spk-scope` — быстрая проверка границ фичи
- `/spk-handoff` — handoff-документ сессии по одной фиче
- `/spk-resume` — минимальное возобновление в сессии: пакет состояния + следующая задача (без артефакта)
- `/spk-recap` — обзор проекта: активные фичи, фаза, следующий шаг
- `/spk-rollback` — откат выполненных задач
- `/spk-hotfix` — экстренное исправление вне цепочки фаз (≤ 3 файлов, без перепланирования)

## Закрытие (только CLI)

- `speckeep archive <slug> .` — по умолчанию: только архив артефактов (PR/ручной merge, deferred/abandoned). Для `completed` нужны закрытые задачи, `Proof:` у каждой `[x]` и `verify: pass` при `workflow.verify: required`; non-completed статусы могут архивировать незавершённую работу. Старые `completed`-фичи без `Proof:` должны добавить evidence или архивироваться с non-completed статусом и причиной.
- `speckeep finish <slug> .` — локальное закрытие: archive + commit снимка + merge `--no-ff` в базовую ветку + удаление локальной feature-ветки. Только когда пользователь явно просит merge/close. Если пользователь называет target-ветку, передай `--into <branch>` (например, `--into dev`); она должна существовать локально. Никогда не push и не удаляет remote-ветки.
