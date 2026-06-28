package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"speckeep/src/internal/config"
)

func CheckArchiveReady(ctx context.Context, cfg config.Config, root, slug, status, reason string) (CheckResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return CheckResult{}, err
	}
	result := CheckResult{}
	if strings.TrimSpace(status) == "" {
		result.AddError("archive status is required")
		return result, nil
	}
	switch status {
	case "completed", "superseded", "abandoned", "rejected", "deferred":
	default:
		result.AddError(fmt.Sprintf("invalid archive status: %s", status))
		return result, nil
	}
	if status != "completed" && strings.TrimSpace(reason) == "" {
		result.AddError("archive reason is required for non-completed statuses")
		return result, nil
	}
	specDisplay, specAbs := resolveSpecDisplayPath(root, cfg.Paths.SpecsDir, slug)
	hotfixDisplay, hotfixAbs := resolveHotfixDisplayPath(root, cfg.Paths.SpecsDir, slug)
	if !fileExists(specAbs) && !fileExists(hotfixAbs) {
		result.AddError(fmt.Sprintf("missing required file: %s (or %s)", specDisplay, hotfixDisplay))
		return result, nil
	}
	_, tasksAbs := resolveTasksDisplayPath(root, cfg.Paths.SpecsDir, slug)
	state, err := State(ctx, root, slug)
	if err != nil {
		return CheckResult{}, err
	}
	if !state.VerifyExists {
		result.AddError(ErrVerifyMissing.Error())
	} else if state.VerifyStatus != StatusPass {
		result.AddError(fmt.Sprintf("verify status is %s - fix before archiving", state.VerifyStatus))
	}
	if fileExists(tasksAbs) {
		taskStateResult, summary, err := VerifyTaskState(ctx, cfg, root, slug)
		if err != nil {
			return CheckResult{}, err
		}
		result.Merge(taskStateResult)
		if summary.Open > 0 {
			result.AddError("open tasks remain - complete before archiving")
		}
	}
	if result.Failed {
		return result, nil
	}
	archiveDisplay := joinDisplay(cfg.Paths.ArchiveDir, slug)
	if err := os.MkdirAll(absFromRoot(root, archiveDisplay), 0o755); err != nil {
		return CheckResult{}, fmt.Errorf("create archive directory %s: %w", archiveDisplay, err)
	}
	result.AddOK(fmt.Sprintf("archive is ready for slug '%s' with status '%s'", slug, status))
	return result, nil
}
