package compose

import (
	"unicode/utf8"
)

// SegKind classifies each ScanSeg: a text run that occupies display columns, an
// ordinary escape (CSI, or a bare ESC plus one char), or an OSC string.
type SegKind uint8

const (
	// SegText is a run of printable runes that together advance the display
	// column by widthIn(s, Start, End).
	SegText SegKind = iota
	// SegEscape is a CSI sequence (ESC [ ... final) or a bare ESC plus one
	// char; it occupies no display columns.
	SegEscape
	// SegOSC is an OSC string (ESC ] payload terminator) terminated by BEL or
	// ESC-backslash, or running to the end of the string when unterminated; it
	// occupies no display columns.
	SegOSC
)

// ScanSeg is one alternating run of a rendered ANSI line produced by ScanANSI.
// Start and End are byte offsets into the scanned string; Col is the display
// column at the segment's start, so text segments begin where the previous
// visible cell left off and escape segments carry the column right before
// them.
type ScanSeg struct {
	Start, End int
	Col        int
	Kind       SegKind
}

// ScanANSI splits a rendered ANSI string into its alternating escape and text
// runs, tracking the display column as it walks so consumers can advance past
// escapes and measure visible cells through a single code path. An OSC segment
// classified by ScanEscape covers the payload and its terminator (both bytes
// of the ESC-backslash terminator); one without a terminator runs to the end
// of the string and reports terminated=false from OSC.
func ScanANSI(s string) []ScanSeg {
	var segs []ScanSeg
	col := 0
	textStart := 0
	flushText := func(end int) {
		if textStart < end {
			segs = append(segs, ScanSeg{Start: textStart, End: end, Col: col, Kind: SegText})
			col += WidthIn(s, textStart, end)
		}
		textStart = end
	}
	for i := 0; i < len(s); {
		if s[i] != '\x1b' {
			_, size := utf8.DecodeRuneInString(s[i:])
			i += size
			continue
		}
		flushText(i)
		end, kind := ScanEscape(s, i)
		segs = append(segs, ScanSeg{Start: i, End: end, Col: col, Kind: kind})
		i = end
		textStart = i
	}
	flushText(len(s))
	return segs
}

// ScanEscape returns the byte offset one past the escape sequence starting at
// i (s[i] must be ESC) and its kind: CSI for an ESC-[ sequence, a bare-ESC
// single-char eat for anything else, and SegOSC for an OSC string (ESC ]),
// which consumes through its BEL or ESC-backslash terminator, or to the end of
// the string when unterminated.
func ScanEscape(s string, i int) (int, SegKind) {
	if i+1 < len(s) && s[i+1] == ']' {
		for j := i + 2; j < len(s); j++ {
			if s[j] == 0x07 {
				return j + 1, SegOSC
			}
			if s[j] == 0x1b && j+1 < len(s) && s[j+1] == '\\' {
				return j + 2, SegOSC
			}
		}
		return len(s), SegOSC
	}
	if i+1 < len(s) && s[i+1] == '[' {
		j := i + 2
		for j < len(s) && !('@' <= s[j] && s[j] <= '~') {
			j++
		}
		if j < len(s) {
			j++
		}
		return j, SegEscape
	}
	if i+1 < len(s) {
		return i + 2, SegEscape
	}
	return i + 1, SegEscape
}

// OSC returns the payload of an OSC segment (between the ESC ] opener and the
// terminator) and whether it was terminated by a BEL or ESC-backslash. An OSC
// that runs to the end of the string without a terminator reports
// terminated=false.
func (seg ScanSeg) OSC(s string) (payload string, terminated bool) {
	if seg.Kind != SegOSC {
		return "", false
	}
	for i := seg.Start + 2; i < seg.End; i++ {
		switch {
		case s[i] == 0x07:
			return s[seg.Start+2 : i], true
		case s[i] == 0x1b && i+1 < seg.End && s[i+1] == '\\':
			return s[seg.Start+2 : i], true
		}
	}
	return s[seg.Start+2 : seg.End], false
}

// WidthIn returns the display-column advance of s[start:end], summing the
// per-rune cell widths so wide and zero-width runes count exactly once in the
// same way the cell walkers do.
func WidthIn(s string, start, end int) int {
	w := 0
	for i := start; i < end; {
		r, size := utf8.DecodeRuneInString(s[i:])
		w += CellWidth(r)
		i += size
	}
	return w
}