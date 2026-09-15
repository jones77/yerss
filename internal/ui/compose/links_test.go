package compose

import (
	"strings"
	"testing"

	"charm.land/glamour/v2"
	"github.com/charmbracelet/x/ansi"
)

// renderLink renders a markdown fragment through the app's glamour style config
// at the given width, mirroring how the article body is rendered.
func renderLink(t *testing.T, md string, w int) string {
	t.Helper()
	return renderLinkStyle(t, md, w, "dark")
}

// renderLinkStyle is renderLink with an explicit dark/light style.
func renderLinkStyle(t *testing.T, md string, w int, style string) string {
	t.Helper()
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(GlamourStyleConfig(style)),
		glamour.WithWordWrap(w),
	)
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	out, err := r.Render(md)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return strings.Trim(out, "\n")
}

// visible returns the stripped, trimmed plain text of a rendered line.
func visible(s string) string {
	return strings.TrimSpace(ansi.Strip(s))
}

func TestLinkColorsBlueInBothThemes(t *testing.T) {
	cases := []struct {
		style   string
		nameSGR string
		hrefSGR string
	}{
		{"dark", "\x1b[94;1m", "\x1b[94m"},
		{"light", "\x1b[34;1m", "\x1b[34m"},
	}
	for _, c := range cases {
		out := renderLinkStyle(t, "[world](https://example.com/post)", 100, c.style)
		if !strings.Contains(out, c.nameSGR+"world") {
			t.Errorf("%s: name span missing %q: %q", c.style, c.nameSGR+"world", out)
		}
		if !strings.Contains(out, c.hrefSGR+"\x1b]8;") {
			t.Errorf("%s: href span missing %q: %q", c.style, c.hrefSGR, out)
		}
	}
}

func TestStripLinkHrefsSingleLink(t *testing.T) {
	url := "https://example.com/post"
	in := renderLink(t, "[world]("+url+")", 100)
	out := StripLinkHrefs(in)
	if got := visible(out); got != "world" {
		t.Errorf("stripped = %q, want %q", got, "world")
	}
	if !strings.Contains(out, "\x1b]8;") || !strings.Contains(out, "\x1b]8;;") {
		t.Errorf("stripped output lost the OSC 8 name span: %q", out)
	}
	if strings.Contains(visible(out), url) {
		t.Errorf("URL text leaked into the visible output: %q", visible(out))
	}
}

func TestStripLinkHrefsTwoLinksSameURL(t *testing.T) {
	url := "https://example.com/post"
	in := renderLink(t, "[one]("+url+") and [two]("+url+")", 100)
	out := StripLinkHrefs(in)
	if got := visible(out); got != "one and two" {
		t.Errorf("stripped = %q, want %q", got, "one and two")
	}
	// Both names keep their OSC 8 spans (each carries the URL as its target);
	// the URL text appears zero times in the visible output.
	if n := strings.Count(visible(out), url); n != 0 {
		t.Errorf("URL text appears %d times in %q", n, visible(out))
	}
	if n := strings.Count(out, url); n != 2 {
		t.Errorf("URL occurrences = %d, want 2 (one per OSC 8 name target)", n)
	}
}

func TestStripLinkHrefsWideCharacters(t *testing.T) {
	in := renderLink(t, "[世界](https://example.com/post)", 100)
	out := StripLinkHrefs(in)
	if got := visible(out); got != "世界" {
		t.Errorf("stripped = %q, want %q", got, "世界")
	}
	if ansi.StringWidth(visible(out)) != ansi.StringWidth("世界") {
		t.Errorf("stripped visible width = %d, want %d", ansi.StringWidth(visible(out)), ansi.StringWidth("世界"))
	}
}

func TestStripLinkHrefsNameEqualsURL(t *testing.T) {
	url := "https://example.com/post"
	in := renderLink(t, "["+url+"]("+url+")", 100)
	out := StripLinkHrefs(in)
	if got := visible(out); got != url {
		t.Errorf("stripped = %q, want the name %q rendered once", got, url)
	}
	if n := strings.Count(visible(out), url); n != 1 {
		t.Errorf("URL text appears %d times, want exactly once", n)
	}
}

func TestStripLinkHrefsESCTerminator(t *testing.T) {
	// glamour emits BEL-terminated OSC 8; the stripper must also handle the
	// ESC-backslash terminator. Build the name/href pair by hand.
	open := ansi.SetHyperlink("https://example.com/post")
	closeSeq := "\x1b]8;;\x1b\\"
	in := open + "\x1b[94;1mworld\x1b[m" + closeSeq +
		"\x1b[37m \x1b[m" +
		"\x1b[94m" + open + "https://example.com/post" + closeSeq + "\x1b[m"
	out := StripLinkHrefs(in)
	if got := visible(out); got != "world" {
		t.Errorf("stripped = %q, want %q", got, "world")
	}
}

