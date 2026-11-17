package articleview

import (
	"charm.land/glamour/v2"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"yerss/internal"
	"yerss/internal/app"
	"yerss/internal/convert"
	"yerss/internal/ui/compose"
	"yerss/internal/ui/imageview"
	"yerss/internal/ui/render"
)

// styles holds the articleview-local style roles. The reader renders only the
// header URL line in the status-bar role; the list and popup styles live in
// their own packages.
type styles struct {
	status lipgloss.Style
}

func newStyles(p render.Palette) styles {
	return styles{status: lipgloss.NewStyle().Foreground(p.StatusBar)}
}

// State is the article reader: the composed content (image blocks spliced into
// the derived text), the viewport over it, the mouse selection, the frame
// cache, and the derivation cache that recomposes must not re-run. Model owns a
// State value and routes input and messages to it; State depends only on
// render, compose, imageview, app, and store.
type State struct {
	sess    *app.Session
	images  *imageview.ImageView
	palette render.Palette
	ascii   bool
	width   int
	height  int
	styles  styles

	contentGen       int
	nativeGen        int
	derivCache       Derivation
	derivKey         derivationKey
	derivValid       bool
	frameKey         FrameKey
	frameValid       bool
	frameStr         string
	frameRenderCount int
	recomposePending bool
	recomposeCount   int
	mdRenderer       *glamour.TermRenderer
	mdRendererW      int
	mdRendererStyle  string

	// LinkExpand is the link display mode this State was composed in. Model
	// owns the session-wide truth (Model.linkExpand) and syncs it here.
	LinkExpand bool

	id          int64
	article     *internal.Article
	lines       []string
	Viewport    viewport.Model
	readMarked  bool
	Sel         Selection
	headerLines int
	// fitHeaderLines is the mode-independent header budget the image-block
	// height fit reserves: the line count of the tallest header the article
	// presents across both link modes. The height fit consumes this instead of
	// headerLines so a photo does not re-size when the user toggles link mode
	// (the expanded header's URL line would otherwise shrink the photo).
	fitHeaderLines int
	imgStart       int
	imgEnd         int
	capStart       int
	nativeImg      bool
	imageBlocks    []compose.ImageBlock
	imageURLs      []string
	// anchorRows records, parallel to imageURLs, the content row where each
	// image's block starts in the current composition (the block's ImgStart
	// when composed, the reserved placeholder row while uncomposed). It is
	// recomputed on every recompose like imageBlocks, so the load frontier can
	// compare it against the viewport without walking the body markdown.
	anchorRows   []int
	inlineImages []convert.InlineImage
	inlineCaps   map[string]string
	links        []Link
	// leadShown reports whether blocks[0] was composed at the top of the
	// content, directly below the header, and shown on open: a genuine lead
	// block, or a suppressed lead whose URL also opens the body as its first
	// inline image (a top-of-article photo reused in the body). When the first
	// content element is text, blocks[0] is a mid-article inline image that
	// was NOT shown on open and the scroll snap must treat it like any other
	// inline image instead of assuming its entry stages were pre-consumed.
	leadShown bool
}

// NewState builds an empty reader State with the package-local styles and the
// reader's session dependencies. Model assigns the result into m.article and
// fills the article fields through Recompose.
func NewState(sess *app.Session, palette render.Palette, ascii bool, images *imageview.ImageView, width, height int) State {
	return State{
		sess:    sess,
		images:  images,
		palette: palette,
		ascii:   ascii,
		styles:  newStyles(palette),
		width:   width,
		height:  height,
	}
}

// SetSize updates the terminal geometry the composition derives content width
// and viewport height from.
func (s *State) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// SetAscii updates the glyph mode; the reader re-derives on the next recompose
// because the derivation key covers ascii.
func (s *State) SetAscii(ascii bool) {
	s.ascii = ascii
}

// glyphs returns the border glyph set, switching the scrollbar fill to the
// double-line style when the display requests it.
func (s *State) glyphs() render.BorderGlyphs {
	g := render.GlyphsFor(s.ascii)
	if s.sess.Config().Display.Scrollbar == "double" && !s.ascii {
		g.Fill = "║"
	}
	return g
}

