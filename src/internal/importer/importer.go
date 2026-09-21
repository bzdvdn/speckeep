package importer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// Source is the origin layout being imported.
type Source string

const (
	// SourceOpenSpec imports from an OpenSpec workspace (openspec/changes/<slug>/).
	SourceOpenSpec Source = "openspec"
	// SourceSpecKit imports from a Spec Kit workspace (specs/<slug>/).
	SourceSpecKit Source = "speckit"
)

// ParseSource normalizes a user-provided source name.
func ParseSource(value string) (Source, error) {
	switch Source(strings.ToLower(strings.TrimSpace(value))) {
	case SourceOpenSpec:
		return SourceOpenSpec, nil
	case SourceSpecKit:
		return SourceSpecKit, nil
	default:
		return "", fmt.Errorf("unsupported import source %q, expected openspec or speckit", value)
	}
}

// Result summarizes one import run.
type Result struct {
	Source   Source            `json:"source"`
	Imported []ImportedFeature `json:"imported"`
	Skipped  []string          `json:"skipped,omitempty"`
}

// ImportedFeature describes one feature package written into specsDir.
type ImportedFeature struct {
	Slug  string   `json:"slug"`
	Files []string `json:"files"`
	Notes []string `json:"notes,omitempty"`

	Spec    string `json:"-"`
	Plan    string `json:"-"`
	Tasks   string `json:"-"`
	Context string `json:"-"`
}

// Import scans a foreign workspace for feature packages and writes speckeep
// artifacts under specsDir (absolute). Existing feature directories are never
// overwritten — they are reported in Skipped instead.
func Import(ctx context.Context, source Source, root, specsDir string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return Result{}, err
	}
	if specsDir == "" {
		return Result{}, fmt.Errorf("specsDir must be provided")
	}

	result := Result{Source: source}
	switch source {
	case SourceOpenSpec:
		result, err = importOpenSpec(ctx, root, specsDir, result)
	case SourceSpecKit:
		result, err = importSpecKit(ctx, root, specsDir, result)
	default:
		return Result{}, fmt.Errorf("unsupported source %q", source)
	}
	if err != nil {
		return Result{}, err
	}
	sort.Slice(result.Imported, func(i, j int) bool { return result.Imported[i].Slug < result.Imported[j].Slug })
	sort.Strings(result.Skipped)
	return result, nil
}

// importOpenSpec maps openspec/changes/<slug>/ packages.
func importOpenSpec(ctx context.Context, root, specsDir string, result Result) (Result, error) {
	changesDir := filepath.Join(root, "openspec", "changes")
	entries, err := os.ReadDir(changesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return result, err
	}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "archive" {
			continue
		}
		slug := slugify(entry.Name())
		if slug == "" {
			result.Skipped = append(result.Skipped, entry.Name()+": empty slug")
			continue
		}
		changeDir := filepath.Join(changesDir, entry.Name())
		if isDir(changeDir) && importedExists(specsDir, slug) {
			result.Skipped = append(result.Skipped, slug+": target feature already exists")
			continue
		}
		feat, err := openSpecFeature(ctx, changeDir, slug)
		if err != nil {
			result.Skipped = append(result.Skipped, slug+": "+err.Error())
			continue
		}
		if err := writeFeature(specsDir, &feat); err != nil {
			return result, err
		}
		result.Imported = append(result.Imported, feat)
	}
	return result, nil
}

// openSpecFeature converts one OpenSpec change package into speckeep artifacts.
func openSpecFeature(ctx context.Context, changeDir, slug string) (ImportedFeature, error) {
	if err := ctx.Err(); err != nil {
		return ImportedFeature{}, err
	}
	feat := ImportedFeature{Slug: slug}

	// Spec: gather requirement/scenario blocks from specs/<*.md> or spec.md.
	specContent, specNotes := translateOpenSpecSpec(ctx, changeDir, slug)
	if feat.Context != "" && specContent != "" {
		// Insert the proposal-derived context between the title and the goal.
		specContent = injectContext(specContent, feat.Context)
	}
	feat.Spec = specContent
	feat.Notes = append(feat.Notes, specNotes...)

	// Context from proposal.md (best effort, truncated).
	proposalPath := filepath.Join(changeDir, "proposal.md")
	if content, err := os.ReadFile(proposalPath); err == nil {
		feat.Context = openSpecProposalContext(string(content))
		if specContent != "" {
			feat.Spec = injectContext(feat.Spec, feat.Context)
		}
	}

	// Plan from design.md (verbatim + compliance footer).
	if design, err := os.ReadFile(filepath.Join(changeDir, "design.md")); err == nil {
		feat.Plan = string(design) + "\n\n## Constitution Compliance\n\n- imported from OpenSpec design.md (no conflicts asserted)\n"
	}

	// Tasks from tasks.md (verbatim + import note).
	if tasks, err := os.ReadFile(filepath.Join(changeDir, "tasks.md")); err == nil {
		feat.Tasks = string(tasks) + importTasksNote("OpenSpec")
	}

	if feat.Spec == "" && feat.Plan == "" && feat.Tasks == "" {
		return ImportedFeature{}, fmt.Errorf("no recognizable artifacts (spec.md/specs/, design.md, tasks.md)")
	}
	return feat, nil
}

