package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"speckeep/src/internal/templates"
)

type DemoOptions struct {
	Shell        string
	AgentTargets []string
}

type DemoResult struct {
	RootAbs      string
	Shell        string
	AgentTargets []string

	ExampleSlug string
	Created     []string // paths relative to the project root
}

func Demo(root string, options DemoOptions) (DemoResult, error) {
	shell := options.Shell
	if shell == "" {
		shell = "sh"
	}

	_, err := Initialize(root, InitOptions{
		InitGit:      false,
		DefaultLang:  "en",
		Shell:        shell,
		AgentTargets: options.AgentTargets,
	})
	if err != nil {
		return DemoResult{}, err
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return DemoResult{}, err
	}

	demoFiles, err := templates.DemoFiles()
	if err != nil {
		return DemoResult{}, fmt.Errorf("load demo files: %w", err)
	}

	var created []string

	for _, file := range demoFiles {
		target := filepath.Join(absRoot, file.TargetPath)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return DemoResult{}, err
		}
		if err := os.WriteFile(target, []byte(file.Content), file.Mode); err != nil {
			return DemoResult{}, err
		}
		created = append(created, rel(absRoot, target))
	}

	return DemoResult{
		RootAbs:      absRoot,
		Shell:        strings.ToLower(shell),
		AgentTargets: append([]string(nil), options.AgentTargets...),
		ExampleSlug:  "export-report",
		Created:      created,
	}, nil
}
