package agents

import (
	"fmt"
	"path/filepath"
	"strings"

	"speckeep/src/internal/templates"
)

// ErrUnsupportedTarget is returned when a target has no skills directory.
var ErrUnsupportedTarget = fmt.Errorf("unsupported agent target")

// targetSkillDirs maps each supported agent target to the directory where its
// runtime expects skill files. `speckeep init --agents <target>` lays the same
// composite `sdd` skill pack into this directory.
var targetSkillDirs = map[string]string{
	"claude":    ".claude/skills",
	"opencode":  ".opencode/skills",
	"codex":     ".codex/skills",
	"kilocode":  ".kilocode/skills",
	"windsurf":  ".windsurf/skills",
	"trae":      ".trae/skills",
	"gemini":    ".gemini/skills",
	"amazonq":   ".amazonq/skills",
	"jules":     ".jules/skills",
	"cline":     ".clinerules/skills",
	"devin":     ".devin/skills",
	"goose":     ".goose/skills",
	"refact":    ".refact/skills",
	"codiumate": ".codiumate/skills",
	"qwen-code": ".qwen/skills",
	"roocode":   ".roo/skills",
	"aider":     ".aider/skills",
	"cursor":    ".cursor/skills",
	"copilot":   ".github/skills",
}

// deprecatedTargets lists targets speckeep no longer generates for.
// Consulted by NormalizeTargets (see its doc comment for why these are
// dropped silently rather than rejected) and by legacy-cleanup paths below.
//
// "continue": multiple secondary sources report Continue.dev was acquired
// by Cursor and discontinued around mid-2026 (GitHub repo made read-only).
// Confirmed by the maintainer as no longer worth supporting.
var deprecatedTargets = map[string]struct{}{
	"continue": {},
}

// DeprecatedTargetsRequested filters values down to whichever ones name a
// deprecated target (see deprecatedTargets), lower-cased and deduplicated —
// so callers can tell the user why an explicit `--agents continue` request
// quietly produced nothing, instead of leaving that silent.
func DeprecatedTargetsRequested(values []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			target := strings.ToLower(strings.TrimSpace(part))
			if _, ok := deprecatedTargets[target]; !ok {
				continue
			}
			if _, dup := seen[target]; dup {
				continue
			}
			seen[target] = struct{}{}
			out = append(out, target)
		}
	}
	return out
}

// LegacyContinueSkillPaths returns the `.continue/skills/...` paths
// speckeep used to generate before "continue" was dropped as a supported
// target (see deprecatedTargets). The base path is inlined rather than
// read from targetSkillDirs since that map no longer has the entry.
func LegacyContinueSkillPaths(commands []CommandDefinition) []string {
	const oldBase = ".continue/skills"
	paths := []string{filepath.ToSlash(filepath.Join(oldBase, "sdd", "SKILL.md"))}
	for _, cmd := range commands {
		paths = append(paths, filepath.ToSlash(filepath.Join(oldBase, "spk-"+cmd.Name, "SKILL.md")))
	}
	return paths
}

// flatCommandTarget describes a target's real (non-Skill) slash/mention
// command mechanism: a flat directory of one-file-per-command, plus the
// character used to invoke it by name.
type flatCommandTarget struct {
	dir          string
	invokePrefix string // "/" or "@"
}

