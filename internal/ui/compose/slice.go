package compose

import (
	"strings"

	xansi "github.com/charmbracelet/x/ansi"
)

// reverse-video on/off, the highlight the article selection wraps around its
// region.
const (
	revOn  = "\x1b[7m"
	revOff = "\x1b[0m"
)

// InvertRange returns s with the display columns [from, to) rendered in reverse
// video, re-emitting the SGR and OSC-8 state active at each boundary so the
// inversion is continuous: a full SGR reset inside the region re-asserts
// reverse video instead of ending the highlight, and hyperlinks opened before or
// inside the region stay clickable. x/ansi's Cut/Truncate do not re-open the
// active state, so a selected fragment across a link carries an orphan reset
// that ends the highlight early; this walks ScanANSI and rebuilds the three
// slices itself. from and to are clamped to the line's width.
func InvertRange(s string, from, to int) string {
	width := xansi.StringWidth(s)
	if from < 0 {
		from = 0
	}
	if to > width {
		to = width
	}
	if from >= to {
		return s
	}
	fromByte := StyledCells(s, from)
	toByte := StyledCells(s, to)
	if fromByte >= toByte {
		return s
	}
	segs := ScanANSI(s)
	sgrFrom, oscFrom := stateAt(s, segs, fromByte)
	sgrTo, oscTo := stateAt(s, segs, toByte)

	var b strings.Builder
	b.WriteString(s[:fromByte])
	if oscFrom != "" {
		b.WriteString(oscClose)
	}
	if sgrFrom != "" {
		b.WriteString(revOff)
	}
	b.WriteString(revOn)
	b.WriteString(sgrFrom)
	b.WriteString(oscFrom)
	b.WriteString(invertInterior(s, segs, fromByte, toByte))
	b.WriteString(revOff)
	if toByte < len(s) {
		b.WriteString(sgrTo)
		b.WriteString(oscTo)
		b.WriteString(s[toByte:])
	}
	return b.String()
}

// invertInterior returns s[lo:hi] with reverse video re-asserted after every
// full SGR reset, so an interior reset does not clear the selection highlight.
func invertInterior(s string, segs []ScanSeg, lo, hi int) string {
	var b strings.Builder
	for _, seg := range segs {
		a, z := max(seg.Start, lo), min(seg.End, hi)
		if a >= z {
			continue
		}
		seq := s[a:z]
		b.WriteString(seq)
		if seg.Kind != SegText && isFullResetSGR(seq) {
			b.WriteString(revOn)
		}
	}
	return b.String()
}

// stateAt returns the SGR sequence and the OSC-8 open sequence active at byte
// offset off, processing every escape segment that ends at or before it. The
// SGR state is the concatenation of the sequences seen since the last full
// reset, which preserves last-wins semantics when re-emitted.
func stateAt(s string, segs []ScanSeg, off int) (sgr, osc string) {
	for _, seg := range segs {
		if seg.End > off {
			break
		}
		if seg.Kind == SegText {
			continue
		}
		seq := s[seg.Start:seg.End]
		if seg.Kind == SegEscape && isSGR(seq) {
			if isFullResetSGR(seq) {
				sgr = ""
			} else {
				sgr += seq
			}
			continue
		}
		if seg.Kind == SegOSC {
			payload, _ := seg.OSC(s)
			if strings.HasPrefix(payload, "8;") {
				if oscURL(payload) == "" {
					osc = ""
				} else {
					osc = seq
				}
			}
		}
	}
	return sgr, osc
}

// oscClose is the OSC 8 sequence that ends the active hyperlink.
const oscClose = "\x1b]8;;\x07"

// oscURL returns the URL carried by an OSC 8 payload ("8;params;url"), or ""
// when the sequence closes the hyperlink.
func oscURL(payload string) string {
	rest := strings.TrimPrefix(payload, "8;")
	if idx := strings.LastIndex(rest, ";"); idx >= 0 {
		return rest[idx+1:]
	}
	return rest
}

// isSGR reports whether seq is a CSI Select Graphic Rendition sequence.
func isSGR(seq string) bool {
	return len(seq) >= 3 && seq[0] == 0x1b && seq[1] == '[' && seq[len(seq)-1] == 'm'
}

// isFullResetSGR reports whether seq resets all graphic rendition (ESC[m or
// ESC[0m, including empty/zero parameter lists).
func isFullResetSGR(seq string) bool {
	if !isSGR(seq) {
		return false
	}
	params := seq[2 : len(seq)-1]
	if params == "" {
		return true
	}
	for _, p := range strings.Split(params, ";") {
		if p != "" && p != "0" {
			return false
		}
	}
	return true
}
