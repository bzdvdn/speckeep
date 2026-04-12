package workflow

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"speckeep/src/internal/config"
	"speckeep/src/internal/featurepaths"
)

type Finding struct {
	Level   string
	Message string
}

var findingSlugPattern = regexp.MustCompile(`\bfor slug ([a-z0-9][a-z0-9-]*)\b`)

func ValidateProject(root string) ([]Finding, error) {
	cfg, err := config.Load(root)
	if err != nil {
		return nil, err
	}

	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return nil, err
	}

	states, err := States(root)
	if err != nil {
		return nil, err
	}

	var findings []Finding
	for _, state := range states {
		specPath, _ := featurepaths.ResolveSpec(specsDir, state.Slug)
		planPath := featurepaths.Plan(specsDir, state.Slug)
		tasksPath := featurepaths.Tasks(specsDir, state.Slug)

		if state.InspectExists && !state.SpecExists {
			findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("inspect report exists without matching spec for slug %s", state.Slug)})
		}
		if state.InspectLegacy {
			findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("legacy inspect report path in use for slug %s; run speckeep feature repair %s or speckeep migrate", state.Slug, state.Slug)})
		}
		if state.PlanExists && !state.InspectExists {
			findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("plan exists without mandatory inspect report for slug %s", state.Slug)})
		}
		if state.TasksExists && !state.PlanExists {
			findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("tasks exist without matching plan for slug %s", state.Slug)})
		}
		if state.VerifyExists && !state.TasksExists {
			findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("verify report exists without tasks for slug %s", state.Slug)})
		}

		if state.InspectExists {
			report, err := ParseReport(state.InspectPath)
			if err != nil {
				findings = append(findings, Finding{Level: "error", Message: err.Error()})
			} else {
				findings = append(findings, validateReport(state.Slug, report, ReportTypeInspect)...)
				findings = append(findings, validateReportSemantics(state.Slug, state.InspectPath, report, state.TasksOpen)...)
				if report.Status == StatusBlocked && (state.PlanExists || state.TasksExists || state.VerifyExists || state.Archived) {
					findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("downstream artifacts exist after blocked inspect for slug %s", state.Slug)})
				}
			}
		}

		if state.VerifyExists {
			report, err := ParseReport(state.VerifyPath)
			if err != nil {
				findings = append(findings, Finding{Level: "error", Message: err.Error()})
			} else {
				findings = append(findings, validateReport(state.Slug, report, ReportTypeVerify)...)
				findings = append(findings, validateReportSemantics(state.Slug, state.VerifyPath, report, state.TasksOpen)...)
				findings = append(findings, validateVerifyReportTraceability(state.Slug, specPath, tasksPath, state.VerifyPath, report)...)
				if report.Status == StatusBlocked && state.Archived {
					findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("archive exists after blocked verify for slug %s", state.Slug)})
				}
			}
		}

		if state.SpecExists {
			findings = append(findings, validateSpec(state.Slug, specPath)...)
		}
		if state.SpecExists && state.TasksExists {
			findings = append(findings, validateTasks(state.Slug, specPath, tasksPath)...)
		}
		if state.PlanExists {
			findings = append(findings, validatePlan(state.Slug, specPath, planPath)...)
		}
	}

	findings = append(findings, validateCrossSpecIDs(states, specsDir)...)

	return findings, nil
}

func ValidateFeature(root, slug string) ([]Finding, error) {
	findings, err := ValidateProject(root)
	if err != nil {
		return nil, err
	}

	filtered := make([]Finding, 0, len(findings))
	for _, finding := range findings {
		if FindingSlug(finding) == slug {
			filtered = append(filtered, finding)
		}
	}
	return filtered, nil
}

func FindingSlug(finding Finding) string {
	match := findingSlugPattern.FindStringSubmatch(finding.Message)
	if len(match) == 2 {
		return match[1]
	}
	return ""
}

