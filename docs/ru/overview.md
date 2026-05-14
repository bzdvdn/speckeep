# Обзор

## Что такое SpecKeep

`speckeep` хранит намерение проекта, спецификации, feature artifacts и задачи в обычных файлах. Цель инструмента — дать людям и агентам один и тот же проектный контекст без жесткого процессного движка.

## Базовые идеи

- Конституция — главный документ проекта.
- Каждая фича проходит через строгую цепочку workflow.
- Каждая фича должна жить в отдельной git-ветке, чтобы и одиночная, и командная работа не упирались в конфликты общего mutable state.
- Генерируемые документы и промты могут быть на английском или русском.
- Агентские workflow должны читать только минимально нужный контекст.
- Проверки готовности лучше выносить в scripts, а не в лишние токены модели.

## Структура workspace

```text
.speckeep/
  speckeep.yaml
  templates/
  scripts/
CONSTITUTION.md
specs/
  active/
    <slug>/
      spec.md
      inspect.md
      hotfix.md
      plan.md
      tasks.md
      data-model.md
      verify.md
      research.md
      contracts/
        api.md
        events.md
  archived/
    <slug>/
      <YYYY-MM-DD>/
        summary.md
        spec.md
        plan.md
        tasks.md
        data-model.md
        research.md
        contracts/
AGENTS.md
```

`tasks.md` — главный operational entrypoint для `implement` и `verify`; `spec.md` и `plan.md` остаются source-артефактами, а `.speckeep/constitution.summary.md` — предпочитаемый компактный policy-контекст, когда нужны правила конституции.

## Публичный CLI

Публичная поверхность CLI намеренно маленькая:

- `speckeep init [path]`
- `speckeep add-agent [path]`
- `speckeep list-agents [path]`
- `speckeep remove-agent [path]`
- `speckeep cleanup-agents [path]`
- `speckeep doctor [path]`
- `speckeep list-specs [path]`
- `speckeep show-spec <name> [path]`

Создание и развитие spec, plan, tasks и implement остаются agent-facing workflow, а не публичными subcommand'ами CLI.
