package compose

import (
	"sort"
	"strings"
)

// StripLinkHrefs removes the visible href span glamour appends to each body
// link (`text url` → `text`), leaving the OSC 8 name span intact so the link
// stays clickable. Glamour always emits a link as a name span (OSC 8 open +
// styled text + close) followed by a styled space and an href span (OSC 8 open
// + URL + close) carrying the same OSC 8 target, so per target the first span
// is kept and each immediately following same-target span — the href — is
// dropped along with its leading space. A link whose name wraps onto the next
// line before its href has no leading space to drop; the drop region then
// starts at the href's OSC 8 open. Runs of styled spaces (glamour's line
// padding) and everything else pass through unchanged.
func StripLinkHrefs(s string) string {
	segs := ScanANSI(s)
	counts := map[string]int{}
	var drops [][2]int
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
		// Glamour emits each link as name-then-href with the same target, so
		// per target the odd occurrences (1st, 3rd, …) are names to keep and
		// the even ones (2nd, 4th, …) are hrefs to drop.
		n := counts[u]
		counts[u] = n + 1
		if n%2 == 0 {
			continue
		}
		// An href span to drop. Extend the drop back over its leading space:
		// the nearest preceding text run, provided it is pure whitespace on a
		// single line (glamour's styled separator). A run carrying a newline is
		// a wrap boundary, so the wrapped-href case starts at the OSC 8 open.
		dropStart := seg.Start
		for k := i - 1; k >= 0; k-- {
			if segs[k].Kind != SegText {
				continue
			}
			t := s[segs[k].Start:segs[k].End]
			if !strings.Contains(t, "\n") && strings.TrimSpace(t) == "" {
				dropStart = segs[k].Start
			}
			break
		}
		// The drop runs through the href's closing OSC 8 (the URL text between
		// open and close), or to the end of the string when unterminated.
		dropEnd := len(s)
		for j := i + 1; j < len(segs); j++ {
			if segs[j].Kind == SegOSC {
				dropEnd = segs[j].End
				i = j
				break
			}
		}
		drops = append(drops, [2]int{dropStart, dropEnd})
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
