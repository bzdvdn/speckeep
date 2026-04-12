package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"speckeep/src/internal/config"
	"speckeep/src/internal/featurepaths"
	"speckeep/src/internal/workflow"
)

type ArchiveResult struct {
	Slug       string   `json:"slug"`
	Status     string   `json:"status"`
	Reason     string   `json:"reason"`
	ArchivedAt string   `json:"archived_at"`
	ArchiveDir string   `json:"archive_dir"`
	Mode       string   `json:"mode"`
	Files      []string `json:"files,omitempty"`
	Restored   []string `json:"restored,omitempty"`
	Error      string   `json:"error,omitempty"`
}

func newArchiveCmd() *cobra.Command {
	var (
		status     string
		reason     string
		copyMode   bool
		restore    bool
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "archive <slug> [path]",
		Short: "Archive or restore a feature package",
		Long: `Archive a completed feature to .speckeep/archive/ or restore from archive.

Archive mode (default):
  - Copies all feature artifacts to .speckeep/archive/<slug>/<YYYY-MM-DD>/
  - Generates summary.md from verify.md
  - Removes active files (unless --copy)

Restore mode (--restore):
  - Copies latest archive snapshot back to active specs/ and plans/
  - Removes the archive entry after successful restore`,
		Example: `  speckeep archive my-feature .
  speckeep archive my-feature . --status completed --reason "AC covered"
  speckeep archive my-feature . --copy --status deferred
  speckeep archive my-feature . --restore`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 2 {
				root = args[1]
			}
			slug := args[0]

			if restore {
				result, err := restoreFeature(root, slug)
				if err != nil {
					return err
				}
				return outputArchiveResult(cmd, result, jsonOutput)
			}

			result, err := archiveFeature(root, slug, status, reason, copyMode)
			if err != nil {
				return err
			}
			return outputArchiveResult(cmd, result, jsonOutput)
		},
	}

	cmd.Flags().StringVar(&status, "status", "completed", "Archive status: completed, superseded, abandoned, rejected, deferred")
	cmd.Flags().StringVar(&reason, "reason", "", "Reason for archiving")
	cmd.Flags().BoolVar(&copyMode, "copy", false, "Keep originals after archiving (copy-only mode)")
	cmd.Flags().BoolVar(&restore, "restore", false, "Restore from archive instead of archiving")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func archiveFeature(root, slug, status, reason string, copyMode bool) (ArchiveResult, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	reason = strings.TrimSpace(reason)

	result := ArchiveResult{
		Slug:   slug,
		Status: status,
		Reason: reason,
		Mode:   "move",
	}
	if copyMode {
		result.Mode = "copy"
	}

	switch status {
	case "completed", "superseded", "abandoned", "rejected", "deferred":
	default:
		return result, fmt.Errorf("invalid archive status: %s", status)
	}
	if status != "completed" && reason == "" {
		return result, fmt.Errorf("archive reason is required for non-completed statuses")
	}

	cfg, err := config.Load(root)
	if err != nil {
		return result, fmt.Errorf("load config: %w", err)
	}

	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return result, err
	}
	archiveDir, err := cfg.ArchiveDir(root)
	if err != nil {
		return result, err
	}

	// Check preconditions
	state, err := workflow.State(root, slug)
	if err != nil {
		return result, fmt.Errorf("get feature state: %w", err)
	}

	if !state.SpecExists {
		return result, fmt.Errorf("no spec found for %s", slug)
	}

	if !state.VerifyExists {
		return result, fmt.Errorf("verify.md not found - run verify before archiving")
	}

	if state.VerifyStatus != workflow.StatusPass && state.VerifyStatus != workflow.StatusConcerns {
		return result, fmt.Errorf("verify status is %s - fix before archiving", state.VerifyStatus)
	}

	if state.TasksOpen > 0 {
		return result, fmt.Errorf("%d tasks still open - complete before archiving", state.TasksOpen)
	}

	// Create archive directory
	dateStr := time.Now().Format("2006-01-02")
	slugArchiveDir := filepath.Join(archiveDir, slug, dateStr)
	if err := os.MkdirAll(slugArchiveDir, 0755); err != nil {
		return result, fmt.Errorf("create archive dir: %w", err)
	}

	// Copy spec artifacts
	specFiles := []string{
		featurepaths.Spec(specsDir, slug),
		featurepaths.Inspect(specsDir, slug),
		featurepaths.Summary(specsDir, slug),
		featurepaths.Hotfix(specsDir, slug),
	}

	specArchiveDir := filepath.Join(slugArchiveDir, "specs", slug)
	if err := os.MkdirAll(specArchiveDir, 0755); err != nil {
		return result, err
	}

	for _, src := range specFiles {
		if _, err := os.Stat(src); err == nil {
			dst := filepath.Join(specArchiveDir, filepath.Base(src))
			if err := copyFile(src, dst); err != nil {
				return result, fmt.Errorf("copy %s: %w", src, err)
			}
			result.Files = append(result.Files, "specs/"+slug+"/"+filepath.Base(src))
		}
	}

	// Copy plan artifacts
	planFiles := []string{
		"plan.md",
		"tasks.md",
		"data-model.md",
		"research.md",
		"verify.md",
	}

	planSourceDir := featurepaths.PlanDir(specsDir, slug)
	planArchiveDir := filepath.Join(slugArchiveDir, "plan")

	for _, name := range planFiles {
		src := filepath.Join(planSourceDir, name)
		if _, err := os.Stat(src); err == nil {
			if err := os.MkdirAll(planArchiveDir, 0755); err != nil {
				return result, err
			}
			dst := filepath.Join(planArchiveDir, name)
			if err := copyFile(src, dst); err != nil {
				return result, fmt.Errorf("copy %s: %w", src, err)
			}
			result.Files = append(result.Files, "plan/"+name)
		}
	}

	// Copy contracts if exist
	contractsSourceDir := featurepaths.ContractsDir(specsDir, slug)
	contractsArchiveDir := filepath.Join(planArchiveDir, "contracts")
	if entries, err := os.ReadDir(contractsSourceDir); err == nil && len(entries) > 0 {
		if err := os.MkdirAll(contractsArchiveDir, 0755); err != nil {
			return result, err
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				src := filepath.Join(contractsSourceDir, entry.Name())
				dst := filepath.Join(contractsArchiveDir, entry.Name())
				if err := copyFile(src, dst); err != nil {
					return result, fmt.Errorf("copy contracts/%s: %w", entry.Name(), err)
				}
				result.Files = append(result.Files, "plan/contracts/"+entry.Name())
			}
		}
	}

	// Generate summary.md
	summaryPath := filepath.Join(slugArchiveDir, "summary.md")
	if err := generateArchiveSummary(summaryPath, slug, status, reason, state, result.Files, cfg); err != nil {
		return result, fmt.Errorf("generate summary: %w", err)
	}
	result.Files = append([]string{"summary.md"}, result.Files...)

	// Remove active files if not copy mode
	if !copyMode {
		// Remove spec artifacts
		for _, src := range specFiles {
			if _, err := os.Stat(src); err == nil {
				if err := os.Remove(src); err != nil {
					return result, fmt.Errorf("remove %s: %w", src, err)
				}
			}
		}
		// Remove entire feature dir (specs/<slug>/) — includes plan/ subdir
		featureDir := featurepaths.SpecDir(specsDir, slug)
		if err := os.RemoveAll(featureDir); err != nil {
			return result, fmt.Errorf("remove feature dir: %w", err)
		}
	}

	result.ArchivedAt = dateStr
	result.ArchiveDir = slugArchiveDir

	return result, nil
}

