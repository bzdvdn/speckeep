package skills

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	manifestVersion = 1
)

var skillIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,63}$`)

type Manifest struct {
	Version int     `yaml:"version"`
	Skills  []Entry `yaml:"skills,omitempty"`
}

type Entry struct {
	ID             string `yaml:"id"`
	Enabled        bool   `yaml:"enabled"`
	Source         string `yaml:"source"`
	Location       string `yaml:"location"`
	Ref            string `yaml:"ref,omitempty"`
	Path           string `yaml:"path,omitempty"`
	CheckoutDir    string `yaml:"checkout_dir,omitempty"`
	Version        string `yaml:"version,omitempty"`
	ResolvedCommit string `yaml:"resolved_commit,omitempty"`
}

type AddOptions struct {
	ID        string
	FromLocal string
	FromGit   string
	Ref       string
	Path      string
	Version   string
	Enabled   bool
}

type AddResult struct {
	Created bool
	Updated bool
	Entry   Entry
}

type RemoveResult struct {
	Removed bool
}

func ManifestPath(root string) string {
	return filepath.Join(root, ".speckeep", "skills", "manifest.yaml")
}

func Load(root string) (Manifest, error) {
	path := ManifestPath(root)
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Manifest{Version: manifestVersion}, nil
		}
		return Manifest{}, fmt.Errorf("read skills manifest: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse skills manifest: %w", err)
	}
	if manifest.Version == 0 {
		manifest.Version = manifestVersion
	}
	if manifest.Version != manifestVersion {
		return Manifest{}, fmt.Errorf("unsupported skills manifest version %d, expected %d", manifest.Version, manifestVersion)
	}
	if manifest.Skills == nil {
		manifest.Skills = []Entry{}
	}
	sort.Slice(manifest.Skills, func(i, j int) bool {
		return manifest.Skills[i].ID < manifest.Skills[j].ID
	})
	return manifest, nil
}

func Save(root string, manifest Manifest) error {
	manifest.Version = manifestVersion
	if manifest.Skills == nil {
		manifest.Skills = []Entry{}
	}
	sort.Slice(manifest.Skills, func(i, j int) bool {
		return manifest.Skills[i].ID < manifest.Skills[j].ID
	})

	path := ManifestPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create skills directory: %w", err)
	}
	content, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("marshal skills manifest: %w", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write skills manifest: %w", err)
	}
	return nil
}

func Add(root string, options AddOptions) (AddResult, error) {
	entry, err := buildEntry(root, options)
	if err != nil {
		return AddResult{}, err
	}

	manifest, err := Load(root)
	if err != nil {
		return AddResult{}, err
	}

	for i := range manifest.Skills {
		if manifest.Skills[i].ID != entry.ID {
			continue
		}
		manifest.Skills[i] = entry
		if err := Save(root, manifest); err != nil {
			return AddResult{}, err
		}
		return AddResult{Updated: true, Entry: entry}, nil
	}

	manifest.Skills = append(manifest.Skills, entry)
	if err := Save(root, manifest); err != nil {
		return AddResult{}, err
	}
	return AddResult{Created: true, Entry: entry}, nil
}

func Remove(root, id string) (RemoveResult, error) {
	skillID, err := normalizeID(id)
	if err != nil {
		return RemoveResult{}, err
	}

	manifest, err := Load(root)
	if err != nil {
		return RemoveResult{}, err
	}

	filtered := make([]Entry, 0, len(manifest.Skills))
	removed := false
	for _, entry := range manifest.Skills {
		if entry.ID == skillID {
			removed = true
			continue
		}
		filtered = append(filtered, entry)
	}
	if !removed {
		return RemoveResult{Removed: false}, nil
	}

	manifest.Skills = filtered
	if err := Save(root, manifest); err != nil {
		return RemoveResult{}, err
	}
	return RemoveResult{Removed: true}, nil
}

func buildEntry(root string, options AddOptions) (Entry, error) {
	id, err := normalizeID(options.ID)
	if err != nil {
		return Entry{}, err
	}

	sourceCount := 0
	if strings.TrimSpace(options.FromLocal) != "" {
		sourceCount++
	}
	if strings.TrimSpace(options.FromGit) != "" {
		sourceCount++
	}
	if sourceCount != 1 {
		return Entry{}, fmt.Errorf("exactly one source must be set: --from-local or --from-git")
	}

	entry := Entry{
		ID:      id,
		Enabled: options.Enabled,
		Version: strings.TrimSpace(options.Version),
	}
	if entry.Path = normalizeSubPath(options.Path); entry.Path == "." {
		entry.Path = ""
	}

	local := strings.TrimSpace(options.FromLocal)
	if local != "" {
		location, err := normalizeLocalLocation(root, local)
		if err != nil {
			return Entry{}, err
		}
		entry.Source = "local"
		entry.Location = location
		return entry, nil
	}

	entry.Source = "git"
	entry.Location = strings.TrimSpace(options.FromGit)
	if entry.Location == "" {
		return Entry{}, fmt.Errorf("git source is empty")
	}
	ref := strings.TrimSpace(options.Ref)
	if ref == "" {
		return Entry{}, fmt.Errorf("git source requires --ref (pin to tag or commit)")
	}
	if isFloatingGitRef(ref) {
		return Entry{}, fmt.Errorf("git ref %q looks floating; use a pinned tag or commit", ref)
	}
	entry.Ref = ref
	checkoutRelPath, resolvedCommit, err := materializeGitSource(root, id, entry.Location, ref)
	if err != nil {
		return Entry{}, err
	}
	entry.CheckoutDir = checkoutRelPath
	entry.ResolvedCommit = resolvedCommit
	return entry, nil
}

func normalizeID(value string) (string, error) {
	id := strings.ToLower(strings.TrimSpace(value))
	if !skillIDPattern.MatchString(id) {
		return "", fmt.Errorf("invalid skill id %q: expected [a-z0-9][a-z0-9_-]{1,63}", value)
	}
	return id, nil
}

func normalizeLocalLocation(root, value string) (string, error) {
	location := strings.TrimSpace(value)
	if location == "" {
		return "", fmt.Errorf("local source is empty")
	}

	resolved := location
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(root, filepath.FromSlash(location))
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("local source %q is not accessible: %w", location, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("local source %q must be a directory", location)
	}

	relPath, err := filepath.Rel(root, resolved)
	if err == nil && !strings.HasPrefix(relPath, ".."+string(filepath.Separator)) && relPath != ".." {
		return filepath.ToSlash(relPath), nil
	}
	return filepath.Clean(resolved), nil
}

func normalizeSubPath(value string) string {
	path := strings.TrimSpace(value)
	if path == "" {
		return "."
	}
	return filepath.ToSlash(filepath.Clean(path))
}

func isFloatingGitRef(ref string) bool {
	lower := strings.ToLower(strings.TrimSpace(ref))
	switch lower {
	case "main", "master", "dev", "develop", "head", "latest", "stable":
		return true
	}
	return strings.HasPrefix(lower, "refs/heads/")
}

func isCommitRef(ref string) bool {
	trimmed := strings.TrimSpace(ref)
	if len(trimmed) < 7 || len(trimmed) > 40 {
		return false
	}
	for _, char := range trimmed {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') && (char < 'A' || char > 'F') {
			return false
		}
	}
	return true
}

func IsFloatingGitRef(ref string) bool {
	return isFloatingGitRef(ref)
}

func IsCommitRef(ref string) bool {
	return isCommitRef(ref)
}

func checkoutPath(root, id string) string {
	return filepath.Join(root, ".speckeep", "skills", "checkouts", id)
}

func materializeGitSource(root, id, sourceURL, ref string) (string, string, error) {
	destination := checkoutPath(root, id)
	if err := os.RemoveAll(destination); err != nil {
		return "", "", fmt.Errorf("reset existing git checkout %q: %w", destination, err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return "", "", fmt.Errorf("create git checkout directory: %w", err)
	}

	if output, err := runGit(root, "clone", sourceURL, destination); err != nil {
		return "", "", fmt.Errorf("git clone failed: %s", strings.TrimSpace(output))
	}
	if output, err := runGit(root, "-C", destination, "checkout", "--force", ref); err != nil {
		return "", "", fmt.Errorf("git checkout %q failed: %s", ref, strings.TrimSpace(output))
	}
	output, err := runGit(root, "-C", destination, "rev-parse", "HEAD")
	if err != nil {
		return "", "", fmt.Errorf("resolve git HEAD failed: %s", strings.TrimSpace(output))
	}
	commit := strings.TrimSpace(output)
	relative, err := filepath.Rel(root, destination)
	if err != nil {
		return "", "", fmt.Errorf("resolve checkout relative path: %w", err)
	}
	return filepath.ToSlash(relative), commit, nil
}

func runGit(root string, args ...string) (string, error) {
	command := exec.Command("git", args...)
	command.Dir = root
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	output := strings.TrimSpace(stdout.String() + "\n" + stderr.String())
	return output, err
}

func ValidateManifest(root string, manifest Manifest) (errors []string, warnings []string) {
	seen := map[string]struct{}{}
	for _, entry := range manifest.Skills {
		if _, ok := seen[entry.ID]; ok {
			errors = append(errors, fmt.Sprintf("duplicate skill id %q in manifest", entry.ID))
		}
		seen[entry.ID] = struct{}{}

		if !skillIDPattern.MatchString(entry.ID) {
			errors = append(errors, fmt.Sprintf("invalid skill id %q", entry.ID))
		}

		switch entry.Source {
		case "local":
			if strings.TrimSpace(entry.Location) == "" {
				errors = append(errors, fmt.Sprintf("skill %q has empty local location", entry.ID))
				continue
			}
			location := entry.Location
			if !filepath.IsAbs(location) {
				location = filepath.Join(root, filepath.FromSlash(location))
			}
			info, err := os.Stat(location)
			if err != nil {
				errors = append(errors, fmt.Sprintf("skill %q local location is not accessible: %s", entry.ID, entry.Location))
				continue
			}
			if !info.IsDir() {
				errors = append(errors, fmt.Sprintf("skill %q local location must be a directory: %s", entry.ID, entry.Location))
				continue
			}

			if subPath := normalizeSubPath(entry.Path); subPath != "." {
				target := filepath.Join(location, filepath.FromSlash(subPath))
				subInfo, err := os.Stat(target)
				if err != nil || !subInfo.IsDir() {
					errors = append(errors, fmt.Sprintf("skill %q path is not accessible under local source: %s", entry.ID, entry.Path))
				}
			}
		case "git":
			if strings.TrimSpace(entry.Location) == "" {
				errors = append(errors, fmt.Sprintf("skill %q has empty git location", entry.ID))
			}
			if strings.TrimSpace(entry.Ref) == "" {
				errors = append(errors, fmt.Sprintf("skill %q git source requires ref", entry.ID))
			} else if isFloatingGitRef(entry.Ref) {
				errors = append(errors, fmt.Sprintf("skill %q uses floating git ref %q", entry.ID, entry.Ref))
			}
			if strings.TrimSpace(entry.ResolvedCommit) == "" {
				warnings = append(warnings, fmt.Sprintf("skill %q has no resolved_commit", entry.ID))
			}
			if strings.TrimSpace(entry.CheckoutDir) == "" {
				warnings = append(warnings, fmt.Sprintf("skill %q has no checkout_dir", entry.ID))
			} else {
				checkout := entry.CheckoutDir
				if !filepath.IsAbs(checkout) {
					checkout = filepath.Join(root, filepath.FromSlash(checkout))
				}
				info, err := os.Stat(checkout)
				if err != nil || !info.IsDir() {
					errors = append(errors, fmt.Sprintf("skill %q checkout_dir is not accessible: %s", entry.ID, entry.CheckoutDir))
				}
			}
		default:
			errors = append(errors, fmt.Sprintf("skill %q has unsupported source %q", entry.ID, entry.Source))
		}
	}
	return errors, warnings
}
