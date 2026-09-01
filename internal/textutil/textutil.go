// Package textutil provides small text helpers shared across the TUI, store,
// and converter packages.
package textutil

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// FormatSize renders a byte count with a binary unit suffix: whole bytes for
// values under 1 KB, otherwise one decimal place (e.g. "450 B", "1.4 MB").
func FormatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	size := float64(bytes)
	for _, u := range []string{"KB", "MB", "GB", "TB"} {
		size /= 1024
		if size < 1024 || u == "TB" {
			return fmt.Sprintf("%.1f %s", size, u)
		}
	}
	return fmt.Sprintf("%.1f TB", size)
}

// EscapeMarkdown backslash-escapes the characters that CommonMark treats as
// special inline (emphasis, code spans, links, autolinks) so the string
// renders literally.
func EscapeMarkdown(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\', '`', '*', '_', '[', ']', '<', '>':
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// MaxLineWidth returns the maximum cell width across the given ANSI-styled
// lines. A nil or empty slice returns 0.
func MaxLineWidth(lines []string) int {
	w := 0
	for _, l := range lines {
		if lw := ansi.StringWidth(l); lw > w {
			w = lw
		}
	}
	return w
}
