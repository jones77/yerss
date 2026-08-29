package convert

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestWrapTextKeepsBoldEscapes(t *testing.T) {
	out := WrapText("\x1b[1mword one two\x1b[22m three four five", 10, "…")
	for _, l := range strings.Split(out, "\n") {
		if ansi.StringWidth(l) > 10 {
			t.Errorf("line exceeds width: %q", l)
		}
	}
	if !strings.Contains(out, "\x1b[1m") || !strings.Contains(out, "\x1b[22m") {
		t.Errorf("bold escapes lost in wrap: %q", out)
	}
}

func TestWrapText(t *testing.T) {
	out := WrapText("one two three four five", 10, "…")
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

func TestWrapTextWidthWithOSC8(t *testing.T) {
	url := "https://example.com/post"
	linked := ansi.SetHyperlink(url) + "world" + ansi.ResetHyperlink()
	out := WrapText(linked+" ", 10, "…")
	for _, l := range strings.Split(out, "\n") {
		if ansi.StringWidth(l) > 10 {
			t.Errorf("line width %d > 10: %q", ansi.StringWidth(l), l)
		}
	}
	if strings.Count(out, "\n") != 0 {
		t.Errorf("a short OSC 8 link should fit one line, got %q", out)
	}
}

func TestWrapTextBreaksURLToNewLine(t *testing.T) {
	url := "example.com/a"
	link := ansi.SetHyperlink(url) + "[text]" + ansi.ResetHyperlink() + "(" + url + ")"
	out := WrapText("abc "+link, 16, "…")
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), out)
	}
	if ansi.Strip(lines[0]) != "abc [text]" {
		t.Errorf("link text should trail the current line, got %q", ansi.Strip(lines[0]))
	}
	if ansi.Strip(lines[1]) != "(example.com/a)" {
		t.Errorf("URL should break onto its own line, got %q", ansi.Strip(lines[1]))
	}
}

func TestWrapTextTruncatesOverlongURL(t *testing.T) {
	url := "https://very-long-example.com/a-really-long-path"
	link := ansi.SetHyperlink(url) + "[text]" + ansi.ResetHyperlink() + "(" + url + ")"
	out := WrapText(link, 10, "…")
	lines := strings.Split(out, "\n")
	for i, l := range lines {
		if ansi.StringWidth(l) > 10 {
			t.Errorf("line %d exceeds width 10: %q", i, l)
		}
	}
	last := ansi.Strip(lines[len(lines)-1])
	if !strings.HasPrefix(last, "(") || !strings.HasSuffix(last, "…)") {
		t.Errorf("overlong URL should be truncated with an ellipsis: %q", last)
	}
}

func TestWrapTextLinkTextStaysTogether(t *testing.T) {
	url := "http://reallylong.com/link/that/stretches/across/a/link"
	link := ansi.SetHyperlink(url) + "[link text followed by]" + ansi.ResetHyperlink() + "(" + url + ")"
	out := WrapText(link, 30, "…")
	lines := strings.Split(out, "\n")
	for i, l := range lines {
		if ansi.StringWidth(l) > 30 {
			t.Errorf("line %d exceeds width 30: %q", i, ansi.Strip(l))
		}
	}
	first := ansi.Strip(lines[0])
	if first != "[link text followed by]" {
		t.Errorf("link text should stay whole on its own line, got %q", first)
	}
	last := ansi.Strip(lines[len(lines)-1])
	if !strings.HasPrefix(last, "(") || !strings.HasSuffix(last, "…)") {
		t.Errorf("URL should break after the bracket and truncate, got %q", last)
	}
}

func TestWrapTextLinkStaysTogetherOnOwnLine(t *testing.T) {
	url := "https://example.com/post"
	link := ansi.SetHyperlink(url) + "[world]" + ansi.ResetHyperlink() + "(" + url + ")"
	out := WrapText("Read the full story here: "+link+" more", 50, "…")
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), out)
	}
	if ansi.Strip(lines[0]) != "Read the full story here: [world]" {
		t.Errorf("link text should trail the current line: %q", ansi.Strip(lines[0]))
	}
	if ansi.Strip(lines[1]) != "(https://example.com/post)" {
		t.Errorf("URL should be on the next line: %q", ansi.Strip(lines[1]))
	}
	if ansi.Strip(lines[2]) != "more" {
		t.Errorf("line 2 = %q", ansi.Strip(lines[2]))
	}
}

