# <Spec Title> Задачи

## Phase Contract

Inputs: plan и минимальные supporting артефакты для этой фичи.
Outputs: упорядоченные исполнимые задачи с покрытием критериев.
Stop if: задачи получаются расплывчатыми или coverage не удается сопоставить.

## Surface Map

| Surface | Tasks |
|---------|-------|
| src/models/feature.ts | T1.1, T1.2, T2.1 |
| src/handlers/feature.ts | T2.1, T2.2 |
| src/tests/feature.test.ts | T3.1 |

## Фаза 1: Основа

Цель: подготовить минимальную структуру, contracts или data prerequisites, чтобы дальнейшая работа была предсказуемой.

- [ ] T1.1 Создать или выровнять базовый каркас фичи — implementation entrypoint существует и соответствует запланированной поверхности. Touches: src/models/feature.ts
- [ ] T1.2 Добавить базовые model, contract, migration или flag-изменения, если без них зависят последующие фазы. Touches: src/models/feature.ts, src/db/migrations/

## Фаза 2: Основная реализация

Цель: реализовать основное поведение фичи и важные edge или failure paths.

- [ ] T2.1 Реализовать основной acceptance path — ключевое поведение работает end to end на нужной поверхности. Touches: src/handlers/feature.ts, src/models/feature.ts
- [ ] T2.2 Реализовать edge, failure, permission или conflicting-state поведение, если оно меняет наблюдаемый результат. Touches: src/handlers/feature.ts

## Фаза 3: Проверка

Цель: доказать, что фича работает, и оставить пакет в reviewable состоянии.

- [ ] T3.1 Добавить или обновить automated coverage — tests или checks подтверждают поведение и страхуют от регрессий. Touches: src/tests/feature.test.ts
- [ ] T3.2 Выполнить verify, cleanup или documentation updates, нужные для review или verify

## Покрытие критериев приемки

- AC-001 -> T1.2, T2.1, T3.1
- AC-002 -> T2.2, T3.1, T3.2

## Заметки

- Сохраняйте порядок задач согласованным с планом и переносите работу в поздние фазы только если она реально зависит от ранних
- Используйте phase-scoped task IDs в формате `T<phase>.<index>`
- Делайте каждую задачу конкретной, измеримой и исполнимой как один связный кусок работы
- Предпочитайте action verbs, связанные с наблюдаемым результатом: implement, add, migrate, validate, remove, backfill
- По возможности ссылайтесь в тексте задач на 1-2 стабильных ID (`AC-*`, `RQ-*`, `DEC-*`)
- Не прячьте proof внутри большой implementation-задачи, а выносите validation отдельно
- Отмечайте задачи выполненными по мере реализации и не оставляйте критерии приемки без покрытия задачами
- Явно укажите, если какая-то фаза осознанно пропущена, потому что фиче она не нужна
