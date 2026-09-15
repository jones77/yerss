package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/app"
	"yerss/internal/config"
	"yerss/internal/store"
	"yerss/internal/ui/articleview"
	"yerss/internal/ui/compose"
	"yerss/internal/ui/render"
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
	m.article = m.article.Recompose(a)
	return m, st
}

func bodyText(m *Model) string {
	return ansi.Strip(strings.Join(m.article.Lines(), "\n"))
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
	m.article.Viewport.SetYOffset(1)
	off := m.article.Viewport.YOffset

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if !strings.Contains(bodyText(m), "world https://example.com/post") {
		t.Errorf("after x: want expanded link: %q", bodyText(m))
	}
	if got := m.article.Viewport.YOffset; got != off {
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
	m.article = m.article.Recompose(a2)
	if !strings.Contains(bodyText(m), "again https://example.com/post") {
		t.Errorf("expanded mode did not persist across article switches: %q", bodyText(m))
	}
}

func TestCompactLinkSpanIsClickable(t *testing.T) {
	// Compact rendering drops the href span but keeps the OSC 8 name span, so
	// the left-click hit-testing reports exactly the name's URL and columns.
	m, _ := openTestArticle(t, "spans", `<p>hello [world](https://example.com/post) end</p>`)
	found := false
	for i, ln := range m.article.Lines() {
		if !strings.Contains(ansi.Strip(ln), "world") {
			continue
		}
		found = true
		spans := articleview.ParseLinkSpans(ln)
		if len(spans) != 1 {
			t.Fatalf("compact line %d has %d link spans, want 1: %+v", i, len(spans), spans)
		}
		sp := spans[0]
		if sp.URL != "https://example.com/post" {
			t.Errorf("compact line %d span url = %q", i, sp.URL)
		}
		name := ansi.Strip(ln)
		start := strings.Index(name, "world")
		if start < 0 || sp.Start != start || sp.End != start+len("world") {
			t.Errorf("compact line %d span (%d,%d) should cover the name at %d..%d: %q", i, sp.Start, sp.End, start, start+len("world"), name)
		}
	}
	if !found {
		t.Fatal("article body missing the link line")
	}
}

func TestReaderHeaderURLRendersBlue(t *testing.T) {
	m, _ := newTestModel(t)
	m.SetLinkExpand(true)
	m.article = m.article.Recompose(store.Article{Title: "t", Link: "https://example.com/post", Content: "<p>x</p>"})
	blue := escapePrefix(lipgloss.NewStyle().Foreground(m.palette.StatusBar).Render("x"))
	if !strings.Contains(m.article.Lines()[0], blue) {
		t.Errorf("header URL line should render in the status-bar blue role: %q", m.article.Lines()[0])
	}
	if first := ansi.Strip(m.article.Lines()[0]); !strings.Contains(first, "example.com/post") {
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
	for i, ln := range m.article.Lines() {
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
	// even when the shared URL wraps across rendered lines. Once the hrefs are
	// stripped the paragraph re-flows, so the names flow back together.
	m, _ := openTestArticle(t, "adjacent", `<p>before [one](`+wrapURL+`) and [two](`+wrapURL+`) after</p>`)
	body := paragraphText(m)
	if got := strings.Join(strings.Fields(body), " "); got != "before one and two after" {
		t.Errorf("compact body = %q, want %q", got, "before one and two after")
	}
	if strings.Contains(body, "example.com") {
		t.Errorf("compact body leaks the URL: %q", body)
	}

	// Non-vacuous guard: the shared URL genuinely wraps at this width in
	// expanded mode, so "after" lands on a later line than "before".
	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	bi, ai := -1, -1
	for i, ln := range m.article.Lines() {
		s := ansi.Strip(ln)
		if bi < 0 && strings.Contains(s, "before") {
			bi = i
		}
		if strings.Contains(s, "after") {
			ai = i
		}
	}
	if ai <= bi {
		t.Fatalf("shared URL did not wrap ('after' on line %d, 'before' on line %d); test is vacuous", ai, bi)
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
	m.article = m.article.Recompose(store.Article{Title: "t", Link: "https://example.com/post", Content: "<p>x</p>"})
	lines := strippedLines(m.article.Lines())
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

// toggleTopFixture builds an article with a Link so the expanded header adds
// the URL line above the title, and scrolls to the very top.
func toggleTopFixture(t *testing.T) (*Model, *store.Store) {
	t.Helper()
	var html strings.Builder
	for i := 1; i <= 30; i++ {
		fmt.Fprintf(&html, `<p>paragraph %d text anchor%d [j](%s)</p>`, i, i, wrapURL)
	}
	m, st := newTestModel(t)
	a := store.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "g-toggle-top",
		Title:   "top-title",
		Link:    "https://example.com/article",
		Content: html.String(),
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id
	m.article = m.article.Recompose(a)
	m.article.Viewport.SetYOffset(0)
	return m, st
}

func TestLinkExpandToggleAtTopShowsURLHeader(t *testing.T) {
	// The title is the first visible line; expanding prepends the URL header
	// above it, and the viewport must show that new first line — not re-anchor
	// on the title and scroll the URL off.
	m, _ := toggleTopFixture(t)
	if got := strings.TrimRight(ansi.Strip(m.article.Lines()[0]), " "); got != "top-title" {
		t.Fatalf("first line before toggle = %q, want the title", got)
	}

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if m.article.Viewport.YOffset != 0 {
		t.Errorf("after x at the top: offset = %d, want 0 so the URL line shows", m.article.Viewport.YOffset)
	}
	first := strings.TrimSpace(ansi.Strip(m.article.Lines()[m.article.Viewport.YOffset]))
	if !strings.Contains(first, "example.com/article") {
		t.Errorf("first visible line after x = %q, want the expanded URL header", first)
	}
	if got := strings.TrimSpace(ansi.Strip(m.article.Lines()[2])); got != "top-title" {
		t.Errorf("title line = %q, want it below the URL header", got)
	}
}

func TestLinkExpandToggleAtTopCollapseStaysAtTop(t *testing.T) {
	m, _ := toggleTopFixture(t)
	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if !strings.Contains(bodyText(m), "example.com/article") {
		t.Fatalf("expected expanded after x: %q", bodyText(m))
	}

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if m.article.Viewport.YOffset != 0 {
		t.Errorf("after collapsing at the top: offset = %d, want 0", m.article.Viewport.YOffset)
	}
	if got := strings.TrimRight(ansi.Strip(m.article.Lines()[0]), " "); got != "top-title" {
		t.Errorf("first visible line after collapse = %q, want the title back at the top", got)
	}
}

func TestLinkExpandToggleKeepsFirstVisibleLine(t *testing.T) {
	m, _ := anchorToggleFixture(t, 30)
	if strings.Contains(bodyText(m), "example.com") {
		t.Fatalf("expected compact by default: %q", bodyText(m))
	}
	oldLen := len(m.article.Lines())
	// Scroll partway down so the toggle genuinely shifts line indices.
	m.article.Viewport.SetYOffset(2)
	oldOffset := m.article.Viewport.YOffset
	anchor := strings.TrimSpace(ansi.Strip(m.article.Lines()[oldOffset]))
	if anchor == "" {
		t.Fatalf("anchor line at offset %d is empty", oldOffset)
	}

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	// Non-vacuous guard: expanded mode really added URL lines.
	if newLen := len(m.article.Lines()); newLen <= oldLen {
		t.Fatalf("link toggle did not grow the article (%d -> %d lines); test is vacuous", oldLen, newLen)
	}
	// Total-delta bump is exactly what the toggle must NOT do.
	inserted := len(m.article.Lines()) - oldLen
	if got := m.article.Viewport.YOffset; got == oldOffset+inserted {
		t.Errorf("toggle jumped to the total-delta offset %d; the first visible line drifted", got)
	}
	// The anchored paragraph is the first visible line again, with its leading
	// words intact (a link grew on the line, so trailing text can move).
	first := strings.TrimSpace(ansi.Strip(m.article.Lines()[m.article.Viewport.YOffset]))
	if !strings.HasPrefix(first, anchor) {
		t.Errorf("first visible line after toggle = %q, want it to start with the anchor %q", first, anchor)
	}
}

func TestLinkExpandToggleAtBottomStaysPinned(t *testing.T) {
	m, _ := anchorToggleFixture(t, 40)
	m.article.Viewport.GotoBottom()
	preBottom := m.article.Viewport.YOffset
	if preBottom == 0 {
		t.Fatalf("article does not scroll; test is vacuous")
	}

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	got := m.article.Viewport.YOffset
	maxOffset := len(m.article.Lines()) - m.article.Viewport.Height
	if got != maxOffset {
		t.Errorf("after x at the bottom: offset = %d, want the new bottom %d (was %d)", got, maxOffset, preBottom)
	}
}

func TestLinkExpandToggleBackReturnsToAnchor(t *testing.T) {
	m, _ := anchorToggleFixture(t, 30)
	m.article.Viewport.SetYOffset(2)
	anchor := strings.TrimSpace(ansi.Strip(m.article.Lines()[m.article.Viewport.YOffset]))

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if !strings.Contains(bodyText(m), "example.com") {
		t.Fatalf("expected expanded after x: %q", bodyText(m))
	}
	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if got := strings.TrimSpace(ansi.Strip(m.article.Lines()[m.article.Viewport.YOffset])); got != anchor {
		t.Errorf("after toggling back: first visible line = %q, want the original anchor %q", got, anchor)
	}
}

func TestLinkExpandToggleBlankFirstLineKeepsSeparator(t *testing.T) {
	m, _ := anchorToggleFixture(t, 30)
	oldLen := len(m.article.Lines())
	blank := -1
	for i := 1; i < oldLen-1; i++ {
		if strings.TrimSpace(ansi.Strip(m.article.Lines()[i])) == "" &&
			strings.TrimSpace(ansi.Strip(m.article.Lines()[i+1])) != "" {
			blank = i
			break
		}
	}
	if blank < 0 {
		t.Fatal("no blank separator line found; test is vacuous")
	}
	m.article.Viewport.SetYOffset(blank)
	oldOffset := m.article.Viewport.YOffset
	if strings.TrimSpace(ansi.Strip(m.article.Lines()[oldOffset])) != "" {
		t.Fatalf("viewport top is not the blank line: %q", ansi.Strip(m.article.Lines()[oldOffset]))
	}
	anchorLine := strings.TrimSpace(ansi.Strip(m.article.Lines()[oldOffset+1]))

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if newLen := len(m.article.Lines()); newLen <= oldLen {
		t.Fatalf("link toggle did not grow the article (%d -> %d lines); test is vacuous", oldLen, newLen)
	}
	if got := strings.TrimSpace(ansi.Strip(m.article.Lines()[m.article.Viewport.YOffset])); got != "" {
		t.Errorf("after toggle the first visible line = %q, want the blank separator", got)
	}
	next := strings.TrimSpace(ansi.Strip(m.article.Lines()[m.article.Viewport.YOffset+1]))
	if !strings.HasPrefix(next, strings.Split(anchorLine, " ")[0]) {
		t.Errorf("paragraph below the separator = %q, want it to start with the anchor %q", next, anchorLine)
	}
}

func TestLinkExpandToggleURLFirstLineNearEndGluesBottom(t *testing.T) {
	m, st := newTestModel(t)
	m.SetLinkExpand(true)
	longURL := "https://example.com/" + strings.Repeat("segment-", 300) + "end"
	var html strings.Builder
	for i := 0; i < 20; i++ {
		fmt.Fprintf(&html, "<p>filler paragraph %d with some words</p>", i)
	}
	fmt.Fprintf(&html, "<p>tail [link](%s)</p>", longURL)
	a := store.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "g-url-tail",
		Title:   "url-tail",
		Content: html.String(),
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id
	m.article = m.article.Recompose(a)

	oldLen := len(m.article.Lines())
	maxOffset := oldLen - m.article.Viewport.Height
	if maxOffset <= 0 {
		t.Fatalf("article does not scroll; test is vacuous")
	}
	// The viewport top must land on a URL-only continuation line, with only URL
	// or blank lines below it, so no anchorable word survives the mode change.
	top := strings.TrimSpace(ansi.Strip(m.article.Lines()[maxOffset]))
	if !strings.Contains(top, "segment-") {
		t.Fatalf("viewport top %d is not a URL continuation line: %q", maxOffset, top)
	}
	m.article.Viewport.SetYOffset(maxOffset)
	if got := m.article.Viewport.YOffset; got != maxOffset {
		t.Fatalf("could not position at %d (got %d)", maxOffset, got)
	}

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if strings.Contains(bodyText(m), "example.com") {
		t.Fatalf("expected compact after toggle: %q", bodyText(m))
	}
	want := len(m.article.Lines()) - m.article.Viewport.Height
	if got := m.article.Viewport.YOffset; got != want {
		t.Errorf("URL-first-line toggle offset = %d, want glued to bottom %d", got, want)
	}
}

func TestLinkExpandToggleReflowReanchorsFirstWord(t *testing.T) {
	// A paragraph whose compact line reflows across modes: the anchor is the
	// pre-toggle first visible line's text, and the new first visible line must
	// contain that line's first word so the reader can re-orient.
	name := "Richard Jebreal and the Democratic Party of Nothing is quite long and keeps going"
	var html strings.Builder
	for i := 0; i < 20; i++ {
		fmt.Fprintf(&html, "<p>filler paragraph %d with some words</p>", i)
	}
	html.WriteString(`<p>before [` + name + `](` + wrapURL + `) after</p>`)
	for i := 0; i < 20; i++ {
		fmt.Fprintf(&html, "<p>trailing paragraph %d with some words</p>", i)
	}
	m, _ := openTestArticle(t, "reflow", html.String())
	oldLen := len(m.article.Lines())
	oldOffset := -1
	for i, ln := range m.article.Lines() {
		if strings.Contains(ansi.Strip(ln), "Richard") {
			oldOffset = i
			break
		}
	}
	if oldOffset < 0 {
		t.Fatal("article body missing the reflow paragraph")
	}
	m.article.Viewport.SetYOffset(oldOffset)
	if got := m.article.Viewport.YOffset; got != oldOffset {
		t.Fatalf("could not position at the reflow line %d (got %d)", oldOffset, got)
	}
	firstWord := strings.Fields(strings.TrimSpace(ansi.Strip(m.article.Lines()[oldOffset])))[0]

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if newLen := len(m.article.Lines()); newLen == oldLen {
		t.Fatalf("link toggle did not change the article; test is vacuous")
	}
	first := strings.TrimSpace(ansi.Strip(m.article.Lines()[m.article.Viewport.YOffset]))
	if !strings.HasPrefix(first, firstWord) {
		t.Errorf("first visible line after toggle = %q, want it to start with %q", first, firstWord)
	}
}

// photoToggleModel builds a model on an article whose lead photo composes a
// block and whose body wraps a long URL across lines, so toggling link mode
// changes the line count while the photo block stays present.
func photoToggleModel(t *testing.T) (*Model, *store.Store) {
	t.Helper()
	a := store.Article{
		FeedURL:  "https://example.com/feed.xml",
		GUID:     "g-photo-toggle",
		Title:    "Headline",
		Author:   "Jane",
		Link:     "https://example.com/a",
		ImageURL: "https://example.com/lead.jpg",
		Content: `<p>intro text with a [wrapped link](` + wrapURL + `) and more</p>` +
			strings.Repeat(`<p>trailing paragraph `+wrapURL+` body text</p>`, 20),
	}
	m, st := newTestModel(t)
	fr := &fakeRenderer{fixed: []string{"IMG1", "IMG2", "IMG3", "IMG4", "IMG5", "IMG6", "IMG7", "IMG8", "IMG9", "IMG10", "IMG11", "IMG12"}}
	m.sess.ImgRenderer = fr
	m.sess.ImgCache.Set(a.ImageURL, testImg())
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id
	m.article = m.article.Recompose(a)
	return m, st
}

func findBlock(m *Model, url string) *compose.ImageBlock {
	for i := range m.article.ImageBlocks() {
		if m.article.ImageBlocks()[i].URL == url {
			return &m.article.ImageBlocks()[i]
		}
	}
	return nil
}

// photoRowCount returns the number of photo rows the block renders (its
// attribution rows follow).
func photoRowCount(b *compose.ImageBlock) int {
	return b.CapStart - b.ImgStart
}

func TestLinkExpandToggleKeepsPhotoOnScreen(t *testing.T) {
	m, _ := photoToggleModel(t)
	b := findBlock(m, "https://example.com/lead.jpg")
	if b == nil {
		t.Fatal("lead photo block not composed")
	}
	const rowRel = 3 // scroll a few rows into the photo
	m.article.Viewport.SetYOffset(b.ImgStart + rowRel)
	oldLen := len(m.article.Lines())

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if newLen := len(m.article.Lines()); newLen <= oldLen {
		t.Fatalf("link toggle did not re-render differently (%d -> %d lines); test is vacuous", oldLen, newLen)
	}
	nb := findBlock(m, "https://example.com/lead.jpg")
	if nb == nil {
		t.Fatal("lead photo block vanished after toggle")
	}
	// The same row of the same photo must lie at the viewport start.
	want := nb.ImgStart + rowRel
	if got := m.article.Viewport.YOffset; got != want {
		t.Errorf("after toggle: offset = %d, want the same photo row at %d (block %d..%d)", got, want, nb.ImgStart, nb.ImgEnd)
	}
	// The photo stays on screen.
	if m.article.Viewport.YOffset > nb.ImgEnd {
		t.Errorf("photo block %d..%d left the viewport (offset %d)", nb.ImgStart, nb.ImgEnd, m.article.Viewport.YOffset)
	}
}

func TestLinkExpandToggleKeepsAttributionOnScreen(t *testing.T) {
	m, _ := photoToggleModel(t)
	b := findBlock(m, "https://example.com/lead.jpg")
	if b == nil {
		t.Fatal("lead photo block not composed")
	}
	// First visible line is the photo's attribution line (still inside the
	// block's span); the photo must stay on screen above it.
	m.article.Viewport.SetYOffset(b.CapStart)
	oldLen := len(m.article.Lines())

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if newLen := len(m.article.Lines()); newLen <= oldLen {
		t.Fatalf("link toggle did not re-render differently (%d -> %d lines); test is vacuous", oldLen, newLen)
	}
	nb := findBlock(m, "https://example.com/lead.jpg")
	if nb == nil {
		t.Fatal("lead photo block vanished after toggle")
	}
	want := nb.ImgStart + (b.CapStart - b.ImgStart)
	if got := m.article.Viewport.YOffset; got != want {
		t.Errorf("after toggle: offset = %d, want the same attribution row at %d (block %d..%d)", got, want, nb.ImgStart, nb.ImgEnd)
	}
	if m.article.Viewport.YOffset > nb.ImgEnd {
		t.Errorf("photo block %d..%d left the viewport (offset %d)", nb.ImgStart, nb.ImgEnd, m.article.Viewport.YOffset)
	}
}

func TestLinkExpandToggleDoesNotResizePhoto(t *testing.T) {
	m, _ := photoToggleModel(t)
	b := findBlock(m, "https://example.com/lead.jpg")
	if b == nil {
		t.Fatal("lead photo block not composed")
	}
	compactRows := photoRowCount(b)

	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	nb := findBlock(m, "https://example.com/lead.jpg")
	if nb == nil {
		t.Fatal("lead photo block vanished after toggle")
	}
	expandedRows := photoRowCount(nb)
	if expandedRows != compactRows {
		t.Errorf("photo re-sized on toggle: %d rows compact vs %d expanded", compactRows, expandedRows)
	}
	if expandedRows < 1 {
		t.Errorf("photo block has no rows; test is vacuous")
	}
}

func TestFitHeaderLinesIsMaxAcrossModes(t *testing.T) {
	m, _ := photoToggleModel(t)
	contentW, _, _ := render.ContentGeom(m.width, m.height, m.sess.Config().Display.PaddingX, m.sess.Config().Display.PaddingY)
	compact := len(strings.Split(m.article.RenderHeaderMode(*m.article.Article(), contentW, false), "\n"))
	expanded := len(strings.Split(m.article.RenderHeaderMode(*m.article.Article(), contentW, true), "\n"))
	want := max(compact, expanded)
	if expanded == compact {
		t.Fatalf("headers identical across modes; test is vacuous")
	}
	if got := m.article.FitHeaderLines(); got != want {
		t.Errorf("fitHeaderLines = %d, want the tallest header across modes (%d)", got, want)
	}
}

func TestLinkExpandModePersistsAcrossRestart(t *testing.T) {
	withLocalZone(t, time.UTC)
	st := openSharedStore(t)
	a := store.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "guid-persist",
		Title:   "persist-title",
		Link:    "https://example.com/article",
		Content: `<p>hello [world](https://example.com/post) end</p>`,
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id

	// Session 1: open the article, toggle to expanded, quit.
	m1 := modelOn(t, st)
	m1.article = m1.article.Recompose(a)
	m1.view = viewArticle
	m1.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if !strings.Contains(bodyText(m1), "example.com/post") {
		t.Fatalf("expected expanded after x: %q", bodyText(m1))
	}
	m1.persistSelection()

	// Session 2: a fresh model runs Init, which loads the saved mode before
	// the selection restore, so the reopened article composes in expanded mode.
	m2 := modelOn(t, st)
	_ = m2.Init()
	if m2.view != viewArticle {
		t.Fatalf("expected article view after restart, got %d", m2.view)
	}
	if !m2.linkExpand {
		t.Errorf("restored linkExpand = false, want true")
	}
	first := strings.TrimRight(ansi.Strip(m2.article.Lines()[0]), " ")
	if !strings.Contains(first, "example.com/article") {
		t.Errorf("first content line = %q, want the URL-above-title header", first)
	}
	if strings.Contains(first, "example.com/post") {
		t.Errorf("first content line leaks the body link URL: %q", first)
	}
	if got := strings.TrimSpace(ansi.Strip(m2.article.Lines()[1])); got != "" {
		t.Errorf("second content line = %q, want a single blank line", got)
	}
	if got := strings.TrimRight(ansi.Strip(m2.article.Lines()[2]), " "); got != "persist-title" {
		t.Errorf("third content line = %q, want the title", got)
	}
}

func TestLinkExpandToggleDoesNotTouchConfigFile(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.toml")
	orig := "# sentinel config — no schema header, untouched by migrate\n[display]\ntheme = \"dark\"\n"
	if err := os.WriteFile(cfgPath, []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	st, err := store.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	m := New(app.New(cfg, st))
	m.width = 80
	m.height = 24

	a := store.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "guid-config",
		Title:   "cfg",
		Content: `<p>hello [world](https://example.com/post) end</p>`,
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id
	m.article = m.article.Recompose(a)
	m.view = viewArticle
	m.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	m.persistSelection()

	got, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != orig {
		t.Errorf("config file changed after toggle+quit:\n got: %q\nwant: %q", got, orig)
	}

	expanded, err := st.LoadLinkExpandMode()
	if err != nil {
		t.Fatalf("LoadLinkExpandMode: %v", err)
	}
	if !expanded {
		t.Errorf("mode was not recorded in the state table after toggle+quit")
	}
}
