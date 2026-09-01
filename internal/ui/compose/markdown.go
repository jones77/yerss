package compose

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
	xansi "github.com/charmbracelet/x/ansi"

	"yerss/internal/ui/render"
)

// GlamourStandardStyle maps the configured display theme to a glamour standard
// style. "auto" resolves via the same light-background detection that drives
// the TUI palette, so the markdown and the rest of the UI agree.
func GlamourStandardStyle(theme string) string {
	switch theme {
	case "light":
		return "light"
	case "dark":
		return "dark"
	default:
		if render.DetectLightBackground() {
			return "light"
		}
		return "dark"
	}
}

// GlamourStyleConfig returns the built-in dark/light style config with the
// document's left margin removed, so the reader's own padX padding is the only
// inset from the border (otherwise glamour's default two-column margin stacks
// on top of it). Links drop their underline.
func GlamourStyleConfig(style string) ansi.StyleConfig {
	var cfg ansi.StyleConfig
	switch style {
	case "light":
		cfg = styles.LightStyleConfig
	default:
		cfg = styles.DarkStyleConfig
	}
	zero := uint(0)
	cfg.Document.Margin = &zero
	// Pin the base body text to the standard text role: ANSI 7 white in dark
	// mode, ANSI 0 black in light mode, overriding glamour's default base
	// foreground. Styled elements (headings, bold, links, blockquotes) keep
	// their own theme colors.
	textColor := "7"
	if style == "light" {
		textColor = "0"
	}
	cfg.Document.Color = &textColor
	truePtr := true
	falsePtr := false
	cfg.Link.Underline = &falsePtr
	// Drop the visible "#"/"##"/"###" markers that glamour's built-in heading
	// styles prepend to each level, so article subtitles render as clean
	// highlighted lines instead of raw markdown. The heading text keeps its
	// own color/bold styling.
	cfg.Heading.Prefix = ""
	cfg.H1.Prefix = ""
	cfg.H2.Prefix = ""
	cfg.H3.Prefix = ""
	cfg.H4.Prefix = ""
	cfg.H5.Prefix = ""
	cfg.H6.Prefix = ""
	// H2 and H3 both inherit the base Heading's bold, so give the depths a
	// weight difference: H2 stays bold, H3 renders plain (same accent color).
	cfg.H2.Bold = &truePtr
	cfg.H3.Bold = &falsePtr
	return cfg
}

