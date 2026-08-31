package ui

import (
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"yerss/convert"
	"yerss/internal/config"
	"yerss/internal/store"
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
	imageBlocks  []imageBlock
	imageURLs    []string
	inlineImages []convert.InlineImage
	inlineCaps   map[string]string
	links        []articleLink
}

// imageBlock is one image block composed into the article content: the lead
// image (index 0, below the header) and each inline image, in document order.
// imgStart is the first image row, capStart the first attribution line (the
// down-snap target), and imgEnd the last composed line of the block. An
// inline block carries nativeImg when its lines are a terminal-side native
// placement.
type imageBlock struct {
	url       string
	imgStart  int
	capStart  int
	imgEnd    int
	nativeImg bool
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
		add(cleanLinkText(sentinelRe.ReplaceAllString(l.text, "")), resolveLink(a, l.url))
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
type textSelection struct {
	active   bool
	tracking bool
	anchorX  int
	anchorY  int
	curX     int
	curY     int
}

// cell is a content-area coordinate.
type cell struct {
	x, y int
}

// rangeFor returns the normalized selection range: lo is the upper-left cell
// and hi the lower-right cell of the selected block.
func (s textSelection) rangeFor() (lo, hi cell) {
	lo = cell{x: s.anchorX, y: s.anchorY}
	hi = cell{x: s.curX, y: s.curY}
	if lo.y > hi.y || (lo.y == hi.y && lo.x > hi.x) {
		lo, hi = hi, lo
	}
	return lo, hi
}

// selectionRange returns the selected column range [from, to) on one row: the
// anchor/cur cells when the selection spans a single row, the full line width
// on the first and last rows of a multi-row selection, and the full line in
// between. lineWidth must be the display width of the row's visible text. It is
// the single source of truth shared by the copy and highlight paths so they
// cannot drift apart.
func selectionRange(lo, hi cell, row, lineWidth int) (from, to int) {
	if lo.y == hi.y {
		return lo.x, hi.x
	}
	if row == lo.y {
		return lo.x, lineWidth
	}
	if row == hi.y {
		return 0, hi.x
	}
	return 0, lineWidth
}

func (m *Model) openArticle() tea.Cmd {
	item, ok := m.articleAtCursor()
	if !ok {
		return nil
	}
	full, err := m.store.GetArticle(item.ID)
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
	padX, padY := m.cfg.Display.PaddingX, m.cfg.Display.PaddingY
	contentW, vpH, _ := contentGeom(m.width, m.height, padX, padY)
	headerLines := 0
	header := m.renderHeader(a, contentW)
	if header != "" {
		headerLines = len(strings.Split(header, "\n"))
	}

	var parts []string
	var blocks []imageBlock
	var captions map[string]string
	lineCount := 0
	addPart := func(s string) {
		parts = append(parts, s)
		lineCount += len(strings.Split(s, "\n"))
	}
	addImageBlock := func(url string, block []string, rows int, native bool) {
		if len(parts) > 0 && parts[len(parts)-1] != "" {
			parts = append(parts, "")
			lineCount++
		} else if len(parts) == 0 {
			parts = append(parts, "")
			lineCount++
		}
		start := lineCount
		addPart(strings.Join(block, "\n"))
		blocks = append(blocks, imageBlock{url: url, imgStart: start, capStart: start + rows, imgEnd: start + len(block) - 1, nativeImg: native})
		parts = append(parts, "")
		lineCount++
	}

	if header != "" {
		addPart(header)
		parts = append(parts, "")
		lineCount++
	}
	if a.ImageURL != "" && m.imagesEnabled() {
		if block, rows, native := m.composeImageBlock(a.ImageURL, articleAttribution(a), contentW, vpH, headerLines); len(block) > 0 {
			addImageBlock(a.ImageURL, block, rows, native)
		}
	}

	bodyMD, inline := convert.ConvertImages(a.Content)
	inline = dedupeInline(a.ImageURL, inline)
	bodyMD = stripDuplicateSentinels(bodyMD, a.ImageURL)
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
		lineCount += lineCountOf(bodyParts)
		blocks = append(blocks, bodyBlocks...)
		captions = inlineCaps
	}

	rendered := strings.Join(parts, "\n")
	vp := viewport.New(contentW, vpH)
	vp.SetContent(rendered)
	imgStart, imgEnd, capStart, nativeImg := -1, -1, -1, false
	if len(blocks) > 0 {
		imgStart, imgEnd, capStart = blocks[0].imgStart, blocks[0].imgEnd, blocks[0].capStart
		nativeImg = blocks[0].nativeImg
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
		st.markRead(m.store)
	}
	return st
}

