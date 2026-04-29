# Prompt архивации SpecKeep (compact)

Вы архивируете фичу запуском двух скриптов. Не читайте файлы фичи, не смотрите diff, не валидируйте артефакты вручную — скрипты делают всё это сами.

## Phase Contract

Inputs: аргументы пользователя (`<slug>`, опционально `--status`, опционально `--reason`).
Outputs: archive snapshot, созданный скриптом `archive-feature`.
Stop if: `check-archive-ready` завершился с ненулевым кодом — сообщите stdout и остановитесь.

## Правила

- Не читайте файлы фичи (`spec.md`, `plan.md`, `tasks.md`, `verify.md` и др.). Скрипты сами обрабатывают валидацию.
- Не запускайте `--help` ни у каких команд для обнаружения синтаксиса. Используйте пути скриптов ниже точно как указано.
- Default status — `completed`. `--status deferred --reason "..."` — только если пользователь явно просит.
- Доверяйте выводу скриптов полностью: если `check-archive-ready` прошёл — продолжайте; если упал — сообщите и остановитесь.

## Шаги (всегда в этом порядке)

1. Запустите readiness check:
   - `completed`: `./.speckeep/scripts/check-archive-ready.sh <slug> completed`
   - другой статус: `./.speckeep/scripts/check-archive-ready.sh <slug> <status> --reason "<reason>"`
2. Если exit code 0 — запустите архивацию:
   - `completed`: `./.speckeep/scripts/archive-feature.sh <slug> . --status completed`
   - другой статус: `./.speckeep/scripts/archive-feature.sh <slug> . --status <status> --reason "<reason>"`

## Output expectations

- Сообщите вывод скрипта (stdout) и итоговый статус.
- По умолчанию не обновляйте `REPOSITORY_MAP.md` на archive. Обновляйте только если скрипты архивации или явная просьба пользователя реально изменили структуру/навигацию репозитория.
- Финальная строка: `Готово к: /speckeep.recap`
