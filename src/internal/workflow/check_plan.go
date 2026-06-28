package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"speckeep/src/internal/config"
)

func CheckPlanReady(ctx context.Context, cfg config.Config, root, slug string) (CheckResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return CheckResult{}, err
	}
	result := CheckResult{}
	specDisplay, specAbs := resolveSpecDisplayPath(root, cfg.Paths.SpecsDir, slug)
	inspectDisplay, inspectAbs := resolveInspectDisplayPath(root, cfg.Paths.SpecsDir, slug)
	legacyInspectDisplay := joinDisplay(cfg.Paths.SpecsDir, slug, "inspect.md")
	templateDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.Plan)
	promptDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.PlanPrompt)
	checkFile(&result, cfg.Project.ConstitutionFile, absFromRoot(root, cfg.Project.ConstitutionFile))
	checkFile(&result, specDisplay, specAbs)
	checkFile(&result, templateDisplay, absFromRoot(root, templateDisplay))
	checkFile(&result, promptDisplay, absFromRoot(root, promptDisplay))
	if fileExists(specAbs) {
		content, err := os.ReadFile(specAbs)
		if err != nil {
			return CheckResult{}, fmt.Errorf("read spec %s: %w", specDisplay, err)
		}
		checkPattern(&result, string(content), `(?m)^## (Критерии приемки|Acceptance Criteria)$`, "spec has acceptance criteria section")
		checkPattern(&result, string(content), acceptanceIDPattern.String(), "spec has stable acceptance IDs")
	}
	inspectDisplayPath := inspectDisplay
	if !fileExists(inspectAbs) {
		legacyAbs := absFromRoot(root, legacyInspectDisplay)
		if fileExists(legacyAbs) {
			inspectDisplayPath = legacyInspectDisplay
			inspectAbs = legacyAbs
		} else {
			result.AddWarn(fmt.Sprintf("no inspect report %s — inspect is optional; run /speckeep.inspect for a deep quality review", inspectDisplay))
		}
	}
	if fileExists(inspectAbs) {
		report, err := ParseReport(ctx, inspectAbs)
		if err != nil {
			result.AddError(err.Error())
		} else {
			if ValidStatus(report.Status) {
				result.AddOK("inspect report has a valid status")
			} else {
				result.AddError("inspect report has a valid status")
			}
			if report.Status == StatusBlocked {
				result.AddError("inspect report is blocked")
			}
		}
		_ = inspectDisplayPath
	}
	if fileExists(specAbs) {
		inspectResult, err := InspectSpec(ctx, cfg, root, specDisplay, "")
		if err != nil {
			return CheckResult{}, err
		}
		result.Merge(inspectResult)
	}
	return result, nil
}

func checkPlanContent(result *CheckResult, slug, specAbs, planDisplay, planContent string) {
	if len(decisionIDPattern.FindAllString(planContent, -1)) == 0 {
		result.AddStructuredWarn("plan_no_decision_ids", CategoryTraceability, planDisplay,
			fmt.Sprintf("plan has no stable decision IDs (DEC-*) for slug %s", slug))
	}
	if !ContainsAny(planContent, "## Implementation Surfaces", "## Поверхности реализации", "## Реализационные поверхности") {
		result.AddStructuredWarn("plan_missing_implementation_surfaces", CategoryStructure, planDisplay,
			fmt.Sprintf("plan is missing Implementation Surfaces section for slug %s", slug))
	}
	if !ContainsAny(planContent, "## Data and Contracts", "## Данные и контракты") {
		result.AddStructuredWarn("plan_missing_data_contracts", CategoryStructure, planDisplay,
			fmt.Sprintf("plan is missing Data and Contracts section for slug %s", slug))
	}
	if !ContainsAny(planContent, "## Acceptance Approach") {
		result.AddStructuredWarn("plan_missing_acceptance_approach", CategoryStructure, planDisplay,
			fmt.Sprintf("plan is missing Acceptance Approach section for slug %s", slug))
	}
	if !ContainsAny(planContent, "## Constitution Compliance", "## Соответствие конституции") {
		result.AddStructuredWarn("plan_missing_constitution_compliance", CategoryStructure, planDisplay,
			fmt.Sprintf("plan is missing Constitution Compliance section for slug %s", slug))
	}
	if !ContainsAny(planContent, "## Acceptance Approach") || !fileExists(specAbs) {
		return
	}
	specContent, err := os.ReadFile(specAbs)
	if err != nil {
		return
	}
	specIDs := ExtractUniqueMatches(string(specContent), `AC-[0-9][0-9][0-9]`)
	if len(specIDs) == 0 {
		return
	}
	approachSection := markdownSection(planContent, "Acceptance Approach")
	planACSet := make(map[string]struct{})
	for _, id := range ExtractUniqueMatches(approachSection, `AC-[0-9][0-9][0-9]`) {
		planACSet[id] = struct{}{}
	}
	specIDSet := make(map[string]struct{}, len(specIDs))
	for _, id := range specIDs {
		specIDSet[id] = struct{}{}
		if _, ok := planACSet[id]; !ok {
			result.AddStructuredWarn("plan_missing_ac_reference", CategoryConsistency, planDisplay,
				fmt.Sprintf("plan Acceptance Approach does not reference %s for slug %s", id, slug), id)
		}
	}
	for id := range planACSet {
		if _, ok := specIDSet[id]; !ok {
			result.AddStructuredWarn("plan_unknown_ac_reference", CategoryConsistency, planDisplay,
				fmt.Sprintf("plan Acceptance Approach references unknown criterion %s for slug %s", id, slug), id)
		}
	}
}
