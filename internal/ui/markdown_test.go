package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"yerss/internal/config"
	"yerss/internal/store"
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
		if got := glamourStandardStyle(c.theme); got != c.want {
			t.Errorf("glamourStandardStyle(%q) = %q, want %q", c.theme, got, c.want)
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
	m.cfg = config.Default()
	m.cfg.Display.Theme = "light"
	r2 := m.markdownRenderer(60)
	if r2 == nil || r2 == r1 {
		t.Errorf("theme change should build a new renderer")
	}
	if m.mdRendererStyle != "light" {
		t.Errorf("cache style not updated: %q", m.mdRendererStyle)
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
		if got := escapeMarkdownText(c.in); got != c.want {
			t.Errorf("escapeMarkdownText(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestMarkdownLink(t *testing.T) {
	if got := markdownLink("https://example.com/a"); got != "[https://example.com/a](https://example.com/a)" {
		t.Errorf("plain link = %q", got)
	}
	if got := markdownLink("https://example.com/a (b)"); got != "[https://example.com/a (b)](<https://example.com/a (b)>)" {
		t.Errorf("link with parens = %q", got)
	}
}

func TestRenderArticleMarkdownEscapesHeader(t *testing.T) {
	md := renderArticleMarkdown(store.Article{
		Title: "Apple's M4 *Pro* chip",
		Link:  "https://example.com/a",
	})
	out := strings.TrimSuffix(md, "\n\n")
	if strings.Contains(out, "**Apple's M4 *Pro*") {
		t.Errorf("un-escaped title leaks markdown: %q", out)
	}
	if !strings.Contains(out, `**Apple's M4 \*Pro\* chip**`) {
		t.Errorf("escaped title missing: %q", out)
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
	if got := indentListContinuations(in); got != in {
		t.Errorf("plain paragraph lines must be untouched, got %q", got)
	}
}
