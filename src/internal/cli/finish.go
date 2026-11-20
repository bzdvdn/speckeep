package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"speckeep/src/internal/config"
	"speckeep/src/internal/gitutil"
	"speckeep/src/internal/workflow"
)

// FinishResult is the machine-readable outcome of `speckeep finish`.
type FinishResult struct {
	Slug             string   `json:"slug"`
	FeatureBranch    string   `json:"feature_branch"`
	TargetBranch     string   `json:"target_branch"`
	ArchiveDir       string   `json:"archive_dir,omitempty"`
	ArchiveFiles     []string `json:"archive_files,omitempty"`
	Archived         bool     `json:"archived"`
	ArchiveCommitted bool     `json:"archive_committed"`
	Merged           bool     `json:"merged"`
	BranchDeleted    bool     `json:"branch_deleted"`
	DryRun           bool     `json:"dry_run,omitempty"`
	PlannedActions   []string `json:"planned_actions,omitempty"`
	NextCommand      string   `json:"next_command,omitempty"`
}

func newFinishCmd() *cobra.Command {
	var (
		intoBranch    string
		featureBranch string
		keepBranch    bool
		dryRun        bool
		jsonOutput    bool
	)

	cmd := &cobra.Command{
		Use:   "finish <slug> [path]",
		Short: "Archive, merge into the base branch, and delete the local feature branch",
		Long: `Archive a completed feature, merge its branch into the base branch with
--no-ff, and delete the local feature branch.

This is the local-merge close. Use ` + "`speckeep archive`" + ` when merging via a pull
request, or for deferred/abandoned features.

Preconditions:
  - the feature is archive-ready (all [x] tasks carry Proof:, verify policy met)
  - the current branch is the feature/hotfix branch
  - the worktree is clean
  - the target branch exists (--into, default: main, then master)

Never pushes and never deletes a remote branch.`,
		Example: `  speckeep finish my-feature .
  speckeep finish my-feature . --into main
  speckeep finish my-feature . --keep-branch
  speckeep finish my-feature . --dry-run
  speckeep finish my-feature . --json`,
		SilenceUsage: true,
		Args:         cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 2 {
				root = args[1]
			}

			result, err := finishFeature(context.Background(), root, args[0], intoBranch, featureBranch, keepBranch, dryRun)
			if err != nil {
				return err
			}

			if jsonOutput {
				payload, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
				return nil
			}

			printFinishResult(cmd, result)
			return nil
		},
	}

	cmd.Flags().StringVar(&intoBranch, "into", "", "Target branch (default: main, then master)")
	cmd.Flags().StringVar(&featureBranch, "branch", "", "Feature/hotfix branch (default: feature/<slug> or hotfix/<slug>)")
	cmd.Flags().BoolVar(&keepBranch, "keep-branch", false, "Keep the local feature branch after merging")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show the plan and validate preconditions without changing anything")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func finishFeature(ctx context.Context, root, slug, intoBranch, featureBranch string, keepBranch, dryRun bool) (FinishResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return FinishResult{}, err
	}

	cfg, err := config.Load(ctx, root)
	if err != nil {
		return FinishResult{}, err
	}

	state, err := workflow.State(ctx, root, slug)
	if err != nil {
		return FinishResult{}, err
	}

	if !gitutil.IsRepository(ctx, root) {
		return FinishResult{}, fmt.Errorf("speckeep finish requires a git repository")
	}

	if featureBranch == "" {
		if state.HotfixExists && !state.SpecExists {
			featureBranch = "hotfix/" + slug
		} else {
			featureBranch = "feature/" + slug
		}
	}
	exists, err := gitutil.BranchExists(ctx, root, featureBranch)
	if err != nil {
		return FinishResult{}, err
	}
	if !exists {
		return FinishResult{}, fmt.Errorf("feature branch %s does not exist", featureBranch)
	}

	current, err := gitutil.CurrentBranch(ctx, root)
	if err != nil {
		return FinishResult{}, err
	}
	if current != featureBranch {
		return FinishResult{}, fmt.Errorf("current branch is %s; switch to %s before running speckeep finish", current, featureBranch)
	}

	if intoBranch == "" {
		intoBranch, err = gitutil.DefaultTargetBranch(ctx, root)
		if err != nil {
			return FinishResult{}, err
		}
	}
	if intoBranch == featureBranch {
		return FinishResult{}, fmt.Errorf("target branch must differ from the feature branch (%s)", featureBranch)
	}
	exists, err = gitutil.BranchExists(ctx, root, intoBranch)
	if err != nil {
		return FinishResult{}, err
	}
	if !exists {
		return FinishResult{}, fmt.Errorf("target branch %s does not exist", intoBranch)
	}

	clean, err := gitutil.WorktreeClean(ctx, root)
	if err != nil {
		return FinishResult{}, err
	}
	if !clean {
		return FinishResult{}, fmt.Errorf("worktree has uncommitted changes; commit or stash them before running speckeep finish")
	}

	result := FinishResult{
		Slug:          slug,
		FeatureBranch: featureBranch,
		TargetBranch:  intoBranch,
		DryRun:        dryRun,
	}

	needArchive := state.SpecExists || state.HotfixExists
	if !needArchive && !state.Archived {
		return FinishResult{}, fmt.Errorf("no active spec/hotfix and no archive found for %s", slug)
	}

	if needArchive {
		check, err := workflow.ValidateArchiveReady(ctx, cfg, root, slug, "completed", "")
		if err != nil {
			return FinishResult{}, err
		}
		if check.Failed {
			var gaps []string
			for _, line := range check.Lines {
				if strings.HasPrefix(line, "ERROR: ") {
					gaps = append(gaps, strings.TrimPrefix(line, "ERROR: "))
				}
			}
			return FinishResult{}, fmt.Errorf("archive readiness failed: %s", strings.Join(gaps, "; "))
		}
	}

	if dryRun {
		if needArchive {
			result.PlannedActions = append(result.PlannedActions, "archive "+slug+" and commit the archive snapshot")
		}
		result.PlannedActions = append(result.PlannedActions,
			fmt.Sprintf("merge %s into %s with --no-ff", featureBranch, intoBranch))
		if !keepBranch {
			result.PlannedActions = append(result.PlannedActions, "delete local branch "+featureBranch)
		}
		result.NextCommand = "/spk-propose"
		return result, nil
	}

	if needArchive {
		archiveResult, err := archiveFeature(root, slug, "completed", "", false, false)
		if err != nil {
			return result, err
		}
		result.Archived = true
		result.ArchiveDir = archiveResult.ArchiveDir
		result.ArchiveFiles = archiveResult.Files

		clean, err = gitutil.WorktreeClean(ctx, root)
		if err != nil {
			return result, err
		}
		if clean {
			return result, fmt.Errorf("archive produced no tracked changes (is the archive directory ignored?); archive and merge manually")
		}
		if err := gitutil.CommitAll(ctx, root, "chore: archive "+slug); err != nil {
			return result, fmt.Errorf("archive created but commit failed: %w (commit the archive changes manually, then rerun speckeep finish)", err)
		}
		result.ArchiveCommitted = true
	}

	if err := gitutil.Checkout(ctx, root, intoBranch); err != nil {
		return result, err
	}
	if err := gitutil.MergeNoFF(ctx, root, featureBranch); err != nil {
		_ = gitutil.AbortMerge(ctx, root)
		_ = gitutil.Checkout(ctx, root, featureBranch)
		return result, fmt.Errorf("merge failed and was aborted; feature branch was kept: %w", err)
	}
	result.Merged = true

	if !keepBranch {
		if err := gitutil.DeleteBranch(ctx, root, featureBranch); err != nil {
			return result, fmt.Errorf("merged into %s, but could not delete local branch %s: %w", intoBranch, featureBranch, err)
		}
		result.BranchDeleted = true
	}

	result.NextCommand = "/spk-propose"
	return result, nil
}

