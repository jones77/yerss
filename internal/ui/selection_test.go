package ui

import (
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal/app"
	"yerss/internal/config"
	"yerss/internal/store"
)

func openSharedStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "sel.sqlite"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func modelOn(t *testing.T, st *store.Store) *Model {
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

func TestRestoreSelectionFallsBackWhenArticleGone(t *testing.T) {
	withLocalZone(t, time.UTC)
	st := openSharedStore(t)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	// Save a selection naming an article that does not exist.
	if err := st.SaveLastSelection(store.LastSelection{View: "article", ArticleID: 9999}); err != nil {
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