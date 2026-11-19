package compose

import (
	"sort"
	"strings"
)

// linkSpan is one OSC 8 hyperlink span of a target: the segment index and byte
// offset of its open, the byte offset just past its closing OSC 8 (or the end
// of the string when unterminated), the visible text it wraps, and whether a
// close was found before the next open.
type linkSpan struct {
	openSeg    int
	openStart  int
	closeEnd   int
	text       string
	terminated bool
}

// StripLinkHrefs removes the visible href span glamour appends to each body
// link (`text url` → `text`), leaving the OSC 8 name span(s) intact so the
// link stays clickable. Glamour always emits a link as a name span (OSC 8
// open + styled text + close) followed by a styled space and an href span
// (OSC 8 open + URL + close) carrying the same OSC 8 target. A link part that
// wraps across lines becomes several same-target spans (glamour re-emits the
// hyperlink after every wrapped newline), so StripLinkHrefs first groups each
// link part's spans into a logical unit — gluing same-target spans together
// only at wrap boundaries, where the text between the previous close and the
// next open is just glamour's styled line padding — and then keeps the first
// unit per target (the name) while dropping each following unit (the href)
// whole: its leading space and every wrapped continuation line.
func StripLinkHrefs(s string) string {
	segs := ScanANSI(s)
	order, byTarget := collectLinkSpans(s, segs)
	units := groupLinkUnits(s, segs, order, byTarget)

	var drops [][2]int
	for _, u := range order {
		for _, un := range units[u] {
			if !un.drop {
				continue
			}
			first, last := byTarget[u][un.first], byTarget[u][un.last]
			// Extend the drop back over the href unit's leading space: the
			// nearest preceding text runs, provided each is pure whitespace on
			// a single line (glamour's styled separator and line-fill padding).
			// A run carrying a newline is a wrap boundary, so a wrapped href
			// starts at its OSC 8 open; a run with real text ends the walk.
			dropStart := first.openStart
			for k := first.openSeg - 1; k >= 0; k-- {
				if segs[k].Kind != SegText {
					continue
				}
				t := s[segs[k].Start:segs[k].End]
				if !strings.Contains(t, "\n") && strings.TrimSpace(t) == "" {
					dropStart = segs[k].Start
					continue
				}
				break
			}
			// The drop runs through the href unit's closing OSC 8 (or to the
			// end of the string when unterminated), and — when nothing but
			// styled whitespace separates the close from the line's newline —
			// through that fill padding and the wrap newline too. Without the
			// extension a wrapped href that ends mid-line leaves a
			// whitespace-only residual line that later splits the paragraph.
			// Real text after the close on the same line (a same-line href
			// followed by more words) must survive, so the extension is
			// limited to the first newline.
			dropEnd := last.closeEnd
			if dropEnd < len(s) {
				end := len(s)
				if nl := strings.IndexByte(s[dropEnd:], '\n'); nl >= 0 {
					end = dropEnd + nl
				}
				if strings.TrimSpace(textBetween(s, segs, dropEnd, end)) == "" {
					dropEnd = end
					// Swallow the wrap newline only when the href is the sole
					// non-whitespace content on its opening line. When real text
					// precedes the dropped href on that line ("... claimed"), the
					// newline is the only word separator left, so it must stay and
					// let the re-wrap re-join the two with a single space.
					lineStart := strings.LastIndexByte(s[:dropStart], '\n') + 1
					if strings.TrimSpace(textBetween(s, segs, lineStart, dropStart)) == "" && dropEnd < len(s) {
						dropEnd++ // swallow the wrap newline
					}
				}
			}
			drops = append(drops, [2]int{dropStart, dropEnd})
		}
	}

	if len(drops) == 0 {
		return s
	}
	sort.Slice(drops, func(a, b int) bool { return drops[a][0] < drops[b][0] })
	var out strings.Builder
	pos := 0
	for _, d := range drops {
		if d[1] <= d[0] || d[0] < pos {
			continue
		}
		out.WriteString(s[pos:d[0]])
		pos = d[1]
	}
	out.WriteString(s[pos:])
	return out.String()
}

// collectLinkSpans returns the distinct OSC 8 hyperlink targets in occurrence
// order and one span per open, grouped by target. A span covers the segment
// index and byte offset of its open, the byte offset just past its closing
// OSC 8 (or the end of the string when unterminated), and the visible text it
// wraps.
func collectLinkSpans(s string, segs []ScanSeg) ([]string, map[string][]linkSpan) {
	var order []string
	byTarget := map[string][]linkSpan{}
	for i := 0; i < len(segs); i++ {
		u, ok := oscLinkTarget(s, segs[i])
		if !ok || u == "" {
			continue
		}
		sp := linkSpan{openSeg: i, openStart: segs[i].Start, closeEnd: len(s)}
		for j := i + 1; j < len(segs); j++ {
			if segs[j].Kind == SegText {
				sp.text += s[segs[j].Start:segs[j].End]
				continue
			}
			if segs[j].Kind != SegOSC {
				continue
			}
			if next, isLink := oscLinkTarget(s, segs[j]); isLink && next != "" {
				// Another open arrives before a close: glamour re-emits the
				// hyperlink across a wrap (for example inside a blockquote)
				// without closing the previous line. Mark the span unterminated
				// and bound it before the next escape so offsets stay ordered.
				sp.closeEnd = segs[j].Start
				sp.terminated = false
			} else {
				sp.closeEnd = segs[j].End
				sp.terminated = true
			}
			break
		}
		if _, ok := byTarget[u]; !ok {
			order = append(order, u)
		}
		byTarget[u] = append(byTarget[u], sp)
	}
	return order, byTarget
}