// targetFlatCommandDirs lists targets whose Skills are auto-discovered but
// NOT slash-invocable — only model-invoked (automatic relevance matching)
// or explicit @mention/tool-call. For these, the tool's own real
// invocable-command mechanism lives in a separate flat-file directory (one
// file = one command), which we generate alongside the skill pack so users
// get a genuine spk-<phase> command there too. Verified against each
// tool's docs:
//   - Windsurf Skills load via @mention/auto-relevance, not "/"; real slash
//     commands are Workflows (.windsurf/workflows/*.md).
//   - OpenCode Skills are loaded by the agent calling a skill() tool, not
//     "/"; real slash commands are Commands (.opencode/commands/*.md).
//   - Cline's .clinerules/ is a context directory, not commands; real
//     slash commands are Workflows (.clinerules/workflows/*.md).
//   - Amazon Q Developer has no confirmed native Skills reader at all
//     (only an unrelated "Agent Toolkit for AWS" plugin uses that term);
//     its real invocable mechanism is Prompts (.amazonq/prompts/*.md),
//     invoked with "@name", not "/name".
var targetFlatCommandDirs = map[string]flatCommandTarget{
	"windsurf": {dir: ".windsurf/workflows", invokePrefix: "/"},
	"opencode": {dir: ".opencode/commands", invokePrefix: "/"},
	"cline":    {dir: ".clinerules/workflows", invokePrefix: "/"},
	"amazonq":  {dir: ".amazonq/prompts", invokePrefix: "@"},
}

// skillPackFiles builds the skill set for one target: a lightweight `sdd`
// overview skill (progressive-disclosure entry point for ambiguous/general
// requests) plus one independent, directly slash-invocable `spk-<phase>`
// skill per command. Each phase skill is its own top-level skill directory
// with its own SKILL.md, inlining the canonical prompt body directly — this
// mirrors how Claude Code (and equivalent skill loaders) only ever
// slash-invoke a top-level `<dir>/SKILL.md`, never a nested resource file.
// A single composite skill with nested `phases/*.md` files (the prior
// layout) made every phase reachable only through model-invocation of the
// root skill, never directly by name — see LegacySkillPhasePaths.
//
// For targets in targetFlatCommandDirs, Skills alone are not enough — they
// are not slash-invocable there at all — so a flat command file per phase
// is generated too, in that tool's real slash-command directory.
func skillPackFiles(target, language string, commands []CommandDefinition) ([]File, error) {
	lang := normalizeLanguage(language)
	base := targetSkillDirs[target]
	if base == "" {
		return nil, nil
	}
	files := []File{{
		Path:    filepath.ToSlash(filepath.Join(base, "sdd", "SKILL.md")),
		Content: renderRootSkill(commands, lang),
		Mode:    0o644,
	}}
	for _, command := range commands {
		content, err := renderPhaseSkill(command, lang)
		if err != nil {
			return nil, err
		}
		files = append(files, File{
			Path:    filepath.ToSlash(filepath.Join(base, "spk-"+command.Name, "SKILL.md")),
			Content: content,
			Mode:    0o644,
		})
	}
	if flat, ok := targetFlatCommandDirs[target]; ok {
		for _, command := range commands {
			content, err := renderFlatCommand(command, lang, flat.invokePrefix)
			if err != nil {
				return nil, err
			}
			files = append(files, File{
				Path:    filepath.ToSlash(filepath.Join(flat.dir, "spk-"+command.Name+".md")),
				Content: content,
				Mode:    0o644,
			})
		}
	}
	// Gemini CLI Skills are auto-discovered but not slash-invocable; its
	// real custom-command mechanism is TOML files, not markdown.
	if target == "gemini" {
		for _, command := range commands {
			content, err := renderGeminiCommand(command, lang)
			if err != nil {
				return nil, err
			}
			files = append(files, File{
				Path:    filepath.ToSlash(filepath.Join(".gemini/commands", "spk-"+command.Name+".toml")),
				Content: content,
				Mode:    0o644,
			})
		}
	}
	// Aider has no skill loader; point its CONVENTIONS file at the pack.
	if target == "aider" {
		files = append(files, File{
			Path:    ".aider/CONVENTIONS.md",
			Content: renderAiderPointer(),
			Mode:    0o644,
		})
	}
	return files, nil
}

func skillPackPaths(target string, commands []CommandDefinition) []string {
	base := targetSkillDirs[target]
	paths := []string{filepath.ToSlash(filepath.Join(base, "sdd", "SKILL.md"))}
	for _, command := range commands {
		paths = append(paths, filepath.ToSlash(filepath.Join(base, "spk-"+command.Name, "SKILL.md")))
	}
	if flat, ok := targetFlatCommandDirs[target]; ok {
		for _, command := range commands {
			paths = append(paths, filepath.ToSlash(filepath.Join(flat.dir, "spk-"+command.Name+".md")))
		}
	}
	if target == "gemini" {
		for _, command := range commands {
			paths = append(paths, filepath.ToSlash(filepath.Join(".gemini/commands", "spk-"+command.Name+".toml")))
		}
	}
	return paths
}

