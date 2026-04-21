# Prompt проверки SpecKeep (compact)

Вы проверяете одну фичу на согласованность и готовность к планированию.

Следуйте базовым правилам в `AGENTS.md`.

## Разрешение путей

- Определите `<specs_dir>` из `.speckeep/speckeep.yaml` (читать ≤ 1 раза за сессию). Если конфиг отсутствует — используйте `.speckeep/specs`.

## Phase Contract

Inputs: `.speckeep/constitution.md`, `<specs_dir>/<slug>/spec.md`.
Outputs: `<specs_dir>/<slug>/inspect.md` со статусом `pass|concerns|blocked` и `<specs_dir>/<slug>/summary.md`.
Stop if: spec отсутствует, slug неоднозначен, или verdict потребовал бы выдумывать product intent.

## Проверки (строго, но дёшево)

- Всегда начинайте с самого дешёвого scope: constitution + spec, затем plan, затем tasks. В код — только если конкретный claim нельзя подтвердить из артефактов.
- Не делайте повторных full-file чтений «для спокойствия»: держите краткие заметки и переоткрывайте только нужные секции.
- Constitution ↔ spec: нет конфликтов с конституцией, workflow-правилами и language policy.
- `AC-*`: каждый AC в Given/When/Then; нет placeholder; нет незакрытых `[NEEDS CLARIFICATION: ...]`.
- Scope: строго одна фича; явные `Вне scope`, `Допущения`, `Открытые вопросы` (или `none`).
- Упоминания технологий: technology names/frameworks/library lists/version pins в spec — это Warning, если это не требование пользователя, не repo-constraint и не внешний contract.
- Неоднозначность: расплывчатые прилагательные (быстро, масштабируемо, безопасно, “удобно”, “надёжно”) без измеримых критериев — Warning; если это блокирует планирование — blocked.
- Плейсхолдеры: `TODO`, `TKTK`, `???`, `<placeholder>` и любые незакрытые маркеры — Error.
- Если есть `<specs_dir>/<slug>/plan/plan.md`: проверьте `spec <-> plan` (цель/scope сохранены; нет новых крупных workstreams).
- Если есть `<specs_dir>/<slug>/plan/tasks.md`: проверьте `plan <-> tasks` и покрытие AC (каждый `AC-*` покрыт ≥ 1 задачей).
- Если есть `<specs_dir>/<slug>/plan/tasks.md`: отсутствие `Touches:` — Warning (дефект token-discipline, который провоцирует широкие чтения на implement).

Если есть `/.speckeep/scripts/check-inspect-ready.*` — запустите (slug первым аргументом) и используйте вывод как baseline. Исходники `/.speckeep/scripts/*` не читать.

## Output expectations

- Запишите `inspect.md` и `summary.md` (summary ≤ ~25 строк: Goal, AC table, Out of Scope).
- В `inspect.md` обязательно: verdict, Errors, Warnings и Next step (если не blocked).
- Для `blocked` не предлагайте следующую фазу; явно укажите, какой refinement нужен.
- В разговоре дайте компактный verdict + непустые Errors/Warnings + Next step.
- Финальная строка: `Готово к: /speckeep.plan <slug>`
