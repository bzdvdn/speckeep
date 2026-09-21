# Агенты

## Поддерживаемые agent targets

SpecKeep умеет генерировать project-local command или prompt files для:

- `claude`
- `codex`
- `copilot`
- `cursor`
- `kilocode`
- `opencode`
- `trae`
- `windsurf`
- `roocode`
- `aider`
- `amazonq`
- `gemini`
- `jules`
- `cline`
- `devin`
- `goose`
- `refact`
- `codiumate`
- `qwen-code`
- `all`

### Уровни поддержки

Все таргеты используют один и тот же генератор (`agents/skill_pack.go`) и получают идентичный композитный `sdd` skill pack, так что каждый таргет функционально полон. Уровни ниже описывают только то, насколько активно проверяется экосистем-специфичная интеграция (конвенции директорий skills, особенности загрузчика) под реальное поведение инструмента:

- **Tier 1 — maintained-first**: `claude`, `cursor`, `codex`, `copilot`, `windsurf`, `opencode`. Им уделяется приоритетное внимание, когда меняется поведение экосистемы агентов (загрузчик skills, конвенции директорий).
- **Tier 2 — best-effort / community**: все остальные таргеты (`kilocode`, `trae`, `roocode`, `aider`, `amazonq`, `gemini`, `jules`, `cline`, `devin`, `goose`, `refact`, `codiumate`, `qwen-code`). Генерируются так же и должны работать, но экосистем-специфичные нюансы чинятся по репорту, а не отслеживаются проактивно.

Если вы используете Tier 2 таргет и находите несоответствие в том, как инструмент реально загружает skills — заведите issue: это и есть сигнал, который поднимает таргет на уровень выше.

### Слэш-вызов: что реально проверено

Каждый target получает идентичный обзорный скилл `sdd/` плюс по одному скиллу `spk-<фаза>/` на каждую фазу в своей `skills/`-эквивалентной директории. А вот работает ли реально `/spk-spec` (или `@spk-spec`) — зависит от того, как конкретно рантайм этого инструмента трактует Skills-директорию — это runtime-поведение вне контроля speckeep и оно отличается по инструментам. Все 19 таргетов теперь проверены индивидуально по официальной документации:

