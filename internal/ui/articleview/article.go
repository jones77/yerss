package articleview

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/convert"
	"yerss/internal/image"
	"yerss/internal/store"
	"yerss/internal/timeutil"
	"yerss/internal/ui/compose"
	"yerss/internal/ui/render"
)

// RendersImage reports whether the given URL is one of the article's images
// (the lead or an inline image) that the article view renders.
func (s *State) RendersImage(url string) bool {
	for _, u := range s.imageURLs {
		if u == url {
			return true
		}
	}
	return false
}

// URLIndex returns the position of url in the article's image list (document
// order: the lead at 0, then inline images), or -1 when the article does not
// render it.
func (s *State) URLIndex(url string) int {
	for i, u := range s.imageURLs {
		if u == url {
			return i
		}
	}
	return -1
}

// Recompose rebuilds the article view state from the article source, bumping
// the content-generation counter (which invalidates the rendered-frame cache)
// and, for a new article, resetting the image load frontier. It returns a new
// State value; callers assign it back. Recomposes of the same article keep the
// load frontier so it stays monotonic within the article open.
func (s *State) Recompose(a store.Article) State {
	s.contentGen++
	if s.article == nil || a.ID != s.id {
		s.images.ResetForNewArticle()
	}
	padX, padY := s.sess.Config().Display.PaddingX, s.sess.Config().Display.PaddingY
	contentW, vpH, _ := render.ContentGeom(s.width, s.height, padX, padY)
	return s.composeArticle(a, s.deriveArticle(a, contentW, vpH), contentW, vpH)
}

// Derivation holds the expensive, content-only display derivation of an
// article — the converted markdown body, rendered text segments, inline-image
// list, harvested links, and rendered header — that image-load recomposes must
// not re-run. Its rendered pieces depend only on the article content and the
// content width; the viewport height feeds the key because the composition that
// consumes it fits image blocks against the viewport.
type Derivation struct {
	content     string
	header      string
	headerLines int
	// fitHeaderLines is the header budget the image-block height fit reserves
	// (see State.fitHeaderLines). It is the tallest header across both
	// link modes so the fit does not depend on the current display mode.
	fitHeaderLines int
	bodyMD         string
	segments       []string
	inline         []convert.InlineImage
	imageURLs      []string
	links          []Link
	captions       map[string]string
	leadURL        string
}

// derivationKey identifies a cached Derivation: the article (id and source
// content), the content geometry, and the renderer inputs that change the
// derived output (ascii mode, palette, the glamour theme style, and the link
// display mode).
type derivationKey struct {
	id         int64
	contentW   int
	vpH        int
	ascii      bool
	palette    render.Palette
	style      string
	linkExpand bool
	content    string
}

// deriveArticle returns the article's derived display content, reusing the
// cached single entry when the article and all derivation inputs match. A
// stale entry (different article, geometry, or renderer inputs) is replaced.
func (s *State) deriveArticle(a store.Article, contentW, vpH int) Derivation {
	key := derivationKey{
		id:         a.ID,
		contentW:   contentW,
		vpH:        vpH,
		ascii:      s.ascii,
		palette:    s.palette,
		style:      compose.GlamourStandardStyle(s.sess.Config().Display.Theme),
		linkExpand: s.LinkExpand,
		content:    a.Content,
	}
	if s.derivValid && s.derivKey == key {
		return s.derivCache
	}
	der := s.computeDerivation(a, contentW)
	s.derivCache = der
	s.derivKey = key
	s.derivValid = true
	return der
}

// stripSoftHyphens removes U+00AD soft hyphens from s. The source HTML uses
// them as invisible line-break hints; the terminal renders each as a visible
// cell where the app counts it as zero-width, so a line padded to the content
// width overflows the frame by one column per soft hyphen and wraps the right
// border (scrollbar) onto the next line's left.
func stripSoftHyphens(s string) string {
	return strings.ReplaceAll(s, "\u00ad", "")
}

