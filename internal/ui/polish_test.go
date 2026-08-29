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
	if !strings.HasSuffix(line, "… ─╖") {
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
		{"", "https://nytimes.com/rss", "nytimes"},
		{"", "", ""},
	}
	for _, c := range cases {
		if got := sourceID(c.link, c.feed); got != c.want {
			t.Errorf("sourceID(%q, %q) = %q, want %q", c.link, c.feed, got, c.want)
		}
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

	dimEscape := escapePrefix(lipgloss.NewStyle().Foreground(m.palette.Dim).Render("x"))
	boldEscape := escapePrefix(lipgloss.NewStyle().Bold(true).Foreground(m.palette.Bold).Render("x"))

	unread := m.renderArticleRow(item, false)
	if !strings.Contains(unread, boldEscape) {
		t.Errorf("unread title should be bold: %q", unread)
	}
	if !strings.Contains(unread, dimEscape+"newrepublic") || !strings.Contains(unread, dimEscape+"15:04") {
		t.Errorf("source and time should be dim: %q", unread)
	}

	item.Read = true
	read := m.renderArticleRow(item, false)
	if strings.Contains(read, boldEscape) {
		t.Errorf("read title should not be bold: %q", read)
	}
	if !strings.Contains(read, dimEscape) {
		t.Errorf("read title should be dim: %q", read)
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

	borderColon := lipgloss.NewStyle().Foreground(m.palette.Border).Render(":")
	brightPipe := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffffff")).Render("|")

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
	line := stripTop(t, bottomBorder(40, g, darkPalette(), scrollState{totalH: 120, viewportH: 20, offset: 100}))
	if !strings.HasPrefix(line, "└o: open in browser") {
		t.Errorf("help hint should be left-aligned: %q", line)
	}
	if !strings.HasSuffix(line, "100% · 120/120╜") {
		t.Errorf("indicator should be 100%% · 120/120 at the bottom: %q", line)
	}

	fits := stripTop(t, bottomBorder(40, g, darkPalette(), scrollState{totalH: 4, viewportH: 20, offset: 0}))
	if !strings.Contains(fits, "100% · 4/4") {
		t.Errorf("short article indicator should be 100%% · 4/4: %q", fits)
	}
}

func TestWrapTextBreaksURLToNewLine(t *testing.T) {
	url := "example.com/a"
	link := ansi.SetHyperlink(url) + "[text]" + ansi.ResetHyperlink() + "(" + url + ")"
	out := wrapText("abc "+link, 16, "…")
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), out)
	}
	if ansi.Strip(lines[0]) != "abc" || ansi.Strip(lines[1]) != "[text]" || ansi.Strip(lines[2]) != "(example.com/a)" {
		t.Errorf("URL should break onto its own line, got %q", out)
	}
}

func TestWrapTextTruncatesOverlongURL(t *testing.T) {
	url := "https://very-long-example.com/a-really-long-path"
	link := ansi.SetHyperlink(url) + "[text]" + ansi.ResetHyperlink() + "(" + url + ")"
	out := wrapText(link, 10, "…")
	lines := strings.Split(out, "\n")
	for i, l := range lines {
		if ansi.StringWidth(l) > 10 {
			t.Errorf("line %d exceeds width 10: %q", i, l)
		}
	}
	last := ansi.Strip(lines[len(lines)-1])
	if !strings.HasPrefix(last, "(") || !strings.HasSuffix(last, "…)") {
		t.Errorf("overlong URL should be truncated with an ellipsis: %q", last)
	}
}

func TestActionLabel(t *testing.T) {
	cases := map[config.Action]string{
		config.OpenArticle:  "Open Article",
		config.MarkAllRead:  "Mark All Read",
		config.ToggleRead:   "Toggle Read",
		config.HalfPageDown: "Half Page Down",
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