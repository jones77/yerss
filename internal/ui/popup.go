package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"yerss/internal/config"
	"yerss/internal/store"
)

type popupState struct {
	tags   []store.TagCount
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
		return m.confirmTagSelection()
	}
	act, ok := m.keys[config.ViewPopup][msg.String()]
	if !ok {
		return m, nil
	}
	switch act {
	case config.Quit:
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

func (m *Model) confirmTagSelection() (tea.Model, tea.Cmd) {
	if len(m.popupData.tags) == 0 {
		m.popup = noPopup
		return m, nil
	}
	t := m.popupData.tags[m.popupData.cursor]
	m.popup = noPopup
	m.list.filter = t.Name
	m.view = viewList
	m.loadList()
	return m, nil
}

func (m *Model) movePopupCursor(delta int) {
	n := len(m.popupData.tags)
	if n == 0 {
		return
	}
	m.popupData.cursor = (m.popupData.cursor + delta) % n
	if m.popupData.cursor < 0 {
		m.popupData.cursor += n
	}
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
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(m.palette.Accent).Render(title))
	for i, t := range m.popupData.tags {
		cursor := "  "
		if i == m.popupData.cursor {
			cursor = "> "
		}
		plain, bold := tagCountText(t.Unread, t.Total)
		boldPart := lipgloss.NewStyle().Bold(true).Foreground(m.palette.Bold).Render(bold)
		plainPart := lipgloss.NewStyle().Foreground(m.palette.Dim).Render(plain)
		lines = append(lines, cursor+t.Name+" "+boldPart+plainPart)
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Accent).
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
	var lines []string
	lines = append(lines, "Help - current keymap")
	lines = append(lines, "")
	for _, a := range []config.Action{
		config.Quit, config.Refresh, config.OpenArticle, config.Back,
		config.MoveUp, config.MoveDown, config.PageUp, config.PageDown,
		config.HalfPageUp, config.HalfPageDown, config.Top, config.Bottom,
		config.TagPopup, config.ToggleRead, config.MarkAllRead,
		config.OpenURL, config.CopyURL, config.Help,
	} {
		keys := strings.Join(m.cfg.Keybindings[a], ", ")
		if keys == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("  %-18s %s", actionLabel(a), keys))
	}
	w := m.width * 3 / 4
	h := m.height * 2 / 3
	if w < 40 {
		w = 40
	}
	if h < 10 {
		h = 10
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Accent).
		Width(w).
		Height(h).
		Padding(0, 1)
	return box.Render(strings.Join(lines, "\n"))
}