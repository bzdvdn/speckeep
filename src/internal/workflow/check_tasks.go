package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"speckeep/src/internal/config"
)

func CheckTasksReady(ctx context.Context, cfg config.Config, root, slug string) (CheckResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return CheckResult{}, err
	}
	result := CheckResult{}
	specDisplay, specAbs := resolveSpecDisplayPath(root, cfg.Paths.SpecsDir, slug)
	planDisplay, planAbs := resolvePlanDisplayPath(root, cfg.Paths.SpecsDir, slug)
	dataModelDisplay, dataModelAbs := resolveDataModelDisplayPath(root, cfg.Paths.SpecsDir, slug)
	tasksTemplateDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.Tasks)
	promptDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.TasksPrompt)
	contractsDisplay, contractsAbs := resolveContractsDisplayPath(root, cfg.Paths.SpecsDir, slug)
	checkFile(&result, cfg.Project.ConstitutionFile, absFromRoot(root, cfg.Project.ConstitutionFile))
	checkFile(&result, specDisplay, specAbs)
	checkFile(&result, planDisplay, planAbs)
	checkFile(&result, dataModelDisplay, dataModelAbs)
	checkFile(&result, tasksTemplateDisplay, absFromRoot(root, tasksTemplateDisplay))
	checkFile(&result, promptDisplay, absFromRoot(root, promptDisplay))
	if isDir(contractsAbs) {
		result.AddOK(contractsDisplay)
	} else {
		result.AddOK("optional contracts directory not present")
	}
	if fileExists(specAbs) {
		content, err := os.ReadFile(specAbs)
		if err != nil {
			return CheckResult{}, fmt.Errorf("read spec %s: %w", specDisplay, err)
		}
		checkPattern(&result, string(content), acceptanceIDPattern.String(), "spec has stable acceptance IDs")
		inspectResult, err := InspectSpec(ctx, cfg, root, specDisplay, "")
		if err != nil {
			return CheckResult{}, err
		}
		result.Merge(inspectResult)
	}
	if fileExists(planAbs) {
		content, err := os.ReadFile(planAbs)
		if err != nil {
			return CheckResult{}, fmt.Errorf("read plan %s: %w", planDisplay, err)
		}
		checkPattern(&result, string(content), decisionIDPattern.String(), "plan has stable decision IDs")
	}
	return result, nil
}

func checkTaskTraceability(result *CheckResult, tasksDisplay, acceptanceBody, tasksContent string) {
	specIDs := ExtractUniqueMatches(acceptanceBody, `AC-[0-9][0-9][0-9]`)
	if len(specIDs) == 0 {
		return
	}
	taskIDs := extractTaskDefinitionIDs(tasksContent)
	taskIDSet := make(map[string]struct{}, len(taskIDs))
	for _, id := range taskIDs {
		taskIDSet[id] = struct{}{}
	}
	if len(taskIDs) > 0 {
		if !ContainsAny(tasksContent, "## Surface Map") {
			result.AddStructuredWarn("surface_map_missing", CategoryTraceability, "tasks", "tasks are missing Surface Map section")
		}
		if countTasksWithoutTouches(tasksContent) > 0 {
			result.AddStructuredWarn("task_touches_missing", CategoryTraceability, "tasks", "tasks contain task lines without Touches: field")
		}
		if hasDuplicateTaskDefinitionIDs(tasksContent) {
			result.AddStructuredWarn("duplicate_task_ids", CategoryTraceability, "tasks", "tasks contain duplicate task IDs")
		}
	}
	coverageSection := acceptanceCoverageSection(tasksContent)
	specIDSet := make(map[string]struct{}, len(specIDs))
	for _, id := range specIDs {
		specIDSet[id] = struct{}{}
	}
	coverageIDs := make(map[string]struct{})
	for _, id := range ExtractUniqueMatches(tasksContent, `AC-[0-9][0-9][0-9]\s*(?:->|→)`) {
		coverageIDs[trimCoverageArrowSuffix(id)] = struct{}{}
	}
	for _, taskID := range ExtractUniqueMatches(coverageSection, `T[0-9]+\.[0-9]+`) {
		if _, ok := taskIDSet[taskID]; !ok {
			result.AddStructuredError("unknown_task_reference", CategoryTraceability, tasksDisplay, fmt.Sprintf("acceptance coverage references unknown task ID %s", taskID), taskID)
		}
	}
	for _, id := range ExtractUniqueMatches(coverageSection, `AC-[0-9][0-9][0-9]`) {
		if _, ok := specIDSet[id]; !ok {
			result.AddStructuredError("unknown_acceptance_reference", CategoryTraceability, tasksDisplay, fmt.Sprintf("acceptance coverage references unknown acceptance criterion %s", id), id)
		}
	}
	for _, id := range specIDs {
		if _, ok := coverageIDs[id]; !ok {
			result.AddStructuredError("acceptance_not_covered", CategoryTraceability, tasksDisplay, fmt.Sprintf("acceptance criterion %s is not covered by tasks", id), id)
		}
	}
}

