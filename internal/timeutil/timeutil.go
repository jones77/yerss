// Package timeutil holds the calendar-day and date-format helpers shared by the
// list view, article view, and JSON export.
package timeutil

import (
	"fmt"
	"time"
)

const (
	// LayoutTime is the clock-only format used for article rows.
	LayoutTime = "15:04"
	// LayoutTimeSeconds is the clock-only format with seconds, used when the
	// show_seconds display option is enabled.
	LayoutTimeSeconds = "15:04:05"
	// LayoutDateTime is the full local timestamp format used in article meta.
	LayoutDateTime = "2006-01-02 15:04:05"
	// LayoutDateTimeMinute is the date-plus-minute format used in the article
	// reader chrome, showing HH:MM without seconds.
	LayoutDateTimeMinute = "2006-01-02 15:04"
	// LayoutDayKey is the calendar-day grouping key format.
	LayoutDayKey = "20060102"
)

// TimeLayout returns the clock-only layout to use for a displayed time:
// minute precision (HH:MM) by default, or second precision (HH:MM:SS) when
// showSeconds is true.
func TimeLayout(showSeconds bool) string {
	if showSeconds {
		return LayoutTimeSeconds
	}
	return LayoutTime
}

// DateTimeLayout returns the date-and-time layout to use for a displayed
// timestamp: minute precision (2006-01-02 15:04) by default, or second
// precision (2006-01-02 15:04:05) when showSeconds is true.
func DateTimeLayout(showSeconds bool) string {
	if showSeconds {
		return LayoutDateTime
	}
	return LayoutDateTimeMinute
}

// DayStart returns t truncated to local midnight.
func DayStart(t time.Time) time.Time {
	t = t.Local()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// DayKey returns the calendar-day grouping key for a publication time: a
// "20060102" key for a non-zero local time, or "undated" for the zero time.
func DayKey(t time.Time) string {
	if t.IsZero() {
		return "undated"
	}
	return DayStart(t).Format(LayoutDayKey)
}

// DayLabel renders a day group's header label: the long local date, prefixed
// with "today, " or "yesterday, " when the day is the reference day or the one
// before it.
func DayLabel(date, now time.Time) string {
	today := DayStart(now)
	switch date {
	case today:
		return "today, " + LongDate(date)
	case today.AddDate(0, 0, -1):
		return "yesterday, " + LongDate(date)
	}
	return LongDate(date)
}

// LongDate renders a local calendar day as `Weekday Day-ordinal Month, Year`,
// for example "Saturday 28th August, 2026".
func LongDate(t time.Time) string {
	d := t.Day()
	return t.Format("Monday ") + fmt.Sprintf("%d%s ", d, ordinal(d)) + t.Format("January, 2006")
}

// ordinal returns the English ordinal suffix for n. The suffix is "th" for
// values ending 11-13 (modulo 100), otherwise it follows the ones digit.
func ordinal(n int) string {
	if n%100 >= 11 && n%100 <= 13 {
		return "th"
	}
	switch n % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	}
	return "th"
}
