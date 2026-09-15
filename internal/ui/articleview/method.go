package articleview

import (
	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal/store"
	"yerss/internal/ui/compose"
)

// RecomputeNativeClear keeps Update's deferred native-clear contract on the
// article state; the computation itself lives on the image machinery, which
// reports whether the native snapshot changed so the frame cache can be
// invalidated. It returns whether the snapshot changed.
func (s *State) RecomputeNativeClear(active bool) bool {
	if s.images.RecomputeNativeClear(active, s) {
		s.nativeGen++
		return true
	}
	return false
}

// FlushRecompose runs the pending article recompose, at most once, when the
// flag is set and the article view is still active, then fires the deferred
// native renders (and newly in-range loads) the recompose's fresh anchors
// brought into view. Image-load message handlers set the recompose-pending
// flag; the caller invokes FlushRecompose after its message switch so the
// per-message work stays bounded to cache updates and the recompose — which
// reuses the cached derivation — runs once per image message. It returns the
// tea.Cmd firing that deferred work, or nil.
func (s *State) FlushRecompose(active bool) tea.Cmd {
	if !s.recomposePending {
		return nil
	}
	s.recomposePending = false
	if active {
		s.RecomposeInPlace()
		return s.images.AdvanceFrontierLoads(s)
	}
	return nil
}

// RecomposeInPlace re-composes the open article, preserving the user's reading
// position: the offset is bumped only by the lines the recompose inserted above
// the viewport top — a block whose placeholder row sits below the current top
// (a load that landed past the fold while the user read above it, or one below
// the restored position on a restart) does not move the reading position — and
// then snapped so a load that landed while the user was reading at its position
// cannot leave the offset inside a photo range or with an image's top
// materialized mid-window.
func (s *State) RecomposeInPlace() {
	s.recomposeCount++
	oldOffset := s.Viewport.YOffset
	oldLen := len(s.lines)
	vpH := s.Viewport.Height
	// A user reading at the article's last screen (GotoBottom, or scrolled to
	// the end) stays at the bottom: re-composition must not let an image load
	// snap the offset up to a photo top that happens to sit mid-window, since
	// that would hide the article's end the user asked to see. The offset bump
	// below already tracks the bottom through the inserted lines; the snap
	// would be the only thing to move it.
	atBottom := oldOffset > 0 && oldOffset >= oldLen-vpH
	oldAnchors := s.anchorRows
	a := *s.article
	ns := s.Recompose(a)
	*s = ns
	if atBottom {
		s.Viewport.GotoBottom()
		return
	}
	if oldOffset > 0 {
		// Insertions above the viewport top: the recompose shifts the reading
		// position by however many lines were added above it. The first image
		// whose old anchor row is at or below the old offset marks the
		// boundary — everything above it was inserted above the top (its own
		// anchor moving by exactly that sum), everything from it down was not.
		// Without such an image the whole delta is above the top.
		inserted := len(s.lines) - oldLen
		for i, a := range oldAnchors {
			if a >= oldOffset {
				if i < len(s.anchorRows) {
					inserted = s.anchorRows[i] - a
				}
				break
			}
		}
		s.Viewport.SetYOffset(compose.SnapAfterRecompose(oldOffset+inserted, vpH, s.imageBlocks, s.leadShown))
	} else {
		s.Viewport.SetYOffset(0)
	}
}

// AnchorAt captures the reading position at offset as content a restart can
// relocate: when the offset sits inside a composed block's span (photo or
// caption rows), the block's URL anchors it as a unit; otherwise the leading
// text of the first content line at or below the offset, with the blank lines
// skipped before it. It mirrors ToggleLinkExpand's anchor selection, so the
// same position can be re-found in a fresh placeholder layout whose image
// blocks have not materialized.
func (s *State) AnchorAt(offset int) (blockURL, anchor string, gap int) {
	if offset < 0 || offset >= len(s.lines) {
		return "", "", 0
	}
	for i := range s.imageBlocks {
		b := &s.imageBlocks[i]
		if offset >= b.ImgStart && offset <= b.ImgEnd {
			return b.URL, "", 0
		}
	}
	anchor, gap, _ = compose.AnchorableLine(s.lines, URLsOf(s.links), offset)
	return "", anchor, gap
}