func TestWrapTextKeepsTrailingPunctWithLink(t *testing.T) {
	url := "https://example.com/post"
	link := ansi.SetHyperlink(url) + "[world]" + ansi.ResetHyperlink() + "(" + url + ")"
	out := WrapText("See "+link+", plus more", 40, "…")
	lines := strings.Split(out, "\n")
	if ansi.Strip(lines[0]) != "See [world](https://example.com/post)," {
		t.Errorf("punctuation should stay on the link line: %q", ansi.Strip(lines[0]))
	}
	if ansi.Strip(lines[1]) != "plus more" {
		t.Errorf("line 1 = %q", ansi.Strip(lines[1]))
	}
}

func TestWrapTextTruncatesURLToKeepPunct(t *testing.T) {
	url := "https://example.com/very-long-path-that-wont-fit"
	link := ansi.SetHyperlink(url) + "[world]" + ansi.ResetHyperlink() + "(" + url + ")"
	out := WrapText(link+".", 30, "…")
	lines := strings.Split(out, "\n")
	last := ansi.Strip(lines[len(lines)-1])
	if ansi.StringWidth(last) > 30 {
		t.Errorf("line exceeds width 30: %q", last)
	}
	if !strings.HasSuffix(last, ").") {
		t.Errorf("punctuation should stay with the truncated link: %q", last)
	}
}

func TestWrapTextKeepsCurlyQuotePunctWithLink(t *testing.T) {
	url := "https://example.com/post"
	link := ansi.SetHyperlink(url) + "[world]" + ansi.ResetHyperlink() + "(" + url + ")"
	out := WrapText(link+",\u201d more", 60, "…")
	lines := strings.Split(out, "\n")
	if ansi.Strip(lines[0]) != "[world](https://example.com/post),\u201d" {
		t.Errorf("curly-quote punctuation should trail the link: %q", ansi.Strip(lines[0]))
	}
	if ansi.Strip(lines[1]) != "more" {
		t.Errorf("line 1 = %q", ansi.Strip(lines[1]))
	}
}

func TestWrapTextPreservesNestedListIndentation(t *testing.T) {
	in := `<ul><li><p>top item</p><ul><li><p>nested one</p></li><li><p>nested two</p></li></ul></li><li><p>next top</p></li></ul>`
	out := WrapText(Convert(in), 60, "…")
	found := false
	for _, l := range strings.Split(out, "\n") {
		stripped := ansi.Strip(l)
		if strings.Contains(stripped, "nested one") || strings.Contains(stripped, "nested two") {
			found = true
			if !strings.HasPrefix(l, "  * ") {
				t.Errorf("nested bullet lost indentation: %q", l)
			}
		}
		if strings.Contains(stripped, "top item") && strings.HasPrefix(l, " ") {
			t.Errorf("top-level bullet wrongly indented: %q", l)
		}
	}
	if !found {
		t.Errorf("nested bullets not found: %q", out)
	}
}

func TestWrapTextPreservesIndentOnWrappedLines(t *testing.T) {
	// A long nested item wraps; every continuation line keeps the indent.
	in := `<ul><li><p>top item</p><ul><li><p>this nested item is long enough that it wraps across several lines when the content width is narrow</p></li></ul></li></ul>`
	out := WrapText(Convert(in), 30, "…")
	lines := strings.Split(out, "\n")
	start := -1
	for i, l := range lines {
		if strings.Contains(ansi.Strip(l), "nested item is long") {
			start = i
			break
		}
	}
	if start < 0 {
		t.Fatalf("nested bullet not found: %q", out)
	}
	if len(lines) <= start+1 {
		t.Errorf("expected the nested item to wrap across multiple lines: %q", out)
	}
	for i := start + 1; i < len(lines); i++ {
		l := lines[i]
		if strings.TrimSpace(ansi.Strip(l)) == "" {
			continue
		}
		if !strings.HasPrefix(l, "  ") {
			t.Errorf("wrapped continuation line %d lost indent: %q", i, l)
		}
	}
}

func TestWrapTextLinkEndsLine(t *testing.T) {
	url := "example.com/a"
	link := ansi.SetHyperlink(url) + "[text]" + ansi.ResetHyperlink() + "(" + url + ")"
	out := WrapText("see "+link+" now", 40, "…")
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), out)
	}
	if ansi.Strip(lines[0]) != "see [text](example.com/a)" {
		t.Errorf("line 0 should end with the link: %q", ansi.Strip(lines[0]))
	}
	if ansi.Strip(lines[1]) != "now" {
		t.Errorf("line 1 should hold the text after the link: %q", ansi.Strip(lines[1]))
	}
}