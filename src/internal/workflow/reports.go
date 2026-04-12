package workflow

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	ReportTypeInspect = "inspect"
	ReportTypeVerify  = "verify"
	StatusPass        = "pass"
	StatusConcerns    = "concerns"
	StatusBlocked     = "blocked"
)

type Report struct {
	Type         string
	Slug         string
	Status       string
	DocsLanguage string
	GeneratedAt  string
}

type reportFrontmatter struct {
	ReportType   string `yaml:"report_type"`
	Slug         string `yaml:"slug"`
	Status       string `yaml:"status"`
	DocsLanguage string `yaml:"docs_language"`
	GeneratedAt  string `yaml:"generated_at"`
}

func ParseReport(path string) (Report, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Report{}, fmt.Errorf("read report %s: %w", path, err)
	}

	report, err := ParseReportContent(string(content))
	if err != nil {
		return Report{}, fmt.Errorf("parse report %s: %w", path, err)
	}
	return report, nil
}

func ParseReportContent(content string) (Report, error) {
	report := Report{}
	metadata, body, err := splitFrontmatter(content)
	if err != nil {
		return Report{}, err
	}

	if metadata != "" {
		var fm reportFrontmatter
		if err := yaml.Unmarshal([]byte(metadata), &fm); err != nil {
			return Report{}, fmt.Errorf("parse report metadata: %w", err)
		}
		report.Type = strings.TrimSpace(fm.ReportType)
		report.Slug = strings.TrimSpace(fm.Slug)
		report.Status = normalizeStatus(fm.Status)
		report.DocsLanguage = strings.TrimSpace(fm.DocsLanguage)
		report.GeneratedAt = strings.TrimSpace(fm.GeneratedAt)
	}

	if report.Type == "" {
		report.Type = inferReportType(body)
	}
	if report.Slug == "" {
		report.Slug = inferReportSlug(body)
	}
	if report.Status == "" {
		report.Status = inferLegacyStatus(body)
	}

	return report, nil
}

func ValidStatus(value string) bool {
	switch normalizeStatus(value) {
	case StatusPass, StatusConcerns, StatusBlocked:
		return true
	default:
		return false
	}
}

func ExtractUniqueMatches(content, pattern string) []string {
	re := regexp.MustCompile(pattern)
	raw := re.FindAllString(content, -1)
	seen := map[string]struct{}{}
	var out []string
	for _, item := range raw {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func ContainsAny(content string, values ...string) bool {
	for _, value := range values {
		if strings.Contains(content, value) {
			return true
		}
	}
	return false
}

func splitFrontmatter(content string) (metadata string, body string, err error) {
	if !strings.HasPrefix(content, "---\n") {
		return "", content, nil
	}

	rest := strings.TrimPrefix(content, "---\n")
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return "", "", fmt.Errorf("unterminated report metadata block")
	}

	metadata = rest[:end]
	body = rest[end+len("\n---\n"):]
	return metadata, body, nil
}

func inferReportType(content string) string {
	switch {
	case strings.Contains(content, "# Inspect Report:"):
		return ReportTypeInspect
	case strings.Contains(content, "# Verify Report:"):
		return ReportTypeVerify
	default:
		return ""
	}
}

func inferReportSlug(content string) string {
	re := regexp.MustCompile(`(?m)^# (Inspect|Verify) Report: (.+)$`)
	match := re.FindStringSubmatch(content)
	if len(match) == 3 {
		return strings.TrimSpace(match[2])
	}
	return ""
}

func inferLegacyStatus(content string) string {
	re := regexp.MustCompile(`(?m)^- status: (pass|concerns|blocked)$`)
	match := re.FindStringSubmatch(content)
	if len(match) == 2 {
		return normalizeStatus(match[1])
	}
	return ""
}

func normalizeStatus(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case StatusPass, StatusConcerns, StatusBlocked:
		return value
	default:
		return ""
	}
}
