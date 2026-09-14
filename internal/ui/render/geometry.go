package render

// ContentGeom computes the article content-area geometry shared by the article
// state builder, the mouse coordinate mapping, and the border renderer. textW
// is the content column width inside the horizontal padding, viewportH the
// number of content rows, and effPadY the clamped vertical padding applied as
// a permanent margin both above and below the content, between the content and
// the top and bottom borders. Keeping the computation in one place guarantees
// mouse hit-testing and rendering agree.
func ContentGeom(w, h, padX, padY int) (textW, viewportH, effPadY int) {
	effPadY = ClampPadY(padY, h)
	textW = max(1, w-2*padX-2)
	viewportH = max(1, h-2*effPadY-2)
	return textW, viewportH, effPadY
}

// WrapIndex advances a cursor by delta within a list of n items, wrapping at
// both ends. A non-positive n leaves the cursor unchanged.
func WrapIndex(i, delta, n int) int {
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

// ClampIndex clamps a cursor into the range [0, n-1] of a list of n items. A
// non-positive n leaves the cursor unchanged. Unlike WrapIndex it does not wrap
// around the ends.
func ClampIndex(i, n int) int {
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
