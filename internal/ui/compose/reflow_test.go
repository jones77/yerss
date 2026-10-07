package compose

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestReflowShortLinesRepairsStrippedParagraph(t *testing.T) {
	const w = 70
	pad := func(s string) string { return s + strings.Repeat(" ", w-ansi.StringWidth(s)) }
	in := pad("from Saint Lucia. Among those identified by a\u00a0consortium of") + "\n" +
		"journalists\u00a0from Venezuelan outlets were a" + "\n" +
		pad("bus driver, a mototaxi driver, men involved in fishing or the seafood")

	out := ReflowShortLines(in, w)
	for i, l := range strings.Split(out, "\n") {
		if got := ansi.StringWidth(l); got > w {
			t.Errorf("line %d exceeds content width: %d > %d: %q", i, got, w, ansi.Strip(l))
		}
	}
	flat := strings.Join(strings.Fields(ansi.Strip(out)), " ")
	if !strings.Contains(flat, "journalists from Venezuelan outlets were a bus driver") {
		t.Errorf("following text not pulled up onto the short line:\n%q", flat)
	}
}

func TestReflowShortLinesSkipsPrefixedParagraphs(t *testing.T) {
	const w = 40
	cases := map[string]string{
		"blockquote": "│ a short quoted line\n│ ”",
		"list item":  "• a short item\ncontinuation",
		"indented":   "  indented line",
	}
	for name, in := range cases {
		if got := ReflowShortLines(in, w); got != in {
			t.Errorf("%s re-flowed:\nin:  %q\nout: %q", name, in, got)
		}
	}
}

func TestReflowShortLinesLeavesPaddedParagraph(t *testing.T) {
	const w = 20
	in := strings.Repeat("x", w) + "\n" + strings.Repeat("y", w)
	if got := ReflowShortLines(in, w); got != in {
		t.Errorf("full-width paragraph re-flowed:\nin:  %q\nout: %q", in, got)
	}
}

func TestReflowAfterStripHrefSingleSpace(t *testing.T) {
	url := "https://example.com/very/long/path/that/exceeds/the/content/width/and/wraps/again"
	md := "the report claimed [the vessel was]( " + url + " ) refueling at sea and more text follows here to fill the line"
	const w = 60

	stripped := StripLinkHrefs(renderLink(t, md, w))
	if strings.Contains(ansi.Strip(stripped), "example.com") {
		t.Fatalf("href not stripped:\n%q", ansi.Strip(stripped))
	}
	out := ansi.Strip(ReflowShortLines(stripped, w))
	if strings.Contains(out, "  ") {
		t.Errorf("reflow left a double space: %q", out)
	}
	if !strings.Contains(out, "claimed the vessel was refueling") {
		t.Errorf("link name and following text not joined with a single space: %q", out)
	}
}

func TestReflowAfterStripPunctuatedLinkName(t *testing.T) {
	url := "https://example.com/very/long/path/that/exceeds/the/content/width/and/wraps/again"
	md := "the report [claimed.]( " + url + " ) was refueling at sea and more text follows here to fill the line"
	const w = 60

	stripped := StripLinkHrefs(renderLink(t, md, w))
	if strings.Contains(ansi.Strip(stripped), "example.com") {
		t.Fatalf("href not stripped:\n%q", ansi.Strip(stripped))
	}
	out := ansi.Strip(ReflowShortLines(stripped, w))
	if strings.Contains(out, "  ") {
		t.Errorf("reflow left a double space after punctuated name: %q", out)
	}
	if !strings.Contains(out, "claimed. was refueling") {
		t.Errorf("punctuated link name not joined with a single space: %q", out)
	}
}

