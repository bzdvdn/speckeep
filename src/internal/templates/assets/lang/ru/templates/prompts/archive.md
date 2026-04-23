# Prompt архивации SpecKeep (compact)

Вы архивируете один feature package в директорию архива (по умолчанию `archive/`).

## Разрешение путей

- Определите `<specs_dir>` из `.speckeep/speckeep.yaml` (читать ≤ 1 раза за сессию). Если конфиг отсутствует — используйте `specs`.

## Phase Contract

Inputs: `<specs_dir>/<slug>/` (spec/inspect/summary/spec.digest) + `<specs_dir>/<slug>/plan/` (plan/plan.digest/tasks/verify и прочие артефакты по наличию). Отчёт verify — это `<specs_dir>/<slug>/plan/verify.md`.
Outputs: snapshot в `<archive_dir>/<slug>/...` (по умолчанию `archive/<slug>/...`, move-based).
Stop if: не завершён verify или статус архивации не может быть обоснован.

## Правила

- Перед архивом предпочтительно запустить `./.speckeep/scripts/check-archive-ready.*` (slug первым аргументом).
- Default status: `completed`. Нестандартные статусы требуют явного `--reason`.
- Архив — это фиксация состояния, не место для новых правок реализации.
- Не ищите «примеры» архивов/снапшотов в других slug ради формата. Достаточно следовать этому prompt и проверкам readiness; summary форматируется/генерируется автоматически командой архивации.
- Скрипты для запуска (ориентир, можно копировать как есть):
  - `./.speckeep/scripts/check-archive-ready.sh <slug> completed`
  - `./.speckeep/scripts/archive-feature.sh <slug> . --status completed`
  - нестандартный статус: `./.speckeep/scripts/archive-feature.sh <slug> . --status deferred --reason "..."` (и соответствующий `check-archive-ready` с тем же статусом/причиной)

## Output expectations

- Создайте snapshot; кратко перечислите перемещённые артефакты и итоговый статус.
- Это терминальный шаг workflow для этой фичи (после verify).
- Финальная строка: `Готово к: /speckeep.recap`