// computeDerivation converts the article's HTML content to marked markdown,
// discovers and dedupes the inline images, harvests the links, renders the
// header and each body text segment through the markdown renderer, and records
// the per-image captions — the work an image-load recompose must not repeat.
func (s *State) computeDerivation(a store.Article, contentW int) Derivation {
	a.Title = stripSoftHyphens(a.Title)
	a.Author = stripSoftHyphens(a.Author)
	a.Link = stripSoftHyphens(a.Link)
	a.Content = stripSoftHyphens(a.Content)
	header := s.renderHeader(a, contentW)
	headerLines := 0
	if header != "" {
		headerLines = len(strings.Split(header, "\n"))
	}
	// The image-block height fit reserves the tallest header the article
	// presents across both link modes, so a photo does not re-size when the
	// link display mode toggles. Rendered in a fixed mode (expanded) here
	// because the expanded header is always at least as tall as the compact
	// one: compact omits the URL line that expanded prepends.
	fitHeaderLines := headerLines
	if expanded := s.renderHeaderExpanded(a, contentW); expanded != "" {
		if n := len(strings.Split(expanded, "\n")); n > fitHeaderLines {
			fitHeaderLines = n
		}
	}
	bodyMD, inline := convert.ConvertImages(a.Content)
	leadURL := ""
	if a.ImageURL != "" {
		leadSrc := compose.CanonicalSource(a.ImageURL)
		suppressed := false
		for _, im := range inline {
			if compose.CanonicalSource(im.URL) == leadSrc {
				suppressed = true
				break
			}
		}
		if !suppressed {
			leadURL = a.ImageURL
		}
	}
	inline = compose.DedupeInline(leadURL, inline)
	bodyMD = compose.StripDuplicateSentinels(bodyMD, leadURL)

	// Render the body's text segments in document order, matching the segment
	// walk spliceBody re-runs so the pre-rendered chunks slot back in exactly.
	var segments []string
	var captions map[string]string
	if len(inline) == 0 {
		segments = append(segments, s.renderMarkdown(bodyMD, contentW))
	} else {
		captions = map[string]string{}
		pos := 0
		idx := 0
		for {
			start, end, url, linkText, ok := compose.NextSentinel(bodyMD, pos)
			if !ok {
				break
			}
			if seg := strings.TrimSpace(bodyMD[pos:start]); seg != "" {
				segments = append(segments, s.renderMarkdown(bodyMD[pos:start], contentW))
			}
			alt := ""
			if idx < len(inline) && inline[idx].URL == url {
				alt = inline[idx].Alt
				idx++
			}
			captions[url] = compose.ResolveImageAttribution(a, url, alt, linkText, false)
			pos = end
		}
		if rest := strings.TrimSpace(bodyMD[pos:]); rest != "" {
			segments = append(segments, s.renderMarkdown(bodyMD[pos:], contentW))
		}
	}
	imageURLs := []string{}
	if leadURL != "" {
		imageURLs = append(imageURLs, leadURL)
	}
	for _, im := range inline {
		imageURLs = append(imageURLs, im.URL)
	}
	// Compact link mode (the default) drops each body link's visible href span
	// so only the clickable blue name renders. This runs on the rendered body
	// segments only — renderMarkdown stays untouched for the header title,
	// author, and popup text.
	if !s.LinkExpand {
		for i, seg := range segments {
			segments[i] = compose.StripLinkHrefs(seg)
		}
		// Stripping a wrapped URL's href spans swallows the line's fill padding
		// and a wrap newline, leaving a short, unpadded glued line. Re-flow
		// affected paragraphs so every line fills the content width again.
		for i, seg := range segments {
			segments[i] = compose.ReflowShortLines(seg, contentW)
		}
	} else {
		// Expanded mode keeps the URLs, but glamour pushes a URL that does not
		// fit next to its name onto its own line, leaving the name line half
		// empty. Re-flow such paragraphs so the URL starts on the name line
		// and wraps at the content width, keeping each wrapped line clickable.
		for i, seg := range segments {
			segments[i] = compose.InlineLinkURLs(seg, contentW)
		}
	}
	return Derivation{
		content:        a.Content,
		header:         header,
		headerLines:    headerLines,
		fitHeaderLines: fitHeaderLines,
		bodyMD:         bodyMD,
		segments:       segments,
		inline:         inline,
		imageURLs:      imageURLs,
		links:          HarvestArticleLinks(a),
		captions:       captions,
		leadURL:        leadURL,
	}
}

