package ui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/jaytaylor/html2text"
)

var (
	imgTagRe = regexp.MustCompile(`(?i)<img[^>]*>`)
	imgAltRe = regexp.MustCompile(`(?i)\balt\s*=\s*"([^"]*)"`)
	linkTagRe = regexp.MustCompile(`(?is)(<a\b[^>]*\bhref\s*=\s*["']([^"']*)[^>]*>)(.*?)(</a>)`)
)

// HTMLToText converts article HTML to plain text, preserving paragraph
// breaks and link URLs. Images are rendered as [alt] (or [image] when no alt
// text is available). Links are wrapped in OSC 8 hyperlink sequences so the
// link text is clickable; the anchor tag is kept so html2text still appends
// the URL as visible fallback text.
func HTMLToText(html string) string {
	if strings.TrimSpace(html) == "" {
		return ""
	}
	html = replaceImages(html)
	html = wrapLinks(html)
	text, err := html2text.FromString(html, html2text.Options{
		PrettyTables:        false,
		PrettyTablesOptions: nil,
	})
	if err != nil {
		return html
	}
	return text
}

// replaceImages substitutes <img> tags with [alt] or [image] placeholders,
// since html2text drops images entirely.
func replaceImages(html string) string {
	return imgTagRe.ReplaceAllStringFunc(html, func(tag string) string {
		m := imgAltRe.FindStringSubmatch(tag)
		if len(m) == 2 && strings.TrimSpace(m[1]) != "" {
			return "[" + strings.TrimSpace(m[1]) + "]"
		}
		return "[image]"
	})
}

// wrapLinks wraps the text of each <a href="url">...</a> element in OSC 8
// hyperlink escape sequences carrying the URL, making the link text clickable
// in terminals that support OSC 8. The anchor tag itself is preserved so
// html2text still appends the URL as plain fallback text.
func wrapLinks(html string) string {
	return linkTagRe.ReplaceAllStringFunc(html, func(m string) string {
		sm := linkTagRe.FindStringSubmatch(m)
		url := strings.TrimSpace(sm[2])
		if url == "" {
			return m
		}
		return sm[1] + ansi.SetHyperlink(url) + sm[3] + ansi.ResetHyperlink() + sm[4]
	})
}

// wrapText soft-wraps text to the given display width on word boundaries.
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	var out strings.Builder
	for _, para := range strings.Split(text, "\n") {
		cur := ""
		curW := 0
		for _, word := range strings.Fields(para) {
			w := ansi.StringWidth(word)
			if cur != "" && curW+w+1 > width {
				out.WriteString(cur)
				out.WriteString("\n")
				cur = ""
				curW = 0
			}
			if cur != "" {
				cur += " "
				curW++
			}
			cur += word
			curW += w
		}
		out.WriteString(cur)
		out.WriteString("\n")
	}
	return strings.TrimSuffix(out.String(), "\n")
}

// truncate truncates s to at most width display columns, keeping ANSI escape
// sequences intact.
func truncate(s string, width int) string {
	return ansi.Truncate(s, width, "")
}

// padRight pads s to at least width display columns.
func padRight(s string, width int) string {
	gap := width - ansi.StringWidth(s)
	if gap <= 0 {
		return s
	}
	return s + strings.Repeat(" ", gap)
}