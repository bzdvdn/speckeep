package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"speckeep/src/internal/project"
)

func newAddSkillCmd() *cobra.Command {
	var id string
	var fromLocal string
	var fromGit string
	var ref string
	var skillPath string
	var version string
	var disabled bool
	var noInstall bool

	cmd := &cobra.Command{
		Use:   "add-skill [path]",
		Short: "Add or update a skill in the project skills manifest",
		Long: `Adds or updates a skill in .speckeep/skills/manifest.yaml.

For git sources, this command clones/checks out the skill under .speckeep/skills/checkouts/<id>.

For git sources, --ref is required to pin the skill version and keep installs reproducible.`,
		Example: "  speckeep add-skill . --id architecture --from-local skills/architecture\n  speckeep add-skill . --id openai-docs --from-git https://example.com/skills.git --ref v1.2.3 --path skills/openai-docs",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}

			result, err := project.AddSkill(root, project.AddSkillOptions{
				ID:        id,
				FromLocal: fromLocal,
				FromGit:   fromGit,
				Ref:       ref,
				Path:      skillPath,
				Version:   version,
				Enabled:   !disabled,
				NoInstall: noInstall,
			})
			if err != nil {
				return err
			}

			for _, line := range result.Messages {
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "skill id (required): lowercase letters, digits, _ or -")
	cmd.Flags().StringVar(&fromLocal, "from-local", "", "local skill directory path")
	cmd.Flags().StringVar(&fromGit, "from-git", "", "git repository url for the skill source")
	cmd.Flags().StringVar(&ref, "ref", "", "pinned git ref (tag or commit) for --from-git")
	cmd.Flags().StringVar(&skillPath, "path", ".", "path inside skill source repository/directory")
	cmd.Flags().StringVar(&version, "version", "", "skill version label")
	cmd.Flags().BoolVar(&disabled, "disabled", false, "add skill as disabled")
	cmd.Flags().BoolVar(&noInstall, "no-install", false, "do not install skill into agent folders after manifest update")
	cmd.MarkFlagRequired("id")

	return cmd
}