func validateReport(slug string, report Report, expectedType string) []Finding {
	var findings []Finding
	if report.Status == "" {
		findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("%s report is missing a valid status for slug %s", expectedType, slug)})
	}
	if report.Type != "" && report.Type != expectedType {
		findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("%s report metadata has wrong report_type for slug %s", expectedType, slug)})
	}
	if report.Slug != "" && report.Slug != slug {
		findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("%s report metadata slug mismatch for slug %s", expectedType, slug)})
	}
	return findings
}

func validateSpec(slug, path string) []Finding {
	content, err := os.ReadFile(path)
	if err != nil {
		return []Finding{{Level: "error", Message: fmt.Sprintf("read spec for slug %s: %v", slug, err)}}
	}

	text := string(content)
	var findings []Finding
	if !ContainsAny(text, "## Goal", "## Цель") {
		findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("spec is missing Goal section for slug %s", slug)})
	}
	if !ContainsAny(text, "## Requirements", "## Требования") {
		findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("spec is missing Requirements section for slug %s", slug)})
	}
	if !ContainsAny(text, "## Acceptance Criteria", "## Критерии приемки") {
		findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("spec is missing Acceptance Criteria section for slug %s", slug)})
	}
	if !ContainsAny(text, "Given") || !ContainsAny(text, "When") || !ContainsAny(text, "Then") {
		findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("spec is missing canonical Given/When/Then markers for slug %s", slug)})
	}
	if len(ExtractUniqueMatches(text, `AC-[0-9][0-9][0-9]`)) == 0 {
		findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("spec has no stable acceptance IDs for slug %s", slug)})
	}
	return findings
}

func validateReportSemantics(slug, path string, report Report, tasksOpen int) []Finding {
	content, err := os.ReadFile(path)
	if err != nil {
		return []Finding{{Level: "error", Message: fmt.Sprintf("read %s report for slug %s: %v", report.Type, slug, err)}}
	}

	_, body, err := splitFrontmatter(string(content))
	if err != nil {
		return []Finding{{Level: "error", Message: fmt.Sprintf("parse %s report for slug %s: %v", report.Type, slug, err)}}
	}

	errorsSection := materialSectionItems(body, "Errors")
	warningsSection := materialSectionItems(body, "Warnings")
	questionsSection := materialSectionItems(body, "Questions")

	var findings []Finding
	switch report.Status {
	case StatusPass:
		if len(errorsSection) > 0 {
			findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("%s report says pass but still lists errors for slug %s", report.Type, slug)})
		}
		if report.Type == ReportTypeVerify && tasksOpen > 0 {
			findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("verify report says pass while open tasks remain for slug %s", slug)})
		}
	case StatusBlocked:
		if len(errorsSection) == 0 {
			findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("%s report says blocked but lists no explicit errors for slug %s", report.Type, slug)})
		}
	case StatusConcerns:
		if len(errorsSection) == 0 && len(warningsSection) == 0 && len(questionsSection) == 0 {
			findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("%s report says concerns but lists no explicit issues for slug %s", report.Type, slug)})
		}
	}

	return findings
}

