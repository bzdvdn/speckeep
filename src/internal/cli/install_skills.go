package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"speckeep/src/internal/project"
)

func newInstallSkillsCmd() *cobra.Command {
	return newSkillsInstallCmd("install-skills [path]", "Install configured skills into agent skill directories")
}

func newSkillsInstallSubCmd() *cobra.Command {
	return newSkillsInstallCmd("install [path]", "Install configured skills into agent skill directories")
}

func newSkillsInstallCmd(use, short string) *cobra.Command {
	var dryRun bool
	var jsonOutput bool
	var targets string
	var includeDisabled bool

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}

			var targetList []string
			if strings.TrimSpace(targets) != "" {
				targetList = splitCommaSeparated(targets)
			}
			result, err := project.InstallSkills(root, project.InstallSkillsOptions{
				Targets:         targetList,
				DryRun:          dryRun,
				IncludeDisabled: includeDisabled,
			})
			if err != nil {
				return err
			}

			if jsonOutput {
				payload, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
				return nil
			}

			for _, line := range result.Messages {
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			printPanel(cmd.OutOrStdout(), "skills install summary", []string{
				fmt.Sprintf("created: %d", len(result.Created)),
				fmt.Sprintf("updated: %d", len(result.Updated)),
				fmt.Sprintf("removed: %d", len(result.Removed)),
				fmt.Sprintf("unchanged: %d", len(result.Unchanged)),
				fmt.Sprintf("dry-run: %t", result.DryRun),
			})
			return nil
		},
	}

	cmd.Flags().StringVar(&targets, "targets", "", "comma-separated agent targets (default: targets from speckeep.yaml)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show which target skill directories would change without writing them")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output install result as JSON")
	cmd.Flags().BoolVar(&includeDisabled, "include-disabled", false, "Install disabled skills too")
	return cmd
}

func splitCommaSeparated(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}
