package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"yerss/internal/config"
	"yerss/internal/store"
)

type popupState struct {
	tags   []store.TagCount
	links  []articleLink
	cursor int
}

func (m *Model) openTagPopup(mode popupMode) {
	tags, err := m.store.ListTags()
	if err != nil {
		m.setStatus("tags error: " + err.Error())
		return
	}
	if mode == popupTagsArticle {
		want := map[string]bool{}
		for _, c := range m.article.article.Categories {
			want[c] = true
		}
		filtered := make([]store.TagCount, 0, len(want))
		for _, t := range tags {
			if want[t.Name] {
				filtered = append(filtered, t)
			}
		}
		tags = filtered
	}
	m.popup = mode
	m.popupData = popupState{tags: tags}
}

func (m *Model) updatePopup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.popup == popupHelp {
		switch msg.String() {
		case "q", "esc", "enter", " ":
			m.popup = noPopup
		case "ctrl+c":
			return m, tea.Quit
		}
		return m, nil
	}
	if msg.String() == "enter" {
		if m.popup == popupLinks {
			return m.confirmLinkSelection()
		}
		return m.confirmTagSelection()
	}
	act, ok := m.keys[config.ViewPopup][msg.String()]
	if !ok {
		return m, nil
	}
	switch act {
	case config.Quit:
		m.persistSelection()
		return m, tea.Quit
	case config.Back:
		m.popup = noPopup
	case config.MoveDown:
		m.movePopupCursor(1)
	case config.MoveUp:
		m.movePopupCursor(-1)
	}
	return m, nil
}

// openLinksPopup opens the links popup over the article view, seeded with the
// links harvested from the current article's markdown source.
func (m *Model) openLinksPopup() {
	m.popup = popupLinks
	m.popupData = popupState{links: m.article.links}
}

// confirmLinkSelection closes the links popup and opens the selected URL in
// the system browser. An empty popup (an article without links) just closes.
func (m *Model) confirmLinkSelection() (tea.Model, tea.Cmd) {
	links := m.popupData.links
	m.popup = noPopup
	if m.popupData.cursor < 0 || m.popupData.cursor >= len(links) {
		return m, nil
	}
	return m, openURLCmd(links[m.popupData.cursor].url)
}

func (m *Model) confirmTagSelection() (tea.Model, tea.Cmd) {
	if len(m.popupData.tags) == 0 {
		m.popup = noPopup
		return m, nil
	}
	t := m.popupData.tags[m.popupData.cursor]
	m.popup = noPopup
	m.list.filter = t.Name
	m.view = viewList
	m.article.sel = textSelection{}
	m.loadList()
	return m, nil
}

func (m *Model) movePopupCursor(delta int) {
	n := len(m.popupData.tags)
	if m.popup == popupLinks {
		n = len(m.popupData.links)
	}
	if n == 0 {
		return
	}
	m.popupData.cursor = wrapIndex(m.popupData.cursor, delta, n)
}

func (m *Model) renderTagPopup() string {
	var h, w int
	if m.popup == popupTagsList {
		h, w = m.height*8/10, m.width*6/10
	} else {
		h, w = m.height*4/10, m.width*5/10
	}
	if h < 3 {
		h = 3
	}
	if w < 12 {
		w = 12
	}
	title := "Tags"
	if m.popup == popupTagsArticle {
		title = "Article tags"
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(m.palette.StatusBar).Render(title))
	for i, t := range m.popupData.tags {
		cursor := "  "
		if i == m.popupData.cursor {
			cursor = "> "
		}
		plain, bold := tagCountText(t.Unread, t.Total)
		boldPart := lipgloss.NewStyle().Bold(true).Foreground(m.palette.Bright).Render(bold)
		plainPart := lipgloss.NewStyle().Foreground(m.palette.Dim).Render(plain)
		lines = append(lines, cursor+t.Name+" "+boldPart+plainPart)
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.StatusBar).
		Width(w).
		Height(h).
		Padding(0, 1)
	return box.Render(strings.Join(lines, "\n"))
}

