package trace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"speckeep/src/internal/config"
	"speckeep/src/internal/featurepaths"
	"speckeep/src/internal/gitutil"
)

const changeSetVersion = 1

var defaultIgnorePrefixes = []string{"vendor/", "node_modules/", "dist/"}

// ChangedFile is one file attributed to a feature.
type ChangedFile struct {
	Path      string   `json:"path"`
	OldPath   string   `json:"old_path,omitempty"`
	Status    string   `json:"status"`
	Additions int      `json:"additions,omitempty"`
	Deletions int      `json:"deletions,omitempty"`
	Binary    bool     `json:"binary,omitempty"`
	Tasks     []string `json:"tasks,omitempty"`
	ACs       []string `json:"acs,omitempty"`
	Anchors   []string `json:"anchors,omitempty"`
}

// ChangeSet is the durable record of what a feature changed.
type ChangeSet struct {
	Version     int              `json:"version"`
	Slug        string           `json:"slug"`
	BaseSHA     string           `json:"base_sha,omitempty"`
	HeadSHA     string           `json:"head_sha,omitempty"`
	BaseBranch  string           `json:"base_branch,omitempty"`
	BaseSource  string           `json:"base_source"`
	Source      string           `json:"source"` // git|declared
	GeneratedAt string           `json:"generated_at"`
	Worktree    bool             `json:"worktree,omitempty"`
	Dirty       bool             `json:"dirty,omitempty"`
	Commits     []gitutil.Commit `json:"commits,omitempty"`
	Files       []ChangedFile    `json:"files"`
	Unmapped    []string         `json:"unmapped,omitempty"`
	NotChanged  []string         `json:"declared_not_changed,omitempty"`
	Artifacts   []string         `json:"artifacts,omitempty"`
}

// BuildChanges computes the actual footprint of an active feature from git,
// mapping changed files to declared task surfaces.
func BuildChanges(ctx context.Context, root, slug string) (ChangeSet, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return ChangeSet{}, err
	}
	cfg, err := config.Load(ctx, root)
	if err != nil {
		return ChangeSet{}, err
	}
	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return ChangeSet{}, err
	}
	tasksContent := ""
	if tasksPath, _ := featurepaths.ResolveTasks(specsDir, slug); tasksPath != "" {
		if content, readErr := os.ReadFile(tasksPath); readErr == nil {
			tasksContent = string(content)
		}
	}
	meta, err := ResolveBase(ctx, root, specsDir, slug)
	if err != nil {
		return ChangeSet{}, err
	}
	return buildChanges(ctx, root, slug, cfg.Paths.SpecsDir, tasksContent, meta, cfg.Trace.Ignore, true)
}

// BuildChangesCommitted computes the footprint of committed history only. It is
// used by archive so changes.json stays consistent with the recorded head SHA.
func BuildChangesCommitted(ctx context.Context, root, slug string) (ChangeSet, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return ChangeSet{}, err
	}
	cfg, err := config.Load(ctx, root)
	if err != nil {
		return ChangeSet{}, err
	}
	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return ChangeSet{}, err
	}
	tasksContent := ""
	if tasksPath, _ := featurepaths.ResolveTasks(specsDir, slug); tasksPath != "" {
		if content, readErr := os.ReadFile(tasksPath); readErr == nil {
			tasksContent = string(content)
		}
	}
	meta, err := ResolveBase(ctx, root, specsDir, slug)
	if err != nil {
		return ChangeSet{}, err
	}
	return buildChanges(ctx, root, slug, cfg.Paths.SpecsDir, tasksContent, meta, cfg.Trace.Ignore, false)
}

// BuildChangesAt computes a change set from explicit revisions. Used by
// rebuild for archived features.
func BuildChangesAt(ctx context.Context, root, slug, specsDirRel, baseSHA, headSHA, tasksContent string, meta Meta) (ChangeSet, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return ChangeSet{}, err
	}
	if meta.BaseSHA == "" {
		meta.BaseSHA = baseSHA
	}
	if meta.BaseSource == "" {
		meta.BaseSource = "inferred"
	}
	ignore := defaultIgnorePrefixes
	if cfg, err := config.Load(ctx, root); err == nil {
		ignore = cfg.Trace.Ignore
	}
	return buildChangesAt(ctx, root, slug, specsDirRel, tasksContent, meta, headSHA, ignore, false)
}

func buildChanges(ctx context.Context, root, slug, specsDirRel, tasksContent string, meta Meta, ignore []string, worktree bool) (ChangeSet, error) {
	head := ""
	if gitutil.IsRepository(ctx, root) {
		if sha, err := gitutil.Head(ctx, root); err == nil {
			head = sha
		}
	}
	return buildChangesAt(ctx, root, slug, specsDirRel, tasksContent, meta, head, ignore, worktree)
}

