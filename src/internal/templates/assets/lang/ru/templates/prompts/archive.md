# Prompt архивации SpecKeep

Архивируете один feature package.

Следуйте базовым правилам в `AGENTS.md` (пути, git, load discipline, readiness scripts, язык, phase discipline).

## Goal

Устойчивый архивный снимок одной фичи или восстановление ранее архивированной фичи обратно в активную разработку.

## Flags

`--copy`: оригиналы остаются на месте (copy-only). По умолчанию оригиналы удаляются; `--copy` сохраняет. Полезно для `deferred`-фич.

`--restore`: копирует последний снимок обратно в активную `specs/<slug>/`, затем удаляет запись из архива.

## Вызов скрипта

**НЕ читайте и НЕ копируйте файлы вручную. Запустите скрипт напрямую — он сам валидирует verify-статус и вернёт ошибку если предусловия не выполнены.**

`--status` по умолчанию `completed`. Валидные: `completed`, `superseded`, `abandoned`, `rejected`, `deferred`. Если статус не `completed` и `--reason` не передан — запросите причину у пользователя.

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

## Output

### Default mode

- `completed` (по умолчанию): `Готово к: ./.speckeep/scripts/archive-feature.sh <slug>`
- Другой статус: `Готово к: ./.speckeep/scripts/archive-feature.sh <slug> --status <status> --reason "<причина>"` (+ `--copy` при необходимости).
- После выполнения: подтвердите успех, суммируйте статус и архивированные файлы.
- Укажите, что archive — terminal workflow step for this feature.

### Restore mode

- После выполнения: подтвердите восстановление, перечислите пути.
- Отметьте, что восстановленный спек не верифицирован — `inspect.md` может быть устаревшим.
- Готово: `Готово к: /speckeep.inspect <slug>` (повторный inspect обязателен перед планированием).
