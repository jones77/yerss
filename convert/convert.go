// Package convert converts article HTML to a markdown display string with
// embedded terminal escape sequences for inline formatting.
package convert

import (
	"bytes"
	"strings"

	"github.com/JohannesKaufmann/dom"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/net/html"
)

var conv = newConverter()

// Convert converts article HTML to a markdown display string with embedded
// terminal escapes: ANSI bold for <strong>/<b>, OSC 8 hyperlinks for <a> link
// text, and [alt]/[image] placeholders for <img>. Structured HTML (lists,
// blockquotes, code blocks, tables) renders as native commonmark markdown.
func Convert(html string) string {
	if strings.TrimSpace(html) == "" {
		return ""
	}
	out, err := conv.ConvertString(html)
	if err != nil {
		return html
	}
	return out
}

// newConverter builds the shared html-to-markdown converter with the custom
// renderers that emit terminal escapes. Escaping is disabled because the
// output is displayed directly in the terminal and never re-parsed as
// markdown; smart escaping would only add backslash noise.
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
	conv.Register.RendererFor("strong", converter.TagTypeInline, renderBold, converter.PriorityEarly)
	conv.Register.RendererFor("b", converter.TagTypeInline, renderBold, converter.PriorityEarly)
	conv.Register.RendererFor("a", converter.TagTypeInline, renderLink, converter.PriorityEarly)
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

// renderBold writes <strong>/<b> content wrapped in ANSI bold escapes so it
// renders bold in the terminal instead of as literal markdown asterisks.
func renderBold(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
	w.WriteString("\x1b[1m")
	ctx.RenderChildNodes(ctx, w, n)
	w.WriteString("\x1b[22m")
	return converter.RenderSuccess
}

// renderLink writes <a> content as an OSC 8 clickable markdown link
// [text](url): the link text is wrapped in OSC 8 hyperlink escapes carrying
// the URL, and the URL follows as visible plain fallback text. A link without
// an href renders its children as plain text.
func renderLink(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
	url := strings.TrimSpace(dom.GetAttributeOr(n, "href", ""))
	if url == "" {
		ctx.RenderChildNodes(ctx, w, n)
		return converter.RenderSuccess
	}
	w.WriteString(ansi.SetHyperlink(url))
	w.WriteRune('[')
	ctx.RenderChildNodes(ctx, w, n)
	w.WriteRune(']')
	w.WriteString(ansi.ResetHyperlink())
	w.WriteRune('(')
	w.WriteString(url)
	w.WriteRune(')')
	return converter.RenderSuccess
}

// renderImage writes <img> as [alt], or [image] when the image has no alt
// text.
func renderImage(_ converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
	alt, ok := dom.GetAttribute(n, "alt")
	if !ok || strings.TrimSpace(alt) == "" {
		w.WriteString("[image]")
		return converter.RenderSuccess
	}
	w.WriteString("[" + strings.TrimSpace(alt) + "]")
	return converter.RenderSuccess
}