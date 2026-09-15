package compose

import (
	"strings"

	xansi "github.com/charmbracelet/x/ansi"
)

// urlBreakpoints are the characters at which a long URL may wrap mid-word, so
// a URL glamour pushed onto its own line re-flows inline after the link name.
const urlBreakpoints = "/-_."

// InlineLinkURLs re-wraps paragraphs in which glamour pushed a link's URL onto
// its own line, so the URL starts on the name's line and wraps at the content
// width like any other long word. It is the expanded-mode counterpart to
// ReflowShortLines: compact mode strips the URL away, expanded mode keeps it,
// and a URL that did not fit next to its name must still read inline instead
// of leaving the name line half empty. Only plain paragraphs are re-wrapped —
// blockquotes, list items, and indented lines keep glamour's exact layout —
// and paragraphs without a wrapped URL pass through byte-for-byte untouched.
// Every wrapped line of a re-flowed URL re-emits its own OSC 8 hyperlink,
// matching the app's convention that a hyperlink must not span a line break.
func InlineLinkURLs(rendered string, contentW int) string {
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
		out = append(out, inlineParagraph(para, contentW)...)
		i = j
	}
	return strings.Join(out, "\n")
}

// inlineParagraph re-wraps one paragraph when it contains a pushed-to-its-own-
// line URL, returning the paragraph unchanged otherwise.
func inlineParagraph(para []string, contentW int) []string {
	first := xansi.Strip(para[0])
	if strings.HasPrefix(first, "│") || strings.HasPrefix(first, " ") || strings.HasPrefix(first, "\t") {
		return para
	}
	if listItemLineRe.MatchString(first) {
		return para
	}
	if !hasWrappedHref(strings.Join(para, "\n")) {
		return para
	}
	return wrapFlow(flattenParagraph(para), contentW)
}

// hasWrappedHref reports whether any href unit in the paragraph begins on a
// line of its own — i.e. glamour pushed the URL onto the next line instead of
// flowing it after the name.
func hasWrappedHref(para string) bool {
	segs := ScanANSI(para)
	order, byTarget := collectLinkSpans(para, segs)
	units := groupLinkUnits(para, segs, order, byTarget)
	for _, u := range order {
		for _, un := range units[u] {
			if !un.drop {
				continue
			}
			first := byTarget[u][un.first]
			start := 0
			if idx := strings.LastIndexByte(para[:first.openStart], '\n'); idx >= 0 {
				start = idx + 1
			}
			if strings.TrimSpace(textBetween(para, segs, start, first.openStart)) == "" {
				return true
			}
		}
	}
	return false
}

// flowChar is one logical display cell of a paragraph's re-flowable text,
// remembering the OSC 8 target, whether that target belongs to a link's name
// or its visible href, and the active SGR escape that styled it, so the
// re-wrap can re-emit both on every line.
type flowChar struct {
	r    rune
	w    int
	url  string
	href bool
	sgr  string
}

// flattenParagraph converts the paragraph's styled, per-line-OSC 8 lines into
// one stream of logical cells, dropping glamour's line-fill padding and the
// wrap newlines. Each original line boundary becomes a separator inserted
// before the next line's first cell: none when the previous line's last cell
// and the next line's first cell belong to the same link part (a name or URL
// that wraps across the line break continues without a word boundary), a
// single space otherwise.
func flattenParagraph(para []string) []flowChar {
	trimmed := make([]string, len(para))
	for i, l := range para {
		trimmed[i] = trimStyledWS(l)
	}
	joined := strings.Join(trimmed, "\n")
	segs := ScanANSI(joined)
	order, byTarget := collectLinkSpans(joined, segs)
	units := groupLinkUnits(joined, segs, order, byTarget)
	hrefSeg := map[int]bool{}
	for _, u := range order {
		for _, un := range units[u] {
			for k := un.first; k <= un.last; k++ {
				hrefSeg[byTarget[u][k].openSeg] = un.drop
			}
		}
	}
	var out []flowChar
	var prevLast *flowChar
	curURL := ""
	curHref := false
	curSGR := ""
	pending := false
	for i := range segs {
		seg := segs[i]
		switch seg.Kind {
		case SegOSC:
			payload, _ := seg.OSC(joined)
			if !strings.HasPrefix(payload, "8;") {
				continue
			}
			rest := payload[2:]
			if idx := strings.LastIndex(rest, ";"); idx >= 0 {
				rest = rest[idx+1:]
			}
			if rest == "" {
				curURL, curHref = "", false
			} else {
				curURL = rest
				curHref = hrefSeg[i]
			}
		case SegEscape:
			esc := joined[seg.Start:seg.End]
			if esc == "\x1b[m" || esc == "\x1b[0m" {
				curSGR = ""
			} else if strings.HasPrefix(esc, "\x1b[") && strings.HasSuffix(esc, "m") {
				curSGR = esc
			}
		case SegText:
			for _, r := range joined[seg.Start:seg.End] {
				if r == '\n' {
					if len(out) > 0 {
						pending = true
						prevLast = &out[len(out)-1]
					}
					continue
				}
				if pending {
					pending = false
					if prevLast == nil || prevLast.url == "" || prevLast.url != curURL || prevLast.href != curHref {
						out = append(out, flowChar{r: ' ', w: 1})
					}
				}
				out = append(out, flowChar{r: r, w: CellWidth(r), url: curURL, href: curHref, sgr: curSGR})
			}
		}
	}
	return out
}

