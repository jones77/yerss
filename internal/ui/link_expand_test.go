package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
)

// openTestArticle inserts an article and opens it in the reader view.
func openTestArticle(t *testing.T, title, content string) (*Model, *store.Store) {
	t.Helper()
	m, st := newTestModel(t)
	a := store.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "guid-" + title,
		Title:   title,
		Content: content,
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id
	m.article = m.newArticleState(a)
	return m, st
}

func bodyText(m *Model) string {
	return ansi.Strip(strings.Join(m.article.lines, "\n"))
}

func TestLinkDisplayDefaultsToCompact(t *testing.T) {
	m, _ := openTestArticle(t, "compact", `<p>hello [world](https://example.com/post) end</p>`)
	body := bodyText(m)
	if !strings.Contains(body, "world") {
		t.Errorf("compact body missing the link name: %q", body)
	}
	if strings.Contains(body, "https://example.com/post") {
		t.Errorf("compact body shows the URL text: %q", body)
	}
}

func TestLinkExpandToggle(t *testing.T) {
	m, _ := openTestArticle(t, "toggle", strings.Join([]string{
		"<p>hello [world](https://example.com/post) end</p>",
		"<p>paragraph two</p>",
		"<p>paragraph three</p>",
		"<p>paragraph four</p>",
		"<p>paragraph five</p>",
		"<p>paragraph six</p>",
	}, ""))
	if strings.Contains(bodyText(m), "https://example.com/post") {
		t.Fatalf("expected compact by default: %q", bodyText(m))
	}

	// Scroll into the article to exercise offset preservation.
	m.article.viewport.SetYOffset(1)
	off := m.article.viewport.YOffset

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if !strings.Contains(bodyText(m), "world https://example.com/post") {
		t.Errorf("after x: want expanded link: %q", bodyText(m))
	}
	if got := m.article.viewport.YOffset; got != off {
		t.Errorf("x moved the scroll: %d -> %d", off, got)
	}

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if strings.Contains(bodyText(m), "https://example.com/post") {
		t.Errorf("second x should return to compact: %q", bodyText(m))
	}
}

func TestLinkExpandModePersistsAcrossArticles(t *testing.T) {
	m, st := openTestArticle(t, "first", `<p>hello [world](https://example.com/post) end</p>`)
	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if !strings.Contains(bodyText(m), "world https://example.com/post") {
		t.Fatalf("expected expanded after x: %q", bodyText(m))
	}
	a2 := store.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "guid-second",
		Title:   "second",
		Content: `<p>see [again](https://example.com/post) here</p>`,
	}
	id2, err := st.UpsertArticle(a2)
	if err != nil {
		t.Fatal(err)
	}
	a2.ID = id2
	m.article = m.newArticleState(a2)
	if !strings.Contains(bodyText(m), "again https://example.com/post") {
		t.Errorf("expanded mode did not persist across article switches: %q", bodyText(m))
	}
}

func TestCompactLinkSpanIsClickable(t *testing.T) {
	// Compact rendering drops the href span but keeps the OSC 8 name span, so
	// the left-click hit-testing reports exactly the name's URL and columns.
	m, _ := openTestArticle(t, "spans", `<p>hello [world](https://example.com/post) end</p>`)
	found := false
	for i, ln := range m.article.lines {
		if !strings.Contains(ansi.Strip(ln), "world") {
			continue
		}
		found = true
		spans := parseLinkSpans(ln)
		if len(spans) != 1 {
			t.Fatalf("compact line %d has %d link spans, want 1: %+v", i, len(spans), spans)
		}
		sp := spans[0]
		if sp.url != "https://example.com/post" {
			t.Errorf("compact line %d span url = %q", i, sp.url)
		}
		name := ansi.Strip(ln)
		start := strings.Index(name, "world")
		if start < 0 || sp.start != start || sp.end != start+len("world") {
			t.Errorf("compact line %d span (%d,%d) should cover the name at %d..%d: %q", i, sp.start, sp.end, start, start+len("world"), name)
		}
	}
	if !found {
		t.Fatal("article body missing the link line")
	}
}

func TestReaderHeaderURLRendersBlue(t *testing.T) {
	m, _ := newTestModel(t)
	m.article = m.newArticleState(store.Article{Title: "t", Link: "https://example.com/post", Content: "<p>x</p>"})
	blue := escapePrefix(lipgloss.NewStyle().Foreground(m.palette.StatusBar).Render("x"))
	if !strings.Contains(m.article.lines[0], blue) {
		t.Errorf("header URL line should render in the status-bar blue role: %q", m.article.lines[0])
	}
}
