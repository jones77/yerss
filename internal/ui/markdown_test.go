package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/ui/compose"
)

func TestGlamourStandardStyleMapping(t *testing.T) {
	cases := []struct {
		theme string
		want  string
	}{
		{"dark", "dark"},
		{"light", "light"},
		{"auto", ""}, // resolved below via detection
	}
	for _, c := range cases {
		if c.theme == "auto" {
			continue
		}
		if got := compose.GlamourStandardStyle(c.theme); got != c.want {
			t.Errorf("compose.GlamourStandardStyle(%q) = %q, want %q", c.theme, got, c.want)
		}
	}
}

func TestRenderMarkdownStylesText(t *testing.T) {
	m, _ := newTestModel(t)
	out := m.renderMarkdown("**bold text**\n\n[world](https://example.com/post)", 60)
	if !strings.Contains(out, "\x1b[") {
		t.Errorf("glamour output should carry ANSI styling: %q", out)
	}
	if !strings.Contains(out, "\x1b]8;") {
		t.Errorf("glamour output should carry OSC 8 links: %q", out)
	}
	stripped := ansi.Strip(out)
	if !strings.Contains(stripped, "bold text") {
		t.Errorf("visible bold text missing: %q", stripped)
	}
	if !strings.Contains(stripped, "world https://example.com/post") {
		t.Errorf("visible link missing: %q", stripped)
	}
	if strings.HasPrefix(out, "\n") || strings.HasSuffix(out, "\n") {
		t.Errorf("glamour output should be trimmed of surrounding newlines: %q", out)
	}
}

func TestMarkdownRendererRebuildsOnWidthChange(t *testing.T) {
	m, _ := newTestModel(t)
	r1 := m.markdownRenderer(60)
	if r1 == nil {
		t.Fatal("renderer should be built")
	}
	r2 := m.markdownRenderer(60)
	if r1 != r2 {
		t.Errorf("same width should reuse the cached renderer")
	}
	r3 := m.markdownRenderer(80)
	if r3 == nil || r3 == r1 {
		t.Errorf("width change should build a new renderer")
	}
	if m.mdRendererW != 80 || m.mdRendererStyle != "dark" {
		t.Errorf("cache keys not updated: w=%d style=%q", m.mdRendererW, m.mdRendererStyle)
	}
}

func TestMarkdownRendererRebuildsOnThemeChange(t *testing.T) {
	m, _ := newTestModel(t)
	r1 := m.markdownRenderer(60)
	m.sess.Config().Display.Theme = "light"
	r2 := m.markdownRenderer(60)
	if r2 == nil || r2 == r1 {
		t.Errorf("theme change should build a new renderer")
	}
	if m.mdRendererStyle != "light" {
		t.Errorf("cache style not updated: %q", m.mdRendererStyle)
	}
}

func TestMarkdownBodyTextUsesStandardBaseColor(t *testing.T) {
	forceTrueColor(t)
	m, _ := newTestModel(t)
	m.sess.Config().Display.Theme = "dark"

	whiteEscape := escapePrefix(lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Render("x"))
	out := m.renderMarkdown("plain body text", 60)
	if !strings.Contains(out, whiteEscape) {
		t.Errorf("dark body text should use ANSI white (7): %q", out)
	}

	m.sess.Config().Display.Theme = "light"
	blackEscape := escapePrefix(lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Render("x"))
	out = m.renderMarkdown("plain body text", 60)
	if !strings.Contains(out, blackEscape) {
		t.Errorf("light body text should use ANSI black (0): %q", out)
	}
}

