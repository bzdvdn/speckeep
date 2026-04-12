package workflow

import "testing"

func TestParseReportContentSupportsMetadataFrontmatter(t *testing.T) {
	content := `---
report_type: inspect
slug: demo
status: pass
docs_language: en
generated_at: 2026-03-30
---
# Inspect Report: demo

## Verdict

- status: pass
`

	report, err := ParseReportContent(content)
	if err != nil {
		t.Fatalf("ParseReportContent returned error: %v", err)
	}
	if report.Type != ReportTypeInspect {
		t.Fatalf("Type = %q, want %q", report.Type, ReportTypeInspect)
	}
	if report.Slug != "demo" {
		t.Fatalf("Slug = %q, want demo", report.Slug)
	}
	if report.Status != StatusPass {
		t.Fatalf("Status = %q, want %q", report.Status, StatusPass)
	}
	if report.DocsLanguage != "en" {
		t.Fatalf("DocsLanguage = %q, want en", report.DocsLanguage)
	}
	if report.GeneratedAt != "2026-03-30" {
		t.Fatalf("GeneratedAt = %q, want 2026-03-30", report.GeneratedAt)
	}
}

func TestParseReportContentSupportsLegacyStatusFallback(t *testing.T) {
	content := `# Verify Report: demo

## Verdict

- status: concerns
`

	report, err := ParseReportContent(content)
	if err != nil {
		t.Fatalf("ParseReportContent returned error: %v", err)
	}
	if report.Type != ReportTypeVerify {
		t.Fatalf("Type = %q, want %q", report.Type, ReportTypeVerify)
	}
	if report.Slug != "demo" {
		t.Fatalf("Slug = %q, want demo", report.Slug)
	}
	if report.Status != StatusConcerns {
		t.Fatalf("Status = %q, want %q", report.Status, StatusConcerns)
	}
}