func validateVerifyReportTraceability(slug, specPath, tasksPath, verifyPath string, report Report) []Finding {
	content, err := os.ReadFile(verifyPath)
	if err != nil {
		return []Finding{{Level: "error", Message: fmt.Sprintf("read verify report for slug %s: %v", slug, err)}}
	}

	_, body, err := splitFrontmatter(string(content))
	if err != nil {
		return []Finding{{Level: "error", Message: fmt.Sprintf("parse verify report for slug %s: %v", slug, err)}}
	}

	var findings []Finding

	archiveReadiness := sectionBulletValue(body, "Verdict", "archive_readiness")
	nextStepItems := materialSectionItems(body, "Next Step")
	notVerifiedItems := materialSectionItems(body, "Not Verified")

	switch report.Status {
	case StatusPass:
		if archiveReadiness == "blocked" {
			findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("verify report says pass but marks archive_readiness blocked for slug %s", slug)})
		}
		if len(notVerifiedItems) > 0 {
			findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("verify report says pass but still lists not-verified areas for slug %s", slug)})
		}
	case StatusConcerns:
		if archiveReadiness == "safe" {
			findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("verify report says concerns but still marks archive_readiness safe for slug %s", slug)})
		}
	case StatusBlocked:
		if archiveReadiness == "safe" {
			findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("verify report says blocked but still marks archive_readiness safe for slug %s", slug)})
		}
		for _, item := range nextStepItems {
			if strings.Contains(strings.ToLower(item), "safe to archive") {
				findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("verify report says blocked but next step still says safe to archive for slug %s", slug)})
				break
			}
		}
	}

	if !ContainsAny(body, "## Not Verified") {
		findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("verify report is missing Not Verified section for slug %s", slug)})
	}

	if fileExists(tasksPath) {
		tasksContent, err := os.ReadFile(tasksPath)
		if err != nil {
			return []Finding{{Level: "error", Message: fmt.Sprintf("read tasks for verify check, slug %s: %v", slug, err)}}
		}
		if len(extractTaskDefinitionIDs(string(tasksContent))) > 0 && len(ExtractUniqueMatches(body, `T[0-9]+\.[0-9]+`)) == 0 {
			findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("verify report does not reference any task IDs for slug %s", slug)})
		}
	}

	if fileExists(specPath) {
		specContent, err := os.ReadFile(specPath)
		if err != nil {
			return []Finding{{Level: "error", Message: fmt.Sprintf("read spec for verify check, slug %s: %v", slug, err)}}
		}
		if len(ExtractUniqueMatches(string(specContent), `AC-[0-9][0-9][0-9]`)) > 0 && len(ExtractUniqueMatches(body, `AC-[0-9][0-9][0-9]`)) == 0 {
			findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("verify report does not reference any acceptance criteria for slug %s", slug)})
		}
	}

	return findings
}

func validatePlan(slug, specPath, path string) []Finding {
	content, err := os.ReadFile(path)
	if err != nil {
		return []Finding{{Level: "error", Message: fmt.Sprintf("read plan for slug %s: %v", slug, err)}}
	}
	text := string(content)

	var findings []Finding
	if len(ExtractUniqueMatches(text, `DEC-[0-9][0-9][0-9]`)) == 0 {
		findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("plan has no stable decision IDs for slug %s", slug)})
	}
	if !ContainsAny(text, "## Acceptance Approach") {
		findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("plan is missing Acceptance Approach section for slug %s", slug)})
	}
	if !ContainsAny(text, "## Constitution Compliance", "## Соответствие конституции") {
		findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("plan is missing Constitution Compliance section for slug %s", slug)})
	}
	if !ContainsAny(text, "## Acceptance Approach") {
		return findings
	}

	if !fileExists(specPath) {
		return findings
	}

	specContent, err := os.ReadFile(specPath)
	if err != nil {
		return []Finding{{Level: "error", Message: fmt.Sprintf("read spec for plan check, slug %s: %v", slug, err)}}
	}
	specIDs := ExtractUniqueMatches(string(specContent), `AC-[0-9][0-9][0-9]`)
	if len(specIDs) == 0 {
		return findings
	}

	acceptanceApproach := markdownSection(text, "Acceptance Approach")
	planACIDs := make(map[string]struct{})
	for _, id := range ExtractUniqueMatches(acceptanceApproach, `AC-[0-9][0-9][0-9]`) {
		planACIDs[id] = struct{}{}
	}

	specIDSet := make(map[string]struct{}, len(specIDs))
	for _, id := range specIDs {
		specIDSet[id] = struct{}{}
		if _, ok := planACIDs[id]; !ok {
			findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("plan does not reference acceptance criterion %s for slug %s", id, slug)})
		}
	}
	for id := range planACIDs {
		if _, ok := specIDSet[id]; !ok {
			findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("plan references unknown acceptance criterion %s for slug %s", id, slug)})
		}
	}

	return findings
}

