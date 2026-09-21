package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"speckeep/src/internal/config"
)

// CheckConvergeReady is the cheap loop gate behind /spk-converge: it re-checks
// a feature whose tasks are claimed complete and turns the gaps into a small
// set of structured findings. The agent (+ repository) fixes those gaps and
// re-runs converge until it is clean — or stops with a concrete reason.
//
// Unlike CheckVerifyReady it does not require the verify report template or
// prompt, and it treats plan.md / data-model.md as optional (express mode).
func CheckConvergeReady(ctx context.Context, cfg config.Config, root, slug string) (CheckResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return CheckResult{}, err
	}
	result := CheckResult{}
	specDisplay, specAbs := resolveSpecDisplayPath(root, cfg.Paths.SpecsDir, slug)
	tasksDisplay, tasksAbs := resolveTasksDisplayPath(root, cfg.Paths.SpecsDir, slug)
	planDisplay, planAbs := resolvePlanDisplayPath(root, cfg.Paths.SpecsDir, slug)
	checkFile(&result, cfg.Project.ConstitutionFile, absFromRoot(root, cfg.Project.ConstitutionFile))
	checkFile(&result, specDisplay, specAbs)
	checkFile(&result, tasksDisplay, tasksAbs)
	if fileExists(planAbs) {
		result.AddOK("plan: " + planDisplay)
	} else {
		result.AddOK("plan: absent (express mode — spec+tasks)")
	}
	if result.Failed {
		result.AddRaw("CONVERGED=0")
		return result, nil
	}

	content, err := os.ReadFile(tasksAbs)
	if err != nil {
		return CheckResult{}, fmt.Errorf("read tasks %s: %w", tasksDisplay, err)
	}
	if hasOpenTasks(content) {
		result.AddError("one or more tasks remain open — implement them or split into follow-up tasks")
	}

	taskStateResult, summary, err := VerifyTaskState(ctx, cfg, root, slug)
	if err != nil {
		return CheckResult{}, err
	}
	result.Merge(taskStateResult)

	proofResult, err := CheckProofs(ctx, cfg, root, slug)
	if err != nil {
		return CheckResult{}, err
	}
	result.Merge(proofResult)

	checkTouchesFilesExist(&result, root, tasksDisplay, string(content))

	if fileExists(specAbs) {
		inspectResult, err := InspectSpec(ctx, cfg, root, specDisplay, tasksDisplay)
		if err != nil {
			return CheckResult{}, err
		}
		result.Merge(inspectResult)
	}
	if fileExists(planAbs) {
		if planContent, readErr := os.ReadFile(planAbs); readErr == nil {
			checkPlanContent(&result, slug, specAbs, planDisplay, string(planContent))
		}
	}

	converged := !result.Failed && summary.Open == 0
	if converged {
		result.AddOK("feature is converged: tasks complete with valid Proof coverage — safe to archive")
		result.AddRaw("CONVERGED=1")
	} else {
		result.AddRaw("CONVERGED=0")
		result.AddRaw("NEXT=append follow-up tasks for the gaps above, then re-run converge")
	}
	return result, nil
}
