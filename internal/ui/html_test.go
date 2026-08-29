package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
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

func TestHTMLToTextListParagraphsRenderedAsBullets(t *testing.T) {
	out := HTMLToText("<ul><li><p>first point</p></li><li><p>second point</p></li></ul><p>after</p>")
	if !strings.Contains(out, "* first point\n") {
		t.Errorf("bullet prefix missing on same line as text: %q", out)
	}
	if !strings.Contains(out, "* second point\n") {
		t.Errorf("second bullet missing: %q", out)
	}
	if strings.Contains(out, "* \n\n") {
		t.Errorf("orphaned bullet marker between paragraphs: %q", out)
	}
	if !strings.Contains(out, "after") {
		t.Errorf("trailing paragraph missing: %q", out)
	}
}

func TestHTMLToTextListWithoutParagraphsUnchanged(t *testing.T) {
	out := HTMLToText("<ul><li>plain</li></ul>")
	if !strings.Contains(out, "* plain") {
		t.Errorf("plain list item should still render as a bullet: %q", out)
	}
}

func TestHTMLToTextNestedListIndents(t *testing.T) {
	out := HTMLToText("<ul><li>top<ul><li>nested one</li><li>nested two</li></ul></li><li>next top</li></ul>")
	want := "* top\n  * nested one\n  * nested two\n\n* next top"
	if out != want {
		t.Errorf("nested list rendering = %q, want %q", out, want)
	}
}

func TestHTMLToTextBulletSpacingIsConsistent(t *testing.T) {
	out := HTMLToText("<ul><li><p>first item</p></li><li><p>second item</p></li></ul>")
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "* ") {
			continue
		}
		if strings.HasPrefix(line, "*") && line != "*" {
			t.Errorf("bullet marker not followed by exactly one space: %q", line)
		}
	}
}

func TestHTMLToTextSVGTagsDoNotBreakLists(t *testing.T) {
	// An SVG <line> tag must not be mistaken for an <li> by the list
	// preprocessor (Substack pages embed SVG icons).
	html := `<button><svg><line x1="21" x2="14" y1="3" y2="10"></line></svg></button><ul><li><p>item one</p></li></ul>`
	out := HTMLToText(html)
	if !strings.Contains(out, "* item one") {
		t.Errorf("list after SVG line tag not rendered: %q", out)
	}
}

func TestHTMLToTextBoldRenderedAsAnsiBold(t *testing.T) {
	out := HTMLToText("<p><strong>bold text</strong> and <b>more</b> plain</p>")
	if !strings.Contains(out, "\x1b[1mbold text\x1b[22m") {
		t.Errorf("strong text should carry ANSI bold: %q", out)
	}
	if !strings.Contains(out, "\x1b[1mmore\x1b[22m") {
		t.Errorf("b text should carry ANSI bold: %q", out)
	}
	if strings.Contains(out, "*bold") || strings.Contains(out, "*more") {
		t.Errorf("markdown asterisks should be gone: %q", out)
	}
}

func TestHTMLToTextBoldInsideLink(t *testing.T) {
	out := HTMLToText(`<p><a href="https://example.com/x"><strong>bold link</strong></a></p>`)
	if !strings.Contains(out, "\x1b[1mbold link\x1b[22m") {
		t.Errorf("bold inside a link should stay ANSI bold: %q", out)
	}
	if strings.Contains(out, "*bold link*") {
		t.Errorf("markdown asterisks should be gone from link text: %q", out)
	}
}

func TestHTMLToTextBoldInListItem(t *testing.T) {
	out := HTMLToText("<ul><li><p><strong>lead</strong> item</p></li></ul>")
	if !strings.Contains(out, "* \x1b[1mlead\x1b[22m item") {
		t.Errorf("bold lead in a list item should render on the bullet line: %q", out)
	}
	if strings.Contains(out, "*\n") {
		t.Errorf("orphaned bullet marker: %q", out)
	}
}

func TestWrapTextKeepsBoldEscapes(t *testing.T) {
	out := wrapText("\x1b[1mword one two\x1b[22m three four five", 10, "…")
	for _, l := range strings.Split(out, "\n") {
		if ansi.StringWidth(l) > 10 {
			t.Errorf("line exceeds width: %q", l)
		}
	}
	if !strings.Contains(out, "\x1b[1m") || !strings.Contains(out, "\x1b[22m") {
		t.Errorf("bold escapes lost in wrap: %q", out)
	}
}

func TestHTMLToTextLinkTextNoInnerSpaces(t *testing.T) {
	url := "https://example.com/post"
	out := HTMLToText(`<p>See <a href="` + url + `"><span>world</span></a> now.</p>`)
	stripped := ansi.Strip(out)
	if !strings.Contains(stripped, "[world]("+url+")") {
		t.Errorf("link text should render as [world](url) with no padding spaces: %q", stripped)
	}
	if strings.Contains(stripped, "[ world ]") {
		t.Errorf("spaces remain between brackets and link text: %q", stripped)
	}
}

func TestHTMLToTextLinkTextKeepsInternalSpaces(t *testing.T) {
	out := HTMLToText(`<p><a href="https://example.com/n">New York Times</a></p>`)
	stripped := ansi.Strip(out)
	if !strings.Contains(stripped, "[New York Times]") {
		t.Errorf("internal word spaces in link text should be kept: %q", stripped)
	}
}

func TestWrapText(t *testing.T) {
	out := wrapText("one two three four five", 10, "…")
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