// oscLinkTarget returns the OSC 8 URI carried by seg and whether seg is an
// OSC 8 sequence at all. A close sequence (empty URI) returns ("", true).
func oscLinkTarget(s string, seg ScanSeg) (string, bool) {
	if seg.Kind != SegOSC {
		return "", false
	}
	payload, _ := seg.OSC(s)
	if !strings.HasPrefix(payload, "8;") {
		return "", false
	}
	rest := payload[2:]
	u := rest
	if idx := strings.LastIndex(rest, ";"); idx >= 0 {
		u = rest[idx+1:]
	}
	return u, true
}

// linkUnit groups a target's spans into one logical link part: within a unit,
// consecutive spans are wrapped continuation lines (they render as one link
// part); a real space between spans — glamour's href prefix on the same line,
// or text like " and " between two links — ends the unit. Units alternate
// name, href, name, href … per target; the drop flag marks the href units.
type linkUnit struct {
	drop        bool
	first, last int
}

// groupLinkUnits partitions each target's spans into name/href units in
// occurrence order.
func groupLinkUnits(s string, segs []ScanSeg, order []string, byTarget map[string][]linkSpan) map[string][]linkUnit {
	units := make(map[string][]linkUnit, len(order))
	for _, u := range order {
		spans := byTarget[u]
		us := []linkUnit{{first: 0, last: 0}}
		for k := 1; k < len(spans); k++ {
			if joinsUnit(s, segs, spans[k-1], spans[k], u, us[len(us)-1].drop) {
				us[len(us)-1].last = k
			} else {
				us = append(us, linkUnit{drop: !us[len(us)-1].drop, first: k, last: k})
			}
		}
		units[u] = us
	}
	return units
}

// joinsUnit reports whether span cur continues the same link part as prev
// toward the same target. A same-line separation always starts a new part,
// because a wrapped continuation line implies a newline in between. Across a
// newline, glamour pads the previous line with styled spaces up to the wrap
// width and re-emits the hyperlink, so a continuation is only recognizable
// when every space in between is that padding (glamour's mid-word wrap always
// resumes with text, never a real space). A real space — the text between two
// links, or a href pushed to its own line before the name's line is padded —
// starts a new part; the href pushed to the next line is identified by its
// text starting at the URL target's beginning. Within an already-open href
// part every same-padding line resumes the URL.
func joinsUnit(s string, segs []ScanSeg, prev, cur linkSpan, target string, prevDrop bool) bool {
	// A span with no close before the next open is one continuous hyperlink
	// that glamour re-emits across a wrap (it may interleave the line's
	// indentation, such as a blockquote border), so the next same-target open
	// is a continuation regardless of the text between them.
	if !prev.terminated {
		return true
	}
	if cur.openStart < prev.closeEnd {
		return false
	}
	if !strings.Contains(s[prev.closeEnd:cur.openStart], "\n") {
		return false
	}
	vis := textBetween(s, segs, prev.closeEnd, cur.openStart)
	if !allSpacesAreLineFill(vis) {
		return false
	}
	if prevDrop {
		return true
	}
	return !strings.HasPrefix(target, strings.TrimSpace(cur.text))
}

// textBetween returns the text runs of s within [lo, hi), excluding escapes.
func textBetween(s string, segs []ScanSeg, lo, hi int) string {
	var b strings.Builder
	for _, seg := range segs {
		if seg.End <= lo {
			continue
		}
		if seg.Start >= hi {
			break
		}
		if seg.Kind != SegText {
			continue
		}
		a, z := seg.Start, seg.End
		if a < lo {
			a = lo
		}
		if z > hi {
			z = hi
		}
		b.WriteString(s[a:z])
	}
	return b.String()
}

// allSpacesAreLineFill reports whether every run of spaces in vis runs up to a
// newline, i.e. no space acts as a real word separator (everything is
// glamour's styled line padding). A bare newline-boundary with no space
// reports true.
func allSpacesAreLineFill(vis string) bool {
	for i := 0; i < len(vis); {
		if vis[i] != ' ' {
			i++
			continue
		}
		j := i
		for j < len(vis) && vis[j] == ' ' {
			j++
		}
		if j >= len(vis) || vis[j] != '\n' {
			return false
		}
		i = j
	}
	return true
}
