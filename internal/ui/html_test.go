package ui

import (
	"strings"
	"testing"
)

func TestHTMLToTextPreservesLinksAndImages(t *testing.T) {
	html := `<p>Hello <a href="https://example.com/post">world</a>.</p><p><img src="cat.png" alt="photo of a cat"></p>`
	out := HTMLToText(html)
	if !strings.Contains(out, "world") {
		t.Errorf("link text not preserved: %q", out)
	}
	if !strings.Contains(out, "https://example.com/post") {
		t.Errorf("link URL not preserved: %q", out)
	}
	if !strings.Contains(out, "photo of a cat") {
		t.Errorf("image alt not rendered: %q", out)
	}
}

func TestHTMLToTextParagraphBreaks(t *testing.T) {
	out := HTMLToText("<p>first para</p><p>second para</p>")
	if !strings.Contains(out, "first para") || !strings.Contains(out, "second para") {
		t.Errorf("paragraph text missing: %q", out)
	}
}

func TestWrapText(t *testing.T) {
	out := wrapText("one two three four five", 10)
	lines := strings.Split(out, "\n")
	for _, l := range lines {
		if len(l) > 10 {
			t.Errorf("line exceeds width: %q", l)
		}
	}
	if len(lines) < 2 {
		t.Errorf("expected multiple lines, got %d", len(lines))
	}
}