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
	case config.OpenURL:
		if (m.popup == popupLinks || m.popup == popupTagsArticle) && m.article.article.Link != "" {
			return m, openURLCmd(m.article.article.Link)
		}
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
	w := m.width * 8 / 10
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
		url := ansi.Truncate(l.url, max(1, innerW-ansi.StringWidth(l.text)-3), m.glyphs().ellipsis)
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

// helpLabel renders an action label for the help popup, replacing the "Half"
// prefix of the half-page actions with the half glyph (½, or 1/2 in ASCII
// mode).
func (m *Model) helpLabel(a config.Action) string {
	label := actionLabel(a)
	if strings.HasPrefix(label, "Half") {
		return m.glyphs().half + strings.TrimPrefix(label, "Half")
	}
	return label
}

// displayKeys renders normalized key strings for the help popup: a literal
// space key (as stored by normalizeKey for runtime matching) is shown as the
// word "space", and the normalized "pgdown" spelling is shown as "pgdn".
func displayKeys(keys []string) []string {
	out := make([]string, len(keys))
	for i, k := range keys {
		switch k {
		case " ":
			out[i] = "space"
		case "pgdown":
			out[i] = "pgdn"
		default:
			out[i] = k
		}
	}
	return out
}

func (m *Model) renderHelp() string {
	heading := lipgloss.NewStyle().Foreground(m.palette.StatusBar)
	keyStyle := lipgloss.NewStyle().Foreground(m.palette.Bright)

	// actionRow renders one binding with its label padded to labelW+2 columns
	// so the keys align across the section, leaving at least two spaces between
	// the label and its bound keys. Each bound key renders in the bright role
	// so it stands out; the commas between keys stay in the default text role.
	actionRow := func(a config.Action, labelW int) string {
		keys := displayKeys(m.cfg.Keybindings[a])
		if len(keys) == 0 {
			return ""
		}
		styled := make([]string, len(keys))
		for i, k := range keys {
			styled[i] = keyStyle.Render(k)
		}
		return fmt.Sprintf("%-*s%s", labelW+2, m.helpLabel(a), strings.Join(styled, ", "))
	}

	// section renders a heading followed by its action rows, aligning the keys
	// under the widest label in the section.
	section := func(label string, actions []config.Action) []string {
		labelW := 0
		for _, a := range actions {
			labelW = max(labelW, ansi.StringWidth(m.helpLabel(a)))
		}
		lines := []string{heading.Render(label)}
		for _, a := range actions {
			if r := actionRow(a, labelW); r != "" {
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

	// The popup including its border must not exceed 70 columns. The box keeps
	// one space of padding on each side.
	gap := 2
	boxW := min(70, m.width, leftW+gap+rightW+4)
	textW := max(1, boxW-4)

	g := m.glyphs()
	style := lipgloss.NewStyle().Foreground(m.palette.StatusBar)
	v := style.Render(g.v)
	pad := " "

	var content []string
	for i := 0; i < n; i++ {
		line := padRight(left[i], leftW) + strings.Repeat(" ", gap) + padRight(right[i], rightW)
		content = append(content, padRight(truncate(line, textW), textW))
	}

	// The top edge inlines the "Key Bindings" title; one blank row pads the
	// top and bottom of the content.
	blank := v + pad + strings.Repeat(" ", textW) + pad + v
	rows := []string{helpTopBorder(boxW, g, m.palette), blank}
	for _, l := range content {
		rows = append(rows, v+pad+l+pad+v)
	}
	rows = append(rows, blank)
	rows = append(rows, style.Render(g.bl+strings.Repeat(g.h, boxW-2)+g.br))
	return strings.Join(rows, "\n")
}

// helpTopBorder renders the help popup's top edge as `┌─ Key Bindings ───┐`,
// centering the title in the bold chrome role and keeping the rails in the
// chrome role.
func helpTopBorder(w int, g borderGlyphs, p Palette) string {
	title := "Key Bindings"
	style := lipgloss.NewStyle().Foreground(p.StatusBar)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(p.StatusBar)
	leftRail := g.tl + g.h
	rightRail := g.h + g.tr
	core := " " + title + " "
	fill := max(0, w-ansi.StringWidth(leftRail)-ansi.StringWidth(core)-ansi.StringWidth(rightRail))
	left := fill / 2
	right := fill - left
	return style.Render(leftRail+strings.Repeat(g.h, left)) +
		titleStyle.Render(core) +
		style.Render(strings.Repeat(g.h, right)+rightRail)
}
