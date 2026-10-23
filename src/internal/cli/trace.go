package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"speckeep/src/internal/config"
	"speckeep/src/internal/featurepaths"
	"speckeep/src/internal/gitutil"
	"speckeep/src/internal/trace"
)

func newTraceCmd() *cobra.Command {
	var jsonOutput bool
	var testsOnly bool
	var filesMode bool
	var filePath string

	cmd := &cobra.Command{
		Use:   "trace [slug] [path]",
		Short: "Trace tasks to Proof entries and map changed files to features",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, slug := parseTraceArgs(args)

			if strings.TrimSpace(filePath) != "" {
				return runTraceFile(cmd, root, filePath, jsonOutput)
			}
			if filesMode {
				if slug == "" {
					return fmt.Errorf("slug required with --files")
				}
				return runTraceFiles(cmd, root, slug, jsonOutput)
			}

			var entries []trace.Entry
			var err error
			if slug != "" {
				entries, err = trace.ParseTasks(context.Background(), root, slug)
			} else {
				entries, err = trace.ParseAll(context.Background(), root)
			}
			if err != nil {
				return err
			}

			if testsOnly {
				var filtered []trace.Entry
				for _, e := range entries {
					if e.Kind == "test" {
						filtered = append(filtered, e)
					}
				}
				entries = filtered
			}

			if jsonOutput {
				payload, err := json.MarshalIndent(entries, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
				return nil
			}

			if len(entries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No Proof entries found in tasks.md files.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Found %d Proof entries:\n\n", len(entries))
			for _, e := range entries {
				anchorPart := ""
				if e.Anchor != "" {
					anchorPart = ":" + e.Anchor
				}
				acPart := ""
				if e.ACID != "" {
					acPart = fmt.Sprintf(" (%s)", e.ACID)
				}
				slugPart := ""
				if slug == "" {
					slugPart = e.Slug + "#"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "- %s%s%s -> %s%s [%s]%s\n", slugPart, e.TaskID, acPart, e.File, anchorPart, e.Kind, traceProblemsSuffix(cmd, root, e))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	cmd.Flags().BoolVar(&testsOnly, "tests", false, "Show only test-kind Proof entries")
	cmd.Flags().BoolVar(&filesMode, "files", false, "Show files changed by a feature, mapped to tasks")
	cmd.Flags().StringVar(&filePath, "file", "", "Show which specs and tasks changed a file")
	cmd.AddCommand(newTraceInitCmd())
	cmd.AddCommand(newTraceRebuildCmd())

	return cmd
}

// parseTraceArgs accepts the historical [slug] [path] forms without confusing
// a path with a slug.
func parseTraceArgs(args []string) (root, slug string) {
	root = "."
	switch len(args) {
	case 1:
		if _, err := os.Stat(args[0]); err == nil {
			root = args[0]
		} else {
			slug = args[0]
		}
	case 2:
		slug = args[0]
		root = args[1]
	}
	return root, slug
}

func runTraceFiles(cmd *cobra.Command, root, slug string, jsonOutput bool) error {
	ctx := context.Background()
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	cfg, err := config.Load(ctx, root)
	if err != nil {
		return err
	}

	changes, err := trace.BuildChanges(ctx, root, slug)
	if err != nil {
		return err
	}
	specsDir, _ := cfg.SpecsDir(root)
	if _, statErr := os.Stat(featurepaths.SpecDir(specsDir, slug)); statErr != nil {
		archiveDir, archErr := cfg.ArchiveDir(root)
		if archErr == nil {
			sets, archErr := trace.ArchivedChangeSets(archiveDir, slug)
			if archErr == nil && len(sets) > 0 {
				changes = sets[0].Changes
			}
		}
	}

	if jsonOutput {
		payload, err := json.MarshalIndent(changes, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(payload))
		return nil
	}

	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "Slug: %s\n", slug)
	if changes.Source == "git" {
		fmt.Fprintf(w, "Source: git (%s..%s", shortSHA(changes.BaseSHA), shortSHA(changes.HeadSHA))
		if changes.BaseBranch != "" {
			fmt.Fprintf(w, ", base %s", changes.BaseBranch)
		}
		fmt.Fprintf(w, ", %s)\n", changes.BaseSource)
	} else {
		fmt.Fprintf(w, "Source: declared (no usable git base)\n")
	}
	if changes.Worktree {
		fmt.Fprintln(w, "Working tree: includes uncommitted tracked and untracked changes")
	} else if changes.Dirty {
		fmt.Fprintln(w, "Note: the working tree was dirty when this snapshot was captured; uncommitted changes are not included")
	}

	fmt.Fprintln(w, "\nFiles:")
	if len(changes.Files) == 0 {
		fmt.Fprintln(w, "  none")
	}
	for _, file := range changes.Files {
		fmt.Fprintf(w, "  %-10s %s", file.Status, file.Path)
		if len(file.Tasks) > 0 {
			fmt.Fprintf(w, "  [%s]", strings.Join(file.Tasks, ", "))
		}
		if len(file.ACs) > 0 {
			fmt.Fprintf(w, "  AC: %s", strings.Join(file.ACs, ", "))
		}
		if file.Additions > 0 || file.Deletions > 0 {
			fmt.Fprintf(w, "  +%d/-%d", file.Additions, file.Deletions)
		}
		fmt.Fprintln(w)
	}

	if len(changes.Unmapped) > 0 {
		fmt.Fprintln(w, "\nUnmapped (changed without a task):")
		for _, file := range changes.Unmapped {
			fmt.Fprintf(w, "  %s\n", file)
		}
	}
	if len(changes.NotChanged) > 0 {
		fmt.Fprintln(w, "\nDeclared but not changed:")
		for _, file := range changes.NotChanged {
			fmt.Fprintf(w, "  %s\n", file)
		}
	}
	if len(changes.Artifacts) > 0 {
		fmt.Fprintln(w, "\nFeature artifacts (excluded):")
		for _, file := range changes.Artifacts {
			fmt.Fprintf(w, "  %s\n", file)
		}
	}
	return nil
}

// TraceHit is one reverse-lookup result for `trace --file`.
type TraceHit struct {
	Source string   `json:"source"` // active|archived
	Slug   string   `json:"slug"`
	Date   string   `json:"date,omitempty"`
	Path   string   `json:"path"`
	Status string   `json:"status"`
	Tasks  []string `json:"tasks,omitempty"`
	ACs    []string `json:"acs,omitempty"`
}

func runTraceFile(cmd *cobra.Command, root, filePath string, jsonOutput bool) error {
	ctx := context.Background()
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	cfg, err := config.Load(ctx, root)
	if err != nil {
		return err
	}
	rel := filePath
	if filepath.IsAbs(filePath) {
		if value, relErr := filepath.Rel(root, filePath); relErr == nil {
			rel = value
		}
	}
	rel = filepath.ToSlash(filepath.Clean(rel))

	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return err
	}
	slugs, _ := featurepaths.ListSpecSlugs(specsDir)
	hits := make([]TraceHit, 0)
	for _, slug := range slugs {
		specDir := featurepaths.SpecDir(specsDir, slug)
		if _, statErr := os.Stat(specDir); statErr != nil {
			continue
		}
		tasksContent := ""
		if tasksPath, _ := featurepaths.ResolveTasks(specsDir, slug); tasksPath != "" {
			if content, readErr := os.ReadFile(tasksPath); readErr == nil {
				tasksContent = string(content)
			}
		}
		surfaces := trace.ParseTaskSurfaces(tasksContent)
		tasks, acs, _ := trace.MatchSurface(rel, surfaces)

		changes, err := trace.BuildChanges(ctx, root, slug)
		status := ""
		if err == nil {
			for _, file := range changes.Files {
				if file.Path == rel || file.OldPath == rel {
					status = file.Status
					if len(tasks) == 0 {
						tasks = file.Tasks
						acs = file.ACs
					}
					break
				}
			}
		}
		if len(tasks) == 0 && status == "" {
			continue
		}
		if status == "" {
			status = "declared"
		}
		hits = append(hits, TraceHit{
			Source: "active",
			Slug:   slug,
			Path:   rel,
			Status: status,
			Tasks:  tasks,
			ACs:    acs,
		})
	}

	archiveDir, err := cfg.ArchiveDir(root)
	if err == nil {
		sets, archErr := trace.ArchivedChangeSets(archiveDir, "")
		if archErr == nil {
			for _, set := range sets {
				for _, file := range set.Changes.Files {
					if !trace.MatchesRef(rel, file.Path) && file.OldPath != rel {
						continue
					}
					hits = append(hits, TraceHit{
						Source: "archived",
						Slug:   set.Slug,
						Date:   set.Date,
						Path:   rel,
						Status: file.Status,
						Tasks:  file.Tasks,
						ACs:    file.ACs,
					})
				}
			}
		}
	}

	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Slug != hits[j].Slug {
			return hits[i].Slug < hits[j].Slug
		}
		return hits[i].Date < hits[j].Date
	})

	if jsonOutput {
		payload, err := json.MarshalIndent(hits, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(payload))
		return nil
	}

	w := cmd.OutOrStdout()
	if len(hits) == 0 {
		fmt.Fprintf(w, "No spec or task references %s\n", rel)
		return nil
	}
	fmt.Fprintf(w, "References to %s:\n\n", rel)
	for _, hit := range hits {
		location := hit.Source
		if hit.Source == "archived" {
			location = "archived " + hit.Date
		}
		tasks := "-"
		if len(hit.Tasks) > 0 {
			tasks = strings.Join(hit.Tasks, ", ")
		}
		acs := ""
		if len(hit.ACs) > 0 {
			acs = " AC: " + strings.Join(hit.ACs, ", ")
		}
		fmt.Fprintf(w, "- %s#%s (%s) [%s]%s\n", hit.Slug, tasks, location, hit.Status, acs)
	}
	return nil
}

