package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"yerss/convert"
	"yerss/internal/config"
	"yerss/internal/store"
)

type articleState struct {
	id         int64
	article    *store.Article
	lines      []string
	viewport   viewport.Model
	readMarked bool
	sel        textSelection
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

func (m *Model) openArticle() {
	rows := m.visibleRows()
	if m.list.cursor < 0 || m.list.cursor >= len(rows) {
		return
	}
	row := rows[m.list.cursor]
	if row.kind != rowArticle {
		return
	}
	item := &m.list.groups[row.groupIdx].articles[row.artIdx]
	full, err := m.store.GetArticle(item.ID)
	if err == nil && full != nil {
		m.article = m.newArticleState(*full)
	} else {
		m.article = m.newArticleState(store.Article{ID: item.ID, Title: item.Title, Read: item.Read})
	}
	m.view = viewArticle
}

func (m *Model) newArticleState(a store.Article) articleState {
	padX, padY := m.cfg.Display.PaddingX, m.cfg.Display.PaddingY
	contentW := m.width - 2*padX - 2
	if contentW < 1 {
		contentW = 1
	}
	vpH := m.height - 2*padY - 2
	if vpH < 1 {
		vpH = 1
	}
	md := renderArticleMarkdown(a)
	rendered := m.renderMarkdown(md, contentW)
	vp := viewport.New(contentW, vpH)
	vp.SetContent(rendered)
	st := articleState{
		id:       a.ID,
		article:  &a,
		lines:    strings.Split(rendered, "\n"),
		viewport: vp,
	}
	if len(st.lines) <= vpH {
		st.markRead(m.store)
	}
	return st
}

// renderArticleMarkdown builds the article's markdown source: a header (bold
// title, italic author, and the article URL as a markdown link when present)
// followed by the converted HTML body. The glamour renderer styles the whole
// document, so links stay OSC 8 clickable.
func renderArticleMarkdown(a store.Article) string {
	var b strings.Builder
	if a.Title != "" {
		b.WriteString("**" + escapeMarkdownText(a.Title) + "**\n")
	}
	if a.Author != "" {
		b.WriteString("*by " + escapeMarkdownText(a.Author) + "*\n")
	}
	if a.Link != "" {
		b.WriteString(markdownLink(a.Link) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(convert.Convert(a.Content))
	return b.String()
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
	case config.Back:
		m.view = viewList
		m.article.sel = textSelection{}
		m.loadList()
	case config.MoveDown:
		m.article.viewport.ScrollDown(1)
		m.article.markRead(m.store)
	case config.MoveUp:
		m.article.viewport.ScrollUp(1)
	case config.PageDown:
		m.article.viewport.PageDown()
		m.article.markRead(m.store)
	case config.PageUp:
		m.article.viewport.PageUp()
	case config.HalfPageDown:
		m.article.viewport.HalfPageDown()
		m.article.markRead(m.store)
	case config.HalfPageUp:
		m.article.viewport.HalfPageUp()
	case config.Top:
		m.article.viewport.GotoTop()
	case config.Bottom:
		m.article.viewport.GotoBottom()
	case config.ToggleRead:
		m.article.toggleRead(m.store)
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
// clipboard and clears the selection. Presses carrying a Ctrl/Alt/Shift
// modifier are not interpreted, so the terminal can activate OSC 8 hyperlinks
// natively.
func (m *Model) updateArticleMouse(msg tea.MouseMsg) tea.Cmd {
	if tea.MouseEvent(msg).IsWheel() && msg.Action == tea.MouseActionPress {
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.article.viewport.ScrollUp(1)
		case tea.MouseButtonWheelDown:
			m.article.viewport.ScrollDown(1)
			m.article.markRead(m.store)
		}
		return nil
	}
	if msg.Shift || msg.Alt || msg.Ctrl {
		return nil
	}
	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button != tea.MouseButtonLeft {
			return nil
		}
		if c, ok := m.contentCell(msg.X, msg.Y); ok {
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

// contentRect returns the screen-space rectangle of the article content area
// (origin column/row plus width/height in cells), matching renderArticleBorder:
// content begins at column padX+1 and row effPadY+1.
func (m *Model) contentRect() (x0, y0, w, h int) {
	padX := m.cfg.Display.PaddingX
	effPadY := clampPadY(m.cfg.Display.PaddingY, m.height)
	w = m.width - 2*padX - 2
	if w < 1 {
		w = 1
	}
	h = m.height - 2*effPadY - 2
	if h < 1 {
		h = 1
	}
	return 1 + padX, 1 + effPadY, w, h
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
		var from, to int
		if lo.y == hi.y {
			from, to = lo.x, hi.x
		} else if row == lo.y {
			from, to = lo.x, ansi.StringWidth(line)
		} else if row == hi.y {
			from, to = 0, hi.x
		} else {
			from, to = 0, ansi.StringWidth(line)
		}
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

func (s *articleState) toggleRead(st *store.Store) {
	s.article.Read = !s.article.Read
	_ = st.SetRead(s.id, s.article.Read)
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
	g := glyphsFor(m.ascii)
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
		var from, to int
		if lo.y == hi.y {
			from, to = lo.x, hi.x
		} else if i == lo.y {
			from, to = lo.x, width
		} else if i == hi.y {
			from, to = 0, hi.x
		} else {
			from, to = 0, width
		}
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
