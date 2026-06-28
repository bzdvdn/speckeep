package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"speckeep/src/internal/config"
)

func CheckInspectReady(ctx context.Context, cfg config.Config, root, slug string) (CheckResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return CheckResult{}, err
	}
	result := CheckResult{}
	specDisplay, specAbs := resolveSpecDisplayPath(root, cfg.Paths.SpecsDir, slug)
	reportTemplateDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.InspectReport)
	promptDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.InspectPrompt)
	checkFile(&result, cfg.Project.ConstitutionFile, absFromRoot(root, cfg.Project.ConstitutionFile))
	checkFile(&result, specDisplay, specAbs)
	checkFile(&result, reportTemplateDisplay, absFromRoot(root, reportTemplateDisplay))
	checkFile(&result, promptDisplay, absFromRoot(root, promptDisplay))
	if !result.Failed {
		inspectResult, err := InspectSpec(ctx, cfg, root, specDisplay, "")
		if err != nil {
			return CheckResult{}, err
		}
		result.Merge(inspectResult)
		if !inspectResult.Failed {
			result.AddOK(fmt.Sprintf("inspect is ready for slug '%s'", slug))
		}
	}
	return result, nil
}

func InspectSpec(ctx context.Context, cfg config.Config, root, specPath, tasksPath string) (CheckResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return CheckResult{}, err
	}
	specDisplay, specAbs := resolveUserPath(root, specPath)
	result := CheckResult{}
	if !fileExists(specAbs) {
		result.AddError(fmt.Sprintf("spec file not found: %s", specDisplay))
		return result, nil
	}
	constitutionAbs := absFromRoot(root, cfg.Project.ConstitutionFile)
	if fileExists(constitutionAbs) {
		if constitutionContent, readErr := os.ReadFile(constitutionAbs); readErr == nil {
			checkConstitutionLanguagePolicy(&result, string(constitutionContent), cfg.Language.Docs)
		}
	}
	content, err := os.ReadFile(specAbs)
	if err != nil {
		return CheckResult{}, fmt.Errorf("read spec %s: %w", specDisplay, err)
	}
	sections := docsSections(cfg.Language.Docs)
	text := string(content)
	checkRequiredHeading(&result, text, sections.Goal)
	checkOptionalHeading(&result, text, sections.Context)
	checkRequiredHeading(&result, text, sections.Requirements)
	checkRequiredHeading(&result, text, sections.Acceptance)
	checkOptionalHeading(&result, text, sections.Questions)
	checkOptionalHeading(&result, text, sections.Assumptions)
	if needsClarificationPattern.MatchString(text) {
		result.AddStructuredError("needs_clarification_marker", CategoryReadiness, "spec", "spec contains unresolved [NEEDS CLARIFICATION] markers")
	}
	acceptanceBody := markdownSection(text, sections.Acceptance)
	if strings.TrimSpace(acceptanceBody) == "" {
		result.AddError("empty acceptance criteria section")
	} else {
		if strings.Contains(acceptanceBody, "Given") {
			result.AddOK("Given marker found")
		} else {
			result.AddError("missing Given marker in acceptance criteria")
		}
		if strings.Contains(acceptanceBody, "When") {
			result.AddOK("When marker found")
		} else {
			result.AddError("missing When marker in acceptance criteria")
		}
		if strings.Contains(acceptanceBody, "Then") {
			result.AddOK("Then marker found")
		} else {
			result.AddError("missing Then marker in acceptance criteria")
		}
	}
	criteriaCount := countMatchingLines(acceptanceBody, `(?m)^### `)
	if criteriaCount > 0 {
		result.AddOK(fmt.Sprintf("acceptance criteria count: %d", criteriaCount))
	} else {
		result.AddWarn("no explicit acceptance criterion headings found")
	}
	acceptanceIDCount := len(acceptanceIDPattern.FindAllString(acceptanceBody, -1))
	if acceptanceIDCount > 0 {
		result.AddOK(fmt.Sprintf("acceptance IDs found: %d", acceptanceIDCount))
	} else {
		result.AddWarn("no stable acceptance IDs found in acceptance criteria")
	}
	requirementsBody := markdownSection(text, sections.Requirements)
	if strings.TrimSpace(requirementsBody) != "" {
		rqIDCount := len(requirementIDPattern.FindAllString(requirementsBody, -1))
		if rqIDCount > 0 {
			result.AddStructuredOK("requirement_ids_present", CategoryTraceability, "spec", fmt.Sprintf("requirement IDs found: %d", rqIDCount))
		} else {
			result.AddStructuredWarn("requirement_ids_missing", CategoryTraceability, "spec", "no stable requirement IDs (RQ-*) found in requirements section")
		}
	}
	checkAmbiguousLanguage(&result, specDisplay, sections.Requirements, requirementsBody)
	checkAmbiguousLanguage(&result, specDisplay, sections.Acceptance, acceptanceBody)
	if strings.TrimSpace(tasksPath) != "" {
		tasksDisplay, tasksAbs := resolveUserPath(root, tasksPath)
		if fileExists(tasksAbs) {
			tasksContentBytes, err := os.ReadFile(tasksAbs)
			if err != nil {
				return CheckResult{}, fmt.Errorf("read tasks %s: %w", tasksDisplay, err)
			}
			tasksContent := string(tasksContentBytes)
			if hasHeading(tasksContent, sections.Coverage) {
				result.AddOK(sections.Coverage)
				coverageBody := markdownSection(tasksContent, sections.Coverage)
				coverageLines := countMatchingLines(coverageBody, `(?m)(?:->|→)`)
				malformedLines := countMalformedCoverageLines(coverageBody)
				if criteriaCount > 0 && coverageLines < criteriaCount {
					result.AddError(fmt.Sprintf("acceptance coverage entries (%d) are fewer than acceptance criteria (%d)", coverageLines, criteriaCount))
				} else {
					result.AddOK(fmt.Sprintf("acceptance coverage entries: %d", coverageLines))
				}
				if acceptanceIDCount > 0 && coverageLines < acceptanceIDCount {
					result.AddError(fmt.Sprintf("acceptance coverage entries (%d) are fewer than acceptance IDs (%d)", coverageLines, acceptanceIDCount))
				}
				if malformedLines > 0 {
					result.AddError("acceptance coverage contains malformed entries; expected lines like AC-001 -> T1.1")
				} else if coverageLines > 0 {
					result.AddOK("acceptance coverage format uses AC and task IDs")
				}
			} else {
				result.AddError(fmt.Sprintf("tasks file is missing required section: %s", sections.Coverage))
			}
			checkTaskTraceability(&result, tasksDisplay, acceptanceBody, tasksContent)
		}
	}
	result.AddRaw(fmt.Sprintf("SUMMARY: errors=%d warnings=%d", result.Errors, result.Warnings))
	return result, nil
}
