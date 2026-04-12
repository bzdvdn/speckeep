package cli

import (
	"fmt"
	"io"
	"strings"

	"speckeep/src/internal/project"
)

func printInitOutput(w io.Writer, result project.InitResult) {
	color := useColor(w)

	agentTargets := "none"
	if len(result.AgentTargets) > 0 {
		agentTargets = strings.Join(result.AgentTargets, ", ")
	}

	fmt.Fprintf(w, "Selected script type: %s\n", result.Shell)
	fmt.Fprintf(w, "Configured languages: docs=%s agent=%s comments=%s\n", result.DocsLang, result.AgentLang, result.CommentsLang)
	if result.SpecsDir != "" || result.ArchiveDir != "" {
		fmt.Fprintf(w, "Configured paths: constitution=%s specs=%s archive=%s\n", result.ConstitutionFile, result.SpecsDir, result.ArchiveDir)
	}
	fmt.Fprintf(w, "enabled agent targets: %s\n\n", agentTargets)

	fmt.Fprintln(w, "Initialize SpecKeep Project")
	steps := initSteps(result)
	for i, step := range steps {
		last := i == len(steps)-1
		printStepLine(w, step, last, color)
	}
	fmt.Fprintln(w)

	if color {
		fmt.Fprintln(w, styleOK(w, "Project ready."))
	} else {
		fmt.Fprintln(w, "Project ready.")
	}
	fmt.Fprintln(w)

	printPanel(w, "Agent Folder Security", []string{
		"Some agents may store credentials, auth tokens, or other private artifacts",
		"inside project-level agent folders (e.g. " + stylePath(w, ".claude/") + ", " + stylePath(w, ".cursor/") + ", " + stylePath(w, ".kilocode/") + ").",
		"Consider adding them (or parts of them) to .gitignore to prevent leaks.",
	})

	printPanel(w, "Next Steps", []string{
		"1. You're already in the project directory.",
		"2. Start using slash commands with your AI agent:",
		"   2.1 " + styleCmd(w, "/speckeep.constitution") + "  - Establish project principles",
		"   2.2 " + styleCmd(w, "/speckeep.spec") + "          - Create a baseline specification",
		"   2.3 " + styleCmd(w, "/speckeep.plan") + "          - Create an implementation plan",
		"   2.4 " + styleCmd(w, "/speckeep.tasks") + "         - Generate actionable tasks",
		"   2.5 " + styleCmd(w, "/speckeep.implement") + "     - Execute implementation",
	})

	printPanel(w, "Enhancement Commands", []string{
		styleCmd(w, "/speckeep.challenge") + " (optional) - adversarial review of spec/plan",
		styleCmd(w, "/speckeep.scope") + " (optional)     - scope boundary check",
		styleCmd(w, "/speckeep.recap") + " (optional)     - recap active features",
	})

	printPanel(w, "Useful CLI Commands", []string{
		styleCmd(w, "speckeep doctor .") + "        - workspace health check",
		styleCmd(w, "speckeep list-specs .") + "    - list active specs",
		styleCmd(w, "speckeep check <slug> .") + "  - feature status",
		styleCmd(w, "speckeep dashboard .") + "     - visual dashboard",
	})
}

type initStep struct {
	Label  string
	Status string // ok, kept, updated, skipped
	Detail string
}

func initSteps(result project.InitResult) []initStep {
	createdCount := len(result.Created)
	keptCount := len(result.Kept)

	templateDetail := fmt.Sprintf("%d created, %d kept", createdCount, keptCount)

	agentsStatus := "skipped"
	agentsDetail := "no targets"
	if len(result.AgentTargets) > 0 {
		agentsStatus = "ok"
		agentsDetail = strings.Join(result.AgentTargets, ", ")
	}

	agentsSnippetStatus := "kept"
	if result.AgentsSnippetChanged {
		agentsSnippetStatus = "updated"
	}

	gitStatus := result.GitRepoStatus
	if gitStatus == "" {
		gitStatus = "ok"
	}

	return []initStep{
		{Label: "Ensure .speckeep directory structure", Status: "ok"},
		{Label: "Install managed templates", Status: "ok", Detail: templateDetail},
		{Label: "Link SpecKeep block in AGENTS.md", Status: agentsSnippetStatus},
		{Label: "Generate agent target artifacts", Status: agentsStatus, Detail: agentsDetail},
		{Label: "Initialize git repository", Status: gitStatus},
		{Label: "Finalize", Status: "ok", Detail: "project ready"},
	}
}

func printStepLine(w io.Writer, step initStep, last bool, color bool) {
	connector := "├─"
	if last {
		connector = "└─"
	}

	status := renderStepStatus(step.Status, color)
	suffix := renderStepSuffix(step.Status, step.Detail)
	fmt.Fprintf(w, "  %s %s %s (%s)\n", connector, status, step.Label, suffix)
}

func renderStepStatus(status string, color bool) string {
	switch status {
	case "ok":
		return colorize("●", ansiGreen, color)
	case "updated":
		return colorize("●", ansiCyan, color)
	case "kept":
		return colorize("●", ansiYellow, color)
	case "skipped":
		return colorize("○", ansiGray, color)
	case "initialized":
		return colorize("●", ansiGreen, color)
	default:
		return colorize("●", ansiGreen, color)
	}
}

func renderStepSuffix(status string, detail string) string {
	label := status
	if status == "initialized" {
		label = "ok"
	}
	if detail == "" {
		return label
	}
	return label + ": " + detail
}
