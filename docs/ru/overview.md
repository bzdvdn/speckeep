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

CLI намеренно операционный: управление воркспейсом, проверки, трассируемость и закрытие. Ежедневная работа со spec/plan/tasks/implement остаётся agent-facing.

- **Workspace**: `init`, `refresh`, `add-agent`, `list-agents`, `remove-agent`, `cleanup-agents`, `doctor`, `migrate`, `import`
- **Фичи**: `list-specs`, `show-spec`, `features`, `feature`, `feature repair`, `status`, `check`, `risk`, `export`, `context`
- **Трассируемость**: `trace`, `trace init`, `trace --files`, `trace --file`, `trace rebuild`
- **Закрытие**: `archive`, `finish`, `converge`, `guard`, `list-archive`
- **Скаффолдинг**: `demo`, `explore`, `schema`
- **Self**: `self check`, `self upgrade`

Каждая команда, флаг и контракт exit code описаны в [Справочнике CLI](/ru/cli).