// lineCountOf returns the number of lines the given parts occupy when joined
// with a newline separator.
func lineCountOf(parts []string) int {
	n := 0
	for _, p := range parts {
		n += len(strings.Split(p, "\n"))
	}
	return n
}

// dedupeInline drops inline images whose URL is the lead image URL or a URL
// already rendered, keeping the first occurrence in document order, so a photo
// attached as the lead and repeated in the body is not shown twice.
func dedupeInline(leadURL string, inline []convert.InlineImage) []convert.InlineImage {
	seen := map[string]bool{}
	if leadURL != "" {
		seen[leadURL] = true
	}
	kept := make([]convert.InlineImage, 0, len(inline))
	for _, im := range inline {
		if seen[im.URL] {
			continue
		}
		seen[im.URL] = true
		kept = append(kept, im)
	}
	return kept
}

// sentinelRe matches the inline-image sentinel token "\x00img:<url>\x00" that
// ConvertImages emits for each <img>, capturing the image URL.
var sentinelRe = regexp.MustCompile(`\x00img:([^\x00]*)\x00`)

// nextSentinel returns the byte span [start, end) of the next inline-image
// sentinel at or after from, together with the image's URL and the text of any
// markdown link the sentinel opens (`<a><img>...text...</a>` renders as
// `[SENTINEL link text](href)`). When the sentinel is link-wrapped, the span is
// extended to consume the whole link construct so no stray `[`, link text, or
// `](...)` fragments render around the image block, and the link text is
// reported for use as the image's caption fallback. ok is false when no
// sentinel remains.
func nextSentinel(md string, from int) (start, end int, url, linkText string, ok bool) {
	loc := sentinelRe.FindStringSubmatchIndex(md[from:])
	if loc == nil {
		return 0, 0, "", "", false
	}
	base := from
	s := base + loc[0]
	e := base + loc[1]
	url = md[base+loc[2] : base+loc[3]]
	// A sentinel immediately preceded by '[' opens the markdown link the <a>
	// renderer produced; '![' (image syntax) is not a link and is left alone.
	if s > 0 && md[s-1] == '[' && (s < 2 || md[s-2] != '!') {
		s--
		if closeIdx := strings.Index(md[e:], "]("); closeIdx >= 0 {
			linkText = cleanLinkText(md[e : e+closeIdx])
			if n := linkDestEnd(md[e+closeIdx+2:]); n >= 0 {
				e += closeIdx + 2 + n
			} else {
				e += closeIdx + 2
			}
		}
	} else if e < len(md) && strings.HasPrefix(md[e:], "](") {
		// A bare '](' immediately after the sentinel outside a link we opened:
		// consume it as a link tail for safety.
		if n := linkDestEnd(md[e+2:]); n >= 0 {
			e += 2 + n
		}
	}
	return s, e, url, linkText, true
}

// cleanLinkText turns raw markdown link text (which can carry hard-break
// backslashes and emphasis markers when the converter wraps block text inside
// a link) into a single whitespace-separated caption line.
func cleanLinkText(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	s = strings.NewReplacer("\\", "", "**", "", "*", "", "`", "").Replace(s)
	return strings.TrimSpace(s)
}

