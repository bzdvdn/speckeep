package project

import (
	"fmt"
	"strings"

	"speckeep/src/internal/skills"
)

func renderManagedAgentsBlockForRoot(root, snippet string) (string, error) {
	withSkills, err := appendSkillsSection(root, snippet)
	if err != nil {
		return "", err
	}
	return renderManagedAgentsBlock(withSkills), nil
}

func appendSkillsSection(root, snippet string) (string, error) {
	manifest, err := skills.Load(root)
	if err != nil {
		return "", err
	}

	base := strings.TrimSpace(snippet)
	lines := []string{base, ""}
	if isRussianSnippet(base) {
		lines = append(lines, "Skills (из `.speckeep/skills/manifest.yaml`):")
	} else {
		lines = append(lines, "Skills (from `.speckeep/skills/manifest.yaml`):")
	}
	if len(manifest.Skills) == 0 {
		if isRussianSnippet(base) {
			lines = append(lines, "- не настроены")
		} else {
			lines = append(lines, "- none configured")
		}
		return strings.Join(lines, "\n"), nil
	}

	for _, entry := range manifest.Skills {
		state := "enabled"
		if !entry.Enabled {
			state = "disabled"
		}
		if isRussianSnippet(base) {
			if entry.Source == "git" && entry.Ref != "" {
				lines = append(lines, fmt.Sprintf("- `%s` (%s, git: `%s@%s`)", entry.ID, state, entry.Location, entry.Ref))
			} else {
				lines = append(lines, fmt.Sprintf("- `%s` (%s, local: `%s`)", entry.ID, state, entry.Location))
			}
		} else {
			if entry.Source == "git" && entry.Ref != "" {
				lines = append(lines, fmt.Sprintf("- `%s` (%s, git: `%s@%s`)", entry.ID, state, entry.Location, entry.Ref))
			} else {
				lines = append(lines, fmt.Sprintf("- `%s` (%s, local: `%s`)", entry.ID, state, entry.Location))
			}
		}
		if strings.TrimSpace(entry.Path) != "" {
			lines = append(lines, fmt.Sprintf("  - path: `%s`", entry.Path))
		}
		if strings.TrimSpace(entry.CheckoutDir) != "" {
			lines = append(lines, fmt.Sprintf("  - checkout: `%s`", entry.CheckoutDir))
		}
	}

	return strings.Join(lines, "\n"), nil
}

func isRussianSnippet(snippet string) bool {
	return strings.Contains(snippet, "Основной контекст:") || strings.Contains(snippet, "Базовые правила:")
}
