## SpecKeep

Контекст: `.speckeep/`. Языки: docs=[DOCS_LANGUAGE], agent=[AGENT_LANGUAGE], comments=[COMMENTS_LANGUAGE]

Цепочка: `constitution → spec → [inspect, опц.] → [plan, опц.] → tasks → implement → archive`; `verify` — опц. аудит (по умолчанию пропускается). `/spk-propose` — дефолт для новой фичи (spec+tasks за проход): не грузи `sdd`-роутер; `/spk-spec` — только для неоднозначных/мульти-фичевых идей; plan — opt-in для дизайна. `workflow.verify` в конфиге: `required` = обязательный гейт перед archive.

Базовые правила:

- ⚠️ **Repository map first**: для навигации не используй `ls`/`find`/glob — сначала `REPOSITORY_MAP.md` (экономит токены). Раз за сессию; перечитывай только после своего обновления.
- Пути/конфиг: `.speckeep/speckeep.yaml` читай ≤ 1 раза за сессию: `<specs_dir>`, `<archive_dir>`, `project.constitution_file`. Нет конфига → `specs/active`, `specs/archived`, `CONSTITUTION.md`.
- Конституция: сначала `.speckeep/constitution.summary.md`, если есть; иначе `project.constitution_file` из конфига.
- Доменный язык: `.speckeep/glossary.md` — раз за сессию, если есть; переиспользуй термины, синонимы не вводи. Создаётся только `/spk-glossary`.
- Ветки: только `/spk-spec`/`/spk-propose` переключают/создают `feature/<slug>` (или `--branch`); остальные фазы уже должны быть на нужной ветке.
- Скрипты: `./.speckeep/scripts/check-ready[SCRIPT_EXT] <phase> <slug>` — только для гейтируемых фаз (`constitution`, `spec`, `propose`, `inspect`, `plan`, `tasks`, `implement`, `verify`, `converge`); aux-командам гейт не нужен. Доверяй exit code; исходники `.speckeep/scripts/*` не читай.
- Скиллы: workflow — skill-pack (`sdd`-роутер + `spk-<phase>/SKILL.md` на фазу). Грузи только нужный фазовый скилл; `sdd` — когда фаза неочевидна.
- Scope/load: только текущий slug; без широких сканов и чужих артефактов; surfaces из `Touches:`. Формы — из `.speckeep/templates/*`.
- Git: без `commit/push/tag` и PR без явной просьбы.
- Done: задача закрыта только с observable proof (файл, тест, вывод команды); артефакт должен быть понятен коллеге без пояснений.
- Proof: под каждой `[x]` задачей в `tasks.md` — строка `Proof: <kind> <path> [<anchor>]` (`code|test|docs|chore`). Без `Proof:` задача не закрыта; `trace`, `doctor` и archive читают только эти строки.
- Trace: после `feature/<slug>` выполни `./.speckeep/scripts/run-speckeep[SCRIPT_EXT] trace init <slug>`. `Touches:` — план; `trace <slug> --files` — факт (`Unmapped` = drift); `trace --file <path>` — кто менял файл.
- Токен-дисциплина: та же фаза → продолжай в сессии (`continue`); не перевызывай skill (перезагрузка промпта). Переход фазы — одна загрузка; `Готово к` называет следующий шаг, а не повтор.
- End block: каждая фаза завершается каноническим блоком ниже (`Вернуться к` при blocked).
  - Канонический end block (единственный источник; prompt'ы ссылаются на него):
    ```
    Slug: <slug>
    Status: <фаза>
    Artifacts: <пути>
    Blockers: <none | причина>
    Готово к: <следующая команда>   (или "Вернуться к: /spk-<phase> <slug>" при blocked)
    ```
  - `Готово к` может содержать `(alt: ...)`; первая команда — основная/безопасная.
- Политика verify-гейта (единственный источник; prompt'ы ссылаются по имени):
  - `required` → verify обязателен перед archive: после implement/hotfix/handoff `Готово к` = `/spk-verify <slug>`; archive требует `verify.md` со `status: pass`.
  - `optional`/нет ключа → on-demand аудит; archive разрешён, когда все `[x]` имеют `Proof:`. `verify.md` ≠ `pass` — вето на archive.
- Discovery: не запускай `speckeep --help` для разведки; используй prompt-файлы и readiness scripts.
- CLI: `./.speckeep/scripts/run-speckeep[SCRIPT_EXT]` — только для реальных CLI-команд (`doctor`, `check`, `trace`, `export`, `refresh`); не запускай `run-speckeep[SCRIPT_EXT] <phase>` — фазы это slash-команды.
- Close: по умолчанию `archive` (PR/ручной merge, deferred/abandoned); `finish` — только по явной просьбе merge/close локальной ветки. Target-ветка → `--into <branch>` (должна существовать локально). Никогда не push и не удаляй remote-ветки. `deferred`/`abandoned` не обходят `Proof:`; non-completed — только для реально отложенной работы.
- Вывод: без больших diff/полных файлов/логов; краткое резюме + список файлов, детали — небольшим фрагментом.
- Don't invent: не добавляй требования, зависимости, scope или AC, которых нет во входных артефактах.
- Эскалация: если фазу нельзя честно завершить (нет артефакта, неоднозначный intent, небезопасный change, заблокированный гейт) — ОСТАНОВИСЬ и сообщи `Вернуться к: /spk-<phase> <slug>` или задай один точечный вопрос. Не выдумывай pass, не прячь пробелы, не угадывай.

Справочник команд: `.speckeep/templates/commands.md` — когда фаза не назвала команду; триггеры repo-map — в `.speckeep/templates/prompts/repo-map.md` (`/spk-repo-map`).
