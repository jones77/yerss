package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"yerss/internal/config"
)

type listState struct {
	articles []articleItem
	cursor   int
	filter   string
}

type articleItem struct {
	ID    int64
	Title string
	Read  bool
}

func (m *Model) loadList() {
	arts, err := m.store.ListArticles(m.list.filter)
	if err != nil {
		m.setStatus("load error: " + err.Error())
		return
	}
	items := make([]articleItem, 0, len(arts))
	for _, a := range arts {
		title := a.Title
		if title == "" {
			title = "(untitled)"
		}
		items = append(items, articleItem{ID: a.ID, Title: title, Read: a.Read})
	}
	m.list.articles = items
	if len(items) == 0 {
		m.list.cursor = 0
	} else if m.list.cursor >= len(items) {
		m.list.cursor = len(items) - 1
	}
}

func (m *Model) renderList() string {
	head := "yerss"
	if m.list.filter != "" {
		head += " · filter: " + m.list.filter
	}
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(m.palette.Accent).Render(head))
	b.WriteString("\n\n")

	if len(m.list.articles) == 0 {
		dim := lipgloss.NewStyle().Foreground(m.palette.Dim)
		b.WriteString(dim.Render("No articles."))
		b.WriteString("\n")
		if m.list.filter != "" {
			b.WriteString(dim.Render("Filter \"" + m.list.filter + "\" has no articles. Press Esc to clear."))
		} else {
			b.WriteString(dim.Render("Add feeds to " + m.cfg.FeedsFile() + " and press R to refresh."))
		}
	} else {
		visible := m.height - 4
		if visible < 1 {
			visible = 1
		}
		start, end := listWindow(len(m.list.articles), m.list.cursor, visible)
		for i := start; i < end; i++ {
			item := m.list.articles[i]
			cursor := "  "
			if i == m.list.cursor {
				cursor = "> "
			}
			title := truncate(item.Title, m.width-4)
			var style lipgloss.Style
			if item.Read {
				style = lipgloss.NewStyle().Foreground(m.palette.Dim)
			} else {
				style = lipgloss.NewStyle().Bold(true).Foreground(m.palette.Bold)
			}
			if i == m.list.cursor {
				style = style.Background(lipgloss.Color("#333333"))
			}
			b.WriteString(style.Render(cursor + title))
			b.WriteString("\n")
		}
	}

	status := m.renderStatusBar()
	content := b.String()
	lines := strings.Split(content, "\n")
	h := m.height
	if h < 1 {
		h = 1
	}
	for len(lines) < h {
		lines = append(lines, "")
	}
	if len(lines) > h {
		lines = lines[:h]
	}
	if len(lines) > 0 {
		lines[len(lines)-1] = status
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderStatusBar() string {
	left := ""
	if m.statusMsg != "" && time.Now().Before(m.statusExpires) {
		left = m.statusMsg
	} else if m.refreshing {
		left = "refreshing..."
	} else {
		left = fmt.Sprintf("%d articles", len(m.list.articles))
	}
	if m.list.filter != "" && (m.statusMsg == "" || !time.Now().Before(m.statusExpires)) {
		left += " · filter " + m.list.filter
	}

	right := ""
	if m.lastRefreshedAt.IsZero() {
		right = "never refreshed"
	} else {
		right = "last refresh " + m.lastRefreshedAt.Format("2006-01-02 15:04")
	}

	pad := m.width - runewidth.StringWidth(left) - runewidth.StringWidth(right)
	if pad < 1 {
		pad = 1
	}
	line := left + strings.Repeat(" ", pad) + right
	line = truncate(line, m.width)
	return lipgloss.NewStyle().Foreground(m.palette.StatusBar).Render(line)
}

func (m *Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	act, ok := m.keys[config.ViewList][msg.String()]
	if !ok {
		return m, nil
	}
	var cmd tea.Cmd
	switch act {
	case config.Quit:
		return m, tea.Quit
	case config.MoveDown:
		m.moveListCursor(1)
	case config.MoveUp:
		m.moveListCursor(-1)
	case config.OpenArticle:
		m.openArticle()
	case config.Back:
		if m.list.filter != "" {
			m.clearFilter()
		}
	case config.Refresh:
		cmd = m.refreshManual()
	case config.PageDown:
		m.moveListCursor(m.pageSize())
	case config.PageUp:
		m.moveListCursor(-m.pageSize())
	case config.HalfPageDown:
		m.moveListCursor(m.pageSize() / 2)
	case config.HalfPageUp:
		m.moveListCursor(-m.pageSize() / 2)
	case config.Top:
		if len(m.list.articles) > 0 {
			m.list.cursor = 0
		}
	case config.Bottom:
		if len(m.list.articles) > 0 {
			m.list.cursor = len(m.list.articles) - 1
		}
	case config.TagPopup:
		m.openTagPopup(popupTagsList)
	case config.ToggleRead:
		m.toggleReadAtCursor()
	case config.MarkAllRead:
		m.markAllVisibleRead()
	case config.Help:
		m.popup = popupHelp
	}
	return m, cmd
}

func (m *Model) pageSize() int {
	n := m.height - 4
	if n < 1 {
		n = 1
	}
	return n
}

func (m *Model) moveListCursor(delta int) {
	n := len(m.list.articles)
	if n == 0 {
		return
	}
	m.list.cursor = (m.list.cursor + delta) % n
	if m.list.cursor < 0 {
		m.list.cursor += n
	}
}

func (m *Model) toggleReadAtCursor() {
	i := m.list.cursor
	if i >= len(m.list.articles) {
		return
	}
	item := &m.list.articles[i]
	item.Read = !item.Read
	_ = m.store.SetRead(item.ID, item.Read)
}

func (m *Model) markAllVisibleRead() {
	if len(m.list.articles) == 0 {
		return
	}
	ids := make([]int64, 0, len(m.list.articles))
	for i := range m.list.articles {
		ids = append(ids, m.list.articles[i].ID)
		m.list.articles[i].Read = true
	}
	_ = m.store.MarkAllRead(ids)
}

func (m *Model) clearFilter() {
	m.list.filter = ""
	m.loadList()
}

// listWindow returns the slice of the list to render around the cursor.
func listWindow(n, cursor, height int) (int, int) {
	if n <= height {
		return 0, n
	}
	start := cursor - height/2
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > n {
		end = n
		start = end - height
	}
	return start, end
}