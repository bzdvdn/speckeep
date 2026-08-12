package trace

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"speckeep/src/internal/config"
	"speckeep/src/internal/featurepaths"
)

// Entry is one Proof record of a completed task, parsed from tasks.md.
//
//   - [x] T1.1 Implement export (AC-001). Touches: src/internal/export.go
//     Proof: code src/internal/export.go RunExport
type Entry struct {
	Slug   string
	TaskID string
	ACID   string
	Kind   string // code|test|docs|chore
	File   string
	Anchor string
}

// LegacyMarker is a leftover @sk-task/@sk-test/@ds-* annotation in source code.
type LegacyMarker struct {
	File string
	Line int
}

// EntryProblem describes a broken Proof entry.
type EntryProblem struct {
	Kind    string // "file-missing" | "anchor-missing"
	Message string
}

var (
	taskLinePattern      = regexp.MustCompile(`^\s*- \[x\]\s+(T[0-9]+\.[0-9]+)\b`)
	checkboxTokenPattern = regexp.MustCompile(`^\s*- \[([ x])\]`)
	taskIDPattern        = regexp.MustCompile(`T[0-9]+\.[0-9]+`)
	proofLinePattern     = regexp.MustCompile(`^\s*Proof:\s*(\S+)\s+(\S+)(?:\s+(\S+))?\s*$`)
	acceptanceIDPattern  = regexp.MustCompile(`AC-[0-9][0-9][0-9]`)
	legacyMarkerPattern  = regexp.MustCompile(`@(?:ds|sk)-(?:task|test)\b`)
)

// ParseTasks reads the tasks.md Proof entries for one feature slug.
func ParseTasks(ctx context.Context, root, slug string) ([]Entry, error) {
	cfg, err := config.Load(context.Background(), root)
	if err != nil {
		return nil, err
	}
	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return nil, err
	}
	tasksPath, _ := featurepaths.ResolveTasks(specsDir, slug)
	return parseFile(root, slug, tasksPath)
}

// ParseAll aggregates Proof entries across all active features.
func ParseAll(ctx context.Context, root string) ([]Entry, error) {
	cfg, err := config.Load(context.Background(), root)
	if err != nil {
		return nil, err
	}
	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return nil, err
	}

	slugs, err := activeSlugs(root, specsDir)
	if err != nil {
		return nil, err
	}

	var out []Entry
	for _, slug := range slugs {
		tasksPath, _ := featurepaths.ResolveTasks(specsDir, slug)
		if !fileExists(tasksPath) {
			continue
		}
		fileEntries, err := parseFile(root, slug, tasksPath)
		if err != nil {
			return nil, err
		}
		out = append(out, fileEntries...)
	}
	return out, nil
}

