package project

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	skillsGitignoreBlockStart = "# >>> speckeep: skills checkouts >>>"
	skillsGitignoreBlockEnd   = "# <<< speckeep: skills checkouts <<<"
)

func syncSkillsGitignore(root string, dryRun bool, result *RefreshResult) error {
	path := filepath.Join(root, ".gitignore")
	block := renderSkillsGitignoreBlock()

	current, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		recordRefreshAction(result, "created", rel(root, path))
		if dryRun {
			return nil
		}
		return os.WriteFile(path, []byte(block), 0o644)
	case err != nil:
		return err
	}

	updated := upsertManagedGitignoreBlock(string(current), block)
	if updated == string(current) {
		recordRefreshAction(result, "unchanged", rel(root, path))
		return nil
	}

	recordRefreshAction(result, "updated", rel(root, path))
	if dryRun {
		return nil
	}
	return os.WriteFile(path, []byte(updated), 0o644)
}

func renderSkillsGitignoreBlock() string {
	return skillsGitignoreBlockStart + "\n" +
		".speckeep/skills/checkouts/\n" +
		skillsGitignoreBlockEnd + "\n"
}

func upsertManagedGitignoreBlock(current, block string) string {
	start := strings.Index(current, skillsGitignoreBlockStart)
	end := strings.Index(current, skillsGitignoreBlockEnd)
	if start >= 0 && end > start {
		end += len(skillsGitignoreBlockEnd)
		prefix := strings.TrimRight(current[:start], "\n")
		suffix := strings.TrimLeft(current[end:], "\n")
		switch {
		case prefix == "" && suffix == "":
			return block
		case prefix == "":
			return block + "\n" + suffix
		case suffix == "":
			return prefix + "\n\n" + block
		default:
			return prefix + "\n\n" + block + "\n" + suffix
		}
	}

	trimmed := strings.TrimRight(current, "\n")
	if trimmed == "" {
		return block
	}
	return trimmed + "\n\n" + block
}