// LegacyAmazonQSkillPaths returns the old `.q/skills/...` paths speckeep
// used before correcting Amazon Q Developer's real dotfile root to
// `.amazonq/`. `.q/` was never confirmed as a path Amazon Q actually reads.
func LegacyAmazonQSkillPaths(commands []CommandDefinition) []string {
	const oldBase = ".q/skills"
	paths := []string{filepath.ToSlash(filepath.Join(oldBase, "sdd", "SKILL.md"))}
	for _, cmd := range commands {
		paths = append(paths, filepath.ToSlash(filepath.Join(oldBase, "spk-"+cmd.Name, "SKILL.md")))
	}
	return paths
}

// LegacySkillPhasePaths returns the pre-flat-skills `sdd/phases/<name>.md`
// paths (nested under one composite skill, unreachable by direct slash
// invocation) superseded by independent `spk-<name>` skills. refresh and
// cleanup remove them; doctor warns if they're still present.
func LegacySkillPhasePaths(commands []CommandDefinition) []string {
	var paths []string
	for _, base := range targetSkillDirs {
		for _, cmd := range commands {
			paths = append(paths, filepath.ToSlash(filepath.Join(base, "sdd", "phases", cmd.Name+".md")))
		}
	}
	return paths
}

func renderRootSkill(commands []CommandDefinition, lang string) string {
	var phaseLines []string
	for _, cmd := range commands {
		phaseLines = append(phaseLines, fmt.Sprintf("- `/spk-%s` — %s", cmd.Name, cmd.Description))
	}
	phases := strings.Join(phaseLines, "\n")
	scripts := "./.speckeep/scripts/check-ready.sh <phase> <slug>"

	if lang == "ru" {
		return fmt.Sprintf(`---
name: sdd
description: SpecKeep — spec-driven development. Use when the user asks to propose, spec, inspect, plan, decompose, implement, converge, or verify a feature.
---

# SpecKeep / SDD

Цепочка: constitution → spec → [inspect, опционально] → plan → tasks → implement → archive. Verify — опциональный аудит по требованию; propose — one-shot быстрая полоса; converge — быстрый цикл закрытия.

Каждая фаза — независимый skill, вызывается напрямую слэш-командой: %[3]s

## Как выполнять фазу

1. Если фаза уже известна — вызови её напрямую: /spk-<фаза> (каждый такой skill самодостаточен, полные инструкции внутри). Этот sdd-skill нужен только когда фаза не очевидна из запроса.
2. Сначала прочитай .speckeep/constitution.summary.md (fallback: CONSTITUTION.md).
3. Branch-first: работай с feature/<slug> (ветку создаёт/переключает только spec/propose).
4. Держи контекст узким: текущий slug + surfaces из Touches:.
5. Запускай readiness-скрипт: %[1]s, доверяй exit-коду.
6. Каждую фазу завершай end block (Slug / Status / Artifacts / Blockers / Готово к) и сохраняй точную финальную строку промпта.

## Гейты (не пропускать)

- speckeep check <slug> перед завершением фазы.
- speckeep converge <slug> (быстрый цикл) или speckeep guard . (CI) перед закрытием.
- Задача выполнена только со строкой Proof: под её [x] в tasks.md.

## Фазы (каждая — отдельный вызываемый skill)

%[2]s

## Ограничения

%[4]s
`, scripts, phases, "`/spk-<phase>`", antiPatternHint(lang))
	}

	return fmt.Sprintf(`---
name: sdd
description: SpecKeep — spec-driven development. Use when the user asks to propose, spec, inspect, plan, decompose, implement, converge, or verify a feature, or starts spec-driven development.
---

# SpecKeep / SDD

Workflow: constitution → spec → [inspect, optional] → plan → tasks → implement → archive. Verify is an optional on-demand audit; propose is the one-shot fast lane; converge is the fast closing loop.

Every phase is its own independent skill, invoked directly with %[3]s

## How to run a phase

1. If the phase is already known, invoke it directly: /spk-<phase> (each is self-contained with full instructions inline). Reach for this sdd overview skill only when the phase isn't obvious from the request.
2. Read .speckeep/constitution.summary.md first (fallback: CONSTITUTION.md).
3. Branch-first: work on feature/<slug> (only spec/propose may create/switch the branch).
4. Keep context narrow: current slug + Touches: surfaces only.
5. Run the readiness script: %[1]s and trust its exit code.
6. End every phase with the end block (Slug / Status / Artifacts / Blockers / Ready for) and preserve the prompt's exact final line.

## Gates (never skip)

- speckeep check <slug> before finishing a phase.
- speckeep converge <slug> (fast loop) or speckeep guard . (CI) before closing.
- A task is done only with a Proof: line under its [x] in tasks.md.

## Phases (each a separately invocable skill)

%[2]s

## Constraints

%[4]s
`, scripts, phases, "`/spk-<phase>`", antiPatternHint(lang))
}

