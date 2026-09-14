package compose

import (
	"strings"
	"unicode"

	xansi "github.com/charmbracelet/x/ansi"
)

// ReflowShortLines re-wraps ordinary paragraphs that contain a line narrower
// than the content width, after compact link mode has stripped link hrefs.
// Glamour pads every paragraph line to the full content width, so a sub-width
// line in a plain paragraph is a stripping artifact: StripLinkHrefs deletes a
// wrapped URL's href spans whole, swallowing the line's fill padding and a wrap
// newline, and glues the following text onto the name line. Re-flowing such a
// paragraph at the content width restores the full-width layout. Blockquotes,
// list items, and indented lines are left untouched (re-flowing would destroy
// their prefixes). The viewport pads lines at render time, so no explicit
// padding is added here.
func ReflowShortLines(rendered string, contentW int) string {
	lines := strings.Split(rendered, "\n")
	var out []string
	for i := 0; i < len(lines); {
		if blankLine(lines[i]) {
			out = append(out, lines[i])
			i++
			continue
		}
		j := i + 1
		for j < len(lines) && !blankLine(lines[j]) {
			j++
		}
		para := lines[i:j]
		if reflowablePara(para, contentW) {
			out = append(out, reflowParagraph(para, contentW)...)
		} else {
			out = append(out, para...)
		}
		i = j
	}
	return strings.Join(out, "\n")
}

// blankLine reports whether the line has no visible text.
func blankLine(line string) bool {
	return strings.TrimSpace(xansi.Strip(line)) == ""
}

// reflowablePara reports whether the paragraph should be re-flowed: it is a
// plain paragraph (not a blockquote, list item, or indented continuation) and
// contains a line whose raw width is narrower than the content width.
func reflowablePara(para []string, contentW int) bool {
	first := xansi.Strip(para[0])
	if strings.HasPrefix(first, "│") || strings.HasPrefix(first, " ") || strings.HasPrefix(first, "\t") {
		return false
	}
	if listItemLineRe.MatchString(first) {
		return false
	}
	for _, line := range para {
		if xansi.StringWidth(line) < contentW {
			return true
		}
	}
	return false
}

// reflowParagraph joins the paragraph's lines (with trailing styled fill
// padding removed) into one flow of text and re-wraps it at the content width.
func reflowParagraph(para []string, contentW int) []string {
	var b strings.Builder
	for i, line := range para {
		trimmed := trimStyledWS(line)
		if i > 0 {
			if trailingPunctRe.MatchString(strings.TrimSpace(xansi.Strip(trimmed))) {
				b.WriteString(trimmed)
				continue
			}
			b.WriteByte(' ')
		}
		b.WriteString(trimmed)
	}
	flow := xansi.Wrap(b.String(), contentW, "")
	if flow == "" {
		return para
	}
	return strings.Split(flow, "\n")
}

// trimStyledWS removes the styled fill spaces glamour appends to pad each line
// to the content width and any leading whitespace left where a hidden link href
// was stripped, keeping every escape sequence around the visible text so OSC 8
// hyperlinks and style resets stay balanced: escapes before the first visible
// rune (the style/link state active at that point) and after the last visible
// rune (resets and link closes) are both preserved.
func trimStyledWS(line string) string {
	segs := ScanANSI(line)
	firstNonWS, lastNonWS := -1, 0
	for _, seg := range segs {
		if seg.Kind != SegText {
			continue
		}
		t := line[seg.Start:seg.End]
		if i := strings.IndexFunc(t, func(r rune) bool { return !unicode.IsSpace(r) }); i >= 0 {
			if firstNonWS < 0 {
				firstNonWS = seg.Start + i
			}
			lastNonWS = seg.End
		}
	}
	if firstNonWS < 0 {
		return ""
	}
	var b strings.Builder
	for _, seg := range segs {
		if seg.Kind != SegText && seg.End <= firstNonWS {
			b.WriteString(line[seg.Start:seg.End])
		}
	}
	b.WriteString(line[firstNonWS:lastNonWS])
	for _, seg := range segs {
		if seg.Kind != SegText && seg.Start >= lastNonWS {
			b.WriteString(line[seg.Start:seg.End])
		}
	}
	return b.String()
}