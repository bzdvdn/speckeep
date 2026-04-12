package cli

import (
	"github.com/spf13/cobra"
	"speckeep/src/internal/project"
)

func newInitCmd() *cobra.Command {
	var initGit bool
	var defaultLang string
	var docsLang string
	var agentLang string
	var commentsLang string
	var shell string
	var specsDir string
	var archiveDir string
	var constitutionFile string
	var agentTargets []string
	var legacyAgentTargets []string

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a .speckeep workspace in the target project",
		Long: `Initializes a SpecKeep workspace inside the target project.

Creates the .speckeep/ directory structure (specs/archive/templates/scripts), inserts/updates a managed block in AGENTS.md, and (optionally) generates agent-target artifacts.

Notes:
  - Template files are created only if missing (existing files are kept).
  - The managed SpecKeep block in AGENTS.md is inserted/updated automatically.`,
		Example: "  speckeep init . --lang en --shell sh --agents codex\n  speckeep init /path/to/repo --git=false --lang en --shell sh",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Print(renderSpecgateBanner(cmd))

			root := "."
			if len(args) == 1 {
				root = args[0]
			}

			result, err := project.Initialize(root, project.InitOptions{
				InitGit:          initGit,
				DefaultLang:      defaultLang,
				DocsLang:         docsLang,
				AgentLang:        agentLang,
				CommentsLang:     commentsLang,
				Shell:            shell,
				SpecsDir:         specsDir,
				ArchiveDir:       archiveDir,
				ConstitutionFile: constitutionFile,
				AgentTargets:     append(agentTargets, legacyAgentTargets...),
			})
			if err != nil {
				return err
			}

			printInitOutput(cmd.OutOrStdout(), result)

			return nil
		},
	}

	cmd.Flags().BoolVar(&initGit, "git", true, "initialize a git repository when one does not exist")
	cmd.Flags().StringVar(&defaultLang, "lang", "en", "base language for generated docs and prompts: en or ru")
	cmd.Flags().StringVar(&docsLang, "docs-lang", "", "language for generated project documentation: en or ru")
	cmd.Flags().StringVar(&agentLang, "agent-lang", "", "language for generated agent prompts and AGENTS guidance: en or ru")
	cmd.Flags().StringVar(&commentsLang, "comments-lang", "", "preferred language for code comments: en or ru")
	cmd.Flags().StringVar(&shell, "shell", "", "shell for generated workflow scripts: sh or powershell")
	cmd.Flags().StringVar(&specsDir, "specs-dir", "", "override specs directory (advanced): e.g. .speckeep/specs")
	cmd.Flags().StringVar(&archiveDir, "archive-dir", "", "override archive directory (advanced): e.g. .speckeep/archive")
	cmd.Flags().StringVar(&constitutionFile, "constitution-file", "", "override constitution file path (advanced): e.g. .speckeep/constitution.md or docs/constitution.md")
	cmd.Flags().StringSliceVar(&agentTargets, "agents", nil, "generate project-local agent command files for one or more targets: claude, codex, copilot, cursor, kilocode, trae, all")
	cmd.Flags().StringSliceVar(&legacyAgentTargets, "agent", nil, "deprecated alias for --agents")
	cmd.Flags().MarkHidden("agent")
	cmd.MarkFlagRequired("shell")

	return cmd
}
