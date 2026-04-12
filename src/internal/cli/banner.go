package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const speckeepASCII = `
 ██████╗ ██████╗ ███████╗ ██████╗██╗  ██╗███████╗███████╗██████╗
██╔════╝ ██╔══██╗██╔════╝██╔════╝██║ ██╔╝██╔════╝██╔════╝██╔══██╗
███████╗ ██████╔╝█████╗  ██║     █████╔╝ █████╗  █████╗  ██████╔╝
╚════██║ ██╔═══╝ ██╔══╝  ██║     ██╔═██╗ ██╔══╝  ██╔══╝  ██╔═══╝
███████║ ██║     ███████╗╚██████╗██║  ██╗███████╗███████╗██║
╚══════╝ ╚═╝     ╚══════╝ ╚═════╝╚═╝  ╚═╝╚══════╝╚══════╝╚═╝
`

func init() {
	cobra.AddTemplateFunc("speckeepBanner", func(cmd *cobra.Command) string {
		return renderSpecgateBanner(cmd)
	})
}

func renderSpecgateBanner(cmd *cobra.Command) string {
	out := cmd.OutOrStdout()
	color := useColor(out)

	art := strings.TrimLeft(speckeepASCII, "\n") + "\n"
	tagline := "SpecKeep — Spec-Driven Development Toolkit\n"
	if cmd != nil && cmd.Root() != nil && cmd != cmd.Root() {
		tagline = fmt.Sprintf("SpecKeep — %s\n", cmd.CommandPath())
	}

	if !color {
		return art + tagline + "\n"
	}

	return renderSpeckeepGradient(art) + ansiYellow + tagline + ansiReset + "\n"
}

func renderSpeckeepGradient(art string) string {
	lines := strings.Split(strings.TrimRight(art, "\n"), "\n")
	if len(lines) == 0 {
		return art
	}

	gradient := []string{
		ansiBlue,
		ansiBlue,
		ansiCyan,
		ansiCyan,
		ansiHiCyan,
		ansiHiCyan,
	}

	var b strings.Builder
	for idx, line := range lines {
		code := gradient[len(gradient)-1]
		if idx < len(gradient) {
			code = gradient[idx]
		}
		b.WriteString(code)
		b.WriteString(line)
		b.WriteString(ansiReset)
		b.WriteString("\n")
	}
	return b.String()
}
