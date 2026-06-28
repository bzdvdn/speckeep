package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"speckeep/src/internal/config"
)

func CheckVerifyReady(ctx context.Context, cfg config.Config, root, slug string) (CheckResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return CheckResult{}, err
	}
	result := CheckResult{}
	specDisplay, specAbs := resolveSpecDisplayPath(root, cfg.Paths.SpecsDir, slug)
	tasksDisplay, tasksAbs := resolveTasksDisplayPath(root, cfg.Paths.SpecsDir, slug)
	reportTemplateDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.VerifyReport)
	promptDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.VerifyPrompt)
	checkFile(&result, cfg.Project.ConstitutionFile, absFromRoot(root, cfg.Project.ConstitutionFile))
	checkFile(&result, specDisplay, specAbs)
	checkFile(&result, tasksDisplay, tasksAbs)
	checkFile(&result, reportTemplateDisplay, absFromRoot(root, reportTemplateDisplay))
	checkFile(&result, promptDisplay, absFromRoot(root, promptDisplay))
	if fileExists(specAbs) && fileExists(tasksAbs) {
		inspectResult, err := InspectSpec(ctx, cfg, root, specDisplay, tasksDisplay)
		if err != nil {
			return CheckResult{}, err
		}
		result.Merge(inspectResult)
		if !result.Failed && inspectResult.Failed {
			result.AddWarn("spec inspection failed — address findings before verify")
		}
	}
	if !result.Failed {
		content, err := os.ReadFile(tasksAbs)
		if err != nil {
			return CheckResult{}, fmt.Errorf("read tasks %s: %w", tasksDisplay, err)
		}
		if hasOpenTasks(content) {
			result.AddError("one or more tasks remain open")
		}
		taskStateResult, _, err := VerifyTaskState(ctx, cfg, root, slug)
		if err != nil {
			return CheckResult{}, err
		}
		result.Merge(taskStateResult)
		if tasksContent, readErr := os.ReadFile(tasksAbs); readErr == nil {
			checkTouchesFilesExist(&result, root, tasksDisplay, string(tasksContent))
		}
	}
	planDisplay, planAbs := resolvePlanDisplayPath(root, cfg.Paths.SpecsDir, slug)
	if fileExists(planAbs) {
		if planContent, readErr := os.ReadFile(planAbs); readErr == nil {
			checkPlanContent(&result, slug, specAbs, planDisplay, string(planContent))
		}
	}
	return result, nil
}

func VerifyTaskState(ctx context.Context, cfg config.Config, root, slug string) (CheckResult, TaskStateSummary, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return CheckResult{}, TaskStateSummary{}, err
	}
	tasksDisplay, tasksAbs := resolveTasksDisplayPath(root, cfg.Paths.SpecsDir, slug)
	result := CheckResult{}
	if !fileExists(tasksAbs) {
		result.AddError(fmt.Sprintf("missing %s", tasksDisplay))
		return result, TaskStateSummary{}, nil
	}
	summary, err := computeTaskState(tasksAbs)
	if err != nil {
		return CheckResult{}, TaskStateSummary{}, err
	}
	result.AddRaw(fmt.Sprintf("TASKS_TOTAL=%d", summary.Total))
	result.AddRaw(fmt.Sprintf("TASKS_COMPLETED=%d", summary.Completed))
	result.AddRaw(fmt.Sprintf("TASKS_OPEN=%d", summary.Open))
	result.AddRaw(fmt.Sprintf("TASK_IDS=%d", summary.TaskIDs))
	result.AddRaw(fmt.Sprintf("AC_COVERAGE_LINES=%d", summary.CoverageLines))
	if summary.Total == 0 {
		result.AddError(fmt.Sprintf("no task checkboxes found in %s", tasksDisplay))
		return result, summary, nil
	}
	if summary.TaskIDs == 0 {
		result.AddError(fmt.Sprintf("no stable task IDs found in %s", tasksDisplay))
		return result, summary, nil
	}
	if summary.CoverageLines == 0 {
		result.AddWarn(fmt.Sprintf("no AC-to-task coverage lines found in %s", tasksDisplay))
	}
	if summary.Open > 0 {
		result.AddWarn(fmt.Sprintf("open tasks remain in %s", tasksDisplay))
		return result, summary, nil
	}
	result.AddOK(fmt.Sprintf("all tasks are marked complete in %s", tasksDisplay))
	return result, summary, nil
}

func hasOpenTasks(content []byte) bool {
	return strings.Contains(string(content), "- [ ]")
}

func computeTaskState(tasksPath string) (TaskStateSummary, error) {
	content, err := os.ReadFile(tasksPath)
	if err != nil {
		return TaskStateSummary{}, fmt.Errorf("read tasks file: %w", err)
	}
	text := string(content)
	total, completed, open, err := taskCounts(tasksPath)
	if err != nil {
		return TaskStateSummary{}, err
	}
	return TaskStateSummary{
		Total:         total,
		Completed:     completed,
		Open:          open,
		TaskIDs:       len(taskIDPattern.FindAllString(text, -1)),
		CoverageLines: len(coverageLinePattern.FindAllString(text, -1)),
	}, nil
}
