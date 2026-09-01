package ui

import "testing"

func TestFormatMB(t *testing.T) {
	cases := []struct {
		bytes int64
		want  string
	}{
		{0, "0MB"},
		{1, "0MB"},
		{1023, "0MB"},
		{1 << 20, "1MB"},
		{1<<20 + 1, "1MB"},
		{1<<21 - 1, "1MB"},
		{1 << 21, "2MB"},
		{1_500_000, "1MB"},
		{47_000_000, "44MB"},
		{320 * 1 << 20, "320MB"},
	}
	for _, c := range cases {
		if got := formatMB(c.bytes); got != c.want {
			t.Errorf("formatMB(%d) = %q, want %q", c.bytes, got, c.want)
		}
	}
}
