# Roadmap

Этот roadmap сфокусирован на ближайших практических итерациях SpecKeep, а не на длинном спекулятивном backlog.

## Направление

SpecKeep должен и дальше занимать позицию между тяжелыми spec-driven системами и более свободными change-driven системами:

- строже, чем OpenSpec
- легче, чем spec-kit
- лучше приспособлен для agent-first workflow на реальных кодовых базах

## Фокус Первого Релиза

Ближайшая задача SpecKeep не в том, чтобы догнать более тяжелые SDD-системы по количеству фаз, артефактов или автоматизации.

Ближайшая задача:

- выпустить легкий первый релиз
- проверить workflow "в бою" на реальных кодовых базах и реальных агентных сессиях
- подтвердить, что strict-by-structure подход работает без большого default context

До первых полевых проверок SpecKeep должен предпочитать:

- узкий default context вместо широкого чтения репозитория
- cheap checks и readiness scripts вместо тяжелого orchestration
- минимально обязательный набор артефактов вместо разрастания feature package
- улучшение traceability и consistency без роста prompt mass

Lightweight guardrails для ближайших итераций:

- у каждой фазы должны быть явные правила `always load`, `load if needed`, `never load by default`
- `implement` должен оставаться task-scoped по умолчанию и открывать более глубокие артефакты только по потребности активной задачи
- `verify` должен оставаться cheap-by-default и углубляться в код или широкий review только по явному запросу
- prerequisite checks нужно по возможности выносить в helper scripts и readiness checks, а не повторять reasoning в prompt
- traceability нужно усиливать через стабильные IDs и явные ссылки, а не через новые shared summary artifacts
- `archive` должен оставаться компактной историей, а не превращаться в новую mutable working-memory layer

До первых полевых проверок SpecKeep не должен спешить с:

- новыми обязательными фазами
- расширением default inspect/verify context
- automation, которая делает workflow тяжелее до подтверждения реальной пользы

## Итерация 1

### Главная цель

Усилить `inspect` как центральный слой качества.

Статус: обязательный persisted inspect report, общий parser для reports и базовая semantic validation уже реализованы.

Release filter: усиливать `inspect` только так, чтобы он оставался дешевым по контексту и не превращался в обязательный тяжелый review engine.

### Планируемая работа

- углублять semantic checks для `constitution <-> spec`
- углублять semantic checks для `spec <-> plan`
- углублять semantic checks для `plan <-> tasks`
- улучшать presentation inspect findings в CLI и docs
- продолжать усиливать acceptance-to-task traceability

### Anti-Bloat Notes

Безопасное направление:

- более сильные structural checks
- более четкая семантика verdict
- лучшая traceability через стабильные acceptance IDs
- дешевые spec <-> plan consistency только по spec.md и plan.md

Требует осторожности:

- читать implementation code по умолчанию во время inspect
- превращать inspect в широкий review engine
- подтягивать data-model, contracts и код в каждый запуск inspect по умолчанию

### Почему это важно

Если `inspect` сильный, все downstream-фазы становятся качественнее при меньшем количестве пустой реализации и переделок.

## Итерация 2

### Главная цель

Добавить легкий post-implementation verification layer.

Статус: легкий contract, prompt, readiness script, report template, evidence-oriented форма отчета и token-safe validator checks уже есть. Дальше нужно усиливать проверки и presentation, не расширяя default context.

Release filter: `verify` должен оставаться легким слоем безопасности, а не новым тяжелым review или QA engine.

### Планируемая работа

- проверять, что завершенные tasks соответствуют реальному состоянию implementation
- проверять, что implementation по-прежнему соответствует intent из spec и plan
- улучшать качество evidence в verify report и согласованность archive-readiness
- улучшать presentation verify findings и status в CLI
- следить, чтобы archived feature state и task state оставались согласованными там, где от этого зависит verification

### Anti-Bloat Notes

Безопасное направление:

- helper-скрипты для проверки состояния задач
- проверки consistency между archive и tasks
- evidence-oriented verify reports без расширения default reads
  Статус: для `verify` уже добавлены грубые helper-based sync checks и базовая semantic validation отчетов.

Требует осторожности:

- читать код по умолчанию во время verify
- превращать verify в тяжелый review или QA engine

### Почему это важно

Это закрывает разрыв между "tasks выполнены" и "фича реально соответствует задуманному дизайну".

## Итерация 3

### Главная цель

Усилить brownfield ergonomics и machine-readable outputs.

Release filter: добавлять automation output только там, где он переиспользует существующие проверки и не тянет новый обязательный контекст.

### Планируемая работа

- улучшить archive summaries и связи архива
- удерживать проверки completed-архива дешевыми за счет переиспользования task-state verification
- добавить machine-readable outputs вроде `doctor --json`
  Статус: уже реализовано для `doctor`; дальше расширять этот подход только там, где output остается дешевым и переиспользует существующие проверки.
- улучшить config-aware поведение scripts и будущих утилит
- продолжить выравнивать многоязычную консистентность docs и prompts

### Anti-Bloat Notes

Безопасное направление:

- machine-readable outputs для уже существующих проверок
- более удобная индексация и summaries архива
- config-aware helpers, уменьшающие повторный reasoning

Требует осторожности:

- archive-flow, который требует чтения широкой истории репозитория
- новые automation outputs, создающие обязательные артефакты
- brownfield helpers, которые незаметно расширяют default context

### Почему это важно

Это сделает SpecKeep удобнее для автоматизации, удобнее для эксплуатации в больших проектах и сильнее для долгоживущих brownfield-кодовых баз.

## Постоянная Работа Над Качеством

Параллельно с feature work SpecKeep должен продолжать улучшать:

- консистентность документации
- unit test coverage
- ergonomics CLI
- ясность prompts и token discipline
- качество brownfield workflow

## Что Пока Не Планируется

SpecKeep не стоит тащить в эти стороны без очень сильной причины:

- тяжелый orchestration engine
- обязательные checkpoint systems
- approval-gate бюрократию
- большие default prompt contexts
- обязательное разрастание артефактов для каждой фичи
- попытка стать "полным process OS" до того, как легкое ядро проверено на практике
