package ui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// truncate truncates s to at most width display columns, keeping ANSI escape
// sequences intact.
func truncate(s string, width int) string {
	return ansi.Truncate(s, width, "")
}

// padRight pads s to at least width display columns.
func padRight(s string, width int) string {
	gap := width - ansi.StringWidth(s)
	if gap <= 0 {
		return s
	}
	return s + strings.Repeat(" ", gap)
}