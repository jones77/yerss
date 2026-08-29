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
		return html
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
	w.WriteString("[" + strings.TrimSpace(alt) + "]")
	return converter.RenderSuccess
}