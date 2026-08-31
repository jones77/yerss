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
// current frame when the frame does not display the open article's native
// image. kitty graphics placements float above text and the erase commands a
// frame repaint issues have no effect on them, so a placement survives
// leaving the article view, opening another article, or scrolling the photo's
// rows out of the viewport; every frame that no longer shows the photo deletes
// the terminal's visible placements so stale photos cannot persist or stack.
// The sequence is empty for protocols whose images are cell-bound (OSC 1337)
// or absent, and when the photo's rows [imgStart, capStart) intersect the
// visible viewport window (that frame re-transmits the image with its stable
// placement id, which replaces the placement). The caption may remain visible
// while the photo is scrolled out; such a frame does not render the transmit
// line, so it deletes the placement.
func (m *Model) nativeImageClear() string {
	if m.imgNative.Protocol != image.ProtocolKitty {
		return ""
	}
	if m.view != viewArticle || !m.article.nativeImg {
		return m.imgNative.Clear()
	}
	vp := m.article.viewport
	if m.article.imgStart < vp.YOffset+vp.Height && m.article.capStart > vp.YOffset {
		return ""
	}
	return m.imgNative.Clear()
}

// articleImageBlock renders the article's lead image as content lines with a
// dim attribution centered directly beneath. On a native-capable terminal it
// serves a previously-rendered native photo block (keyed by the current content
// width and viewport height) once the off-thread NativeCmd has finished;
// otherwise it renders the halfblock block from the in-memory decoded image, or
// serves the stored text block when its width matches the content width. It
// returns nil when images are disabled, no URL is set, or no image source is
// available yet; the int reports how many of the returned lines are image rows
// (the attribution lines follow them) and the bool reports whether the lines
// carry a native inline-image escape (a terminal-side placement that outlives
// the frame).
func (m *Model) articleImageBlock(a store.Article, contentW, vpH, headerLines int) ([]string, int, bool) {
	if !m.imagesEnabled() || a.ImageURL == "" {
		return nil, 0, false
	}
	attr := articleAttribution(a)
	if m.nativeImages() {
		if lines, ok := m.imgNatives.Get(image.NativeKey(a.ImageURL, contentW, vpH)); ok {
			return m.composeBlock(lines, attr, contentW), len(lines), true
		}
	}
	if img, ok := m.imgCache.Get(a.ImageURL); ok {
		block, rows := m.fitBlock(func(maxH int) []string {
			lines, err := m.imgRenderer.Render(img, contentW, maxH)
			if err != nil {
				return nil
			}
			return lines
		}, attr, contentW, vpH, headerLines)
		return block, rows, false
	}
	if lines, ok := m.imgBlocks.Get(a.ImageURL); ok && image.BlockWidth(lines) == contentW {
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
	// body line. The attribution line count is unknown until the block's
	// width is known (it wraps to the block), so render optimistically
	// reserving one attribution line, then re-render smaller if the wrapped
	// attribution needs more.
	base := vpH - headerLines - 3
	maxH := max(1, base-1)
	if attr == "" {
		maxH = max(1, base)
	}
	var rendered []string
	for i := 0; ; i++ {
		rendered = renderFn(maxH)
		if rendered == nil {
			return nil, 0
		}
		imgW := image.BlockWidth(rendered)
		attrLines := 0
		if attr != "" {
			attrLines = len(image.LayoutWrapLines(attr, imgW))
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
// wrapped attribution lines, each centered beneath the block.
func (m *Model) composeBlock(rendered []string, attr string, contentW int) []string {
	imgW := image.BlockWidth(rendered)
	photoPad := max(0, (contentW-imgW)/2)
	out := make([]string, 0, len(rendered)+2)
	for _, l := range rendered {
		out = append(out, strings.Repeat(" ", photoPad)+l)
	}
	if attr != "" {
		style := lipgloss.NewStyle().Foreground(m.palette.Dim)
		for _, al := range image.LayoutWrapLines(attr, imgW) {
			pad := photoPad + max(0, (imgW-ansi.StringWidth(al))/2)
			out = append(out, strings.Repeat(" ", pad)+style.Render(al))
		}
	}
	return out
}

// articleAttribution returns the photo attribution for the article's lead
// image: the credit extracted from the article's HTML when present, otherwise
// "photo: <source>" derived by the list view's source-identifier rules. It
// returns "" when neither is available.
func articleAttribution(a store.Article) string {
	if credit := convert.ImageCredit(a.Content, a.ImageURL); credit != "" {
		return credit
	}
	if src := sourceID(a.Link, a.FeedURL); src != "" {
		return "photo: " + src
	}
	return ""
}

// fireImageLoad returns a tea.Cmd that loads the article's lead image through
// the cache hierarchy when image rendering is enabled and a URL is present. On
// a native-capable terminal it fires both the block placeholder command and
// the photo fetch command; otherwise it fires the block command with network
// fetching enabled.
func (m *Model) fireImageLoad(a store.Article) tea.Cmd {
	if !m.imagesEnabled() || a.ImageURL == "" {
		return nil
	}
	if m.imgLoading[a.ImageURL] {
		return nil
	}
	return m.loadImageCmd(a)
}

// loadImageCmd starts the image load for a, marking the URL in flight so
// duplicate loads are not started (for example on resize).
func (m *Model) loadImageCmd(a store.Article) tea.Cmd {
	width, vpH, _ := contentGeom(m.width, m.height, m.cfg.Display.PaddingX, m.cfg.Display.PaddingY)
	m.imgLoading[a.ImageURL] = true
	if m.nativeImages() {
		return tea.Batch(
			image.BlockCmd(m.imgCache, m.imgBlocks, m.store, a.ID, a.ImageURL, width, vpH, false),
			image.PhotoCmd(m.imgCache, m.imgBlocks, m.imgPhotos, m.store, a.ID, a.ImageURL, width, vpH),
		)
	}
	return image.BlockCmd(m.imgCache, m.imgBlocks, m.store, a.ID, a.ImageURL, width, vpH, true)
}

// ensureImageSource fires the image load on resize when the article has a lead
// image URL and no cached source renders at the current content geometry (for
// example a stored block rendered at a different width, or a native render keyed
// to an earlier viewport size).
func (m *Model) ensureImageSource(a store.Article) tea.Cmd {
	if !m.imagesEnabled() || a.ImageURL == "" {
		return nil
	}
	url := a.ImageURL
	if m.imgLoading[url] {
		return nil
	}
	width, vpH, _ := contentGeom(m.width, m.height, m.cfg.Display.PaddingX, m.cfg.Display.PaddingY)
	if m.hasImageSource(url, width, vpH) {
		return nil
	}
	return m.loadImageCmd(a)
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
// content), it is bumped by the number of inserted image lines.
func (m *Model) recomposeArticle() {
	oldOffset := m.article.viewport.YOffset
	oldLen := len(m.article.lines)
	a := *m.article.article
	m.article = m.newArticleState(a)
	inserted := len(m.article.lines) - oldLen
	if oldOffset > 0 {
		m.article.viewport.SetYOffset(oldOffset + inserted)
	} else {
		m.article.viewport.SetYOffset(0)
	}
}

// onBlockLoaded re-composes the open article with its freshly available block,
// preserving the reading position.
func (m *Model) onBlockLoaded(msg image.BlockMsg) {
	m.imgLoading[msg.Key] = false
	m.imgBlocks.Set(msg.Key, msg.Lines)
	if m.view != viewArticle {
		return
	}
	a := *m.article.article
	if msg.Key != a.ImageURL {
		return
	}
	m.recomposeArticle()
}

// nativeRenderCmd returns a tea.Cmd that renders the open article's lead photo
// to a native block for the current content geometry, off the UI goroutine. It
// reuses the article state's header line count so the height fit is consistent
// with the composition. It returns nil when no article or native renderer is
// active.
func (m *Model) nativeRenderCmd() tea.Cmd {
	if m.view != viewArticle || m.article.article == nil || !m.nativeImages() {
		return nil
	}
	a := *m.article.article
	if a.ImageURL == "" {
		return nil
	}
	width, vpH, _ := contentGeom(m.width, m.height, m.cfg.Display.PaddingX, m.cfg.Display.PaddingY)
	return image.NativeCmd(m.imgNative, m.imgPhotos, m.imgNatives, a.ImageURL, width, vpH, articleAttribution(a), m.article.headerLines)
}

// onPhotoLoaded fires the off-thread native render for the open article's
// fetched photo, replacing the on-thread render the UI used to perform. The
// resulting NativeMsg re-composes the article with the cached native block.
func (m *Model) onPhotoLoaded(msg image.PhotoMsg) tea.Cmd {
	m.imgLoading[msg.Key] = false
	if m.view != viewArticle {
		return nil
	}
	a := *m.article.article
	if msg.Key != a.ImageURL {
		return nil
	}
	return m.nativeRenderCmd()
}

// onNativeLoaded re-composes the open article with its freshly rendered native
// block, replacing the placeholder halfblock and preserving the reading
// position. It always clears the in-flight guard, even when the article changed
// or the view departed meanwhile (the render is still cached for later).
func (m *Model) onNativeLoaded(msg image.NativeMsg) {
	m.imgLoading[msg.Key] = false
	if m.view != viewArticle {
		return
	}
	a := *m.article.article
	if msg.Key != a.ImageURL {
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
	a := *m.article.article
	if msg.Key != a.ImageURL {
		return
	}
	if _, ok := m.imgBlocks.Get(msg.Key); ok {
		m.recomposeArticle()
	}
}

// snapYOffset implements atomic image scroll: after a downward move that either
// lands the offset in the photo's line range [imgStart, capStart-1] or leaves
// the full photo block on screen (window top at or above imgStart and window
// bottom reaching imgEnd), a downward move skips the photo onto the first
// non-photo line (capStart, the first attribution line, or the line past the
// block when no attribution is composed). An upward move landing within the
// photo rows [imgStart, capStart-1] reveals the full image (imgStart), while
// landing on a caption line keeps the caption scrolling upward line by line
// (unchanged). On a full-image-capable terminal (native), an upward move while
// the photo is fully visible (and the offset is not already at the article top)
// skips straight to the article top (0), because re-transmitting a native photo
// on every header step is expensive; halfblock art is cheap, so it still scrolls
// the header line by line. An imgStart of -1 marks "no image block". Offsets
// outside the ranges (or moves that never enter them) are unchanged.
func snapYOffset(yOffset, vpH, imgStart, imgEnd, capStart, direction int, native bool) int {
	if imgStart < 0 || imgEnd < imgStart {
		return yOffset
	}
	if direction > 0 {
		if yOffset >= imgStart && yOffset < capStart {
			return capStart
		}
		if yOffset <= imgStart && yOffset+vpH >= imgEnd {
			return capStart
		}
	}
	if direction < 0 {
		if native && yOffset > 0 && yOffset <= imgStart && yOffset+vpH >= imgEnd {
			return 0
		}
		if yOffset >= imgStart && yOffset < capStart {
			return imgStart
		}
	}
	return yOffset
}

// scrollArticle runs a viewport scroll operation then snaps the offset out of
// the photo's row range when the move partially clips the photo, using the
// move's direction to decide where to land: down onto the caption (or past
// the block when there is none), up to the full-image reveal (or to the
// article top when a native photo is fully visible).
func (m *Model) scrollArticle(fn func()) {
	before := m.article.viewport.YOffset
	fn()
	after := m.article.viewport.YOffset
	dir := 0
	if after > before {
		dir = 1
	} else if after < before {
		dir = -1
	}
	snapped := snapYOffset(after, m.article.viewport.Height, m.article.imgStart, m.article.imgEnd, m.article.capStart, dir, m.article.nativeImg)
	if snapped != after {
		m.article.viewport.SetYOffset(snapped)
	}
}
