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
// breaks. Images are rendered as [alt] (or [image] when no alt text is
// available). Links are rendered as markdown `[text](url)` with the link text
// wrapped in OSC 8 hyperlink sequences so it is clickable, and the URL kept as
// visible plain fallback text.
func HTMLToText(html string) string {
	if strings.TrimSpace(html) == "" {
		return ""
	}
	html = replaceImages(html)
	html = wrapLinks(html)
	text, err := html2text.FromString(html, html2text.Options{
		PrettyTables:        false,
		PrettyTablesOptions: nil,
		OmitLinks:           true,
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

// wrapLinks renders each <a href="url">...</a> element as a markdown link:
// the link text is wrapped in OSC 8 hyperlink escape sequences carrying the
// URL (so it is clickable) and the URL follows as visible plain fallback text.
func wrapLinks(html string) string {
	return linkTagRe.ReplaceAllStringFunc(html, func(m string) string {
		sm := linkTagRe.FindStringSubmatch(m)
		url := strings.TrimSpace(sm[2])
		if url == "" {
			return m
		}
		return ansi.SetHyperlink(url) + "[" + sm[3] + "]" + ansi.ResetHyperlink() + "(" + url + ")"
	})
}

// wrapText soft-wraps text to the given display width on word boundaries.
// Markdown links are link-aware: a link always ends its line (any following
// text starts on a new line); when a link's `[text](url)` form does not fit on
// the current line, the `(url)` breaks onto a new line, and a URL too long for
// a single line is truncated with the ellipsis glyph.
func wrapText(text string, width int, ellipsis string) string {
	if width <= 0 {
		return text
	}
	var out strings.Builder
	linkBoundary := "]" + ansi.ResetHyperlink() + "("
	for _, para := range strings.Split(text, "\n") {
		cur := ""
		curW := 0
		forceBreak := false
		tokens := linkTokens(para)
		for i := 0; i < len(tokens); i++ {
			word := tokens[i]
			if forceBreak {
				out.WriteString(cur)
				out.WriteString("\n")
				cur, curW = "", 0
				forceBreak = false
			}
			if j := strings.Index(word, linkBoundary); j >= 0 {
				head := word[:j+1+len(ansi.ResetHyperlink())]
				tail := word[j+1+len(ansi.ResetHyperlink()):]
				hw := ansi.StringWidth(head)
				tw := ansi.StringWidth(tail)
				// Keep trailing punctuation on the same line as the link,
				// truncating the URL with an ellipsis to make room for it.
				punct := ""
				for i+1 < len(tokens) && isTrailingPunct(tokens[i+1]) {
					punct += tokens[i+1]
					i++
				}
				pw := ansi.StringWidth(punct)
				full := hw + tw + pw
				if (cur == "" && full <= width) || (cur != "" && curW+full+1 <= width) {
					if cur != "" {
						cur += " "
						curW++
					}
					cur += head + tail + punct
					curW += full
				} else {
					// The link does not fit on the current line; flush it and
					// keep it together on a fresh line when it fits there,
					// otherwise break [text] and (url) apart.
					if cur != "" {
						out.WriteString(cur)
						out.WriteString("\n")
					}
					if full <= width {
						cur = head + tail + punct
						curW = full
					} else {
						out.WriteString(head)
						out.WriteString("\n")
						avail := width - pw
						if avail < 1 {
							avail = 1
						}
						cur = truncateURL(tail, avail, ellipsis) + punct
						curW = ansi.StringWidth(cur)
					}
				}
				forceBreak = true
				continue
			}
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

// isTrailingPunct reports whether s is punctuation that belongs on the same
// line as a preceding markdown link.
func isTrailingPunct(s string) bool {
	for _, r := range s {
		switch r {
		case ',', '.', ';', ':', '!', '?', ')', '"', '\'':
		default:
			return false
		}
	}
	return len(s) > 0
}

// linkOpenSeq is the start of an OSC 8 hyperlink escape sequence.
const linkOpenSeq = "\x1b]8;"

// linkTokens splits a paragraph into wrap tokens, treating each markdown link
// (with its OSC 8 wrapping) as a single token so spaces inside the link text
// are not break points. Plain text outside links is split on whitespace. The
// link still wraps as `[text]` then `(url)` via wrapText's link handling.
func linkTokens(para string) []string {
	var tokens []string
	i := 0
	for i < len(para) {
		rel := strings.Index(para[i:], linkOpenSeq)
		if rel < 0 {
			tokens = append(tokens, strings.Fields(para[i:])...)
			break
		}
		start := i + rel
		tokens = append(tokens, strings.Fields(para[i:start])...)
		// The link is OSC-open + [text] + OSC-reset + (url); find the closing
		// paren after the "](url)" boundary so the whole link is one token.
		boundary := strings.Index(para[start:], "]" + ansi.ResetHyperlink() + "(")
		if boundary < 0 {
			tokens = append(tokens, para[start:])
			break
		}
		urlStart := start + boundary + 1 + len(ansi.ResetHyperlink())
		end := strings.IndexByte(para[urlStart:], ')')
		if end < 0 {
			tokens = append(tokens, para[start:])
			break
		}
		linkEnd := urlStart + end + 1
		tokens = append(tokens, para[start:linkEnd])
		i = linkEnd
	}
	return tokens
}

// truncateURL truncates a markdown URL tail `(url)` so it fits width display
// columns, keeping the parentheses and cutting an overlong inner URL with the
// ellipsis glyph.
func truncateURL(tail string, width int, ellipsis string) string {
	if ansi.StringWidth(tail) <= width {
		return tail
	}
	inner := tail[1 : len(tail)-1]
	avail := width - 2 - ansi.StringWidth(ellipsis)
	if avail < 1 {
		avail = 1
	}
	return "(" + ansi.Truncate(inner, avail, ellipsis) + ")"
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