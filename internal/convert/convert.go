// Package convert converts article HTML to a markdown display string. The
// output is plain commonmark: structured HTML (headings, lists, blockquotes,
// code blocks, tables) renders as native markdown, and inline <img> elements
// render as [alt]/[image] placeholders. Terminal styling is applied downstream
// by the glamour renderer, so the converter emits no ANSI escapes.
package convert

import (
	"bytes"
	"strings"

	"github.com/JohannesKaufmann/dom"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	"golang.org/x/net/html"

	"yerss/internal/textutil"
)

var conv = newConverter()

// Convert converts article HTML to a plain commonmark markdown string.
// Structured HTML (headings, lists, blockquotes, code blocks, tables) renders
// as native markdown and [alt]/[image] placeholders stand in for <img>.
func Convert(html string) string {
	if strings.TrimSpace(html) == "" {
		return ""
	}
	out, err := conv.ConvertString(html)
	if err != nil {
		return plainText(html)
	}
	return out
}

// InlineImage is one inline <img> element discovered in an article's content:
// its source URL and its alt text (which may be empty). They are returned in
// document order so the article view can render each image in place with its
// own attribution.
type InlineImage struct {
	URL string
	Alt string
}

// ConvertImages converts article HTML to markdown where each inline <img> is
// replaced by a sentinel token "\x00img:<url>\x00" (so the caller can split
// the body on it and interleave image blocks), and returns that marked
// markdown together with the ordered list of inline images. Unlike Convert it
// does not lose the image source; both paths share the same converter
// construction, so Convert is unchanged.
func ConvertImages(html string) (string, []InlineImage) {
	if strings.TrimSpace(html) == "" {
		return "", nil
	}
	var imgs []InlineImage
	conv := newMarkedConverter(func(src, alt string) {
		imgs = append(imgs, InlineImage{URL: src, Alt: alt})
	})
	out, err := conv.ConvertString(html)
	if err != nil {
		return plainText(html), imgs
	}
	return out, imgs
}

// newConverter builds the shared html-to-markdown converter. Escaping is
// disabled because the output is fed to a markdown renderer (glamour) and
// never re-parsed for display; smart escaping would only add backslash noise.
func newConverter() *converter.Converter {
	return buildConverter(renderImage)
}

// newMarkedConverter builds a converter whose <img> renderer writes a
// "\x00img:<url>\x00" sentinel instead of the [alt]/[image] placeholder and
// reports each image (source and alt) to onImg as it is emitted, in document
// order. URLs cannot contain NUL, so the sentinel is unambiguous.
func newMarkedConverter(onImg func(src, alt string)) *converter.Converter {
	return buildConverter(func(_ converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
		src, _ := dom.GetAttribute(n, "src")
		alt, _ := dom.GetAttribute(n, "alt")
		onImg(src, strings.TrimSpace(alt))
		w.WriteString("\x00img:" + src + "\x00")
		return converter.RenderSuccess
	})
}

// buildConverter constructs the html-to-markdown converter with img handled by
// the given renderer. Escaping is disabled because the output is fed to a
// markdown renderer (glamour) and never re-parsed for display; smart escaping
// would only add backslash noise.
func buildConverter(imgRender func(converter.Context, converter.Writer, *html.Node) converter.RenderStatus) *converter.Converter {
	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(
				commonmark.WithBulletListMarker("*"),
				commonmark.WithListEndComment(false),
			),
			table.NewTablePlugin(),
		),
		converter.WithEscapeMode(converter.EscapeModeDisabled),
	)
	conv.Register.PostRenderer(collapseBlankLines, converter.PriorityLate)
	conv.Register.PreRenderer(suppressSingleImageCaptions, converter.PriorityEarly)
	conv.Register.RendererFor("img", converter.TagTypeInline, imgRender, converter.PriorityEarly)
	return conv
}

// suppressSingleImageCaptions removes the figcaption of any figure that holds
// exactly one <img> — its own photo's caption — from the body markdown, because
// the attribution path composes that caption centered beneath the photo;
// emitting it as body text would duplicate it. A figcaption over multiple
// images (a gallery's shared credit, a photo-grid label) belongs to no single
// image and stays as body text.
func suppressSingleImageCaptions(_ converter.Context, doc *html.Node) {
	for _, fig := range findAll(doc, "figure") {
		if len(findAll(fig, "img")) != 1 {
			continue
		}
		var caps []*html.Node
		for c := fig.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "figcaption" {
				caps = append(caps, c)
			}
		}
		for _, c := range caps {
			dom.RemoveNode(c)
		}
	}
}

// collapseBlankLines turns whitespace-only lines into empty lines. The
// commonmark list renderer indents blank lines inside nested list items;
// collapsing them keeps the terminal output clean. Content-bearing lines keep
// their trailing whitespace so <br> hard breaks ("  \n") survive.
func collapseBlankLines(_ converter.Context, content []byte) []byte {
	lines := bytes.Split(content, []byte("\n"))
	for i, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			lines[i] = nil
		}
	}
	return bytes.Join(lines, []byte("\n"))
}

// renderImage writes <img> as [alt], or [image] when the image has no alt
// text.
func renderImage(_ converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
	alt, ok := dom.GetAttribute(n, "alt")
	if !ok || strings.TrimSpace(alt) == "" {
		w.WriteString("[image]")
		return converter.RenderSuccess
	}
	w.WriteString("[" + textutil.EscapeMarkdown(strings.TrimSpace(alt)) + "]")
	return converter.RenderSuccess
}

// plainText extracts the visible text from an HTML document. It is the graceful
// fallback used when html-to-markdown conversion fails, so the reader shows
// readable text with markup removed rather than the raw HTML source. Block
// elements are separated by spaces.
func plainText(htmlStr string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return strings.TrimSpace(htmlStr)
	}
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
		if n.Type == html.ElementNode && isBlockElement(n.Data) {
			b.WriteByte(' ')
		}
	}
	walk(doc)
	return strings.TrimSpace(b.String())
}

// isBlockElement reports whether a tag separates its content from surrounding
// text with whitespace when rendered as plain text.
func isBlockElement(tag string) bool {
	switch tag {
	case "p", "div", "br", "li", "ul", "ol", "blockquote", "h1", "h2", "h3",
		"h4", "h5", "h6", "table", "tr", "section", "article", "header",
		"footer", "body", "html", "hr":
		return true
	}
	return false
}
