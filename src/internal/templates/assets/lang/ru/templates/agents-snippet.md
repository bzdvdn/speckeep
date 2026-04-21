## SpecKeep

Основной контекст: `.speckeep/`. Языки: docs=[DOCS_LANGUAGE], agent=[AGENT_LANGUAGE], comments=[COMMENTS_LANGUAGE]

Цепочка workflow: `constitution → spec → inspect → plan → tasks → implement → verify → archive`

Базовые правила:
- Пути/конфиг: используйте дефолты `.speckeep/`; читайте `.speckeep/speckeep.yaml` ≤ 1 раза за сессию.
- Ветки: только `/speckeep.spec` может переключать/создавать `feature/<slug>` (или `--branch`). Остальные фазы должны уже быть на нужной ветке.
- Скрипты: запускайте readiness scripts; доверяйте stdout/exit code; исходники `/.speckeep/scripts/*` не читать.
- Scope/load: по умолчанию только текущий slug; без широких репо-сканов; предпочитайте surfaces из `Touches:`.
- Git safety: не делать `git commit/push/tag` и PR без явной просьбы.
- CLI: используйте `./.speckeep/scripts/run-speckeep.sh` (PowerShell: `./.speckeep/scripts/run-speckeep.ps1`).
- Вывод в чат: не вставляйте большие `git diff`/полные файлы/простыни логов. Давайте краткое резюме изменений + список затронутых файлов; если нужны детали — покажите только небольшой фрагмент вокруг места правки.
- Scope: не читайте и не меняйте артефакты других slug/спек, если текущая задача явно не требует (иначе это scope violation).

Команды:
- `/speckeep.constitution` → конституция
- `/speckeep.spec` → spec (branch-first)
- `/speckeep.inspect` → inspect
- `/speckeep.plan` → plan package
- `/speckeep.tasks` → tasks
- `/speckeep.implement` → implement
- `/speckeep.verify` → verify
- `/speckeep.archive` → archive
