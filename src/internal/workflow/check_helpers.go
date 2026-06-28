package workflow

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"speckeep/src/internal/featurepaths"
)

func resolveUserPath(root, value string) (string, string) {
	display := filepath.ToSlash(strings.TrimSpace(value))
	if filepath.IsAbs(value) {
		return display, value
	}
	return display, filepath.Join(root, filepath.FromSlash(value))
}

func joinDisplay(parts ...string) string {
	return filepath.ToSlash(filepath.Join(parts...))
}

func absFromRoot(root, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(root, filepath.FromSlash(rel))
}

func resolveSpecDisplayPath(root, specsDir, slug string) (string, string) {
	display := joinDisplay(specsDir, slug, "spec.md")
	abs := absFromRoot(root, display)
	if fileExists(abs) {
		return display, abs
	}
	legacyDisplay := joinDisplay(specsDir, slug+".md")
	legacyAbs := absFromRoot(root, legacyDisplay)
	if fileExists(legacyAbs) {
		return legacyDisplay, legacyAbs
	}
	return display, abs
}

func resolveHotfixDisplayPath(root, specsDir, slug string) (string, string) {
	display := joinDisplay(specsDir, slug, "hotfix.md")
	abs := absFromRoot(root, display)
	if fileExists(abs) {
		return display, abs
	}
	legacyDisplay := joinDisplay(specsDir, slug+".hotfix.md")
	legacyAbs := absFromRoot(root, legacyDisplay)
	if fileExists(legacyAbs) {
		return legacyDisplay, legacyAbs
	}
	return display, abs
}

func resolveInspectDisplayPath(root, specsDir, slug string) (string, string) {
	display := joinDisplay(specsDir, slug, "inspect.md")
	abs := absFromRoot(root, display)
	if fileExists(abs) {
		return display, abs
	}
	legacyDisplay := joinDisplay(specsDir, slug+".inspect.md")
	legacyAbs := absFromRoot(root, legacyDisplay)
	if fileExists(legacyAbs) {
		return legacyDisplay, legacyAbs
	}
	return display, abs
}

func resolvePlanDisplayPath(root, specsDir, slug string) (string, string) {
	return resolveFeatureArtifactPath(root, featurepaths.Plan(specsDir, slug), featurepaths.LegacyPlan(specsDir, slug))
}

func resolveTasksDisplayPath(root, specsDir, slug string) (string, string) {
	return resolveFeatureArtifactPath(root, featurepaths.Tasks(specsDir, slug), featurepaths.LegacyTasks(specsDir, slug))
}

func resolveDataModelDisplayPath(root, specsDir, slug string) (string, string) {
	return resolveFeatureArtifactPath(root, featurepaths.DataModel(specsDir, slug), featurepaths.LegacyDataModel(specsDir, slug))
}

func resolveContractsDisplayPath(root, specsDir, slug string) (string, string) {
	return resolveFeatureArtifactDirPath(root, featurepaths.ContractsDir(specsDir, slug), featurepaths.LegacyContractsDir(specsDir, slug))
}

func resolveFeatureArtifactPath(root, canonicalPath, legacyPath string) (string, string) {
	canonicalDisplay := filepath.ToSlash(canonicalPath)
	canonicalAbs := absFromRoot(root, canonicalDisplay)
	if fileExists(canonicalAbs) {
		return canonicalDisplay, canonicalAbs
	}
	legacyDisplay := filepath.ToSlash(legacyPath)
	legacyAbs := absFromRoot(root, legacyDisplay)
	if fileExists(legacyAbs) {
		return legacyDisplay, legacyAbs
	}
	return canonicalDisplay, canonicalAbs
}

func resolveFeatureArtifactDirPath(root, canonicalPath, legacyPath string) (string, string) {
	canonicalDisplay := filepath.ToSlash(canonicalPath)
	canonicalAbs := absFromRoot(root, canonicalDisplay)
	if isDir(canonicalAbs) {
		return canonicalDisplay, canonicalAbs
	}
	legacyDisplay := filepath.ToSlash(legacyPath)
	legacyAbs := absFromRoot(root, legacyDisplay)
	if isDir(legacyAbs) {
		return legacyDisplay, legacyAbs
	}
	return canonicalDisplay, canonicalAbs
}

func checkFile(result *CheckResult, displayPath, absolutePath string) {
	if fileExists(absolutePath) {
		result.AddStructuredOK("file_present", CategoryReadiness, displayPath, displayPath)
		return
	}
	result.AddStructuredError("file_missing", CategoryReadiness, displayPath, fmt.Sprintf("missing %s", displayPath))
}

func checkOptionalFile(result *CheckResult, displayPath, absolutePath, missingMsg string) {
	if fileExists(absolutePath) {
		result.AddStructuredOK("file_present", CategoryReadiness, displayPath, displayPath)
		return
	}
	result.AddStructuredWarn("optional_file_missing", CategoryReadiness, displayPath, missingMsg)
}

func checkPattern(result *CheckResult, content, pattern, label string) {
	if regexp.MustCompile(pattern).FindStringIndex(content) != nil {
		result.AddStructuredOK("pattern_present", CategoryStructure, "", label)
		return
	}
	result.AddStructuredError("pattern_missing", CategoryStructure, "", label)
}

func hasHeading(content, section string) bool {
	return strings.Contains(content, "\n## "+section+"\n") || strings.HasPrefix(content, "## "+section+"\n") || strings.HasSuffix(content, "\n## "+section)
}

func checkRequiredHeading(result *CheckResult, content, section string) {
	if hasHeading(content, section) {
		result.AddStructuredOK("required_section_present", CategoryStructure, "spec", section, section)
	} else {
		result.AddStructuredError("required_section_missing", CategoryStructure, "spec", fmt.Sprintf("missing required section: %s", section), section)
	}
}

func checkOptionalHeading(result *CheckResult, content, section string) {
	if hasHeading(content, section) {
		result.AddStructuredOK("optional_section_present", CategoryStructure, "spec", section, section)
	} else {
		result.AddStructuredWarn("optional_section_missing", CategoryStructure, "spec", fmt.Sprintf("missing section: %s", section), section)
	}
}

func markdownSection(content, section string) string {
	lines := strings.Split(content, "\n")
	var captured []string
	inSection := false
	target := "## " + section
	for _, line := range lines {
		if line == target {
			inSection = true
			continue
		}
		if inSection && strings.HasPrefix(line, "## ") {
			break
		}
		if inSection {
			captured = append(captured, line)
		}
	}
	return strings.Join(captured, "\n")
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
