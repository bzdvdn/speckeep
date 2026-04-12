package main

import (
	"errors"
	"fmt"
	"os"

	"speckeep/src/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		var exitErr interface{ ExitCode() int }
		if errors.As(err, &exitErr) {
			if err.Error() != "" {
				fmt.Fprintln(os.Stderr, err)
			}
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
