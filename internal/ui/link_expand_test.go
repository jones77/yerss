package ui

import (
	"fmt"
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
	m.linkExpand = true
	m.article = m.newArticleState(store.Article{Title: "t", Link: "https://example.com/post", Content: "<p>x</p>"})
	blue := escapePrefix(lipgloss.NewStyle().Foreground(m.palette.StatusBar).Render("x"))
	if !strings.Contains(m.article.lines[0], blue) {
		t.Errorf("header URL line should render in the status-bar blue role: %q", m.article.lines[0])
	}
	if first := ansi.Strip(m.article.lines[0]); !strings.Contains(first, "example.com/post") {
		t.Errorf("first content line should be the URL in expanded mode: %q", first)
	}
}

// wrapURL is long enough to wrap inside the test model's content width (74
// columns) but short enough that its last rendered chunk leaves room for
// trailing paragraph text.
const wrapURL = "https://example.com/very/long/path/that/exceeds/the/content/width/and/wraps/again/more/still"

// paragraphText returns the article body stripped of ANSI, cut at the first
// occurrence of "before" so the header (title) lines are left out.
func paragraphText(m *Model) string {
	body := bodyText(m)
	if i := strings.Index(body, "before"); i >= 0 {
		body = body[i:]
	}
	return body
}

func TestCompactWrapHidesWrappedURL(t *testing.T) {
	// The URL spans multiple rendered lines in expanded mode; in compact mode
	// no part of it may appear, not even a wrapped continuation line.
	m, _ := openTestArticle(t, "longhref", `<p>before [link](`+wrapURL+`) after</p>`)
	body := paragraphText(m)
	if got := strings.Join(strings.Fields(body), " "); got != "before link after" {
		t.Errorf("compact body = %q, want %q", got, "before link after")
	}
	if strings.Contains(body, "example.com") {
		t.Errorf("compact body leaks the wrapped URL: %q", body)
	}

	// Non-vacuous guard: the URL genuinely wraps at this width in expanded
	// mode, so the trailing "after" lands on a later line than "before".
	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	bi, ai := -1, -1
	for i, ln := range m.article.lines {
		s := ansi.Strip(ln)
		if bi < 0 && strings.Contains(s, "before") {
			bi = i
		}
		if strings.Contains(s, "after") {
			ai = i
		}
	}
	if ai <= bi {
		t.Fatalf("URL did not wrap ('after' on line %d, 'before' on line %d); test is vacuous", ai, bi)
	}

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if body := paragraphText(m); strings.Contains(body, "example.com") {
		t.Errorf("compact body leaks the wrapped URL after toggling back: %q", body)
	}
}

func TestCompactWrapKeepsWrappedName(t *testing.T) {
	// A name longer than the content width wraps across lines; the full name
	// must render and the URL must stay hidden.
	name := "Richard Jebreal and the Democratic Party of Nothing is quite long and keeps going"
	m, _ := openTestArticle(t, "longname", `<p>before [`+name+`](`+wrapURL+`) after</p>`)
	body := paragraphText(m)
	got := strings.Join(strings.Fields(body), " ")
	want := strings.Join(strings.Fields("before "+name+" after"), " ")
	if got != want {
		t.Errorf("compact body = %q, want %q", got, want)
	}
	if strings.Contains(got, "example.com") {
		t.Errorf("compact body leaks the URL: %q", got)
	}
	if n := nonEmptyLines(body); n < 2 {
		t.Fatalf("name did not wrap (%d non-empty lines); test is vacuous", n)
	}
}

func TestCompactWrapAdjacentSameURL(t *testing.T) {
	// Two links to the same URL must each keep their name and hide their href,
	// even when the shared URL wraps across rendered lines.
	m, _ := openTestArticle(t, "adjacent", `<p>before [one](`+wrapURL+`) and [two](`+wrapURL+`) after</p>`)
	body := paragraphText(m)
	if got := strings.Join(strings.Fields(body), " "); got != "before one and two after" {
		t.Errorf("compact body = %q, want %q", got, "before one and two after")
	}
	if strings.Contains(body, "example.com") {
		t.Errorf("compact body leaks the URL: %q", body)
	}
	if n := nonEmptyLines(body); n < 3 {
		t.Fatalf("shared URL did not wrap (+ adjacent links; %d non-empty lines); test is vacuous", n)
	}
}

func nonEmptyLines(s string) int {
	n := 0
	for _, ln := range strings.Split(s, "\n") {
		if strings.TrimSpace(ln) != "" {
			n++
		}
	}
	return n
}