| Target | Вызывается по имени? | Заметки |
| --- | --- | --- |
| `claude` | Да, `/name` | Подтверждено по официальной документации Claude Code. |
| `cursor` | Да, `/name` | Подтверждено по официальной документации Cursor. |
| `qwen-code` | Да, `/name` | Подтверждено — Skills реально слэш-вызываемы, путь совпадает с текущим выводом speckeep (`.qwen/skills/`). |
| `codex` | Вероятно, `/name` | Skills слэш-вызываемы по докам OpenAI, но *канонический* задокументированный путь — `.agents/skills/`, не `.codex/skills/` — `.codex/skills/` задокументирован только в community-источниках и не подтверждён независимо как читаемый Codex CLI. |
| `copilot` | Да (CLI), `/name` | Подтверждено для Copilot **CLI** по официальной документации. У VS Code Copilot Chat своя отдельная документация "Agent Skills", паритет с ней не перепроверен. |
| `windsurf` | Нет через Skills — **да через сгенерированный Workflow** | Windsurf Skills загружаются автоматически или через `@mention`, никогда через `/`. speckeep также генерирует `.windsurf/workflows/spk-<фаза>.md` (реальный slash-механизм Windsurf), так что `/spk-spec` там реально работает. |
| `opencode` | Нет через `/` — фазы через `/skills` | OpenCode Skills загружаются вызовом агентом tool `skill()`, никогда через `/`; палитра `/` строится из `.opencode/commands/*` и явно пропускает `source === "skill"`. speckeep **поставляет OpenCode в режиме skills-only**: дублирующий набор `.opencode/commands/spk-<фаза>.md` удалён, поэтому фазы выбираются через `/skills`, а не набираются как `/spk-spec`. |
| `cline` | Нет через Skills (`.clinerules/` — это контекст, не команды) — **да через сгенерированный Workflow** | speckeep также генерирует `.clinerules/workflows/spk-<фаза>.md` (реальный slash-механизм Cline, вызывается как `/spk-<фаза>.md`). |
| `amazonq` | Нет через `/`, но **да через сгенерированный Prompt с `@name`** | У Amazon Q Developer нет подтверждённого нативного Skills-ридера вообще (единственная официальная документация "Skills" принадлежит отдельному плагину "Agent Toolkit for AWS"). speckeep теперь пишет в реальный dotfile root `.amazonq/` (не `.q/` — старый путь, вероятно, никогда не читался) и генерирует `.amazonq/prompts/spk-<фаза>.md`, вызывается через `@spk-<фаза>`, не слэшем. |
| `gemini` | Нет через Skills — **да через сгенерированную TOML-команду** | Gemini CLI Skills — только model-invoked. speckeep также генерирует `.gemini/commands/spk-<фаза>.toml` (реальный формат команд Gemini — TOML, не markdown), так что `/spk-spec` там реально работает. |
| `devin` | Нет, только `@skills:name` | Skills загружаются автоматически или через явное упоминание `@skills:name`, никогда через `/`. `.agents/skills/` — задокументированный канонический путь; `.devin/skills/` у speckeep — подтверждённо валидный fallback-путь, просто не основной. |
| `goose` | Нет, только meta-команда `/skills <name>` | Не вызывается напрямую как `/name` — доступ через `/skills <name>` или авто-триггер. Реальные slash-команды Goose — отдельная система "recipes" (параметризованные YAML-задачи с пользовательскими алиасами), которую speckeep не генерирует. `.agents/skills/` — канонический путь; `.goose/skills/` — подтверждённо валидный fallback. |
| `kilocode` | Не проверено / вероятно нет через Skills | Community-документация описывает Skills как только model-invoked, с реальными slash-командами по пути **`.kilo/commands/`** — но та же документация также намекает, что dotfile root продукта мог смениться с `.kilocode/` на `.kilo/`, что не подтверждено независимо. speckeep **не стал** менять `.kilocode/skills/` на основании одного источника; считайте эту строку открытой до подтверждения. |
| `roocode` | Нет, и плоского command-fallback не существует | Skills — только model-invoked. В отличие от Windsurf/Cline, у Roo Code нет задокументированного one-file-per-command механизма — единственная project-local конвенция (`.roo/rules/*.md`) — это конкатенированный свободный контекст, не вызываемый по отдельности. Сейчас у speckeep нет способа дать Roo Code реальный `/spk-spec`. |
| `trae` | Не проверено, вероятно нет | Документация Trae описывает natural-language/авто-выбор триггеринг, а не `/name`. Открытый upstream issue также говорит, что drop-in `SKILL.md`-файлы могут не подхватываться автоматически без регистрации через UI/CLI Trae. |
| `aider` | N/A — нет Skills-загрузчика | Подтверждено: у Aider нет ни Skills, ни custom-slash-command механизма. Он читает `CONVENTIONS.md` только когда явно попросить (`/read`, `--read` или `.aider.conf.yml`) — существующий указатель `.aider/CONVENTIONS.md` остаётся верным подходом. |
| `jules` | N/A — механизма нет вообще | Jules — async/неинтерактивный агент; он читает только `AGENTS.md` (fallback на `README.md`) для контекста. Ни Skills, ни slash-команды не применимы — вывод `.jules/skills/` у speckeep сейчас не читается, но и не вредит. |
| `refact` | Не проверено, вероятно нет | Ни Skills, ни SKILL.md, ни project-local command-конвенции не найдено нигде в докам Refact.ai или на GitHub. Кастомизация там UI-driven ("AI Toolbox", `Alt+T`), а не repo-local файл. Вывод `.refact/skills/` у speckeep не подтверждён как читаемый вообще. |
| `codiumate` | Не проверено | **Продукт переименован**: CodiumAI → Qodo → "Codiumate" теперь называется "Qodo Gen". У Qodo есть система, совместимая с открытым Agent-Skills-стандартом, но ни один официальный источник не подтверждает установленный project-path или слэш-вызываемость — считайте неподтверждённым, не предполагайте паритет. |

Если у вас нужная команда не появляется на каком-то таргете — заведите issue с тем, что вы видите: это и есть сигнал, который переводит строку из "не проверено" в подтверждённую (или, как с windsurf/cline/amazonq/gemini, в target-специфичный фикс).

**`continue` убран из поддерживаемых таргетов.** Несколько вторичных источников сообщили, что Continue.dev был поглощён Cursor и свёрнут примерно в середине 2026 (репозиторий на GitHub переведён в read-only) — поддерживать генерацию для него не имеет смысла. `speckeep doctor` подсвечивает оставшийся вывод `.continue/skills/` с момента до удаления, а `speckeep refresh`/`cleanup-agents` его убирают; запись `continue`, оставшаяся в существующем `speckeep.yaml`, при следующем refresh тихо отбрасывается, а не приводит к ошибке.

## Куда пишутся файлы

Каждый target получает один и тот же набор скиллов в `<skills-dir-таргета>/`: лёгкий обзорный скилл `sdd/` плюс по одному независимому, самодостаточному скиллу `spk-<фаза>/` на каждую фазу (например `spk-spec/`, `spk-plan/`, `spk-implement/`, ...), задуманному как напрямую слэш-вызываемый — какие таргеты это реально подтверждают, см. таблицу выше:

