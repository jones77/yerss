package feed

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFeedsFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "feeds.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadFeedsStripsSchemesAndRewrites(t *testing.T) {
	path := writeFeedsFile(t, "# my feeds\n"+
		"https://example.com/feed\n"+
		"\n"+
		"HTTP://Example.com/Other\n"+
		"bare.example.org/rss\n")

	urls, err := LoadFeeds(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"example.com/feed", "Example.com/Other", "bare.example.org/rss"}
	if len(urls) != len(want) {
		t.Fatalf("urls = %v, want %v", urls, want)
	}
	for i := range want {
		if urls[i] != want[i] {
			t.Errorf("urls[%d] = %q, want %q", i, urls[i], want[i])
		}
	}

	// The file is rewritten in place with the scheme-less canonical form,
	// preserving comments and blank lines verbatim.
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	wantFile := "# my feeds\nexample.com/feed\n\nExample.com/Other\nbare.example.org/rss\n"
	if string(got) != wantFile {
		t.Errorf("rewritten file = %q, want %q", string(got), wantFile)
	}

	// A second load is a no-op: the file is already canonical.
	urls, err = LoadFeeds(path)
	if err != nil {
		t.Fatal(err)
	}
	got, _ = os.ReadFile(path)
	if string(got) != wantFile || len(urls) != len(want) {
		t.Errorf("idempotent load changed the file to %q or the urls to %v", string(got), urls)
	}
}

func TestLoadFeedsLeavesCanonicalFileUntouched(t *testing.T) {
	content := "# feeds\nexample.com/feed\n"
	path := writeFeedsFile(t, content)
	if _, err := LoadFeeds(path); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != content {
		t.Errorf("canonical file was rewritten: %q", string(got))
	}
}

func TestStripScheme(t *testing.T) {
	cases := map[string]string{
		"https://example.com/f": "example.com/f",
		"HTTPS://example.com/f": "example.com/f",
		"http://example.com/f":  "example.com/f",
		"Http://example.com/f":  "example.com/f",
		"example.com/f":         "example.com/f",
	}
	for in, want := range cases {
		if got := stripScheme(in); got != want {
			t.Errorf("stripScheme(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCanonicalURL(t *testing.T) {
	cases := map[string]string{
		"example.com/feed":     "https://example.com/feed",
		"example.com/feed.xml": "https://example.com/feed.xml",
	}
	for in, want := range cases {
		if got := CanonicalURL(in); got != want {
			t.Errorf("CanonicalURL(%q) = %q, want %q", in, got, want)
		}
	}
	if got := CanonicalURL("ftp://example.com/f"); !strings.HasPrefix(got, "ftp://") {
		t.Errorf("CanonicalURL should pass through entries that carry a scheme, got %q", got)
	}
}