// linkDestEnd returns the length of a markdown link destination beginning at
// s, including its closing ')', or -1 when the destination is unterminated. It
// honors angle-bracketed destinations and balanced parentheses.
func linkDestEnd(s string) int {
	if strings.HasPrefix(s, "<") {
		for i := 1; i < len(s); i++ {
			if s[i] == '>' {
				if i+1 < len(s) && s[i+1] == ')' {
					return i + 2
				}
				return -1
			}
		}
		return -1
	}
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			if depth == 0 {
				return i + 1
			}
			depth--
		}
	}
	return -1
}

// stripDuplicateSentinels removes every inline-image sentinel whose URL has
// already appeared (the lead image URL, or an earlier occurrence in document
// order), keeping the first occurrence of each image, along with the markdown
// link wrapping a removed sentinel. A skipped duplicate renders no block while
// the surrounding paragraphs keep their separation.
func stripDuplicateSentinels(md, leadURL string) string {
	seen := map[string]bool{}
	if leadURL != "" {
		seen[leadURL] = true
	}
	var b strings.Builder
	pos := 0
	for {
		start, end, url, _, ok := nextSentinel(md, pos)
		if !ok {
			b.WriteString(md[pos:])
			break
		}
		b.WriteString(md[pos:start])
		if !seen[url] {
			b.WriteString(md[start:end])
		}
		seen[url] = true
		pos = end
	}
	return b.String()
}

