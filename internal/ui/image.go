package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yerss/convert"
	"yerss/internal/image"
	"yerss/internal/store"
)

// imagesEnabled reports whether lead image rendering is active: the config mode
// is not off and ASCII fallback is not forcing images off (the -a flag and
// terminal detection both set m.ascii).
func (m *Model) imagesEnabled() bool {
	return !m.ascii && m.cfg.Display.Images != "off"
}

// nativeImages reports whether the terminal can render real photos natively.
func (m *Model) nativeImages() bool {
	return m.imgNative.Protocol != image.ProtocolNone
}

// nativeImageClear returns the native-image clear sequence to prepend to the
// current frame for every native image the open article renders that the frame
// does not display. kitty graphics placements float above text and the erase
// commands a frame repaint issues have no effect on them, so a placement
// survives scrolling the photo's rows out of the viewport unless explicitly
// deleted; the frame deletes each scrolled-out image's placement by its stable
// id so stale photos cannot persist or stack. A placement is kept only when its
// rows [imgStart, capStart) are fully contained in the visible viewport window:
// a partially visible image (its top above the fold or its bottom below it)
// would draw past the window and overlap the article border, so the frame
// deletes it and the transmit on its first line is suppressed by
// suppressClippedNativeTransmits. The sequence is empty for protocols whose
// images are cell-bound (OSC 1337) or absent, and for fully contained images
// (that frame re-transmits the image with its stable placement id, which
// replaces the placement). Leaving the article view (or an article with no
// composed native image) clears every placement on screen so the previous
// article's photos cannot linger.
func (m *Model) nativeImageClear() string {
	if m.imgNative.Protocol != image.ProtocolKitty {
		return ""
	}
	if m.view != viewArticle {
		return m.imgNative.Clear()
	}
	vp := m.article.viewport
	var sb strings.Builder
	seen := 0
	for _, b := range m.article.imageBlocks {
		if !b.nativeImg {
			continue
		}
		seen++
		if b.imgStart >= vp.YOffset && b.capStart <= vp.YOffset+vp.Height {
			continue
		}
		sb.WriteString(image.DeletePlacement(b.url))
	}
	if seen == 0 {
		return m.imgNative.Clear()
	}
	return sb.String()
}

// suppressClippedNativeTransmits blanks the native-image transmit escape on the
// first line of each native image that is only partially visible in the article
// viewport (its top above the fold or its bottom below it). The placement for
// such an image is deleted by nativeImageClear, but the transmit escape carried
// in the block's first line would re-create it every frame and draw the image
// over the article border; stripping the escape leaves the row blank so the
// partially visible image never paints outside the content area. Fully
// contained images keep their transmit.
func (m *Model) suppressClippedNativeTransmits(lines []string) {
	if m.imgNative.Protocol != image.ProtocolKitty {
		return
	}
	vp := m.article.viewport
	for _, b := range m.article.imageBlocks {
		if !b.nativeImg {
			continue
		}
		if b.imgStart >= vp.YOffset && b.capStart <= vp.YOffset+vp.Height {
			continue
		}
		if idx := b.imgStart - vp.YOffset; idx >= 0 && idx < len(lines) {
			lines[idx] = ansi.Strip(lines[idx])
		}
		m.previewClippedImage(lines, b)
	}
}

// previewClippedImage fills the visible rows of a partially visible native
// image with its halfblock preview, so the portion of a photo that is in the
// window shows a blocky preview rather than a blank strip (the native placement
// is deleted for a partially visible image because it would draw over the
// article border). The preview renders the cached decoded image at the native
// block's row count, so it aligns with the native's cell box and centers the
// same way. When the decoded image is not cached the rows stay blank.
func (m *Model) previewClippedImage(lines []string, b imageBlock) {
	vp := m.article.viewport
	img, ok := m.imgCache.Get(b.url)
	if !ok {
		return
	}
	rows := b.capStart - b.imgStart
	pre, err := m.imgRenderer.Render(img, vp.Width, rows)
	if err != nil {
		return
	}
	imgW := image.BlockWidth(pre)
	pad := max(0, (vp.Width-imgW)/2)
	lo := max(b.imgStart, vp.YOffset)
	hi := min(b.capStart, vp.YOffset+vp.Height)
	for r := lo; r < hi; r++ {
		idx := r - vp.YOffset
		k := r - b.imgStart
		if idx < 0 || idx >= len(lines) || k >= len(pre) {
			continue
		}
		lines[idx] = strings.Repeat(" ", pad) + pre[k]
	}
}