func printFinishResult(cmd *cobra.Command, result FinishResult) {
	w := cmd.OutOrStdout()

	printPanel(w, "speckeep finish", []string{
		"slug: " + result.Slug,
		"feature_branch: " + result.FeatureBranch,
		"target_branch: " + result.TargetBranch,
	})
	if result.ArchiveDir != "" {
		printPanel(w, "Archive", []string{"archive_dir: " + stylePath(w, result.ArchiveDir)})
	}
	if result.DryRun {
		fmt.Fprintln(w)
		for _, action := range result.PlannedActions {
			fmt.Fprintln(w, "- would "+action)
		}
		fmt.Fprintf(w, "\n%s Dry run: no changes made\n", styleOK(w, "✓"))
		return
	}

	parts := []string{}
	if result.ArchiveCommitted {
		parts = append(parts, "archived")
	}
	if result.Merged {
		parts = append(parts, "merged into "+result.TargetBranch)
	}
	if result.BranchDeleted {
		parts = append(parts, "deleted "+result.FeatureBranch)
	}
	fmt.Fprintf(w, "\n%s Finished %s: %s\n", styleOK(w, "✓"), result.Slug, strings.Join(parts, "; "))
	if result.NextCommand != "" {
		fmt.Fprintf(w, "Next: %s\n", styleCmd(w, result.NextCommand))
	}
}