// wrapFlow packs the paragraph's logical cells into lines no wider than
// contentW, breaking after spaces (word boundaries) and after URL breakpoint
// characters so a long URL wraps instead of overflowing, and re-emits each
// line as self-contained styled text with per-line OSC 8 hyperlinks.
func wrapFlow(chars []flowChar, contentW int) []string {
	var lines [][]flowChar
	line := []flowChar{}
	lineW := 0
	breakIdx := -1
	breakIsSpace := false
	for _, c := range chars {
		if lineW > 0 && lineW+c.w > contentW && breakIdx >= 0 {
			if breakIsSpace {
				lines = append(lines, line[:breakIdx])
			} else {
				lines = append(lines, line[:breakIdx+1])
			}
			line = line[breakIdx+1:]
			lineW = lineWidth(line)
			breakIdx = -1
		}
		if c.r == ' ' && len(line) == 0 {
			continue
		}
		line = append(line, c)
		lineW += c.w
		if (c.r == ' ' || strings.ContainsRune(urlBreakpoints, c.r)) && lineW <= contentW {
			breakIdx = len(line) - 1
			breakIsSpace = c.r == ' '
		}
		if lineW > contentW && breakIdx < 0 {
			if len(line) > 1 {
				lines = append(lines, line[:len(line)-1])
				line = line[len(line)-1:]
				lineW = lineWidth(line)
			} else {
				lines = append(lines, line)
				line = nil
				lineW = 0
			}
		}
	}
	if len(line) > 0 {
		lines = append(lines, line)
	}
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		out = append(out, renderFlowLine(l))
	}
	return out
}

// lineWidth sums the display width of the line's cells.
func lineWidth(line []flowChar) int {
	w := 0
	for _, c := range line {
		w += c.w
	}
	return w
}

// renderFlowLine emits one wrapped line: consecutive cells sharing a style and
// hyperlink form one run, each run wrapped in its OSC 8 target (if any) and
// its SGR, so a hyperlink and its styling never span a line break. Trailing
// spaces are dropped (the viewport pads at render time).
func renderFlowLine(chars []flowChar) string {
	for len(chars) > 0 && chars[len(chars)-1].r == ' ' {
		chars = chars[:len(chars)-1]
	}
	var b strings.Builder
	i := 0
	for i < len(chars) {
		j := i
		for j < len(chars) && chars[j].url == chars[i].url && chars[j].sgr == chars[i].sgr {
			j++
		}
		g := chars[i:j]
		if g[0].url != "" {
			b.WriteString("\x1b]8;;")
			b.WriteString(g[0].url)
			b.WriteByte(0x07)
		}
		if g[0].sgr != "" {
			b.WriteString(g[0].sgr)
		}
		for _, c := range g {
			b.WriteRune(c.r)
		}
		if g[0].sgr != "" {
			b.WriteString("\x1b[m")
		}
		if g[0].url != "" {
			b.WriteString("\x1b]8;;\a")
		}
		i = j
	}
	return b.String()
}