func buildChangesAt(ctx context.Context, root, slug, specsDirRel, tasksContent string, meta Meta, headSHA string, ignore []string, worktree bool) (ChangeSet, error) {
	surfaces := ParseTaskSurfaces(tasksContent)
	cs := ChangeSet{
		Version:     changeSetVersion,
		Slug:        slug,
		BaseSHA:     meta.BaseSHA,
		HeadSHA:     headSHA,
		BaseBranch:  meta.BaseBranch,
		BaseSource:  meta.BaseSource,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Files:       []ChangedFile{},
	}
	if gitutil.IsRepository(ctx, root) && !worktree {
		cs.Dirty = gitutil.TrackedDirty(ctx, root)
		if untracked, err := gitutil.UntrackedFiles(ctx, root); err == nil && len(untracked) > 0 {
			cs.Dirty = true
		}
	}

	if meta.BaseSHA == "" || (headSHA == "" && !worktree) {
		cs.Source = "declared"
		cs.Files = declaredFiles(surfaces)
		return cs, nil
	}

	to := headSHA
	hasWorktreeChanges := false
	if worktree {
		to = ""
		hasWorktreeChanges = cs.Dirty
	}
	diff, err := gitutil.DiffFiles(ctx, root, meta.BaseSHA, to)
	if err != nil {
		cs.Source = "declared"
		cs.Files = declaredFiles(surfaces)
		return cs, nil
	}
	cs.Source = "git"
	if headSHA != "" {
		cs.Commits, _ = gitutil.Commits(ctx, root, meta.BaseSHA, headSHA)
	}

	specPrefix := strings.Trim(filepathToSlash(specsDirRel), "/")
	seen := map[string]bool{}
	changed := make([]string, 0, len(diff))
	for _, entry := range diff {
		if ignoredPath(entry.Path, ignore) {
			continue
		}
		seen[entry.Path] = true
		if specPrefix != "" && strings.HasPrefix(entry.Path, specPrefix+"/") {
			cs.Artifacts = append(cs.Artifacts, entry.Path)
			continue
		}
		tasks, acs, anchors := MatchSurface(entry.Path, surfaces)
		if len(tasks) == 0 && entry.OldPath != "" {
			tasks, acs, anchors = MatchSurface(entry.OldPath, surfaces)
		}
		file := ChangedFile{
			Path:      entry.Path,
			OldPath:   entry.OldPath,
			Status:    entry.Status,
			Additions: entry.Additions,
			Deletions: entry.Deletions,
			Binary:    entry.Binary,
			Tasks:     tasks,
			ACs:       acs,
			Anchors:   anchors,
		}
		cs.Files = append(cs.Files, file)
		changed = append(changed, entry.Path)
		if len(tasks) == 0 {
			cs.Unmapped = append(cs.Unmapped, entry.Path)
		}
	}

	if worktree {
		if untracked, err := gitutil.UntrackedFiles(ctx, root); err == nil {
			for _, path := range untracked {
				if seen[path] || ignoredPath(path, ignore) {
					continue
				}
				if specPrefix != "" && strings.HasPrefix(path, specPrefix+"/") {
					cs.Artifacts = append(cs.Artifacts, path)
					continue
				}
				tasks, acs, anchors := MatchSurface(path, surfaces)
				cs.Files = append(cs.Files, ChangedFile{
					Path:    path,
					Status:  "?",
					Tasks:   tasks,
					ACs:     acs,
					Anchors: anchors,
				})
				changed = append(changed, path)
				hasWorktreeChanges = true
				if len(tasks) == 0 {
					cs.Unmapped = append(cs.Unmapped, path)
				}
			}
		}
	}
	cs.Worktree = hasWorktreeChanges

	sort.Slice(cs.Files, func(i, j int) bool { return cs.Files[i].Path < cs.Files[j].Path })
	sort.Strings(cs.Unmapped)
	sort.Strings(cs.Artifacts)
	cs.NotChanged = declaredNotChanged(surfaces, changed)
	return cs, nil
}

func declaredFiles(surfaces []TaskSurface) []ChangedFile {
	refs := DeclaredSurfaces(surfaces)
	files := make([]ChangedFile, 0, len(refs))
	for _, ref := range refs {
		tasks, acs, anchors := matchRef(ref, surfaces)
		files = append(files, ChangedFile{
			Path:    ref,
			Status:  "declared",
			Tasks:   tasks,
			ACs:     acs,
			Anchors: anchors,
		})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files
}

func matchRef(ref string, surfaces []TaskSurface) (tasks, acs, anchors []string) {
	for _, surface := range surfaces {
		matched := false
		for _, touch := range surface.Touches {
			if touch == ref || strings.TrimSuffix(touch, "/") == strings.TrimSuffix(ref, "/") {
				matched = true
				break
			}
		}
		if !matched {
			for _, proof := range surface.Proofs {
				if filepathToSlash(proof.File) == ref {
					matched = true
					if proof.Anchor != "" {
						anchors = append(anchors, proof.Anchor)
					}
				}
			}
		}
		if matched {
			tasks = append(tasks, surface.ID)
			acs = append(acs, surface.ACs...)
		}
	}
	return uniqueSorted(tasks), uniqueSorted(acs), uniqueSorted(anchors)
}

func declaredNotChanged(surfaces []TaskSurface, changed []string) []string {
	var out []string
	for _, surface := range surfaces {
		if !surface.Closed {
			continue
		}
		for _, ref := range surfaceRefs(surface) {
			found := false
			for _, file := range changed {
				if refMatches(file, ref) {
					found = true
					break
				}
			}
			if !found {
				out = append(out, ref)
			}
		}
	}
	return uniqueSorted(out)
}

func ignoredPath(file string, ignore []string) bool {
	if len(ignore) == 0 {
		ignore = defaultIgnorePrefixes
	}
	for _, prefix := range ignore {
		prefix = filepathToSlash(strings.TrimSpace(prefix))
		if prefix == "" {
			continue
		}
		if strings.HasPrefix(file, prefix) {
			return true
		}
	}
	return false
}

// MarshalJSON renders a change set as indented JSON with a trailing newline.
func (cs ChangeSet) MarshalJSONLine() ([]byte, error) {
	content, err := json.MarshalIndent(cs, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(content, '\n'), nil
}

// WriteChanges persists a change set next to archive artifacts.
func WriteChanges(path string, cs ChangeSet) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content, err := cs.MarshalJSONLine()
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

// ReadChanges reads a changes.json file.
func ReadChanges(path string) (ChangeSet, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return ChangeSet{}, err
	}
	var cs ChangeSet
	if err := json.Unmarshal(content, &cs); err != nil {
		return ChangeSet{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return cs, nil
}
