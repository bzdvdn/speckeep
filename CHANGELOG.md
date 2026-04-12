# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versions follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0] - 2026-04-11

### Added

- Initial release of the Speckeep CLI (specification-driven development kit for agents and humans)
- Canonical workspace under `.speckeep/` with file-based artifacts (specs, plan packages, reports, scripts, templates)
- Strict phase chain: `constitution → spec → inspect → plan → tasks → implement → verify → archive`
- Bilingual templates/prompts: English (`en`) and Russian (`ru`)
- Shell support: `sh` and `powershell`
- Core CLI:
  - `speckeep init`, `speckeep refresh`, `speckeep doctor`
  - `speckeep list-specs`, `speckeep show-spec`, `speckeep check`, `speckeep trace`
  - `speckeep feature`, `speckeep feature repair`, `speckeep features`, `speckeep migrate`
  - `speckeep export`, `speckeep demo`, `speckeep archive`, `speckeep list-archive`
- Managed agent integrations (generated wrapper files + prompts) for: Claude, Codex, Copilot, Cursor, Kilocode, Trae, Windsurf, Roocode, Aider
- Phase readiness scripts and internal CLI plumbing (`__internal`) to keep wrappers cheap and deterministic
- Stable IDs for traceability: `RQ-*`, `AC-*`, `DEC-*`, `T*` + acceptance coverage mapping (`AC-* -> T*`)
- Migration support from legacy `.draftspec/` workspace into `.speckeep/` (safe move/copy + path canonicalization)

[0.1.0]: https://github.com/bzdvdn/speckeep/releases/tag/v0.1.0