func TestStripLinkHrefsWrappedHref(t *testing.T) {
	// A long URL pushes the href onto the next line; the wrap leaves no leading
	// space, so the drop starts at the href's OSC 8 open and the name survives.
	url := "https://example.com/very/long/path/that/exceeds/the/content/width"
	in := renderLink(t, "[word]("+url+")", 30)
	if !strings.Contains(in, "\n") {
		t.Skip("width did not force a wrap")
	}
	out := StripLinkHrefs(in)
	if got := visible(out); got != "word" {
		t.Errorf("stripped = %q, want the name %q", got, "word")
	}
	if strings.Contains(visible(out), url) {
		t.Errorf("URL text leaked into visible output: %q", visible(out))
	}
	if n := strings.Count(out, url); n != 1 {
		t.Errorf("URL occurrences = %d, want 1 (the OSC 8 name target only)", n)
	}
}

func TestStripLinkHrefsWrappedNameKept(t *testing.T) {
	// A name long enough to wrap becomes several same-target spans; the
	// stripper must keep every one of them and still drop the href that
	// follows, line for line.
	name := "Richard Jebreal and the Democratic Party of Nothing is quite long"
	url := "https://example.com/very/long/path/that/exceeds/the/content/width"
	in := renderLink(t, "["+name+"]("+url+")", 35)
	if !strings.Contains(in, "\n") {
		t.Skip("width did not force a wrap")
	}
	out := StripLinkHrefs(in)
	if got := strings.Join(strings.Fields(visible(out)), " "); got != name {
		t.Errorf("stripped = %q, want the full wrapped name %q", got, name)
	}
	if strings.Contains(visible(out), url) {
		t.Errorf("URL text leaked into visible output: %q", visible(out))
	}
}

func TestStripLinkHrefsAdjacentSameURLWrapped(t *testing.T) {
	// Adjacent links to the same URL in a wrapped paragraph must each keep
	// their name and drop their (possibly multi-line) href.
	url := "https://example.com/very/long/path/that/exceeds/the/content/width"
	in := renderLink(t, "[one]("+url+") and [two]("+url+")", 30)
	if !strings.Contains(in, "\n") {
		t.Skip("width did not force a wrap")
	}
	out := StripLinkHrefs(in)
	if got := strings.Join(strings.Fields(visible(out)), " "); got != "one and two" {
		t.Errorf("stripped = %q, want %q", got, "one and two")
	}
	if strings.Contains(visible(out), url) {
		t.Errorf("URL text leaked into visible output: %q", visible(out))
	}
}

func TestStripLinkHrefsLinklessPassesThrough(t *testing.T) {
	in := renderLink(t, "plain body text with **bold** and `code`", 100)
	if out := StripLinkHrefs(in); out != in {
		t.Errorf("link-less text was modified:\n in: %q\nout: %q", in, out)
	}
}

func TestStripLinkHrefsWrappedHrefLeavesNoWhitespaceOnlyLine(t *testing.T) {
	// A link whose name ends a line while the href wraps to the next lines:
	// the strip must consume the href's trailing fill padding and wrap
	// newline, or a whitespace-only residual line survives and splits the
	// paragraph during reflow (the compact-mode "Hegseth told" bug).
	const w = 70
	url := "https://example.com/very/long/path/that/exceeds/the/content/width/and/wraps/again"
	md := "Hegseth [told](" + url + ") coalition partners that the new task force would continue the targeting and killing"
	in := renderLink(t, md, w)
	if !strings.Contains(in, "\n") {
		t.Skip("width did not force a wrap")
	}
	stripped := StripLinkHrefs(in)
	for i, l := range strings.Split(stripped, "\n") {
		if strings.TrimSpace(ansi.Strip(l)) == "" {
			t.Errorf("stripped output left whitespace-only line %d (href fill residual): %q", i, ansi.Strip(l))
		}
	}
	if got := strings.Contains(ansi.Strip(stripped), url); got {
		t.Errorf("URL text leaked into the visible output: %q", ansi.Strip(stripped))
	}
	out := strings.Join(strings.Fields(ansi.Strip(ReflowShortLines(stripped, w))), " ")
	if !strings.Contains(out, "Hegseth told coalition partners") {
		t.Errorf("name and following text not joined on one flow: %q", out)
	}
}
