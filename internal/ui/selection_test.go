package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal"
	"yerss/internal/app"
	"yerss/internal/config"
)

func openSharedStore(t *testing.T) *internal.Store {
	t.Helper()
	st, err := internal.Open(filepath.Join(t.TempDir(), "sel.sqlite"))
	if err != nil {
		t.Fatalf("internal.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func modelOn(t *testing.T, st *internal.Store) *Model {
	t.Helper()
	m := New(app.New(config.Default(), st))
	m.width = 80
	m.height = 24
	return m
}

func TestRestoreSelectionRestoresListCursor(t *testing.T) {
	withLocalZone(t, time.UTC)
	st := openSharedStore(t)
	m1 := modelOn(t, st)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "b", time.Date(2026, 8, 28, 11, 0, 0, 0, time.UTC))
	m1.list.Load()
	if m1.list.RowCount() != 3 {
		t.Fatalf("expected 3 rows, got %d", m1.list.RowCount())
	}
	m1.list.SetCursor(2) // article b
	m1.persistSelection()

	m2 := modelOn(t, st)
	m2.list.Load()
	m2.restoreSelection()
	if m2.list.Cursor() != 2 {
		t.Errorf("restored cursor = %d, want 2", m2.list.Cursor())
	}
	if m2.view != viewList {
		t.Errorf("restored view = %d, want list", m2.view)
	}
}

func TestRestoreSelectionRestoresHeader(t *testing.T) {
	withLocalZone(t, time.UTC)
	st := openSharedStore(t)
	m1 := modelOn(t, st)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	m1.list.Load()
	m1.list.SetCursor(0) // day header
	m1.persistSelection()

	m2 := modelOn(t, st)
	m2.list.Load()
	m2.restoreSelection()
	if m2.list.Cursor() != 0 {
		t.Errorf("restored header cursor = %d, want 0", m2.list.Cursor())
	}
	if !m2.list.CursorIsHeader() {
		t.Errorf("restored row should be a day header")
	}
}

func TestRestoreSelectionReopensArticle(t *testing.T) {
	withLocalZone(t, time.UTC)
	st := openSharedStore(t)
	m1 := modelOn(t, st)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "b", time.Date(2026, 8, 28, 11, 0, 0, 0, time.UTC))
	m1.list.Load()
	m1.list.SetCursor(2) // article b
	m1.openArticle()
	if m1.view != viewArticle {
		t.Fatalf("expected article view, got %d", m1.view)
	}
	m1.article.Viewport.ScrollDown(7)
	wantOffset := m1.article.Viewport.YOffset
	m1.persistSelection()

	m2 := modelOn(t, st)
	m2.list.Load()
	m2.restoreSelection()
	if m2.view != viewArticle {
		t.Errorf("restored view = %d, want article", m2.view)
	}
	if m2.article.ID() != m1.article.ID() {
		t.Errorf("restored article id = %d, want %d", m2.article.ID(), m1.article.ID())
	}
	if m2.article.Viewport.YOffset != wantOffset {
		t.Errorf("restored scroll offset = %d, want %d", m2.article.Viewport.YOffset, wantOffset)
	}

	m2.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if m2.view != viewList {
		t.Errorf("back from restored article should return to list, view = %d", m2.view)
	}
	if m2.list.CursorIsHeader() || m2.list.ArticleIDAtRow(m2.list.Cursor()) != m2.article.ID() {
		t.Errorf("back should land on the restored article row")
	}
}

