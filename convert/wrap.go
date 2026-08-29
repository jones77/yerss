package convert

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"
)

// WrapText soft-wraps text to the given display width on word boundaries.
// Markdown links are link-aware: a link always ends its line (any following
// text starts on a new line); when a link's `[text](url)` form does not fit on
// the current line, the `(url)` breaks onto a new line, and a URL too long for
// a single line is truncated with the ellipsis glyph. Leading indentation on a
// paragraph (e.g. nested list items) is preserved and re-applied to every
// wrapped line, since markdown structure indents with leading spaces.
func WrapText(text string, width int, ellipsis string) string {
	if width <= 0 {
		return text
	}
	var out strings.Builder
	linkBoundary := "]" + ansi.ResetHyperlink() + "("
	for _, para := range strings.Split(text, "\n") {
		indent := para[:len(para)-len(strings.TrimLeft(para, " \t"))]
		effW := width - ansi.StringWidth(indent)
		if effW < 1 {
			effW = 1
		}
		emit := func(s string) {
			out.WriteString(indent)
			out.WriteString(s)
			out.WriteString("\n")
		}
		cur := ""
		curW := 0
		forceBreak := false
		tokens := linkTokens(para)
		for i := 0; i < len(tokens); i++ {
			word := tokens[i]
			if forceBreak {
				emit(cur)
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
				if (cur == "" && full <= effW) || (cur != "" && curW+full+1 <= effW) {
					if cur != "" {
						cur += " "
						curW++
					}
					cur += head + tail + punct
					curW += full
				} else {
					// The link does not fit on the current line: [text] trails
					// the current line when it fits there, otherwise it goes on
					// its own line, and (url) follows on the next line.
					if cur != "" && curW+hw+1 <= effW {
						cur += " " + head
						emit(cur)
					} else {
						if cur != "" {
							emit(cur)
						}
						emit(head)
					}
					avail := effW - pw
					if avail < 1 {
						avail = 1
					}
					cur = truncateURL(tail, avail, ellipsis) + punct
					curW = ansi.StringWidth(cur)
				}
				forceBreak = true
				continue
			}
			w := ansi.StringWidth(word)
			if cur != "" && curW+w+1 > effW {
				emit(cur)
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
		emit(cur)
	}
	return strings.TrimSuffix(out.String(), "\n")
}

// isTrailingPunct reports whether s is punctuation that belongs on the same
// line as a preceding markdown link. It covers ASCII and Unicode punctuation
// (curly quotes, ellipsis, dashes), so tokens like `,”` stay with the link.
func isTrailingPunct(s string) bool {
	for _, r := range s {
		if !unicode.IsPunct(r) {
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
// link still wraps as `[text]` then `(url)` via WrapText's link handling.
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
		boundary := strings.Index(para[start:], "]"+ansi.ResetHyperlink()+"(")
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