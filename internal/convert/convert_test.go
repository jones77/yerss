package convert

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestConvertPreservesLinksAndImages(t *testing.T) {
	html := `<p>Hello <a href="https://example.com/post">world</a>.</p><p><img src="cat.png" alt="photo of a cat"></p>`
	out := Convert(html)
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

func TestConvertParagraphBreaks(t *testing.T) {
	out := Convert("<p>first para</p><p>second para</p>")
	if !strings.Contains(out, "first para") || !strings.Contains(out, "second para") {
		t.Errorf("paragraph text missing: %q", out)
	}
}

func TestConvertListParagraphsRenderedAsBullets(t *testing.T) {
	out := Convert("<ul><li><p>first point</p></li><li><p>second point</p></li></ul><p>after</p>")
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

func TestConvertListWithoutParagraphsUnchanged(t *testing.T) {
	out := Convert("<ul><li>plain</li></ul>")
	if !strings.Contains(out, "* plain") {
		t.Errorf("plain list item should still render as a bullet: %q", out)
	}
}

func TestConvertNestedListIndents(t *testing.T) {
	out := Convert("<ul><li>top<ul><li>nested one</li><li>nested two</li></ul></li><li>next top</li></ul>")
	want := "* top\n\n  * nested one\n  * nested two\n* next top"
	if out != want {
		t.Errorf("nested list rendering = %q, want %q", out, want)
	}
}

func TestConvertBulletSpacingIsConsistent(t *testing.T) {
	out := Convert("<ul><li><p>first item</p></li><li><p>second item</p></li></ul>")
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "* ") {
			continue
		}
		if strings.HasPrefix(line, "*") && line != "*" {
			t.Errorf("bullet marker not followed by exactly one space: %q", line)
		}
	}
}

func TestConvertSVGTagsDoNotBreakLists(t *testing.T) {
	// An SVG <line> tag must not be mistaken for an <li> by the list
	// renderer (Substack pages embed SVG icons).
	html := `<button><svg><line x1="21" x2="14" y1="3" y2="10"></line></svg></button><ul><li><p>item one</p></li></ul>`
	out := Convert(html)
	if !strings.Contains(out, "* item one") {
		t.Errorf("list after SVG line tag not rendered: %q", out)
	}
}

func TestConvertBoldRenderedAsMarkdown(t *testing.T) {
	out := Convert("<p><strong>bold text</strong> and <b>more</b> plain</p>")
	if !strings.Contains(out, "**bold text**") {
		t.Errorf("strong text should render as markdown bold: %q", out)
	}
	if !strings.Contains(out, "**more**") {
		t.Errorf("b text should render as markdown bold: %q", out)
	}
	if strings.Contains(out, "\x1b") {
		t.Errorf("converter output should not contain ANSI escapes: %q", out)
	}
}

func TestConvertBoldInsideLink(t *testing.T) {
	out := Convert(`<p><a href="https://example.com/x"><strong>bold link</strong></a></p>`)
	if !strings.Contains(out, "[**bold link**](https://example.com/x)") {
		t.Errorf("bold inside a link should stay markdown bold: %q", out)
	}
	if strings.Contains(out, "\x1b") {
		t.Errorf("converter output should not contain ANSI escapes: %q", out)
	}
}

func TestConvertBoldInListItem(t *testing.T) {
	out := Convert("<ul><li><p><strong>lead</strong> item</p></li></ul>")
	if !strings.Contains(out, "* **lead** item") {
		t.Errorf("bold lead in a list item should render on the bullet line: %q", out)
	}
	if strings.Contains(out, "*\n") {
		t.Errorf("orphaned bullet marker: %q", out)
	}
}

func TestConvertRendersLinksAsMarkdown(t *testing.T) {
	url := "https://example.com/post"
	html := `<p>Hello <a href="` + url + `">world</a>.</p>`
	out := Convert(html)

	if !strings.Contains(out, "[world]("+url+")") {
		t.Errorf("link should render as [world](url): %q", out)
	}
	if strings.Contains(out, "\x1b") {
		t.Errorf("converter output should not contain ANSI escapes: %q", out)
	}
	stripped := out
	if !strings.Contains(stripped, "Hello") {
		t.Errorf("non-link text should be unaffected: %q", stripped)
	}
}

func TestConvertSingleQuotedHref(t *testing.T) {
	url := "https://example.com/post"
	html := `<p>See <a href='` + url + `'>post</a> now.</p>`
	out := Convert(html)
	if !strings.Contains(out, "[post]("+url+")") {
		t.Errorf("single-quoted href should render as markdown link: %q", out)
	}
}

func TestConvertPlainParagraphHasNoEscapes(t *testing.T) {
	out := Convert("<p>just plain text</p>")
	if strings.Contains(out, "\x1b") {
		t.Errorf("plain text should not contain escape sequences: %q", out)
	}
}

