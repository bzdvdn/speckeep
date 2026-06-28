package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"speckeep/src/internal/config"
)

func CheckImplementReady(ctx context.Context, cfg config.Config, root, slug string) (CheckResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return CheckResult{}, err
	}
	result := CheckResult{}
	specDisplay, specAbs := resolveSpecDisplayPath(root, cfg.Paths.SpecsDir, slug)
	planDisplay, planAbs := resolvePlanDisplayPath(root, cfg.Paths.SpecsDir, slug)
	tasksDisplay, tasksAbs := resolveTasksDisplayPath(root, cfg.Paths.SpecsDir, slug)
	dataModelDisplay, dataModelAbs := resolveDataModelDisplayPath(root, cfg.Paths.SpecsDir, slug)
	promptDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.ImplementPrompt)
	contractsDisplay, contractsAbs := resolveContractsDisplayPath(root, cfg.Paths.SpecsDir, slug)
	checkFile(&result, cfg.Project.ConstitutionFile, absFromRoot(root, cfg.Project.ConstitutionFile))
	checkFile(&result, specDisplay, specAbs)
	checkFile(&result, planDisplay, planAbs)
	checkFile(&result, tasksDisplay, tasksAbs)
	checkFile(&result, dataModelDisplay, dataModelAbs)
	checkFile(&result, promptDisplay, absFromRoot(root, promptDisplay))
	if isDir(contractsAbs) {
		result.AddOK(contractsDisplay)
	} else {
		result.AddOK("optional contracts directory not present")
	}
	if fileExists(tasksAbs) {
		content, err := os.ReadFile(tasksAbs)
		if err != nil {
			return CheckResult{}, fmt.Errorf("read tasks %s: %w", tasksDisplay, err)
		}
		checkPattern(&result, string(content), `(?m)^## (Покрытие критериев приемки|Acceptance Coverage)$`, "tasks include acceptance coverage section")
		checkPattern(&result, string(content), `(?m)^## (Implementation Context|Контекст реализации)$`, "tasks include implementation context section")
		checkPattern(&result, string(content), taskIDPattern.String(), "tasks include phase-scoped task IDs")
		checkPattern(&result, string(content), coverageLinePattern.String(), "tasks include AC-to-task coverage lines")
	}
	if fileExists(specAbs) && fileExists(tasksAbs) {
		inspectResult, err := InspectSpec(ctx, cfg, root, specDisplay, tasksDisplay)
		if err != nil {
			return CheckResult{}, err
		}
		result.Merge(inspectResult)
	}
	if fileExists(planAbs) && fileExists(tasksAbs) {
		planContent, err := os.ReadFile(planAbs)
		if err != nil {
			return CheckResult{}, fmt.Errorf("read plan %s: %w", planDisplay, err)
		}
		tasksContent, err := os.ReadFile(tasksAbs)
		if err != nil {
			return CheckResult{}, fmt.Errorf("read tasks %s: %w", tasksDisplay, err)
		}
		checkPlanTaskSurfaceConsistency(&result, planDisplay, string(planContent), tasksDisplay, string(tasksContent))
	}
	if fileExists(planAbs) {
		planContent, err := os.ReadFile(planAbs)
		if err != nil {
			return CheckResult{}, fmt.Errorf("read plan for content check %s: %w", planDisplay, err)
		}
		checkPlanContent(&result, slug, specAbs, planDisplay, string(planContent))
	}
	constitutionResult, err := CheckConstitution(ctx, cfg, root, cfg.Project.ConstitutionFile)
	if err != nil {
		return CheckResult{}, err
	}
	result.Lines = append(result.Lines, constitutionResult.Lines...)
	result.Warnings += constitutionResult.Warnings
	return result, nil
}