func checkPlanTaskSurfaceConsistency(result *CheckResult, planDisplay, planContent, tasksDisplay, tasksContent string) {
	planSurfaces := extractPlanSurfaceRefs(planContent)
	taskSurfaces := extractTaskSurfaceRefs(tasksContent)
	if len(planSurfaces) == 0 || len(taskSurfaces) == 0 {
		return
	}
	taskSurfaceSet := make(map[string]struct{}, len(taskSurfaces))
	for _, surface := range taskSurfaces {
		taskSurfaceSet[surface] = struct{}{}
	}
	for _, surface := range planSurfaces {
		if _, ok := taskSurfaceSet[surface]; !ok {
			result.AddStructuredWarn("plan_surface_missing_from_tasks", CategoryConsistency, tasksDisplay, fmt.Sprintf("planned surface %s is not referenced in tasks surface map or Touches:", surface), surface)
		}
	}
	planSurfaceSet := make(map[string]struct{}, len(planSurfaces))
	for _, surface := range planSurfaces {
		planSurfaceSet[surface] = struct{}{}
	}
	for _, surface := range taskSurfaces {
		if _, ok := planSurfaceSet[surface]; !ok {
			result.AddStructuredWarn("task_surface_missing_from_plan", CategoryConsistency, planDisplay, fmt.Sprintf("task surface %s is not referenced in plan implementation surfaces", surface), surface)
		}
	}
}

func extractPlanSurfaceRefs(planContent string) []string {
	section := markdownSection(planContent, "Implementation Surfaces")
	if strings.TrimSpace(section) == "" {
		section = markdownSection(planContent, "Реализационные поверхности")
	}
	if strings.TrimSpace(section) == "" {
		section = markdownSection(planContent, "Поверхности реализации")
	}
	return extractSurfaceRefs(section)
}

func extractTaskSurfaceRefs(tasksContent string) []string {
	var refs []string
	refs = append(refs, extractSurfaceRefs(markdownSection(tasksContent, "Surface Map"))...)
	refs = append(refs, extractTouchesRefs(tasksContent)...)
	return uniqueStrings(refs)
}

func extractSurfaceRefs(content string) []string {
	re := regexp.MustCompile(`(?:^|[\s` + "`" + `:(])([A-Za-z0-9_./-]+\.[A-Za-z0-9_./-]+|[A-Za-z0-9_./-]+/)`)
	raw := re.FindAllStringSubmatch(content, -1)
	refs := make([]string, 0, len(raw))
	for _, match := range raw {
		if len(match) != 2 {
			continue
		}
		value := strings.TrimSpace(match[1])
		if value == "" {
			continue
		}
		if taskIDPattern.MatchString(value) {
			continue
		}
		if acceptanceIDPattern.MatchString(value) || decisionIDPattern.MatchString(value) || requirementIDPattern.MatchString(value) {
			continue
		}
		refs = append(refs, value)
	}
	return uniqueStrings(refs)
}

func extractTouchesRefs(tasksContent string) []string {
	lines := strings.Split(tasksContent, "\n")
	var refs []string
	for _, line := range lines {
		idx := strings.Index(line, "Touches:")
		if idx < 0 {
			continue
		}
		part := strings.TrimSpace(line[idx+len("Touches:"):])
		for _, piece := range strings.Split(part, ",") {
			value := strings.TrimSpace(piece)
			if value == "" {
				continue
			}
			refs = append(refs, value)
		}
	}
	return uniqueStrings(refs)
}

func checkTouchesFilesExist(result *CheckResult, root, tasksDisplay, tasksContent string) {
	for _, line := range strings.Split(tasksContent, "\n") {
		idx := strings.Index(line, "Touches:")
		if idx < 0 {
			continue
		}
		part := strings.TrimSpace(line[idx+len("Touches:"):])
		for _, piece := range strings.Split(part, ",") {
			value := strings.TrimSpace(strings.Trim(strings.TrimSpace(piece), "`"))
			if value == "" || strings.ContainsAny(value, " \t") {
				continue
			}
			absPath := absFromRoot(root, value)
			if !fileExists(absPath) && !isDir(absPath) {
				result.AddStructuredWarn("touches_file_missing", CategoryTraceability, tasksDisplay,
					fmt.Sprintf("Touches: references non-existent path: %s", value), value)
			}
		}
	}
}

func countMatchingLines(content, pattern string) int {
	return len(regexp.MustCompile(pattern).FindAllString(content, -1))
}

func countMalformedCoverageLines(content string) int {
	lines := strings.Split(content, "\n")
	count := 0
	for _, line := range lines {
		if !containsCoverageArrow(line) {
			continue
		}
		if !coverageLinePattern.MatchString(line) {
			count++
		}
	}
	return count
}

func containsCoverageArrow(line string) bool {
	return strings.Contains(line, "->") || strings.Contains(line, "→")
}

func trimCoverageArrowSuffix(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, "->")
	value = strings.TrimSuffix(value, "→")
	return strings.TrimSpace(value)
}

func checkAmbiguousLanguage(result *CheckResult, path, section, content string) {
	lower := strings.ToLower(content)
	refs := ExtractUniqueMatches(content, `(?:AC|RQ)-[0-9][0-9][0-9]`)
	for _, phrase := range ambiguityPhrases {
		if !strings.Contains(lower, phrase) {
			continue
		}
		message := fmt.Sprintf("ambiguous wording detected in %s: %q", section, phrase)
		result.Findings = append(result.Findings, CheckFinding{
			Code:     "ambiguous_wording",
			Severity: SeverityWarning,
			Category: CategoryAmbiguity,
			Artifact: "spec",
			Path:     path,
			Section:  section,
			Message:  message,
			Refs:     append([]string(nil), refs...),
		})
		result.Warnings++
		result.Lines = append(result.Lines, "WARN: "+message)
	}
}
