package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/image"
	"yerss/internal/store"
	"yerss/internal/ui/compose"
	"yerss/internal/ui/render"
)

// imgConcurrency bounds the number of concurrent image fetch, decode, and
// render operations (the block, photo, and native commands) via the shared
// imgSem semaphore. It is a small constant independent of the article's image
// count, so a photo-heavy article cannot spawn one operation per image at once.
const imgConcurrency = 3

// imagesEnabled reports whether lead image rendering is active: the config mode
// is not off and ASCII fallback is not forcing images off (the -a flag and
// terminal detection both set m.ascii).
func (m *Model) imagesEnabled() bool {
	return !m.ascii && m.sess.Config().Display.Images != "off"
}

// nativeImages reports whether the terminal can render real photos natively.
func (m *Model) nativeImages() bool {
	return m.sess.ImgNative.Protocol != image.ProtocolNone
}

// recomputeNativeClear computes the native-image clear sequence for the current
// frame and stores it on the model for View() to prepend, moving the kitty
// cleanup into the update path so render stays a pure function of state. kitty
// graphics placements float above text and the erase commands a frame repaint
// issues have no effect on them, so a placement survives scrolling the photo's
// rows out of the viewport unless explicitly deleted. Each block records the
// kitty image id it was rendered under, and nativeSent tracks the id last
// transmitted per URL; a frame deletes — by that exact id — every held image it
// does not re-transmit: scrolled-out or clipped placements (a partially visible
// image would draw past the window and overlap the article border, so it is
// deleted and its transmit suppressed by suppressClippedNativeTransmits),
// blocks superseded by a re-render at a new size (the prior id is deleted
// before the new transmit, so the terminal never holds two sizes of one image),
// and blocks that reverted to a placeholder. Leaving the article view frees
// every held image by its id — delete-all only clears visible placements and
// can leave the terminal's cached image data to accumulate; on Ghostty, whose
// by-id delete handling is partial, leaving the article view additionally
// clears every visible placement so none can survive over the list. An article
// with no
// composed native image falls back to the delete-all clear so the previous
// article's photos cannot linger, but only on the transition into that article
// (nativeClearArticleID), not on every no-native-image frame. The sequence is
// empty for protocols whose images are cell-bound (OSC 1337) or absent, and for
// fully contained images (that frame re-transmits the image with its stable
// placement id, which replaces the placement).
func (m *Model) recomputeNativeClear() {
	// Snapshot the transmitted-id tracking before this frame's computeNativeClear
	// advances it, so the render path can tell a placement reference re-show (the
	// image was transmitted on an earlier frame) from this frame's fresh transmit.
	m.nativeSentPrev = make(map[string]uint32, len(m.nativeSent))
	for url, id := range m.nativeSent {
		m.nativeSentPrev[url] = id
	}
	m.nativeClearPrefix = m.computeNativeClear()
}

