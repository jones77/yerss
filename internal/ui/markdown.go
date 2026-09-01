package ui

import (
	"strings"

	"charm.land/glamour/v2"

	"yerss/internal/ui/compose"
)

// markdownRenderer returns the glamour renderer for the given content width,
// caching it across renders and rebuilding it only when the width or the
// resolved theme changes (glamour binds its word wrap at construction). It
// returns nil when a renderer cannot be built, in which case callers fall back
// to the raw markdown.
func (m *Model) markdownRenderer(contentW int) *glamour.TermRenderer {
	style := compose.GlamourStandardStyle(m.cfg.Display.Theme)
	if m.mdRenderer != nil && m.mdRendererW == contentW && m.mdRendererStyle == style {
		return m.mdRenderer
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(compose.GlamourStyleConfig(style)),
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
		return compose.MarkdownToPlainText(md)
	}
	return compose.FixBlockquoteRewrap(compose.IndentListContinuations(strings.Trim(out, "\n")))
}