// inlineBodyParts renders the article body with inline image blocks interleaved
// at the sentinel positions ConvertImages emitted. It returns the body parts
// (each joined into the final content with newline separators) and the ordered
// inline image blocks with content-relative line indices. An inline image whose
// source is not yet composed renders a blank placeholder line in its place so
// the text flow stays stable until its load lands and re-composes the block. A
// sentinel wrapped in a markdown link (`<a><img></a>`) consumes the whole link
// construct, so no stray `[`/`](...)` renders around the image.
func (m *Model) inlineBodyParts(a store.Article, md string, inline []convert.InlineImage, contentW, vpH, headerLines, base int) ([]string, []imageBlock, map[string]string) {
	var parts []string
	var blocks []imageBlock
	captions := map[string]string{}
	lineCount := base
	addPart := func(s string) {
		parts = append(parts, s)
		lineCount += len(strings.Split(s, "\n"))
	}
	addImageBlock := func(url string, block []string, rows int, native bool) {
		if len(parts) > 0 && parts[len(parts)-1] != "" {
			parts = append(parts, "")
			lineCount++
		}
		start := lineCount
		addPart(strings.Join(block, "\n"))
		blocks = append(blocks, imageBlock{url: url, imgStart: start, capStart: start + rows, imgEnd: start + len(block) - 1, nativeImg: native})
		parts = append(parts, "")
		lineCount++
	}
	pos := 0
	idx := 0
	for {
		start, end, url, linkText, ok := nextSentinel(md, pos)
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
		attr := inlineAttribution(a, url, alt, linkText)
		captions[url] = attr
		if block, rows, native := m.composeImageBlock(url, attr, contentW, vpH, headerLines); len(block) > 0 {
			addImageBlock(url, block, rows, native)
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

// inlineAttribution returns the caption for an inline image: the image's alt
// text when present, otherwise the credit extracted from the article's HTML
// when present, otherwise the text of any markdown link that wraps the image
// (promo images like the TomDispatch banner render as a linked image whose
// link text labels it), otherwise "photo: <source>" derived by the list view's
// source-identifier rules. It returns "" when none is available.
func inlineAttribution(a store.Article, url, alt, linkText string) string {
	if alt != "" {
		return alt
	}
	if credit := convert.ImageCredit(a.Content, url); credit != "" {
		return credit
	}
	if linkText != "" {
		return linkText
	}
	if src := sourceID(a.Link, a.FeedURL); src != "" {
		return "photo: " + src
	}
	return ""
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
		meta = append(meta, m.renderMarkdown("**"+escapeMarkdownText(a.Title)+"**", contentW))
	}
	if a.Author != "" {
		meta = append(meta, m.renderMarkdown("by "+escapeMarkdownText(a.Author), contentW))
	}
	metaBlock := strings.Join(meta, "\n")
	if a.Link == "" {
		return metaBlock
	}
	urlPart := m.renderMarkdown(headerLink(a.Link), contentW)
	if metaBlock == "" {
		return urlPart
	}
	return urlPart + "\n\n" + metaBlock
}

// headerLink renders a URL the markdown renderer prints exactly once. The
// glamour renderer emits `[text](url)` as "text url", so the URL is emitted
// bare and left to autolink detection; the angle-bracket form is used only
// when the URL contains characters that would break plain autolink parsing.
func headerLink(url string) string {
	if strings.ContainsAny(url, " <>") || strings.Contains(url, ")") {
		return "<" + url + ">"
	}
	return url
}

func (m *Model) updateArticle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	act, ok := m.keys[config.ViewArticle][msg.String()]
	if !ok {
		return m, nil
	}
	var cmd tea.Cmd
	switch act {
	case config.Quit:
		m.persistSelection()
		return m, tea.Quit
	case config.Back, config.CloseArticle:
		m.backToList()
	case config.LinkPopup:
		m.openLinksPopup()
	case config.MoveDown:
		m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
		m.article.markRead(m.store)
	case config.MoveUp:
		m.scrollArticle(func() { m.article.viewport.ScrollUp(1) }, true)
	case config.PageDown:
		m.scrollArticle(func() { m.article.viewport.PageDown() }, false)
		m.article.markRead(m.store)
	case config.PageUp:
		m.scrollArticle(func() { m.article.viewport.PageUp() }, false)
	case config.HalfPageDown:
		m.scrollArticle(func() { m.article.viewport.HalfPageDown() }, false)
		m.article.markRead(m.store)
	case config.HalfPageUp:
		m.scrollArticle(func() { m.article.viewport.HalfPageUp() }, false)
	case config.Top:
		m.article.viewport.GotoTop()
	case config.Bottom:
		m.scrollArticle(func() { m.article.viewport.GotoBottom() }, false)
	case config.OpenURL:
		if m.article.article.Link != "" {
			cmd = openURLCmd(m.article.article.Link)
		}
	case config.CopyURL:
		if m.article.article.Link != "" {
			cmd = copyURLCmd(m.article.article.Link)
		}
	case config.CopyArticleText:
		cmd = copyTextCmd(ansi.Strip(strings.Join(m.article.lines, "\n")), "copied article text")
	case config.TagPopup:
		m.openTagPopup(popupTagsArticle)
	case config.Help:
		m.popup = popupHelp
	}
	return m, cmd
}

// updateArticleMouse handles mouse events in the article view: wheel up/down
// scrolls the viewport by one line, with downward scroll marking the article
// read like the Down key. An unmodified left-button press anchors a text
// selection, drag extends it, and release copies the selected text to the
// clipboard and clears the selection. An unmodified right-button press
// returns to the list view like the Back keys. Presses carrying a Ctrl/Alt/
// Shift modifier are not interpreted, so the terminal can activate OSC 8
// hyperlinks natively.
func (m *Model) updateArticleMouse(msg tea.MouseMsg) tea.Cmd {
	if tea.MouseEvent(msg).IsWheel() && msg.Action == tea.MouseActionPress {
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.scrollArticle(func() { m.article.viewport.ScrollUp(1) }, true)
		case tea.MouseButtonWheelDown:
			m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
			m.article.markRead(m.store)
		}
		return nil
	}
	if msg.Shift || msg.Alt || msg.Ctrl {
		return nil
	}
	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonRight {
			m.backToList()
			return nil
		}
		if msg.Button != tea.MouseButtonLeft {
			return nil
		}
		if c, ok := m.contentCell(msg.X, msg.Y); ok {
			if url := m.linkAtContentCell(c.x, c.y); url != "" {
				return openURLCmd(url)
			}
			m.article.sel = textSelection{
				active:   true,
				tracking: true,
				anchorX:  c.x,
				anchorY:  c.y,
				curX:     c.x,
				curY:     c.y,
			}
		}
	case tea.MouseActionMotion:
		if !m.article.sel.tracking {
			return nil
		}
		c := m.contentCellClamped(msg.X, msg.Y)
		m.article.sel.curX, m.article.sel.curY = c.x, c.y
	case tea.MouseActionRelease:
		if !m.article.sel.tracking {
			return nil
		}
		text := m.article.selectedText()
		m.article.sel = textSelection{}
		if text != "" {
			return copyTextCmd(text, "copied selection")
		}
	}
	return nil
}

