package ui

import (
	"testing"
	"time"

	"yerss/internal"
)

func TestContentGeomStateAndRectAgree(t *testing.T) {
	m, _ := newTestModel(t)
	m.width = 30
	m.height = 6
	m.article.SetSize(m.width, m.height)
	m.sess.Config().Display.PaddingX = 2
	m.sess.Config().Display.PaddingY = 2

	st := m.article.Recompose(internal.Article{ID: 1})
	_, _, w, h := m.article.ContentRect()
	if st.Viewport.Width != w || st.Viewport.Height != h {
		t.Fatalf("newArticleState viewport (%d,%d) != contentRect (%d,%d)",
			st.Viewport.Width, st.Viewport.Height, w, h)
	}
}

func TestDayKey(t *testing.T) {
	if got := internal.DayKey(time.Time{}); got != "undated" {
		t.Errorf("DayKey(zero) = %q, want %q", got, "undated")
	}
	day := time.Date(2026, 8, 28, 15, 4, 0, 0, time.Local)
	if got := internal.DayKey(day); got != "20260828" {
		t.Errorf("DayKey(%v) = %q, want %q", day, got, "20260828")
	}
	nextDay := time.Date(2026, 8, 29, 0, 0, 0, 0, time.Local)
	if got := internal.DayKey(nextDay); got != "20260829" {
		t.Errorf("DayKey(%v) = %q, want %q", nextDay, got, "20260829")
	}
}

func TestDayStart(t *testing.T) {
	in := time.Date(2026, 8, 28, 23, 59, 59, 999999999, time.Local)
	got := internal.DayStart(in)
	want := time.Date(2026, 8, 28, 0, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Errorf("DayStart(%v) = %v, want %v", in, got, want)
	}
}
