package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"yerss/internal/config"
	"yerss/internal/store"
)

// escapePrefix returns the ANSI escape prefix of a lipgloss-rendered string
// (everything before the first printable rune).
func escapePrefix(rendered string) string {
	return strings.Split(rendered, "x")[0]
}

// forceTrueColor makes lipgloss emit TrueColor escape codes so color-rule
// tests can assert on the rendered ANSI, restoring the prior profile on
// cleanup.
func forceTrueColor(t *testing.T) {
	t.Helper()
	orig := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(orig) })
}

func TestTopBorderBulletGlyph(t *testing.T) {
	u := stripTop(t, topBorder(40, glyphsFor(false), darkPalette(), "2026-01-02 15:04:05", "short"))
	if !strings.Contains(u, "·") {
		t.Errorf("unicode top border missing · bullet: %q", u)
	}
	a := stripTop(t, topBorder(40, glyphsFor(true), darkPalette(), "2026-01-02 15:04:05", "short"))
	if !strings.Contains(a, "2026-01-02 15:04:05 . ") {
		t.Errorf("ascii top border missing . bullet: %q", a)
	}
	if strings.Contains(a, "·") {
		t.Errorf("ascii top border should not contain ·: %q", a)
	}
}

func TestTopBorderShowsTime(t *testing.T) {
	g := glyphsFor(false)
	line := stripTop(t, topBorder(40, g, darkPalette(), "2026-01-02 15:04:05", strings.Repeat("x", 100)))
	if !strings.Contains(line, "2026-01-02 15:04:05") {
		t.Errorf("date+time prefix must stay intact: %q", line)
	}
	if !strings.HasSuffix(line, "… ─┐") {
		t.Errorf("expected ellipsis and single dash before the right corner: %q", line)
	}
}

func TestSourceID(t *testing.T) {
	cases := []struct {
		link, feed, want string
	}{
		{"https://newrepublic.com/story/1", "", "newrepublic"},
		{"https://www.nytimes.com/x", "", "nytimes"},
		{"https://sueddeutsche.de/x", "", "sueddeutsche"},
		{"https://reallylongnewspaperdomainname.net/x", "", "reallylongne"},
		{"https://tribunemag.com./story/1", "", "tribunemag"},
		{"https://WWW.TRIBUNEMAG.COM/x", "", "tribunemag"},
		{"https://tribunemag.co.uk/story/1", "", "tribunemag"},
		{"https://www.bbc.co.uk/news/x", "", "bbc"},
		{"https://feeds.example.com/x", "", "example"},
		{"", "https://nytimes.com/rss", "nytimes"},
		{"", "", ""},
	}
	for _, c := range cases {
		if got := sourceID(c.link, c.feed); got != c.want {
			t.Errorf("sourceID(%q, %q) = %q, want %q", c.link, c.feed, got, c.want)
		}
	}
}

func TestArticleRowSourceRightAligned(t *testing.T) {
	withLocalZone(t, time.UTC)
	m, st := newTestModel(t)
	insertTimedArticle(t, st, "timed", time.Date(2026, 8, 28, 15, 4, 5, 0, time.UTC))
	m.loadList()
	item := &m.list.groups[0].articles[0]
	item.Link = "https://newrepublic.com/story/1"

	line := ansi.Strip(m.renderArticleRow(item, false))
	if !strings.HasSuffix(line, "newrepublic") {
		t.Errorf("source should be the only right-aligned field: %q", line)
	}
	if strings.Contains(line, glyphsFor(m.ascii).bullet) {
		t.Errorf("row should not contain bullets: %q", line)
	}
	if ansi.StringWidth(line) != m.width {
		t.Errorf("row width = %d, want %d: %q", ansi.StringWidth(line), m.width, line)
	}
}

func TestArticleRowTitleTruncatedWithEllipsis(t *testing.T) {
	withLocalZone(t, time.UTC)
	m, st := newTestModel(t)
	insertTimedArticle(t, st, strings.Repeat("x", 200), time.Date(2026, 8, 28, 15, 4, 5, 0, time.UTC))
	m.loadList()
	item := &m.list.groups[0].articles[0]
	item.Link = "https://newrepublic.com/story/1"

	g := glyphsFor(m.ascii)
	line := ansi.Strip(m.renderArticleRow(item, false))
	if !strings.HasSuffix(line, g.ellipsis+" newrepublic") {
		t.Errorf("truncated title should end with an ellipsis and keep a space before the source: %q", line)
	}
	if ansi.StringWidth(line) != m.width {
		t.Errorf("row width = %d, want %d: %q", ansi.StringWidth(line), m.width, line)
	}
}

