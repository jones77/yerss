package textutil

import "testing"

func TestFormatSize(t *testing.T) {
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}
	for _, tc := range tests {
		if got := FormatSize(tc.in); got != tc.want {
			t.Errorf("FormatSize(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestEscapeMarkdown(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"hello", "hello"},
		{"a*b", `a\*b`},
		{`a\b`, `a\\b`},
		{"[link]", `\[link\]`},
		{"<http>", `\<http\>`},
		{"_em_", `\_em\_`},
		{"`code`", "\\`code\\`"},
	}
	for _, tc := range tests {
		if got := EscapeMarkdown(tc.in); got != tc.want {
			t.Errorf("EscapeMarkdown(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestMaxLineWidth(t *testing.T) {
	if w := MaxLineWidth(nil); w != 0 {
		t.Errorf("MaxLineWidth(nil) = %d, want 0", w)
	}
	if w := MaxLineWidth([]string{"ab", "abcd", "a"}); w != 4 {
		t.Errorf("MaxLineWidth = %d, want 4", w)
	}
}
