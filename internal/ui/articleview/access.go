package articleview

import (
	"charm.land/glamour/v2"

	"yerss/internal"
	"yerss/internal/ui/render"
)

// DerivationKey is an exported, comparable mirror of the derivation cache key.
// The unexported derivationKey remains the stored field; the accessor converts
// so package ui tests can compare keys and render them.
type DerivationKey struct {
	ID         int64
	ContentW   int
	VpH        int
	Ascii      bool
	Palette    render.Palette
	Style      string
	LinkExpand bool
	Content    string
}

// Segments returns the derivation's rendered body segments.
func (d Derivation) Segments() []string { return d.segments }

// SetRecomposeCount sets the recompose counter (test support).
func (s *State) SetRecomposeCount(n int) { s.recomposeCount = n }

// SetFrameRenderCount sets the frame render counter (test support).
func (s *State) SetFrameRenderCount(n int) { s.frameRenderCount = n }

// SetFrameValid forces a frame-cache miss (test support).
func (s *State) SetFrameValid(b bool) { s.frameValid = b }

// SetDerivValid forces a derivation-cache miss (test support).
func (s *State) SetDerivValid(b bool) { s.derivValid = b }

// DerivValid reports whether the derivation cache is populated.
func (s *State) DerivValid() bool { return s.derivValid }

// DerivCache returns the cached derivation.
func (s *State) DerivCache() Derivation { return s.derivCache }

// DerivKey returns an exported view of the derivation cache key.
func (s *State) DerivKey() DerivationKey {
	dk := s.derivKey
	return DerivationKey{
		ID:         dk.id,
		ContentW:   dk.contentW,
		VpH:        dk.vpH,
		Ascii:      dk.ascii,
		Palette:    dk.palette,
		Style:      dk.style,
		LinkExpand: dk.linkExpand,
		Content:    dk.content,
	}
}

// Derive returns the derivation for the article, cached by key.
func (s *State) Derive(a internal.Article, contentW, vpH int) Derivation {
	return s.deriveArticle(a, contentW, vpH)
}

// RenderMarkdown renders markdown to ANSI text at the given width.
func (s *State) RenderMarkdown(md string, contentW int) string {
	return s.renderMarkdown(md, contentW)
}

// MarkdownRenderer returns the cached glamour renderer for the width.
func (s *State) MarkdownRenderer(contentW int) *glamour.TermRenderer {
	return s.markdownRenderer(contentW)
}

// MarkdownRendererW returns the cached renderer width (test support).
func (s *State) MarkdownRendererW() int { return s.mdRendererW }

// MarkdownRendererStyle returns the cached renderer style (test support).
func (s *State) MarkdownRendererStyle() string { return s.mdRendererStyle }

// RenderHeaderMode renders the article header in compact or expanded link
// mode at the given content width.
func (s *State) RenderHeaderMode(a internal.Article, contentW int, linkExpand bool) string {
	return s.renderHeaderMode(a, contentW, linkExpand)
}

// SetNativeImg overrides the lead-image native-render flag (test support).
func (s *State) SetNativeImg(b bool) { s.nativeImg = b }

// SetLines replaces the composed article lines (test support).
func (s *State) SetLines(lines []string) { s.lines = lines }

// HighlightSelection applies the inverted selection style to the selected cell
// range of each visible content line.
func HighlightSelection(lines []string, sel Selection) []string {
	return highlightSelection(lines, sel)
}
