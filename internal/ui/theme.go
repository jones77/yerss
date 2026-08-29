package ui

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Palette holds the colors used across the TUI for one theme.
type Palette struct {
	Accent    lipgloss.Color
	Dim       lipgloss.Color
	Bold      lipgloss.Color
	Border    lipgloss.Color
	StatusBar lipgloss.Color
}

func darkPalette() Palette {
	return Palette{
		Accent:    lipgloss.Color("#f7768e"),
		Dim:       lipgloss.Color("#565f89"),
		Bold:      lipgloss.Color("#c0caf5"),
		Border:    lipgloss.Color("#3b4261"),
		StatusBar: lipgloss.Color("#7aa2f7"),
	}
}

func lightPalette() Palette {
	return Palette{
		Accent:    lipgloss.Color("#d70062"),
		Dim:       lipgloss.Color("#8a8a8a"),
		Bold:      lipgloss.Color("#1c1c1c"),
		Border:    lipgloss.Color("#a0a0a0"),
		StatusBar: lipgloss.Color("#005f87"),
	}
}

// resolvePalette picks the palette for the configured theme mode. "auto"
// detects the terminal's preferred background from the environment.
func resolvePalette(mode string) Palette {
	switch mode {
	case "light":
		return lightPalette()
	case "dark":
		return darkPalette()
	default:
		if detectLightBackground() {
			return lightPalette()
		}
		return darkPalette()
	}
}

// detectLightBackground reports whether the terminal prefers a light
// background, inferred from the COLORFGBG environment variable ("fg;bg").
func detectLightBackground() bool {
	fgbg := os.Getenv("COLORFGBG")
	if fgbg == "" {
		return false
	}
	parts := strings.Split(fgbg, ";")
	if len(parts) != 2 {
		return false
	}
	switch strings.TrimSpace(parts[1]) {
	case "15", "7":
		return true
	}
	return false
}

// detectAsciiNeeded reports whether the terminal is unlikely to render
// Unicode box-drawing characters and ASCII fallback should be used.
func detectAsciiNeeded() bool {
	term := os.Getenv("TERM")
	if strings.HasPrefix(term, "linux") || strings.Contains(term, "dumb") {
		return true
	}
	lang := os.Getenv("LANG")
	if lang != "" && !strings.Contains(strings.ToUpper(lang), "UTF") {
		return true
	}
	return false
}