func TestConvertLinkTextNoInnerSpaces(t *testing.T) {
	url := "https://example.com/post"
	out := Convert(`<p>See <a href="` + url + `"><span>world</span></a> now.</p>`)
	stripped := ansi.Strip(out)
	if !strings.Contains(stripped, "[world]("+url+")") {
		t.Errorf("link text should render as [world](url) with no padding spaces: %q", stripped)
	}
	if strings.Contains(stripped, "[ world ]") {
		t.Errorf("spaces remain between brackets and link text: %q", stripped)
	}
}

func TestConvertLinkTextKeepsInternalSpaces(t *testing.T) {
	out := Convert(`<p><a href="https://example.com/n">New York Times</a></p>`)
	stripped := ansi.Strip(out)
	if !strings.Contains(stripped, "[New York Times]") {
		t.Errorf("internal word spaces in link text should be kept: %q", stripped)
	}
}

func TestConvertImgAltEscaped(t *testing.T) {
	out := Convert(`<p><img alt="a *b* [c]"></p>`)
	if !strings.Contains(out, "[a \\*b\\* \\[c\\]]") {
		t.Errorf("image alt text should be escaped so special chars render literally: %q", out)
	}
	if strings.Contains(out, "[a *b* [c]]") {
		t.Errorf("unescaped alt text would be interpreted as markdown: %q", out)
	}
}

func TestConvertErrorFallsBackToPlainText(t *testing.T) {
	got := plainText("<p>Hello</p><div>World</div>")
	if got != "Hello World" {
		t.Errorf("plainText should join block text with spaces, got %q", got)
	}
	if strings.Contains(got, "<") || strings.Contains(got, ">") {
		t.Errorf("plainText should strip HTML tags, got %q", got)
	}
}

func TestConvertImagesReturnsSentinelsInDocumentOrder(t *testing.T) {
	html := `<p>before</p><p><img src="a.jpg" alt="first"></p><p>middle</p><p><img src="b.jpg"></p><p>after</p>`
	out, imgs := ConvertImages(html)
	// Sentinels appear in document order.
	iA := strings.Index(out, "\x00img:a.jpg\x00")
	iB := strings.Index(out, "\x00img:b.jpg\x00")
	if iA < 0 || iB < 0 {
		t.Fatalf("sentinel tokens missing from output: %q", out)
	}
	if iA > iB {
		t.Errorf("sentinel order wrong: a.jpg at %d after b.jpg at %d", iA, iB)
	}
	if !strings.Contains(out, "before") || !strings.Contains(out, "middle") || !strings.Contains(out, "after") {
		t.Errorf("surrounding text not preserved: %q", out)
	}
	if len(imgs) != 2 {
		t.Fatalf("inline images = %d, want 2", len(imgs))
	}
	if imgs[0].URL != "a.jpg" || imgs[1].URL != "b.jpg" {
		t.Errorf("urls = %+v, want [a.jpg b.jpg]", imgs)
	}
	if imgs[0].Alt != "first" {
		t.Errorf("first alt = %q, want %q", imgs[0].Alt, "first")
	}
	if imgs[1].Alt != "" {
		t.Errorf("second alt = %q, want empty", imgs[1].Alt)
	}
}

func TestConvertImagesNoImages(t *testing.T) {
	out, imgs := ConvertImages("<p>just text</p>")
	if imgs != nil {
		t.Errorf("imgs = %v, want nil", imgs)
	}
	if !strings.Contains(out, "just text") {
		t.Errorf("text not preserved: %q", out)
	}
}

func TestConvertImagesLinkedImageWithLinkText(t *testing.T) {
	// A promo banner <a> wraps the image and the banner text; the marked
	// converter renders the sentinel inside the link with the text between the
	// sentinel and the destination, which the article view consumes whole.
	html := `<p><a href="/tomdispatch"><span><img src="logo.jpg"></span><div><p>Read Our Complete Coverage</p><h2>TomDispatch</h2></div></a></p>`
	out, imgs := ConvertImages(html)
	if !strings.Contains(out, "\x00img:logo.jpg\x00") {
		t.Errorf("sentinel missing from linked-image output: %q", out)
	}
	if !strings.Contains(out, "](/tomdispatch)") {
		t.Errorf("link tail missing from linked-image output: %q", out)
	}
	if len(imgs) != 1 || imgs[0].URL != "logo.jpg" {
		t.Errorf("imgs = %+v, want [logo.jpg]", imgs)
	}
}

func TestConvertStillEmitsImagePlaceholders(t *testing.T) {
	// Regression: the plain Convert path must keep emitting [alt]/[image]
	// placeholders, not sentinels.
	out := Convert(`<p><img src="a.jpg" alt="first"></p><p><img src="b.jpg"></p>`)
	if strings.Contains(out, "\x00img:") {
		t.Errorf("Convert must not emit sentinels: %q", out)
	}
	if !strings.Contains(out, "[first]") {
		t.Errorf("alt placeholder missing: %q", out)
	}
	if !strings.Contains(out, "[image]") {
		t.Errorf("no-alt image placeholder missing: %q", out)
	}
}
