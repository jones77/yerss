package ui

import (
	"regexp"
	"strings"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
	xansi "github.com/charmbracelet/x/ansi"
)

// glamourStandardStyle maps the configured display theme to a glamour standard
// style. "auto" resolves via the same light-background detection that drives
// the TUI palette, so the markdown and the rest of the UI agree.
func glamourStandardStyle(theme string) string {
	switch theme {
	case "light":
		return "light"
	case "dark":
		return "dark"
	default:
		if detectLightBackground() {
			return "light"
		}
		return "dark"
	}
}

// glamourStyleConfig returns the built-in dark/light style config with the
// document's left margin removed, so the reader's own padX padding is the only
// inset from the border (otherwise glamour's default two-column margin stacks
// on top of it). Links drop their underline.
func glamourStyleConfig(style string) ansi.StyleConfig {
	var cfg ansi.StyleConfig
	switch style {
	case "light":
		cfg = styles.LightStyleConfig
	default:
		cfg = styles.DarkStyleConfig
	}
	zero := uint(0)
	cfg.Document.Margin = &zero
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

// markdownRenderer returns the glamour renderer for the given content width,
// caching it across renders and rebuilding it only when the width or the
// resolved theme changes (glamour binds its word wrap at construction). It
// returns nil when a renderer cannot be built, in which case callers fall back
// to the raw markdown.
func (m *Model) markdownRenderer(contentW int) *glamour.TermRenderer {
	style := glamourStandardStyle(m.cfg.Display.Theme)
	if m.mdRenderer != nil && m.mdRendererW == contentW && m.mdRendererStyle == style {
		return m.mdRenderer
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(glamourStyleConfig(style)),
		glamour.WithWordWrap(contentW),
	)
	if err != nil {
		return nil
	}
	m.mdRenderer = r
	m.mdRendererW = contentW
	m.mdRendererStyle = style
	return r
}

// indentListContinuations shifts the wrapped continuation lines of list items
// two columns to the right so they align under the text that follows the
// bullet or number. Glamour wraps list content inside the same block as its
// marker, so every level's continuation lands on the marker column; adding two
// cells re-aligns it with the item's text. The article frame later truncates
// each line back to the content width, so no width bookkeeping is needed here.
func indentListContinuations(rendered string) string {
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

// escapeMarkdownText escapes the characters that CommonMark treats as special
// inline (emphasis, code spans, links, autolinks) in feed-controlled header
// text so it renders literally. Characters that are only special at the start
// of a line or inside a link destination are left alone.
func escapeMarkdownText(s string) string {
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

// markdownLink renders a link whose visible text is the URL itself, escaping
// the text and wrapping the destination in angle brackets when it contains
// characters (spaces, parentheses) that would break the plain `(url)` form.
func markdownLink(url string) string {
	text := escapeMarkdownText(url)
	dest := url
	if strings.ContainsAny(dest, " )") {
		dest = "<" + dest + ">"
	}
	return "[" + text + "](" + dest + ")"
}

// renderMarkdown renders markdown to styled terminal text at the content
// width, trimming glamour's surrounding blank lines. It returns the markdown
// unchanged when no renderer is available or rendering fails.
func (m *Model) renderMarkdown(md string, contentW int) string {
	r := m.markdownRenderer(contentW)
	if r == nil {
		return md
	}
	out, err := r.Render(md)
	if err != nil {
		return md
	}
	return indentListContinuations(strings.Trim(out, "\n"))
}