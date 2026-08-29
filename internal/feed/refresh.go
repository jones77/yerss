package feed

import "time"

// Gate encodes the refresh lifecycle: a startup interval that gates
// auto-refresh and a cooldown that throttles manual refresh.
type Gate struct {
	MinInterval time.Duration
	Cooldown    time.Duration
}

// NeedsStartupRefresh reports whether startup should fetch feeds because more
// than MinInterval has elapsed (or no refresh has ever happened).
func (g Gate) NeedsStartupRefresh(last, now time.Time) bool {
	if last.IsZero() {
		return true
	}
	return now.Sub(last) > g.MinInterval
}

// ManualRefreshAllowed bypasses the startup interval and applies the cooldown.
// When not allowed it returns the time remaining until the next refresh is
// permitted.
func (g Gate) ManualRefreshAllowed(last, now time.Time) (bool, time.Duration) {
	if last.IsZero() {
		return true, 0
	}
	elapsed := now.Sub(last)
	if elapsed < g.Cooldown {
		return false, g.Cooldown - elapsed
	}
	return true, 0
}