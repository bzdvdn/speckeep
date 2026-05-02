# Архитектура

На этой странице описано внутреннее устройство SpecKeep.

## Основные Слои

SpecKeep разбит на несколько практических слоев:

- `src/cmd/speckeep/` для CLI entrypoint
- `src/internal/cli/` для Cobra-команд и пользовательской сборки CLI
- `src/internal/project/` для lifecycle-операций workspace, таких как `init`, обслуживание агентов и cleanup
- `src/internal/config/` для загрузки, применения defaults и сохранения `.speckeep/speckeep.yaml`
- `src/internal/templates/` для локализованных generated assets и сборки файлов
- `src/internal/agents/` для генерации project-local command или prompt files
- `src/internal/specs/` для чтения spec files, используемых публичным CLI
- `src/internal/trace/` для сканирования аннотаций прослеживаемости в коде
- `src/internal/doctor/` для health-check логики

## CLI-Слой

Публичный CLI намеренно остается маленьким. В `src/internal/cli/` подключаются команды:

- `init`
- `add-agent`
- `list-agents`
- `remove-agent`
- `cleanup-agents`
- `doctor`
- `dashboard`
- `trace`
- `list-specs`
- `show-spec`

CLI-слой должен оставаться тонким. Основное поведение лучше держать в пакетах project, config, templates, agents и doctor.

## Слой Конфигурации

Источник истины для config лежит в `.speckeep/speckeep.yaml`.

`src/internal/config/config.go` отвечает за:

- схему config
- применение defaults
- резолв ключевых путей workspace
- загрузку config с диска
- сохранение обновленного config обратно на диск

Это позволяет остальному коду не захардкодить слишком много путей вида `.speckeep/...`.

## Слой Templates и Assets

SpecKeep генерирует файлы из локализованных assets, которые лежат в:

- `src/internal/templates/assets/lang/en/`
- `src/internal/templates/assets/lang/ru/`

Туда входят:

- `constitution.md`
- шаблоны spec, plan, tasks и archive
- prompts для `constitution`, `spec`, `inspect`, `plan`, `tasks`, `implement` и `verify`
- локализованный `agents-snippet.md`

Общие shell scripts лежат в:

- `src/internal/templates/assets/scripts/`

Пакет `templates` собирает эти assets в generated workspace `.speckeep/`.

## Слой Генерации Агентов

`src/internal/agents/files.go` генерирует project-local файлы для поддерживаемых targets:

- `claude`
- `codex`
- `copilot`
- `cursor`
- `kilocode`
- `trae`

Эти generated files являются обертками, которые ссылаются на канонические prompts SpecKeep в `.speckeep/templates/prompts/`.

Так сохраняется один главный источник истины для workflow prompts при поддержке нескольких агентных экосистем.

## Слой Жизненного Цикла Проекта

`src/internal/project/init.go` управляет lifecycle workspace:

- инициализация проекта
- добавление agent targets
- получение списка agent targets
- удаление agent targets
- очистка orphaned agent artifacts
- создание вспомогательных скриптов в `.speckeep/scripts/`

Этот слой отвечает за:

- создание layout `.speckeep/`
- запись generated files
- обновление `AGENTS.md`
- синхронизацию enabled agent targets в config

## Слой Прослеживаемости и Верификации

Пакет `src/internal/trace/` реализует основную логику связи кода с требованиями:

- сканирует файлы на наличие аннотаций `@sk-task` и `@sk-test`
- фильтрует находки по слагу фичи
- поддерживает JSON-вывод для интеграции с workflow верификации

Этот слой позволяет фазе `verify` оставаться экономной по токенам, заменяя ручной анализ кода детерминированным сканированием метаданных.

## Слой Здоровья и Обслуживания

`src/internal/doctor/doctor.go` и `src/internal/gitutil/` проверяют здоровье и согласованность workspace.

Он валидирует:

- наличие обязательных директорий и файлов
- корректность настроенных языков
- наличие generated files для включенных agent targets
- отсутствие незамеченных stale artifacts от отключенных targets
- **Smart Branching**: проверяет соответствие текущей ветки Git слагу фичи (ожидается `feature/<slug>`)
- **Traceability Integrity**: предупреждает об осиротевших аннотациях `@sk-task` или невалидных ссылках на `AC-*`

Связанные maintenance-команды:

- `remove-agent` отключает target и удаляет его generated files
- `cleanup-agents` удаляет orphaned leftovers для отключенных targets
- `doctor` сообщает `error`, `warning` и `ok`
- `dashboard` предоставляет визуальную сводку прогресса, статуса и здоровья веток проекта

## Принципы Дизайна

Внутренняя архитектура следует нескольким важным принципам:

- держать публичный CLI маленьким
- делать generated assets language-aware, но структурно согласованными
- выносить readiness checks в shell scripts, когда это возможно
- держать один канонический источник prompt-логики вместо дублирования
- сохранять strict workflow phases без тяжелого orchestration engine
- экономить токены через контроль read sets и scope артефактов

## Anti-Bloat Checklist

Используйте этот checklist перед добавлением новой фичи, prompt-правила, скрипта или артефакта:

- Увеличивает ли это default read set? Если да, по умолчанию это риск.
- Можно ли решить задачу дешевым deterministic helper вместо дополнительного reasoning?
- Делает ли это новый артефакт обязательным для каждой фичи? Если да, стоит пересмотреть решение.
- Требует ли это чтения implementation code по умолчанию? Если да, решение, скорее всего, слишком тяжелое.
- Можно ли начинать workflow от constitution, spec, plan или tasks до чтения кода?
- Расширяет ли это публичный CLI без явной пользы для повседневной работы?
- Добавляет ли это совершенно новый gate, или можно усилить уже существующую фазу?
- Можно ли объяснить ценность изменения одним коротким предложением? Если нет, есть риск лишней процессной сложности.
- Приближает ли это SpecKeep к бюрократии в стиле spec-kit без сопоставимой пользы?
- Улучшает ли это brownfield ergonomics на практике?

Изменение обычно хорошо вписывается в SpecKeep, если выполняет хотя бы одно из двух:

- повышает качество без расширения default context
- заменяет дорогой reasoning дешевой структурной проверкой
