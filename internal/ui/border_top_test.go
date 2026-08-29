package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

func stripTop(t *testing.T, line string) string {
	t.Helper()
	return ansi.Strip(line)
}

func TestTopBorderTitleFits(t *testing.T) {
	g := glyphsFor(false)
	p := darkPalette()
	line := stripTop(t, topBorder(40, g, p, "2026-01-02", "short"))

	if w := runewidth.StringWidth(line); w != 40 {
		t.Errorf("top border width = %d, want 40: %q", w, line)
	}
	if !strings.HasPrefix(line, "┌─ ") {
		t.Errorf("expected left rail ┌─ : %q", line)
	}
	if !strings.HasSuffix(line, "┐") {
		t.Errorf("expected right corner ┐: %q", line)
	}
	if !strings.Contains(line, "───") {
		t.Errorf("expected fill dashes when the title fits: %q", line)
	}
	if strings.Contains(line, "…") {
		t.Errorf("no ellipsis expected when the title fits: %q", line)
	}
}

func TestTopBorderTitleTruncated(t *testing.T) {
	g := glyphsFor(false)
	p := darkPalette()
	line := stripTop(t, topBorder(40, g, p, "2026-01-02", strings.Repeat("x", 100)))

	if w := runewidth.StringWidth(line); w != 40 {
		t.Errorf("top border width = %d, want 40: %q", w, line)
	}
	if !strings.Contains(line, "2026-01-02") {
		t.Errorf("date must stay intact: %q", line)
	}
	if !strings.HasSuffix(line, "… ─┐") {
		t.Errorf("expected … then a single dash before the right corner: %q", line)
	}
}

func TestTopBorderTitleTruncatedASCII(t *testing.T) {
	g := glyphsFor(true)
	p := darkPalette()
	line := stripTop(t, topBorder(40, g, p, "2026-01-02", strings.Repeat("x", 100)))

	if w := runewidth.StringWidth(line); w != 40 {
		t.Errorf("top border width = %d, want 40: %q", w, line)
	}
	if !strings.HasSuffix(line, "... -+") {
		t.Errorf("expected ASCII ellipsis ... and single dash before +: %q", line)
	}
	if strings.Contains(line, "…") {
		t.Errorf("no unicode ellipsis expected in ASCII mode: %q", line)
	}
}

func TestTopBorderCJKTitle(t *testing.T) {
	g := glyphsFor(false)
	p := darkPalette()
	title := "日本語のタイトルです" // 10 runes, 20 display columns
	line := stripTop(t, topBorder(30, g, p, "2026-01-02", title))

	if w := runewidth.StringWidth(line); w != 30 {
		t.Errorf("top border width = %d, want 30: %q", w, line)
	}
	if !strings.Contains(line, "日本語のタ") {
		t.Errorf("expected 4 CJK chars (10 cols) to survive truncation: %q", line)
	}
	if strings.Contains(line, "イトルです") {
		t.Errorf("title should be truncated before the remaining chars: %q", line)
	}
	if !strings.HasSuffix(line, "… ─┐") {
		t.Errorf("expected ellipsis and single dash: %q", line)
	}
}

func TestTopBorderEmptyDate(t *testing.T) {
	g := glyphsFor(false)
	p := darkPalette()
	line := stripTop(t, topBorder(40, g, p, "", "x"))

	if w := runewidth.StringWidth(line); w != 40 {
		t.Errorf("top border width = %d, want 40: %q", w, line)
	}
	if strings.Contains(line, "·") {
		t.Errorf("no dangling separator expected when date is empty: %q", line)
	}
	if !strings.Contains(line, " x ") {
		t.Errorf("title should still render: %q", line)
	}
}