func TestArticleScrollOffsetSurvivesWindowSize(t *testing.T) {
	withLocalZone(t, time.UTC)
	st := openSharedStore(t)
	m1 := modelOn(t, st)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	m1.list.Load()
	m1.list.SetCursor(1)
	m1.openArticle()
	m1.article.Viewport.ScrollDown(7)
	wantOffset := m1.article.Viewport.YOffset
	m1.persistSelection()

	m2 := modelOn(t, st)
	m2.list.Load()
	m2.restoreSelection()
	if m2.view != viewArticle {
		t.Fatalf("expected article view after restore, got %d", m2.view)
	}
	if m2.article.Viewport.YOffset != wantOffset {
		t.Fatalf("restored offset = %d, want %d", m2.article.Viewport.YOffset, wantOffset)
	}

	// The terminal's first WindowSizeMsg recreates the article state; the
	// scroll offset must survive it.
	m2.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if m2.article.Viewport.YOffset != wantOffset {
		t.Errorf("offset after WindowSizeMsg = %d, want %d", m2.article.Viewport.YOffset, wantOffset)
	}
}

func TestRestoreSelectionAnchorsToContentWithImages(t *testing.T) {
	withLocalZone(t, time.UTC)
	st := openSharedStore(t)
	a := photoHeavyArticle(4)
	var body strings.Builder
	for i := 0; i < 60; i++ {
		fmt.Fprintf(&body, "<p>paragraph-%d filler text</p>", i)
	}
	a.Content = "<p>lead-in</p>" + a.Content + body.String() + "<p>trailing-distinct-line</p>"
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	for pos := 0; pos < 5; pos++ {
		url := a.ImageURL
		if pos > 0 {
			url = fmt.Sprintf("https://example.com/img%d.jpg", pos)
		}
		if err := st.SetArticleImage(id, internal.ArticleImage{
			Position: pos, URL: url, Photo: pngBytes(t, 320, 160),
		}); err != nil {
			t.Fatal(err)
		}
	}
	full, err := st.GetArticle(id)
	if err != nil || full == nil {
		t.Fatalf("GetArticle(%d): %v", id, err)
	}

	// Quit state: fully composed, all images loaded, scrolled to a mid-body
	// text position (a unique paragraph line so the anchor is unambiguous).
	m1 := modelOn(t, st)
	m1.article = m1.article.Recompose(*full)
	m1.view = viewArticle
	drain(m1, m1.images.FireImageLoad(*full, &m1.article))
	want := -1
	for i, ln := range m1.article.Lines() {
		if strings.Contains(ansi.Strip(ln), "paragraph-30") {
			want = i
			break
		}
	}
	if want < 0 {
		t.Fatal("no paragraph line found in composed article")
	}
	m1.article.Viewport.SetYOffset(want)
	wantTop := strings.TrimSpace(ansi.Strip(m1.article.Lines()[want]))
	m1.persistSelection()

	sel, err := st.LoadLastSelection()
	if err != nil {
		t.Fatal(err)
	}
	if sel.ArticleAnchor == "" {
		t.Fatal("expected a content anchor to be persisted")
	}

	// Restart: fresh model, empty in-memory image caches -> placeholder
	// layout; restoreSelection relocates the anchor and fires image loads.
	m2 := modelOn(t, st)
	m2.list.Load()
	drain(m2, m2.restoreSelection())
	if m2.view != viewArticle {
		t.Fatalf("expected article view after restore, got %d", m2.view)
	}
	got := strings.TrimSpace(ansi.Strip(m2.article.Lines()[m2.article.Viewport.YOffset]))
	if got != wantTop {
		t.Errorf("restored top line = %q, want %q (offset %d -> %d)",
			got, wantTop, want, m2.article.Viewport.YOffset)
	}
	if m2.article.Viewport.YOffset != want {
		t.Errorf("restored offset = %d, want %d", m2.article.Viewport.YOffset, want)
	}
}

