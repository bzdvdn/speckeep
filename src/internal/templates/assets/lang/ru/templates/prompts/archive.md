# Prompt архивации SpecKeep

Вы архивируете один feature package.

## Goal

Создайте устойчивый архивный снимок одной фичи, или восстановите ранее архивированную фичу обратно в активную разработку.

## Примечание о путях

Пути в этом промпте показаны для layout по умолчанию. Если в `.speckeep/speckeep.yaml` переопределены `paths.specs_dir` или `paths.archive_dir`, всегда следуйте путям из конфигурации, а не примерам по умолчанию.
Читайте `.speckeep/speckeep.yaml` максимум один раз за сессию для резолва путей; не перечитывайте его без необходимости (только если конфиг изменился или путь неоднозначен).

## Flags

`--copy`: оригиналы остаются на месте после архивации (режим только копирования). По умолчанию оригиналы удаляются после архивации; `--copy` их сохраняет. Полезно для `deferred` фич, которые могут вернуться в активную разработку.

`--restore`: обратная операция — копирует последний снимок обратно в активную `specs/<slug>/`, затем удаляет запись из архива. См. Restore Rules ниже.

## Вызов скрипта

**НЕ читайте и НЕ копируйте файлы вручную. Запустите скрипт напрямую — он сам валидирует статус verify и вернёт ошибку если предусловия не выполнены.**

По умолчанию используйте `--status completed`, если пользователь явно не передал другой. Валидные значения: `completed`, `superseded`, `abandoned`, `rejected`, `deferred`.

Если статус не `completed` и `--reason` не передан — запросите причину у пользователя перед запуском.

**Unix/macOS:**
```bash
./.speckeep/scripts/archive-feature.sh <slug> --status <status> [--reason "<причина>"]
```

**Windows (PowerShell):**
```powershell
.\.speckeep\scripts\powershell\archive-feature.ps1 <slug> --status <status> [--reason "<причина>"]
```

Примеры (Unix):
- `./.speckeep/scripts/archive-feature.sh my-feature`
- `./.speckeep/scripts/archive-feature.sh my-feature --status completed`
- `./.speckeep/scripts/archive-feature.sh my-feature --status deferred --reason "Перенесено на Q3" --copy`
- `./.speckeep/scripts/archive-feature.sh my-feature --restore`

Примеры (Windows):
- `.\.speckeep\scripts\powershell\archive-feature.ps1 my-feature`
- `.\.speckeep\scripts\powershell\archive-feature.ps1 my-feature --status completed`
- `.\.speckeep\scripts\powershell\archive-feature.ps1 my-feature --status deferred --reason "Перенесено на Q3" --copy`
- `.\.speckeep\scripts\powershell\archive-feature.ps1 my-feature --restore`

## Output expectations

### Режим по умолчанию

- Сообщите: `Готово к: ./.speckeep/scripts/archive-feature.sh <slug> --status <status>` (добавьте `--reason "..."`
  если статус не `completed`, или если пользователь явно хочет сохранить причину даже для `completed`)
- После выполнения: подтвердите успех, суммируйте статус и архивированные файлы
- Укажите, что archive — это терминальный шаг workflow для этой фичи

### Режим restore

- После выполнения: подтвердите восстановление файлов и перечислите их пути
- Укажите, что восстановленный спек не верифицирован — существующий `inspect.md` может быть устаревшим относительно текущей кодовой базы
- Завершите строкой: `Готово к: /speckeep.inspect <slug>` (повторный inspect обязателен перед возобновлением планирования)