func TestArticleRowTitleColorChangesWithRead(t *testing.T) {
	forceTrueColor(t)
	withLocalZone(t, time.UTC)
	m, st := newTestModel(t)
	insertTimedArticle(t, st, "timed", time.Date(2026, 8, 28, 15, 4, 5, 0, time.UTC))
	m.loadList()
	item := &m.list.groups[0].articles[0]
	item.Link = "https://newrepublic.com/story/1"

	textEscape := escapePrefix(lipgloss.NewStyle().Foreground(m.palette.Text).Render("x"))
	boldEscape := escapePrefix(lipgloss.NewStyle().Bold(true).Foreground(m.palette.Bright).Render("x"))

	unread := m.renderArticleRow(item, false)
	if !strings.Contains(unread, boldEscape) {
		t.Errorf("unread title should be bold: %q", unread)
	}
	if !strings.Contains(unread, textEscape+"newrepublic") || !strings.Contains(unread, textEscape+"15:04") {
		t.Errorf("source and time should be white: %q", unread)
	}

	item.Read = true
	read := m.renderArticleRow(item, false)
	if strings.Contains(read, boldEscape) {
		t.Errorf("read title should not be bold: %q", read)
	}
	if !strings.Contains(read, textEscape) {
		t.Errorf("read title should be white: %q", read)
	}
}

func TestArticleRowSelectionBackground(t *testing.T) {
	forceTrueColor(t)
	withLocalZone(t, time.UTC)
	m, st := newTestModel(t)
	insertTimedArticle(t, st, "timed", time.Date(2026, 8, 28, 15, 4, 5, 0, time.UTC))
	m.loadList()
	item := &m.list.groups[0].articles[0]

	bgEscape := escapePrefix(lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Render("x"))
	// The selected rail renders its grey foreground and the selection
	// background in a single escape; expect that combined prefix.
	selEscape := escapePrefix(lipgloss.NewStyle().Foreground(m.palette.Text).Background(lipgloss.Color("#333333")).Render("x"))
	unselected := m.renderArticleRow(item, false)
	selected := m.renderArticleRow(item, true)
	if strings.Contains(unselected, bgEscape) {
		t.Errorf("unselected row should have no background highlight: %q", unselected)
	}
	if !strings.Contains(selected, bgEscape) {
		t.Errorf("selected row should use the full-row background highlight: %q", selected)
	}
	// The highlight must start at the row's first column (the tree glyph).
	if !strings.HasPrefix(selected, selEscape) {
		t.Errorf("selection background should cover the rail column: %q", selected)
	}
}

func TestArticleFrameBorderColors(t *testing.T) {
	forceTrueColor(t)
	m, _ := newTestModel(t)
	m.ascii = true
	m.article = m.newArticleState(store.Article{Title: "long", Content: "<p>x</p>"})
	m.article.viewport.SetContent(strings.Join(make([]string, 70), "\n"))
	s := m.renderArticle()
	lines := strings.Split(s, "\n")

	borderColon := lipgloss.NewStyle().Foreground(m.palette.Dim).Render(":")
	brightPipe := lipgloss.NewStyle().Bold(true).Foreground(m.palette.StatusBar).Render("|")

	if !strings.HasPrefix(lines[2], borderColon) {
		t.Errorf("left border should be a grey :, got %q", lines[2][:20])
	}
	if !strings.HasSuffix(lines[2], brightPipe) {
		t.Errorf("thumb row should end with a bright |: %q", lines[2])
	}
	if !strings.HasSuffix(lines[8], borderColon) {
		t.Errorf("track row should end with a grey :, got %q", lines[8])
	}
}

func TestBottomBorderHelpHintAndIndicator(t *testing.T) {
	g := glyphsFor(false)
	line := stripTop(t, bottomBorder(80, g, darkPalette(), scrollState{totalH: 120, viewportH: 20, offset: 100}))
	if !strings.HasPrefix(line, "└─ o: open in browser") {
		t.Errorf("hint should be inset past a horizontal line: %q", line)
	}
	if !strings.HasSuffix(line, "100% · 120/120 ─┘") {
		t.Errorf("indicator should be inset past a horizontal line: %q", line)
	}
	if !strings.Contains(line, "browser ─") {
		t.Errorf("space between the hint and the border fill: %q", line)
	}
	if !strings.Contains(line, "─ 100%") {
		t.Errorf("space between the border fill and the percentage: %q", line)
	}
	if ansi.StringWidth(line) != 80 {
		t.Errorf("bottom border width = %d, want 80: %q", ansi.StringWidth(line), line)
	}

	fits := stripTop(t, bottomBorder(80, g, darkPalette(), scrollState{totalH: 4, viewportH: 20, offset: 0}))
	if !strings.Contains(fits, "100% · 4/4") {
		t.Errorf("short article indicator should be 100%% · 4/4: %q", fits)
	}
}

func TestActionLabel(t *testing.T) {
	cases := map[config.Action]string{
		config.OpenArticle:     "Open Article",
		config.HalfPageDown:    "Half Page Down",
		config.CopyArticleText: "Copy Article Text",
	}
	for a, want := range cases {
		if got := actionLabel(a); got != want {
			t.Errorf("actionLabel(%s) = %q, want %q", a, got, want)
		}
	}
}

func TestLowercaseTOpensTagPopup(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", []string{"tech"})
	m.loadList()
	m.updateList(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	if m.popup != popupTagsList {
		t.Errorf("t should open the tag popup, popup = %d", m.popup)
	}
}
