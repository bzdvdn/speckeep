package gitutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func EnsureRepository(ctx context.Context, root string) (bool, error) {
	gitDir := filepath.Join(root, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	if _, _, err := run(ctx, root, "git", "init"); err != nil {
		return false, err
	}

	return true, nil
}

func EnsureBranch(ctx context.Context, root, branch string) (string, error) {
	exists, err := branchExists(ctx, root, branch)
	if err != nil {
		return "", err
	}

	if exists {
		if _, _, err := run(ctx, root, "git", "checkout", branch); err != nil {
			return "", err
		}
		return fmt.Sprintf("switched to existing branch %s", branch), nil
	}

	hasCommits, err := hasCommits(ctx, root)
	if err != nil {
		return "", err
	}

	if hasCommits {
		if _, _, err := run(ctx, root, "git", "checkout", "-b", branch); err != nil {
			return "", err
		}
		return fmt.Sprintf("created and switched to new branch %s", branch), nil
	}

	if _, _, err := run(ctx, root, "git", "checkout", "--orphan", branch); err != nil {
		return "", err
	}

	return fmt.Sprintf("created and switched to new orphan branch %s", branch), nil
}

func CurrentBranch(ctx context.Context, root string) (string, error) {
	stdout, _, err := run(ctx, root, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout), nil
}

// IsRepository reports whether root is inside a git work tree.
func IsRepository(ctx context.Context, root string) bool {
	_, _, err := run(ctx, root, "git", "rev-parse", "--is-inside-work-tree")
	return err == nil
}

// Head returns the current HEAD commit SHA.
func Head(ctx context.Context, root string) (string, error) {
	stdout, _, err := run(ctx, root, "git", "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout), nil
}

// ResolveRef resolves a revision (branch, tag, SHA) to a full commit SHA.
func ResolveRef(ctx context.Context, root, ref string) (string, error) {
	stdout, _, err := run(ctx, root, "git", "rev-parse", "--verify", ref+"^{commit}")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout), nil
}

// MergeBase returns the best common ancestor of two revisions.
func MergeBase(ctx context.Context, root, a, b string) (string, error) {
	stdout, _, err := run(ctx, root, "git", "merge-base", a, b)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout), nil
}

// RootCommit returns the first commit reachable from ref.
func RootCommit(ctx context.Context, root, ref string) (string, error) {
	stdout, _, err := run(ctx, root, "git", "rev-list", "--max-parents=0", ref)
	if err != nil {
		return "", err
	}
	lines := strings.Fields(stdout)
	if len(lines) == 0 {
		return "", fmt.Errorf("git rev-list --max-parents=0 %s: no commits", ref)
	}
	return lines[len(lines)-1], nil
}

// CommitCount returns the number of commits reachable from ref.
func CommitCount(ctx context.Context, root, ref string) (int, error) {
	stdout, _, err := run(ctx, root, "git", "rev-list", "--count", ref)
	if err != nil {
		return 0, err
	}
	count := 0
	if _, err := fmt.Sscanf(strings.TrimSpace(stdout), "%d", &count); err != nil {
		return 0, fmt.Errorf("parse commit count %q: %w", stdout, err)
	}
	return count, nil
}

// IsAncestor reports whether ancestor is reachable from descendant.
func IsAncestor(ctx context.Context, root, ancestor, descendant string) bool {
	_, _, err := run(ctx, root, "git", "merge-base", "--is-ancestor", ancestor, descendant)
	return err == nil
}

// DiffFile describes one path changed between two revisions.
type DiffFile struct {
	Status    string // A|M|D|R|C|T|U
	Path      string // new path (old path for deletions)
	OldPath   string // set for renames and copies
	Additions int
	Deletions int
	Binary    bool
}

// DiffFiles returns the changed files between from and to, preserving renames.
// An empty to means the working tree.
func DiffFiles(ctx context.Context, root, from, to string) ([]DiffFile, error) {
	rangeArg := from + ".." + to
	if strings.TrimSpace(to) == "" {
		rangeArg = from
	}

	nameOut, _, err := run(ctx, root, "git", "diff", "--name-status", "-z", "-M", "--no-ext-diff", rangeArg)
	if err != nil {
		return nil, err
	}
	files := parseNameStatus(nameOut)
	if len(files) == 0 {
		return files, nil
	}

	numOut, _, err := run(ctx, root, "git", "diff", "--numstat", "-z", "-M", "--no-ext-diff", rangeArg)
	if err != nil {
		return nil, err
	}
	stats := parseNumstat(numOut)
	for i := range files {
		if stat, ok := stats[files[i].Path]; ok {
			files[i].Additions = stat.add
			files[i].Deletions = stat.del
			files[i].Binary = stat.binary
		}
		if files[i].OldPath != "" {
			if stat, ok := stats[files[i].OldPath]; ok && files[i].Additions == 0 && files[i].Deletions == 0 && !files[i].Binary {
				files[i].Additions = stat.add
				files[i].Deletions = stat.del
				files[i].Binary = stat.binary
			}
		}
	}
	return files, nil
}

