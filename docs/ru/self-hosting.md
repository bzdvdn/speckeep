# Self-Hosting и Разработка

## Разработка самого SpecKeep

Внутри репозитория SpecKeep сгенерированные `/.speckeep/`, `/AGENTS.md` и `/TESTS/` считаются локальными development artifacts. Они полезны для smoke-тестов и self-hosting сценариев, но не являются исходниками продукта.

## Рекомендуемый локальный цикл

```bash
go test ./...
go build -o bin/speckeep ./src/cmd/speckeep
./bin/speckeep init TESTS/demo --git=false --lang en --agents claude --agents cursor
./bin/speckeep doctor TESTS/demo
./bin/speckeep doctor TESTS/demo --json
```

## Текущее Покрытие Тестами

В репозитории уже есть unit-тесты для:

- загрузки config, defaults, save и path resolution
- инициализации проекта и lifecycle-операций для агентов
- health-check логики через `doctor`
- списка spec, чтения spec и создания файлов из шаблонов
- согласованности локализованных template assets
- генерации agent files
- CLI-поведения на уровне команд

## Зачем нужны `doctor` и `cleanup-agents`

Когда ты тестируешь несколько agent targets, в проекте легко остаются stale generated files. SpecKeep разделяет эти задачи:

- `remove-agent` обновляет config и удаляет файлы для выбранных включенных targets
- `cleanup-agents` удаляет leftover artifacts для targets, которые уже не включены
- `doctor` показывает missing files как `error`, а leftover artifacts как `warning`

## Источники истины

Главные источники истины в этом репозитории:

- `src/` для Go-кода
- `src/internal/templates/assets/lang/` для локализованных generated assets
- `README.md` для краткого позиционирования продукта
- `MVP.md` для текущей продуктовой модели
- `docs/` для расширенной документации
