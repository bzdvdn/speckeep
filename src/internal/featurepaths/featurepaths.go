package featurepaths

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Artifact struct {
	Name          string
	CanonicalPath string
	LegacyPath    string
}

func SpecDir(specsDir, slug string) string {
	return filepath.Join(specsDir, slug)
}

func Spec(specsDir, slug string) string {
	return filepath.Join(SpecDir(specsDir, slug), "spec.md")
}

func Inspect(specsDir, slug string) string {
	return filepath.Join(SpecDir(specsDir, slug), "inspect.md")
}

func Summary(specsDir, slug string) string {
	return filepath.Join(SpecDir(specsDir, slug), "summary.md")
}

func Hotfix(specsDir, slug string) string {
	return filepath.Join(SpecDir(specsDir, slug), "hotfix.md")
}

func SpecDigest(specsDir, slug string) string {
	return filepath.Join(SpecDir(specsDir, slug), "spec.digest.md")
}

func PlanDir(specsDir, slug string) string {
	return filepath.Join(specsDir, slug, "plan")
}

func Plan(specsDir, slug string) string {
	return filepath.Join(PlanDir(specsDir, slug), "plan.md")
}

func Tasks(specsDir, slug string) string {
	return filepath.Join(PlanDir(specsDir, slug), "tasks.md")
}

func DataModel(specsDir, slug string) string {
	return filepath.Join(PlanDir(specsDir, slug), "data-model.md")
}

func Research(specsDir, slug string) string {
	return filepath.Join(PlanDir(specsDir, slug), "research.md")
}

func Verify(specsDir, slug string) string {
	return filepath.Join(PlanDir(specsDir, slug), "verify.md")
}

func ContractsDir(specsDir, slug string) string {
	return filepath.Join(PlanDir(specsDir, slug), "contracts")
}

func LegacySpec(specsDir, slug string) string {
	return filepath.Join(specsDir, slug+".md")
}

func LegacyInspect(specsDir, slug string) string {
	return filepath.Join(specsDir, slug+".inspect.md")
}

func LegacySummary(specsDir, slug string) string {
	return filepath.Join(specsDir, slug+".summary.md")
}

func LegacyHotfix(specsDir, slug string) string {
	return filepath.Join(specsDir, slug+".hotfix.md")
}

func Artifacts(specsDir, slug string) []Artifact {
	return []Artifact{
		{Name: "spec", CanonicalPath: Spec(specsDir, slug), LegacyPath: LegacySpec(specsDir, slug)},
		{Name: "inspect report", CanonicalPath: Inspect(specsDir, slug), LegacyPath: LegacyInspect(specsDir, slug)},
		{Name: "summary", CanonicalPath: Summary(specsDir, slug), LegacyPath: LegacySummary(specsDir, slug)},
		{Name: "hotfix", CanonicalPath: Hotfix(specsDir, slug), LegacyPath: LegacyHotfix(specsDir, slug)},
	}
}

func ResolveSpec(specsDir, slug string) (string, bool) {
	return resolve(Spec(specsDir, slug), LegacySpec(specsDir, slug))
}

func ResolveInspect(specsDir, slug string) (string, bool) {
	return resolve(Inspect(specsDir, slug), LegacyInspect(specsDir, slug))
}

func ResolveSummary(specsDir, slug string) (string, bool) {
	return resolve(Summary(specsDir, slug), LegacySummary(specsDir, slug))
}

func ResolveHotfix(specsDir, slug string) (string, bool) {
	return resolve(Hotfix(specsDir, slug), LegacyHotfix(specsDir, slug))
}

func ListSpecSlugs(specsDir string) ([]string, error) {
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("read specs directory: %w", err)
	}

	slugSet := map[string]struct{}{}
	for _, entry := range entries {
		switch {
		case entry.IsDir():
			if fileExists(filepath.Join(specsDir, entry.Name(), "spec.md")) {
				slugSet[entry.Name()] = struct{}{}
			}
		default:
			slug, ok := slugFromLegacySpec(entry.Name())
			if ok {
				slugSet[slug] = struct{}{}
			}
		}
	}

	slugs := make([]string, 0, len(slugSet))
	for slug := range slugSet {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)
	return slugs, nil
}

func resolve(canonicalPath, legacyPath string) (string, bool) {
	switch {
	case fileExists(canonicalPath):
		return canonicalPath, false
	case fileExists(legacyPath):
		return legacyPath, true
	default:
		return canonicalPath, false
	}
}

func slugFromLegacySpec(name string) (string, bool) {
	if !strings.HasSuffix(name, ".md") {
		return "", false
	}
	switch {
	case strings.HasSuffix(name, ".inspect.md"):
		return "", false
	case strings.HasSuffix(name, ".summary.md"):
		return "", false
	case strings.HasSuffix(name, ".hotfix.md"):
		return "", false
	default:
		return strings.TrimSuffix(name, ".md"), true
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
