package project

import (
	"fmt"
	"path/filepath"
	"strings"
)

// resolveSpecsRoot expands the --specs-root convenience option into the active
// specs directory and the archive directory. Explicit specsDir/archiveDir
// values override the corresponding derived value; an empty specsRoot keeps
// the existing exact-path behavior.
func resolveSpecsRoot(specsRoot, specsDir, archiveDir string) (string, string, error) {
	root := strings.TrimSpace(specsRoot)
	if root == "" {
		return specsDir, archiveDir, nil
	}
	if filepath.IsAbs(root) {
		return "", "", fmt.Errorf("specs-root must be a relative path, got %q", root)
	}

	root = filepath.ToSlash(filepath.Clean(root))
	active := filepath.ToSlash(filepath.Join(root, "active"))
	archived := filepath.ToSlash(filepath.Join(root, "archived"))

	if strings.TrimSpace(specsDir) != "" {
		active = specsDir
	}
	if strings.TrimSpace(archiveDir) != "" {
		archived = archiveDir
	}
	return active, archived, nil
}
