# Агенты

## Поддерживаемые agent targets

SpecKeep умеет генерировать project-local command или prompt files для:

- `claude`
- `codex`
- `copilot`
- `cursor`
- `kilocode`
- `trae`
- `all`

## Куда пишутся файлы

- Claude: `.claude/commands/`
- Codex: `.codex/prompts/`
- Copilot: `.github/prompts/`
- Cursor: `.cursor/rules/`
- Kilo Code: `.kilocode/rules/`
- Trae: `.trae/project_rules.md`

Эти generated files являются тонкой оберткой над каноническими промтами в `.speckeep/templates/prompts/`.

## Агентская дисциплина

Agent-facing workflow в SpecKeep:

- `constitution`
- `spec`
- `inspect`
- `plan`
- `tasks`
- `implement`
- `verify`
- `archive`

Каждый prompt должен:

- читать только минимально нужный контекст
- останавливаться при отсутствии prerequisites
- уважать configured documentation language и agent language
- считать конституцию документом с наивысшим приоритетом

Каждая сгенерированная обёртка для агента включает:

- **Hint цепочки workflow**: `constitution → spec → inspect → plan → tasks → implement → verify → archive` — не позволяет агентам пропускать фазы или забегать вперёд
- **Дисциплина выполнения скриптов**: явная инструкция выполнять скрипты как shell-команды, доверять stdout/exit code и никогда не читать исходники скриптов
- **Блок anti-patterns**: типичные ошибки, которых следует избегать — пропуск readiness scripts, перепланирование во время implement, отметка задач без observable proof, чтение всего репозитория когда нужен минимальный контекст

`spec` должен оставаться branch-first:

- перед записью `specs/<slug>/spec.md` он должен создавать или переключать `feature/<slug>`, когда окружение это позволяет
- он должен поддерживать `--name`, optional `--slug` и optional `--branch` для chat-oriented ввода
- если `/speckeep.spec` вызван с `--name`, но без достаточного описания, он должен сохранить контекст и запросить или принять следующее сообщение как продолжение spec-запроса
- если вход приходит из локального prompt-файла, он должен предпочитать `name:` и опциональный `slug:` в начале файла вместо generic filename
- если запрос неоднозначен, охватывает несколько фич, похож на URL или пытается вывести одну spec из нескольких изменений конституции, он должен остановиться и запросить одно конкретное изменение

`verify` специально сделан легким:

- он стартует от `tasks.md`
- он может использовать `.speckeep/scripts/verify-task-state.sh <slug>` как дешевый helper первого прохода
- обёртки `.speckeep/scripts/*` вычисляют корень проекта и передают его через `--root`, поэтому их можно запускать из любого текущего каталога
- более глубокие артефакты читаются только когда нужно подтвердить конкретный вывод
- его задача — подтвердить готовность к архивированию или следующему refine-циклу, а не превращаться в тяжелый review engine

Для `archive` статус по умолчанию — `completed`; для non-`completed` статусов нужен явный `--reason`. Для `completed` нормально (и дешево) сначала переиспользовать `verify-task-state.sh` перед созданием снимка.

## Команды обслуживания

Управлять agent targets лучше через публичный CLI:

```bash
speckeep add-agent my-project --agents claude --agents cursor
speckeep list-agents my-project
speckeep remove-agent my-project --agents cursor
speckeep cleanup-agents my-project
speckeep doctor my-project
```
