package ui

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/convert"
	"yerss/internal/image"
	"yerss/internal/store"
	"yerss/internal/ui/compose"
	"yerss/internal/ui/render"
)

type articleState struct {
	id           int64
	article      *store.Article
	lines        []string
	viewport     viewport.Model
	readMarked   bool
	sel          textSelection
	headerLines  int
	imgStart     int
	imgEnd       int
	capStart     int
	nativeImg    bool
	imageBlocks  []compose.ImageBlock
	imageURLs    []string
	// anchorRows records, parallel to imageURLs, the content row where each
	// image's block starts in the current composition (the block's ImgStart
	// when composed, the reserved placeholder row while uncomposed). It is
	// recomputed on every recompose like imageBlocks, so the load frontier can
	// compare it against the viewport without walking the body markdown.
	anchorRows   []int
	inlineImages []convert.InlineImage
	inlineCaps   map[string]string
	links        []articleLink
	// leadShown reports whether blocks[0] was composed at the top of the
	// content, directly below the header, and shown on open: a genuine lead
	// block, or a suppressed lead whose URL also opens the body as its first
	// inline image (a top-of-article photo reused in the body). When the first
	// content element is text, blocks[0] is a mid-article inline image that
	// was NOT shown on open and the scroll snap must treat it like any other
	// inline image instead of assuming its entry stages were pre-consumed.
	leadShown bool
}

// rendersImage reports whether the given URL is one of the article's images
// (the lead or an inline image) that the article view renders.
func (s *articleState) rendersImage(url string) bool {
	for _, u := range s.imageURLs {
		if u == url {
			return true
		}
	}
	return false
}

// urlIndex returns the position of url in the article's image list (document
// order: the lead at 0, then inline images), or -1 when the article does not
// render it.
func (s *articleState) urlIndex(url string) int {
	for i, u := range s.imageURLs {
		if u == url {
			return i
		}
	}
	return -1
}

// articleLink is one hyperlink harvested from the article's markdown source.
type articleLink struct {
	text string
	url  string
}

// harvestLinks extracts the hyperlinks from the article's markdown source in
// document order, deduplicating by URL. It scans the `[text](url)` forms the
// converter emits, tolerating `<...>`-wrapped destinations.
func harvestLinks(md string) []articleLink {
	re := regexp.MustCompile(`!?\[([^\]]*)\]\(([^)]+)\)`)
	seen := map[string]bool{}
	var out []articleLink
	for _, m := range re.FindAllStringSubmatch(md, -1) {
		dest := strings.TrimPrefix(strings.TrimSuffix(m[2], ">"), "<")
		if dest == "" || seen[dest] {
			continue
		}
		seen[dest] = true
		out = append(out, articleLink{text: m[1], url: dest})
	}
	return out
}

// harvestArticleLinks collects every hyperlink rendered in the article body —
// the markdown links, not the article's own header URL — deduplicated by URL,
// with relative destinations resolved to absolute against the article's origin
// so the links popup and open-url work. It scans the same marked markdown the
// body renders (ConvertImages), so a link that wraps an inline image (promo
// banners render as a linked image) is harvested like any other, with the
// image sentinel stripped from the link text.
func harvestArticleLinks(a store.Article) []articleLink {
	seen := map[string]bool{}
	var out []articleLink
	add := func(text, url string) {
		if url == "" || seen[url] {
			return
		}
		seen[url] = true
		out = append(out, articleLink{text: text, url: url})
	}
	md, _ := convert.ConvertImages(a.Content)
	for _, l := range harvestLinks(md) {
		add(compose.CleanLinkText(compose.SentinelRe.ReplaceAllString(l.text, "")), resolveLink(a, l.url))
	}
	return out
}

// resolveLink resolves a link destination harvested from an article's HTML
// against the article's origin: a relative target like "/tomdispatch" becomes
// absolute (https://theintercept.com/tomdispatch) so opening it navigates to
// the right page. Absolute URLs pass through; targets with no resolvable base
// are returned unchanged.
func resolveLink(a store.Article, href string) string {
	u, err := url.Parse(href)
	if err != nil || u.IsAbs() {
		return href
	}
	for _, base := range []string{a.Link, a.FeedURL} {
		bu, err := url.Parse(base)
		if err != nil || bu.Scheme == "" || bu.Host == "" {
			continue
		}
		return bu.ResolveReference(u).String()
	}
	return href
}

