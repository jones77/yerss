package compose

import (
	"yerss/internal/convert"
	"yerss/internal/store"
)

// ImageBlock is one image block composed into the article content: the lead
// image (index 0, below the header) and each inline image, in document order.
// ImgStart is the first image row, CapStart the first attribution line, and
// ImgEnd the last composed line of the block — the last line of the wrapped
// caption, or of the photo when there is no caption. Snapping is defined by
// two boundaries: TOP (ImgStart, the image's first line at the viewport's
// first row) and BOTTOM (ImgEnd - vpH + 1, the last line of the wrapped
// caption at the viewport's last row), with EXIT (CapStart) and OFF
// (ImgStart - vpH) as the down- and up-exits. An inline block carries NativeImg
// when its lines are a terminal-side native placement, and NativeID the kitty
// image id that render was transmitted under (0 for halfblock or OSC 1337
// renders), so cleanup can delete the exact image.
type ImageBlock struct {
	URL       string
	ImgStart  int
	CapStart  int
	ImgEnd    int
	NativeImg bool
	NativeID  uint32
}

// CaptionWidth returns the standard caption width for the article content area:
// a fixed fraction of the content width, independent of any photo's own width,
// so a long caption wraps predictably and does not inflate a narrow photo's
// block. The fit and the composition both derive the caption width from the
// content width, so they always agree.
func CaptionWidth(contentW int) int {
	return max(1, contentW*7/10)
}

// ResolveImageAttribution returns the caption for an image (lead or inline)
// using a single fallback chain: figure caption first, then alt text (inline
// only), then link text (inline only), then "photo: <source>". The figure
// caption wins over alt text so an editorial figcaption (a gallery photo's own
// caption) renders instead of the image's alt. A lead image skips the alt and
// link-text steps; an image in a captioned figure whose caption belongs to a
// later image (a photo-grid's first photo) returns "" before any other source,
// so it never borrows that figure's caption.
func ResolveImageAttribution(a store.Article, url, alt, linkText string, isLead bool) string {
	cap, inFig := convert.ImageCaption(a.Content, url)
	if inFig && cap == "" {
		return ""
	}
	if inFig {
		return cap
	}
	if isLead {
		if src := store.SourceLabel(a.Link, a.FeedURL); src != "" {
			return "photo: " + src
		}
		return ""
	}
	if alt != "" {
		return alt
	}
	if linkText != "" {
		return linkText
	}
	if src := store.SourceLabel(a.Link, a.FeedURL); src != "" {
		return "photo: " + src
	}
	return ""
}

// SnapPos is the set of canonical snap offsets for one image block at a given
// viewport height: bottom (the last line of the wrapped caption at the
// viewport's last row), top (the image's first line at the viewport's first
// row), exit (the block scrolled out — for the lead the caption at the
// viewport top, for an inline block the line immediately after the wrapped
// caption, so the image and its caption leave as one unit — the down-exit),
// and off (the image's top at the fold, the block fully below the viewport —
// the up-exit). The two boundaries, bottom and top, are the snap stages; exit
// and off are where a block leaves the viewport in each direction.
type snapPos struct {
	bottom int
	top    int
	exit   int
	off    int
}

func snapPositions(i int, b ImageBlock, vpH int) snapPos {
	exit := b.CapStart
	if i > 0 {
		exit = b.ImgEnd + 1
	}
	return snapPos{
		bottom: b.ImgEnd - vpH + 1,
		top:    b.ImgStart,
		exit:   exit,
		off:    b.ImgStart - vpH,
	}
}

// offRevealTop reports whether an off-scroll target t (the offset that would
// put image i's top at the fold) would re-show an inline image above i without
// that image being flush to the viewport top, and if so returns that image's
// index and its TOP boundary. Two regimes are caught: the off target landing
// inside a previous inline image's photo range (leaving it partially visible,
// a blank strip), and — for adjacent images nearer than the viewport height —
// the target landing just above a previous inline image's top, leaving it fully
// visible but aligned a few rows down from the viewport top. In both cases the
// caller lands on that image's TOP so the reverse stage is aligned rather than
// skipped. The scan runs downward from the immediately preceding image, so the
// nearest image is revealed: consecutive images each show in sequence on the
// way up rather than the scan skipping the nearer one for a higher image whose
// range the target also touches.
func offRevealTop(t, i, vpH int, blocks []ImageBlock) (int, bool) {
	for j := i - 1; j > 0; j-- {
		nb := blocks[j]
		partially := t >= nb.ImgStart && t < nb.CapStart
		flush := t <= nb.ImgStart && nb.ImgEnd <= t+vpH
		if partially || flush {
			return j, true
		}
	}
	return 0, false
}