// UntrackedFiles returns repo-relative untracked files (excluding ignored).
func UntrackedFiles(ctx context.Context, root string) ([]string, error) {
	stdout, _, err := run(ctx, root, "git", "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return nil, err
	}
	var files []string
	for _, file := range strings.Split(stdout, "\x00") {
		file = strings.TrimSpace(file)
		if file != "" {
			files = append(files, filepath.ToSlash(file))
		}
	}
	sort.Strings(files)
	return files, nil
}

// TrackedDirty reports whether the working tree has uncommitted changes to
// tracked files.
func TrackedDirty(ctx context.Context, root string) bool {
	stdout, _, err := run(ctx, root, "git", "status", "--porcelain", "--untracked-files=no")
	return err == nil && strings.TrimSpace(stdout) != ""
}

func parseNameStatus(out string) []DiffFile {
	fields := strings.Split(out, "\x00")
	var files []DiffFile
	for i := 0; i < len(fields); i++ {
		status := strings.TrimSpace(fields[i])
		if status == "" {
			continue
		}
		kind := status[:1]
		if kind == "R" || kind == "C" {
			if i+2 >= len(fields) {
				break
			}
			files = append(files, DiffFile{
				Status:  kind,
				OldPath: filepath.ToSlash(fields[i+1]),
				Path:    filepath.ToSlash(fields[i+2]),
			})
			i += 2
			continue
		}
		if i+1 >= len(fields) {
			break
		}
		files = append(files, DiffFile{
			Status: kind,
			Path:   filepath.ToSlash(fields[i+1]),
		})
		i++
	}
	return files
}

type numstat struct {
	add    int
	del    int
	binary bool
}

func parseNumstat(out string) map[string]numstat {
	fields := strings.Split(out, "\x00")
	stats := map[string]numstat{}
	for i := 0; i < len(fields); i++ {
		chunk := strings.TrimLeft(fields[i], "\n")
		if chunk == "" {
			continue
		}
		parts := strings.SplitN(chunk, "\t", 3)
		if len(parts) < 3 {
			continue
		}
		s := numstat{}
		s.binary = parts[0] == "-" || parts[1] == "-"
		if !s.binary {
			fmt.Sscanf(parts[0], "%d", &s.add)
			fmt.Sscanf(parts[1], "%d", &s.del)
		}
		stats[filepath.ToSlash(parts[2])] = s
	}
	return stats
}

// Commit is a compact commit record used in change attribution.
type Commit struct {
	SHA     string `json:"sha"`
	Subject string `json:"subject"`
}

// Commits returns commits reachable from to but not from from, newest first.
func Commits(ctx context.Context, root, from, to string) ([]Commit, error) {
	rangeArg := to
	if strings.TrimSpace(from) != "" {
		rangeArg = from + ".." + to
	}
	stdout, _, err := run(ctx, root, "git", "log", "--format=%H%x09%s", rangeArg)
	if err != nil {
		return nil, err
	}
	var commits []Commit
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		commit := Commit{SHA: parts[0]}
		if len(parts) == 2 {
			commit.Subject = parts[1]
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

// FileAt returns the content of path at the given revision (via git show).
func FileAt(ctx context.Context, root, revision, path string) ([]byte, error) {
	stdout, stderr, err := run(ctx, root, "git", "show", revision+":"+path)
	if err != nil {
		if strings.TrimSpace(stderr) != "" {
			return nil, fmt.Errorf("git show %s:%s: %w: %s", revision, path, err, strings.TrimSpace(stderr))
		}
		return nil, fmt.Errorf("git show %s:%s: %w", revision, path, err)
	}
	return []byte(stdout), nil
}

func branchExists(ctx context.Context, root, branch string) (bool, error) {
	_, _, err := run(ctx, root, "git", "rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
	if err == nil {
		return true, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() != 0 {
		return false, nil
	}

	return false, err
}

func hasCommits(ctx context.Context, root string) (bool, error) {
	_, _, err := run(ctx, root, "git", "rev-parse", "--verify", "HEAD")
	if err == nil {
		return true, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() != 0 {
		return false, nil
	}

	return false, err
}

func run(ctx context.Context, dir string, name string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = strings.TrimSpace(stdout.String())
		}
		if message != "" {
			return stdout.String(), stderr.String(), fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, message)
		}
		return stdout.String(), stderr.String(), fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}

	return stdout.String(), stderr.String(), nil
}
