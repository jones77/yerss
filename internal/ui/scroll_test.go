package ui

import "testing"

func TestScrollStateThumbBounds(t *testing.T) {
	cases := []struct {
		name               string
		totalH, viewportH, trackH int
	}{
		{"long article", 100, 20, 22},
		{"very long", 1000, 5, 30},
		{"barely scrollable", 21, 20, 22},
		{"one extra line", 50, 49, 22},
		{"tiny viewport", 40, 1, 22},
		{"degenerate track", 100, 20, 1},
	}
	for _, c := range cases {
		if c.totalH <= c.viewportH {
			continue
		}
		maxH := max(1, c.trackH-1)
		for offset := 0; offset <= c.totalH-c.viewportH; offset++ {
			s := scrollState{totalH: c.totalH, viewportH: c.viewportH, offset: offset}
			top, h := s.thumb(c.trackH)
			if h < 1 || h > maxH {
				t.Errorf("%s (offset %d): thumb height %d out of bounds [1,%d]", c.name, offset, h, maxH)
			}
			if top < 0 || top+h > c.trackH {
				t.Errorf("%s (offset %d): thumb %d+%d outside track %d", c.name, offset, top, h, c.trackH)
			}
		}
	}
}

func TestScrollStateThumbPositions(t *testing.T) {
	// trackH 22, totalH 70, viewportH 20 → thumbH = 22*20/70 = 6, travel 16.
	top, h := (scrollState{totalH: 70, viewportH: 20, offset: 0}).thumb(22)
	if top != 0 || h != 6 {
		t.Errorf("top thumb = (%d,%d), want (0,6)", top, h)
	}
	// 21 of 50 scrollable rows (42%) → top = 16*21/50 = 6.
	top, h = (scrollState{totalH: 70, viewportH: 20, offset: 21}).thumb(22)
	if top != 6 || h != 6 {
		t.Errorf("42%% thumb = (%d,%d), want (6,6)", top, h)
	}
	// max offset 50 → top = travel = 16, sits at the bottom.
	top, h = (scrollState{totalH: 70, viewportH: 20, offset: 50}).thumb(22)
	if top != 16 || h != 6 {
		t.Errorf("100%% thumb = (%d,%d), want (16,6)", top, h)
	}
}

func TestScrollStateThumbFitsViewport(t *testing.T) {
	top, h := (scrollState{totalH: 10, viewportH: 20, offset: 0}).thumb(22)
	if top != 0 || h != 22 {
		t.Errorf("fits thumb = (%d,%d), want (0,22)", top, h)
	}
	top, h = (scrollState{totalH: 10, viewportH: 10, offset: 5}).thumb(22)
	if top != 0 || h != 22 {
		t.Errorf("exact-fit thumb = (%d,%d), want (0,22)", top, h)
	}
}

func TestScrollStatePercent(t *testing.T) {
	if got := (scrollState{totalH: 10, viewportH: 20, offset: 0}).percent(); got != 100 {
		t.Errorf("fits percent = %d, want 100", got)
	}
	if got := (scrollState{totalH: 10, viewportH: 10, offset: 5}).percent(); got != 100 {
		t.Errorf("exact-fit percent = %d, want 100", got)
	}
	if got := (scrollState{totalH: 70, viewportH: 20, offset: 0}).percent(); got != 0 {
		t.Errorf("top percent = %d, want 0", got)
	}
	if got := (scrollState{totalH: 70, viewportH: 20, offset: 21}).percent(); got != 42 {
		t.Errorf("42%% percent = %d, want 42", got)
	}
	if got := (scrollState{totalH: 70, viewportH: 20, offset: 50}).percent(); got != 100 {
		t.Errorf("bottom percent = %d, want 100", got)
	}
}