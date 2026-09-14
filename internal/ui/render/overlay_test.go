package render

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func TestOverlayCentersByDisplayWidth(t *testing.T) {
	base := "abcdefghijklmnopqrstuvwxyz" // 26 wide
	pop := "XXX"                         // 3 wide
	got := Overlay(base, pop)
	want := "abcdefghijk" + "XXX" + "opqrstuvwxyz"
	if got != want {
		t.Errorf("overlay = %q, want %q", got, want)
	}
}

func TestOverlayPreservesBaseOutsidePopup(t *testing.T) {
	base := strings.Join([]string{
		strings.Repeat("a", 20),
		strings.Repeat("b", 20),
		strings.Repeat("c", 20),
	}, "\n")
	pop := "YY"
	got := Overlay(base, pop)
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), got)
	}
	if lines[0] != strings.Repeat("a", 20) {
		t.Errorf("top line modified: %q", lines[0])
	}
	if lines[2] != strings.Repeat("c", 20) {
		t.Errorf("bottom line modified: %q", lines[2])
	}
	// popup width 2, base width 20 → centered at columns 9..10.
	if want := "bbbbbbbbb" + "YY" + "bbbbbbbbb"; lines[1] != want {
		t.Errorf("middle line not centered: %q, want %q", lines[1], want)
	}
}

func TestOverlayIgnoresANSIWidth(t *testing.T) {
	base := "aaaaaaaaaa" // 10 wide
	pop := lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Render("XX")
	got := Overlay(base, pop)
	stripped := ansi.Strip(got)
	if stripped != "aaaaXXaaaa" {
		t.Errorf("styled overlay stripped = %q, want %q", stripped, "aaaaXXaaaa")
	}
}