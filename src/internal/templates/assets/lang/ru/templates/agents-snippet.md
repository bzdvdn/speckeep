## SpecKeep

Основной контекст проекта — `.speckeep/`. Языки: docs=[DOCS_LANGUAGE], agent=[AGENT_LANGUAGE], comments=[COMMENTS_LANGUAGE]

Цепочка workflow: `constitution → spec → inspect → plan → tasks → implement → verify → archive`
- `/speckeep.constitution`: создать или обновить `.speckeep/constitution.md`
- `/speckeep.spec`: создать или уточнить `.speckeep/specs/<slug>/spec.md`; `--amend` для точечных правок (**обязательно** branch-first: до записи любого файла переключиться/создать `feature/<slug>` или ветку из `--branch`)
- `/speckeep.inspect`: проверить одну фичу на согласованность и качество
- `/speckeep.plan`: создать или обновить `.speckeep/specs/<slug>/plan/`; `--update` для точечных правок, `--research` для research-first
- `/speckeep.tasks`: создать или обновить `.speckeep/specs/<slug>/plan/tasks.md`
- `/speckeep.implement`: выполнить незавершённые задачи из `tasks.md`
- `/speckeep.verify`: проверить один feature package; `--deep` для полной per-AC валидации по коду
- `/speckeep.archive`: архивировать в `.speckeep/archive/` (move-based); `--copy` оставляет оригиналы, `--restore` восстанавливает

Опциональные (в любой момент): `/speckeep.challenge` (адверсариальная проверка; `--spec`/`--plan`), `/speckeep.handoff` (передача сессии), `/speckeep.hotfix` (экстренное исправление ≤ 3 файлов), `/speckeep.scope` (проверка границ; `--plan`/`--tasks`), `/speckeep.recap` (обзор проекта)

Дисциплина чтения:
- Не пропускайте фазы; по умолчанию загружайте только текущий feature slug
- Предпочитайте readiness scripts перед чтением более глубоких артефактов; для CLI используйте `./.speckeep/scripts/run-speckeep.sh`
- Не загружайте: нерелевантные specs/plans, широкие сканы репозитория, исходники scripts, файлы уже прочитанные в сессии (если сами не редактировали)
- Используйте настроенный язык комментариев для нового/изменяемого кода; сохраняйте существующие соглашения файла

Перед значимыми изменениями: просмотрите `constitution.md`, релевантный `specs/<slug>/spec.md` и `specs/<slug>/plan/` если есть. После изменений: поддерживайте согласованность specs, plans, tasks и реализации.