func restoreFeature(root, slug string) (ArchiveResult, error) {
	result := ArchiveResult{
		Slug: slug,
		Mode: "restore",
	}

	cfg, err := config.Load(root)
	if err != nil {
		return result, fmt.Errorf("load config: %w", err)
	}

	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return result, err
	}
	archiveDir, err := cfg.ArchiveDir(root)
	if err != nil {
		return result, err
	}

	slugArchiveDir := filepath.Join(archiveDir, slug)
	entries, err := os.ReadDir(slugArchiveDir)
	if err != nil {
		return result, fmt.Errorf("read archive dir: %w", err)
	}
	if len(entries) == 0 {
		return result, fmt.Errorf("no archive snapshots found for %s", slug)
	}

	// Find most recent snapshot
	var latestSnapshot string
	for _, entry := range entries {
		if entry.IsDir() {
			if latestSnapshot == "" || entry.Name() > latestSnapshot {
				latestSnapshot = entry.Name()
			}
		}
	}
	if latestSnapshot == "" {
		return result, fmt.Errorf("no dated snapshot found in archive")
	}

	snapshotDir := filepath.Join(slugArchiveDir, latestSnapshot)

	// Check for existing active files
	specPath := featurepaths.Spec(specsDir, slug)
	planDir := featurepaths.PlanDir(specsDir, slug)
	if _, err := os.Stat(specPath); err == nil {
		return result, fmt.Errorf("active spec already exists for %s - delete first or use different slug", slug)
	}
	if _, err := os.Stat(planDir); err == nil {
		return result, fmt.Errorf("active plan already exists for %s - delete first", slug)
	}

	// Restore specs
	specArchiveDir := filepath.Join(snapshotDir, "specs", slug)
	if entries, err := os.ReadDir(specArchiveDir); err == nil {
		specTargetDir := filepath.Dir(specPath)
		if err := os.MkdirAll(specTargetDir, 0755); err != nil {
			return result, err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			src := filepath.Join(specArchiveDir, entry.Name())
			dst := filepath.Join(specTargetDir, entry.Name())
			if err := copyFile(src, dst); err != nil {
				return result, fmt.Errorf("restore spec/%s: %w", entry.Name(), err)
			}
			result.Restored = append(result.Restored, "specs/"+slug+"/"+entry.Name())
		}
	}

	// Restore plan
	planArchiveDir := filepath.Join(snapshotDir, "plan")
	if entries, err := os.ReadDir(planArchiveDir); err == nil {
		if err := os.MkdirAll(planDir, 0755); err != nil {
			return result, err
		}
		for _, entry := range entries {
			src := filepath.Join(planArchiveDir, entry.Name())
			dst := filepath.Join(planDir, entry.Name())
			if entry.IsDir() {
				if entry.Name() == "contracts" {
					contractsSrc := filepath.Join(planArchiveDir, "contracts")
					contractsDst := filepath.Join(planDir, "contracts")
					if err := os.MkdirAll(contractsDst, 0755); err != nil {
						return result, err
					}
					contractFiles, _ := os.ReadDir(contractsSrc)
					for _, cf := range contractFiles {
						if !cf.IsDir() {
							cfs := filepath.Join(contractsSrc, cf.Name())
							cfd := filepath.Join(contractsDst, cf.Name())
							if err := copyFile(cfs, cfd); err != nil {
								return result, fmt.Errorf("restore contracts/%s: %w", cf.Name(), err)
							}
							result.Restored = append(result.Restored, "specs/"+slug+"/plan/contracts/"+cf.Name())
						}
					}
				}
				continue
			}
			if err := copyFile(src, dst); err != nil {
				return result, fmt.Errorf("restore plan/%s: %w", entry.Name(), err)
			}
			result.Restored = append(result.Restored, "specs/"+slug+"/plan/"+entry.Name())
		}
	}

	// Remove archive snapshot after successful restore
	if err := os.RemoveAll(snapshotDir); err != nil {
		return result, fmt.Errorf("remove archive snapshot: %w", err)
	}

	// Remove slug archive dir if empty
	remaining, _ := os.ReadDir(slugArchiveDir)
	if len(remaining) == 0 {
		os.Remove(slugArchiveDir)
	}

	result.ArchivedAt = latestSnapshot

	return result, nil
}