// composeArticle builds the article view state from the cached derivation:
// it renders the image blocks (served from the image caches), splices them
// into the derived text segments at the sentinel positions, and builds the
// viewport. This is the per-recompose part — image blocks change as loads land
// while the derived body stays put.
func (s *State) composeArticle(a store.Article, der Derivation, contentW, vpH int) State {
	var parts []string
	var blocks []compose.ImageBlock
	var captions map[string]string
	var anchors []int
	lineCount := 0
	addPart := func(part string) {
		parts = append(parts, part)
		lineCount += len(strings.Split(part, "\n"))
	}
	addImageBlock := func(url string, block []string, rows int, native bool, id uint32) int {
		if len(parts) > 0 && parts[len(parts)-1] != "" {
			parts = append(parts, "")
			lineCount++
		} else if len(parts) == 0 {
			parts = append(parts, "")
			lineCount++
		}
		start := lineCount
		addPart(strings.Join(block, "\n"))
		blocks = append(blocks, compose.ImageBlock{URL: url, ImgStart: start, CapStart: start + rows, ImgEnd: start + len(block) - 1, NativeImg: native, NativeID: id})
		parts = append(parts, "")
		lineCount++
		return start
	}

	if der.header != "" {
		addPart(der.header)
		parts = append(parts, "")
		lineCount++
	}
	if der.leadURL != "" {
		leadAnchor := der.headerLines + 1
		if s.images.Enabled() {
			if block, rows, native, id := s.images.ComposeBlock(a.ImageURL, compose.ResolveImageAttribution(a, a.ImageURL, "", "", true), contentW, vpH, der.fitHeaderLines); len(block) > 0 {
				leadAnchor = addImageBlock(a.ImageURL, block, rows, native, id)
			} else {
				// The lead is uncomposed: its block will land on the blank
				// line below the header (or the top of the content when the
				// header is empty), so that row is its anchor.
				leadAnchor = der.headerLines + 1
			}
		}
		anchors = append(anchors, leadAnchor)
	}

	if len(der.inline) == 0 {
		addPart(der.segments[0])
	} else {
		bodyParts, bodyBlocks, bodyAnchors, inlineCaps := s.spliceBody(der, contentW, vpH, der.fitHeaderLines, lineCount)
		parts = append(parts, bodyParts...)
		lineCount += compose.LineCountOf(bodyParts)
		blocks = append(blocks, bodyBlocks...)
		anchors = append(anchors, bodyAnchors...)
		captions = inlineCaps
	}

	rendered := strings.Join(parts, "\n")
	vp := viewport.New(contentW, vpH)
	vp.SetContent(rendered)
	// leadShown reports whether blocks[0] is composed at the top of the
	// content, directly below the header (one row past its blank line), and
	// shown on open: a genuine lead block, or a suppressed lead whose URL also
	// opens the body as its first inline image — a top-of-article photo reused
	// in the body, as ProPublica's features do. The scroll snap treats such a
	// block's entry stages as pre-consumed (rise flush to the viewport top,
	// then skip past its caption) instead of snapping it like a mid-article
	// inline image that entered from below. When the first content element is
	// text, blocks[0] is a mid-article inline image that was never shown on
	// open and the snap must treat it as a regular inline image.
	leadShown := len(blocks) > 0 && blocks[0].ImgStart == der.headerLines+1
	imgStart, imgEnd, capStart, nativeImg := -1, -1, -1, false
	if len(blocks) > 0 {
		imgStart, imgEnd, capStart = blocks[0].ImgStart, blocks[0].ImgEnd, blocks[0].CapStart
		nativeImg = blocks[0].NativeImg
	}
	st := s.base()
	st.id = a.ID
	st.article = &a
	st.lines = strings.Split(rendered, "\n")
	st.Viewport = vp
	st.headerLines = der.headerLines
	st.fitHeaderLines = der.fitHeaderLines
	st.imgStart = imgStart
	st.imgEnd = imgEnd
	st.capStart = capStart
	st.nativeImg = nativeImg
	st.imageBlocks = blocks
	st.imageURLs = der.imageURLs
	st.anchorRows = anchors
	st.inlineImages = der.inline
	st.inlineCaps = captions
	st.links = der.links
	st.leadShown = leadShown
	if len(st.lines) <= vpH {
		st.MarkRead(s.sess.Store())
	}
	return st
}