// renderPhaseSkill inlines the canonical prompt body directly into the phase
// skill file, so a skill-driven agent gets the full instructions from a
// single read instead of following a pointer to .speckeep/templates/prompts/.
// The prompt text stays the single authored source (embedded at build time);
// this only changes where its content is delivered from at runtime.
func renderPhaseSkill(command CommandDefinition, lang string) (string, error) {
	promptPath := ".speckeep/templates/prompts/" + command.Name + ".md"
	checkReady := "./.speckeep/scripts/check-ready.sh " + command.Name + " [<slug>]"

	body, err := templates.PromptContent(lang, command.Name)
	if err != nil {
		return "", fmt.Errorf("render phase skill %q: %w", command.Name, err)
	}
	body = strings.TrimSpace(dropLeadingHeading(body))

	if lang == "ru" {
		return fmt.Sprintf(`---
name: spk-%s
description: SpecKeep-фаза «%s» — %s.
---

# /spk-%s

%s

---

Напоминания:

- readiness: %s (запусти, доверяй exit-коду).
- Создавай/правь только артефакты, которые называет промпт выше; контекст — текущий slug и surfaces из Touches:.
- Не расширяй scope, не перепланируй, не коммить без явной просьбы.
- Заверши фазу end block и сохрани точную финальную строку промпта.
- Гейт: speckeep check <slug> → исправь находки или сообщи blocker.
- Канонический источник (синхронизируется автоматически): %s

%s
`, command.Name, command.Name, command.Description, command.Name, body, checkReady, promptPath, proofHint(lang)), nil
	}

	return fmt.Sprintf(`---
name: spk-%s
description: SpecKeep phase "%s" — %s.
---

# /spk-%s

%s

---

Reminders:

- readiness: %s (run it, trust the exit code).
- Write/patch only the artifacts named above; keep context to the current slug and Touches: surfaces.
- Do not expand scope, re-plan, or commit without being asked.
- End with the end block and preserve the prompt's exact final line.
- Gate: speckeep check <slug> → fix findings or report a blocker.
- Canonical source (kept in sync automatically): %s

%s
`, command.Name, command.Name, command.Description, command.Name, body, checkReady, promptPath, proofHint(lang)), nil
}