- Claude: `.claude/skills/{sdd,spk-*}/`
- Codex: `.codex/skills/{sdd,spk-*}/`
- Copilot: `.github/skills/{sdd,spk-*}/`
- Cursor: `.cursor/skills/{sdd,spk-*}/`
- Kilo Code: `.kilocode/skills/{sdd,spk-*}/`
- OpenCode: `.opencode/skills/{sdd,spk-*}/` (skills-only; без генерации slash-команд)
- Trae: `.trae/skills/{sdd,spk-*}/`
- Windsurf: `.windsurf/skills/{sdd,spk-*}/` + `.windsurf/workflows/spk-*.md` (real slash mechanism)
- Roo Code: `.roo/skills/{sdd,spk-*}/`
- Aider: `.aider/skills/{sdd,spk-*}/` (+ указатель `.aider/CONVENTIONS.md`)
- Amazon Q: `.amazonq/skills/{sdd,spk-*}/` + `.amazonq/prompts/spk-*.md` (real mechanism, invoked with `@name`)
- Gemini: `.gemini/skills/{sdd,spk-*}/` + `.gemini/commands/spk-*.toml` (real slash mechanism)
- Jules: `.jules/skills/{sdd,spk-*}/`
- Cline: `.clinerules/skills/{sdd,spk-*}/` + `.clinerules/workflows/spk-*.md` (real slash mechanism)
- Devin: `.devin/skills/{sdd,spk-*}/`
- Goose: `.goose/skills/{sdd,spk-*}/`
- Refact: `.refact/skills/{sdd,spk-*}/`
- Codiumate: `.codiumate/skills/{sdd,spk-*}/`
- Qwen Code: `.qwen/skills/{sdd,spk-*}/`

`<target>/sdd/SKILL.md` — обзор (цепочка workflow, гейты) на случай, когда фаза не очевидна из запроса; каждый `<target>/spk-<фаза>/SKILL.md` — свой top-level скилл, напрямую слэш-вызываемый (например `/spk-spec`), а не доступный только через открытие моделью связанного файла — инлайнит тело канонического промпта из `.speckeep/templates/prompts/`. CLI остаётся детерминированным проверяющим остовом — скилл никогда не заменяет вердикты `check`/`guard`. Для инструментов, которые не индексируют skills-директорию автоматически (напр., Aider), генерируется один loader-указатель; managed-блок в `AGENTS.md` тоже всегда ссылается на скиллы.

## Агентская дисциплина

Agent-facing workflow в SpecKeep:

- `constitution`
- `spec`
- `inspect`
- `plan`
- `tasks`
- `implement`
- `verify`

Каждый prompt должен:

- читать только минимально нужный контекст
- останавливаться при отсутствии prerequisites
- уважать configured documentation language и agent language
- считать конституцию документом с наивысшим приоритетом

Каждая сгенерированная обёртка для агента включает:

- **Hint цепочки workflow**: `constitution → spec → [inspect, опционально] → plan → tasks → implement → verify → archive` — не позволяет агентам пропускать обязательные фазы или забегать вперёд, а archive оставляет явным CLI-шагом после agent verify
- **Дисциплина выполнения скриптов**: явная инструкция выполнять скрипты как shell-команды, доверять stdout/exit code и никогда не читать исходники скриптов
- **Блок anti-patterns**: типичные ошибки, которых следует избегать — пропуск readiness scripts, перепланирование во время implement, отметка задач без observable proof, чтение всего репозитория когда нужен минимальный контекст

`spec` должен оставаться branch-first:

- перед записью `specs/active/<slug>/spec.md` он должен создавать или переключать `feature/<slug>`, когда окружение это позволяет
- он должен поддерживать `--name`, optional `--slug` и optional `--branch` для chat-oriented ввода
- если `/spk-spec` вызван с `--name`, но без достаточного описания, он должен сохранить контекст и запросить или принять следующее сообщение как продолжение spec-запроса
- если вход приходит из локального prompt-файла, он должен предпочитать `name:` и опциональный `slug:` в начале файла вместо generic filename
- если запрос неоднозначен, охватывает несколько фич, похож на URL или пытается вывести одну spec из нескольких изменений конституции, он должен остановиться и запросить одно конкретное изменение

`verify` специально сделан легким:

- он стартует от `tasks.md`
- он может использовать `.speckeep/scripts/verify-task-state.sh <slug>` как дешевый helper первого прохода
- обёртки `.speckeep/scripts/*` вычисляют корень проекта и передают его через `--root`, поэтому их можно запускать из любого текущего каталога
- более глубокие артефакты читаются только когда нужно подтвердить конкретный вывод
- его задача — подтвердить готовность к архивированию или следующему refine-циклу, а не превращаться в тяжелый review engine

После `verify: pass` предпочитайте явный CLI-шаг `speckeep archive <slug> .`, чтобы архивирование оставалось вне reasoning-loop агента. Для `archive` статус по умолчанию — `completed`; для non-`completed` статусов нужен явный `--reason`. Для `completed` нормально (и дешево) сначала переиспользовать `verify-task-state.sh` перед созданием снимка.

## Команды обслуживания

Управлять agent targets лучше через публичный CLI:

```bash
speckeep add-agent my-project --agents claude --agents cursor
speckeep list-agents my-project
speckeep remove-agent my-project --agents cursor
speckeep cleanup-agents my-project
speckeep doctor my-project
```
