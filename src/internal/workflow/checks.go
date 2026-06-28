package workflow

import (
	"fmt"
	"regexp"
	"strings"
)

type CheckResult struct {
	Lines    []string
	Findings []CheckFinding
	Errors   int
	Warnings int
	Failed   bool
}

type FindingSeverity string
type FindingCategory string

const (
	SeverityOK      FindingSeverity = "ok"
	SeverityWarning FindingSeverity = "warning"
	SeverityError   FindingSeverity = "error"

	CategoryStructure    FindingCategory = "structure"
	CategoryTraceability FindingCategory = "traceability"
	CategoryAmbiguity    FindingCategory = "ambiguity"
	CategoryConsistency  FindingCategory = "consistency"
	CategoryReadiness    FindingCategory = "readiness"
)

type CheckFinding struct {
	Code     string          `json:"code"`
	Severity FindingSeverity `json:"severity"`
	Category FindingCategory `json:"category"`
	Artifact string          `json:"artifact,omitempty"`
	Path     string          `json:"path,omitempty"`
	Section  string          `json:"section,omitempty"`
	Message  string          `json:"message"`
	Refs     []string        `json:"refs,omitempty"`
}

type TaskStateSummary struct {
	Total         int
	Completed     int
	Open          int
	TaskIDs       int
	CoverageLines int
}

type docSections struct {
	Goal         string
	Context      string
	Requirements string
	Acceptance   string
	Questions    string
	Coverage     string
	Assumptions  string
}

var (
	taskIDPattern             = regexp.MustCompile(`T[0-9]+\.[0-9]+`)
	coverageLinePattern       = regexp.MustCompile(`AC-[0-9][0-9][0-9].*(?:->|→).*T[0-9]+\.[0-9]+`)
	placeholderPattern        = regexp.MustCompile(`\[[A-Z0-9_][A-Z0-9_ -]*\]`)
	acceptanceIDPattern       = regexp.MustCompile(`AC-[0-9][0-9][0-9]`)
	requirementIDPattern      = regexp.MustCompile(`RQ-[0-9][0-9][0-9]`)
	decisionIDPattern         = regexp.MustCompile(`DEC-[0-9][0-9][0-9]`)
	needsClarificationPattern = regexp.MustCompile(`\[NEEDS CLARIFICATION`)

	ambiguityPhrases = []string{
		"should",
		"appropriate",
		"fast",
		"user-friendly",
		"as needed",
		"if possible",
		"when appropriate",
		"при необходимости",
		"если возможно",
		"удобн",
		"быстр",
		"понятн",
	}
)

func (r *CheckResult) AddOK(message string) {
	r.Lines = append(r.Lines, "OK: "+message)
}

func (r *CheckResult) AddWarn(message string) {
	r.Findings = append(r.Findings, CheckFinding{
		Code:     "warn",
		Severity: SeverityWarning,
		Message:  message,
	})
	r.Lines = append(r.Lines, "WARN: "+message)
	r.Warnings++
}

func (r *CheckResult) AddError(message string) {
	r.Findings = append(r.Findings, CheckFinding{
		Code:     "error",
		Severity: SeverityError,
		Message:  message,
	})
	r.Lines = append(r.Lines, "ERROR: "+message)
	r.Errors++
	r.Failed = true
}

func (r *CheckResult) AddRaw(line string) {
	r.Lines = append(r.Lines, line)
}

func (r *CheckResult) Merge(other CheckResult) {
	r.Lines = append(r.Lines, other.Lines...)
	r.Findings = append(r.Findings, other.Findings...)
	r.Errors += other.Errors
	r.Warnings += other.Warnings
	if other.Failed {
		r.Failed = true
	}
}

func (r *CheckResult) AddFinding(finding CheckFinding) {
	r.Findings = append(r.Findings, finding)
	switch finding.Severity {
	case SeverityError:
		r.Errors++
		r.Failed = true
		r.Lines = append(r.Lines, "ERROR: "+finding.Message)
	case SeverityWarning:
		r.Warnings++
		r.Lines = append(r.Lines, "WARN: "+finding.Message)
	default:
		r.Lines = append(r.Lines, "OK: "+finding.Message)
	}
}

func (r *CheckResult) AddStructuredOK(code string, category FindingCategory, artifact string, message string, refs ...string) {
	r.AddFinding(CheckFinding{
		Code:     code,
		Severity: SeverityOK,
		Category: category,
		Artifact: artifact,
		Message:  message,
		Refs:     refs,
	})
}

func (r *CheckResult) AddStructuredWarn(code string, category FindingCategory, artifact string, message string, refs ...string) {
	r.AddFinding(CheckFinding{
		Code:     code,
		Severity: SeverityWarning,
		Category: category,
		Artifact: artifact,
		Message:  message,
		Refs:     refs,
	})
}

func (r *CheckResult) AddStructuredError(code string, category FindingCategory, artifact string, message string, refs ...string) {
	r.AddFinding(CheckFinding{
		Code:     code,
		Severity: SeverityError,
		Category: category,
		Artifact: artifact,
		Message:  message,
		Refs:     refs,
	})
}

func checkConstitutionLanguagePolicy(result *CheckResult, constitutionContent, targetLanguage string) {
	re := regexp.MustCompile(`(?m)^-\s*docs:\s*(\S+)`)
	matches := re.FindStringSubmatch(constitutionContent)
	if len(matches) < 2 {
		return
	}
	docsLang := strings.TrimSpace(matches[1])
	if strings.EqualFold(docsLang, targetLanguage) {
		result.AddStructuredOK("constitution_language_consistent", CategoryConsistency, "constitution",
			fmt.Sprintf("constitution docs language is %s", docsLang))
	} else {
		result.AddStructuredWarn("constitution_language_mismatch", CategoryConsistency, "constitution",
			fmt.Sprintf("constitution docs language is %s, but project is configured for %s", docsLang, targetLanguage))
	}
}