// renderFlatCommand renders a plain, frontmatter-free markdown command file
// for targets whose Skills mechanism isn't invocable by name (see
// targetFlatCommandDirs) — the tool treats the file itself as the command
// body, keyed by filename. invokePrefix is "/" (Windsurf Workflows, OpenCode
// Commands, Cline Workflows) or "@" (Amazon Q Prompts).
func renderFlatCommand(command CommandDefinition, lang, invokePrefix string) (string, error) {
	promptPath := ".speckeep/templates/prompts/" + command.Name + ".md"
	checkReady := "./.speckeep/scripts/check-ready.sh " + command.Name + " [<slug>]"
	invoke := invokePrefix + "spk-" + command.Name

	body, err := templates.PromptContent(lang, command.Name)
	if err != nil {
		return "", fmt.Errorf("render flat command %q: %w", command.Name, err)
	}
	body = strings.TrimSpace(dropLeadingHeading(body))

	if lang == "ru" {
		return fmt.Sprintf(`# %s

%s

%s

---

Напоминания:

- readiness: %s (запусти, доверяй exit-коду).
- Создавай/правь только артефакты, которые называет промпт выше; контекст — текущий slug и surfaces из Touches:.
- Не расширяй scope, не перепланируй, не коммить без явной просьбы.
- Заверши фазу end block и сохрани точную финальную строку промпта.
- Гейт: speckeep check <slug> → исправь находки или сообщи blocker.
- Канонический источник (синхронизируется автоматически): %s
`, invoke, command.Description, body, checkReady, promptPath), nil
	}

	return fmt.Sprintf(`# %s

%s

%s

---

Reminders:

- readiness: %s (run it, trust the exit code).
- Write/patch only the artifacts named above; keep context to the current slug and Touches: surfaces.
- Do not expand scope, re-plan, or commit without being asked.
- End with the end block and preserve the prompt's exact final line.
- Gate: speckeep check <slug> → fix findings or report a blocker.
- Canonical source (kept in sync automatically): %s
`, invoke, command.Description, body, checkReady, promptPath), nil
}

// renderGeminiCommand renders a Gemini CLI custom-command TOML file
// (.gemini/commands/<name>.toml) — Gemini's real invocable-command format,
// distinct from its (non-slash-invocable) Skills mechanism.
func renderGeminiCommand(command CommandDefinition, lang string) (string, error) {
	body, err := templates.PromptContent(lang, command.Name)
	if err != nil {
		return "", fmt.Errorf("render gemini command %q: %w", command.Name, err)
	}
	body = strings.TrimSpace(dropLeadingHeading(body))
	checkReady := "./.speckeep/scripts/check-ready.sh " + command.Name + " [<slug>]"
	promptPath := ".speckeep/templates/prompts/" + command.Name + ".md"

	reminders := "Reminders: run " + checkReady + " first; write/patch only the named artifacts; " +
		"do not expand scope or commit without being asked; end with the end block; " +
		"gate with speckeep check <slug>; canonical source: " + promptPath
	if lang == "ru" {
		reminders = "Напоминания: сначала запусти " + checkReady + "; создавай/правь только названные артефакты; " +
			"не расширяй scope и не коммить без просьбы; заверши end block'ом; " +
			"гейт speckeep check <slug>; канонический источник: " + promptPath
	}

	description := strings.ReplaceAll(command.Description, `"`, `'`)
	promptBody := strings.ReplaceAll(body, `"""`, `'''`) + "\n\n---\n\n" + reminders

	return fmt.Sprintf("description = \"%s\"\nprompt = \"\"\"\n%s\n\"\"\"\n", description, promptBody), nil
}

// dropLeadingHeading strips the first line of content when it is a markdown
// heading, so the inlined prompt body doesn't duplicate the skill's own H1.
func dropLeadingHeading(content string) string {
	head, rest, found := strings.Cut(content, "\n")
	if !found || !strings.HasPrefix(strings.TrimSpace(head), "#") {
		return content
	}
	return rest
}

func renderAiderPointer() string {
	return `# SpecKeep Conventions

Load the SpecKeep skills:
- overview: .aider/skills/sdd/SKILL.md
- one self-contained skill per phase: .aider/skills/spk-<phase>/SKILL.md
  (e.g. .aider/skills/spk-spec/SKILL.md, .aider/skills/spk-plan/SKILL.md, ...)

Follow the canonical prompts in .speckeep/templates/prompts/. Read
.speckeep/constitution.summary.md first. Run readiness scripts and never mark a
task done without a Proof: line in tasks.md.
`
}
