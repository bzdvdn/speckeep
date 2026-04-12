package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"speckeep/src/internal/specs"
)

func newListSpecsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-specs [path]",
		Short: "List available specifications",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}

			names, err := specs.List(root)
			if err != nil {
				return err
			}

			for _, name := range names {
				fmt.Fprintln(cmd.OutOrStdout(), name)
			}

			return nil
		},
	}
}