func TestReflowAfterStripMissingSpace(t *testing.T) {
	// A link name that shares its line with the href while the following word
	// wraps to the next line: stripping must not fuse the name with the word
	// (the compact-mode "claimedsearching" bug).
	url := "https://media.freedom.press/media/documents/2026-3444_cost_estimate_2.pdf"
	md := "The city\u2019s response is beyond belief. In its response to the request on Monday, the Eugene Police Department [claimed](" + url + ") searching for records and processing the request would take at least 44,816 hours of staff time and cost approximately $2,484,599.04."
	const w = 100

	stripped := StripLinkHrefs(renderLink(t, md, w))
	if strings.Contains(ansi.Strip(stripped), "freedom.press") {
		t.Fatalf("href not stripped:\n%q", ansi.Strip(stripped))
	}
	out := ansi.Strip(ReflowShortLines(stripped, w))
	if strings.Contains(out, "claimedsearching") {
		t.Errorf("link name fused with following word: %q", out)
	}
	if strings.Contains(out, "claimed  searching") {
		t.Errorf("reflow left a double space: %q", out)
	}
	if !strings.Contains(out, "claimed searching") {
		t.Errorf("link name and following text not joined with a single space: %q", out)
	}
}

func TestReflowAfterStripLeadingComma(t *testing.T) {
	// A link name ends a line, the href wraps to the next line, and the markdown
	// comma is stranded right after the href: stripping the href leaves a line
	// starting with ", for a grand …". Re-flow must butt the comma against the
	// name ("$195,869.52,"), never insert a space before it.
	url := "https://media.freedom.press/media/documents/2026-3443_cost_estimate_4.pdf"
	md := "A public records request for similar records concerning a former police officer, Eric Klinko, was met with a comparatively modest estimate of [$195,869.52](" + url + "), for a grand total of nearly $2.7 million for both requests.)"
	const w = 90

	stripped := StripLinkHrefs(renderLink(t, md, w))
	if strings.Contains(ansi.Strip(stripped), "freedom.press") {
		t.Fatalf("href not stripped:\n%q", ansi.Strip(stripped))
	}
	out := ansi.Strip(ReflowShortLines(stripped, w))
	if strings.Contains(out, " ,") {
		t.Errorf("reflow left a space before the comma: %q", out)
	}
	if !strings.Contains(out, "$195,869.52, for a grand total") {
		t.Errorf("comma not butted against the link name: %q", out)
	}
}

func TestTrimStyledWSKeepsTrailingEscapes(t *testing.T) {
	line := "\x1b[37mhello\x1b[m\x1b[37m   \x1b[m"
	got := trimStyledWS(line)
	if ansi.Strip(got) != "hello" {
		t.Errorf("trimStyledWS visible = %q, want %q", ansi.Strip(got), "hello")
	}
	if !strings.HasSuffix(got, "\x1b[m") {
		t.Errorf("trimStyledWS dropped trailing reset: %q", got)
	}
}

func TestReflowAfterStripWrappedHref(t *testing.T) {
	url := "https://www.elclip.org/guiria-venezuela-vueltas-narcotrafico-bombardeos-caribe/?lang=en"
	md := "from Saint Lucia. Among those identified by a\u00a0[consortium of journalists](" + url + ")\u00a0from Venezuelan outlets were a bus driver, a mototaxi driver, men involved in fishing or the seafood trade, and a futsal player."
	const w = 70

	stripped := StripLinkHrefs(renderLink(t, md, w))
	if strings.Contains(ansi.Strip(stripped), "elclip.org") {
		t.Fatalf("href not stripped:\n%q", stripped)
	}
	out := ReflowShortLines(stripped, w)

	flat := strings.Join(strings.Fields(ansi.Strip(out)), " ")
	if !strings.Contains(flat, "journalists from Venezuelan outlets were a bus driver") {
		t.Errorf("paragraph not re-flowed across the stripped href:\n%q", ansi.Strip(out))
	}
	for i, l := range strings.Split(out, "\n") {
		if ansi.StringWidth(l) > w {
			t.Errorf("line %d exceeds width: %d > %d", i, ansi.StringWidth(l), w)
		}
	}
	if !strings.Contains(out, "\x1b]8;") {
		t.Errorf("reflow dropped the OSC 8 hyperlink span: %q", out)
	}
}
