package articleview

import (
	"strings"
	"testing"
	"time"

	"yerss/internal"
)

func bottomBorderLine(frame string) string {
	lines := strings.Split(frame, "\n")
	return lines[len(lines)-1]
}

func TestNavHintReplacesOpenHintAndReverts(t *testing.T) {
	s, _ := newTestState(t)
	s = s.Recompose(internal.Article{Title: "t", Content: "<p>body</p>"})

	s.navHint = navHintNext
	s.navHintExpires = time.Now().Add(time.Minute)
	last := bottomBorderLine(s.Render())
	if !strings.Contains(last, "K: next") {
		t.Errorf("active navigation hint missing: %q", last)
	}
	if strings.Contains(last, "o: open article in browser") {
		t.Errorf("open hint must be replaced while a navigation hint shows: %q", last)
	}

	s.navHintExpires = time.Now().Add(-time.Minute)
	last = bottomBorderLine(s.Render())
	if got := s.NavHint(); got != "" {
		t.Errorf("NavHint() = %q after expiry, want empty", got)
	}
	if !strings.Contains(last, "o: open article in browser") {
		t.Errorf("open hint must return after expiry: %q", last)
	}
}