// captionWidth returns the standard caption width for the article content area:
// a fixed fraction of the content width, independent of any photo's own width,
// so a long caption wraps predictably and does not inflate a narrow photo's
// block. The fit and the composition both derive the caption width from the
// content width, so they always agree.
func captionWidth(contentW int) int {
	return max(1, contentW*7/10)
}

// composeImageBlock renders the image at url as content lines with the given
// attribute centered directly beneath, for both the lead image and inline
// images. On a native-capable terminal it serves a previously-rendered native
// photo block (keyed by the current content width and viewport height) once
// the off-thread NativeCmd has finished; otherwise it renders the halfblock
// block from the in-memory decoded image, or serves the stored text block when
// its width matches the content width. It returns nil when images are
// disabled, no URL is set, or no image source is available yet; the int
// reports how many of the returned lines are image rows (the attribution lines
// follow them) and the bool reports whether the lines carry a native
// inline-image escape (a terminal-side placement that outlives the frame).
func (m *Model) composeImageBlock(url, attr string, contentW, vpH, headerLines int) ([]string, int, bool) {
	if !m.imagesEnabled() || url == "" {
		return nil, 0, false
	}
	if m.nativeImages() {
		if lines, ok := m.imgNatives.Get(image.NativeKey(url, contentW, vpH)); ok {
			return m.composeBlock(lines, attr, contentW), len(lines), true
		}
	}
	if img, ok := m.imgCache.Get(url); ok {
		block, rows := m.fitBlock(func(maxH int) []string {
			lines, err := m.imgRenderer.Render(img, contentW, maxH)
			if err != nil {
				return nil
			}
			return lines
		}, attr, contentW, vpH, headerLines)
		return block, rows, false
	}
	if lines, ok := m.imgBlocks.Get(url); ok && image.BlockWidth(lines) == contentW {
		return m.composeBlock(lines, attr, contentW), len(lines), false
	}
	return nil, 0, false
}

// fitBlock renders the block through renderFn, shrinking the height cap until
// the block plus its wrapped attribution fits the viewport budget alongside
// the header and one line of body text, then composes the centered block. It
// returns the composed lines and the number of image rows they contain (the
// attribution lines follow). Bounded at three iterations.
func (m *Model) fitBlock(renderFn func(maxH int) []string, attr string, contentW, vpH, headerLines int) ([]string, int) {
	// Budget: blank line above + attribution lines + blank line below + one
	// body line. The attribution line count is unknown until the caption width
	// is known (the caption wraps at the standard caption width, not the
	// image's), so render optimistically reserving one attribution line, then
	// re-render smaller if the wrapped attribution needs more.
	base := vpH - headerLines - 3
	maxH := max(1, base-1)
	if attr == "" {
		maxH = max(1, base)
	}
	captionW := captionWidth(contentW)
	var rendered []string
	for i := 0; ; i++ {
		rendered = renderFn(maxH)
		if rendered == nil {
			return nil, 0
		}
		attrLines := 0
		if attr != "" {
			attrLines = len(image.LayoutWrapLines(attr, captionW))
		}
		want := max(1, base-attrLines)
		if want == maxH || i == 2 {
			break
		}
		maxH = want
	}
	return m.composeBlock(rendered, attr, contentW), len(rendered)
}

// composeBlock centers rendered block lines within contentW and appends the
// wrapped attribution lines, each centered beneath the image at the standard
// caption width.
func (m *Model) composeBlock(rendered []string, attr string, contentW int) []string {
	imgW := image.BlockWidth(rendered)
	photoPad := max(0, (contentW-imgW)/2)
	out := make([]string, 0, len(rendered)+2)
	for _, l := range rendered {
		out = append(out, strings.Repeat(" ", photoPad)+l)
	}
	if attr != "" {
		style := lipgloss.NewStyle().Foreground(m.palette.Dim)
		captionW := captionWidth(contentW)
		left := photoPad + (imgW-captionW)/2
		if left < 0 {
			left = 0
		}
		for _, al := range image.LayoutWrapLines(attr, captionW) {
			pad := left + max(0, (captionW-ansi.StringWidth(al))/2)
			out = append(out, strings.Repeat(" ", pad)+style.Render(al))
		}
	}
	return out
}

