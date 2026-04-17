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

## Базовые правила (применяются к каждой slash-команде, если prompt не переопределяет)

- **Пути**: используйте дефолты `.speckeep/`, если `.speckeep/speckeep.yaml` не переопределяет `paths.specs_dir`, `paths.archive_dir` или `project.constitution_file`. Читайте конфиг ≤ 1 раза за сессию.
- **Git**: ветки создаёт/переключает только `/speckeep.spec` (в `feature/<slug>` или явный `--branch`). Остальные фазы должны уже быть на нужной ветке — иначе остановитесь и сообщите; не создавайте ветки. Не запускайте `git commit`/`push`/`tag`, не открывайте PR без явного запроса пользователя. Для CLI используйте `./.speckeep/scripts/run-speckeep.sh` (PowerShell: `./.speckeep/scripts/run-speckeep.ps1`).
- **Load discipline**: по умолчанию грузите только текущий feature slug. Не читайте нерелевантные specs/plans, широкие репо-сканы, исходники `/.speckeep/scripts/*`, файлы, уже прочитанные в этой сессии (если сами не редактировали). Предпочитайте вывод readiness-скрипта чтению исходников или более глубоких артефактов.
- **Readiness scripts**: если `/.speckeep/scripts/check-<phase>-ready.*` существует, запускайте со slug как первый аргумент (например `bash ./.speckeep/scripts/check-plan-ready.sh <slug>` или `.\.speckeep\scripts\check-plan-ready.ps1 <slug>`). Используйте его вывод как основной слой структурных свидетельств; не перевыводите их самостоятельно.
- **Язык**: используйте настроенный язык документации для новых/редактируемых артефактов и настроенный язык комментариев для нового/редактируемого кода. Сохраняйте существующие соглашения файла; не смешивайте языки внутри одного артефакта без веской причины.
- **Phase discipline**: не дрейфуйте в работу других фаз — каждая команда пишет только свои артефакты.

Перед значимыми изменениями: просмотрите `constitution.md`, релевантный `specs/<slug>/spec.md` и `specs/<slug>/plan/` если есть. После изменений: поддерживайте согласованность specs, plans, tasks и реализации.
