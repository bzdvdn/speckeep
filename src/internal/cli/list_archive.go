package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"speckeep/src/internal/config"
)

type archiveEntry struct {
	Slug       string `json:"slug"`
	Status     string `json:"status"`
	Reason     string `json:"reason,omitempty"`
	ArchivedAt string `json:"archived_at"`
	Snapshot   string `json:"snapshot"`
}

type archiveSummaryFrontmatter struct {
	ReportType string `yaml:"report_type"`
	Slug       string `yaml:"slug"`
	Status     string `yaml:"status"`
	Reason     string `yaml:"reason"`
	ArchivedAt string `yaml:"archived_at"`
}

func newListArchiveCmd() *cobra.Command {
	var (
		filterStatus string
		since        string
		jsonOutput   bool
	)

	cmd := &cobra.Command{
		Use:   "list-archive [path]",
		Short: "List archived features",
		Long: `List archived features from .speckeep/archive/.

Shows one entry per slug (most recent snapshot). Use --status to filter
by archive status, and --since to filter by date.`,
		Example: `  speckeep list-archive .
  speckeep list-archive . --status deferred
  speckeep list-archive . --since 2026-01-01
  speckeep list-archive . --json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}

			var sinceTime time.Time
			if since != "" {
				t, err := time.Parse("2006-01-02", since)
				if err != nil {
					return fmt.Errorf("invalid --since date %q: use YYYY-MM-DD format", since)
				}
				sinceTime = t
			}

			entries, err := loadArchiveEntries(root, filterStatus, sinceTime)
			if err != nil {
				return err
			}

			if jsonOutput {
				payload, err := json.MarshalIndent(entries, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
				return nil
			}

			printArchiveList(cmd, entries, filterStatus, since)
			return nil
		},
	}

	cmd.Flags().StringVar(&filterStatus, "status", "", "Filter by status: completed, superseded, abandoned, rejected, deferred")
	cmd.Flags().StringVar(&since, "since", "", "Filter by archive date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func loadArchiveEntries(root, filterStatus string, since time.Time) ([]archiveEntry, error) {
	cfg, err := config.Load(root)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	archiveDir, err := cfg.ArchiveDir(root)
	if err != nil {
		return nil, err
	}

	slugDirs, err := os.ReadDir(archiveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read archive dir: %w", err)
	}

	var entries []archiveEntry

	for _, slugDir := range slugDirs {
		if !slugDir.IsDir() {
			continue
		}
		slug := slugDir.Name()

		snapshot, entry, ok := latestArchiveEntry(archiveDir, slug)
		if !ok {
			continue
		}
		entry.Snapshot = snapshot

		if filterStatus != "" && entry.Status != filterStatus {
			continue
		}

		if !since.IsZero() {
			entryDate, err := time.Parse("2006-01-02", snapshot)
			if err == nil && entryDate.Before(since) {
				continue
			}
		}

		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Snapshot > entries[j].Snapshot
	})

	return entries, nil
}

func latestArchiveEntry(archiveDir, slug string) (string, archiveEntry, bool) {
	slugDir := filepath.Join(archiveDir, slug)
	snapshots, err := os.ReadDir(slugDir)
	if err != nil {
		return "", archiveEntry{}, false
	}

	var latest string
	for _, s := range snapshots {
		if s.IsDir() && s.Name() > latest {
			latest = s.Name()
		}
	}
	if latest == "" {
		return "", archiveEntry{}, false
	}

	summaryPath := filepath.Join(slugDir, latest, "summary.md")
	entry, ok := parseArchiveSummary(summaryPath, slug, latest)
	return latest, entry, ok
}

func parseArchiveSummary(path, slug, snapshot string) (archiveEntry, bool) {
	content, err := os.ReadFile(path)
	if err != nil {
		return archiveEntry{Slug: slug, Snapshot: snapshot}, true
	}

	raw := string(content)
	if !strings.HasPrefix(raw, "---\n") {
		return archiveEntry{Slug: slug, Snapshot: snapshot}, true
	}

	rest := strings.TrimPrefix(raw, "---\n")
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return archiveEntry{Slug: slug, Snapshot: snapshot}, true
	}

	var fm archiveSummaryFrontmatter
	if err := yaml.Unmarshal([]byte(rest[:end]), &fm); err != nil {
		return archiveEntry{Slug: slug, Snapshot: snapshot}, true
	}

	return archiveEntry{
		Slug:       slug,
		Status:     fm.Status,
		Reason:     fm.Reason,
		ArchivedAt: snapshot,
	}, true
}

func printArchiveList(cmd *cobra.Command, entries []archiveEntry, filterStatus, since string) {
	w := cmd.OutOrStdout()

	if len(entries) == 0 {
		label := "archive"
		if filterStatus != "" {
			label = filterStatus + " archive"
		}
		fmt.Fprintf(w, "No entries in %s.\n", label)
		return
	}

	// Column widths
	slugW, statusW, dateW, reasonW := 4, 6, 11, 6
	for _, e := range entries {
		if l := utf8.RuneCountInString(e.Slug); l > slugW {
			slugW = l
		}
		if l := utf8.RuneCountInString(e.Status); l > statusW {
			statusW = l
		}
		if l := utf8.RuneCountInString(e.ArchivedAt); l > dateW {
			dateW = l
		}
		if l := utf8.RuneCountInString(e.Reason); l > reasonW {
			reasonW = l
		}
	}

	header := fmt.Sprintf("%-*s  %-*s  %-*s  %s",
		slugW, "SLUG",
		statusW, "STATUS",
		dateW, "ARCHIVED_AT",
		"REASON",
	)
	sep := strings.Repeat("─", slugW+statusW+dateW+reasonW+8)

	fmt.Fprintln(w, styleMuted(w, header))
	fmt.Fprintln(w, styleMuted(w, sep))

	for _, e := range entries {
		statusStyled := e.Status
		switch e.Status {
		case "completed":
			statusStyled = styleOK(w, e.Status)
		case "deferred":
			statusStyled = styleWarn(w, e.Status)
		case "abandoned", "rejected":
			statusStyled = styleError(w, e.Status)
		}

		fmt.Fprintf(w, "%-*s  %s%s  %-*s  %s\n",
			slugW, e.Slug,
			statusStyled,
			strings.Repeat(" ", statusW-utf8.RuneCountInString(e.Status)),
			dateW, e.ArchivedAt,
			styleMuted(w, e.Reason),
		)
	}

	fmt.Fprintf(w, "\n%d archived feature(s)", len(entries))
	if filterStatus != "" {
		fmt.Fprintf(w, " [status: %s]", filterStatus)
	}
	if since != "" {
		fmt.Fprintf(w, " [since: %s]", since)
	}
	fmt.Fprintln(w)
}
