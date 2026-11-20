## SpecKeep

Основной контекст: `.speckeep/`. Языки: docs=[DOCS_LANGUAGE], agent=[AGENT_LANGUAGE], comments=[COMMENTS_LANGUAGE]

Цепочка: `constitution → spec → [inspect, опц.] → [plan, опц.] → tasks → implement → archive`; `verify` — опциональный аудит (по умолчанию пропускается). `/spk-propose` — дефолт для новой фичи (spec+tasks за один проход): используй напрямую и не грузи `sdd`-роутер; `/spk-spec` — только для неоднозначных/мульти-фичевых идей; plan — opt-in для фич с дизайном. `workflow.verify` в `.speckeep/speckeep.yaml`: `required` делает verify обязательным гейтом перед archive.

Базовые правила:

- ⚠️ **Repository map first**: для первичной навигации не используй `ls`/`find`/glob — сначала прочитай `REPOSITORY_MAP.md` (полная карта, экономит токены). Читай раз за сессию; перечитывай только после своего обновления карты.
- Пути/конфиг: `.speckeep/speckeep.yaml` читай ≤ 1 раза за сессию — он задаёт `<specs_dir>`, `<archive_dir>` и `project.constitution_file`. Если конфига нет: `specs/active`, `specs/archived`, `CONSTITUTION.md`.
- Конституция: сначала `.speckeep/constitution.summary.md`, если есть; иначе `project.constitution_file` из конфига.
- Доменный язык: `.speckeep/glossary.md` читай раз за сессию и переиспользуй термины в spec/plan/tasks; не вводи синоним уже определённому термину. Файл опционален, создаётся только через `/spk-glossary`.
- Ветки: только `/spk-spec`/`/spk-propose` переключают/создают `feature/<slug>` (или `--branch`); остальные фазы уже должны быть на нужной ветке.
- Скрипты: `./.speckeep/scripts/check-ready[SCRIPT_EXT] <phase> <slug>` — только для гейтируемых фаз (`constitution`, `spec`, `propose`, `inspect`, `plan`, `tasks`, `implement`, `verify`, `converge`; `archive` — CLI). Aux-командам (`repo-map`, `module-context`, `glossary`, `challenge`, `handoff`, `recap`, `scope`, `rollback`, `hotfix`) гейт не нужен. Доверяй exit code; исходники `.speckeep/scripts/*` не читай.
- Скиллы: workflow — skill-pack (`sdd`-роутер + `spk-<phase>/SKILL.md` на фазу). Грузи только нужный фазовый скилл; `sdd` — когда фаза неочевидна.
- Scope/load: только текущий slug; без широких сканов; surfaces из `Touches:`.
- Git: без `commit/push/tag` и PR без явной просьбы.
- Done: задача закрыта только с observable proof (файл, тест, вывод команды); артефакт должен быть понятен коллеге без пояснений.
- Proof: под каждой `[x]` задачей в `tasks.md` — строка `Proof: <kind> <path> [<anchor>]` (`kind` = `code|test|docs|chore`). Без `Proof:` задача не закрыта; `speckeep trace`, `doctor` и archive читают только эти строки.
- Trace: сразу после `feature/<slug>` выполни `./.speckeep/scripts/run-speckeep[SCRIPT_EXT] trace init <slug>`. `Touches:` — план; `trace <slug> --files` — факт (`Unmapped` = drift); `trace --file <path>` — кто менял файл.
- Токен-дисциплина: в той же фазе продолжай в сессии (`continue`); не перевызывай тот же slash/skill — это перезагрузка промпта. Переход фазы — ровно одна загрузка; новую команду — только при смене фазы или после очистки контекста. `Готово к` называет следующую фазу + `continue?`, а не повтор текущей команды.
- End block: каждая фаза завершается каноническим блоком ниже (при blocked — `Вернуться к`).
  - Канонический end block (точная форма, единственный источник — prompt'ы ссылаются на него, не выводи локальный вариант):
    ```
    Slug: <slug>
    Status: <фаза>
    Artifacts: <пути>
    Blockers: <none | причина>
    Готово к: <следующая команда>   (или "Вернуться к: /spk-<phase> <slug>" при blocked)
    ```
- Политика verify-гейта (единственный источник; prompt'ы ссылаются по имени):
  - `workflow.verify: required` → verify обязателен перед archive: после implement/hotfix/handoff `Готово к` = `/spk-verify <slug>`; archive требует `verify.md` со `status: pass`.
  - `workflow.verify: optional` (или нет ключа) → verify on-demand; archive разрешён, когда все `[x]` имеют `Proof:`. `verify.md` со status ≠ `pass` — вето на archive.
- Discovery: не запускай `speckeep --help` для разведки; используй prompt-файлы и readiness scripts.
- CLI: `./.speckeep/scripts/run-speckeep[SCRIPT_EXT]` — только для реальных CLI-команд (`doctor`, `check`, `trace`, `export`, `refresh`). Не запускай `run-speckeep[SCRIPT_EXT] <phase>`: фазы — slash-команды, артефакты пишутся напрямую.
- Close: по умолчанию `speckeep archive` (PR/ручной merge, deferred/abandoned). Только если пользователь явно просит merge/close локальной feature-ветки — `speckeep finish` (archive + merge --no-ff + удаление локальной ветки). Если пользователь называет target-ветку (`dev`, `release/1.4`), передай `--into <branch>`; она должна существовать локально. Никогда не push и не удаляй remote-ветки.
- Вывод: без больших `git diff`, полных файлов и простыней логов; краткое резюме + список файлов, детали — небольшим фрагментом.
- Scope: не читай/меняй артефакты других slug без явной необходимости.
- Don't invent: не добавляй требования, зависимости, scope или AC, которых нет во входных артефактах.
- Эскалация: если фазу нельзя честно завершить (нет артефакта, неоднозначный intent, небезопасно, гейт заблокирован) — ОСТАНОВИСЬ и сообщи `Вернуться к: /spk-<phase> <slug>` или задай один точечный вопрос. Не выдумывай AC, не прячь пробелы и не угадывай шаг — остановка с точной причиной корректна.

Справочник команд: `.speckeep/templates/commands.md` — только когда нужной команды нет в текущей фазе. Триггеры repo-map — в `.speckeep/templates/prompts/repo-map.md` (`/spk-repo-map`).
