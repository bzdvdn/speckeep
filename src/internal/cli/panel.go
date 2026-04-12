package cli

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func printPanel(w io.Writer, title string, body []string) {
	maxLen := visibleRuneLen(title)
	for _, line := range body {
		if l := visibleRuneLen(line); l > maxLen {
			maxLen = l
		}
	}

	border := strings.Repeat("═", maxLen+2)
	fmt.Fprintf(w, "╔%s╗\n", border)
	fmt.Fprintf(w, "║ %s%s ║\n", title, strings.Repeat(" ", maxLen-visibleRuneLen(title)))
	for _, line := range body {
		fmt.Fprintf(w, "║ %s%s ║\n", line, strings.Repeat(" ", maxLen-visibleRuneLen(line)))
	}
	fmt.Fprintf(w, "╚%s╝\n\n", border)
}

func visibleRuneLen(s string) int {
	return utf8.RuneCountInString(stripANSI(s))
}

func stripANSI(s string) string {
	if strings.IndexByte(s, 0x1b) < 0 {
		return s
	}
	return ansiRegexp.ReplaceAllString(s, "")
}