// importSpecKit maps Spec Kit specs/<slug>/ packages (spec.md, plan.md, tasks.md).
func importSpecKit(ctx context.Context, root, specsDir string, result Result) (Result, error) {
	srcDir := filepath.Join(root, "specs")
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return result, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		slug := slugify(entry.Name())
		if slug == "" || importedExists(specsDir, slug) {
			if slug != "" {
				result.Skipped = append(result.Skipped, slug+": target feature already exists")
			}
			continue
		}
		pkgDir := filepath.Join(srcDir, entry.Name())
		feat := ImportedFeature{Slug: slug}
		if content, err := os.ReadFile(filepath.Join(pkgDir, "spec.md")); err == nil {
			feat.Spec = string(content) + "\n\n## Assumptions\n\n- imported from Spec Kit spec.md\n"
		}
		if content, err := os.ReadFile(filepath.Join(pkgDir, "plan.md")); err == nil {
			feat.Plan = string(content) + "\n\n## Constitution Compliance\n\n- imported from Spec Kit plan.md (no conflicts asserted)\n"
		}
		if content, err := os.ReadFile(filepath.Join(pkgDir, "tasks.md")); err == nil {
			feat.Tasks = string(content) + importTasksNote("Spec Kit")
		}
		if feat.Spec == "" && feat.Plan == "" && feat.Tasks == "" {
			result.Skipped = append(result.Skipped, slug+": no spec.md/plan.md/tasks.md found")
			continue
		}
		if err := writeFeature(specsDir, &feat); err != nil {
			return result, err
		}
		result.Imported = append(result.Imported, feat)
	}
	return result, nil
}

func importedExists(specsDir, slug string) bool {
	for _, name := range []string{"spec.md", "plan.md", "tasks.md"} {
		if fileExists(filepath.Join(specsDir, slug, name)) {
			return true
		}
	}
	return false
}

func writeFeature(specsDir string, feat *ImportedFeature) error {
	targetDir := filepath.Join(specsDir, feat.Slug)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}
	for _, artifact := range []struct {
		name  string
		value string
	}{{name: "spec.md", value: feat.Spec}, {name: "plan.md", value: feat.Plan}, {name: "tasks.md", value: feat.Tasks}} {
		if artifact.value == "" {
			continue
		}
		target := filepath.Join(targetDir, artifact.name)
		if err := os.WriteFile(target, []byte(artifact.value), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", artifact.name, err)
		}
		feat.Files = append(feat.Files, filepath.ToSlash(filepath.Join(filepath.Base(specsDir), feat.Slug, artifact.name)))
	}
	return nil
}

func importTasksNote(source string) string {
	return "\n\n<!-- IMPORTED from " + source + " — best-effort. Run /spk-tasks to regenerate with Touches:, ## Surface Map and ## Acceptance Coverage. -->\n"
}

func translateOpenSpecSpec(ctx context.Context, changeDir, slug string) (string, []string) {
	files := openSpecSpecFiles(changeDir)
	if len(files) == 0 {
		return "", nil
	}
	var (
		reqs      []requirement
		scenOrder []string
		seen      = map[string]string{}
	)
	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		parseOpenSpecRequirements(string(content), &reqs, &scenOrder, seen)
	}
	if len(reqs) == 0 {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", titleFromSlug(slug)))
	sb.WriteString("## Goal\n\n")
	if len(scenOrder) > 0 {
		sb.WriteString("Deliver: " + strings.Join(scenOrder, "; ") + ".\n")
	} else {
		sb.WriteString("Imported feature.\n")
	}
	sb.WriteString("\n## Requirements\n\n")
	for i, req := range reqs {
		sb.WriteString(fmt.Sprintf("- RQ-%03d %s: %s\n", i+1, req.title, firstLine(req.body)))
	}
	sb.WriteString("\n## Acceptance Criteria\n\n")
	k := 0
	for _, title := range scenOrder {
		sc := seen[title]
		if sc == "" {
			continue
		}
		k++
		sb.WriteString(fmt.Sprintf("\n### AC-%03d %s\n\n", k, title))
		given, when, then := splitScenario(sc)
		sb.WriteString(fmt.Sprintf("- **Given** %s\n", given))
		sb.WriteString(fmt.Sprintf("- **When** %s\n", when))
		if then != "" {
			sb.WriteString(fmt.Sprintf("- **Then** %s\n", then))
		}
	}
	sb.WriteString("\n## Assumptions\n\n")
	sb.WriteString("- Imported from OpenSpec (requirements/scenarios). Run `/spk-inspect` to deepen quality.\n")
	notes := []string{"spec.md rebuilt from OpenSpec Requirement/Scenario blocks; validate AC wording"}
	return sb.String(), notes
}

