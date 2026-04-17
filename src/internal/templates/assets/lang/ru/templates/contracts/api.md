# API Контракт

> **Создавайте этот файл только когда** фича трогает API-границу (новый endpoint, изменённая форма request/response, новый error-contract, breaking client-visible изменение). Если API contract не нужен — не создавайте файл.

## Scope

- Связанные `AC-*`: `AC-001`
- Связанные `DEC-*`: `DEC-001`

## API-001 Граница 1

- Назначение:
- Trigger:
- Inputs:
  - `field_name` - тип или shape, required или optional, смысл
- Outputs:
  - `field_name` - тип или shape, смысл
- Ошибки:
- Idempotency / Ordering:
- Notes:
- Связанные `AC-*`:

## API-002 Граница 2 — та же структура (Назначение / Trigger / Inputs / Outputs / Ошибки / Idempotency / Notes / Связанные AC)

## Заметки

- Фиксируйте только те границы API, которые существенно влияют на реализацию
