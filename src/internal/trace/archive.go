package trace

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
)

// ArchivedChangeSet is a changes.json loaded from one archive snapshot.
type ArchivedChangeSet struct {
	Slug    string
	Date    string
	Path    string
	Changes ChangeSet
}

// ArchivedSnapshotsWithoutChanges lists archive snapshots that predate
// change attribution and can be backfilled by `speckeep trace rebuild`.
func ArchivedSnapshotsWithoutChanges(archiveDir string) ([]string, error) {
	slugEntries, err := os.ReadDir(archiveDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, slugEntry := range slugEntries {
		if !slugEntry.IsDir() {
			continue
		}
		slugDir := filepath.Join(archiveDir, slugEntry.Name())
		dateEntries, err := os.ReadDir(slugDir)
		if err != nil {
			continue
		}
		for _, dateEntry := range dateEntries {
			if !dateEntry.IsDir() {
				continue
			}
			dateDir := filepath.Join(slugDir, dateEntry.Name())
			if _, err := os.Stat(filepath.Join(dateDir, "changes.json")); err == nil {
				continue
			}
			if _, err := os.Stat(filepath.Join(dateDir, "summary.md")); err != nil {
				continue
			}
			rel, relErr := filepath.Rel(archiveDir, dateDir)
			if relErr != nil {
				rel = dateDir
			}
			out = append(out, filepath.ToSlash(rel))
		}
	}
	sort.Strings(out)
	return out, nil
}

// ArchivedChangeSets reads archived change sets for one slug, or for all
// slugs when slug is empty. Snapshots without changes.json are skipped.
func ArchivedChangeSets(archiveDir, slug string) ([]ArchivedChangeSet, error) {
	var out []ArchivedChangeSet
	slugs := []string{slug}
	if slug == "" {
		entries, err := os.ReadDir(archiveDir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
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

	for _, name := range slugs {
		slugDir := filepath.Join(archiveDir, name)
		dates, err := os.ReadDir(slugDir)
		if err != nil {
			continue
		}
		for _, date := range dates {
			if !date.IsDir() {
				continue
			}
			path := filepath.Join(slugDir, date.Name(), "changes.json")
			changes, err := ReadChanges(path)
			if err != nil {
				continue
			}
			out = append(out, ArchivedChangeSet{
				Slug:    name,
				Date:    date.Name(),
				Path:    path,
				Changes: changes,
			})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Slug != out[j].Slug {
			return out[i].Slug < out[j].Slug
		}
		return out[i].Date > out[j].Date
	})
	return out, nil
}
