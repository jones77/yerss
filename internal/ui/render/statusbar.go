package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal"
)

// StatusBarLeft builds the status bar's left segment: the transient status
// message when one is active, otherwise the last-refresh timestamp (or a
// "never refreshed" placeholder), and the active list filter when one is set
// and no status message is competing for the space.
func StatusBarLeft(base, dim lipgloss.Style, g BorderGlyphs, statusMsg string, statusActive, refreshing bool, lastRefreshedAt time.Time, showSeconds bool, filter string) string {
	left := ""
	if statusMsg != "" && statusActive {
		left = base.Render(statusMsg)
	} else if refreshing {
		left = base.Render("refreshing...")
	} else if lastRefreshedAt.IsZero() {
		left = base.Render("never refreshed")
	} else {
		t := lastRefreshedAt
		left = base.Render(t.Format(internal.TimeLayout(showSeconds))+" "+internal.LongDate(t)) + " " + dim.Render("last refresh")
	}
	if filter != "" && (statusMsg == "" || !statusActive) {
		left += " " + dim.Render(g.Bullet) + " " + base.Render("filter "+filter)
	}
	return left
}

// StatusBarRight builds the status bar's right segment, which runs
// `?: help • DB <n>MB • <percent>% • <n>/<total>`: the help hint first, then
// the on-disk database size (DB) as a whole number of megabytes, then the
// scroll percentage and position. Each element is prepended so the first
// right-aligned element is `?: help`.
func StatusBarRight(base, dim lipgloss.Style, bullet string, dbSize int64, n, total int) string {
	pct := 0
	if total > 0 {
		pct = (n*100 + total/2) / total
	}
	right := base.Render(fmt.Sprintf("%d/%d", n, total))
	right = base.Render(fmt.Sprintf("%d%%", pct)) + " " + dim.Render(bullet) + " " + right
	right = base.Render("DB "+FormatMB(dbSize)) + " " + dim.Render(bullet) + " " + right
	right = base.Render("?: help") + " " + dim.Render(bullet) + " " + right
	return right
}

// StatusBarLine joins the left and right segments with at least one space of
// padding, truncating the result to width. The whole line is rendered in base.
func StatusBarLine(width int, base lipgloss.Style, left, right string) string {
	pad := width - ansi.StringWidth(left) - ansi.StringWidth(right)
	if pad < 1 {
		pad = 1
	}
	line := left + strings.Repeat(" ", pad) + right
	line = Truncate(line, width)
	return base.Render(line)
}