// computeNativeClear returns the native-image clear sequence for the current
// frame and advances the nativeSent/nativeClearArticleID tracking; it is
// invoked ONLY through recomputeNativeClear so the mutation never reaches
// View(), keeping render a pure function of state.
func (m *Model) computeNativeClear() string {
	if m.sess.ImgNative.Protocol != image.ProtocolKitty {
		return ""
	}
	if m.view != viewArticle {
		var sb strings.Builder
		for _, id := range m.nativeSent {
			sb.WriteString(image.DeleteByID(id))
		}
		m.nativeSent = make(map[string]uint32)
		// Ghostty's by-id delete handling is partial, so a placement can
		// survive the per-id deletes and float over the list; the delete-all
		// clears every visible placement as a belt-and-suspenders guarantee.
		if m.sess.ImgNative.Ghostty {
			sb.WriteString(m.sess.ImgNative.Clear())
		}
		// Leaving the article view: the next entry into a no-native article
		// must re-emit the delete-all.
		m.nativeClearArticleID = -1
		return sb.String()
	}
	vp := m.article.viewport
	var sb strings.Builder
	seen := 0
	for _, b := range m.article.imageBlocks {
		if !b.NativeImg {
			continue
		}
		seen++
		if b.ImgStart >= vp.YOffset && b.CapStart <= vp.YOffset+vp.Height {
			// The frame re-transmits this block under its recorded id. If a
			// prior size's id is still held for the URL, delete it before the
			// transmit, then record the current id for later cleanup.
			if held, ok := m.nativeSent[b.URL]; ok && held != 0 && held != b.NativeID {
				sb.WriteString(image.DeleteByID(held))
			}
			if b.NativeID != 0 {
				m.nativeSent[b.URL] = b.NativeID
			}
			continue
		}
		// The block is scrolled out or clipped: delete its placement by the id
		// it was rendered under, freeing the terminal's cached image data, and
		// forget the transmitted id so the frame that scrolls it back in
		// transmits the payload once before referencing again.
		sb.WriteString(image.DeleteByID(b.NativeID))
		delete(m.nativeSent, b.URL)
	}
	if seen == 0 {
		m.nativeSent = make(map[string]uint32)
		if m.nativeClearArticleID == m.article.id {
			return ""
		}
		m.nativeClearArticleID = m.article.id
		return m.sess.ImgNative.Clear()
	}
	m.nativeClearArticleID = 0
	return sb.String()
}

