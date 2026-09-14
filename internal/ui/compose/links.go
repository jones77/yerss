package compose

import (
	"sort"
	"strings"
)

// linkSpan is one OSC 8 hyperlink span of a target: the segment index and byte
// offset of its open, the byte offset just past its closing OSC 8 (or the end
// of the string when unterminated), and the visible text it wraps.
type linkSpan struct {
	openSeg   int
	openStart int
	closeEnd  int
	text      string
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

	// Collect one span per OSC 8 hyperlink open, grouped by target in
	// occurrence order.
	var order []string
	byTarget := map[string][]linkSpan{}
	for i := 0; i < len(segs); i++ {
		seg := segs[i]
		if seg.Kind != SegOSC {
			continue
		}
		payload, _ := seg.OSC(s)
		if !strings.HasPrefix(payload, "8;") {
			continue
		}
		rest := payload[2:]
		u := rest
		if idx := strings.LastIndex(rest, ";"); idx >= 0 {
			u = rest[idx+1:]
		}
		if u == "" {
			continue
		}
		sp := linkSpan{openSeg: i, openStart: seg.Start, closeEnd: len(s)}
		for j := i + 1; j < len(segs); j++ {
			if segs[j].Kind == SegOSC {
				sp.closeEnd = segs[j].End
				break
			}
			if segs[j].Kind == SegText {
				sp.text += s[segs[j].Start:segs[j].End]
			}
		}
		if _, ok := byTarget[u]; !ok {
			order = append(order, u)
		}
		byTarget[u] = append(byTarget[u], sp)
	}

	// Group each target's spans into units: within a unit, consecutive spans
	// are wrapped continuation lines (they render as one link part); a real
	// space between spans — glamour's href prefix on the same line, or text
	// like " and " between two links — ends the unit. Units alternate name,
	// href, name, href … per target.
	var drops [][2]int
	for _, u := range order {
		spans := byTarget[u]
		type unit struct {
			drop       bool
			first, last int
		}
		units := []unit{{first: 0, last: 0}}
		for k := 1; k < len(spans); k++ {
			if joinsUnit(s, segs, spans[k-1], spans[k], u, units[len(units)-1].drop) {
				units[len(units)-1].last = k
			} else {
				units = append(units, unit{drop: !units[len(units)-1].drop, first: k, last: k})
			}
		}
		for _, un := range units {
			if !un.drop {
				continue
			}
			first, last := spans[un.first], spans[un.last]
			// Extend the drop back over the href unit's leading space: the
			// nearest preceding text run, provided it is pure whitespace on a
			// single line (glamour's styled separator). A run carrying a
			// newline is a wrap boundary, so a wrapped href starts at its
			// OSC 8 open and the name's fill padding survives.
			dropStart := first.openStart
			for k := first.openSeg - 1; k >= 0; k-- {
				if segs[k].Kind != SegText {
					continue
				}
				t := s[segs[k].Start:segs[k].End]
				if !strings.Contains(t, "\n") && strings.TrimSpace(t) == "" {
					dropStart = segs[k].Start
				}
				break
			}
			// The drop runs through the href unit's closing OSC 8, or to the
			// end of the string when unterminated.
			drops = append(drops, [2]int{dropStart, last.closeEnd})
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