// articleAttribution returns the photo attribution for the article's lead
// image: the credit extracted from the article's HTML when present, otherwise
// "photo: <source>" derived by the list view's source-identifier rules. It
// returns "" when neither is available. A lead image that sits in a captioned
// figure whose caption belongs to a later image (a photo-grid's first photo)
// returns "" rather than the source label.
func articleAttribution(a store.Article) string {
	if cap, inFig := convert.ImageCaption(a.Content, a.ImageURL); inFig {
		return cap
	}
	if src := sourceID(a.Link, a.FeedURL); src != "" {
		return "photo: " + src
	}
	return ""
}

// fireImageLoad returns a tea.Cmd that loads every image the open article
// renders (the lead and each inline image, in order) through the cache
// hierarchy when image rendering is enabled, with one in-flight guard per URL.
func (m *Model) fireImageLoad(a store.Article) tea.Cmd {
	if !m.imagesEnabled() {
		return nil
	}
	var cmds []tea.Cmd
	for pos, url := range m.article.imageURLs {
		if url == "" || m.imgLoading[url] {
			continue
		}
		cmds = append(cmds, m.loadImageCmd(a, url, pos))
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// loadImageCmd starts the image load for the article image at url and position,
// marking the URL in flight so duplicate loads are not started (for example on
// resize).
func (m *Model) loadImageCmd(a store.Article, url string, position int) tea.Cmd {
	width, vpH, _ := contentGeom(m.width, m.height, m.cfg.Display.PaddingX, m.cfg.Display.PaddingY)
	m.imgLoading[url] = true
	if m.nativeImages() {
		return tea.Batch(
			image.BlockCmd(m.imgCache, m.imgBlocks, m.store, a.ID, url, width, vpH, false, position),
			image.PhotoCmd(m.imgCache, m.imgBlocks, m.imgPhotos, m.store, a.ID, url, width, vpH, position),
		)
	}
	return image.BlockCmd(m.imgCache, m.imgBlocks, m.store, a.ID, url, width, vpH, true, position)
}

// ensureImageSource fires the image load on resize for every image the open
// article renders (the lead and each inline image) that has no cached source
// at the current content geometry (for example a stored block rendered at a
// different width, or a native render keyed to an earlier viewport size).
func (m *Model) ensureImageSource(a store.Article) tea.Cmd {
	if !m.imagesEnabled() {
		return nil
	}
	width, vpH, _ := contentGeom(m.width, m.height, m.cfg.Display.PaddingX, m.cfg.Display.PaddingY)
	var cmds []tea.Cmd
	for pos, url := range m.article.imageURLs {
		if url == "" || m.imgLoading[url] {
			continue
		}
		if m.hasImageSource(url, width, vpH) {
			continue
		}
		cmds = append(cmds, m.loadImageCmd(a, url, pos))
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// hasImageSource reports whether a cached source renders the image at the given
// content geometry: a native render at that size, a fetched photo (native), a
// decoded image, or a width-matched stored block.
func (m *Model) hasImageSource(url string, contentW, vpH int) bool {
	if _, ok := m.imgNatives.Get(image.NativeKey(url, contentW, vpH)); ok {
		return true
	}
	if _, ok := m.imgPhotos.Get(url); ok {
		return true
	}
	if _, ok := m.imgCache.Get(url); ok {
		return true
	}
	if lines, ok := m.imgBlocks.Get(url); ok && image.BlockWidth(lines) == contentW {
		return true
	}
	return false
}

// recomposeArticle re-composes the open article, preserving the user's reading
// position: if the offset was already past the insertion point (the top of the
// content), it is bumped by the number of inserted image lines and then snapped
// so a load that landed while the user was reading at its position cannot leave
// the offset inside a photo range or with an image's top materialized
// mid-window.
func (m *Model) recomposeArticle() {
	oldOffset := m.article.viewport.YOffset
	oldLen := len(m.article.lines)
	vpH := m.article.viewport.Height
	// A user reading at the article's last screen (GotoBottom, or scrolled to
	// the end) stays at the bottom: re-composition must not let an image load
	// snap the offset up to a photo top that happens to sit mid-window, since
	// that would hide the article's end the user asked to see. The offset bump
	// below already tracks the bottom through the inserted lines; the snap
	// would be the only thing to move it.
	atBottom := oldOffset > 0 && oldOffset >= oldLen-vpH
	a := *m.article.article
	m.article = m.newArticleState(a)
	if atBottom {
		m.article.viewport.GotoBottom()
		return
	}
	inserted := len(m.article.lines) - oldLen
	if oldOffset > 0 {
		m.article.viewport.SetYOffset(snapAfterRecompose(oldOffset+inserted, vpH, m.article.imageBlocks))
	} else {
		m.article.viewport.SetYOffset(0)
	}
}

// snapAfterRecompose repositions the offset after a re-composition so an image
// load that lands while the user is reading at its position cannot leave the
// offset inside a photo range or with an image's top materialized mid-window.
// An offset inside the lead photo snaps onto its caption (the lead was shown on
// open); an offset inside an inline photo, or below an inline image whose top
// sits within the window, snaps forward to the inline image's top (motion 1),
// matching the two-motion scroll contract so the image is shown rather than
// left partially clipped.
func snapAfterRecompose(offset, vpH int, blocks []imageBlock) int {
	for i, b := range blocks {
		if offset >= b.imgStart && offset < b.capStart {
			if i == 0 {
				return b.capStart
			}
			return b.imgStart
		}
	}
	// An inline image whose top materialized inside the window below the
	// offset snaps forward to its top (the nearest such image), so it is shown
	// rather than scrolled past unseen. The lead block is excluded: its top
	// sitting below the header is the normal open state and the header must
	// stay visible.
	best := -1
	for i, b := range blocks {
		if i > 0 && b.imgStart > offset && b.imgStart < offset+vpH {
			if best < 0 || b.imgStart < blocks[best].imgStart {
				best = i
			}
		}
	}
	if best >= 0 {
		return blocks[best].imgStart
	}
	return offset
}

// onBlockLoaded re-composes the open article when the message's URL is one of
// its images, preserving the reading position.
func (m *Model) onBlockLoaded(msg image.BlockMsg) {
	m.imgLoading[msg.Key] = false
	m.imgBlocks.Set(msg.Key, msg.Lines)
	if m.view != viewArticle {
		return
	}
	if !m.article.rendersImage(msg.Key) {
		return
	}
	m.recomposeArticle()
}

// inlineAttrFor returns the composed caption for an inline image by URL: the
// caption recorded when the block was composed (alt, else figure credit, else
// link text, else the source fallback), so the native render's height fit
// reserves exactly the lines the composed block carries. It falls back to the
// per-image derivation when the article state has no recorded caption. It
// returns "" when url is not one of the open article's inline images.
func (m *Model) inlineAttrFor(url string) string {
	if cap := m.article.inlineCaps[url]; cap != "" {
		return cap
	}
	for _, im := range m.article.inlineImages {
		if im.URL == url {
			return inlineAttribution(*m.article.article, url, im.Alt, "")
		}
	}
	return ""
}

// nativeRenderCmd returns a tea.Cmd that renders the open article's photo at
// url to a native block for the current content geometry, off the UI goroutine.
// It reuses the article state's header line count so the height fit is
// consistent with the composition. It returns nil when no article or native
// renderer is active, or the URL is empty.
func (m *Model) nativeRenderCmd(url string) tea.Cmd {
	if m.view != viewArticle || m.article.article == nil || !m.nativeImages() || url == "" {
		return nil
	}
	a := *m.article.article
	attr := articleAttribution(a)
	if url != a.ImageURL {
		attr = m.inlineAttrFor(url)
	}
	width, vpH, _ := contentGeom(m.width, m.height, m.cfg.Display.PaddingX, m.cfg.Display.PaddingY)
	return image.NativeCmd(m.imgNative, m.imgPhotos, m.imgNatives, url, width, vpH, attr, m.article.headerLines, captionWidth(width))
}

// onPhotoLoaded fires the off-thread native render for the photo at msg's URL
// when it belongs to the open article, replacing the on-thread render the UI
// used to perform. The resulting NativeMsg re-composes the article with the
// cached native block.
func (m *Model) onPhotoLoaded(msg image.PhotoMsg) tea.Cmd {
	m.imgLoading[msg.Key] = false
	if m.view != viewArticle {
		return nil
	}
	if !m.article.rendersImage(msg.Key) {
		return nil
	}
	return m.nativeRenderCmd(msg.Key)
}

// onNativeLoaded re-composes the open article with its freshly rendered native
// block for msg's URL, replacing the placeholder halfblock and preserving the
// reading position. It always clears the in-flight guard, even when the article
// changed or the view departed meanwhile (the render is still cached for later).
func (m *Model) onNativeLoaded(msg image.NativeMsg) {
	m.imgLoading[msg.Key] = false
	if m.view != viewArticle {
		return
	}
	if !m.article.rendersImage(msg.Key) {
		return
	}
	m.recomposeArticle()
}

// onImageFailed clears the in-flight guard and, when a block became available
// meanwhile (the native placeholder path), re-composes so the block appears;
// otherwise the article already renders without an image block and the failure
// is never surfaced.
func (m *Model) onImageFailed(msg image.FailedMsg) {
	m.imgLoading[msg.Key] = false
	if m.view != viewArticle {
		return
	}
	if !m.article.rendersImage(msg.Key) {
		return
	}
	if _, ok := m.imgBlocks.Get(msg.Key); ok {
		m.recomposeArticle()
	}
}

// snapPos is the set of canonical snap offsets for one image block at a given
// viewport height: bottom (the last line of the wrapped caption at the
// viewport's last row), top (the image's first line at the viewport's first
// row), exit (the photo scrolled out, the caption at the viewport top — the
// down-exit), and off (the image's top at the fold, the block fully below the
// viewport — the up-exit). The two boundaries, bottom and top, are the snap
// stages; exit and off are where a block leaves the viewport in each
// direction.
type snapPos struct {
	bottom int
	top    int
	exit   int
	off    int
}

func snapPositions(b imageBlock, vpH int) snapPos {
	return snapPos{
		bottom: b.imgEnd - vpH + 1,
		top:    b.imgStart,
		exit:   b.capStart,
		off:    b.imgStart - vpH,
	}
}

// snapYOffset implements the image snap for single-line scroll moves across
// the article's ordered image blocks (lead at index 0, then inline in document
// order). Each block defines two snap boundaries — TOP (its first line at the
// viewport's first row) and BOTTOM (the last line of its wrapped caption at
// the viewport's last row) — and the snap steps between them. The lead block
// skips onto its caption in a single downward press — it was shown on open,
// so its boundaries are pre-consumed — and a fully visible lead photo (window
// top at or above imgStart and window bottom reaching imgEnd) likewise skips
// onto its caption. Inline blocks snap in two stages so an image is never
// skipped without first being shown: a downward move that leaves an inline
// image partially visible in the window — its top inside the window, its last
// line below the fold, whether it entered from below or a skip of the
// preceding image left it there — snaps its BOTTOM boundary so the wrapped
// caption is fully visible, the next downward move snaps its TOP boundary, and
// a move landing in its photo range then skips it onto its caption (EXIT).
// Upward moves mirror it: an image entering from above snaps its TOP boundary,
// the next upward move snaps its BOTTOM boundary, an up move that cuts its
// last line below the fold snaps it fully below the fold (OFF) so it scrolls
// off cleanly, and a move landing in its photo range reveals it (TOP). A
// downward move landing in an inline photo's range entered from above snaps to
// its TOP (the second stage); starting at or inside it skips onto its caption.
// Caption rows scroll normally both ways. On a full-image-capable terminal
// (native), an upward move while the lead photo is fully visible (and the
// offset is not already at the article top) skips straight to the article top
// (0), because re-transmitting a native photo on every header step is
// expensive; halfblock art is cheap, so it still scrolls the header line by
// line. before, the offset before the move, gates each stage so the snaps
// cannot re-fire and loop. Multi-line moves (page, half-page, goto-bottom) do
// not call this function at all: they land wherever they land, even mid-photo.
// An empty block list means no image; offsets outside the ranges (or moves
// that never enter them) are unchanged.
func snapYOffset(yOffset, vpH, before int, blocks []imageBlock, direction int, native bool) int {
	if len(blocks) == 0 {
		return yOffset
	}
	pos := make([]snapPos, len(blocks))
	for i, b := range blocks {
		pos[i] = snapPositions(b, vpH)
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
			if yOffset >= b.imgStart && yOffset < b.capStart {
				if i > 0 && before < b.imgStart {
					return snap(pos[i].top)
				}
				return snap(pos[i].exit)
			}
		}
		// A fully visible lead photo (index 0) is skipped in one press: it was
		// shown on open. Inline blocks are excluded — the fully-visible skip
		// there would preempt the two-stage snap and skip an image that was
		// never shown.
		if len(blocks) > 0 {
			if yOffset <= blocks[0].imgStart && yOffset+vpH >= blocks[0].imgEnd {
				return snap(pos[0].exit)
			}
		}
		// Rise stage: an inline image sitting fully visible at its BOTTOM
		// boundary (the bottom snap landed here) moves to its TOP boundary on
		// the next down move.
		for i, b := range blocks {
			if i > 0 && before == pos[i].bottom && yOffset <= b.imgStart {
				return snap(pos[i].top)
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
			if i > 0 && b.imgStart >= yOffset && b.imgStart < yOffset+vpH && b.imgEnd > yOffset+vpH {
				if best < 0 || b.imgStart > blocks[best].imgStart {
					best = i
				}
			}
		}
		if best >= 0 {
			return snap(pos[best].bottom)
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
			if i > 0 && b.imgStart-yOffset == 1 && b.imgStart == blocks[i-1].imgEnd+2 {
				return snap(b.imgStart)
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
			if yOffset > 0 && yOffset <= b.imgStart && yOffset+vpH >= b.imgEnd {
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
				// photo range, leaving it partially visible; reveal that image
				// instead so the transition stays clean.
				for j, nb := range blocks {
					if j > 0 && j < i && t >= nb.imgStart && t < nb.capStart {
						return snap(pos[j].top)
					}
				}
				if t < 0 {
					t = 0
				}
				return snap(t)
			}
		}
		// Up landing in a photo range reveals the image at its TOP boundary.
		for i, b := range blocks {
			if yOffset >= b.imgStart && yOffset < b.capStart {
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
			if i > 0 && before == b.imgStart && yOffset >= b.imgEnd-vpH {
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
				for j, nb := range blocks {
					if j > 0 && j < i && off >= nb.imgStart && off < nb.capStart {
						return snap(pos[j].top)
					}
				}
				return snap(off)
			}
		}
		// Top snap (up): the inline image whose bottom just entered the
		// viewport from above (it was entirely above the window before the
		// move) aligns its TOP boundary to the viewport top. before gates the
		// entry so the snap cannot re-fire and loop.
		best := -1
		for i, b := range blocks {
			if i > 0 && b.imgEnd > yOffset && b.imgEnd <= before {
				if best < 0 || b.imgEnd < blocks[best].imgEnd {
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
			if i > 0 && b.imgStart >= yOffset && b.imgStart < yOffset+vpH && b.imgEnd > yOffset+vpH {
				if best < 0 || b.imgStart > blocks[best].imgStart {
					best = i
				}
			}
		}
		if best >= 0 {
			target := pos[best].off
			// With images spaced closer than the viewport height, the
			// scroll-off target can land inside the previous inline image's
			// photo range, leaving it partially visible (blank); reveal that
			// image instead so the transition stays clean.
			for i, b := range blocks {
				if i > 0 && i < best && target >= b.imgStart && target < b.capStart {
					return snap(pos[i].top)
				}
			}
			return snap(target)
		}
		return yOffset
	}
	return yOffset
}

// scrollArticle runs a viewport scroll operation then, for single-line moves
// (snap true), applies the two-stage image snap: down brings an inline image in
// at the viewport bottom then moves it to the top, up mirrors it, and a move
// that lands inside a photo range skips it onto its caption (or reveals it).
// Multi-line moves (page, half-page, goto-bottom) pass snap false and land
// wherever they land, even mid-photo, so paging never skips the text between
// images.
func (m *Model) scrollArticle(fn func(), snap bool) {
	before := m.article.viewport.YOffset
	fn()
	after := m.article.viewport.YOffset
	if !snap {
		return
	}
	dir := 0
	if after > before {
		dir = 1
	} else if after < before {
		dir = -1
	}
	snapped := snapYOffset(after, m.article.viewport.Height, before, m.article.imageBlocks, dir, m.article.nativeImg)
	if snapped != after {
		m.article.viewport.SetYOffset(snapped)
	}
}
