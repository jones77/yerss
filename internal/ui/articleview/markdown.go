package articleview

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
func (s *State) markdownRenderer(contentW int) *glamour.TermRenderer {
	style := compose.GlamourStandardStyle(s.sess.Config().Display.Theme)
	if s.mdRenderer != nil && s.mdRendererW == contentW && s.mdRendererStyle == style {
		return s.mdRenderer
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(compose.GlamourStyleConfig(style)),
		glamour.WithWordWrap(contentW),
	)
	if err != nil {
		return nil
	}
	s.mdRenderer = r
	s.mdRendererW = contentW
	s.mdRendererStyle = style
	return r
}

// renderMarkdown renders markdown to styled terminal text at the content
// width, trimming glamour's surrounding blank lines. It returns the markdown
// unchanged when no renderer is available or rendering fails.
func (s *State) renderMarkdown(md string, contentW int) string {
	r := s.markdownRenderer(contentW)
	if r == nil {
		return md
	}
	out, err := r.Render(md)
	if err != nil {
		return compose.MarkdownToPlainText(md)
	}
	return compose.FixBlockquoteRewrap(compose.IndentListContinuations(strings.Trim(out, "\n")))
}