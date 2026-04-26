package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"speckeep/src/internal/project"
)

func newListSkillsCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list-skills [path]",
		Short: "List configured skills from the project skills manifest",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}

			result, err := project.ListSkills(root)
			if err != nil {
				return err
			}

			if jsonOutput {
				payload, err := json.MarshalIndent(struct {
					Skills interface{} `json:"skills"`
				}{Skills: result.Skills}, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
				return nil
			}

			if len(result.Skills) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no skills configured")
				return nil
			}
			for _, skill := range result.Skills {
				status := "enabled"
				if !skill.Enabled {
					status = "disabled"
				}
				location := skill.Location
				if skill.Source == "git" && skill.Ref != "" {
					location = skill.Location + "@" + skill.Ref
				}
				line := fmt.Sprintf("%s\t%s\t%s", skill.ID, status, location)
				if strings.TrimSpace(skill.Path) != "" {
					line += "\tpath=" + skill.Path
				}
				if strings.TrimSpace(skill.CheckoutDir) != "" {
					line += "\tcheckout=" + skill.CheckoutDir
				}
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output skills as JSON")
	return cmd
}