func TestEscapeMarkdownText(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Apple's M4 *Pro* chip", `Apple's M4 \*Pro\* chip`},
		{"a_b_c", `a\_b\_c`},
		{"[bracketed] text", `\[bracketed\] text`},
		{"em <em>not html</em>", `em \<em\>not html\</em\>`},
		{"plain text 123", "plain text 123"},
		{"https://example.com/a_(b)", `https://example.com/a\_(b)`},
	}
	for _, c := range cases {
		if got := compose.EscapeMarkdownText(c.in); got != c.want {
			t.Errorf("compose.EscapeMarkdownText(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIndentListContinuationsHangsIndent(t *testing.T) {
	m, _ := newTestModel(t)
	md := "- top bullet that is long enough to wrap around onto a second line\n" +
		"  - nested bullet also long enough to wrap a bit\n" +
		"  continuation text\n" +
		"\n" +
		"plain paragraph that wraps around but must stay unindented"
	stripped := ansi.Strip(m.renderMarkdown(md, 30))
	if !strings.Contains(stripped, "\n  enough to wrap around onto a") {
		t.Errorf("top-level continuation should hang-indent 2 below the bullet:\n%q", stripped)
	}
	if !strings.Contains(stripped, "\n    enough to wrap a bit") {
		t.Errorf("nested continuation should hang-indent under nested text:\n%q", stripped)
	}
	if strings.Contains(stripped, "\n  around but must stay") {
		t.Errorf("paragraph continuation must not be indented:\n%q", stripped)
	}
}

func TestIndentListContinuationsSkipsPlainParagraph(t *testing.T) {
	in := "first line of plain text\nsecond line of the same paragraph"
	if got := compose.IndentListContinuations(in); got != in {
		t.Errorf("plain paragraph lines must be untouched, got %q", got)
	}
}

func TestRenderMarkdownHeadingHasNoHashPrefix(t *testing.T) {
	m, _ := newTestModel(t)
	out := m.renderMarkdown("# Title\n\n## Section\n\n### Sub", 40)
	for _, level := range []string{"Title", "Section", "Sub"} {
		if strings.Contains(out, "# "+level) || strings.Contains(out, "## "+level) || strings.Contains(out, "### "+level) {
			t.Errorf("heading %q should not carry a '#' prefix:\n%q", level, out)
		}
		if !strings.Contains(out, level) {
			t.Errorf("heading text %q missing: %q", level, out)
		}
	}
}

func TestH2BoldH3Plain(t *testing.T) {
	m, _ := newTestModel(t)
	h2 := m.renderMarkdown("## Section", 40)
	h3 := m.renderMarkdown("### Sub", 40)
	if !strings.Contains(h2, ";1m") {
		t.Errorf("H2 heading should render bold, got: %q", h2)
	}
	if strings.Contains(h3, ";1m") {
		t.Errorf("H3 heading should render plain (not bold), got: %q", h3)
	}
}

func TestRenderMarkdownBlockquoteKeepsSolidBar(t *testing.T) {
	// glamour re-wraps a blockquote wider than its inner paragraph at many
	// content widths, stranding short words ("to", "of", "by") on their own line
	// that lose the "│" prefix and break the solid quote bar. renderMarkdown
	// must fold those orphans back so every blockquote line stays barred with no
	// content lost or clipped.
	src := "> What changes must now occur, in our way of looking at things, in our notions! Even the elementary concepts of time and space have begun to vacillate. Space is killed by the railways, and we are left with time alone."
	m, _ := newTestModel(t)
	for _, w := range []int{110, 90, 86, 81, 70, 60, 50, 46, 41, 40} {
		stripped := ansi.Strip(m.renderMarkdown(src, w))
		inQuote := false
		var words []string
		for _, line := range strings.Split(stripped, "\n") {
			if strings.TrimSpace(line) == "" {
				inQuote = false
				continue
			}
			if strings.HasPrefix(line, "│") {
				inQuote = true
				words = append(words, strings.Fields(strings.TrimPrefix(line, "│"))...)
				continue
			}
			if inQuote {
				t.Errorf("width %d: blockquote line missing '│' bar: %q", w, line)
			}
		}
		// No blockquote content may be dropped or clipped by the fix.
		compact := strings.Join(words, " ")
		for _, want := range []string{"notions", "vacillate", "railways", "alone"} {
			if !strings.Contains(compact, want) {
				t.Errorf("width %d: blockquote lost %q:\n%q", w, want, stripped)
			}
		}
	}
}

func TestFixBlockquoteRewrapUnit(t *testing.T) {
	// Emulate glamour's broken output: an orphaned word ("to", "time") on its
	// own line that lost the "│" bar, plus intact bar lines.
	bar := "\x1b[38;5;252m│ \x1b[m"
	in := bar + "notions! Even the elementary concepts of time and space have begun        \n" +
		"to\n" +
		bar + "vacillate. Space is killed by the railways, and we are left with   \n" +
		"time\n" +
		bar + "alone."
	got := ansi.Strip(compose.FixBlockquoteRewrap(in))
	lines := strings.Split(got, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "to" || strings.TrimSpace(line) == "time" {
			t.Errorf("orphan word still on its own line after rewrap: %q", line)
		}
		if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "│") {
			t.Errorf("blockquote line missing '│' bar: %q", line)
		}
	}
	compact := strings.Join(strings.Fields(got), " ")
	for _, want := range []string{"have begun to", "vacillate", "left with time", "alone"} {
		if !strings.Contains(compact, want) {
			t.Errorf("rewrap should keep %q in order:\n%q", want, got)
		}
	}
}

func TestRenderMarkdownBlockquoteNoClip(t *testing.T) {
	// A full-width bar line must never have content clipped to absorb an
	// orphaned word: "the lines to" must survive intact at every width.
	m, _ := newTestModel(t)
	md := "> What changes must now occur, in our way of looking at things, in our notions! Even the elementary concepts of time and space have begun to vacillate. Space is killed by the railways, and we are left with time alone. … Now you can travel to Orléans in four and a half hours, and it takes no longer to get to Rouen. Just imagine what will happen when the lines to Belgium and Germany are completed and connected up with their railways! I feel as if the mountains and forests of all countries were advancing on Paris. Even now, I can smell the German linden trees; the North Sea’s breakers are rolling against my door."
	for _, w := range []int{110, 90, 86, 81, 70, 60, 50, 46, 41, 40} {
		stripped := ansi.Strip(m.renderMarkdown(md, w))
		// Collapse to logical words (drop the "│" bar and spacing) so line
		// boundaries do not break substring checks.
		var sb strings.Builder
		for _, f := range strings.Fields(stripped) {
			if f == "│" {
				continue
			}
			sb.WriteString(f)
			sb.WriteByte(' ')
		}
		compact := strings.TrimSpace(sb.String())
		if !strings.Contains(compact, "the lines to Belgium") {
			t.Errorf("width %d: 'the lines to Belgium' corrupted:\n%q", w, compact)
		}
	}
}

func TestMarkdownToPlainTextFallback(t *testing.T) {
	got := compose.MarkdownToPlainText("**bold** and [world](https://example.com/post)\n\n# Title\n\n> quote\n\n* item")
	if strings.Contains(got, "**") {
		t.Errorf("plain-text fallback should not contain emphasis markers: %q", got)
	}
	if strings.Contains(got, "[") || strings.Contains(got, "]") {
		t.Errorf("plain-text fallback should not contain link syntax: %q", got)
	}
	if !strings.Contains(got, "bold") || !strings.Contains(got, "world") || !strings.Contains(got, "https://example.com/post") {
		t.Errorf("plain-text fallback should keep visible text and URL: %q", got)
	}
	if !strings.Contains(got, "Title") || !strings.Contains(got, "quote") || !strings.Contains(got, "item") {
		t.Errorf("plain-text fallback should keep headings/blockquote/list text: %q", got)
	}
}

func TestFixBlockquoteRewrapConsecutiveOrphans(t *testing.T) {
	// glamour can strand two words back-to-back before the next bar line; the
	// run scan must fold both under the quote bar with no content lost.
	bar := "\x1b[38;5;252m│ \x1b[m"
	in := bar + "notions! Even the elementary concepts of time and space have begun        \n" +
		"to\n" +
		"time\n" +
		bar + "vacillate. Space is killed by the railways, and we are left with   \n" +
		bar + "alone."
	got := ansi.Strip(compose.FixBlockquoteRewrap(in))
	for _, line := range strings.Split(got, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "to" || trimmed == "time" {
			t.Errorf("consecutive orphan words still stranded: %q", line)
		}
		if trimmed != "" && !strings.HasPrefix(line, "│") {
			t.Errorf("blockquote line missing '│' bar: %q", line)
		}
	}
	compact := strings.Join(strings.Fields(got), " ")
	for _, want := range []string{"have begun to", "to time", "vacillate", "left with", "alone"} {
		if !strings.Contains(compact, want) {
			t.Errorf("rewrap should keep %q in order:\n%q", want, got)
		}
	}
}

func TestStyledCellsSkipsEscapesAndWideChars(t *testing.T) {
	// The bar prefix is escape-heavy: cutting 2 visible cells must land exactly
	// after the styled "│ ", regardless of the bytes the escapes occupy.
	bar := "\x1b[38;5;252m│ \x1b[m"
	if got, want := compose.StyledCells(bar+"world", 2), len("\x1b[38;5;252m│ "); got != want {
		t.Errorf("styledCells over ANSI prefix = %d, want %d", got, want)
	}
	if compose.CellWidth('世') != 2 {
		t.Errorf("wide rune width = %d, want 2", compose.CellWidth('世'))
	}
	if compose.CellWidth('\u0301') != 0 {
		t.Errorf("combining mark width = %d, want 0", compose.CellWidth('\u0301'))
	}
}
