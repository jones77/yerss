package render

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
//     (#707070 in both modes).
//   - Selection: the selected-row background highlight
//     (#707070 in both modes).
//   - StatusBar: the blue role covering the status bar, border inline text,
//     scrollbar thumb, and popup chrome (ANSI 12 bright blue / 4 blue).
type Palette struct {
	Text      lipgloss.Color
	Bright    lipgloss.Color
	Dim       lipgloss.Color
	Selection lipgloss.Color
	StatusBar lipgloss.Color
}

// DarkPalette returns the dark-theme palette.
func DarkPalette() Palette {
	return Palette{
		Text:      lipgloss.Color("7"),
		Bright:    lipgloss.Color("15"),
		Dim:       lipgloss.Color("#707070"),
		Selection: lipgloss.Color("#707070"),
		StatusBar: lipgloss.Color("12"),
	}
}

// LightPalette returns the light-theme palette.
func LightPalette() Palette {
	return Palette{
		Text:      lipgloss.Color("0"),
		Bright:    lipgloss.Color("0"),
		Dim:       lipgloss.Color("#707070"),
		Selection: lipgloss.Color("#707070"),
		StatusBar: lipgloss.Color("4"),
	}
}

// ResolvePalette picks the palette for the configured theme mode. "auto"
// detects the terminal's preferred background from the environment.
func ResolvePalette(mode string) Palette {
	switch mode {
	case "light":
		return LightPalette()
	case "dark":
		return DarkPalette()
	default:
		if DetectLightBackground() {
			return LightPalette()
		}
		return DarkPalette()
	}
}

// DetectLightBackground reports whether the terminal prefers a light
// background, inferred from the COLORFGBG environment variable ("fg;bg").
func DetectLightBackground() bool {
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

// DetectAsciiNeeded reports whether the terminal is unlikely to render
// Unicode box-drawing characters and ASCII fallback should be used.
func DetectAsciiNeeded() bool {
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