// textSelection tracks a mouse text selection in the article viewport. Cells
// are in content-area coordinates (relative to the visible viewport window,
// i.e. 0..viewportH rows and 0..contentW columns). active means there is a
// range to highlight; tracking means the mouse button is currently held and
// dragging extends the range.
func (m *Model) openArticle() tea.Cmd {
	item, ok := m.articleAtCursor()
	if !ok {
		return nil
	}
	full, err := m.sess.GetArticle(item.ID)
	if err == nil && full != nil {
		m.article = m.newArticleState(*full)
		m.view = viewArticle
		return m.fireImageLoad(*full)
	}
	m.article = m.newArticleState(store.Article{ID: item.ID, Title: item.Title, Read: item.Read})
	m.view = viewArticle
	return nil
}

func (m *Model) newArticleState(a store.Article) articleState {
	// A different article (or the first) opens a fresh load frontier and
	// deferred-native-render set; recomposes and resizes of the same article
	// keep them so the frontier stays monotonic within the article open.
	if m.article.article == nil || a.ID != m.article.id {
		m.imgFrontier = -1
		m.nativePending = make(map[string]bool)
	}
	padX, padY := m.sess.Config().Display.PaddingX, m.sess.Config().Display.PaddingY
	contentW, vpH, _ := render.ContentGeom(m.width, m.height, padX, padY)
	return m.composeArticle(a, m.deriveArticle(a, contentW, vpH), contentW, vpH)
}

// articleDerivation holds the expensive, content-only display derivation of an
// article — the converted markdown body, rendered text segments, inline-image
// list, harvested links, and rendered header — that image-load recomposes must
// not re-run. Its rendered pieces depend only on the article content and the
// content width; the viewport height feeds the key because the composition that
// consumes it fits image blocks against the viewport.
type articleDerivation struct {
	content     string
	header      string
	headerLines int
	bodyMD      string
	segments    []string
	inline      []convert.InlineImage
	imageURLs   []string
	links       []articleLink
	captions    map[string]string
	leadURL     string
}

// derivationKey identifies a cached articleDerivation: the article (id and
// source content), the content geometry, and the renderer inputs that change
// the derived output (ascii mode, palette, and the glamour theme style).
type derivationKey struct {
	id       int64
	contentW int
	vpH      int
	ascii    bool
	palette  render.Palette
	style    string
	content  string
}

// deriveArticle returns the article's derived display content, reusing the
// cached single entry when the article and all derivation inputs match. A
// stale entry (different article, geometry, or renderer inputs) is replaced.
func (m *Model) deriveArticle(a store.Article, contentW, vpH int) articleDerivation {
	key := derivationKey{
		id:       a.ID,
		contentW: contentW,
		vpH:      vpH,
		ascii:    m.ascii,
		palette:  m.palette,
		style:    compose.GlamourStandardStyle(m.sess.Config().Display.Theme),
		content:  a.Content,
	}
	if m.derivValid && m.derivKey == key {
		return m.derivCache
	}
	der := m.computeDerivation(a, contentW)
	m.derivCache = der
	m.derivKey = key
	m.derivValid = true
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
func (m *Model) computeDerivation(a store.Article, contentW int) articleDerivation {
	a.Title = stripSoftHyphens(a.Title)
	a.Author = stripSoftHyphens(a.Author)
	a.Link = stripSoftHyphens(a.Link)
	a.Content = stripSoftHyphens(a.Content)
	header := m.renderHeader(a, contentW)
	headerLines := 0
	if header != "" {
		headerLines = len(strings.Split(header, "\n"))
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
		segments = append(segments, m.renderMarkdown(bodyMD, contentW))
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
				segments = append(segments, m.renderMarkdown(bodyMD[pos:start], contentW))
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
			segments = append(segments, m.renderMarkdown(bodyMD[pos:], contentW))
		}
	}
	imageURLs := []string{}
	if leadURL != "" {
		imageURLs = append(imageURLs, leadURL)
	}
	for _, im := range inline {
		imageURLs = append(imageURLs, im.URL)
	}
	return articleDerivation{
		content:     a.Content,
		header:      header,
		headerLines: headerLines,
		bodyMD:      bodyMD,
		segments:    segments,
		inline:      inline,
		imageURLs:   imageURLs,
		links:       harvestArticleLinks(a),
		captions:    captions,
		leadURL:     leadURL,
	}
}

