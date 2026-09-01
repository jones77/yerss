package timeutil

import (
	"strings"
	"testing"
	"time"
)

func TestDayKey(t *testing.T) {
	if got := DayKey(time.Time{}); got != "undated" {
		t.Errorf("DayKey(zero) = %q, want %q", got, "undated")
	}
	day := time.Date(2026, 8, 28, 15, 4, 0, 0, time.Local)
	if got := DayKey(day); got != "20260828" {
		t.Errorf("DayKey(%v) = %q, want %q", day, got, "20260828")
	}
	nextDay := time.Date(2026, 8, 29, 0, 0, 0, 0, time.Local)
	if got := DayKey(nextDay); got != "20260829" {
		t.Errorf("DayKey(%v) = %q, want %q", nextDay, got, "20260829")
	}
}

func TestDayStart(t *testing.T) {
	in := time.Date(2026, 8, 28, 23, 59, 59, 999999999, time.Local)
	got := DayStart(in)
	want := time.Date(2026, 8, 28, 0, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Errorf("DayStart(%v) = %v, want %v", in, got, want)
	}
}

func TestDayLabel(t *testing.T) {
	old := time.Local
	time.Local = time.UTC
	t.Cleanup(func() { time.Local = old })
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	day := func(d int) time.Time { return time.Date(2026, 8, d, 0, 0, 0, 0, time.UTC) }
	cases := []struct {
		name string
		date time.Time
		want string
	}{
		{"today", day(28), "today, " + LongDate(day(28))},
		{"yesterday", day(27), "yesterday, " + LongDate(day(27))},
		{"older", day(26), LongDate(day(26))},
	}
	for _, c := range cases {
		if got := DayLabel(c.date, now); got != c.want {
			t.Errorf("DayLabel(%v, now) = %q, want %q", c.date, got, c.want)
		}
	}
}

func TestLongDateOrdinals(t *testing.T) {
	cases := []struct {
		day  int
		want string
	}{
		{1, "1st"}, {2, "2nd"}, {3, "3rd"},
		{11, "11th"}, {12, "12th"}, {13, "13th"},
		{21, "21st"}, {22, "22nd"}, {23, "23rd"},
		{30, "30th"}, {31, "31st"},
	}
	for _, c := range cases {
		tm := time.Date(2026, 8, c.day, 0, 0, 0, 0, time.UTC)
		got := LongDate(tm)
		if !strings.Contains(got, c.want) {
			t.Errorf("LongDate(%d) = %q, missing ordinal %q", c.day, got, c.want)
		}
		if !strings.Contains(got, "August, 2026") {
			t.Errorf("LongDate(%d) = %q, missing month/year", c.day, got)
		}
		if !strings.Contains(got, " ") {
			t.Errorf("LongDate(%d) = %q, expected a weekday prefix", c.day, got)
		}
	}
}

// TestOrdinalModulo100 pins the suffix beyond two-digit days: the "th" band is
// 11-13 modulo 100, so 111/112/113 take "th" instead of following the ones digit.
func TestOrdinalModulo100(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{101, "st"}, {102, "nd"}, {103, "rd"},
		{111, "th"}, {112, "th"}, {113, "th"},
		{121, "st"}, {122, "nd"}, {123, "rd"},
	}
	for _, c := range cases {
		if got := ordinal(c.n); got != c.want {
			t.Errorf("ordinal(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}
