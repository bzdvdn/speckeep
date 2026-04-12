package trace

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Finding struct {
	Type        string `json:"type"` // "task" or "test"
	File        string `json:"file"`
	Line        int    `json:"line"`
	TaskID      string `json:"task_id"`
	ACID        string `json:"ac_id,omitempty"`
	Description string `json:"description,omitempty"`
}

type TraceResult struct {
	Findings []Finding `json:"findings"`
}

// tracePattern matches:
//
//	// @sk-task T1.1: Description (AC-001)
//	// @sk-test T1.1: TestName (AC-001)
//
// Legacy annotations (@ds-task/@ds-test) are also accepted.
var tracePattern = regexp.MustCompile(`@(?:ds|sk)-(task|test)\s+([A-Z0-9.]+)(?::\s*([^(]*))?(?:\s*\((AC-[0-9]+)\))?`)

func Scan(root string) (TraceResult, error) {
	var result TraceResult

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}
		if shouldSkip(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		findings, err := scanFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}
		result.Findings = append(result.Findings, findings...)
		return nil
	})

	return result, err
}

func scanFile(path string) ([]Finding, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var findings []Finding
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		matches := tracePattern.FindStringSubmatch(line)
		if len(matches) > 2 {
			finding := Finding{
				Type:        matches[1],
				File:        path,
				Line:        lineNumber,
				TaskID:      matches[2],
				Description: strings.TrimSpace(matches[3]),
				ACID:        matches[4],
			}
			findings = append(findings, finding)
		}
	}

	return findings, scanner.Err()
}

func shouldSkip(path string) bool {
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
	return false
}

func FilterBySlug(findings []Finding, taskIDs map[string]struct{}) []Finding {
	var filtered []Finding
	for _, f := range findings {
		if _, ok := taskIDs[f.TaskID]; ok {
			filtered = append(filtered, f)
		}
	}
	return filtered
}