// composeArticle builds the article view state from the cached derivation:
// it renders the image blocks (served from the image caches), splices them
// into the derived text segments at the sentinel positions, and builds the
// viewport. This is the per-recompose part — image blocks change as loads land
// while the derived body stays put.
func (m *Model) composeArticle(a store.Article, der articleDerivation, contentW, vpH int) articleState {
	var parts []string
	var blocks []compose.ImageBlock
	var captions map[string]string
	var anchors []int
	lineCount := 0
	addPart := func(s string) {
		parts = append(parts, s)
		lineCount += len(strings.Split(s, "\n"))
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
		if m.imagesEnabled() {
			if block, rows, native, id := m.composeImageBlock(a.ImageURL, compose.ResolveImageAttribution(a, a.ImageURL, "", "", true), contentW, vpH, der.headerLines); len(block) > 0 {
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
		bodyParts, bodyBlocks, bodyAnchors, inlineCaps := m.spliceBody(der, contentW, vpH, der.headerLines, lineCount)
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
	st := articleState{
		id:           a.ID,
		article:      &a,
		lines:        strings.Split(rendered, "\n"),
		viewport:     vp,
		headerLines:  der.headerLines,
		imgStart:     imgStart,
		imgEnd:       imgEnd,
		capStart:     capStart,
		nativeImg:    nativeImg,
		imageBlocks:  blocks,
		imageURLs:    der.imageURLs,
		anchorRows:   anchors,
		inlineImages: der.inline,
		inlineCaps:   captions,
		links:        der.links,
		leadShown:    leadShown,
	}
	if len(st.lines) <= vpH {
		st.markRead(m.sess.Store())
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
func (m *Model) spliceBody(der articleDerivation, contentW, vpH, headerLines, base int) ([]string, []compose.ImageBlock, []int, map[string]string) {
	var parts []string
	var blocks []compose.ImageBlock
	var anchors []int
	lineCount := base
	segIdx := 0
	addPart := func(s string) {
		parts = append(parts, s)
		lineCount += len(strings.Split(s, "\n"))
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
		if block, rows, native, id := m.composeImageBlock(url, der.captions[url], contentW, vpH, headerLines); len(block) > 0 {
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

// renderHeader renders the reader header: the article URL as the very first
// line, a blank line, the bold title, and the author line directly beneath it
// with no blank between them. The URL is shown with its scheme stripped but is
// wrapped in an OSC 8 hyperlink whose target keeps the full URL, so it stays
// clickable. A long URL wraps at the content width onto subsequent lines so the
// whole display string is visible, each wrapped line carrying its own OSC 8
// hyperlink (some terminals drop hyperlinks across a line break). Title and
// author are rendered as separate documents and joined because the markdown
// renderer folds line breaks inside a paragraph into spaces.
func (m *Model) renderHeader(a store.Article, contentW int) string {
	var meta []string
	if a.Title != "" {
		meta = append(meta, m.renderMarkdown("**"+compose.EscapeMarkdownText(a.Title)+"**", contentW))
	}
	if a.Author != "" {
		meta = append(meta, m.renderMarkdown("by "+compose.EscapeMarkdownText(a.Author), contentW))
	}
	metaBlock := strings.Join(meta, "\n")
	if a.Link == "" {
		return metaBlock
	}
	urlPart := m.renderHeaderLink(a.Link, contentW)
	if metaBlock == "" {
		return urlPart
	}
	return urlPart + "\n\n" + metaBlock
}

// renderHeaderLink wraps the stripped article URL at the content width and
// emits each line as an OSC 8 hyperlink carrying the full URL, rendered in the
// status-bar blue role.
func (m *Model) renderHeaderLink(url string, contentW int) string {
	var parts []string
	for _, l := range image.LayoutWrapLines(stripURLScheme(url), contentW) {
		parts = append(parts, m.styles.status.Render(ansi.SetHyperlink(url)+l+ansi.ResetHyperlink()))
	}
	return strings.Join(parts, "\n")
}

func (m *Model) renderArticle() string {
	st := &m.article
	a := st.article
	date := ""
	if !a.PublishedAt.IsZero() {
		date = a.PublishedAt.Local().Format("2006-01-02 15:04:05")
	}
	title := a.Title
	if title == "" {
		title = "(untitled)"
	}
	g := m.glyphs()
	content := m.article.viewport.View()
	lines := strings.Split(content, "\n")
	m.rewriteNativeReShows(lines)
	m.suppressClippedNativeTransmits(lines)
	lines = highlightSelection(lines, m.article.sel)
	sc := render.ScrollState{
		TotalH:    m.article.viewport.TotalLineCount(),
		ViewportH: m.article.viewport.Height,
		Offset:    m.article.viewport.YOffset,
	}
	return render.RenderArticleBorder(m.width, m.height, m.sess.Config().Display.PaddingX, m.sess.Config().Display.PaddingY,
		g, m.palette, date, title, sc, lines)
}

// highlightSelection applies the inverted selection style to the selected cell
// range of each visible content line. Lines outside the range are returned
// unchanged; ANSI and OSC 8 escape sequences on a line are preserved by
// ansi.Cut/ansi.Truncate. The reverse-video SGR is emitted directly so the
// highlight works regardless of the terminal profile (lipgloss's Reverse is a
// no-op when no TTY is detected).
