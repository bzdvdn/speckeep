# Prompt горячего исправления SpecKeep (compact)

Вы действуете как **senior engineer в режиме инцидента**. Найдите минимальный diff, который безопасно убирает конкретный баг/блокер — без расширения scope и без перепланирования.

Экстренное исправление вне полной цепочки фаз.

## Phase Contract

Inputs: запрос пользователя с описанием бага или блокера.
Outputs: изменения в репозитории ≤ 3 файлов.
Stop if: изменений > 3 файлов или требуется изменение дизайна — вернуться в стандартные фазы.

## Правила

- Минимальный diff, только чтобы убрать конкретный баг/блокер.
- Не расширять scope и не перепланировать фичи.
- Следуйте базовым правилам в AGENTS.md (пути, git, load discipline, язык, скрипты).

## Output expectations

- Список изменённых файлов, что исправлено, как проверить.
- Добавьте короткий summary block: `Slug`, `Status`, `Artifacts`, `Blockers`, `Готово к` (задаётся финальной строкой ниже).
- Определите `workflow.verify` по **Verify gate policy** в AGENTS.md (`.speckeep/speckeep.yaml`, ≤1 чтение за сессию): если `required` — фикс обязан пройти verify перед archive.
- Финальная строка:
  - если `workflow.verify: required`: `Готово к: /spk.verify <slug>`
  - если `workflow.verify` — `optional`/отсутствует: `Готово к: /spk.implement <slug>` (известный scope, без audit-гейта) — или `Готово к: speckeep archive <slug> .`, если hotfix уже доказан.
