package compose

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestFixBlockquoteTrailingPunctuationStaysOnBar(t *testing.T) {
	// Emulate glamour stranding the closing curly quote: the blockquote's final
	// word "on.”" wraps right after the period, leaving a bare "”" line whose
	// next line is ordinary paragraph text (not another bar line).
	bar := "\x1b[38;5;252m│ \x1b[m"
	in := bar + "“The ocean swallows up all the information we would usually rely on.\n" +
		"”\n" +
		"Twenty-six of the 227 total killed have been identified."
	out := ansi.Strip(FixBlockquoteRewrap(in))

	var sawQuote, sawPunct bool
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "ocean swallows") {
			sawQuote = true
			if !strings.HasPrefix(line, "│") {
				t.Errorf("blockquote text line missing bar: %q", line)
			}
		}
		if strings.HasSuffix(trimmed, "”") {
			sawPunct = true
			if !strings.HasPrefix(line, "│") {
				t.Errorf("stranded closing quote not folded under the bar: %q", line)
			}
		}
		if strings.Contains(trimmed, "Twenty-six") && strings.HasPrefix(line, "│") {
			t.Errorf("following paragraph absorbed into the blockquote: %q", line)
		}
	}
	if !sawQuote || !sawPunct {
		t.Fatalf("expected the quote text and its closing punctuation:\n%q", out)
	}
}
