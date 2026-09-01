package ui

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
