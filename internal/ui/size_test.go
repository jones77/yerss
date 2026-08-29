package ui

import "testing"

func TestFormatSize(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{450, "450 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1500, "1.5 KB"},
		{15360, "15.0 KB"},
		{1500000, "1.4 MB"},
		{1048576, "1.0 MB"},
		{2000000000, "1.9 GB"},
		{1099511627776, "1.0 TB"},
	}
	for _, c := range cases {
		if got := formatSize(c.in); got != c.want {
			t.Errorf("formatSize(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}