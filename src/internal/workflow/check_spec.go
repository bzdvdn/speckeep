package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"speckeep/src/internal/config"
	"speckeep/src/internal/gitutil"
)

func CheckSpecReady(ctx context.Context, cfg config.Config, root string) (CheckResult, error) {
	return CheckSpecReadyForSlug(ctx, cfg, root, "")
}

func CheckSpecReadyForSlug(ctx context.Context, cfg config.Config, root, slug string) (CheckResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return CheckResult{}, err
	}
	result := CheckResult{}
	constitutionDisplay := cfg.Project.ConstitutionFile
	templateDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.Spec)
	promptDisplay := joinDisplay(cfg.Paths.TemplatesDir, cfg.Templates.SpecPrompt)
	checkFile(&result, constitutionDisplay, absFromRoot(root, constitutionDisplay))
	checkFile(&result, templateDisplay, absFromRoot(root, templateDisplay))
	checkFile(&result, promptDisplay, absFromRoot(root, promptDisplay))
	templateAbs := absFromRoot(root, templateDisplay)
	if fileExists(templateAbs) {
		content, err := os.ReadFile(templateAbs)
		if err != nil {
			return CheckResult{}, fmt.Errorf("read spec template %s: %w", templateDisplay, err)
		}
		checkPattern(&result, string(content), `(?m)^## (Требования|Requirements)$`, "spec template has requirements section")
		checkPattern(&result, string(content), requirementIDPattern.String(), "spec template includes requirement IDs")
		checkPattern(&result, string(content), acceptanceIDPattern.String(), "spec template includes acceptance IDs")
		checkPattern(&result, string(content), `Given`, "spec template includes Given marker")
		checkPattern(&result, string(content), `When`, "spec template includes When marker")
		checkPattern(&result, string(content), `Then`, "spec template includes Then marker")
	}
	if strings.TrimSpace(slug) != "" {
		expectedBranch := "feature/" + strings.TrimSpace(slug)
		branch, err := gitutil.CurrentBranch(context.Background(), root)
		if err != nil {
			result.AddWarn("git branch check skipped (git not available)")
		} else if branch == "HEAD" {
			result.AddError(fmt.Sprintf("detached HEAD: switch/create %s before writing any spec file", expectedBranch))
		} else if branch != expectedBranch {
			result.AddError(fmt.Sprintf("expected to be on branch %s, got %s", expectedBranch, branch))
		} else {
			result.AddOK(fmt.Sprintf("on expected feature branch %s", expectedBranch))
		}
	}
	return result, nil
}