func generateArchiveSummary(path, slug, status, reason string, state workflow.FeatureState, files []string, cfg config.Config) error {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString("report_type: archive_summary\n")
	sb.WriteString(fmt.Sprintf("slug: %s\n", slug))
	sb.WriteString(fmt.Sprintf("status: %s\n", status))
	if reason != "" {
		sb.WriteString(fmt.Sprintf("reason: %s\n", reason))
	}
	sb.WriteString(fmt.Sprintf("docs_language: %s\n", cfg.Language.Docs))
	sb.WriteString(fmt.Sprintf("archived_at: %s\n", time.Now().Format("2006-01-02T15:04:05Z")))
	sb.WriteString("---\n\n")

	sb.WriteString(fmt.Sprintf("# Archive Summary: %s\n\n", slug))

	sb.WriteString("## Status\n\n")
	sb.WriteString(fmt.Sprintf("- status: %s\n", status))
	if reason != "" {
		sb.WriteString(fmt.Sprintf("- reason: %s\n", reason))
	}
	sb.WriteString(fmt.Sprintf("- tasks: %d/%d completed\n", state.TasksCompleted, state.TasksTotal))
	sb.WriteString(fmt.Sprintf("- verify: %s\n", state.VerifyStatus))
	sb.WriteString("\n")

	sb.WriteString("## Snapshot\n\n")
	sb.WriteString(fmt.Sprintf("- path: `%s`\n", filepath.Dir(path)))
	sb.WriteString("- mode: move-based (active files removed after archive)\n")
	sb.WriteString("\n")

	sb.WriteString("## Contents\n\n")
	for _, f := range files {
		sb.WriteString(fmt.Sprintf("- %s\n", f))
	}
	sb.WriteString("\n")

	sb.WriteString("## Evidence\n\n")
	sb.WriteString("See verify.md for detailed verification evidence.\n")
	sb.WriteString("\n")

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0644)
}

func outputArchiveResult(cmd *cobra.Command, result ArchiveResult, jsonOutput bool) error {
	if jsonOutput {
		payload, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(payload))
		return nil
	}

	w := cmd.OutOrStdout()

	if result.Mode == "restore" {
		printPanel(w, "speckeep archive --restore", []string{
			"slug: " + result.Slug,
			"restored_from: " + result.ArchivedAt,
		})
		if len(result.Restored) > 0 {
			printPanel(w, "Restored Files", result.Restored)
		}
		fmt.Fprintf(w, "\n%s Ready for: /speckeep.inspect %s\n", styleOK(w, "✓"), result.Slug)
	} else {
		printPanel(w, "speckeep archive", []string{
			"slug: " + result.Slug,
			"status: " + result.Status,
			"mode: " + result.Mode,
			"archive_dir: " + stylePath(w, result.ArchiveDir),
		})
		if result.Reason != "" {
			printPanel(w, "Reason", []string{result.Reason})
		}
		if len(result.Files) > 0 {
			printPanel(w, "Archived Files", result.Files)
		}
		fmt.Fprintf(w, "\n%s Feature archived successfully\n", styleOK(w, "✓"))
	}

	return nil
}
