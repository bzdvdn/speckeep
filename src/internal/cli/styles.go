package cli

import (
	"io"
	"os"
	"strings"
)

const (
	ansiReset  = "\x1b[0m"
	ansiBlue   = "\x1b[34m"
	ansiCyan   = "\x1b[36m"
	ansiHiCyan = "\x1b[96m"
	ansiYellow = "\x1b[33m"
	ansiGreen  = "\x1b[32m"
	ansiRed    = "\x1b[31m"
	ansiGray   = "\x1b[90m"
)

func useColor(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func colorize(s string, code string, enabled bool) string {
	if !enabled {
		return s
	}
	return code + s + ansiReset
}

func styleCmd(w io.Writer, s string) string {
	return colorize(s, ansiCyan, useColor(w))
}

func stylePath(w io.Writer, s string) string {
	return colorize(s, ansiYellow, useColor(w))
}

func styleWarn(w io.Writer, s string) string {
	return colorize(s, ansiYellow, useColor(w))
}

func styleMuted(w io.Writer, s string) string {
	return colorize(s, ansiGray, useColor(w))
}

func styleOK(w io.Writer, s string) string {
	return colorize(s, ansiGreen, useColor(w))
}

func styleError(w io.Writer, s string) string {
	return colorize(s, ansiRed, useColor(w))
}