func TestRestoreSelectionSurvivesStartupResize(t *testing.T) {
	withLocalZone(t, time.UTC)
	st := openSharedStore(t)
	a := photoHeavyArticle(4)
	var body strings.Builder
	for i := 0; i < 60; i++ {
		fmt.Fprintf(&body, "<p>paragraph-%d filler text</p>", i)
	}
	a.Content = "<p>lead-in</p>" + a.Content + body.String() + "<p>trailing-distinct-line</p>"
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	for pos := 0; pos < 5; pos++ {
		url := a.ImageURL
		if pos > 0 {
			url = fmt.Sprintf("https://example.com/img%d.jpg", pos)
		}
		if err := st.SetArticleImage(id, internal.ArticleImage{
			Position: pos, URL: url, Photo: pngBytes(t, 320, 160),
		}); err != nil {
			t.Fatal(err)
		}
	}
	full, err := st.GetArticle(id)
	if err != nil || full == nil {
		t.Fatalf("GetArticle(%d): %v", id, err)
	}

	// Quit state: fully composed at the terminal's real size (76x22), scrolled
	// to a unique mid-body text line.
	const realW, realH = 76, 22
	m1 := modelOn(t, st)
	m1.article.SetSize(realW, realH)
	m1.article = m1.article.Recompose(*full)
	m1.view = viewArticle
	drain(m1, m1.images.FireImageLoad(*full, &m1.article))
	want := -1
	for i, ln := range m1.article.Lines() {
		if strings.Contains(ansi.Strip(ln), "paragraph-30") {
			want = i
			break
		}
	}
	if want < 0 {
		t.Fatal("no paragraph line found in composed article")
	}
	m1.article.Viewport.SetYOffset(want)
	wantTop := strings.TrimSpace(ansi.Strip(m1.article.Lines()[want]))
	m1.persistSelection()

	// Restart: fresh model, restoreSelection runs at the hardcoded default
	// size, then the terminal reports its real size, then image loads land.
	m2 := modelOn(t, st)
	m2.list.Load()
	drain(m2, m2.restoreSelection())
	if m2.view != viewArticle {
		t.Fatalf("expected article view after restore, got %d", m2.view)
	}
	_, cmd := m2.Update(tea.WindowSizeMsg{Width: realW, Height: realH})
	drain(m2, cmd)
	drain(m2, nil)
	got := strings.TrimSpace(ansi.Strip(m2.article.Lines()[m2.article.Viewport.YOffset]))
	if got != wantTop {
		t.Errorf("restored top line = %q, want %q (offset %d -> %d)",
			got, wantTop, want, m2.article.Viewport.YOffset)
	}
}