// backToList returns from the article view to the list view: the previously
// viewed article stays selected and any active mouse selection is cleared.
func (m *Model) backToList() {
	m.view = viewList
	m.article.sel = textSelection{}
	m.loadList()
}

// contentRect returns the screen-space rectangle of the article content area
// (origin column/row plus width/height in cells), matching renderArticleBorder:
// content begins at column padX+1 and row 1, directly beneath the top border
// with no padding above it.
func (m *Model) contentRect() (x0, y0, w, h int) {
	padX := m.cfg.Display.PaddingX
	w, h, _ = contentGeom(m.width, m.height, padX, m.cfg.Display.PaddingY)
	return 1 + padX, 1, w, h
}

// contentCell maps a screen cell to a content-area cell, reporting ok=false
// when the press is outside the content area (border, padding, status bar).
func (m *Model) contentCell(x, y int) (cell, bool) {
	x0, y0, w, h := m.contentRect()
	if x < x0 || x >= x0+w || y < y0 || y >= y0+h {
		return cell{}, false
	}
	return cell{x: x - x0, y: y - y0}, true
}

// contentCellClamped maps a screen cell to a content-area cell, clamping the
// result into the content area so a drag that leaves it stays in range.
func (m *Model) contentCellClamped(x, y int) cell {
	x0, y0, w, h := m.contentRect()
	cx := x - x0
	if cx < 0 {
		cx = 0
	}
	if cx >= w {
		cx = w - 1
	}
	cy := y - y0
	if cy < 0 {
		cy = 0
	}
	if cy >= h {
		cy = h - 1
	}
	return cell{x: cx, y: cy}
}

// linkSpan is a hyperlink covering the display columns [start, end) on a
// single rendered line.
type linkSpan struct {
	start, end int
	url        string
}

// linkAtContentCell reports the URL of any hyperlink covering content cell
// (x, y), or "" when the cell is not on a link. The cell is indexed against the
// viewport's visible lines, matching the layout the user sees.
func (m *Model) linkAtContentCell(x, y int) string {
	visible := strings.Split(m.article.viewport.View(), "\n")
	if y < 0 || y >= len(visible) {
		return ""
	}
	for _, sp := range parseLinkSpans(visible[y]) {
		if x >= sp.start && x < sp.end {
			return sp.url
		}
	}
	return ""
}

// parseLinkSpans scans a rendered ANSI line for OSC 8 hyperlink escapes and
// returns the display-column span each URL covers. It tracks the current
// column as it walks printable runes (honoring wide characters), clips spans
// when a link is reset, and skips CSI/SGR and other escape sequences. The
// terminator may be BEL or the ESC-backslash string terminator.
func parseLinkSpans(line string) []linkSpan {
	var spans []linkSpan
	url := ""
	start := 0
	flush := func(x int) {
		if url != "" {
			spans = append(spans, linkSpan{start: start, end: x, url: url})
		}
	}
	x := 0
	for i := 0; i < len(line); {
		switch line[i] {
		case 0x1b:
			if i+1 < len(line) && line[i+1] == ']' {
				end := oscEnd(line, i+2)
				if end < 0 {
					flush(x)
					return spans
				}
				payload := line[i+2 : end]
				flush(x)
				if strings.HasPrefix(payload, "8;") {
					rest := payload[2:]
					u := rest
					if idx := strings.LastIndex(rest, ";"); idx >= 0 {
						u = rest[idx+1:]
					}
					if u == "" {
						url = ""
					} else {
						url = u
						start = x
					}
				}
				if line[end] == 0x07 {
					i = end + 1
				} else {
					i = end + 2
				}
				continue
			}
			// Any other escape/CSI sequence: skip it via the shared helper.
			i = skipEscape(line, i)
			continue
		default:
			r, size := utf8.DecodeRuneInString(line[i:])
			x += cellWidth(r)
			i += size
		}
	}
	flush(x)
	return spans
}

