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
	m1.loadList()
	if len(m1.visibleRows()) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(m1.visibleRows()))
	}
	m1.list.cursor = 2 // article b
	m1.persistSelection()

	m2 := modelOn(t, st)
	m2.loadList()
	m2.restoreSelection()
	if m2.list.cursor != 2 {
		t.Errorf("restored cursor = %d, want 2", m2.list.cursor)
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
	m1.loadList()
	m1.list.cursor = 0 // day header
	m1.persistSelection()

	m2 := modelOn(t, st)
	m2.loadList()
	m2.restoreSelection()
	if m2.list.cursor != 0 {
		t.Errorf("restored header cursor = %d, want 0", m2.list.cursor)
	}
	rows := m2.visibleRows()
	if rows[m2.list.cursor].kind != rowHeader {
		t.Errorf("restored row should be a day header")
	}
}

func TestRestoreSelectionReopensArticle(t *testing.T) {
	withLocalZone(t, time.UTC)
	st := openSharedStore(t)
	m1 := modelOn(t, st)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	insertTimedArticle(t, st, "b", time.Date(2026, 8, 28, 11, 0, 0, 0, time.UTC))
	m1.loadList()
	m1.list.cursor = 2 // article b
	m1.openArticle()
	if m1.view != viewArticle {
		t.Fatalf("expected article view, got %d", m1.view)
	}
	m1.article.viewport.ScrollDown(7)
	wantOffset := m1.article.viewport.YOffset
	m1.persistSelection()

	m2 := modelOn(t, st)
	m2.loadList()
	m2.restoreSelection()
	if m2.view != viewArticle {
		t.Errorf("restored view = %d, want article", m2.view)
	}
	if m2.article.id != m1.article.id {
		t.Errorf("restored article id = %d, want %d", m2.article.id, m1.article.id)
	}
	if m2.article.viewport.YOffset != wantOffset {
		t.Errorf("restored scroll offset = %d, want %d", m2.article.viewport.YOffset, wantOffset)
	}

	m2.updateArticle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if m2.view != viewList {
		t.Errorf("back from restored article should return to list, view = %d", m2.view)
	}
	rows := m2.visibleRows()
	row := rows[m2.list.cursor]
	if row.kind != rowArticle || m2.list.groups[row.groupIdx].articles[row.artIdx].ID != m2.article.id {
		t.Errorf("back should land on the restored article row")
	}
}

func TestArticleScrollOffsetSurvivesWindowSize(t *testing.T) {
	withLocalZone(t, time.UTC)
	st := openSharedStore(t)
	m1 := modelOn(t, st)
	insertTimedArticle(t, st, "a", time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC))
	m1.loadList()
	m1.list.cursor = 1
	m1.openArticle()
	m1.article.viewport.ScrollDown(7)
	wantOffset := m1.article.viewport.YOffset
	m1.persistSelection()

	m2 := modelOn(t, st)
	m2.loadList()
	m2.restoreSelection()
	if m2.view != viewArticle {
		t.Fatalf("expected article view after restore, got %d", m2.view)
	}
	if m2.article.viewport.YOffset != wantOffset {
		t.Fatalf("restored offset = %d, want %d", m2.article.viewport.YOffset, wantOffset)
	}

	// The terminal's first WindowSizeMsg recreates the article state; the
	// scroll offset must survive it.
	m2.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if m2.article.viewport.YOffset != wantOffset {
		t.Errorf("offset after WindowSizeMsg = %d, want %d", m2.article.viewport.YOffset, wantOffset)
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
	m.loadList()
	m.restoreSelection()
	if m.list.cursor != 0 {
		t.Errorf("cursor should fall back to 0, got %d", m.list.cursor)
	}
	if m.view != viewList {
		t.Errorf("view should stay list when the article is gone, got %d", m.view)
	}
}