type requirement struct {
	title string
	body  string
}

func parseOpenSpecRequirements(content string, reqs *[]requirement, scenOrder *[]string, seen map[string]string) {
	lines := strings.Split(content, "\n")
	var curReq *requirement
	curScenario := ""
	for _, raw := range lines {
		line := raw
		switch {
		case strings.HasPrefix(line, "### Requirement"):
			title := strings.TrimSpace(strings.TrimPrefix(line, "### Requirement"))
			title = strings.TrimPrefix(title, ":")
			title = strings.TrimSpace(title)
			*reqs = append(*reqs, requirement{title: title})
			curReq = &(*reqs)[len(*reqs)-1]
			curScenario = ""
		case strings.HasPrefix(line, "#### Scenario"):
			title := strings.TrimSpace(strings.TrimPrefix(line, "#### Scenario"))
			title = strings.TrimPrefix(title, ":")
			title = strings.TrimSpace(title)
			curScenario = title
			*scenOrder = append(*scenOrder, title)
			seen[title] = ""
			curReq = nil
		case strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "# "):
			curReq = nil
			curScenario = ""
		case curScenario != "" && strings.Contains(line, "**"):
			seen[curScenario] = seen[curScenario] + "\n" + line
		case curReq != nil && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "-"):
			curReq.body += " " + strings.TrimSpace(line)
		}
	}
}

func splitScenario(block string) (given, when, then string) {
	given, when, then = "the feature is active", "", ""
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimLeft(strings.TrimSpace(line), "- ")
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		switch {
		case strings.HasPrefix(lower, "**when**"):
			when = stripScenarioMarker(trimmed)
		case strings.HasPrefix(lower, "**then**"):
			then = stripScenarioMarker(trimmed)
		case strings.HasPrefix(lower, "**given**"):
			given = stripScenarioMarker(trimmed)
		}
	}
	return given, when, then
}

// stripScenarioMarker removes a `**WHEN**`/`**When**` style leading marker
// case-insensitively, plus any trailing colon.
func stripScenarioMarker(line string) string {
	for _, marker := range []string{"**when**", "**given**", "**then**"} {
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, marker) {
			line = strings.TrimSpace(line[len(marker):])
			break
		}
	}
	for _, candidate := range []string{":", "-"} {
		if strings.HasPrefix(line, candidate) {
			line = strings.TrimSpace(line[len(candidate):])
		}
	}
	return line
}

// openSpecSpecFiles lists the spec sources for a change: specs/<*.md> else spec.md.
func openSpecSpecFiles(changeDir string) []string {
	specsDir := filepath.Join(changeDir, "specs")
	if entries, err := os.ReadDir(specsDir); err == nil && len(entries) > 0 {
		var files []string
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				files = append(files, filepath.Join(specsDir, e.Name()))
			}
		}
		if len(files) > 0 {
			sort.Strings(files)
			return files
		}
	}
	if fileExists(filepath.Join(changeDir, "spec.md")) {
		return []string{filepath.Join(changeDir, "spec.md")}
	}
	return nil
}

func openSpecProposalContext(content string) string {
	lines := strings.Split(content, "\n")
	out := []string{}
	inBody := false
	skipLines := 0
	for _, rawLine := range lines {
		line := strings.TrimRight(rawLine, " \t\r")
		if strings.HasPrefix(line, "---") {
			if !inBody {
				inBody = true
				skipLines = 1
				continue
			}
			inBody = false
			continue
		}
		if skipLines > 0 {
			skipLines--
			continue
		}
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "# why", "## why", "# what changes", "## what changes", "# proposal":
			continue
		}
		out = append(out, line)
		if len(out) >= 20 {
			out = append(out, "…")
			break
		}
	}
	context := strings.TrimSpace(strings.Join(out, "\n"))
	if context == "" {
		return ""
	}
	return "\n## Context\n\n" + context + "\n"
}

func firstLine(value string) string {
	value = strings.TrimSpace(value)
	if idx := strings.IndexByte(value, '\n'); idx >= 0 {
		value = value[:idx]
	}
	return strings.TrimSpace(value)
}

// injectContext inserts a `## Context` section right after the H1 title,
// before any other top-level section.
func injectContext(spec, context string) string {
	if context == "" {
		return spec
	}
	lines := strings.SplitN(spec, "\n", 2)
	if !strings.HasPrefix(lines[0], "# ") {
		return context + "\n" + spec
	}
	return lines[0] + "\n" + context + "\n" + lines[1]
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
		case r == ' ' || r == '-' || r == '_':
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
	for i := range parts {
		if parts[i] != "" {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, " ")
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
