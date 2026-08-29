package ui

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Palette holds the standard terminal colors used across the TUI for one
// theme, named for the role each color plays:
//
//   - Text: the article body foreground (ANSI 7 white / 0 black).
//   - Bright: emphasis such as unread titles (ANSI 15 bright white / 0 bold).
//   - Dim: the grey role covering the border, rails, and muted text
//     (ANSI 8 dark grey in both modes).
//   - StatusBar: the blue role covering the status bar, border inline text,
//     scrollbar thumb, and popup chrome (ANSI 12 bright blue / 4 blue).
type Palette struct {
	Text      lipgloss.Color
	Bright    lipgloss.Color
	Dim       lipgloss.Color
	StatusBar lipgloss.Color
}

func darkPalette() Palette {
	return Palette{
		Text:      lipgloss.Color("7"),
		Bright:    lipgloss.Color("15"),
		Dim:       lipgloss.Color("8"),
		StatusBar: lipgloss.Color("12"),
	}
}

func lightPalette() Palette {
	return Palette{
		Text:      lipgloss.Color("0"),
		Bright:    lipgloss.Color("0"),
		Dim:       lipgloss.Color("8"),
		StatusBar: lipgloss.Color("4"),
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