// RestoreAt repositions a freshly re-composed article to the saved content
// anchor: an image block anchors to the placeholder row the article reserves
// for its URL (its top materializes at the viewport top when the block loads),
// and a text anchor is re-found in the new placeholder layout and clamped back
// by the saved gap. It returns false when the saved selection carries no usable
// anchor — an offset-only selection saved before anchors existed — so the
// caller falls back to the raw offset.
func (s *State) RestoreAt(sel store.LastSelection) bool {
	vpH := s.Viewport.Height
	if sel.ArticleBlock != "" {
		if i := s.URLIndex(sel.ArticleBlock); i >= 0 {
			if rows := s.AnchorRows(); i < len(rows) {
				if idx := rows[i]; idx >= 0 && idx <= len(s.lines)-vpH {
					s.Viewport.SetYOffset(idx)
					return true
				}
			}
		}
		return false
	}
	if sel.ArticleAnchor == "" {
		return false
	}
	idx, ok := compose.FindAnchorLine(s.lines, sel.ArticleAnchor)
	if !ok {
		return false
	}
	idx -= sel.ArticleGap
	if maxOffset := len(s.lines) - vpH; idx > maxOffset {
		idx = maxOffset
	}
	if idx < 0 {
		idx = 0
	}
	s.Viewport.SetYOffset(idx)
	return true
}

// ToggleLinkExpand flips the compact/expanded link mode and re-composes the
// article, keeping the first visible content line anchored: the line that was
// first visible before the toggle stays first visible after it. RecomposeInPlace
// alone preserves a reading position by bumping the offset by the article's
// total line-count delta, which is only correct for insertions at the top of
// the content; a link-mode toggle widens lines across the whole body, so the
// handle is re-anchored on the first visible line's visible text instead. A
// reader pinned at the bottom stays pinned by RecomposeInPlace's bottom path.
// It returns the new LinkExpand value.
func (s *State) ToggleLinkExpand() bool {
	oldOffset := s.Viewport.YOffset
	oldLen := len(s.lines)
	vpH := s.Viewport.Height
	atBottom := oldOffset > 0 && oldOffset >= oldLen-vpH
	anchor := ""
	gap := 0
	anchorFound := false
	// A photo block anchors as a unit: its photo rows strip to blank/escape
	// noise, so the text anchor would miss them, and anchoring a caption line
	// at the viewport top would push the photo above it off screen. When the
	// first visible line lies inside a composed block's span — its photo rows
	// and its attribution/caption rows both — remember the URL and the row the
	// viewport starts at within the block, so the same photo region stays on
	// screen after the mode change.
	blockURL, blockRow := "", -1
	if !atBottom && oldOffset >= 0 && oldOffset < oldLen {
		for i := range s.imageBlocks {
			b := &s.imageBlocks[i]
			if oldOffset >= b.ImgStart && oldOffset <= b.ImgEnd {
				blockURL, blockRow = b.URL, oldOffset-b.ImgStart
				break
			}
		}
		if blockRow < 0 {
			anchor, gap, anchorFound = compose.AnchorableLine(s.lines, URLsOf(s.links), oldOffset)
		}
	}
	s.LinkExpand = !s.LinkExpand
	s.RecomposeInPlace()
	if blockRow >= 0 {
		for i := range s.imageBlocks {
			if s.imageBlocks[i].URL == blockURL {
				idx := s.imageBlocks[i].ImgStart + blockRow
				if maxOffset := len(s.lines) - vpH; idx > maxOffset {
					idx = maxOffset
				}
				if idx >= 0 {
					s.Viewport.SetYOffset(idx)
				}
				return s.LinkExpand
			}
		}
		// The URL's block vanished after the recompose (e.g. dedup dropped
		// it); fall through to the text anchor with no better option.
	}
	if !anchorFound {
		// Nothing below the offset survives a mode change (the article tail is
		// blank or URL text only), so there is no line to re-anchor on; glue to
		// the bottom rather than let the total-delta offset drift.
		s.Viewport.GotoBottom()
		return s.LinkExpand
	}
	idx, ok := compose.FindAnchorLine(s.lines, anchor)
	if !ok {
		// The anchor line vanished (for example a link name removed by the
		// mode change); the total-delta offset is no worse than today's
		// behavior, so leave it.
		return s.LinkExpand
	}
	idx -= gap
	if maxOffset := len(s.lines) - vpH; idx > maxOffset {
		idx = maxOffset
	}
	if idx < 0 {
		idx = 0
	}
	s.Viewport.SetYOffset(idx)
	return s.LinkExpand
}
