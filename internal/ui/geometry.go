package ui

import "time"

// wrapIndex advances a cursor by delta within a list of n items, wrapping at
// both ends. A non-positive n leaves the cursor unchanged.
func wrapIndex(i, delta, n int) int {
	if n <= 0 {
		return i
	}
	i += delta
	i %= n
	if i < 0 {
		i += n
	}
	return i
}

// clampIndex clamps a cursor into the range [0, n-1] of a list of n items. A
// non-positive n leaves the cursor unchanged. Unlike wrapIndex it does not wrap
// around the ends.
func clampIndex(i, n int) int {
	if n <= 0 {
		return i
	}
	if i < 0 {
		return 0
	}
	if i >= n {
		return n - 1
	}
	return i
}

// dayStart returns t truncated to local midnight.
func dayStart(t time.Time) time.Time {
	t = t.Local()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// dayKey returns the calendar-day grouping key for a publication time: a
// "20060102" key for a non-zero local time, or "undated" for the zero time.
func dayKey(t time.Time) string {
	if t.IsZero() {
		return "undated"
	}
	return dayStart(t).Format("20060102")
}
