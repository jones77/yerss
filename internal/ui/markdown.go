package ui

import (
	"strings"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/ansi"
	"charm.land/glamour/v2/styles"
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
// on top of it).
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
	return strings.Trim(out, "\n")
}