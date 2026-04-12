package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"speckeep/src/internal/config"
)

func newSchemaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema [path]",
		Short: "Show or customize the artifact dependency graph",
		Long: `Display or customize the artifact dependency graph for the project.

A schema defines which artifacts are required for each phase and in what order.
This allows teams to adapt the workflow without modifying core templates.

Examples:
  speckeep schema
  speckeep schema --show
  speckeep schema --set research-first
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

			schema := cfg.Workflow.Schema
			if schema == "" {
				schema = "default"
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Current schema: %s\n\n", schema)

			phases := getSchemaPhases(schema)
			fmt.Fprintf(cmd.OutOrStdout(), "Phase order:\n")
			for i, phase := range phases {
				fmt.Fprintf(cmd.OutOrStdout(), "  %d. %s\n", i+1, phase)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\nAvailable schemas:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  default       - constitution -> spec -> inspect -> plan -> tasks -> implement -> verify -> archive\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  research-first - constitution -> spec -> inspect -> plan -> research -> tasks -> implement -> verify -> archive\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  lite          - spec -> plan -> tasks -> implement (no inspect/verify)\n")

			return nil
		},
	}

	var setSchema string
	cmd.Flags().StringVar(&setSchema, "set", "", "Set the project schema to a predefined type")
	cmd.Flags().Bool("show", false, "Show the current schema configuration")

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if setSchema != "" {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}

			cfg, err := config.Load(root)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			validSchemas := []string{"default", "research-first", "lite"}
			valid := false
			for _, s := range validSchemas {
				if setSchema == s {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid schema %q, expected one of: %s", setSchema, strings.Join(validSchemas, ", "))
			}

			if cfg.Workflow.Schema == setSchema {
				fmt.Fprintf(cmd.OutOrStdout(), "Schema already set to %s\n", setSchema)
				return nil
			}

			cfg.Workflow.Schema = setSchema
			if err := config.Save(root, cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Schema set to: %s\n", setSchema)
		}
		return nil
	}

	return cmd
}

func getSchemaPhases(schema string) []string {
	switch schema {
	case "research-first":
		return []string{"constitution", "spec", "inspect", "plan", "research", "tasks", "implement", "verify", "archive"}
	case "lite":
		return []string{"spec", "plan", "tasks", "implement"}
	default:
		return []string{"constitution", "spec", "inspect", "plan", "tasks", "implement", "verify", "archive"}
	}
}