// spliceBody renders the cached body text segments with inline image blocks
// interleaved at the sentinel positions the derivation walked, so the joined
// output is byte-identical to the pre-cache composition. An inline image whose
// source is not yet composed renders a blank placeholder line in its place so
// the text flow stays stable until its load lands and re-composes the block. A
// sentinel wrapped in a markdown link (`<a><img></a>`) is already consumed by
// the sentinel walk, so no stray `[`/`](...)` renders around the image. The
// returned int slice holds each inline image's anchor row (the block's start
// row when composed, the reserved placeholder row while uncomposed), parallel
// to the inline images in document order.
func (s *State) spliceBody(der Derivation, contentW, vpH, headerLines, base int) ([]string, []compose.ImageBlock, []int, map[string]string) {
	var parts []string
	var blocks []compose.ImageBlock
	var anchors []int
	lineCount := base
	segIdx := 0
	addPart := func(part string) {
		parts = append(parts, part)
		lineCount += len(strings.Split(part, "\n"))
	}
	addImageBlock := func(url string, block []string, rows int, native bool, id uint32) int {
		if len(parts) > 0 && parts[len(parts)-1] != "" {
			parts = append(parts, "")
			lineCount++
		}
		start := lineCount
		addPart(strings.Join(block, "\n"))
		blocks = append(blocks, compose.ImageBlock{URL: url, ImgStart: start, CapStart: start + rows, ImgEnd: start + len(block) - 1, NativeImg: native, NativeID: id})
		parts = append(parts, "")
		lineCount++
		return start
	}
	pos := 0
	for {
		start, end, url, _, ok := compose.NextSentinel(der.bodyMD, pos)
		if !ok {
			break
		}
		if seg := strings.TrimSpace(der.bodyMD[pos:start]); seg != "" {
			addPart(der.segments[segIdx])
			segIdx++
		}
		if block, rows, native, id := s.images.ComposeBlock(url, der.captions[url], contentW, vpH, headerLines); len(block) > 0 {
			anchors = append(anchors, addImageBlock(url, block, rows, native, id))
		} else {
			// Placeholder: keep the paragraph break so the body text stays
			// stable and the block inserts at the right offset when the load
			// lands and the article re-composes. The placeholder blank is the
			// row the block will start at.
			if len(parts) == 0 || parts[len(parts)-1] != "" {
				parts = append(parts, "")
				lineCount++
			}
			anchors = append(anchors, lineCount)
		}
		pos = end
	}
	if rest := strings.TrimSpace(der.bodyMD[pos:]); rest != "" {
		addPart(der.segments[segIdx])
	}
	return parts, blocks, anchors, der.captions
}

// stripURLScheme removes a leading "https://" or "http://" from a displayed
// URL so the header line shows the domain and path without the scheme prefix.
// The scheme adds no information for a reader-facing link line and wastes
// columns on a narrow terminal.
func stripURLScheme(url string) string {
	if strings.HasPrefix(url, "https://") {
		return url[len("https://"):]
	}
	if strings.HasPrefix(url, "http://") {
		return url[len("http://"):]
	}
	return url
}

// renderHeader renders the reader header: in expanded link mode the article URL
// as the very first line, a blank line, the bold title, and the author line
// directly beneath it with no blank between them; in compact link mode the URL
// line is omitted so the title is the first content line. The URL is shown with
// its scheme stripped but is wrapped in an OSC 8 hyperlink whose target keeps
// the full URL, so it stays clickable. A long URL wraps at the content width
// onto subsequent lines so the whole display string is visible, each wrapped
// line carrying its own OSC 8 hyperlink (some terminals drop hyperlinks across
// a line break). Title and author are rendered as separate documents and joined
// because the markdown renderer folds line breaks inside a paragraph into
// spaces.
func (s *State) renderHeader(a store.Article, contentW int) string {
	return s.renderHeaderMode(a, contentW, s.LinkExpand)
}

// renderHeaderExpanded renders the header in the fixed expanded mode, used to
// compute the mode-independent fitHeaderLines budget.
func (s *State) renderHeaderExpanded(a store.Article, contentW int) string {
	return s.renderHeaderMode(a, contentW, true)
}

func (s *State) renderHeaderMode(a store.Article, contentW int, linkExpand bool) string {
	var meta []string
	if a.Title != "" {
		meta = append(meta, s.renderMarkdown("**"+compose.EscapeMarkdownText(a.Title)+"**", contentW))
	}
	if a.Author != "" {
		meta = append(meta, s.renderMarkdown("by "+compose.EscapeMarkdownText(a.Author), contentW))
	}
	metaBlock := strings.Join(meta, "\n")
	if a.Link == "" || !linkExpand {
		return metaBlock
	}
	urlPart := s.renderHeaderLink(a.Link, contentW)
	if metaBlock == "" {
		return urlPart
	}
	return urlPart + "\n\n" + metaBlock
}