// IndentListContinuations shifts the wrapped continuation lines of list items
// two columns to the right so they align under the text that follows the
// bullet or number. Glamour wraps list content inside the same block as its
// marker, so every level's continuation lands on the marker column; adding two
// cells re-aligns it with the item's text. The article frame later truncates
// each line back to the content width, so no width bookkeeping is needed here.
func IndentListContinuations(rendered string) string {
	const indent = "  "
	prevItem := false
	lines := strings.Split(rendered, "\n")
	for i, line := range lines {
		vis := xansi.Strip(line)
		if strings.Trim(vis, " \t") == "" {
			prevItem = false
			continue
		}
		if listItemLineRe.MatchString(vis) {
			prevItem = true
			continue
		}
		if prevItem {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

// listItemLineRe matches a rendered list-item marker at the start of a line,
// after any leading whitespace: an unordered bullet (•, ‣, ▪, *, +, -) or an
// ordered number (1. 2) followed by a space.
var listItemLineRe = regexp.MustCompile(`^[ \t]*(?:[•‣▪*+\-]|\d+[.)])[ \t]`)

// CellWidth returns the display width of a single rune, using the
// grapheme-aware ANSI width primitive shared across the article rendering path so
// visible-cell measurement is consistent everywhere it is needed.
func CellWidth(r rune) int {
	return xansi.StringWidth(string(r))
}

// StyledCells returns the byte offset into s covering at most n visible cells,
// skipping ANSI escape sequences so the cut point lands on a character boundary
// and not inside a style escape. It is the single source of truth for "advance N
// visible cells" used by the blockquote bar helpers.
func StyledCells(s string, n int) int {
	i := 0
	seen := 0
	for i < len(s) && seen < n {
		if s[i] == '\x1b' {
			i = SkipEscape(s, i)
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		r, _ := utf8.DecodeRuneInString(s[i:])
		seen += CellWidth(r)
		i += size
	}
	return i
}

// FixBlockquoteRewrap repairs a glamour wrapping bug where, at many content
// widths, a short word inside a blockquote is split onto its own line that
// loses the blockquote "│" prefix (glamour re-wraps the blockquote wider than
// its inner paragraph, stranding the first word of a wrapped line). Such lines
// break the solid vertical quote bar. Rather than guess at the split, each
// blockquote run is reassembled into its logical text, word-wrapped once at the
// blockquote content width, and re-prefixed with the "│" bar, so no word is
// stranded and no line overflows.
func FixBlockquoteRewrap(rendered string) string {
	lines := strings.Split(rendered, "\n")
	for i := 0; i < len(lines); i++ {
		if !isBarLine(lines[i]) {
			continue
		}
		j := i + 1
		var bar string
		var content strings.Builder
		contentW := 0
		bar, _ = splitBar(lines[i])
		content.WriteString(barContent(lines[i]))
		if w := xansi.StringWidth(lines[i]); w > contentW {
			contentW = w
		}
		for j < len(lines) && (isBarLine(lines[j]) || isOrphanLine(lines, j)) {
			if isBarLine(lines[j]) {
				if w := xansi.StringWidth(lines[j]); w > contentW {
					contentW = w
				}
				content.WriteString(" ")
				content.WriteString(barContent(lines[j]))
			} else {
				content.WriteString(" ")
				content.WriteString(strings.TrimSpace(xansi.Strip(lines[j])))
			}
			j++
		}
		wrapped := xansi.Wrap(strings.TrimSpace(content.String()), contentW-2, " ,.;-+|")
		fixed := make([]string, 0, j-i)
		for _, l := range strings.Split(wrapped, "\n") {
			if strings.TrimSpace(xansi.Strip(l)) == "" {
				continue
			}
			fixed = append(fixed, bar+l)
		}
		lines = append(append(lines[:i], fixed...), lines[j:]...)
		i += len(fixed) - 1
	}
	return strings.Join(lines, "\n")
}

// isBarLine reports whether the line starts with a blockquote "│" bar.
func isBarLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(xansi.Strip(line)), "│")
}

// isOrphanLine reports whether the line is a short, bar-less word stranded
// inside a blockquote by glamour's mis-wrapping. An orphan is the head of a
// contiguous run of stranded words that ends at the next bar line, so consecutive
// orphans are folded too. The width threshold (smaller than the nearest
// preceding bar line) keeps ordinary paragraph text from being swallowed.
func isOrphanLine(lines []string, i int) bool {
	if i == 0 || i+1 >= len(lines) {
		return false
	}
	vis := strings.TrimSpace(xansi.Strip(lines[i]))
	if vis == "" || strings.HasPrefix(vis, "│") {
		return false
	}
	if !isBarLine(lines[i+1]) && !isOrphanLine(lines, i+1) {
		return false
	}
	for p := i - 1; p >= 0; p-- {
		if isBarLine(lines[p]) {
			return xansi.StringWidth(vis) < xansi.StringWidth(lines[p])
		}
	}
	return false
}

// splitBar splits a processed blockquote line into its styled "│ " prefix and
// the remaining styled content, walking visible cells (not bytes) so ANSI
// escape sequences in the prefix do not offset the content.
func splitBar(line string) (prefix, content string) {
	k := StyledCells(line, 2)
	return line[:k], line[k:]
}

// barContent returns the styled content of a blockquote line minus its "│ "
// prefix and trailing padding (which glamour writes as styled spaces).
func barContent(line string) string {
	_, c := splitBar(line)
	keep := xansi.StringWidth(strings.TrimRight(xansi.Strip(c), " "))
	return cutStyledWidth(c, keep)
}

// cutStyledWidth returns the styled substring of s covering at most w visible
// cells, walking visible cells (not bytes) so ANSI escape sequences do not
// shift the cut point.
func cutStyledWidth(s string, w int) string {
	return s[:StyledCells(s, w)]
}

// SkipEscape advances i past the escape/OSC/CSI sequence starting at i.
func SkipEscape(s string, i int) int {
	if s[i] != '\x1b' {
		return i
	}
	i++
	if i < len(s) && s[i] == ']' {
		i++
		for i < len(s) && s[i] != '\x07' && !(s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '\\') {
			i++
		}
		if i < len(s) {
			i++
		}
	} else if i < len(s) && s[i] == '[' {
		i++
		for i < len(s) && !('@' <= s[i] && s[i] <= '~') {
			i++
		}
		if i < len(s) {
			i++
		}
	} else if i < len(s) {
		i++
	}
	return i
}

// EscapeMarkdownText escapes the characters that CommonMark treats as special
// inline (emphasis, code spans, links, autolinks) in feed-controlled header
// text so it renders literally. Characters that are only special at the start
// of a line or inside a link destination are left alone.
func EscapeMarkdownText(s string) string {
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

// MarkdownToPlainText strips commonmark block and inline markers from markdown
// source, leaving readable plain text. It is used only as a graceful fallback
// when the glamour renderer fails, so it prioritizes readability over fidelity;
// visible link text and URLs are preserved as plain text.
func MarkdownToPlainText(md string) string {
	reLink := regexp.MustCompile(`!?\[([^\]]*)\]\(([^)]*)\)`)
	reCode := regexp.MustCompile("`[^`]*`")
	reBlock := regexp.MustCompile(`(?m)^#{1,6}\s+|^>\s+|^[ \t]*[-*+][ \t]+|^[ \t]*\d+[.)][ \t]+`)
	md = reLink.ReplaceAllString(md, "$1 $2") // [text](url) / ![alt](url) -> "text url"
	md = reCode.ReplaceAllString(md, "")      // `code` removed
	md = reBlock.ReplaceAllString(md, "")     // headings, blockquotes, list markers
	md = regexp.MustCompile(`[*_]+`).ReplaceAllString(md, "")
	md = strings.ReplaceAll(md, "\\", "")
	return strings.TrimSpace(md)
}