func TestCompactHeaderHidesURL(t *testing.T) {
	m, _ := newTestModel(t)
	m.article = m.newArticleState(store.Article{Title: "t", Link: "https://example.com/post", Content: "<p>x</p>"})
	lines := strippedLines(m.article.lines)
	if first := strings.TrimRight(lines[0], " "); first != "t" {
		t.Errorf("first content line = %q, want the title", first)
	}
	if joined := strings.Join(lines, "\n"); strings.Contains(joined, "example.com/post") {
		t.Errorf("compact header should not render the URL line: %q", joined)
	}
}

// anchorToggleFixture builds an article whose paragraphs each start with a
// stable "paragraph <n> text anchor<n>" line and end in a wrapped hyperlink, so
// toggling link mode adds URL lines that push the text off its compact position
// without changing each paragraph's leading words.
func anchorToggleFixture(t *testing.T, n int) (*Model, *store.Store) {
	t.Helper()
	var html strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&html, `<p>paragraph %d text anchor%d [j](%s)</p>`, i, i, wrapURL)
	}
	return openTestArticle(t, "anchor-toggle", html.String())
}

func TestLinkExpandToggleKeepsFirstVisibleLine(t *testing.T) {
	m, _ := anchorToggleFixture(t, 30)
	if strings.Contains(bodyText(m), "example.com") {
		t.Fatalf("expected compact by default: %q", bodyText(m))
	}
	oldLen := len(m.article.lines)
	// Scroll partway down so the toggle genuinely shifts line indices.
	m.article.viewport.SetYOffset(2)
	oldOffset := m.article.viewport.YOffset
	anchor := strings.TrimSpace(ansi.Strip(m.article.lines[oldOffset]))
	if anchor == "" {
		t.Fatalf("anchor line at offset %d is empty", oldOffset)
	}

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	// Non-vacuous guard: expanded mode really added URL lines.
	if newLen := len(m.article.lines); newLen <= oldLen {
		t.Fatalf("link toggle did not grow the article (%d -> %d lines); test is vacuous", oldLen, newLen)
	}
	// Total-delta bump is exactly what the toggle must NOT do.
	inserted := len(m.article.lines) - oldLen
	if got := m.article.viewport.YOffset; got == oldOffset+inserted {
		t.Errorf("toggle jumped to the total-delta offset %d; the first visible line drifted", got)
	}
	// The anchored paragraph is the first visible line again, with its leading
	// words intact (a link grew on the line, so trailing text can move).
	first := strings.TrimSpace(ansi.Strip(m.article.lines[m.article.viewport.YOffset]))
	if !strings.HasPrefix(first, anchor) {
		t.Errorf("first visible line after toggle = %q, want it to start with the anchor %q", first, anchor)
	}
}

func TestLinkExpandToggleAtBottomStaysPinned(t *testing.T) {
	m, _ := anchorToggleFixture(t, 40)
	m.article.viewport.GotoBottom()
	preBottom := m.article.viewport.YOffset
	if preBottom == 0 {
		t.Fatalf("article does not scroll; test is vacuous")
	}

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	got := m.article.viewport.YOffset
	maxOffset := len(m.article.lines) - m.article.viewport.Height
	if got != maxOffset {
		t.Errorf("after x at the bottom: offset = %d, want the new bottom %d (was %d)", got, maxOffset, preBottom)
	}
}

func TestLinkExpandToggleBackReturnsToAnchor(t *testing.T) {
	m, _ := anchorToggleFixture(t, 30)
	m.article.viewport.SetYOffset(2)
	anchor := strings.TrimSpace(ansi.Strip(m.article.lines[m.article.viewport.YOffset]))

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if !strings.Contains(bodyText(m), "example.com") {
		t.Fatalf("expected expanded after x: %q", bodyText(m))
	}
	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if got := strings.TrimSpace(ansi.Strip(m.article.lines[m.article.viewport.YOffset])); got != anchor {
		t.Errorf("after toggling back: first visible line = %q, want the original anchor %q", got, anchor)
	}
}

func TestFindAnchorLine(t *testing.T) {
	lines := []string{
		"title",
		"",
		"paragraph 1 text anchor1 https://example.com/very/long",
		"paragraph 2 text anchor2",
		"a wrapped",
	}
	if idx, ok := findAnchorLine(lines, "paragraph 2 text anchor2"); !ok || idx != 3 {
		t.Errorf("exact match: got (%d, %v), want (3, true)", idx, ok)
	}
	if idx, ok := findAnchorLine(lines, "paragraph 1 text anchor1"); !ok || idx != 2 {
		t.Errorf("prefix match over URL-extended line: got (%d, %v), want (2, true)", idx, ok)
	}
	if idx, ok := findAnchorLine(lines, "a wrapped name shifted"); !ok || idx != 4 {
		t.Errorf("common-prefix fallback: got (%d, %v), want (4, true)", idx, ok)
	}
	if _, ok := findAnchorLine(lines, "nothing matches this at all"); ok {
		t.Errorf("no-match scan should report false")
	}
}