func newTraceInitCmd() *cobra.Command {
	var jsonOutput bool
	var base string
	var force bool

	cmd := &cobra.Command{
		Use:   "init <slug> [path]",
		Short: "Pin the feature base commit for change attribution",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			root := "."
			if len(args) == 2 {
				root = args[1]
			}
			root, err := filepath.Abs(root)
			if err != nil {
				return err
			}
			meta, changed, err := trace.InitMeta(context.Background(), root, slug, base, force)
			if err != nil {
				return err
			}

			if jsonOutput {
				payload, err := json.MarshalIndent(meta, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
				return nil
			}

			w := cmd.OutOrStdout()
			if !changed {
				fmt.Fprintf(w, "Base already pinned for %s: %s\n", slug, shortSHA(meta.BaseSHA))
				return nil
			}
			if meta.BaseSHA == "" {
				fmt.Fprintf(w, "Recorded %s in declared mode (no git base found)\n", slug)
				return nil
			}
			fmt.Fprintf(w, "Pinned base for %s: %s (%s", slug, shortSHA(meta.BaseSHA), meta.BaseSource)
			if meta.BaseBranch != "" {
				fmt.Fprintf(w, ", branch %s", meta.BaseBranch)
			}
			fmt.Fprintln(w, ")")

			if branch, err := gitutil.CurrentBranch(context.Background(), root); err == nil {
				expected := "feature/" + slug
				if branch != expected && branch != "hotfix/"+slug {
					fmt.Fprintf(w, "Warning: current branch is %s, expected %s\n", branch, expected)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	cmd.Flags().StringVar(&base, "base", "", "Base ref to pin (default: merge-base with main/master)")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite an existing base pin")
	return cmd
}

func newTraceRebuildCmd() *cobra.Command {
	var all bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "rebuild [slug] [path]",
		Short: "Backfill changes.json for archived features",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			slug := ""
			if len(args) == 1 {
				if _, err := os.Stat(args[0]); err == nil {
					root = args[0]
				} else {
					slug = args[0]
				}
			} else if len(args) == 2 {
				slug = args[0]
				root = args[1]
			}
			if slug == "" && !all {
				return fmt.Errorf("provide a slug or --all")
			}
			root, err := filepath.Abs(root)
			if err != nil {
				return err
			}
			results, err := rebuildArchives(context.Background(), root, slug)
			if err != nil {
				return err
			}
			if jsonOutput {
				payload, err := json.MarshalIndent(results, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
				return nil
			}
			if len(results) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No archived features need changes.json.")
				return nil
			}
			for _, result := range results {
				fmt.Fprintf(cmd.OutOrStdout(), "- %s (%s): %s\n", result.Slug, result.Date, result.Status)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Rebuild every archived feature")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	return cmd
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	if sha == "" {
		return "-"
	}
	return sha
}

func traceProblemsSuffix(cmd *cobra.Command, root string, e trace.Entry) string {
	if !trace.FileExists(root, e) {
		return " (file missing)"
	}
	if e.Anchor == "" {
		return ""
	}
	for _, problem := range trace.CheckEntry(root, e) {
		if problem.Kind == "anchor-missing" {
			return " (anchor missing)"
		}
	}
	return ""
}