func parseFile(root, slug, tasksPath string) ([]Entry, error) {
	content, err := os.ReadFile(tasksPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var (
		entries   []Entry
		curTask   string
		curACID   string
		curClosed bool
	)

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if matches := checkboxTokenPattern.FindStringSubmatch(line); len(matches) > 0 {
			taskID := firstMatch(taskIDPattern, line)
			curTask = ""
			curClosed = false
			if taskID != "" {
				curTask = taskID
				curClosed = matches[1] == "x"
				if firstACID := firstMatch(acceptanceIDPattern, line); len(firstACID) > 0 {
					curACID = firstACID
				} else {
					curACID = ""
				}
			}
			continue
		}

		proof := proofLinePattern.FindStringSubmatch(line)
		if proof == nil || curTask == "" || !curClosed {
			continue
		}

		entry := Entry{
			Slug:   slug,
			TaskID: curTask,
			ACID:   curACID,
			Kind:   proof[1],
			File:   proof[2],
			Anchor: proof[3],
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func firstMatch(re *regexp.Regexp, line string) string {
	match := re.FindString(line)
	return strings.TrimSpace(match)
}

// ResolveFile returns the absolute path of a Proof entry file relative to root.
func ResolveFile(root string, e Entry) string {
	value := strings.Trim(strings.TrimSpace(e.File), "`")
	if value == "" {
		return ""
	}
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(root, filepath.FromSlash(value))
}

// FileExists reports whether the Proof entry file exists.
func FileExists(root string, e Entry) bool {
	path := ResolveFile(root, e)
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// CheckEntry verifies that a Proof entry points to an existing file and,
// when an anchor is present, that the anchor symbol can be found in it.
func CheckEntry(root string, e Entry) []EntryProblem {
	var problems []EntryProblem
	if !FileExists(root, e) {
		problems = append(problems, EntryProblem{
			Kind:    "file-missing",
			Message: nameEntry(e) + " references missing file: " + e.File,
		})
		return problems
	}
	if e.Anchor == "" {
		return problems
	}
	path := ResolveFile(root, e)
	content, err := os.ReadFile(path)
	if err != nil {
		return problems
	}
	pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(e.Anchor) + `\b`)
	if !pattern.Match(content) {
		problems = append(problems, EntryProblem{
			Kind:    "anchor-missing",
			Message: nameEntry(e) + " anchor " + e.Anchor + " not found in " + e.File,
		})
	}
	return problems
}

func nameEntry(e Entry) string {
	if e.Slug == "" {
		return e.TaskID
	}
	return e.Slug + "#" + e.TaskID
}

// FindLegacyMarkers scans the codebase for deprecated @ds-*/@sk-task/@sk-test markers.
func FindLegacyMarkers(ctx context.Context, root string) ([]LegacyMarker, error) {
	var markers []LegacyMarker
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if shouldSkip(path, info) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if legacyMarkerPattern.MatchString(line) {
				markers = append(markers, LegacyMarker{File: path, Line: i + 1})
			}
		}
		return nil
	})
	return markers, err
}

// CompleteTasks returns the set of closed ( [x] ) task IDs in a file.
func CompleteTasks(content string) []string {
	var ids []string
	for _, line := range strings.Split(content, "\n") {
		matches := taskLinePattern.FindStringSubmatch(line)
		if len(matches) == 2 {
			ids = append(ids, matches[1])
		}
	}
	return ids
}

// CompletedTaskLines returns closed task IDs together with their Proof presence.
func CompletedTaskLines(content string) (ids []string, withProof map[string]bool) {
	withProof = map[string]bool{}
	curTask := ""
	curClosed := false
	for _, line := range strings.Split(content, "\n") {
		if matches := checkboxTokenPattern.FindStringSubmatch(line); len(matches) > 0 {
			taskID := firstMatch(taskIDPattern, line)
			curTask = ""
			curClosed = false
			if taskID != "" {
				curTask = taskID
				curClosed = matches[1] == "x"
				if curClosed {
					ids = append(ids, taskID)
					if _, ok := withProof[taskID]; !ok {
						withProof[taskID] = false
					}
				}
			}
			continue
		}
		if proofLinePattern.MatchString(line) && curTask != "" && curClosed {
			withProof[curTask] = true
		}
	}
	ids = uniqueStrings(ids)
	return ids, withProof
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func activeSlugs(root, specsDir string) ([]string, error) {
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var slugs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		tasksPath, _ := featurepaths.ResolveTasks(specsDir, entry.Name())
		if fileExists(tasksPath) {
			slugs = append(slugs, entry.Name())
		}
	}
	sort.Strings(slugs)
	return slugs, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func shouldSkip(path string, info os.FileInfo) bool {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") && base != "." {
		return true
	}
	skipDirs := []string{"node_modules", "vendor", "dist", "bin", "obj", ".git", ".speckeep"}
	for _, dir := range skipDirs {
		if base == dir {
			return true
		}
	}
	// Markdown files are guidance/docs, not source — they may legitimately describe
	// the deprecated markers in instructional text (e.g. "do NOT place @sk-task").
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown", ".mdc", ".mdown", ".mkd":
		return true
	}
	_ = info
	return false
}
