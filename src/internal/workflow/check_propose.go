package workflow

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"speckeep/src/internal/config"
	"speckeep/src/internal/gitutil"
)

// CheckProposeReady validates that a workspace can run the one-shot propose
// lane: constitution + spec/tasks templates + prompts present, and (when a slug
// is given) the expected feature branch.
func CheckProposeReady(ctx context.Context, cfg config.Config, root, slug string) (CheckResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return CheckResult{}, err
	}
	result := CheckResult{}
	constitutionDisplay := cfg.Project.ConstitutionFile
	specTemplateDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.Spec)
	specPromptDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.SpecPrompt)
	tasksTemplateDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.Tasks)
	tasksPromptDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.TasksPrompt)
	checkFile(&result, constitutionDisplay, absFromRoot(root, constitutionDisplay))
	checkFile(&result, specTemplateDisplay, absFromRoot(root, specTemplateDisplay))
	checkFile(&result, specPromptDisplay, absFromRoot(root, specPromptDisplay))
	checkFile(&result, tasksTemplateDisplay, absFromRoot(root, tasksTemplateDisplay))
	checkFile(&result, tasksPromptDisplay, absFromRoot(root, tasksPromptDisplay))
	if strings.TrimSpace(slug) != "" {
		expectedBranch := "feature/" + strings.TrimSpace(slug)
		branch, err := gitutil.CurrentBranch(context.Background(), root)
		if err != nil {
			result.AddWarn("git branch check skipped (git not available)")
		} else if branch == "HEAD" {
			result.AddError(fmt.Sprintf("detached HEAD: switch/create %s before proposing a feature", expectedBranch))
		} else if branch != expectedBranch && branch != "main" && branch != "master" {
			result.AddError(fmt.Sprintf("expected to be on branch %s, got %s (create it or use /spk.spec)", expectedBranch, branch))
		} else {
			result.AddOK(fmt.Sprintf("on branch %s (ok for propose)", branch))
		}
	}
	return result, nil
}
