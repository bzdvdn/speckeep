package trace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"speckeep/src/internal/config"
	"speckeep/src/internal/gitutil"
)

const metaVersion = 1

// Meta is the machine-written trace metadata of one feature. It pins the
// commit the feature branched from so change attribution survives merges.
type Meta struct {
	Version    int    `json:"version"`
	Slug       string `json:"slug"`
	BaseSHA    string `json:"base_sha,omitempty"`
	BaseBranch string `json:"base_branch,omitempty"`
	BaseSource string `json:"base_source"` // explicit|inferred|declared
	CreatedAt  string `json:"created_at"`
}

// ArchiveMeta augments Meta with the archived commit pointer so archived
// features can be rebuilt without a compact snapshot.sha.
type ArchiveMeta struct {
	Meta
	HeadSHA string `json:"head_sha,omitempty"`
	Branch  string `json:"branch,omitempty"`
}

// MetaPath returns the location of a feature's trace meta file.
func MetaPath(specsDir, slug string) string {
	return filepath.Join(specsDir, slug, "trace-meta.json")
}

// LoadMeta reads trace metadata for a feature. The boolean reports presence.
func LoadMeta(specsDir, slug string) (Meta, bool, error) {
	content, err := os.ReadFile(MetaPath(specsDir, slug))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Meta{}, false, nil
		}
		return Meta{}, false, err
	}
	var meta Meta
	if err := json.Unmarshal(content, &meta); err != nil {
		return Meta{}, false, fmt.Errorf("parse %s: %w", MetaPath(specsDir, slug), err)
	}
	return meta, true, nil
}

// WriteMeta persists trace metadata for a feature.
func WriteMeta(specsDir, slug string, meta Meta) error {
	return WriteMetaFile(MetaPath(specsDir, slug), meta)
}

// WriteMetaFile writes trace metadata to an explicit path.
func WriteMetaFile(path string, meta Meta) error {
	return writeJSONFile(path, meta)
}

// WriteArchiveMetaFile writes archive metadata to an explicit path.
func WriteArchiveMetaFile(path string, meta ArchiveMeta) error {
	return writeJSONFile(path, meta)
}

func writeJSONFile(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(path, content, 0o644)
}

// InitMeta pins the feature base commit. It is idempotent: an existing base is
// kept unless force is set. baseRef overrides auto-detection (main/master).
func InitMeta(ctx context.Context, root, slug, baseRef string, force bool) (Meta, bool, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Meta{}, false, err
	}
	cfg, err := config.Load(ctx, root)
	if err != nil {
		return Meta{}, false, err
	}
	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return Meta{}, false, err
	}

	existing, exists, err := LoadMeta(specsDir, slug)
	if err != nil {
		return Meta{}, false, err
	}
	if exists && existing.BaseSHA != "" && !force {
		return existing, false, nil
	}

	meta := Meta{
		Version:   metaVersion,
		Slug:      slug,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if !gitutil.IsRepository(ctx, root) {
		meta.BaseSource = "declared"
		if err := WriteMeta(specsDir, slug, meta); err != nil {
			return Meta{}, false, err
		}
		return meta, true, nil
	}

	head, err := gitutil.Head(ctx, root)
	if err != nil {
		// Orphan/first-feature branch with no commits yet: record declared mode.
		meta.BaseSource = "declared"
		if err := WriteMeta(specsDir, slug, meta); err != nil {
			return Meta{}, false, err
		}
		return meta, true, nil
	}

	baseBranch := strings.TrimSpace(baseRef)
	if baseBranch == "" {
		baseBranch = detectBaseBranch(ctx, root)
	}
	meta.BaseBranch = baseBranch
	meta.BaseSHA = head
	meta.BaseSource = "explicit"

	if baseBranch != "" {
		if baseSHA, err := gitutil.MergeBase(ctx, root, baseBranch, "HEAD"); err == nil && baseSHA != "" {
			meta.BaseSHA = baseSHA
			meta.BaseSource = "inferred"
		}
	}
	if baseRef != "" {
		baseSHA, err := gitutil.ResolveRef(ctx, root, baseRef)
		if err != nil {
			return Meta{}, false, fmt.Errorf("resolve base %q: %w", baseRef, err)
		}
		meta.BaseSHA = baseSHA
		meta.BaseSource = "explicit"
	}
	if existing.CreatedAt != "" {
		meta.CreatedAt = existing.CreatedAt
	}

	if err := WriteMeta(specsDir, slug, meta); err != nil {
		return Meta{}, false, err
	}
	return meta, true, nil
}

// ResolveBase returns the base SHA for a feature, falling back to inference
// when meta is absent. The second result reports how the base was determined.
func ResolveBase(ctx context.Context, root, specsDir, slug string) (Meta, error) {
	meta, exists, err := LoadMeta(specsDir, slug)
	if err != nil {
		return Meta{}, err
	}
	if exists && meta.BaseSHA != "" {
		return meta, nil
	}

	meta.Version = metaVersion
	meta.Slug = slug
	meta.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	if !gitutil.IsRepository(ctx, root) {
		meta.BaseSource = "declared"
		return meta, nil
	}

	baseBranch := detectBaseBranch(ctx, root)
	if baseBranch != "" {
		if baseSHA, err := gitutil.MergeBase(ctx, root, baseBranch, "HEAD"); err == nil && baseSHA != "" {
			meta.BaseSHA = baseSHA
			meta.BaseBranch = baseBranch
			meta.BaseSource = "inferred"
			return meta, nil
		}
	}

	count, err := gitutil.CommitCount(ctx, root, "HEAD")
	if err == nil && count <= 1 {
		if baseSHA, err := gitutil.RootCommit(ctx, root, "HEAD"); err == nil {
			meta.BaseSHA = baseSHA
			meta.BaseSource = "inferred"
			return meta, nil
		}
	}

	meta.BaseSource = "declared"
	return meta, nil
}

// detectBaseBranch picks the conventional integration branch when present.
func detectBaseBranch(ctx context.Context, root string) string {
	for _, candidate := range []string{"main", "master"} {
		if _, err := gitutil.ResolveRef(ctx, root, candidate); err == nil {
			return candidate
		}
	}
	return ""
}
