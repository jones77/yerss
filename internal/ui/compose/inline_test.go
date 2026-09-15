package compose

import (
	"strings"
	"testing"

	xansi "github.com/charmbracelet/x/ansi"
)

// urlOSC8Targets returns the set of OSC 8 URLs opened on a line, one per
// open sequence, in order.
func urlOSC8Targets(line string) []string {
	var out []string
	for _, seg := range ScanANSI(line) {
		if seg.Kind != SegOSC {
			continue
		}
		payload, _ := seg.OSC(line)
		if u := oscURL(payload); u != "" {
			out = append(out, u)
		}
	}
	return out
}

func TestInlineLinkURLsReflowsPushedURLOntoNameLine(t *testing.T) {
	// A URL longer than the remaining space on the name's line: glamour pushes
	// it onto its own line below a half-empty name line. InlineLinkURLs must
	// re-flow the paragraph so the URL starts on the name's line and wraps at
	// the content width.
	const w = 70
	url := "https://example.com/a/very/long/path/that/extends/beyond/the/width/and/wraps/again"
	md := "Hegseth [told](" + url + ") coalition partners that the new task force would continue the targeting and killing"
	in := renderLink(t, md, w)
	if !hasWrappedHref(in) {
		t.Skip("width did not push the URL onto its own line")
	}

	out := InlineLinkURLs(in, w)
	lines := strings.Split(out, "\n")
	stripped := xansi.Strip(out)
	// The name and the URL share a line: the name must be followed by the URL
	// on the same line, never by a bare URL line under a half-empty name line.
	if idx := strings.Index(stripped, "Hegseth told"); idx >= 0 {
		lineEnd := strings.IndexByte(stripped[idx:], '\n')
		seg := stripped[idx:]
		if lineEnd >= 0 {
			seg = seg[:lineEnd]
		}
		if !strings.Contains(seg, "example.com") {
			t.Errorf("name line does not start the URL:\n%q", stripped)
		}
	} else {
		t.Errorf("re-flowed text lost the name: %q", stripped)
	}
	// The full URL is present in the visible text (concatenated across the
	// wrapped lines), and every line fits within the content width.
	if !strings.Contains(strings.Join(strings.Fields(stripped), ""), url) {
		t.Errorf("re-flowed text lost the URL: %q", stripped)
	}
	for i, l := range lines {
		if xansi.StringWidth(l) > w {
			t.Errorf("line %d width %d exceeds content width %d: %q", i, xansi.StringWidth(l), w, l)
		}
	}
}

func TestInlineLinkURLsKeepsWrappedURLLinesClickable(t *testing.T) {
	// Every wrapped line of a re-flowed URL must stay an OSC 8 hyperlink
	// carrying the full URL.
	const w = 60
	url := "https://example.com/very/long/path/that/exceeds/the/content/width/and/wraps/again"
	md := "see [this]( " + url + " ) for the full report and more details"
	in := renderLink(t, md, w)
	if !hasWrappedHref(in) {
		t.Skip("width did not push the URL onto its own line")
	}

	out := InlineLinkURLs(in, w)
	urlSegments := 0
	for i, l := range strings.Split(out, "\n") {
		targets := urlOSC8Targets(l)
		hasURL := false
		for _, target := range targets {
			if target == url {
				hasURL = true
			} else {
				t.Errorf("line %d OSC 8 target = %q, want %q: %q", i, target, url, l)
			}
		}
		if !hasURL {
			continue
		}
		urlSegments++
		if strings.TrimSpace(xansi.Strip(l)) == "" {
			t.Errorf("line %d carries the URL OSC 8 span but no visible URL text: %q", i, l)
		}
	}
	if urlSegments < 2 {
		t.Errorf("URL did not wrap across multiple lines (segments=%d), test is not exercising per-line OSC 8: %q", urlSegments, xansi.Strip(out))
	}
}
