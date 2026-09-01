package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"yerss/internal/config"
	"yerss/internal/ui/render"
)

// Article).
func actionLabel(a config.Action) string {
	return cases.Title(language.Und).String(strings.ReplaceAll(string(a), "_", " "))
}

// helpLabel renders an action label for the help popup, replacing the "Half"
// prefix of the half-page actions with the half glyph (½, or 1/2 in ASCII
// mode).
func (m *Model) helpLabel(a config.Action) string {
	label := actionLabel(a)
	if strings.HasPrefix(label, "Half") {
		return m.glyphs().Half + strings.TrimPrefix(label, "Half")
	}
	return label
}

// displayKeys renders normalized key strings for the help popup: a literal
// space key (as stored by normalizeKey for runtime matching) is shown as the
// word "space", and the normalized "pgdown" spelling is shown as "pgdn".
func displayKeys(keys []string) []string {
	out := make([]string, len(keys))
	for i, k := range keys {
		switch k {
		case " ":
			out[i] = "space"
		case "pgdown":
			out[i] = "pgdn"
		default:
			out[i] = k
		}
	}
	return out
}

func (m *Model) renderHelp() string {
	heading := lipgloss.NewStyle().Foreground(m.palette.StatusBar)
	keyStyle := lipgloss.NewStyle().Foreground(m.palette.Bright)

	// actionRow renders one binding with its label padded to labelW+2 columns
	// so the keys align across the section, leaving at least two spaces between
	// the label and its bound keys. Each bound key renders in the bright role
	// so it stands out; the commas between keys stay in the default text role.
	actionRow := func(a config.Action, labelW int) string {
		keys := displayKeys(m.cfg.Keybindings[a])
		if len(keys) == 0 {
			return ""
		}
		styled := make([]string, len(keys))
		for i, k := range keys {
			styled[i] = keyStyle.Render(k)
		}
		return fmt.Sprintf("%-*s%s", labelW+2, m.helpLabel(a), strings.Join(styled, ", "))
	}

	// section renders a heading followed by its action rows, aligning the keys
	// under the widest label in the section.
	section := func(label string, actions []config.Action) []string {
		labelW := 0
		for _, a := range actions {
			labelW = max(labelW, ansi.StringWidth(m.helpLabel(a)))
		}
		lines := []string{heading.Render(label)}
		for _, a := range actions {
			if r := actionRow(a, labelW); r != "" {
				lines = append(lines, r)
			}
		}
		return lines
	}

	// Global fills the left column; List view and Article view stack in the
	// right column, separated by a blank line.
	left := section("Global", config.GlobalActions())
	right := section("List view", config.ListActions())
	right = append(right, "")
	right = append(right, section("Article view", config.ArticleActions())...)

	n := max(len(left), len(right))
	for len(left) < n {
		left = append(left, "")
	}
	for len(right) < n {
		right = append(right, "")
	}

	leftW := 0
	for _, l := range left {
		leftW = max(leftW, ansi.StringWidth(l))
	}
	rightW := 0
	for _, l := range right {
		rightW = max(rightW, ansi.StringWidth(l))
	}

	// The popup including its border must not exceed 70 columns. The box keeps
	// one space of padding on each side.
	gap := 2
	boxW := min(70, m.width, leftW+gap+rightW+4)
	textW := max(1, boxW-4)

	g := m.glyphs()
	style := lipgloss.NewStyle().Foreground(m.palette.StatusBar)
	v := style.Render(g.V)
	pad := " "

	var content []string
	for i := 0; i < n; i++ {
		line := render.PadRight(left[i], leftW) + strings.Repeat(" ", gap) + render.PadRight(right[i], rightW)
		content = append(content, render.PadRight(render.Truncate(line, textW), textW))
	}

	// The top edge inlines the "Key Bindings" title; one blank row pads the
	// top and bottom of the content.
	blank := v + pad + strings.Repeat(" ", textW) + pad + v
	rows := []string{inlineTitleBorder("Key Bindings", boxW, g, m.palette), blank}
	for _, l := range content {
		rows = append(rows, v+pad+l+pad+v)
	}
	rows = append(rows, blank)
	rows = append(rows, style.Render(g.BL+strings.Repeat(g.H, boxW-2)+g.BR))
	return strings.Join(rows, "\n")
}

// inlineTitleBorder renders a popup's top edge as `┌─ Title ───┐`,
// centering the title in the bold chrome role and keeping the rails in the
// chrome role.
func inlineTitleBorder(title string, w int, g render.BorderGlyphs, p render.Palette) string {
	style := lipgloss.NewStyle().Foreground(p.StatusBar)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(p.StatusBar)
	leftRail := g.TL + g.H
	rightRail := g.H + g.TR
	core := " " + title + " "
	fill := max(0, w-ansi.StringWidth(leftRail)-ansi.StringWidth(core)-ansi.StringWidth(rightRail))
	left := fill / 2
	right := fill - left
	return style.Render(leftRail+strings.Repeat(g.H, left)) +
		titleStyle.Render(core) +
		style.Render(strings.Repeat(g.H, right)+rightRail)
}
