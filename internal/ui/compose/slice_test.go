package compose

import (
	"strings"
	"testing"
	"unicode/utf8"

	xansi "github.com/charmbracelet/x/ansi"
)

// styledLine is a line with plain text, a styled hyperlink, and more plain
// text: "before link after" with the link wrapped in OSC 8 and a color.
const styledLine = "\x1b[37mbefore \x1b[m\x1b]8;;https://example.com/x\x07\x1b[96;1mlink\x1b[m\x1b]8;;\x07\x1b[37m after\x1b[m"

// reverseColumns walks a rendered line and reports, per display column, whether
// reverse video is active when the column is drawn. It interprets just the SGR
// attributes the highlight path emits (7 on, 27/0 off).
func reverseColumns(s string) []bool {
	rev := false
	var out []bool
	for _, seg := range ScanANSI(s) {
		if seg.Kind == SegText {
			for i := seg.Start; i < seg.End; {
				r, size := utf8.DecodeRuneInString(s[i:])
				for k := 0; k < CellWidth(r); k++ {
					out = append(out, rev)
				}
				i += size
			}
			continue
		}
		seq := s[seg.Start:seg.End]
		if !isSGR(seq) {
			continue
		}
		for _, p := range strings.Split(seq[2:len(seq)-1], ";") {
			switch p {
			case "7":
				rev = true
			case "27", "", "0":
				rev = false
			}
		}
	}
	return out
}

// endState walks a rendered line and reports the reverse-video and hyperlink
// state left active at its end.
func endState(s string) (reverse, linkOpen bool) {
	for _, seg := range ScanANSI(s) {
		if seg.Kind == SegOSC {
			payload, _ := seg.OSC(s)
			if strings.HasPrefix(payload, "8;") {
				linkOpen = oscURL(payload) != ""
			}
			continue
		}
		if seg.Kind != SegEscape {
			continue
		}
		seq := s[seg.Start:seg.End]
		if !isSGR(seq) {
			continue
		}
		for _, p := range strings.Split(seq[2:len(seq)-1], ";") {
			switch p {
			case "7":
				reverse = true
			case "27", "", "0":
				reverse = false
			}
		}
	}
	return reverse, linkOpen
}

func TestInvertRangeContinuousAcrossLink(t *testing.T) {
	got := InvertRange(styledLine, 3, 14)
	if xansi.Strip(got) != xansi.Strip(styledLine) {
		t.Fatalf("visible text changed:\n got %q\nwant %q", xansi.Strip(got), xansi.Strip(styledLine))
	}
	rev := reverseColumns(got)
	if len(rev) != 17 {
		t.Fatalf("column count = %d, want 17", len(rev))
	}
	for i := 0; i < len(rev); i++ {
		want := i >= 3 && i < 14
		if rev[i] != want {
			t.Errorf("column %d reverse = %v, want %v (line %q)", i, rev[i], want, got)
		}
	}
	if !strings.Contains(got, "\x1b]8;;https://example.com/x") {
		t.Errorf("link open dropped inside the selection: %q", got)
	}
}

func TestInvertRangeMidLinkStaysBalanced(t *testing.T) {
	// Start and end inside the link name: the before slice must close the open
	// link and the after slice must re-open it so the escapes stay balanced.
	got := InvertRange(styledLine, 7, 9)
	if xansi.Strip(got) != xansi.Strip(styledLine) {
		t.Fatalf("visible text changed: %q", xansi.Strip(got))
	}
	rev := reverseColumns(got)
	for i := 7; i < 9; i++ {
		if !rev[i] {
			t.Errorf("column %d (link name) not inverted: %q", i, got)
		}
	}
	for i := 0; i < len(rev); i++ {
		if i >= 7 && i < 9 {
			continue
		}
		if rev[i] {
			t.Errorf("column %d outside the selection inverted: %q", i, got)
		}
	}
	if reverse, linkOpen := endState(got); reverse || linkOpen {
		t.Errorf("line ends with state active: reverse=%v linkOpen=%v (%q)", reverse, linkOpen, got)
	}
}

func TestInvertRangeWholeLinkBalanced(t *testing.T) {
	// The selection exactly covers the link name: it must stay inverted and the
	// link must remain clickable across the selection.
	got := InvertRange(styledLine, 7, 11)
	rev := reverseColumns(got)
	for i := 7; i < 11; i++ {
		if !rev[i] {
			t.Errorf("column %d not inverted: %q", i, got)
		}
	}
	if !strings.Contains(got, "\x1b]8;;https://example.com/x") {
		t.Errorf("link dropped: %q", got)
	}
	if reverse, linkOpen := endState(got); reverse || linkOpen {
		t.Errorf("line ends with state active: reverse=%v linkOpen=%v", reverse, linkOpen)
	}
}
