package gitutil

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func EnsureRepository(root string) (bool, error) {
	gitDir := filepath.Join(root, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	if _, _, err := run(root, "git", "init"); err != nil {
		return false, err
	}

	return true, nil
}

func EnsureBranch(root, branch string) (string, error) {
	exists, err := branchExists(root, branch)
	if err != nil {
		return "", err
	}

	if exists {
		if _, _, err := run(root, "git", "checkout", branch); err != nil {
			return "", err
		}
		return fmt.Sprintf("switched to existing branch %s", branch), nil
	}

	hasCommits, err := hasCommits(root)
	if err != nil {
		return "", err
	}

	if hasCommits {
		if _, _, err := run(root, "git", "checkout", "-b", branch); err != nil {
			return "", err
		}
		return fmt.Sprintf("created and switched to new branch %s", branch), nil
	}

	if _, _, err := run(root, "git", "checkout", "--orphan", branch); err != nil {
		return "", err
	}

	return fmt.Sprintf("created and switched to new orphan branch %s", branch), nil
}

func CurrentBranch(root string) (string, error) {
	stdout, _, err := run(root, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout), nil
}

func branchExists(root, branch string) (bool, error) {
	_, _, err := run(root, "git", "rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
	if err == nil {
		return true, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() != 0 {
		return false, nil
	}

	return false, err
}

func hasCommits(root string) (bool, error) {
	_, _, err := run(root, "git", "rev-parse", "--verify", "HEAD")
	if err == nil {
		return true, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() != 0 {
		return false, nil
	}

	return false, err
}

func run(dir string, name string, args ...string) (string, string, error) {
	cmd := exec.Command(name, args...)
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
