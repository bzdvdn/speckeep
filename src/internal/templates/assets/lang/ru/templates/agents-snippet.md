## SpecKeep

Основной контекст: `.speckeep/`. Языки: docs=[DOCS_LANGUAGE], agent=[AGENT_LANGUAGE], comments=[COMMENTS_LANGUAGE]

Цепочка workflow: `constitution → spec → [inspect, опционально] → plan → tasks → implement → archive`; `verify` — опциональный on-demand аудит (всегда доступен, по умолчанию пропускается). Уважайте `workflow.verify` в `.speckeep/speckeep.yaml`: `required` возвращает verify как обязательный гейт перед archive.

Базовые правила:

- ⚠️ **CRITICAL — Repository map first**: **НЕ** используйте `ls`, `find`, glob для первичной навигации. Прочитайте `REPOSITORY_MAP.md` в первую очередь — он содержит полную карту репозитория. Это экономит токены и соблюдает workflow discipline. Читайте карту один раз за сессию, переиспользуйте заметки; перечитывайте только если сами обновили карту в этой же сессии.
- Пути/конфиг: читайте `.speckeep/speckeep.yaml` ≤ 1 раза за сессию; если конфига нет, defaults: `<specs_dir>=specs/active`, `<archive_dir>=specs/archived`, constitution=`CONSTITUTION.md`.
- Конституция: загружайте `.speckeep/constitution.summary.md` сначала, если файл существует; только при его отсутствии переходите к `project.constitution_file` (по умолчанию `CONSTITUTION.md`).
- Ветки: только `/spk.spec` может переключать/создавать `feature/<slug>` (или `--branch`). Остальные фазы должны уже быть на нужной ветке.
- Скрипты: перед каждой фазой запускайте `check-ready.* <phase> <slug>` (и любые extras из секции Команды); доверяйте stdout/exit code; исходники `.speckeep/scripts/*` не читать.
- Scope/load: по умолчанию только текущий slug; без широких репо-сканов; предпочитайте surfaces из `Touches:`.
- Git safety: не делать `git commit/push/tag` и PR без явной просьбы.
- Done: никогда не отмечать задачу выполненной без observable proof (путь файла, вывод теста или команды). Каждый артефакт должен быть понятен коллеге без дополнительных объяснений.
- Proof: доказательство каждой завершённой задачи — строка `Proof:` в `tasks.md` непосредственно под отмеченной задачей: `Proof: <kind> <path> [<anchor>]` (`kind` = `code|test|docs|chore`). Задача `[x]` без `Proof:` — ещё не завершена. `speckeep trace`, `speckeep doctor` и архивные проверки читают evidence только из строк `Proof:` в `tasks.md`.
- End block: каждая фаза завершается компактным summary: `Slug`, `Status`, `Artifacts`, `Blockers`, `Готово к`. `Готово к` — это `speckeep archive`, когда все задачи `[x]` имеют `Proof:` (или после `verify: pass`), КРОМЕ случая `workflow.verify: required`, когда сначала `/spk.verify`. При blocked используйте `Вернуться к`.
  - Канонический end block (точная форма, единственный источник — фазовые prompt'ы ссылаются на него, не выводите локальный вариант):
    ```
    Slug: <slug>
    Status: <фаза>
    Artifacts: <пути>
    Blockers: <none | причина>
    Готово к: <следующая команда>   (или "Вернуться к: /spk.<phase> <slug>" при blocked)
    ```
- Политика verify-гейта (единственный источник — prompt'ы ссылаются на неё по имени; не ведите приватные копии):
  - `workflow.verify: required` → verify — **обязательный гейт перед archive**: после implement/hotfix/handoff строка `Готово к` должна быть `/spk.verify <slug>`, а `speckeep archive` требует `verify.md` со `status: pass`.
  - `workflow.verify: optional` (или отсутствует, по умолчанию) → verify — on-demand аудит; archive разрешён, когда все задачи `[x]` имеют `Proof:`. Присутствующий `verify.md` со status ≠ `pass` по-прежнему вето на archive.
- Discovery: не запускать `speckeep ... --help` для разведки; используйте prompt-файлы и readiness scripts.
- CLI: используйте `./.speckeep/scripts/run-speckeep.sh` (PowerShell: `./.speckeep/scripts/run-speckeep.ps1`) только для настоящих CLI-команд (напр. `doctor`, `check`, `trace`, `export`, `refresh`). Не запускайте `run-speckeep.* <phase>` вроде `spec`/`plan`/`tasks` — фазы выполняются как slash-команды, а артефакты пишутся напрямую.
- Вывод в чат: не вставляйте большие `git diff`/полные файлы/простыни логов. Давайте краткое резюме изменений + список затронутых файлов; если нужны детали — покажите только небольшой фрагмент вокруг места правки.
- Scope: не читайте и не меняйте артефакты других slug/спек, если текущая задача явно не требует (иначе это scope violation).
- Don't invent: не вводите требований, зависимостей, scope или критериев приёмки, отсутствующих во входных артефактах текущей фазы.
- Эскалация: если вы не можете честно завершить текущую фазу (нет артефакта, неоднозначный intent, небезопасное изменение, заблокированный гейт) — ОСТАНОВИТЕСЬ и сообщите `Вернуться к: /spk.<phase> <slug>` или задайте один точечный вопрос. Никогда не выдумывайте criterion приёмки, не прячьте пробелы и не угадывайте следующий шаг, чтобы избежать остановки — остановка с точной причиной это корректный результат.

Команды (префикс: `/spk.`):

- `/spk.constitution` → конституция
- `/spk.spec` → spec (branch-first)
- `/spk.inspect` → опциональная глубокая проверка качества
- `/spk.plan` → plan artifacts
- `/spk.tasks` → tasks
- `/spk.implement` → implement
- `/spk.verify` → verify
- `/spk.challenge` → адверсариальная проверка spec/plan (слепые зоны, непроверяемые AC)
- `/spk.scope` → быстрая проверка границ фичи (in/out, риски)
- `/spk.rollback` → откат выполненных задач фичи, возврат в незавершённое состояние
- `/spk.recap` → обзор проекта: активные фичи, фаза, следующий шаг
- `/spk.handoff` → handoff-документ сессии по одной фиче (продолжение без догадок)
- `/spk.hotfix` → экстренное исправление вне цепочки фаз (≤ 3 файлов, без перепланирования)
- `speckeep archive <slug> .` → CLI-only архив когда фича детерминированно доказана (все задачи `[x]` имеют `Proof:`) или после `verify: pass`; при `workflow.verify: required` архив требует `verify: pass`
- `/spk.repo-map` → обновить `REPOSITORY_MAP.md` (см. выделенный prompt для политики + шаблона)

Чеклист триггеров обновления (запускайте `/spk.repo-map`, если истинно хотя бы одно):

- Добавлена или удалена верхнеуровневая кодовая директория/модуль.
- Перемещены/переименованы ключевые исходники, меняющие навигацию.
- Добавлены/удалены runtime/service/CLI entrypoints.
- Существенно изменены границы подсистем (заметно поменялись where-to-edit пути).
- Пользователь явно попросил обновить repo map.
