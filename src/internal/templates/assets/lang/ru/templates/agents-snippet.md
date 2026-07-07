## SpecKeep

Основной контекст: `.speckeep/`. Языки: docs=[DOCS_LANGUAGE], agent=[AGENT_LANGUAGE], comments=[COMMENTS_LANGUAGE]

Цепочка workflow: `constitution → spec → [inspect, опционально] → plan → tasks → implement → verify → archive (CLI-only после verify)`

Базовые правила:

- Пути/конфиг: читайте `.speckeep/speckeep.yaml` ≤ 1 раза за сессию; если конфига нет, defaults: `<specs_dir>=specs/active`, `<archive_dir>=specs/archived`, constitution=`CONSTITUTION.md`.
- Конституция: загружайте `.speckeep/constitution.summary.md` сначала, если файл существует; только при его отсутствии переходите к `project.constitution_file` (по умолчанию `CONSTITUTION.md`).
- Ветки: только `/spk.spec` может переключать/создавать `feature/<slug>` (или `--branch`). Остальные фазы должны уже быть на нужной ветке.
- Скрипты: перед каждой фазой запускайте `check-ready.* <phase> <slug>` (и любые extras из секции Команды); доверяйте stdout/exit code; исходники `.speckeep/scripts/*` не читать.
- Scope/load: по умолчанию только текущий slug; без широких репо-сканов; предпочитайте surfaces из `Touches:`.
- ⚠️ **CRITICAL — Repository map first**: **НЕ** используйте `ls`, `find`, glob для первичной навигации. Прочитайте `REPOSITORY_MAP.md` в первую очередь — он содержит полную карту репозитория. Это экономит токены и соблюдает workflow discipline. Читайте карту один раз за сессию, переиспользуйте заметки; перечитывайте только если сами обновили карту в этой же сессии.
- Git safety: не делать `git commit/push/tag` и PR без явной просьбы.
- Done: никогда не отмечать задачу выполненной без observable proof (путь файла, вывод теста или команды). Каждый артефакт должен быть понятен коллеге без дополнительных объяснений.
- Traceability: для каждой нетривиально завершённой задачи trace-маркеры обязательны. Нет `@sk-task` в изменённом коде или нет `@sk-test` в изменённых тестах для этой задачи — задача ещё не завершена.
- Placement: trace-маркеры запрещено ставить на уровень `package`, `import` или file-header comment; ставьте их над owning function/method/test/type declaration.
- End block: каждая фаза завершается компактным summary: `Slug`, `Status`, `Artifacts`, `Blockers`, `Готово к` (или `Вернуться к` при blocked / `speckeep archive` только после `verify: pass`).
- Discovery: не запускать `speckeep ... --help` для разведки; используйте prompt-файлы и readiness scripts.
- CLI: используйте `./.speckeep/scripts/run-speckeep.sh` (PowerShell: `./.speckeep/scripts/run-speckeep.ps1`) только для настоящих CLI-команд (напр. `doctor`, `check`, `trace`, `export`, `refresh`). Не запускайте `run-speckeep.* <phase>` вроде `spec`/`plan`/`tasks` — фазы выполняются как slash-команды, а артефакты пишутся напрямую.
- Вывод в чат: не вставляйте большие `git diff`/полные файлы/простыни логов. Давайте краткое резюме изменений + список затронутых файлов; если нужны детали — покажите только небольшой фрагмент вокруг места правки.
- Scope: не читайте и не меняйте артефакты других slug/спек, если текущая задача явно не требует (иначе это scope violation).
- Don't invent: не вводите требований, зависимостей, scope или критериев приёмки, отсутствующих во входных артефактах текущей фазы.

Команды (префикс: `/spk.`):

- `/spk.constitution` → конституция
- `/spk.spec` → spec (branch-first)
- `/spk.inspect` → опциональная глубокая проверка качества
- `/spk.plan` → plan artifacts
- `/spk.tasks` → tasks
- `/spk.implement` → implement
- `/spk.verify` → verify
- `/spk.challenge` → адверсариальная проверка spec/plan (слепые зоны, непроверяемые AC)
- `/spk.rollback` → откат выполненных задач фичи, возврат в незавершённое состояние
- `/spk.recap` → обзор проекта: активные фичи, фаза, следующий шаг
- `speckeep archive <slug> .` → CLI-only архив после `verify: pass`
- `/spk.repo-map` → обновить `REPOSITORY_MAP.md` (см. выделенный prompt для политики + шаблона)

Чеклист триггеров обновления (запускайте `/spk.repo-map`, если истинно хотя бы одно):

- Добавлена или удалена верхнеуровневая кодовая директория/модуль.
- Перемещены/переименованы ключевые исходники, меняющие навигацию.
- Добавлены/удалены runtime/service/CLI entrypoints.
- Существенно изменены границы подсистем (заметно поменялись where-to-edit пути).
- Пользователь явно попросил обновить repo map.