func TestRestoreSelectionTopExpandedStaysAtTop(t *testing.T) {
	withLocalZone(t, time.UTC)
	st := openSharedStore(t)
	// The reported bug: an expanded-mode article whose header is the URL line
	// plus a long wrapping title, and whose lead photo also opens the body, so
	// it composes as a suppressed-lead block after a few intro lines (not a
	// lead directly under the header). Quitting at the very top and restarting
	// must land back at the very top once the photo load lands — not snap the
	// viewport forward to the photo and scroll the header off.
	a := internal.Article{
		FeedURL:  "https://example.com/feed.xml",
		GUID:     "g-restore-top-expanded",
		Title:    "The U.S. Has Killed 227 Civilians in Boat Strikes During \u201cA Year of Murders\u201d",
		Author:   "Nick Turse",
		Link:     "https://theintercept.com/2026/09/02/military-boat-strikes-killings-year/",
		ImageURL: "https://example.com/lead.jpg",
		Content: "<p>American troops have long plied the waters off the coast of West Africa, " +
			"hunting for arms smugglers, drug runners, and the shadowy networks that supply them. " +
			"But in the year since the United States dramatically escalated its patrols, the boats " +
			"that crews have fired on have killed more than two hundred civilians.</p>" +
			`<figure><img src="https://example.com/lead.jpg"><figcaption>Photo: Nick Turse</figcaption></figure>` +
			"<p>The investigation drew on interviews with dozens of people inside and outside the military. " +
			"Officials described a campaign that has grown more aggressive with each passing month, " +
			"even as the civilian toll has mounted.</p>" +
			"<p>Read more: [theintercept.com/2026/09/02/military-boat-strikes-killings-year](https://theintercept.com/2026/09/02/military-boat-strikes-killings-year/).</p>",
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.SetArticleImage(id, internal.ArticleImage{
		Position: 0, URL: a.ImageURL, Photo: pngBytes(t, 440, 440),
	}); err != nil {
		t.Fatal(err)
	}
	full, err := st.GetArticle(id)
	if err != nil || full == nil {
		t.Fatalf("GetArticle(%d): %v", id, err)
	}

	// Session 1: expanded, photo loaded, sitting at the very top.
	const realW, realH = 76, 22
	m1 := modelOn(t, st)
	m1.SetLinkExpand(true)
	m1.article.SetSize(realW, realH)
	m1.article = m1.article.Recompose(*full)
	m1.view = viewArticle
	drain(m1, m1.images.FireImageLoad(*full, &m1.article))
	// The fixture must be the suppressed-lead geometry the bug needs: a single
	// photo block that materializes BELOW the header, not a lead under it.
	blocks := m1.article.ImageBlocks()
	if len(blocks) != 1 {
		t.Fatalf("expected one photo block, got %d", len(blocks))
	}
	if blocks[0].ImgStart <= m1.article.HeaderLines() {
		t.Fatalf("fixture photo block starts at %d (header %d); want it below the header", blocks[0].ImgStart, m1.article.HeaderLines())
	}
	if blocks[0].ImgStart >= realH {
		t.Fatalf("fixture photo block starts at %d; want it inside the 22-row window so the snap case applies", blocks[0].ImgStart)
	}
	m1.article.Viewport.SetYOffset(0)
	m1.persistSelection()
	sel, err := st.LoadLastSelection()
	if err != nil {
		t.Fatal(err)
	}
	if sel.ArticleOffset != 0 {
		t.Fatalf("persisted offset = %d, want 0 (the top)", sel.ArticleOffset)
	}

	// Session 2: fresh model restores the expanded mode, restoreSelection runs
	// at the hardcoded default size, the terminal reports its real size BEFORE
	// the photo load lands (the real app's message ordering), then the load
	// lands. The viewport must stay at the top through all of it.
	m2 := modelOn(t, st)
	m2.list.Load()
	if expanded, err := m2.sess.LoadLinkExpandMode(); err == nil {
		m2.SetLinkExpand(expanded)
	}
	imgLoad := m2.restoreSelection()
	if m2.view != viewArticle {
		t.Fatalf("expected article view after restore, got %d", m2.view)
	}
	_, cmd := m2.Update(tea.WindowSizeMsg{Width: realW, Height: realH})
	drain(m2, imgLoad) // the photo load lands after the resize
	drain(m2, cmd)
	drain(m2, nil)

	if got := m2.article.Viewport.YOffset; got != 0 {
		t.Errorf("after restart at the top: offset = %d, want 0 so the header stays visible", got)
	}
	first := strings.TrimSpace(ansi.Strip(m2.article.Lines()[m2.article.Viewport.YOffset]))
	if !strings.Contains(first, "theintercept.com/2026/09/02") {
		t.Errorf("first visible line = %q, want the expanded URL header line", first)
	}
}

func TestRestoreSelectionFallsBackWhenArticleGone(t *testing.T) {
	withLocalZone(t, time.UTC)
	st := openSharedStore(t)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	// Save a selection naming an article that does not exist.
	if err := st.SaveLastSelection(internal.LastSelection{View: "article", ArticleID: 9999}); err != nil {
		t.Fatalf("SaveLastSelection: %v", err)
	}

	m := modelOn(t, st)
	m.list.Load()
	m.restoreSelection()
	if m.list.Cursor() != 0 {
		t.Errorf("cursor should fall back to 0, got %d", m.list.Cursor())
	}
	if m.view != viewList {
		t.Errorf("view should stay list when the article is gone, got %d", m.view)
	}
}
