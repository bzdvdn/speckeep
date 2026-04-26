package agents

import "path/filepath"

func SkillBasePath(target string) (string, bool) {
	switch target {
	case "codex":
		return filepath.ToSlash(filepath.Join(".codex", "skills")), true
	case "claude":
		return filepath.ToSlash(filepath.Join(".claude", "skills")), true
	case "kilocode":
		return filepath.ToSlash(filepath.Join(".kilocode", "skills")), true
	case "windsurf":
		return filepath.ToSlash(filepath.Join(".windsurf", "skills")), true
	case "trae":
		return filepath.ToSlash(filepath.Join(".trae", "skills")), true
	default:
		return "", false
	}
}
