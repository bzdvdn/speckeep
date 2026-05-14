# Prompt проверки SpecKeep (compact)

Вы проводите опциональную глубокую проверку качества спецификации перед планированием. Эта фаза не обязательна — если спека прошла самопроверку и выглядит надёжно, пользователь может перейти напрямую к `/speckeep.plan`. Используйте inspect при неоднозначностях, сложном домене или когда нужен формальный quality gate.

Следуйте базовым правилам в `AGENTS.md`.

## Phase Contract

Inputs: `.speckeep/constitution.summary.md` (предпочтительно, если файл есть) или `project.constitution_file` (по умолчанию `CONSTITUTION.md`), `<specs_dir>/<slug>/spec.md`.
Outputs: `<specs_dir>/<slug>/inspect.md` со статусом `pass|concerns|blocked`.
Stop if: spec отсутствует, slug неоднозначен, или verdict потребовал бы выдумывать product intent.

## Проверки (строго, но дёшево)

- Всегда начинайте с самого дешёвого scope: constitution + spec, затем plan, затем tasks. В код — только если конкретный claim нельзя подтвердить из артефактов.
- Не делайте повторных full-file чтений «для спокойствия»: держите краткие заметки и переоткрывайте только нужные секции.
- Формат отчёта берите из `.speckeep/templates/inspect.md`. Не ищите «примеры» inspect-отчётов в других slug ради формы: это лишний токен‑расход и scope drift.
- Constitution ↔ spec: нет конфликтов с конституцией, workflow-правилами и language policy.
- Если нужен контекст конституции, сначала загрузите `.speckeep/constitution.summary.md`, если файл существует; только при его отсутствии переходите к `project.constitution_file`.
- `AC-*`: каждый AC в Given/When/Then; нет placeholder; нет незакрытых `[NEEDS CLARIFICATION: ...]`.
- Scope: строго одна фича; явные `Вне scope`, `Допущения`, `Открытые вопросы` (или `none`).
- Упоминания технологий: technology names/frameworks/library lists/version pins в spec — это Warning, если это не требование пользователя, не repo-constraint и не внешний contract.
- Неоднозначность: расплывчатые прилагательные (быстро, масштабируемо, безопасно, “удобно”, “надёжно”) без измеримых критериев — Warning; если это блокирует планирование — blocked.
- Плейсхолдеры: `TODO`, `TKTK`, `???`, `<placeholder>` и любые незакрытые маркеры — Error.
- Если есть `<specs_dir>/<slug>/plan.md`: проверьте `spec <-> plan` (цель/scope сохранены; нет новых крупных workstreams).
- Если есть `<specs_dir>/<slug>/tasks.md`: проверьте `plan <-> tasks` и покрытие AC (каждый `AC-*` покрыт ≥ 1 задачей).
- Если есть `<specs_dir>/<slug>/tasks.md`: отсутствие `Touches:` — Warning (дефект token-discipline, который провоцирует широкие чтения на implement).

Если есть `./.speckeep/scripts/check-inspect-ready.*` — запустите (slug первым аргументом) и используйте вывод как baseline. Исходники `./.speckeep/scripts/*` не читать.

## Output expectations

- Запишите `inspect.md`.
- Если нужен компактный recap по AC или scope, держите его внутри `inspect.md`; не создавайте отдельный `summary.md`.
- В `inspect.md` обязательно: verdict, Errors, Warnings и Next step (если не blocked).
- Для `blocked` не предлагайте следующую фазу; явно укажите, какой refinement нужен.
- В разговоре дайте компактный verdict + непустые Errors/Warnings + Next step.
- Финальная строка:
  - если `pass|concerns`: `Готово к: /speckeep.plan <slug>`
  - если `blocked`: `Вернуться к: /speckeep.spec <slug>`
