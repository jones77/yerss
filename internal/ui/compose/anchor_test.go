package compose

import "testing"

func TestAnchorableLineSkipsBlankAndURL(t *testing.T) {
	urls := []string{"https://example.com/on-of-narco-terrorist-refueling-vessel-aug-28-2026/"}
	lines := []string{
		"",
		"",
		"https://example.com/on-of-narco-terrorist-refueling-vessel-aug-28-2026/",
		"on-of-narco-terrorist-refueling-vessel-aug-28-2026/",
		"claimed was refueling",
	}
	anchor, gap, found := AnchorableLine(lines, urls, 0)
	if !found {
		t.Fatal("expected an anchorable line")
	}
	if anchor != "claimed was refueling" {
		t.Errorf("anchor = %q, want the first real text line", anchor)
	}
	if gap != 2 {
		t.Errorf("gap = %d, want 2 blank separators skipped", gap)
	}

	// A URL-leading line anchors on the first real word that follows it.
	anchor, _, found = AnchorableLine([]string{"https://example.com/x tail words here"}, urls, 0)
	if !found || anchor != "tail words here" {
		t.Errorf("URL-leading line anchor = %q, found = %v; want %q", anchor, found, "tail words here")
	}

	// A tail that is blank or URL only reports no anchor.
	if _, _, found := AnchorableLine(lines[:3], urls, 0); found {
		t.Error("blank/URL-only tail should report no anchorable word")
	}
}

func TestFindAnchorLine(t *testing.T) {
	lines := []string{
		"title",
		"",
		"paragraph 1 text anchor1 https://example.com/very/long",
		"paragraph 2 text anchor2",
		"a wrapped",
	}
	if idx, ok := FindAnchorLine(lines, "paragraph 2 text anchor2"); !ok || idx != 3 {
		t.Errorf("exact match: got (%d, %v), want (3, true)", idx, ok)
	}
	if idx, ok := FindAnchorLine(lines, "paragraph 1 text anchor1"); !ok || idx != 2 {
		t.Errorf("prefix match over URL-extended line: got (%d, %v), want (2, true)", idx, ok)
	}
	if idx, ok := FindAnchorLine(lines, "a wrapped name shifted"); !ok || idx != 4 {
		t.Errorf("common-prefix fallback: got (%d, %v), want (4, true)", idx, ok)
	}
	if _, ok := FindAnchorLine(lines, "nothing matches this at all"); ok {
		t.Errorf("no-match scan should report false")
	}
}

func TestFindAnchorLineSplitAcrossReflow(t *testing.T) {
	// The anchor text was a full line in the old layout; after a reflow its
	// first word moved to the end of the previous line and the rest begins the
	// next. None of the exact/prefix/common-run passes match, so the split pair
	// must be found on the earlier line.
	text := "plane above provided live video of the vessel and the 11 men on board."
	lines := []string{
		"title",
		"",
		"early last September. A plane",
		"above provided live video of the vessel and the 11 men on board. In",
		"the video, a uniformed officer",
	}
	if idx, ok := FindAnchorLine(lines, text); !ok || idx != 2 {
		t.Errorf("split anchor: got (%d, %v), want (2, true)", idx, ok)
	}
	// A text that no pair reconstructs still reports false.
	if _, ok := FindAnchorLine(lines, "no such words exist here at all"); ok {
		t.Error("split scan should report false for an absent anchor")
	}
}
