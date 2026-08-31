package ui

import (
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
	id          int64
	article     *store.Article
	lines       []string
	viewport    viewport.Model
	readMarked  bool
	sel         textSelection
	headerLines int
	imgStart    int
	imgEnd      int
	capStart    int
	nativeImg   bool
	links       []articleLink
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
// the markdown links, not the article's own header URL — deduplicated by URL.
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
	for _, l := range harvestLinks(convert.Convert(a.Content)) {
		add(l.text, l.url)
	}
	return out
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
	imgStart, imgEnd, capStart := -1, -1, -1
	headerLines := 0
	header := m.renderHeader(a, contentW)
	if header != "" {
		headerLines = len(strings.Split(header, "\n"))
	}
	imgBlock, imgRows, imgNative := m.articleImageBlock(a, contentW, vpH, headerLines)
	body := m.renderMarkdown(convert.Convert(a.Content), contentW)
	rendered := body
	if header != "" {
		rendered = header + "\n\n" + body
	}
	if len(imgBlock) > 0 {
		// The block sits below the header with one blank line on each side;
		// it covers the image lines plus the attribution lines, and capStart
		// (the down-snap target) is the first attribution line, or the line
		// past the block when no attribution is composed.
		imgStart, imgEnd = headerLines+1, headerLines+len(imgBlock)
		capStart = imgStart + imgRows
		rendered = header + "\n\n" + strings.Join(imgBlock, "\n") + "\n\n" + body
	}
	vp := viewport.New(contentW, vpH)
	vp.SetContent(rendered)
	st := articleState{
		id:          a.ID,
		article:     &a,
		lines:       strings.Split(rendered, "\n"),
		viewport:    vp,
		headerLines: headerLines,
		imgStart:    imgStart,
		imgEnd:      imgEnd,
		capStart:    capStart,
		nativeImg:   imgNative,
		links:       harvestArticleLinks(a),
	}
	if len(st.lines) <= vpH {
		st.markRead(m.store)
	}
	return st
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
		m.scrollArticle(func() { m.article.viewport.ScrollDown(1) })
		m.article.markRead(m.store)
	case config.MoveUp:
		m.scrollArticle(func() { m.article.viewport.ScrollUp(1) })
	case config.PageDown:
		m.scrollArticle(func() { m.article.viewport.PageDown() })
		m.article.markRead(m.store)
	case config.PageUp:
		m.scrollArticle(func() { m.article.viewport.PageUp() })
	case config.HalfPageDown:
		m.scrollArticle(func() { m.article.viewport.HalfPageDown() })
		m.article.markRead(m.store)
	case config.HalfPageUp:
		m.scrollArticle(func() { m.article.viewport.HalfPageUp() })
	case config.Top:
		m.article.viewport.GotoTop()
	case config.Bottom:
		m.scrollArticle(func() { m.article.viewport.GotoBottom() })
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
			m.scrollArticle(func() { m.article.viewport.ScrollUp(1) })
		case tea.MouseButtonWheelDown:
			m.scrollArticle(func() { m.article.viewport.ScrollDown(1) })
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
	lines := highlightSelection(strings.Split(content, "\n"), m.article.sel)
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
