package ui

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"yerss/convert"
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
	inlineImages []convert.InlineImage
	inlineCaps   map[string]string
	links        []articleLink
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
	padX, padY := m.sess.Config().Display.PaddingX, m.sess.Config().Display.PaddingY
	contentW, vpH, _ := render.ContentGeom(m.width, m.height, padX, padY)
	headerLines := 0
	header := m.renderHeader(a, contentW)
	if header != "" {
		headerLines = len(strings.Split(header, "\n"))
	}

	var parts []string
	var blocks []compose.ImageBlock
	var captions map[string]string
	lineCount := 0
	addPart := func(s string) {
		parts = append(parts, s)
		lineCount += len(strings.Split(s, "\n"))
	}
	addImageBlock := func(url string, block []string, rows int, native bool, id uint32) {
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
	}

	if header != "" {
		addPart(header)
		parts = append(parts, "")
		lineCount++
	}
	if a.ImageURL != "" && m.imagesEnabled() {
		if block, rows, native, id := m.composeImageBlock(a.ImageURL, compose.ResolveImageAttribution(a, a.ImageURL, "", "", true), contentW, vpH, headerLines); len(block) > 0 {
			addImageBlock(a.ImageURL, block, rows, native, id)
		}
	}

	bodyMD, inline := convert.ConvertImages(a.Content)
	inline = compose.DedupeInline(a.ImageURL, inline)
	bodyMD = compose.StripDuplicateSentinels(bodyMD, a.ImageURL)
	imageURLs := []string{}
	if a.ImageURL != "" {
		imageURLs = append(imageURLs, a.ImageURL)
	}
	for _, im := range inline {
		imageURLs = append(imageURLs, im.URL)
	}
	if len(inline) == 0 {
		addPart(m.renderMarkdown(bodyMD, contentW))
	} else {
		bodyParts, bodyBlocks, inlineCaps := m.inlineBodyParts(a, bodyMD, inline, contentW, vpH, headerLines, lineCount)
		parts = append(parts, bodyParts...)
		lineCount += compose.LineCountOf(bodyParts)
		blocks = append(blocks, bodyBlocks...)
		captions = inlineCaps
	}

	rendered := strings.Join(parts, "\n")
	vp := viewport.New(contentW, vpH)
	vp.SetContent(rendered)
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
		headerLines:  headerLines,
		imgStart:     imgStart,
		imgEnd:       imgEnd,
		capStart:     capStart,
		nativeImg:    nativeImg,
		imageBlocks:  blocks,
		imageURLs:    imageURLs,
		inlineImages: inline,
		inlineCaps:   captions,
		links:        harvestArticleLinks(a),
	}
	if len(st.lines) <= vpH {
		st.markRead(m.sess.Store())
	}
	return st
}

// inlineBodyParts renders the article body with inline image blocks interleaved
// at the sentinel positions ConvertImages emitted. It returns the body parts
// (each joined into the final content with newline separators) and the ordered
// inline image blocks with content-relative line indices. An inline image whose
// source is not yet composed renders a blank placeholder line in its place so
// the text flow stays stable until its load lands and re-composes the block. A
// sentinel wrapped in a markdown link (`<a><img></a>`) consumes the whole link
// construct, so no stray `[`/`](...)` renders around the image.
func (m *Model) inlineBodyParts(a store.Article, md string, inline []convert.InlineImage, contentW, vpH, headerLines, base int) ([]string, []compose.ImageBlock, map[string]string) {
	var parts []string
	var blocks []compose.ImageBlock
	captions := map[string]string{}
	lineCount := base
	addPart := func(s string) {
		parts = append(parts, s)
		lineCount += len(strings.Split(s, "\n"))
	}
	addImageBlock := func(url string, block []string, rows int, native bool, id uint32) {
		if len(parts) > 0 && parts[len(parts)-1] != "" {
			parts = append(parts, "")
			lineCount++
		}
		start := lineCount
		addPart(strings.Join(block, "\n"))
		blocks = append(blocks, compose.ImageBlock{URL: url, ImgStart: start, CapStart: start + rows, ImgEnd: start + len(block) - 1, NativeImg: native, NativeID: id})
		parts = append(parts, "")
		lineCount++
	}
	pos := 0
	idx := 0
	for {
		start, end, url, linkText, ok := compose.NextSentinel(md, pos)
		if !ok {
			break
		}
		if seg := strings.TrimSpace(md[pos:start]); seg != "" {
			addPart(m.renderMarkdown(md[pos:start], contentW))
		}
		alt := ""
		if idx < len(inline) && inline[idx].URL == url {
			alt = inline[idx].Alt
			idx++
		}
		attr := compose.ResolveImageAttribution(a, url, alt, linkText, false)
		captions[url] = attr
		if block, rows, native, id := m.composeImageBlock(url, attr, contentW, vpH, headerLines); len(block) > 0 {
			addImageBlock(url, block, rows, native, id)
		} else {
			// Placeholder: keep the paragraph break so the body text stays
			// stable and the block inserts at the right offset when the load
			// lands and the article re-composes.
			if len(parts) == 0 || parts[len(parts)-1] != "" {
				parts = append(parts, "")
				lineCount++
			}
		}
		pos = end
	}
	if rest := strings.TrimSpace(md[pos:]); rest != "" {
		addPart(m.renderMarkdown(rest, contentW))
	}
	return parts, blocks, captions
}

// renderHeader renders the reader header: the article URL as the very first
// line (rendered exactly once by the markdown renderer's autolink handling),
// a blank line, the bold title, and the author line directly beneath it with
// no blank between them. Title and author are rendered as separate documents
// and joined because the markdown renderer folds line breaks inside a
// paragraph into spaces.
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
	urlPart := m.renderMarkdown(compose.HeaderLink(a.Link), contentW)
	if metaBlock == "" {
		return urlPart
	}
	return urlPart + "\n\n" + metaBlock
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