func validateTasks(slug, specPath, tasksPath string) []Finding {
	specContent, err := os.ReadFile(specPath)
	if err != nil {
		return []Finding{{Level: "error", Message: fmt.Sprintf("read spec for coverage check, slug %s: %v", slug, err)}}
	}
	tasksContent, err := os.ReadFile(tasksPath)
	if err != nil {
		return []Finding{{Level: "error", Message: fmt.Sprintf("read tasks for coverage check, slug %s: %v", slug, err)}}
	}

	specIDs := ExtractUniqueMatches(string(specContent), `AC-[0-9][0-9][0-9]`)
	if len(specIDs) == 0 {
		return nil
	}

	tasksText := string(tasksContent)
	if !ContainsAny(tasksText, "## Acceptance Coverage", "## Покрытие критериев приемки") {
		return []Finding{{Level: "error", Message: fmt.Sprintf("tasks are missing Acceptance Coverage section for slug %s", slug)}}
	}

	taskIDs := extractTaskDefinitionIDs(tasksText)
	taskIDSet := make(map[string]struct{}, len(taskIDs))
	for _, id := range taskIDs {
		taskIDSet[id] = struct{}{}
	}

	var findings []Finding
	if len(taskIDs) > 0 {
		if !ContainsAny(tasksText, "## Surface Map") {
			findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("tasks are missing Surface Map section for slug %s", slug)})
		}
		if countTasksWithoutTouches(tasksText) > 0 {
			findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("tasks contain task lines without Touches: field for slug %s", slug)})
		}
	}

	if hasDuplicateTaskDefinitionIDs(tasksText) {
		// Duplicate IDs make downstream references ambiguous.
		findings = append(findings, Finding{Level: "warning", Message: fmt.Sprintf("tasks contain duplicate task IDs for slug %s", slug)})
		findings = append(findings, validateTaskCoverage(slug, specIDs, tasksText, taskIDSet)...)
		return findings
	}

	findings = append(findings, validateTaskCoverage(slug, specIDs, tasksText, taskIDSet)...)
	return findings
}

func validateTaskCoverage(slug string, specIDs []string, tasksText string, taskIDSet map[string]struct{}) []Finding {
	coverageIDs := make(map[string]struct{})
	for _, id := range ExtractUniqueMatches(tasksText, `AC-[0-9][0-9][0-9]\s*(?:->|→)`) {
		coverageIDs[trimCoverageArrowSuffix(id)] = struct{}{}
	}
	specIDSet := make(map[string]struct{}, len(specIDs))
	for _, id := range specIDs {
		specIDSet[id] = struct{}{}
	}

	coverageSection := acceptanceCoverageSection(tasksText)
	for _, taskID := range ExtractUniqueMatches(coverageSection, `T[0-9]+\.[0-9]+`) {
		if _, ok := taskIDSet[taskID]; !ok {
			findings := []Finding{{Level: "error", Message: fmt.Sprintf("acceptance coverage references unknown task ID %s for slug %s", taskID, slug)}}
			for _, id := range ExtractUniqueMatches(coverageSection, `AC-[0-9][0-9][0-9]`) {
				if _, ok := specIDSet[id]; !ok {
					findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("acceptance coverage references unknown acceptance criterion %s for slug %s", id, slug)})
				}
			}
			for _, id := range specIDs {
				if _, ok := coverageIDs[id]; !ok {
					findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("acceptance criterion %s is not covered by tasks for slug %s", id, slug)})
				}
			}
			return findings
		}
	}

	var findings []Finding
	for _, id := range ExtractUniqueMatches(coverageSection, `AC-[0-9][0-9][0-9]`) {
		if _, ok := specIDSet[id]; !ok {
			findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("acceptance coverage references unknown acceptance criterion %s for slug %s", id, slug)})
		}
	}
	for _, id := range specIDs {
		if _, ok := coverageIDs[id]; !ok {
			findings = append(findings, Finding{Level: "error", Message: fmt.Sprintf("acceptance criterion %s is not covered by tasks for slug %s", id, slug)})
		}
	}
	return findings
}

