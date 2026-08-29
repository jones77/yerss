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

// newConverter builds the shared html-to-markdown converter. Escaping is
// disabled because the output is fed to a markdown renderer (glamour) and
// never re-parsed for display; smart escaping would only add backslash noise.
func newConverter() *converter.Converter {
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
	conv.Register.RendererFor("img", converter.TagTypeInline, renderImage, converter.PriorityEarly)
	return conv
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
	w.WriteString("[" + escapeAlt(strings.TrimSpace(alt)) + "]")
	return converter.RenderSuccess
}

// escapeAlt backslash-escapes the markdown-significant characters in alt text so
// that image alt text renders literally inside the [alt] placeholder instead of
// being interpreted as styling or link syntax by the downstream renderer.
func escapeAlt(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\', '`', '*', '_', '[', ']', '<', '>':
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
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