// renderLinksPopup renders the article links popup: each row shows the link
// text followed by the URL in the dim style, truncated to the box width.
func (m *Model) renderLinksPopup() string {
	h := m.height * 6 / 10
	w := m.width * 6 / 10
	if h < 3 {
		h = 3
	}
	if w < 12 {
		w = 12
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(m.palette.StatusBar).Render("Links"))
	dim := lipgloss.NewStyle().Foreground(m.palette.Dim)
	for i, l := range m.popupData.links {
		cursor := "  "
		if i == m.popupData.cursor {
			cursor = "> "
		}
		// text + separator + dim URL, truncated to the interior width.
		innerW := w - 6
		url := ansi.Truncate(l.url, max(1, innerW-ansi.StringWidth(l.text)-3), glyphsFor(m.ascii).ellipsis)
		line := l.text + dim.Render(" · "+url)
		lines = append(lines, cursor+ansi.Truncate(line, innerW, ""))
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.StatusBar).
		Width(w).
		Height(h).
		Padding(0, 1)
	return box.Render(strings.Join(lines, "\n"))
}

// tagCountText splits the tag count display into a plain and a bold portion.
// All-unread and all-read tags collapse to a single total; mixed tags show
// unread/total with the unread count bold.
func tagCountText(unread, total int) (plain, bold string) {
	if unread == 0 {
		return fmt.Sprintf("(%d)", total), ""
	}
	if total == unread {
		return "", fmt.Sprintf("(%d)", total)
	}
	return fmt.Sprintf("/%d)", total), fmt.Sprintf("(%d", unread)
}

// actionLabel renders an action identifier as a display label: underscores
// become spaces and each word is capitalized (e.g. open_article -> Open
// Article).
func actionLabel(a config.Action) string {
	return cases.Title(language.Und).String(strings.ReplaceAll(string(a), "_", " "))
}

func (m *Model) renderHelp() string {
	heading := lipgloss.NewStyle().Bold(true).Foreground(m.palette.StatusBar)

	// actionRow renders one binding as `label  keys`, indented two spaces.
	actionRow := func(a config.Action) string {
		keys := strings.Join(m.cfg.Keybindings[a], ", ")
		if keys == "" {
			return ""
		}
		return fmt.Sprintf("  %-11s %s", actionLabel(a), keys)
	}

	// section renders a heading followed by its action rows.
	section := func(label string, actions []config.Action) []string {
		lines := []string{heading.Render(label)}
		for _, a := range actions {
			if r := actionRow(a); r != "" {
				lines = append(lines, r)
			}
		}
		return lines
	}

	// Global fills the left column; List view and Article view stack in the
	// right column, separated by a blank line.
	left := section("Global", config.GlobalActions())
	right := section("List view", config.ListActions())
	right = append(right, "")
	right = append(right, section("Article view", config.ArticleActions())...)

	n := max(len(left), len(right))
	for len(left) < n {
		left = append(left, "")
	}
	for len(right) < n {
		right = append(right, "")
	}

	leftW := 0
	for _, l := range left {
		leftW = max(leftW, ansi.StringWidth(l))
	}
	rightW := 0
	for _, l := range right {
		rightW = max(rightW, ansi.StringWidth(l))
	}

	// The popup including its border must not exceed 70 columns.
	boxW := min(70, m.width, leftW+2+rightW+4)
	innerW := max(1, boxW-4)

	lines := []string{padRight("Help - current keymap", innerW)}
	for i := 0; i < n; i++ {
		line := padRight(left[i], leftW) + "  " + padRight(right[i], rightW)
		lines = append(lines, padRight(truncate(line, innerW), innerW))
	}

	h := len(lines) + 2
	if h < 6 {
		h = 6
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.StatusBar).
		Height(h).
		Padding(0, 1)
	return box.Render(strings.Join(lines, "\n"))
}