func materialSectionItems(content, section string) []string {
	body := markdownSection(content, section)
	var items []string
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		if value == "" || strings.EqualFold(value, "none") {
			continue
		}
		items = append(items, value)
	}
	return items
}

func sectionBulletValue(content, section, key string) string {
	body := markdownSection(content, section)
	wantPrefix := "- " + key + ":"
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, wantPrefix) {
			continue
		}
		return strings.TrimSpace(strings.TrimPrefix(trimmed, wantPrefix))
	}
	return ""
}

func hasDuplicateTaskDefinitionIDs(content string) bool {
	matches := extractTaskDefinitionIDs(content)
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		if _, ok := seen[match]; ok {
			return true
		}
		seen[match] = struct{}{}
	}
	return false
}

func countTasksWithoutTouches(content string) int {
	re := regexp.MustCompile(`(?m)^- \[[ x]\]\s+T[0-9]+\.[0-9]+\b`)
	matches := re.FindAllStringIndex(content, -1)
	count := 0
	for _, loc := range matches {
		// Find the end of this line.
		lineEnd := strings.Index(content[loc[0]:], "\n")
		var line string
		if lineEnd == -1 {
			line = content[loc[0]:]
		} else {
			line = content[loc[0] : loc[0]+lineEnd]
		}
		if !strings.Contains(line, "Touches:") {
			count++
		}
	}
	return count
}

func extractTaskDefinitionIDs(content string) []string {
	re := regexp.MustCompile(`(?m)^- \[[ x]\]\s+(T[0-9]+\.[0-9]+)\b`)
	raw := re.FindAllStringSubmatch(content, -1)
	ids := make([]string, 0, len(raw))
	for _, match := range raw {
		if len(match) == 2 {
			ids = append(ids, match[1])
		}
	}
	return ids
}

func acceptanceCoverageSection(tasksText string) string {
	if ContainsAny(tasksText, "## Acceptance Coverage") {
		return markdownSection(tasksText, "Acceptance Coverage")
	}
	return markdownSection(tasksText, "Покрытие критериев приемки")
}

func validateCrossSpecIDs(states []FeatureState, specsDir string) []Finding {
	type occurrence struct{ slugs []string }
	acRegistry := map[string]*occurrence{}
	rqRegistry := map[string]*occurrence{}

	for _, state := range states {
		if !state.SpecExists {
			continue
		}
		specPath, _ := featurepaths.ResolveSpec(specsDir, state.Slug)
		content, err := os.ReadFile(specPath)
		if err != nil {
			continue
		}
		text := string(content)
		for _, id := range ExtractUniqueMatches(text, `AC-[0-9][0-9][0-9]`) {
			if acRegistry[id] == nil {
				acRegistry[id] = &occurrence{}
			}
			acRegistry[id].slugs = append(acRegistry[id].slugs, state.Slug)
		}
		for _, id := range ExtractUniqueMatches(text, `RQ-[0-9][0-9][0-9]`) {
			if rqRegistry[id] == nil {
				rqRegistry[id] = &occurrence{}
			}
			rqRegistry[id].slugs = append(rqRegistry[id].slugs, state.Slug)
		}
	}

	var findings []Finding
	for id, occ := range acRegistry {
		if len(occ.slugs) > 1 {
			sort.Strings(occ.slugs)
			findings = append(findings, Finding{
				Level:   "warning",
				Message: fmt.Sprintf("stable ID %s appears in multiple specs: %s", id, strings.Join(occ.slugs, ", ")),
			})
		}
	}
	for id, occ := range rqRegistry {
		if len(occ.slugs) > 1 {
			sort.Strings(occ.slugs)
			findings = append(findings, Finding{
				Level:   "warning",
				Message: fmt.Sprintf("stable ID %s appears in multiple specs: %s", id, strings.Join(occ.slugs, ", ")),
			})
		}
	}
	sort.Slice(findings, func(i, j int) bool { return findings[i].Message < findings[j].Message })
	return findings
}
