package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
	"yerss/internal/ui/popup"
	"yerss/internal/ui/render"
)

func TestRenderList(t *testing.T) {
	m, st := newTestModel(t)
	insertArticle(t, st, "one", nil)
	insertArticle(t, st, "two", nil)
	m.list.Load()
	s := m.renderList()
	if !strings.Contains(s, "one") || !strings.Contains(s, "two") {
		t.Errorf("list render missing articles: %q", s)
	}
	if !strings.Contains(s, "0% "+render.GlyphsFor(m.ascii).Bullet+" 0/2") {
		t.Errorf("status bar missing position count: %q", s)
	}
}

func TestRenderArticleBorder(t *testing.T) {
	m, st := newTestModel(t)
	a := store.Article{
		FeedURL:     "https://example.com/feed.xml",
		GUID:        "g1",
		Title:       "Hello world article title",
		Link:        "https://example.com/1",
		PublishedAt: mustParseTime(t, "2026-01-02T15:04:05Z"),
		Content:     "<p>a short paragraph</p>",
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id
	m.SetLinkExpand(true)
	m.article = m.article.Recompose(a)

	s := m.article.Render()
	lines := strings.Split(s, "\n")
	if len(lines) != m.height {
		t.Fatalf("expected %d lines, got %d", m.height, len(lines))
	}
	if !strings.Contains(lines[0], "Hello world article title") {
		t.Errorf("top border missing title: %q", lines[0])
	}
	wantDate := mustParseTime(t, "2026-01-02T15:04:05Z").Local().Format("2006-01-02 15:04")
	if !strings.Contains(lines[0], wantDate) {
		t.Errorf("top border missing date+time %q: %q", wantDate, lines[0])
	}
	if !strings.Contains(lines[len(lines)-1], "o: open article in browser") {
		t.Errorf("bottom border missing help hint: %q", lines[len(lines)-1])
	}
	if !strings.Contains(lines[len(lines)-1], "100% · 5/5") {
		t.Errorf("bottom border missing line indicator: %q", lines[len(lines)-1])
	}
}

func TestRenderArticleBorderShowSeconds(t *testing.T) {
	m, st := newTestModel(t)
	m.sess.Config().Display.ShowSeconds = true
	a := store.Article{
		FeedURL:     "https://example.com/feed.xml",
		GUID:        "g-seconds",
		Title:       "Seconds title",
		PublishedAt: mustParseTime(t, "2026-01-02T15:04:05Z"),
		Content:     "<p>a short paragraph</p>",
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id
	m.article = m.article.Recompose(a)

	s := m.article.Render()
	lines := strings.Split(s, "\n")
	wantDate := mustParseTime(t, "2026-01-02T15:04:05Z").Local().Format("2006-01-02 15:04:05")
	if !strings.Contains(lines[0], wantDate) {
		t.Errorf("top border missing seconds date+time %q: %q", wantDate, lines[0])
	}
}

func TestRenderArticleBorderASCII(t *testing.T) {
	m, _ := newTestModel(t)
	m.SetAscii(true)
	a := store.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "g2",
		Title:   "ascii",
		Content: "<p>x</p>",
	}
	m.article = m.article.Recompose(a)
	s := m.article.Render()
	lines := strings.Split(s, "\n")
	if !strings.Contains(lines[0], "+") {
		t.Errorf("ascii top border missing +: %q", lines[0])
	}
}

func TestRenderArticleBorderInteriorPadding(t *testing.T) {
	m, _ := newTestModel(t)
	m.SetAscii(true)
	m.article = m.article.Recompose(store.Article{
		Title:   "Hello world article title",
		Content: "<p>a short paragraph</p>",
	})
	lines := strings.Split(m.article.Render(), "\n")
	first := ansi.Strip(lines[2])
	// padX contributes two leading spaces; glamour's document margin is removed
	// so it does not stack on top.
	if !strings.HasPrefix(first, ":  Hello world article title") {
		t.Errorf("content missing two-space left padding: %q", first)
	}
	if got := first[len(first)-1]; got != ':' {
		t.Errorf("right edge not a border line at column %d: got %q", len(first)-1, first)
	}
	if len(first) != 80 {
		t.Errorf("content row width = %d, want 80: %q", len(first), first)
	}
	if !strings.HasPrefix(ansi.Strip(lines[0]), "+") {
		t.Errorf("top border has leading margin: %q", lines[0])
	}
}

// trackGlyphs returns the right-edge column of every interior row of a rendered
// article frame, with ANSI escapes stripped. In ASCII fallback mode the thumb
// rows read '|' and the plain border track rows read ':'.
func trackGlyphs(t *testing.T, s string) string {
	t.Helper()
	lines := strings.Split(s, "\n")
	var b strings.Builder
	for _, l := range lines[1 : len(lines)-1] {
		st := ansi.Strip(l)
		b.WriteByte(st[len(st)-1])
	}
	return b.String()
}

func TestRenderArticleThumbPosition(t *testing.T) {
	m, _ := newTestModel(t)
	m.SetAscii(true)
	m.article = m.article.Recompose(store.Article{Title: "long", Content: "<p>x</p>"})
	// 70 content lines vs viewport height 20 (height 24, padY 1 above and
	// below): scrollable range 50, track height 22, thumb height 22*20/70 = 6.
	m.article.Viewport.SetContent(strings.Join(make([]string, 70), "\n"))

	cases := []struct {
		name      string
		offset    int
		wantTrack string
		wantLabel string
	}{
		{"on open", 0, "||||||::::::::::::::::", "0% . 20/70"},
		{"42%", 21, "::::::||||||::::::::::", "42% . 41/70"},
		{"100%", 50, "::::::::::::::::||||||", "100% . 70/70"},
	}
	for _, c := range cases {
		m.article.Viewport.SetYOffset(c.offset)
		s := m.article.Render()
		lines := strings.Split(s, "\n")
		if got := trackGlyphs(t, s); got != c.wantTrack {
			t.Errorf("%s: track glyphs = %q, want %q", c.name, got, c.wantTrack)
		}
		if !strings.Contains(lines[len(lines)-1], c.wantLabel) {
			t.Errorf("%s: bottom border missing %q: %q", c.name, c.wantLabel, lines[len(lines)-1])
		}
	}
}

func TestRenderArticleShortArticleNoThumb(t *testing.T) {
	m, _ := newTestModel(t)
	m.SetAscii(true)
	m.article = m.article.Recompose(store.Article{Title: "short", Content: "<p>x</p>"})
	s := m.article.Render()
	lines := strings.Split(s, "\n")
	if got := trackGlyphs(t, s); got != strings.Repeat(":", 22) {
		t.Errorf("fitting article track = %q, want all %q", got, strings.Repeat(":", 22))
	}
	if !strings.Contains(lines[len(lines)-1], "100% . 3/3") {
		t.Errorf("bottom border missing 100%%: %q", lines[len(lines)-1])
	}
}

func TestRenderTagPopup(t *testing.T) {
	m, _ := newTestModel(t)
	m.popupData.SetKind(popup.KindTagsList)
	m.popupData.SetTags([]store.TagCount{{Name: "tech", Total: 3, Unread: 1}})
	s := m.popupData.Render()
	if !strings.Contains(s, "tech") {
		t.Errorf("popup missing tag: %q", s)
	}
}

func mustParseTime(t *testing.T, v string) time.Time {
	t.Helper()
	tm, err := time.Parse(time.RFC3339, v)
	if err != nil {
		t.Fatal(err)
	}
	return tm
}
