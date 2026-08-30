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

// elideMiddle shortens s to at most maxWidth display columns, keeping the
// start and end readable and removing the middle with the given ellipsis.
// Strings already within the limit (or with no room for an ellipsis) are
// returned unchanged.
func elideMiddle(s string, maxWidth int, ellipsis string) string {
	w := ansi.StringWidth(s)
	e := ansi.StringWidth(ellipsis)
	if w <= maxWidth || e >= maxWidth {
		return s
	}
	keep := maxWidth - e
	left := (keep + 1) / 2
	right := keep - left
	return ansi.Truncate(s, left, "") + ellipsis + ansi.TruncateLeft(s, w-right, "")
}