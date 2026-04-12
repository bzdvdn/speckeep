package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newHelpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "help [command]",
		Short: "Show help for any command",
		Long: `Shows help for any speckeep command.

Examples:
  speckeep help
  speckeep help init
  speckeep help refresh`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _, err := cmd.Root().Find(args)
			if target == nil || err != nil {
				unknown := strings.Join(args, " ")
				if unknown == "" {
					unknown = "<root>"
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "Unknown help topic %q\n", unknown)
				_ = cmd.Root().Usage()
				return newExitError(1, "")
			}

			title := "speckeep help"
			if target != cmd.Root() {
				title = "speckeep help: " + target.CommandPath()
			}
			printPanel(cmd.OutOrStdout(), title, []string{
				"Tip: add " + styleCmd(cmd.OutOrStdout(), "--help") + " to any command.",
			})

			target.InitDefaultHelpFlag()
			target.InitDefaultVersionFlag()
			return target.Help()
		},
	}
	return cmd
}