// SnapYOffset implements the image snap for single-line scroll moves across
// the article's ordered image blocks (lead at index 0, then inline in document
// order). Each block defines two snap boundaries — TOP (its first line at the
// viewport's first row) and BOTTOM (the last line of its wrapped caption at
// the viewport's last row) — and the snap steps between them. The lead block
// skips onto its caption in a single downward press — it was shown on open,
// so its boundaries are pre-consumed — and a fully visible lead photo (window
// top at or above ImgStart and window bottom reaching ImgEnd) likewise skips
// onto its caption. Inline blocks snap in two stages so an image is never
// skipped without first being shown: a downward move that leaves an inline
// image partially visible in the window — its top inside the window, its last
// line below the fold, whether it entered from below or a skip of the
// preceding image left it there — snaps its BOTTOM boundary so the wrapped
// caption is fully visible, the next downward move snaps its TOP boundary, and
// a move landing in its photo range then skips the whole image-and-caption
// block onto the line after its caption (EXIT), so the caption leaves with the
// image as one unit rather than being left at the viewport top.
// Upward moves mirror it: an image entering from above snaps its TOP boundary —
// firing as soon as the block's last line enters the window from above, even
// at its first row, so the whole image and caption appear at once — the next
// upward move snaps its BOTTOM boundary, an up move that cuts its
// last line below the fold snaps it fully below the fold (OFF) so it scrolls
// off cleanly, and a move landing in its photo range reveals it (TOP). A
// downward move landing in an inline photo's range entered from above snaps to
// its TOP (the second stage); starting at or inside it skips the block past
// its caption.
// A short inline block (shorter than the viewport) whose top a downward move
// brings into the window from below snaps straight to its TOP boundary,
// mirroring the upward entry snap, so it lands flush at the viewport top and
// the next downward move skips the whole block past its caption. An upward
// scroll-off that would land inside a preceding image's range reveals the
// immediately preceding image at its TOP, so consecutive images each show in
// sequence. On a full-image-capable terminal
// (native), an upward move while the lead photo is fully visible (and the
// offset is not already at the article top) skips straight to the article top
// (0), because re-transmitting a native photo on every header step is
// expensive; halfblock art is cheap, so it still scrolls the header line by
// line. before, the offset before the move, gates each stage so the snaps
// cannot re-fire and loop. Multi-line moves (page, half-page, goto-bottom) do
// not call this function at all: they land wherever they land, even mid-photo.
// An empty block list means no image; offsets outside the ranges (or moves
// that never enter them) are unchanged.
func SnapYOffset(yOffset, vpH, before int, blocks []ImageBlock, direction int, native bool) int {
	if len(blocks) == 0 {
		return yOffset
	}
	pos := make([]snapPos, len(blocks))
	for i, b := range blocks {
		pos[i] = snapPositions(i, b, vpH)
	}
	if direction > 0 {
		// A snap must not return the offset before the move: that target would
		// re-fire identically on the next move and loop (a block that fills the
		// viewport has its BOTTOM boundary equal to its TOP one, so the two
		// stages bottom out at the pre-move offset). A small backward
		// adjustment that shows an image (the reveal-to-top) is kept.
		snap := func(t int) int {
			if t != before {
				return t
			}
			return yOffset
		}
		// Down landing in a photo range. The lead block (index 0) skips onto
		// its caption in one press (it was shown on open; its boundaries are
		// pre-consumed). Inline blocks: a move entering the photo from above
		// snaps to its TOP (second stage), a move starting at or inside it
		// skips it onto its caption (EXIT).
		for i, b := range blocks {
			if yOffset >= b.ImgStart && yOffset < b.CapStart {
				if i > 0 && before < b.ImgStart {
					return snap(pos[i].top)
				}
				// A caption-less image (a photo-grid's first photo) followed
				// immediately by an adjacent inline image: its EXIT (one past
				// its last photo row, the separator gap) would leave the next
				// image resting one line below the viewport top, needing an
				// extra scroll to snap flush. Skip straight to the next
				// image's TOP instead.
				if b.CapStart == b.ImgEnd+1 && i+1 < len(blocks) && blocks[i+1].ImgStart == b.ImgEnd+2 {
					return snap(pos[i+1].top)
				}
				return snap(pos[i].exit)
			}
		}
		// A fully visible lead photo (index 0) is skipped in one press: it was
		// shown on open. Inline blocks are excluded — the fully-visible skip
		// there would preempt the two-stage snap and skip an image that was
		// never shown.
		if len(blocks) > 0 {
			if yOffset <= blocks[0].ImgStart && yOffset+vpH >= blocks[0].ImgEnd {
				return snap(pos[0].exit)
			}
		}
		// Rise stage: an inline image sitting fully visible at its BOTTOM
		// boundary (the bottom snap landed here) moves to its TOP boundary on
		// the next down move.
		for i, b := range blocks {
			if i > 0 && before == pos[i].bottom && yOffset <= b.ImgStart {
				return snap(pos[i].top)
			}
		}
		// Down entry snap for a short inline image: a single-line down move
		// that brings a short block's top into the window from below (its top
		// sat at or below the fold before the move, its block is shorter than
		// the viewport) snaps its TOP boundary so it lands flush at the
		// viewport top, mirroring the upward entry snap — a short photo
		// entered from above is aligned flush rather than scrolled into view
		// line-by-line — and the next down move skips it onto its caption. The
		// partial-visibility bottom snap below still handles blocks taller
		// than the viewport (which the top entering at the fold leaves
		// genuinely clipped) and images whose top was already inside the
		// window (a skip of the preceding image). before gates the entry so
		// the snap cannot re-fire and loop.
		for i, b := range blocks {
			if i > 0 &&
				b.ImgEnd-b.ImgStart+1 < vpH &&
				b.ImgStart >= yOffset && b.ImgStart < yOffset+vpH &&
				b.ImgStart >= before+vpH {
				return snap(b.ImgStart)
			}
		}
		// Bottom snap: an inline image that is partially visible in the window
		// — its top inside the window and not above the fold, its last line
		// below the fold — aligns its BOTTOM boundary so it is fully visible.
		// This fires both when a down move first brings an image's top into
		// view from below and when a skip of the preceding image leaves the
		// next image's top already inside the window (consecutive tall images
		// are spaced closer than the viewport height), so a partially visible
		// image never sits blank awaiting manual scrolling. A fully visible
		// image does not fire it, and an image entirely below the fold does
		// not fire it.
		best := -1
		for i, b := range blocks {
			if i > 0 && b.ImgStart >= yOffset && b.ImgStart < yOffset+vpH && b.ImgEnd > yOffset+vpH {
				if best < 0 || b.ImgStart > blocks[best].ImgStart {
					best = i
				}
			}
		}
		if best >= 0 {
			// A lower inline image whose top is inside the window is skipped
			// only when every inline image above it has been shown and passed
			// (scrolled onto its caption). When a higher inline image is still
			// fully visible — the first of a double photo, both near viewport
			// height and adjacent — the snap must not jump to the lower one's
			// bottom, or the higher image's own flush positions are never
			// reached. Fall through so the higher image's natural progression
			// (entry snap onto its top, then its caption) continues.
			skip := false
			for j, nb := range blocks {
				if j > 0 && j < best && nb.ImgStart >= yOffset && nb.ImgEnd <= yOffset+vpH {
					skip = true
					break
				}
			}
			if !skip {
				return snap(pos[best].bottom)
			}
		}
		// The separator gap between two adjacent image blocks (no body text
		// between them) is a single blank line: the next block's top is
		// exactly two lines past the previous block's last line. A down move
		// that lands on that blank line — the inline image's top one line
		// below the viewport top — snaps to the image's top so scrolling
		// never rests on the gap. A partially visible image is caught by the
		// bottom snap above, so only fully visible adjacent images reach
		// here.
		for i, b := range blocks {
			if i > 0 && b.ImgStart-yOffset == 1 && b.ImgStart == blocks[i-1].ImgEnd+2 {
				return snap(b.ImgStart)
			}
		}
		return yOffset
	}
	if direction < 0 {
		// A snap must not return the offset before the move, which would re-fire
		// identically on the next move and loop; unchanged offsets fall through
		// to normal scrolling.
		snap := func(t int) int {
			if t != before {
				return t
			}
			return yOffset
		}
		// A native photo that is fully visible and not at the article top
		// skips straight to the top, because re-transmitting a native photo
		// per header step is expensive. This applies only to the lead image:
		// the article top is above it, and inline images must scroll normally.
		if native && len(blocks) > 0 {
			b := blocks[0]
			if yOffset > 0 && yOffset <= b.ImgStart && yOffset+vpH >= b.ImgEnd {
				return snap(0)
			}
		}
		// Up from an inline image's BOTTOM boundary (its last line on the
		// viewport's last row, reached by the down bottom-snap or the up
		// sink) snaps it fully below the fold (OFF), completing the two-stage
		// exit in one press: a scroll-k from the caption-on-bottom-line state
		// clears the photo out of view rather than pausing on the caption or
		// revealing the previous image with this one still partially visible.
		// The target puts the image's top at the fold, clamped to the article
		// top.
		for i := range blocks {
			if i > 0 && before == pos[i].bottom {
				t := pos[i].off
				// With images spaced closer than the viewport height, the
				// scroll-off target can land inside a previous inline image's
				// photo range, leaving it partially visible — or, for a double
				// photo, just above it, leaving it fully visible but not flush
				// to the viewport top. Reveal that image at its top so the
				// reverse stage is aligned.
				if j, ok := offRevealTop(t, i, vpH, blocks); ok {
					return snap(pos[j].top)
				}
				if t < 0 {
					t = 0
				}
				return snap(t)
			}
		}
		// Up landing in a photo range reveals the image at its TOP boundary.
		for i, b := range blocks {
			if yOffset >= b.ImgStart && yOffset < b.CapStart {
				return snap(pos[i].top)
			}
		}
		// Sink stage: an inline image sitting fully visible at its TOP
		// boundary (the top snap landed here) moves to its BOTTOM boundary on
		// the next up move. A block that fills the viewport has its BOTTOM
		// boundary equal to its TOP one, so there is no distinct bottom stage:
		// the next up move scrolls the image fully off (revealing a previous
		// image whose range the target would land in).
		for i, b := range blocks {
			if i > 0 && before == b.ImgStart && yOffset >= b.ImgEnd-vpH {
				target := pos[i].bottom
				if target < yOffset {
					return snap(target)
				}
				// The block fills the viewport: its BOTTOM boundary is at or
				// past the offset just reached, so there is no distinct bottom
				// stage (top and bottom show the same full-screen image). The
				// next up move scrolls the image fully off (revealing a
				// previous image whose range the target would land in).
				off := pos[i].off
				if j, ok := offRevealTop(off, i, vpH, blocks); ok {
					return snap(pos[j].top)
				}
				return snap(off)
			}
		}
		// Top snap (up): the inline image whose bottom just entered the
		// viewport from above (it was entirely above the window before the
		// move) aligns its TOP boundary to the viewport top, so the whole
		// image and its caption appear at once. The bottom entering at the
		// window's first row counts, so a caption-only frame (only the wrapped
		// caption's last line visible) is never shown. before gates the entry
		// so the snap cannot re-fire and loop.
		best := -1
		for i, b := range blocks {
			if i > 0 && b.ImgEnd >= yOffset && b.ImgEnd <= before {
				if best < 0 || b.ImgEnd < blocks[best].ImgEnd {
					best = i
				}
			}
		}
		if best >= 0 {
			return snap(pos[best].top)
		}
		// Bottom scroll-off: an inline image partially visible with its top
		// inside the window and its last line below the fold (an up move
		// scrolled its last line past the fold, off the bottom edge) snaps
		// fully below the fold (OFF) so it scrolls off cleanly rather than
		// rendering a blank strip while its placement is deleted. The target
		// puts the image's top at the fold, so it cannot re-fire and loop.
		best = -1
		for i, b := range blocks {
			if i > 0 && b.ImgStart >= yOffset && b.ImgStart < yOffset+vpH && b.ImgEnd > yOffset+vpH {
				if best < 0 || b.ImgStart > blocks[best].ImgStart {
					best = i
				}
			}
		}
		if best >= 0 {
			target := pos[best].off
			// With images spaced closer than the viewport height, the
			// scroll-off target can land inside the previous inline image's
			// photo range, leaving it partially visible (blank) — or, for a
			// double photo, just above it, leaving it fully visible but not
			// flush to the viewport top. Reveal that image instead so the
			// transition stays clean.
			if j, ok := offRevealTop(target, best, vpH, blocks); ok {
				return snap(pos[j].top)
			}
			return snap(target)
		}
		return yOffset
	}
	return yOffset
}