// rewriteNativeReShows rewrites the first line of every fully visible native
// image whose payload was transmitted on an earlier frame (its id is recorded
// in nativeSentPrev, the state before this frame's transmits) from the full
// transmit escape the cached block carries into a placement reference (`a=p`
// naming the same image id), so later frames re-show the cached photo instead
// of re-transmitting its base64 payload. It is a pure per-frame transform of
// the visible viewport lines: the cached block keeps its full transmit, so a
// recompose or a scroll-out/scroll-back cycle restores it, and render stays a
// pure function of state (the snapshot is set by recomputeNativeClear in the
// update path). A full transmit stays on the first frame, after a scroll-out
// delete cleared the id, and after a re-render at a new size (the prior id was
// deleted and the new id is not yet recorded).
func (m *Model) rewriteNativeReShows(lines []string) {
	if m.sess.ImgNative.Protocol != image.ProtocolKitty {
		return
	}
	vp := m.article.viewport
	for _, b := range m.article.imageBlocks {
		if !b.NativeImg {
			continue
		}
		if b.ImgStart < vp.YOffset || b.CapStart > vp.YOffset+vp.Height {
			continue
		}
		// The image was already transmitted before this frame under exactly
		// this render id, so the terminal holds its payload; re-show it by
		// placement reference.
		if held, ok := m.nativeSentPrev[b.URL]; !ok || held == 0 || held != b.NativeID {
			continue
		}
		if idx := b.ImgStart - vp.YOffset; idx >= 0 && idx < len(lines) {
			lines[idx] = image.KittyPlacementReference(lines[idx])
		}
	}
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
	if m.sess.ImgNative.Protocol != image.ProtocolKitty {
		return
	}
	vp := m.article.viewport
	for _, b := range m.article.imageBlocks {
		if !b.NativeImg {
			continue
		}
		if b.ImgStart >= vp.YOffset && b.CapStart <= vp.YOffset+vp.Height {
			continue
		}
		if idx := b.ImgStart - vp.YOffset; idx >= 0 && idx < len(lines) {
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
func (m *Model) previewClippedImage(lines []string, b compose.ImageBlock) {
	vp := m.article.viewport
	img, ok := m.sess.ImgCache.Get(b.URL)
	if !ok {
		return
	}
	rows := b.CapStart - b.ImgStart
	pre, err := m.sess.ImgRenderer.Render(img, vp.Width, rows)
	if err != nil {
		return
	}
	imgW := image.BlockWidth(pre)
	pad := max(0, (vp.Width-imgW)/2)
	lo := max(b.ImgStart, vp.YOffset)
	hi := min(b.CapStart, vp.YOffset+vp.Height)
	for r := lo; r < hi; r++ {
		idx := r - vp.YOffset
		k := r - b.ImgStart
		if idx < 0 || idx >= len(lines) || k >= len(pre) {
			continue
		}
		lines[idx] = strings.Repeat(" ", pad) + pre[k]
	}
}

// composeImageBlock renders the image at url as content lines with the given
// attribute centered directly beneath, for both the lead image and inline
// images. On a native-capable terminal it serves a previously-rendered native
// photo block (keyed by the current content width and viewport height) once
// the off-thread NativeCmd has finished; otherwise it serves the width-matched
// stored text block, and only when neither a native render nor a stored block
// matches the current width does it re-render the halfblock from the in-memory
// decoded image (a resize or a first load). It returns nil when images are
// disabled, no URL is set, or no image source is available yet; the int
// reports how many of the returned lines are image rows (the attribution lines
// follow them), the bool reports whether the lines carry a native
// inline-image escape (a terminal-side placement that outlives the frame), and
// the uint32 is the kitty image id the native render was transmitted under (0
// for halfblock or OSC 1337 renders), for deleting the exact image on exit or
// when a re-render supersedes it.
func (m *Model) composeImageBlock(url, attr string, contentW, vpH, headerLines int) ([]string, int, bool, uint32) {
	if !m.imagesEnabled() || url == "" {
		return nil, 0, false, 0
	}
	if m.nativeImages() {
		if lines, ok := m.sess.ImgNatives.Get(image.NativeKey(url, contentW, vpH)); ok {
			id := image.NativeRenderID(lines)
			return m.composeBlock(lines, attr, contentW), len(lines), true, id
		}
	}
	// A width-matched rendered block is served before re-rendering from the
	// decoded image, so a recompose after an image-load message never re-runs
	// the halfblock render on the UI thread for an image whose block is already
	// composed at this width. The width guard keeps a stale-width block (a
	// resize re-renders it) out of the serve path: a full-width block must
	// match the content width exactly, while a small image's block — rendered
	// at its content-width-independent capped width — is served when the cached
	// decoded image confirms its width is what a render at the current content
	// width would produce.
	if lines, ok := m.sess.ImgBlocks.Get(url); ok {
		if w := image.BlockWidth(lines); w == contentW {
			return m.composeBlock(lines, attr, contentW), len(lines), false, 0
		} else if img, ok := m.sess.ImgCache.Get(url); ok {
			b := img.Bounds()
			if w == image.RenderWidth(b.Dx(), b.Dy(), contentW) {
				return m.composeBlock(lines, attr, contentW), len(lines), false, 0
			}
		}
	}
	if img, ok := m.sess.ImgCache.Get(url); ok {
		block, rows := m.fitBlock(func(maxH int) []string {
			lines, err := m.sess.ImgRenderer.Render(img, contentW, maxH)
			if err != nil {
				return nil
			}
			return lines
		}, attr, contentW, vpH, headerLines)
		return block, rows, false, 0
	}
	return nil, 0, false, 0
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
	captionW := compose.CaptionWidth(m.width, contentW)
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
		captionW := compose.CaptionWidth(m.width, contentW)
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

// fireImageLoad returns a tea.Cmd that loads the images the open article
// renders at or before the load frontier (the lead and the inline images in or
// near the visible viewport, in document order) through the cache hierarchy
// when image rendering is enabled, with one in-flight guard per URL. Images
// beyond the frontier are not fetched, decoded, or rendered until the reader
// scrolls toward them.
func (m *Model) fireImageLoad(a store.Article) tea.Cmd {
	if !m.imagesEnabled() {
		return nil
	}
	m.advanceFrontier()
	return m.fireImagesTo(a, m.imgFrontier, true)
}

// loadImageCmd starts the image load for the article image at url and position,
// marking the URL in flight so duplicate loads are not started (for example on
// resize), and gating the fetch/decode/render work with the shared concurrency
// bound.
func (m *Model) loadImageCmd(a store.Article, url string, position int) tea.Cmd {
	width, vpH, _ := render.ContentGeom(m.width, m.height, m.sess.Config().Display.PaddingX, m.sess.Config().Display.PaddingY)
	m.imgLoading[url] = true
	if m.nativeImages() {
		return tea.Batch(
			m.gateCmd(image.BlockCmd(m.sess.ImgCache, m.sess.ImgBlocks, m.sess.Store(), a.ID, url, width, vpH, false, position)),
			m.gateCmd(image.PhotoCmd(m.sess.ImgCache, m.sess.ImgBlocks, m.sess.ImgPhotos, m.sess.Store(), a.ID, url, width, vpH, position)),
		)
	}
	return m.gateCmd(image.BlockCmd(m.sess.ImgCache, m.sess.ImgBlocks, m.sess.Store(), a.ID, url, width, vpH, true, position))
}

// ensureImageSource fires the image load on resize for every image the open
// article renders at or before the load frontier that has no cached source at
// the current content geometry (for example a stored block rendered at a
// different width, or a native render keyed to an earlier viewport size).
func (m *Model) ensureImageSource(a store.Article) tea.Cmd {
	if !m.imagesEnabled() {
		return nil
	}
	m.advanceFrontier()
	return m.fireImagesTo(a, m.imgFrontier, false)
}

// advanceFrontier recomputes the load frontier from the open article's anchor
// rows and current viewport: the frontier is the largest image index whose
// anchor row is within the viewport plus one viewport height of lookahead
// below the fold. It only grows within an article open, so a recompose that
// shifts rows (or a scroll that returns toward the top) can never retract a
// load an earlier position already brought in range.
func (m *Model) advanceFrontier() {
	if m.article.article == nil {
		return
	}
	vp := m.article.viewport
	limit := vp.YOffset + 2*vp.Height
	k := -1
	for i, row := range m.article.anchorRows {
		if row >= limit {
			break
		}
		k = i
	}
	if k > m.imgFrontier {
		m.imgFrontier = k
	}
}

// advanceFrontierLoads advances the load frontier to the current viewport and
// returns a tea.Cmd firing the newly in-range image loads and the deferred
// native renders now in view, or nil when nothing is newly in range. It is the
// scroll hook: every viewport move runs it, and the recompose after a load
// landing runs it too so a deferred native render fires the moment its photo's
// block is within reach.
func (m *Model) advanceFrontierLoads() tea.Cmd {
	if !m.imagesEnabled() {
		return nil
	}
	m.advanceFrontier()
	a := m.article.article
	if a == nil {
		return nil
	}
	return m.fireImagesTo(*a, m.imgFrontier, false)
}

// fireImagesTo returns a tea.Cmd firing the image work for the article's image
// URLs at indices 0..k: deferred native renders now in view, and image loads
// through the cache hierarchy. open selects the open-path policy — fire a load
// for every in-range URL so the native render path runs even when a source is
// already available — versus the scroll/resize policy that skips URLs already
// served at the current geometry (their blocks are already composed). imgLoading
// remains the per-URL in-flight guard, so a frontier advance never double-fires
// a URL whose load or render is still running.
func (m *Model) fireImagesTo(a store.Article, k int, open bool) tea.Cmd {
	width, vpH, _ := render.ContentGeom(m.width, m.height, m.sess.Config().Display.PaddingX, m.sess.Config().Display.PaddingY)
	var cmds []tea.Cmd
	for pos, url := range m.article.imageURLs {
		if pos > k || url == "" {
			continue
		}
		if m.nativePending[url] {
			if m.photoBlockInView(url) && !m.imgLoading[url] {
				if nc := m.nativeRenderCmd(url); nc != nil {
					cmds = append(cmds, nc)
					delete(m.nativePending, url)
				}
			}
			continue
		}
		if m.imgLoading[url] {
			continue
		}
		if !open && m.hasImageSource(url, width, vpH) {
			continue
		}
		cmds = append(cmds, m.loadImageCmd(a, url, pos))
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// photoBlockInView reports whether the open article's photo at url has its
// block within the viewport plus the lookahead margin (the same condition the
// load frontier uses), so its native render is worth starting.
func (m *Model) photoBlockInView(url string) bool {
	idx := m.article.urlIndex(url)
	if idx < 0 || idx >= len(m.article.anchorRows) {
		return false
	}
	vp := m.article.viewport
	return m.article.anchorRows[idx] < vp.YOffset+2*vp.Height
}

// gateCmd wraps an image command so it acquires the shared in-flight slot
// before running its fetch/decode/render work and releases it after, bounding
// the number of concurrent operations across the block, photo, and native
// command types. Firing more commands than the bound just queues them; the
// semaphore, not the batch structure, enforces the limit.
func (m *Model) gateCmd(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		m.imgSem <- struct{}{}
		defer func() { <-m.imgSem }()
		return cmd()
	}
}

// hasImageSource reports whether a cached source renders the image at the given
// content geometry: a native render at that size, a fetched photo (native), a
// decoded image, or a width-matched stored block.
func (m *Model) hasImageSource(url string, contentW, vpH int) bool {
	if _, ok := m.sess.ImgNatives.Get(image.NativeKey(url, contentW, vpH)); ok {
		return true
	}
	if _, ok := m.sess.ImgPhotos.Get(url); ok {
		return true
	}
	if _, ok := m.sess.ImgCache.Get(url); ok {
		return true
	}
	if lines, ok := m.sess.ImgBlocks.Get(url); ok && image.BlockWidth(lines) == contentW {
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
	m.recomposeCount++
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
		m.article.viewport.SetYOffset(compose.SnapAfterRecompose(oldOffset+inserted, vpH, m.article.imageBlocks, m.article.leadShown))
	} else {
		m.article.viewport.SetYOffset(0)
	}
}

// flushRecompose runs the pending article recompose, at most once, when the
// flag is set and the article view is still active, then fires the deferred
// native renders (and newly in-range loads) the recompose's fresh anchors
// brought into view. Image-load message handlers set recomposePending; Update
// calls flushRecompose after the message switch so the per-message work stays
// bounded to cache updates and the recompose — which reuses the cached
// derivation — runs once per image message. It returns the tea.Cmd firing that
// deferred work, or nil.
func (m *Model) flushRecompose() tea.Cmd {
	if !m.recomposePending {
		return nil
	}
	m.recomposePending = false
	if m.view == viewArticle {
		m.recomposeArticle()
		return m.advanceFrontierLoads()
	}
	return nil
}

// onBlockLoaded marks the article for recomposition when the message's URL is
// one of its images; the recompose runs via flushRecompose after the Update
// switch. The block cache is always updated so a later open renders the block.
func (m *Model) onBlockLoaded(msg image.BlockMsg) {
	m.imgLoading[msg.Key] = false
	m.sess.ImgBlocks.Set(msg.Key, msg.Lines)
	if m.view != viewArticle {
		return
	}
	if !m.article.rendersImage(msg.Key) {
		return
	}
	m.recomposePending = true
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
			return compose.ResolveImageAttribution(*m.article.article, url, im.Alt, "", false)
		}
	}
	return ""
}

// nativeRenderCmd returns a tea.Cmd that renders the open article's photo at
// url to a native block for the current content geometry, off the UI goroutine.
// It reuses the article state's header line count so the height fit is
// consistent with the composition. The URL is marked in flight so a scroll or
// recompose does not start a duplicate render, and the render work is gated by
// the shared concurrency bound. It returns nil when no article or native
// renderer is active, or the URL is empty.
func (m *Model) nativeRenderCmd(url string) tea.Cmd {
	if m.view != viewArticle || m.article.article == nil || !m.nativeImages() || url == "" {
		return nil
	}
	m.imgLoading[url] = true
	a := *m.article.article
	attr := ""
	if cap := m.inlineAttrFor(url); cap != "" {
		attr = cap
	} else {
		attr = compose.ResolveImageAttribution(a, a.ImageURL, "", "", true)
	}
	width, vpH, _ := render.ContentGeom(m.width, m.height, m.sess.Config().Display.PaddingX, m.sess.Config().Display.PaddingY)
	return m.gateCmd(image.NativeCmd(m.sess.ImgCache, m.sess.ImgNative, m.sess.ImgPhotos, m.sess.ImgNatives, url, width, vpH, attr, m.article.headerLines, compose.CaptionWidth(m.width, width)))
}

// onPhotoLoaded fires the off-thread native render for the photo at msg's URL
// when it belongs to the open article and its block is in or near view,
// replacing the on-thread render the UI used to perform. A photo whose block is
// out of view keeps its halfblock placeholder: the URL is recorded in
// nativePending and the render starts when the reader scrolls it into view. The
// resulting NativeMsg re-composes the article with the cached native block.
func (m *Model) onPhotoLoaded(msg image.PhotoMsg) tea.Cmd {
	m.imgLoading[msg.Key] = false
	if m.view != viewArticle {
		return nil
	}
	if !m.article.rendersImage(msg.Key) {
		return nil
	}
	m.recomposePending = true
	if m.photoBlockInView(msg.Key) {
		return m.nativeRenderCmd(msg.Key)
	}
	m.nativePending[msg.Key] = true
	return nil
}

// onNativeLoaded marks the article for recomposition with its freshly rendered
// native block for msg's URL, replacing the placeholder halfblock; the recompose
// runs via flushRecompose after the Update switch. It always clears the
// in-flight guard, even when the article changed or the view departed meanwhile
// (the render is still cached for later).
func (m *Model) onNativeLoaded(msg image.NativeMsg) {
	m.imgLoading[msg.Key] = false
	if m.view != viewArticle {
		return
	}
	if !m.article.rendersImage(msg.Key) {
		return
	}
	m.recomposePending = true
}

// onImageFailed clears the in-flight guard and, when a block became available
// meanwhile (the native placeholder path), marks the article for recomposition
// so the block appears; otherwise the article already renders without an image
// block and the failure is never surfaced.
func (m *Model) onImageFailed(msg image.FailedMsg) {
	m.imgLoading[msg.Key] = false
	if m.view != viewArticle {
		return
	}
	if !m.article.rendersImage(msg.Key) {
		return
	}
	if _, ok := m.sess.ImgBlocks.Get(msg.Key); ok {
		m.recomposePending = true
	}
}

// scrollArticle runs a viewport scroll operation then, for single-line moves
// (snap true), applies the two-stage image snap: down brings an inline image in
// at the viewport bottom then moves it to the top, up mirrors it, and a move
// that lands inside a photo range skips it onto its caption (or reveals it).
// Multi-line moves (page, half-page, goto-bottom) pass snap false and land
// wherever they land, even mid-photo, so paging never skips the text between
// images. Every move then advances the load frontier and returns the command
// firing the newly in-range image loads and deferred native renders, or nil.
func (m *Model) scrollArticle(fn func(), snap bool) tea.Cmd {
	before := m.article.viewport.YOffset
	fn()
	after := m.article.viewport.YOffset
	// A scroll that cannot move — the article is already at its start or end —
	// is a no-op: no content is newly revealed, so snapping and advancing the
	// load frontier would only fire image work that recomposes an unchanged
	// article, keeping the UI busy after the user has stopped scrolling at the
	// boundary.
	if after == before {
		return nil
	}
	if snap {
		dir := 0
		if after > before {
			dir = 1
		} else if after < before {
			dir = -1
		}
		snapped := compose.SnapYOffset(after, m.article.viewport.Height, before, m.article.imageBlocks, dir, m.article.nativeImg, m.article.leadShown)
		if snapped != after {
			m.article.viewport.SetYOffset(snapped)
		}
	}
	return m.advanceFrontierLoads()
}
