## SpecKeep

Основной контекст: `.speckeep/`. Языки: docs=[DOCS_LANGUAGE], agent=[AGENT_LANGUAGE], comments=[COMMENTS_LANGUAGE]

Цепочка workflow: `constitution → spec → [inspect, опционально] → plan → tasks → implement → verify → archive`

Базовые правила:
- Пути/конфиг: читайте `.speckeep/speckeep.yaml` ≤ 1 раза за сессию; если конфига нет, defaults: `<specs_dir>=specs`, `<archive_dir>=archive`, constitution=`CONSTITUTION.md`.
- Конституция: в любой фазе предпочитайте `.speckeep/constitution.summary.md` вместо `CONSTITUTION.md` если файл есть.
- Ветки: только `/speckeep.spec` может переключать/создавать `feature/<slug>` (или `--branch`). Остальные фазы должны уже быть на нужной ветке.
- Скрипты: запускайте readiness scripts; доверяйте stdout/exit code; исходники `.speckeep/scripts/*` не читать.
- Scope/load: по умолчанию только текущий slug; без широких репо-сканов; предпочитайте surfaces из `Touches:`.
- Repository map first: если есть `REPOSITORY_MAP.md`, прочитайте его до широкого поиска по файлам. Читайте карту один раз за сессию и переиспользуйте заметки; перечитывайте только если сами обновили карту в этой же сессии.
- Git safety: не делать `git commit/push/tag` и PR без явной просьбы.
- Done: никогда не отмечать задачу выполненной без observable proof (путь файла, вывод теста или команды).
- Discovery: не запускать `speckeep ... --help` для разведки; используйте prompt-файлы и readiness scripts.
- CLI: используйте `./.speckeep/scripts/run-speckeep.sh` (PowerShell: `./.speckeep/scripts/run-speckeep.ps1`) только для настоящих CLI-команд (напр. `doctor`, `check`, `trace`, `export`, `refresh`). Не запускайте `run-speckeep.* <phase>` вроде `spec`/`plan`/`tasks` — фазы выполняются как slash-команды, а артефакты пишутся напрямую.
- Вывод в чат: не вставляйте большие `git diff`/полные файлы/простыни логов. Давайте краткое резюме изменений + список затронутых файлов; если нужны детали — покажите только небольшой фрагмент вокруг места правки.
- Scope: не читайте и не меняйте артефакты других slug/спек, если текущая задача явно не требует (иначе это scope violation).

Команды:
- `/speckeep.constitution` → конституция
- `/speckeep.spec` → spec (branch-first)
- `/speckeep.inspect` → опциональная глубокая проверка качества
- `/speckeep.plan` → plan package
- `/speckeep.tasks` → tasks
- `/speckeep.implement` → implement
- `/speckeep.verify` → verify
- `/speckeep.repo-map` → обновить `REPOSITORY_MAP.md` по компактному шаблону ниже

Политика repository map (`/speckeep.repo-map`):
- Держите `REPOSITORY_MAP.md` компактным и code-only (пути + короткие роли).
- Language-agnostic: определяйте стек по маркерам репозитория (напр. `go.mod`, `package.json`, `pyproject.toml`, `Cargo.toml`, `pom.xml`, `*.csproj`) и адаптируйте секции под найденный стек.
- Не предполагайте Go-структуру для не-Go проектов.
- Жесткий лимит размера: целевой объем до 180 строк; если карта растет — сжимайте, а не расширяйте prose.
- Обновляйте in-place (минимальный diff): сохраняйте неизменные строки/порядок и правьте только затронутые записи/секции.
- Не переписывайте файл целиком, если изменилась только часть карты.
- Если `REPOSITORY_MAP.md` отсутствует — создайте по шаблону; если существует — патчите существующее содержимое.
- Исключайте из индексации: `src/internal/agents/**`, `.speckeep/**`, `specs/**`, `archive/**`, `.git/**`, `bin/**`, `demo/**`, `docs/**`, `TESTS/**`, `node_modules/**`, `vendor/**`, `dist/**`, `build/**`, `coverage/**`.
- Важно: проектные настройки уже читаются из `.speckeep/speckeep.yaml`; не дублируйте этот конфиг в карте.

Чеклист триггеров обновления (запускайте `/speckeep.repo-map`, если истинно хотя бы одно):
- Добавлена или удалена верхнеуровневая кодовая директория/модуль.
- Перемещены/переименованы ключевые исходники, меняющие навигацию.
- Добавлены/удалены runtime/service/CLI entrypoints.
- Существенно изменены границы подсистем (заметно поменялись where-to-edit пути).
- Пользователь явно попросил обновить repo map.

Шаблон repository map:
```md
# Repository Map

## Entry Points
- `<path>` — `<runtime/service/cli entrypoint>`

## Top-Level Code
- `<path>` — `<module role>`

## Key Paths
- `<path>` — `<what is implemented here>`

## Where To Edit
- `<change type>` — `<likely paths>`

## Excluded
- `<glob>` — `excluded from indexing`
```