// base copies the session infrastructure of s onto a fresh State so
// composeArticle can fill in the article fields without repeating it.
func (s *State) base() State {
	return State{
		sess: s.sess, images: s.images, palette: s.palette, ascii: s.ascii,
		styles: s.styles, width: s.width, height: s.height,
		contentGen: s.contentGen, nativeGen: s.nativeGen,
		derivCache: s.derivCache, derivKey: s.derivKey, derivValid: s.derivValid,
		frameKey: s.frameKey, frameValid: s.frameValid, frameStr: s.frameStr,
		frameRenderCount: s.frameRenderCount,
		recomposePending: s.recomposePending, recomposeCount: s.recomposeCount,
		mdRenderer: s.mdRenderer, mdRendererW: s.mdRendererW, mdRendererStyle: s.mdRendererStyle,
		LinkExpand: s.LinkExpand,
	}
}

// ID returns the article id.
func (s *State) ID() int64 { return s.id }

// Article returns the article being read, or nil before the first Recompose.
func (s *State) Article() *internal.Article { return s.article }

// Lines returns the composed content lines (the viewport source).
func (s *State) Lines() []string { return s.lines }

// ViewportY returns the current scroll offset.
func (s *State) ViewportY() int { return s.Viewport.YOffset }

// ViewportHeight returns the viewport height in rows.
func (s *State) ViewportHeight() int { return s.Viewport.Height }

// ViewportWidth returns the viewport width in columns.
func (s *State) ViewportWidth() int { return s.Viewport.Width }

// SetViewportY sets the scroll offset.
func (s *State) SetViewportY(y int) { s.Viewport.SetYOffset(y) }

// ImageBlocks returns the composed image blocks in content order.
func (s *State) ImageBlocks() []compose.ImageBlock { return s.imageBlocks }

// NativeImg reports whether blocks[0] is a native terminal image.
func (s *State) NativeImg() bool { return s.nativeImg }

// LeadShown reports whether blocks[0] was shown on open (see the field doc).
func (s *State) LeadShown() bool { return s.leadShown }

// FitHeaderLines returns the mode-independent header budget.
func (s *State) FitHeaderLines() int { return s.fitHeaderLines }

// AnchorRows returns the per-image anchor rows, parallel to ImageURLs.
func (s *State) AnchorRows() []int { return s.anchorRows }

// ImageURLs returns the article's image URLs (lead first, then inline).
func (s *State) ImageURLs() []string { return s.imageURLs }

// InlineCaps returns the per-inline-image caption map.
func (s *State) InlineCaps() map[string]string { return s.inlineCaps }

// InlineImages returns the inline image list from the content conversion.
func (s *State) InlineImages() []convert.InlineImage { return s.inlineImages }

// HeaderLines returns the rendered header line count.
func (s *State) HeaderLines() int { return s.headerLines }

// ImgStart returns the first block's start row, or -1 with no blocks.
func (s *State) ImgStart() int { return s.imgStart }

// ImgEnd returns the first block's end row, or -1 with no blocks.
func (s *State) ImgEnd() int { return s.imgEnd }

// CapStart returns the first block's caption start row, or -1 with no blocks.
func (s *State) CapStart() int { return s.capStart }

// ReadMarked reports whether the article was already marked read this session.
func (s *State) ReadMarked() bool { return s.readMarked }

// Links returns the article's harvested links.
func (s *State) Links() []Link { return s.links }

// SetRecomposePending queues a recompose to run on the next flush.
func (s *State) SetRecomposePending(b bool) { s.recomposePending = b }

// ContentGen returns the content-generation counter (frame-cache input).
func (s *State) ContentGen() int { return s.contentGen }

// NativeGen returns the native re-show generation counter.
func (s *State) NativeGen() int { return s.nativeGen }

// FrameRenderCount returns how many uncached frames were rendered.
func (s *State) FrameRenderCount() int { return s.frameRenderCount }

// RecomposeCount returns how many image-load recomposes have run.
func (s *State) RecomposeCount() int { return s.recomposeCount }
