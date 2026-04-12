package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"speckeep/src/internal/config"
)

func newExploreCmd() *cobra.Command {
	var slug string
	var branch string
	var name string

	cmd := &cobra.Command{
		Use:   "explore [path]",
		Short: "Create an exploration workspace for unstructured investigation",
		Long: `Create a lightweight exploration workspace before committing to a full spec.
Use this when requirements are unclear and you need to investigate before writing a spec.

Examples:
  speckeep explore .
  speckeep explore . --name "Investigate auth performance"
  speckeep explore . --slug "auth-perf" --name "Auth performance investigation"
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}

			cfg, err := config.Load(root)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if slug == "" {
				if branch == "" {
					branch, err = getCurrentBranch(root)
					if err != nil {
						branch = "explore"
					}
				}
				slug = deriveSlugFromBranch(branch)
				if slug == "" {
					slug = fmt.Sprintf("explore-%s", time.Now().Format("2006-01-02"))
				}
			}

			if name == "" {
				name = fmt.Sprintf("Exploration: %s", slug)
			}

			specsDir, err := cfg.SpecsDir(root)
			if err != nil {
				return err
			}

			featureDir := filepath.Join(specsDir, slug)
			if err := os.MkdirAll(featureDir, 0o755); err != nil {
				return err
			}

			explorePath := filepath.Join(featureDir, "explore.md")
			if _, err := os.Stat(explorePath); err == nil {
				return fmt.Errorf("exploration already exists at %s", explorePath)
			}

			exploreContent := generateExploreTemplate(name, slug, cfg.Language.Docs)
			if err := os.WriteFile(explorePath, []byte(exploreContent), 0o644); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created exploration workspace:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Slug: %s\n", slug)
			fmt.Fprintf(cmd.OutOrStdout(), "  File: %s\n", relPath(root, explorePath))
			fmt.Fprintf(cmd.OutOrStdout(), "\nNext steps:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  1. Use /speckeep.explore %s to investigate\n", slug)
			fmt.Fprintf(cmd.OutOrStdout(), "  2. When ready, call /speckeep.spec %s to write the spec\n", slug)
			fmt.Fprintf(cmd.OutOrStdout(), "  3. Or call /speckeep.spec %s --from-explore to convert exploration to spec\n", slug)

			return nil
		},
	}

	cmd.Flags().StringVar(&slug, "slug", "", "Feature slug (default: derived from branch)")
	cmd.Flags().StringVar(&branch, "branch", "", "Feature branch name")
	cmd.Flags().StringVar(&name, "name", "", "Exploration name")

	return cmd
}

func generateExploreTemplate(name, slug, language string) string {
	if language == "ru" {
		return fmt.Sprintf(`# %s

**Slug:** %s
**Created:** %s
**Status:** exploring

## Цель исследования

Опишите, что нужно исследовать и почему это важно.

## Ключевые вопросы

- Какие существуют варианты реализации?
- Какие есть ограничения и зависимости?
- Какие риски и неизвестные факторы?

## Найденные факты

### Факт 1: [Заголовок]

**Описание:** Что обнаружено
**Источник:** Где найдено (код, документация, тесты)
**Влияние:** Как это влияет на реализацию

### Факт 2: [Заголовок]

**Описание:** 
**Источник:** 
**Влияние:** 

## Варианты реализации

### Вариант A: [Название]

**Плюсы:**
- 

**Минусы:**
- 

**Сложность:** [низкая/средняя/высокая]

### Вариант B: [Название]

**Плюсы:**
- 

**Минусы:**
- 

**Сложность:** [низкая/средняя/высокая]

## Рекомендация

Какой вариант предпочтителен и почему:

## Готовность к спецификации

- [ ] Ключевые вопросы исследованы
- [ ] Варианты реализации оценены
- [ ] Рекомендация сформулирована
- [ ] Готов перейти к /speckeep.spec

## Связанные артефакты

- Spec: specs/%s/spec.md (не создан)
- Plan: specs/%s/plan/plan.md (не создан)
`, name, slug, time.Now().Format("2006-01-02"), slug, slug)
	}

	return fmt.Sprintf(`# %s

**Slug:** %s
**Created:** %s
**Status:** exploring

## Investigation Goal

Describe what needs to be investigated and why it matters.

## Key Questions

- What implementation options exist?
- What are the constraints and dependencies?
- What are the risks and unknowns?

## Discovered Facts

### Fact 1: [Title]

**Description:** What was discovered
**Source:** Where found (code, docs, tests)
**Impact:** How this affects implementation

### Fact 2: [Title]

**Description:** 
**Source:** 
**Impact:** 

## Implementation Options

### Option A: [Name]

**Pros:**
- 

**Cons:**
- 

**Complexity:** [low/medium/high]

### Option B: [Name]

**Pros:**
- 

**Cons:**
- 

**Complexity:** [low/medium/high]

## Recommendation

Which option is preferred and why:

## Readiness for Specification

- [ ] Key questions investigated
- [ ] Implementation options evaluated
- [ ] Recommendation formulated
- [ ] Ready to proceed to /speckeep.spec

## Related Artifacts

- Spec: specs/%s/spec.md (not created)
- Plan: specs/%s/plan/plan.md (not created)
`, name, slug, time.Now().Format("2006-01-02"), slug, slug)
}

func deriveSlugFromBranch(branch string) string {
	if branch == "" || branch == "main" || branch == "master" {
		return ""
	}
	parts := strings.Split(branch, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return branch
}

func getCurrentBranch(root string) (string, error) {
	gitDir := filepath.Join(root, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return "", fmt.Errorf("not a git repository")
	}

	headPath := filepath.Join(gitDir, "HEAD")
	content, err := os.ReadFile(headPath)
	if err != nil {
		return "", err
	}

	head := strings.TrimSpace(string(content))
	if strings.HasPrefix(head, "ref: ") {
		ref := strings.TrimPrefix(head, "ref: ")
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1], nil
		}
	}

	return "", fmt.Errorf("could not determine branch")
}

func relPath(root, target string) string {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return target
	}
	return rel
}
