package ui

import (
	"testing"
	"time"

	"yerss/internal/store"
)

func TestContentGeom(t *testing.T) {
	cases := []struct {
		name                         string
		w, h, padX, padY             int
		wantTextW, wantViewportH, wantEffPadY int
	}{
		{"normal", 80, 24, 2, 1, 74, 20, 1},
		{"no padding", 80, 24, 0, 0, 78, 22, 0},
		{"degenerate height", 80, 2, 2, 1, 74, 1, 0},
		{"padding exceeds height", 30, 6, 2, 2, 24, 2, 1},
		{"zero width", 0, 24, 2, 1, 1, 20, 1},
		{"zero height", 80, 0, 2, 1, 74, 1, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			textW, viewportH, effPadY := contentGeom(tc.w, tc.h, tc.padX, tc.padY)
			if textW != tc.wantTextW || viewportH != tc.wantViewportH || effPadY != tc.wantEffPadY {
				t.Fatalf("contentGeom(%d,%d,%d,%d) = (%d,%d,%d), want (%d,%d,%d)",
					tc.w, tc.h, tc.padX, tc.padY, textW, viewportH, effPadY,
					tc.wantTextW, tc.wantViewportH, tc.wantEffPadY)
			}
		})
	}
}

func TestContentGeomStateAndRectAgree(t *testing.T) {
	m, _ := newTestModel(t)
	m.width = 30
	m.height = 6
	m.cfg.Display.PaddingX = 2
	m.cfg.Display.PaddingY = 2

	st := m.newArticleState(store.Article{ID: 1})
	_, _, w, h := m.contentRect()
	if st.viewport.Width != w || st.viewport.Height != h {
		t.Fatalf("newArticleState viewport (%d,%d) != contentRect (%d,%d)",
			st.viewport.Width, st.viewport.Height, w, h)
	}
}

func TestWrapIndex(t *testing.T) {
	cases := []struct {
		i, delta, n, want int
	}{
		{0, 1, 5, 1},
		{4, 1, 5, 0},
		{0, -1, 5, 4},
		{2, 5, 5, 2},
		{0, -6, 5, 4},
		{0, 0, 5, 0},
		{3, 0, 0, 3},
		{1, 2, 1, 0},
		{0, 1, 1, 0},
	}
	for _, tc := range cases {
		if got := wrapIndex(tc.i, tc.delta, tc.n); got != tc.want {
			t.Errorf("wrapIndex(%d,%d,%d) = %d, want %d", tc.i, tc.delta, tc.n, got, tc.want)
		}
	}
}

func TestDayKey(t *testing.T) {
	if got := dayKey(time.Time{}); got != "undated" {
		t.Errorf("dayKey(zero) = %q, want %q", got, "undated")
	}
	day := time.Date(2026, 8, 28, 15, 4, 0, 0, time.Local)
	if got := dayKey(day); got != "20260828" {
		t.Errorf("dayKey(%v) = %q, want %q", day, got, "20260828")
	}
	nextDay := time.Date(2026, 8, 29, 0, 0, 0, 0, time.Local)
	if got := dayKey(nextDay); got != "20260829" {
		t.Errorf("dayKey(%v) = %q, want %q", nextDay, got, "20260829")
	}
}

func TestDayStart(t *testing.T) {
	in := time.Date(2026, 8, 28, 23, 59, 59, 999999999, time.Local)
	got := dayStart(in)
	want := time.Date(2026, 8, 28, 0, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Errorf("dayStart(%v) = %v, want %v", in, got, want)
	}
}
