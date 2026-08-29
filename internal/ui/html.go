package ui

import (
	"regexp"
	"strings"

	"github.com/jaytaylor/html2text"
	"github.com/mattn/go-runewidth"
)

var (
	imgTagRe = regexp.MustCompile(`(?i)<img[^>]*>`)
	imgAltRe = regexp.MustCompile(`(?i)\balt\s*=\s*"([^"]*)"`)
)

// HTMLToText converts article HTML to plain text, preserving paragraph
// breaks and link URLs. Images are rendered as [alt] (or [image] when no alt
// text is available).
func HTMLToText(html string) string {
	if strings.TrimSpace(html) == "" {
		return ""
	}
	html = replaceImages(html)
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
			w := runewidth.StringWidth(word)
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

// truncate truncates s to at most width display columns.
func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= width {
		return s
	}
	var b strings.Builder
	w := 0
	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if w+rw > width {
			break
		}
		b.WriteRune(r)
		w += rw
	}
	return b.String()
}

// padRight pads s to at least width display columns.
func padRight(s string, width int) string {
	gap := width - runewidth.StringWidth(s)
	if gap <= 0 {
		return s
	}
	return s + strings.Repeat(" ", gap)
}