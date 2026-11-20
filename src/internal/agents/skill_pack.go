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
//   - Cline's .clinerules/ is a context directory, not commands; real
//     slash commands are Workflows (.clinerules/workflows/*.md).
//   - Amazon Q Developer has no confirmed native Skills reader at all
//     (only an unrelated "Agent Toolkit for AWS" plugin uses that term);
//     its real invocable mechanism is Prompts (.amazonq/prompts/*.md),
//     invoked with "@name", not "/name".
//
// OpenCode is deliberately NOT listed: its Skills are likewise only
// model-invoked (loaded through the agent's skill() tool, never "/"), but
// the duplicate `.opencode/commands/spk-<phase>.md` set is redundant as a
// second entry point, so speckeep ships OpenCode skills-only. Phase skills
// stay reachable through the `/skills` picker; stale command files are
// removed via LegacyOpenCodeCommandPaths.
var targetFlatCommandDirs = map[string]flatCommandTarget{
	"windsurf": {dir: ".windsurf/workflows", invokePrefix: "/"},
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
func skillPackFiles(target, language, shell string, commands []CommandDefinition) ([]File, error) {
	lang := normalizeLanguage(language)
	base := targetSkillDirs[target]
	if base == "" {
		return nil, nil
	}
	files := []File{{
		Path:    filepath.ToSlash(filepath.Join(base, "sdd", "SKILL.md")),
		Content: renderRootSkill(commands, lang, shell),
		Mode:    0o644,
	}}
	for _, command := range commands {
		content, err := renderPhaseSkill(command, lang, shell)
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
			content, err := renderFlatCommand(command, lang, flat.invokePrefix, shell)
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
			content, err := renderGeminiCommand(command, lang, shell)
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

// LegacyOpenCodeCommandPaths returns the flat `.opencode/commands/spk-<phase>.md`
// slash-command files speckeep used to generate for OpenCode. OpenCode ships
// skills-only now (see targetFlatCommandDirs); refresh/cleanup remove these
// leftovers and doctor warns when they're still present.
func LegacyOpenCodeCommandPaths(commands []CommandDefinition) []string {
	var paths []string
	for _, cmd := range commands {
		paths = append(paths, filepath.ToSlash(filepath.Join(".opencode/commands", "spk-"+cmd.Name+".md")))
	}
	return paths
}

func renderRootSkill(commands []CommandDefinition, lang, shell string) string {
	var phaseNames, auxNames []string
	for _, cmd := range commands {
		if hasReadyCheck(cmd.Name) {
			phaseNames = append(phaseNames, cmd.Name)
			continue
		}
		auxNames = append(auxNames, cmd.Name)
	}
	phases := strings.Join(phaseNames, ", ")
	aux := strings.Join(auxNames, ", ")
	readiness := checkReadyScript(shell) + " <phase> <slug>"

	if lang == "ru" {
		return fmt.Sprintf(`---
name: sdd
description: SpecKeep — SDD-роутер. Используй только когда фаза неочевидна; каждый /spk-<phase> самодостаточен.
---

# SpecKeep / SDD

Цепочка: constitution → spec → [inspect, опц.] → [plan, опц.] → tasks → implement → archive. Новая фича/идея → сразу /spk-propose (роутер не грузи); plan — opt-in для дизайна.

Маршрут: новая фича → propose; неясно → spec; сырая/рискованная идея → grill (опц.); tech unknown → research (опц.); баг с неизвестной причиной → diagnose (опц.); дизайн → plan; есть tasks → implement; close → archive (PR) или finish (local); застрял → check/recap.

Если фаза известна — вызывай /spk-<phase> напрямую. Гейты: speckeep check <slug> по фазам; speckeep converge <slug> или speckeep guard . перед закрытием. Readiness: %[3]s (только гейтируемые фазы). End block/close — в AGENTS.md. Справочник: .speckeep/templates/commands.md.

Фазы (префикс /spk-): %[1]s

Вспомогательные (без гейта): %[2]s
`, phases, aux, readiness)
	}

	return fmt.Sprintf(`---
name: sdd
description: SpecKeep — SDD router. Use only when the phase is unclear; each /spk-<phase> is self-contained.
---

# SpecKeep / SDD

Chain: constitution → spec → [inspect, optional] → [plan, optional] → tasks → implement → archive. New feature idea → /spk-propose directly (skip this router); plan is opt-in for design.

Route: new feature → propose; unclear → spec; raw/high-risk idea → grill (optional); tech unknown → research (optional); bug with unknown cause → diagnose (optional); design → plan; tasks exist → implement; close → archive (PR) or finish (local); stuck → check/recap.

If the phase is known, invoke /spk-<phase> directly. Gates: speckeep check <slug> per phase; speckeep converge <slug> or speckeep guard . before closing. Readiness: %[3]s (gated phases only). End block/close are in AGENTS.md. Reference: .speckeep/templates/commands.md.

Phases (prefix /spk-): %[1]s

Auxiliary (no gate): %[2]s
`, phases, aux, readiness)
}

// readyCheckPhases lists the phases that have a real readiness gate behind
// `check-ready.sh <phase>` (mirrors the `__internal check-<phase>-ready`
// subcommands). Auxiliary commands — handoff, challenge, scope, glossary,
// recap, hotfix, repo-map, rollback — have no readiness check, so their
// generated artifacts must not advertise one: invoking it would fail with
// "unknown phase"/"unknown flag: --root".
var readyCheckPhases = map[string]struct{}{
	"constitution": {},
	"spec":         {},
	"propose":      {},
	"inspect":      {},
	"plan":         {},
	"tasks":        {},
	"implement":    {},
	"verify":       {},
	"converge":     {},
}

func hasReadyCheck(phase string) bool {
	_, ok := readyCheckPhases[phase]
	return ok
}

// checkReadyScript returns the readiness-script path for the configured shell,
// so generated reminders point at check-ready.sh or check-ready.ps1 as
// appropriate.
func checkReadyScript(shell string) string {
	ext := ".sh"
	if normalizeShell(shell) == "powershell" {
		ext = ".ps1"
	}
	return "./.speckeep/scripts/check-ready" + ext
}

// readinessLine returns the readiness reminder bullet (with trailing newline)
// for a phase, or "" when the phase has no readiness check.
func readinessLine(phase, lang, shell string) string {
	if !hasReadyCheck(phase) {
		return ""
	}
	checkReady := checkReadyScript(shell) + " " + phase + " [<slug>]"
	if lang == "ru" {
		return "- readiness: " + checkReady + " (запусти, доверяй exit-коду).\n"
	}
	return "- readiness: " + checkReady + " (run it, trust the exit code).\n"
}

// renderPhaseSkill inlines the canonical prompt body directly into the phase
// skill file, so a skill-driven agent gets the full instructions from a
// single read instead of following a pointer to .speckeep/templates/prompts/.
// The prompt text stays the single authored source (embedded at build time);
// this only changes where its content is delivered from at runtime.
func renderPhaseSkill(command CommandDefinition, lang, shell string) (string, error) {
	promptPath := ".speckeep/templates/prompts/" + command.Name + ".md"
	checkReady := readinessLine(command.Name, lang, shell)

	body, err := templates.PromptContent(lang, command.Name, shell)
	if err != nil {
		return "", fmt.Errorf("render phase skill %q: %w", command.Name, err)
	}
	body = strings.TrimSpace(dropLeadingHeading(body))

	// Auxiliary commands have no readiness gate and often define their own
	// output shape (recap/scope/resume explicitly forbid the end block), so
	// the generic gate/end-block reminder would contradict them.
	followEN := "Follow AGENTS.md; gate: speckeep check <slug> → fix findings or report a blocker. End with the canonical end block and keep the prompt's exact final line."
	followRU := "Следуй AGENTS.md; гейт: speckeep check <slug> → исправь находки или сообщи blocker. Заверши каноническим end block'ом и сохрани точную финальную строку промпта."
	if !hasReadyCheck(command.Name) {
		followEN = "Follow AGENTS.md. No readiness gate. Keep the prompt's exact final line and output format."
		followRU = "Следуй AGENTS.md. Гейта готовности нет. Сохрани точную финальную строку и формат вывода промпта."
	}

	if lang == "ru" {
		return fmt.Sprintf(`---
name: spk-%s
description: SpecKeep-фаза «%s» — %s.
---

# /spk-%s

%s

---

%s Источник: %s.

%s`, command.Name, command.Name, command.Description, command.Name, body, followRU, promptPath, checkReady), nil
	}

	return fmt.Sprintf(`---
name: spk-%s
description: SpecKeep phase "%s" — %s.
---

# /spk-%s

%s

---

%s Source: %s.

%s`, command.Name, command.Name, command.Description, command.Name, body, followEN, promptPath, checkReady), nil
}

// renderFlatCommand renders a plain, frontmatter-free markdown command file
// for targets whose Skills mechanism isn't invocable by name (see
// targetFlatCommandDirs) — the tool treats the file itself as the command
// body, keyed by filename. invokePrefix is "/" (Windsurf Workflows, Cline
// Workflows) or "@" (Amazon Q Prompts).
func renderFlatCommand(command CommandDefinition, lang, invokePrefix, shell string) (string, error) {
	promptPath := ".speckeep/templates/prompts/" + command.Name + ".md"
	checkReady := readinessLine(command.Name, lang, shell)
	invoke := invokePrefix + "spk-" + command.Name

	body, err := templates.PromptContent(lang, command.Name, shell)
	if err != nil {
		return "", fmt.Errorf("render flat command %q: %w", command.Name, err)
	}
	body = strings.TrimSpace(dropLeadingHeading(body))

	followEN := "Follow AGENTS.md; gate: speckeep check <slug>. End with the canonical end block."
	followRU := "Следуй AGENTS.md; гейт: speckeep check <slug>. Заверши каноническим end block'ом."
	if !hasReadyCheck(command.Name) {
		followEN = "Follow AGENTS.md. No readiness gate; keep the prompt's output format."
		followRU = "Следуй AGENTS.md. Гейта готовности нет; сохрани формат вывода промпта."
	}

	if lang == "ru" {
		return fmt.Sprintf(`# %s

%s

%s

%s%s Источник: %s.
`, invoke, command.Description, body, checkReady, followRU, promptPath), nil
	}

	return fmt.Sprintf(`# %s

%s

%s

%s%s Source: %s.
`, invoke, command.Description, body, checkReady, followEN, promptPath), nil
}

// renderGeminiCommand renders a Gemini CLI custom-command TOML file
// (.gemini/commands/<name>.toml) — Gemini's real invocable-command format,
// distinct from its (non-slash-invocable) Skills mechanism.
func renderGeminiCommand(command CommandDefinition, lang, shell string) (string, error) {
	body, err := templates.PromptContent(lang, command.Name, shell)
	if err != nil {
		return "", fmt.Errorf("render gemini command %q: %w", command.Name, err)
	}
	body = strings.TrimSpace(dropLeadingHeading(body))
	promptPath := ".speckeep/templates/prompts/" + command.Name + ".md"
	runFirst := ""
	if hasReadyCheck(command.Name) {
		runFirst = "run " + checkReadyScript(shell) + " " + command.Name + " [<slug>] first; "
	}

	reminders := "Reminders: " + runFirst + "write/patch only the named artifacts; " +
		"do not expand scope or commit without being asked; end with the end block; " +
		"gate with speckeep check <slug>; canonical source: " + promptPath
	if !hasReadyCheck(command.Name) {
		reminders = "Reminders: write/patch only the named artifacts; " +
			"do not expand scope or commit without being asked; no readiness gate; " +
			"keep the prompt's output format; canonical source: " + promptPath
	}
	if lang == "ru" {
		runFirst = ""
		if hasReadyCheck(command.Name) {
			runFirst = "сначала запусти " + checkReadyScript(shell) + " " + command.Name + " [<slug>]; "
		}
		reminders = "Напоминания: " + runFirst + "создавай/правь только названные артефакты; " +
			"не расширяй scope и не коммить без просьбы; заверши end block'ом; " +
			"гейт speckeep check <slug>; канонический источник: " + promptPath
		if !hasReadyCheck(command.Name) {
			reminders = "Напоминания: создавай/правь только названные артефакты; " +
				"не расширяй scope и не коммить без просьбы; гейта готовности нет; " +
				"сохрани формат вывода промпта; канонический источник: " + promptPath
		}
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
