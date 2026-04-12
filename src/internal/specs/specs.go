package specs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"speckeep/src/internal/config"
	"speckeep/src/internal/featurepaths"
	"speckeep/src/internal/gitutil"
)

type CreateOptions struct {
	CreateBranch bool
	BranchPrefix string
}

type CreateResult struct {
	Messages []string
}

type ResolvedInput struct {
	Title string
	Slug  string
}

func List(root string) ([]string, error) {
	cfg, err := config.Load(root)
	if err != nil {
		return nil, err
	}

	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return nil, err
	}
	return featurepaths.ListSpecSlugs(specsDir)
}

func Show(root, name string) (string, error) {
	cfg, err := config.Load(root)
	if err != nil {
		return "", err
	}

	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return "", err
	}

	specPath, _ := featurepaths.ResolveSpec(specsDir, name)
	content, err := os.ReadFile(specPath)
	if err != nil {
		return "", fmt.Errorf("read spec %q: %w", name, err)
	}

	return string(content), nil
}

func Create(root, name string, options CreateOptions) (CreateResult, error) {
	resolved, err := ResolveInput(name)
	if err != nil {
		return CreateResult{}, err
	}

	cfg, err := config.Load(root)
	if err != nil {
		return CreateResult{}, err
	}

	var messages []string
	if options.CreateBranch {
		message, err := gitutil.EnsureBranch(root, branchName(resolved.Slug, options.BranchPrefix))
		if err != nil {
			return CreateResult{}, err
		}
		messages = append(messages, message)
	} else {
		messages = append(messages, "skipped feature branch creation")
	}

	title := resolved.Title
	templatesDir, err := cfg.TemplatesDir(root)
	if err != nil {
		return CreateResult{}, err
	}
	specTemplatePath := filepath.Join(templatesDir, "spec.md")
	tasksTemplatePath := filepath.Join(templatesDir, "tasks.md")
	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return CreateResult{}, err
	}

	specTemplate, err := os.ReadFile(specTemplatePath)
	if err != nil {
		return CreateResult{}, fmt.Errorf("read spec template: %w", err)
	}
	tasksTemplate, err := os.ReadFile(tasksTemplatePath)
	if err != nil {
		return CreateResult{}, fmt.Errorf("read tasks template: %w", err)
	}

	specPath := featurepaths.Spec(specsDir, resolved.Slug)
	tasksPath := featurepaths.Tasks(specsDir, resolved.Slug)

	created, err := writeFilledTemplate(specPath, string(specTemplate), title)
	if err != nil {
		return CreateResult{}, err
	}
	if created {
		messages = append(messages, fmt.Sprintf("created %s", displayPath(root, specPath)))
	} else {
		messages = append(messages, fmt.Sprintf("kept existing %s", displayPath(root, specPath)))
	}

	created, err = writeFilledTemplate(tasksPath, string(tasksTemplate), title)
	if err != nil {
		return CreateResult{}, err
	}
	if created {
		messages = append(messages, fmt.Sprintf("created %s", displayPath(root, tasksPath)))
	} else {
		messages = append(messages, fmt.Sprintf("kept existing %s", displayPath(root, tasksPath)))
	}

	return CreateResult{Messages: messages}, nil
}

func ResolveInput(input string) (ResolvedInput, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return ResolvedInput{}, fmt.Errorf("spec input cannot be empty")
	}

	if looksLikeURL(input) {
		return ResolvedInput{}, fmt.Errorf("spec input %q looks like a URL; provide a short feature name or add name:/slug: metadata in a local prompt file", input)
	}

	if info, err := os.Stat(input); err == nil && !info.IsDir() {
		return resolveFileInput(input)
	}

	slug := slugify(input)
	if err := validateSlug(slug, input); err != nil {
		return ResolvedInput{}, err
	}

	return ResolvedInput{
		Title: titleFromSlug(slug),
		Slug:  slug,
	}, nil
}

func writeFilledTemplate(path, templateContent, title string) (bool, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}

	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}

	content := strings.ReplaceAll(templateContent, "<Spec Title>", title)
	return true, os.WriteFile(path, []byte(content), 0o644)
}

func resolveFileInput(path string) (ResolvedInput, error) {
	name, slug, err := readPromptMetadata(path)
	if err != nil {
		return ResolvedInput{}, err
	}

	if slug != "" {
		if err := validateSlug(slug, path); err != nil {
			return ResolvedInput{}, err
		}
		if name == "" {
			name = titleFromSlug(slug)
		}
		return ResolvedInput{Title: name, Slug: slug}, nil
	}

	if name != "" {
		slug = slugify(name)
		if err := validateSlug(slug, path); err != nil {
			return ResolvedInput{}, err
		}
		return ResolvedInput{Title: name, Slug: slug}, nil
	}

	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	slug = slugify(base)
	if isGenericSlug(slug) {
		return ResolvedInput{}, fmt.Errorf("prompt file %q needs a top-level name: or slug: because %q is too generic for a safe feature branch", path, base)
	}
	if err := validateSlug(slug, path); err != nil {
		return ResolvedInput{}, err
	}

	return ResolvedInput{
		Title: titleFromSlug(slug),
		Slug:  slug,
	}, nil
}

func readPromptMetadata(path string) (name string, slug string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", fmt.Errorf("read prompt file %q: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for lineNo := 0; scanner.Scan() && lineNo < 20; lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" && lineNo > 0 {
			break
		}

		switch {
		case strings.HasPrefix(strings.ToLower(line), "name:"):
			value := strings.TrimSpace(line[len("name:"):])
			if value != "" {
				name = value
			}
		case strings.HasPrefix(strings.ToLower(line), "slug:"):
			value := strings.TrimSpace(line[len("slug:"):])
			if value != "" {
				slug = slugify(value)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("scan prompt file %q: %w", path, err)
	}

	return name, slug, nil
}

func branchName(slug, prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "feature"
	}
	return prefix + "/" + slug
}

func looksLikeURL(input string) bool {
	lower := strings.ToLower(input)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func validateSlug(slug, source string) error {
	if slug == "" {
		return fmt.Errorf("spec input %q produced an empty slug", source)
	}
	if len(slug) > 60 {
		return fmt.Errorf("derived slug %q is too long; provide a shorter feature name or explicit slug:", slug)
	}
	if isGenericSlug(slug) {
		return fmt.Errorf("derived slug %q is too generic; provide a more specific feature name or explicit slug:", slug)
	}
	return nil
}

func isGenericSlug(slug string) bool {
	switch slug {
	case "prompt", "spec", "spec-prompt", "request", "input", "notes", "draft", "index", "tmp", "file", "untitled":
		return true
	default:
		return false
	}
}

func displayPath(root, target string) string {
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return target
	}
	return filepath.ToSlash(relative)
}

func slugify(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '-' || r == '_' || r == '/':
			if b.Len() > 0 && !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

func titleFromSlug(slug string) string {
	parts := strings.Split(slug, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}