// renderHeaderLink wraps the stripped article URL at the content width and
// emits each line as an OSC 8 hyperlink carrying the full URL, rendered in the
// status-bar blue role.
func (s *State) renderHeaderLink(url string, contentW int) string {
	var parts []string
	for _, l := range image.LayoutWrapLines(stripURLScheme(url), contentW) {
		parts = append(parts, s.styles.status.Render(ansi.SetHyperlink(url)+l+ansi.ResetHyperlink()))
	}
	return strings.Join(parts, "\n")
}

// FrameKey is a comparable signature of the render inputs the article frame
// body draws from. It is a value-keyed signature rather than a manual dirty
// flag, so the cache stays correct for the SetYOffset calls that bypass
// scrollArticle (resize, recompose snap, selection rebuild): the signature
// changes exactly when the rendered output would. The read set it covers: the
// article content and viewport composition (contentGen), the native
// re-show/suppress transforms' dependence on nativeSentPrev (nativeGen), the
// viewport offset and terminal geometry (yOffset, width, height), the mouse
// selection highlight (sel), and the glyph/link display modes (ascii,
// linkExpand). palette and ShowSeconds are fixed for the session and the
// padding is fixed config, so they are omitted.
type FrameKey struct {
	ArticleID  int64
	ContentGen int
	NativeGen  int
	YOffset    int
	Width      int
	Height     int
	Sel        Selection
	Ascii      bool
	LinkExpand bool
}

// KeyFor builds the frame-cache key from the current reader state.
func (s *State) KeyFor() FrameKey {
	return FrameKey{
		ArticleID:  s.id,
		ContentGen: s.contentGen,
		NativeGen:  s.nativeGen,
		YOffset:    s.Viewport.YOffset,
		Width:      s.width,
		Height:     s.height,
		Sel:        s.Sel,
		Ascii:      s.ascii,
		LinkExpand: s.LinkExpand,
	}
}

// Render returns the article reader's frame, memoized: when the state
// signature matches the frame already rendered, the cached frame string is
// returned without re-running the viewport composition, native re-show and
// clipped-transmit transforms, selection highlight, or border render. View()
// keeps prepending nativeClearPrefix to the cached body, so the prefix's
// per-frame transitions do not defeat the cache.
func (s *State) Render() string {
	if s.frameValid && s.frameKey == s.KeyFor() {
		return s.frameStr
	}
	frame := s.RenderUncached()
	s.frameStr = frame
	s.frameKey = s.KeyFor()
	s.frameValid = true
	return frame
}

// RenderUncached runs the full article frame render pipeline with no
// memoization: the viewport view, the native re-show and clipped-transmit
// transforms, the selection highlight, and the article border.
func (s *State) RenderUncached() string {
	s.frameRenderCount++
	a := s.article
	date := ""
	if !a.PublishedAt.IsZero() {
		date = a.PublishedAt.Local().Format(timeutil.DateTimeLayout(s.sess.Config().Display.ShowSeconds))
	}
	title := a.Title
	if title == "" {
		title = "(untitled)"
	}
	g := s.glyphs()
	content := s.Viewport.View()
	lines := strings.Split(content, "\n")
	s.images.RewriteNativeReShows(lines, s)
	s.images.SuppressClippedNativeTransmits(lines, s)
	lines = highlightSelection(lines, s.Sel)
	sc := render.ScrollState{
		TotalH:    s.Viewport.TotalLineCount(),
		ViewportH: s.Viewport.Height,
		Offset:    s.Viewport.YOffset,
	}
	return render.RenderArticleBorder(s.width, s.height, s.sess.Config().Display.PaddingX, s.sess.Config().Display.PaddingY,
		g, s.palette, date, title, sc, lines)
}

// highlightSelection applies the inverted selection style to the selected cell
// range of each visible content line. Lines outside the range are returned
// unchanged; ANSI and OSC 8 escape sequences on a line are preserved by
// ansi.Cut/ansi.Truncate. The reverse-video SGR is emitted directly so the
// highlight works regardless of the terminal profile (lipgloss's Reverse is a
// no-op when no TTY is detected).