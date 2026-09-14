package render

import "testing"

func TestContentGeom(t *testing.T) {
	cases := []struct {
		name                                  string
		w, h, padX, padY                      int
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
			textW, viewportH, effPadY := ContentGeom(tc.w, tc.h, tc.padX, tc.padY)
			if textW != tc.wantTextW || viewportH != tc.wantViewportH || effPadY != tc.wantEffPadY {
				t.Fatalf("ContentGeom(%d,%d,%d,%d) = (%d,%d,%d), want (%d,%d,%d)",
					tc.w, tc.h, tc.padX, tc.padY, textW, viewportH, effPadY,
					tc.wantTextW, tc.wantViewportH, tc.wantEffPadY)
			}
		})
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
		if got := WrapIndex(tc.i, tc.delta, tc.n); got != tc.want {
			t.Errorf("WrapIndex(%d,%d,%d) = %d, want %d", tc.i, tc.delta, tc.n, got, tc.want)
		}
	}
}