package compose

import (
	"reflect"
	"testing"
)

func TestScanANSI(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []ScanSeg
	}{
		{"plain text", "abc", []ScanSeg{{0, 3, 0, SegText}}},
		{"empty", "", nil},
		{"CSI prefix then text", "\x1b[38;5;252m│ \x1b[mworld", []ScanSeg{
			{0, 11, 0, SegEscape},
			{11, 15, 0, SegText},
			{15, 18, 2, SegEscape},
			{18, 23, 2, SegText},
		}},
		{"bare esc eat", "a\x1bZb", []ScanSeg{
			{0, 1, 0, SegText},
			{1, 3, 1, SegEscape},
			{3, 4, 1, SegText},
		}},
		{"OSC BEL terminator", "x\x1b]8;;https://e/1\x07y", []ScanSeg{
			{0, 1, 0, SegText},
			{1, 18, 1, SegOSC},
			{18, 19, 1, SegText},
		}},
		{"OSC ST terminator", "x\x1b]8;;https://e/1\x1b\\y", []ScanSeg{
			{0, 1, 0, SegText},
			{1, 19, 1, SegOSC},
			{19, 20, 1, SegText},
		}},
		{"unterminated OSC eats rest", "x\x1b]8;;https://e/1", []ScanSeg{
			{0, 1, 0, SegText},
			{1, 17, 1, SegOSC},
		}},
		{"wide and zero-width runes", "界\u0301a", []ScanSeg{{0, 6, 0, SegText}}},
		{"adjacent escapes", "\x1b[1m\x1b[2m", []ScanSeg{
			{0, 4, 0, SegEscape},
			{4, 8, 0, SegEscape},
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ScanANSI(c.in)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("ScanANSI(%q) = %+v, want %+v", c.in, got, c.want)
			}
		})
	}
}

// segmentsTile checks that the scanner's segments cover the input exactly and
// in order: the first starts at 0, the last ends at len(s), and each segment
// begins where the previous one ended.
func TestSegmentsTileTheInput(t *testing.T) {
	inputs := []string{
		"",
		"plain",
		"\x1b[1m",
		"a\x1b[1mb\x1b[0m",
		"\x1b]8;;https://e/\x1b\\x\x1b]8;;\x1b\\",
		"界\u0301\x1b[1m\x07x",
	}
	for _, in := range inputs {
		segs := ScanANSI(in)
		at := 0
		for _, seg := range segs {
			if seg.Start != at {
				t.Errorf("ScanANSI(%q): segment %+v starts at %d, want %d", in, seg, seg.Start, at)
			}
			if seg.End <= seg.Start || seg.End > len(in) {
				t.Errorf("ScanANSI(%q): segment %+v out of range", in, seg)
			}
			at = seg.End
		}
		if at != len(in) {
			t.Errorf("ScanANSI(%q): segments cover %d/%d bytes", in, at, len(in))
		}
	}
}

func TestScanEscapeMatchesSkipEscapeOffsets(t *testing.T) {
	// The CSI and bare-ESC cases below replicate the offsets the removed
	// SkipEscape returned, so the scanner stays behavior-compatible.
	csi := "\x1b[38;5;252m"
	if end, kind := ScanEscape(csi, 0); end != len(csi) || kind != SegEscape {
		t.Errorf("ScanEscape(csi) = %d/%v, want %d/SegEscape", end, kind, len(csi))
	}
	bare := "abc\x1bZ"
	if end, kind := ScanEscape(bare, 3); end != 5 || kind != SegEscape {
		t.Errorf("ScanEscape(bare) = %d/%v, want 5/SegEscape", end, kind)
	}
}

func TestScanSegOSC(t *testing.T) {
	cases := []struct {
		name           string
		in             string
		wantPayload    string
		wantTerminated bool
	}{
		{"BEL", "\x1b]8;;https://e/\x07", "8;;https://e/", true},
		{"ST", "\x1b]8;;https://e/\x1b\\", "8;;https://e/", true},
		{"unterminated", "\x1b]8;;https://e/", "8;;https://e/", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var seg ScanSeg
			for _, s := range ScanANSI(c.in) {
				if s.Kind == SegOSC {
					seg = s
					break
				}
			}
			if seg.Kind != SegOSC {
				t.Fatalf("no OSC segment in %q", c.in)
			}
			payload, terminated := seg.OSC(c.in)
			if payload != c.wantPayload || terminated != c.wantTerminated {
				t.Errorf("OSC() = %q/%v, want %q/%v", payload, terminated, c.wantPayload, c.wantTerminated)
			}
		})
	}
}

func TestWidthInCountsWideAndZeroWidth(t *testing.T) {
	if got := WidthIn("ab", 0, 2); got != 2 {
		t.Errorf("WidthIn(ab) = %d, want 2", got)
	}
	if got := WidthIn("界", 0, 3); got != 2 {
		t.Errorf("WidthIn(wide) = %d, want 2", got)
	}
	if got := WidthIn("a\u0301", 0, 3); got != 1 {
		t.Errorf("WidthIn(combining) = %d, want 1", got)
	}
}

func TestStyledCellsSkipsAllEscapeKinds(t *testing.T) {
	// CSI, OSC (BEL and ST), and bare-ESC sequences must not count cells.
	csi := "\x1b[1mworld"
	if got, want := StyledCells(csi, 5), len(csi); got != want {
		t.Errorf("StyledCells over CSI = %d, want %d", got, want)
	}
	osc := "ab\x1b]8;;x\x1b\\cd"
	if got, want := StyledCells(osc, 4), len(osc); got != want {
		t.Errorf("StyledCells over OSC = %d, want %d", got, want)
	}
	bare := "a\x1bZbc"
	if got, want := StyledCells(bare, 3), len(bare); got != want {
		t.Errorf("StyledCells over bare ESC = %d, want %d", got, want)
	}
}