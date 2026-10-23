package trace

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeArchiveSnapshot(t *testing.T, archiveDir, slug, date string, changes *ChangeSet, withSummary bool) string {
	t.Helper()
	dir := filepath.Join(archiveDir, slug, date)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if withSummary {
		if err := os.WriteFile(filepath.Join(dir, "summary.md"), []byte("# summary\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if changes != nil {
		if err := WriteChanges(filepath.Join(dir, "changes.json"), *changes); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestArchivedChangeSets(t *testing.T) {
	archiveDir := t.TempDir()
	writeArchiveSnapshot(t, archiveDir, "demo", "2026-01-01", &ChangeSet{Version: 1, Slug: "demo"}, true)
	writeArchiveSnapshot(t, archiveDir, "demo", "2026-01-02", &ChangeSet{Version: 1, Slug: "demo"}, true)
	writeArchiveSnapshot(t, archiveDir, "other", "2026-01-03", &ChangeSet{Version: 1, Slug: "other"}, true)

	sets, err := ArchivedChangeSets(archiveDir, "demo")
	if err != nil {
		t.Fatalf("ArchivedChangeSets: %v", err)
	}
	if len(sets) != 2 {
		t.Fatalf("got %d demo sets, want 2", len(sets))
	}
	if sets[0].Date != "2026-01-02" || sets[1].Date != "2026-01-01" {
		t.Errorf("sets not sorted newest first: %+v", sets)
	}

	all, err := ArchivedChangeSets(archiveDir, "")
	if err != nil {
		t.Fatalf("ArchivedChangeSets(all): %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("got %d sets, want 3", len(all))
	}

	if sets, err := ArchivedChangeSets(filepath.Join(archiveDir, "missing"), ""); err != nil || len(sets) != 0 {
		t.Errorf("missing archive dir = %+v, %v", sets, err)
	}
}

func TestArchivedSnapshotsWithoutChanges(t *testing.T) {
	archiveDir := t.TempDir()
	writeArchiveSnapshot(t, archiveDir, "demo", "2026-01-01", &ChangeSet{Version: 1}, true)
	writeArchiveSnapshot(t, archiveDir, "demo", "2026-01-02", nil, true)
	writeArchiveSnapshot(t, archiveDir, "other", "2026-01-03", nil, false)

	missing, err := ArchivedSnapshotsWithoutChanges(archiveDir)
	if err != nil {
		t.Fatalf("ArchivedSnapshotsWithoutChanges: %v", err)
	}
	if !reflect.DeepEqual(missing, []string{"demo/2026-01-02"}) {
		t.Errorf("missing = %v", missing)
	}

	if missing, err := ArchivedSnapshotsWithoutChanges(filepath.Join(archiveDir, "missing")); err != nil || len(missing) != 0 {
		t.Errorf("missing archive dir = %v, %v", missing, err)
	}
}

func TestReadChangesRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "changes.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadChanges(path); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if _, err := ReadChanges(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected error for missing file")
	}
}
