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
// current frame for every native image the terminal holds that the frame does
// not transmit. kitty graphics placements float above text and the erase
// commands a frame repaint issues have no effect on them, so a placement
// survives scrolling the photo's rows out of the viewport unless explicitly
// deleted. Each block records the kitty image id it was rendered under, and
// nativeSent tracks the id last transmitted per URL; a frame deletes — by that
// exact id — every held image it does not re-transmit: scrolled-out or clipped
// placements (a partially visible image would draw past the window and overlap
// the article border, so it is deleted and its transmit suppressed by
// suppressClippedNativeTransmits), blocks superseded by a re-render at a new
// size (the prior id is deleted before the new transmit, so the terminal never
// holds two sizes of one image), and blocks that reverted to a placeholder.
// The sequence is empty for protocols whose images are cell-bound (OSC 1337) or
// absent, and for fully contained images (that frame re-transmits the image
// with its stable placement id, which replaces the placement). Leaving the
// article view frees every held image by its id — delete-all only clears
// visible placements and can leave the terminal's cached image data to
// accumulate — and an article with no composed native image falls back to the
// delete-all clear so the previous article's photos cannot linger.
func (m *Model) nativeImageClear() string {
	if m.imgNative.Protocol != image.ProtocolKitty {
		return ""
	}
	if m.view != viewArticle {
		var sb strings.Builder
		for _, id := range m.nativeSent {
			sb.WriteString(image.DeleteByID(id))
		}
		m.nativeSent = make(map[string]uint32)
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
		// it was rendered under.
		sb.WriteString(image.DeleteByID(b.NativeID))
	}
	if seen == 0 {
		m.nativeSent = make(map[string]uint32)
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
	img, ok := m.imgCache.Get(b.URL)
	if !ok {
		return
	}
	rows := b.CapStart - b.ImgStart
	pre, err := m.imgRenderer.Render(img, vp.Width, rows)
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
// the off-thread NativeCmd has finished; otherwise it renders the halfblock
// block from the in-memory decoded image, or serves the stored text block when
// its width matches the content width. It returns nil when images are
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
		if lines, ok := m.imgNatives.Get(image.NativeKey(url, contentW, vpH)); ok {
			id := image.NativeRenderID(lines)
			return m.composeBlock(lines, attr, contentW), len(lines), true, id
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
		return block, rows, false, 0
	}
	if lines, ok := m.imgBlocks.Get(url); ok && image.BlockWidth(lines) == contentW {
		return m.composeBlock(lines, attr, contentW), len(lines), false, 0
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
	captionW := compose.CaptionWidth(contentW)
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
		captionW := compose.CaptionWidth(contentW)
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
	width, vpH, _ := render.ContentGeom(m.width, m.height, m.cfg.Display.PaddingX, m.cfg.Display.PaddingY)
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
	width, vpH, _ := render.ContentGeom(m.width, m.height, m.cfg.Display.PaddingX, m.cfg.Display.PaddingY)
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
		m.article.viewport.SetYOffset(compose.SnapAfterRecompose(oldOffset+inserted, vpH, m.article.imageBlocks))
	} else {
		m.article.viewport.SetYOffset(0)
	}
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
			return compose.InlineAttribution(*m.article.article, url, im.Alt, "")
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
	attr := compose.ArticleAttribution(a)
	if url != a.ImageURL {
		attr = m.inlineAttrFor(url)
	}
	width, vpH, _ := render.ContentGeom(m.width, m.height, m.cfg.Display.PaddingX, m.cfg.Display.PaddingY)
	return image.NativeCmd(m.imgNative, m.imgPhotos, m.imgNatives, url, width, vpH, attr, m.article.headerLines, compose.CaptionWidth(width))
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
	snapped := compose.SnapYOffset(after, m.article.viewport.Height, before, m.article.imageBlocks, dir, m.article.nativeImg)
	if snapped != after {
		m.article.viewport.SetYOffset(snapped)
	}
}