// oscEnd returns the index of the terminator that closes an OSC string
// beginning at from: either a BEL byte or the index of the ESC in the
// ESC-backslash string terminator. It returns -1 when no terminator is found.
func oscEnd(line string, from int) int {
	for i := from; i < len(line); i++ {
		if line[i] == 0x07 {
			return i
		}
		if line[i] == 0x1b && i+1 < len(line) && line[i+1] == '\\' {
			return i
		}
	}
	return -1
}

// selectedText returns the plain text within the selection range, stripped of
// ANSI styling and OSC 8 escapes, preserving the wrapped line breaks. It
// returns "" for an inactive or zero-width selection.
func (s *articleState) selectedText() string {
	if !s.sel.active {
		return ""
	}
	visible := strings.Split(s.viewport.View(), "\n")
	lo, hi := s.sel.rangeFor()
	var b strings.Builder
	for row := lo.y; row <= hi.y; row++ {
		if row < 0 || row >= len(visible) {
			continue
		}
		line := ansi.Strip(visible[row])
		from, to := selectionRange(lo, hi, row, ansi.StringWidth(line))
		if from < to {
			slice := ansi.Cut(line, from, to)
			if row < hi.y {
				slice = strings.TrimRight(slice, " ")
			}
			b.WriteString(slice)
		}
		if row < hi.y {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// markRead implements the read-status rules: an article that fits the
// viewport is marked read on open; a longer article is marked read on the
// first downward scroll.
func (s *articleState) markRead(st *store.Store) {
	if s.readMarked || s.article.Read {
		s.readMarked = true
		return
	}
	s.readMarked = true
	s.article.Read = true
	_ = st.SetRead(s.id, true)
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
	sc := scrollState{
		totalH:    m.article.viewport.TotalLineCount(),
		viewportH: m.article.viewport.Height,
		offset:    m.article.viewport.YOffset,
	}
	return renderArticleBorder(m.width, m.height, m.cfg.Display.PaddingX, m.cfg.Display.PaddingY,
		g, m.palette, date, title, sc, lines)
}

// highlightSelection applies the inverted selection style to the selected cell
// range of each visible content line. Lines outside the range are returned
// unchanged; ANSI and OSC 8 escape sequences on a line are preserved by
// ansi.Cut/ansi.Truncate. The reverse-video SGR is emitted directly so the
// highlight works regardless of the terminal profile (lipgloss's Reverse is a
// no-op when no TTY is detected).
func highlightSelection(lines []string, sel textSelection) []string {
	if !sel.active {
		return lines
	}
	lo, hi := sel.rangeFor()
	const revOn, revOff = "\x1b[7m", "\x1b[0m"
	out := make([]string, len(lines))
	for i, line := range lines {
		if i < lo.y || i > hi.y {
			out[i] = line
			continue
		}
		width := ansi.StringWidth(line)
		from, to := selectionRange(lo, hi, i, width)
		if to > width {
			to = width
		}
		if from >= to {
			out[i] = line
			continue
		}
		before := ansi.Truncate(line, from, "")
		selSeg := ansi.Cut(line, from, to)
		after := ansi.Cut(line, to, width)
		out[i] = before + revOn + selSeg + revOff + after
	}
	return out
}