// SnapAfterRecompose repositions the offset after a re-composition so an image
// load that lands while the user is reading at its position cannot leave the
// offset inside a photo range or with an image's top materialized mid-window.
// An offset inside the lead photo snaps onto its caption (the lead was shown on
// open); an offset inside an inline photo, or below an inline image whose top
// sits within the window, snaps forward to the inline image's top (motion 1),
// matching the two-motion scroll contract so the image is shown rather than
// left partially clipped.
func SnapAfterRecompose(offset, vpH int, blocks []ImageBlock) int {
	for i, b := range blocks {
		if offset >= b.ImgStart && offset < b.CapStart {
			if i == 0 {
				return b.CapStart
			}
			return b.ImgStart
		}
	}
	// An inline image whose top materialized inside the window below the
	// offset snaps forward to its top (the nearest such image), so it is shown
	// rather than scrolled past unseen. The lead block is excluded: its top
	// sitting below the header is the normal open state and the header must
	// stay visible.
	best := -1
	for i, b := range blocks {
		if i > 0 && b.ImgStart > offset && b.ImgStart < offset+vpH {
			if best < 0 || b.ImgStart < blocks[best].ImgStart {
				best = i
			}
		}
	}
	if best >= 0 {
		return blocks[best].ImgStart
	}
	return offset
}
