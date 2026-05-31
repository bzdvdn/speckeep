# SpecKeep Repository Map

Обновить `REPOSITORY_MAP.md` — компактный, code-only навигационный индекс.

## Phase Contract

Inputs: состояние файловой системы проекта (spec не требуется).
Outputs: обновлённый `REPOSITORY_MAP.md` в корне проекта.
Stop if: структурных изменений нет (сначала проверьте чеклист триггеров).

## Политика

- Держите `REPOSITORY_MAP.md` компактным и code-only (пути + короткие роли).
- Language-agnostic: определяйте стек по маркерам репозитория (напр. `go.mod`, `package.json`, `pyproject.toml`, `Cargo.toml`, `pom.xml`, `*.csproj`) и адаптируйте секции под найденный стек.
- Не предполагайте Go-структуру для не-Go проектов.
- Жесткий лимит размера: целевой объем до 180 строк; если карта растет — сжимайте, а не расширяйте prose.
- Обновляйте in-place (минимальный diff): сохраняйте неизменные строки/порядок и правьте только затронутые записи/секции.
- Не переписывайте файл целиком, если изменилась только часть карты.
- Если `REPOSITORY_MAP.md` отсутствует — создайте по шаблону; если существует — патчите существующее содержимое.
- Исключайте из индексации: `src/internal/agents/**`, `.speckeep/**`, `specs/archived/**`, `.git/**`, `bin/**`, `demo/**`, `docs/**`, `TESTS/**`, `node_modules/**`, `vendor/**`, `dist/**`, `build/**`, `coverage/**`.
- Важно: проектные настройки уже читаются из `.speckeep/speckeep.yaml`; не дублируйте этот конфиг в карте.

## Шаблон

```md
# Repository Map

## Entry Points
- `<path>` — `<runtime/service/cli entrypoint>`

## Top-Level Code
- `<path>` — `<module role>`

## Key Paths
- `<path>` — `<what is implemented here>`

## Where To Edit
- `<change type>` — `<likely paths>`

## Excluded
- `<glob>` — `excluded from indexing`
```

## Ожидаемый вывод

- Перечислите изменённые/добавленные/удалённые записи.
- Подтвердите, что карта актуальна и укладывается в лимит размера.
- Включите компактный summary: `Slug`, `Status`, `Artifacts`, `Blockers`.
- Финальная строка: `Готово к: <следующая фаза>`.
