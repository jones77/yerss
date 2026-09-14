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
