package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"speckeep/src/internal/config"
	"speckeep/src/internal/gitutil"
	"speckeep/src/internal/trace"
)

type rebuildResult struct {
	Slug   string `json:"slug"`
	Date   string `json:"date"`
	Status string `json:"status"`
}

// rebuildArchives backfills changes.json for archived snapshots that predate
// the change-attribution feature. It is best-effort: compact snapshots carry a
// commit pointer, move-based snapshots need an archived trace-meta.json.
func rebuildArchives(ctx context.Context, root, slugFilter string) ([]rebuildResult, error) {
	cfg, err := config.Load(ctx, root)
	if err != nil {
		return nil, err
	}
	archiveDir, err := cfg.ArchiveDir(root)
	if err != nil {
		return nil, err
	}

	slugs := []string{slugFilter}
	if slugFilter == "" {
		entries, err := os.ReadDir(archiveDir)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			return nil, err
		}
		slugs = slugs[:0]
		for _, entry := range entries {
			if entry.IsDir() {
				slugs = append(slugs, entry.Name())
			}
		}
	}

	var results []rebuildResult
	for _, slug := range slugs {
		slugDir := filepath.Join(archiveDir, slug)
		dates, err := os.ReadDir(slugDir)
		if err != nil {
			continue
		}
		for _, date := range dates {
			if !date.IsDir() {
				continue
			}
			dateDir := filepath.Join(slugDir, date.Name())
			changesPath := filepath.Join(dateDir, "changes.json")
			if _, err := os.Stat(changesPath); err == nil {
				continue
			}
			status := rebuildSnapshot(ctx, root, cfg.Paths.SpecsDir, slug, dateDir)
			results = append(results, rebuildResult{Slug: slug, Date: date.Name(), Status: status})
		}
	}
	return results, nil
}

func rebuildSnapshot(ctx context.Context, root, specsDirRel, slug, dateDir string) string {
	branch, headSHA := readSnapshotPointer(dateDir)
	archiveMeta, tasksContent := archivedInputs(ctx, root, specsDirRel, slug, dateDir, headSHA)
	meta := archiveMeta.Meta
	if branch == "" {
		branch = archiveMeta.Branch
	}
	if headSHA == "" {
		headSHA = archiveMeta.HeadSHA
	}

	if meta.Version == 0 {
		meta = trace.Meta{Version: 1, Slug: slug, CreatedAt: ""}
	}
	if meta.Slug == "" {
		meta.Slug = slug
	}
	if meta.BaseSHA == "" && headSHA != "" {
		meta.BaseSHA = inferBaseFor(ctx, root, headSHA)
		meta.BaseSource = "inferred"
	}
	if headSHA == "" && meta.BaseSHA == "" {
		return "skipped: no commit pointer and no archived base"
	}
	if meta.BaseBranch == "" && branch != "" {
		meta.BaseBranch = branch
	}
	if headSHA == "" {
		return "skipped: no commit pointer"
	}
	if _, err := gitutil.ResolveRef(ctx, root, headSHA); err != nil {
		return "skipped: commit unavailable"
	}

	changes, err := trace.BuildChangesAt(ctx, root, slug, specsDirRel, meta.BaseSHA, headSHA, tasksContent, meta)
	if err != nil {
		return "error: " + err.Error()
	}
	if err := trace.WriteChanges(filepath.Join(dateDir, "changes.json"), changes); err != nil {
		return "error: " + err.Error()
	}
	return "rebuilt"
}

// readSnapshotPointer parses the compact archive pointer file.
func readSnapshotPointer(dateDir string) (branch, sha string) {
	content, err := os.ReadFile(filepath.Join(dateDir, "snapshot.sha"))
	if err != nil {
		return "", ""
	}
	for _, line := range strings.Split(string(content), "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "branch":
			branch = strings.TrimSpace(value)
		case "sha":
			sha = strings.TrimSpace(value)
		}
	}
	return branch, sha
}

func archivedInputs(ctx context.Context, root, specsDirRel, slug, dateDir, headSHA string) (trace.ArchiveMeta, string) {
	var meta trace.ArchiveMeta
	metaCandidates := []string{
		filepath.Join(dateDir, "trace-meta.json"),
		filepath.Join(dateDir, "specs", slug, "trace-meta.json"),
	}
	for _, metaPath := range metaCandidates {
		if content, err := os.ReadFile(metaPath); err == nil {
			_ = json.Unmarshal(content, &meta)
			break
		}
	}
	if meta.Version == 0 && headSHA != "" {
		rel := filepath.ToSlash(filepath.Join(specsDirRel, slug, "trace-meta.json"))
		if content, err := gitutil.FileAt(ctx, root, headSHA, rel); err == nil {
			_ = json.Unmarshal(content, &meta)
		}
	}

	tasksContent := ""
	tasksPath := filepath.Join(dateDir, "plan", "tasks.md")
	if content, err := os.ReadFile(tasksPath); err == nil {
		tasksContent = string(content)
	} else if headSHA != "" {
		for _, rel := range []string{
			filepath.ToSlash(filepath.Join(specsDirRel, slug, "tasks.md")),
			filepath.ToSlash(filepath.Join(specsDirRel, slug, "plan", "tasks.md")),
		} {
			if content, err := gitutil.FileAt(ctx, root, headSHA, rel); err == nil {
				tasksContent = string(content)
				break
			}
		}
	}
	return meta, tasksContent
}

func inferBaseFor(ctx context.Context, root, headSHA string) string {
	for _, branch := range []string{"main", "master"} {
		if sha, err := gitutil.MergeBase(ctx, root, branch, headSHA); err == nil && sha != "" {
			return sha
		}
	}
	return ""
}
