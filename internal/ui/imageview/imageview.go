package imageview

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/app"
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

// Concurrency is the exported image command concurrency bound (imgConcurrency).
const Concurrency = imgConcurrency

// ImageView owns the open article's image state and methods: the load frontier
// and in-flight bookkeeping, the halfblock preview memo, and the native
// transmit/clear sequencing. It is a value field on Model (m.images), so the
// image methods live off the model as one cohesive type. It depends only on
// render/compose and the session; the article state it reads and writes is
// passed in explicitly through the ArticleView interface, and the recompose
// and link-mode-toggle work that re-composes the article is owned by the caller
// (the Model, and later articleview), reached through return signals rather
// than a back reference.
type ImageView struct {
	sess    *app.Session
	palette render.Palette
	ascii   bool
	width   int
	height  int

	// loading guards per-URL in-flight image loads and native renders, so a
	// frontier advance or a resize never double-fires a URL whose load or
	// render is still running.
	loading map[string]bool
	// frontier is the load frontier index into the open article's imageURLs:
	// the largest index whose image's anchor row is within the viewport plus
	// the lookahead margin. Images at or before it are eligible to load;
	// images past it are not fetched, decoded, or rendered until the reader
	// scrolls toward them. It only grows within an article open.
	frontier int
	// nativePending holds the URLs of the open article's photos whose bytes
	// are fetched but whose native render is deferred because the block is out
	// of view; the render fires when the reader scrolls the block into view.
	nativePending map[string]bool
	// sem bounds the number of concurrent image fetch/decode/render operations
	// (block, photo, and native commands) to imgConcurrency, regardless of how
	// many image commands are fired.
	sem chan struct{}
	// previewCache memoizes the halfblock preview rendered for each partially
	// visible native image (keyed by URL, content width, and preview row
	// count), so the clipped-preview path renders each image at most once
	// instead of once per frame while it is partially visible.
	previewCache map[string][]string
	// nativeSent tracks the kitty image id last transmitted per URL, so a
	// frame can delete the prior size's image before a re-render at a new
	// size, and leaving the article can free every image the terminal holds.
	nativeSent map[string]uint32
	// nativeSentPrev snapshots nativeSent as it stood at the start of the
	// current frame's recomputeNativeClear, before computeNativeClear advanced
	// it with this frame's transmits. The render path uses it to tell a
	// placement-reference re-show (the image was transmitted on an earlier
	// frame) from this frame's fresh transmit, so the first frame after a
	// transmit (or a scroll-back or resize) always carries the full payload.
	nativeSentPrev map[string]uint32
	// nativeClearPrefix holds the precomputed kitty image clear sequence for
	// the current frame, computed in the update path so View() stays pure.
	nativeClearPrefix string
	// nativeClearArticleID records the article the delete-all was last emitted
	// for; it gates the delete-all so it fires only on the transition into an
	// article with no native images, not on every no-native-image frame. -1
	// means none emitted yet (initial, or after leaving the article view).
	nativeClearArticleID int64
}

// New builds the image machinery bound to a session.
func New(sess *app.Session, palette render.Palette, ascii bool, width, height int) *ImageView {
	return &ImageView{
		sess:    sess,
		palette: palette,
		ascii:   ascii,
		width:   width,
		height:  height,
		// frontier starts at -1 so the first advance loads every in-range
		// image (the open path fires even URLs without a cached source).
		frontier:       -1,
		loading:        make(map[string]bool),
		nativePending:  make(map[string]bool),
		sem:            make(chan struct{}, imgConcurrency),
		previewCache:   make(map[string][]string),
		nativeSent:     make(map[string]uint32),
		nativeSentPrev: make(map[string]uint32),
		// -1 means no delete-all has been emitted for a no-native article yet,
		// so the first entry into one fires it. Reset to -1 when leaving the
		// article view so re-entering re-clears.
		nativeClearArticleID: -1,
	}
}

// SetSize updates the terminal geometry the image loads and block fits size
// against.
func (v *ImageView) SetSize(width, height int) {
	v.width, v.height = width, height
}

// SetAscii updates the ASCII fallback mode, which disables image rendering
// when forced.
func (v *ImageView) SetAscii(ascii bool) {
	v.ascii = ascii
}

// ResetForNewArticle starts a fresh load frontier and deferred-render set when
// a different article opens; recomposes and resizes of the same article keep
// them so the frontier stays monotonic within the article open.
func (v *ImageView) ResetForNewArticle() {
	v.frontier = -1
	v.nativePending = make(map[string]bool)
	v.previewCache = make(map[string][]string)
}

// ClearNativePending drops the deferred native-render set, so a reopened
// article starts fresh (the photos remain cached, so in-view renders re-fire
// through the open path).
func (v *ImageView) ClearNativePending() {
	v.nativePending = make(map[string]bool)
}

// NativeClearPrefix returns the precomputed kitty image clear sequence for the
// current frame, computed in the update path so View() stays pure.
func (v *ImageView) NativeClearPrefix() string {
	return v.nativeClearPrefix
}

// Enabled reports whether lead image rendering is active: the config mode is
// not off and ASCII fallback is not forcing images off (the -a flag and
// terminal detection both set ascii).
func (v *ImageView) Enabled() bool {
	return !v.ascii && v.sess.Config().Display.Images != "off"
}

// NativeEnabled reports whether the terminal can render real photos natively.
func (v *ImageView) NativeEnabled() bool {
	return v.sess.ImgNative.Protocol != image.ProtocolNone
}

// RecomputeNativeClear computes the native-image clear sequence for the current
// frame and stores it on the image machinery for View() to prepend, moving the
// kitty cleanup into the update path so render stays a pure function of state.
// kitty graphics placements float above text and the erase commands a frame
// repaint issues have no effect on them, so a placement survives scrolling the
// photo's rows out of the viewport unless explicitly deleted. Each block
// records the kitty image id it was rendered under, and nativeSent tracks the
// id last transmitted per URL; a frame deletes — by that exact id — every held
// image it does not re-transmit: scrolled-out or clipped placements (a
// partially visible image would draw past the window and overlap the article
// border, so it is deleted and its transmit suppressed by
// SuppressClippedNativeTransmits), blocks superseded by a re-render at a new
// size (the prior id is deleted before the new transmit, so the terminal never
// holds two sizes of one image), and blocks that reverted to a placeholder.
// Leaving the article view frees every held image by its id — delete-all only
// clears visible placements and can leave the terminal's cached image data to
// accumulate; on Ghostty, whose by-id delete handling is partial, leaving the
// article view additionally clears every visible placement so none can survive
// over the list. An article with no composed native image falls back to the
// delete-all clear so the previous article's photos cannot linger, but only on
// the transition into that article (nativeClearArticleID), not on every
// no-native-image frame. The sequence is empty for protocols whose images are
// cell-bound (OSC 1337) or absent, and for fully contained images (that frame
// re-transmits the image with its stable placement id, which replaces the
// placement). active reports whether the article view is current (false when
// the frame renders the list). It returns whether the native snapshot the
// frame body renders against changed, so the caller can bump its native
// generation counter and invalidate the cached article frame.
func (v *ImageView) RecomputeNativeClear(active bool, view ArticleView) bool {
	// Keep the snapshot the previous frame rendered under, then snapshot the
	// transmitted-id tracking before this frame's computeNativeClear advances
	// it, so the render path can tell a placement reference re-show (the image
	// was transmitted on an earlier frame) from this frame's fresh transmit.
	prev := v.nativeSentPrev
	sentPrev := make(map[string]uint32, len(v.nativeSent))
	for url, id := range v.nativeSent {
		sentPrev[url] = id
	}
	v.nativeSentPrev = sentPrev
	clearPrev := v.nativeClearArticleID
	v.nativeClearPrefix = v.computeNativeClear(active, view)
	// The article frame body renders against nativeSentPrev, so the frame cache
	// is valid only while that snapshot matches the one its body was rendered
	// under. A snapshot change — the transition from a fresh transmit to a
	// re-show by placement reference, a scroll-out delete, a re-render at a new
	// size, or the re-entry into a no-native article — is reported so the caller
	// re-renders against the new snapshot instead of serving the stale body. In
	// steady state (no-op boundary scrolls, block staying in or out of view) the
	// snapshot is unchanged and the cached frame keeps serving.
	return !v.nativeSentEqual(prev, sentPrev) || clearPrev != v.nativeClearArticleID
}

// nativeSentEqual reports whether two native transmitted-id maps hold the same
// (url -> image id) entries.
func (v *ImageView) nativeSentEqual(a, b map[string]uint32) bool {
	if len(a) != len(b) {
		return false
	}
	for url, id := range a {
		if b[url] != id {
			return false
		}
	}
	return true
}

// computeNativeClear returns the native-image clear sequence for the current
// frame and advances the nativeSent/nativeClearArticleID tracking; it is
// invoked ONLY through RecomputeNativeClear so the mutation never reaches
// View(), keeping render a pure function of state.
func (v *ImageView) computeNativeClear(active bool, view ArticleView) string {
	if v.sess.ImgNative.Protocol != image.ProtocolKitty {
		return ""
	}
	if !active {
		var sb strings.Builder
		for _, id := range v.nativeSent {
			sb.WriteString(image.DeleteByID(id))
		}
		v.nativeSent = make(map[string]uint32)
		// Ghostty's by-id delete handling is partial, so a placement can
		// survive the per-id deletes and float over the list; the delete-all
		// clears every visible placement as a belt-and-suspenders guarantee.
		if v.sess.ImgNative.Ghostty {
			sb.WriteString(v.sess.ImgNative.Clear())
		}
		// Leaving the article view: the next entry into a no-native article
		// must re-emit the delete-all.
		v.nativeClearArticleID = -1
		return sb.String()
	}
	vp := view.ViewportY()
	vpH := view.ViewportHeight()
	var sb strings.Builder
	seen := 0
	for _, b := range view.ImageBlocks() {
		if !b.NativeImg {
			continue
		}
		seen++
		if b.ImgStart >= vp && b.CapStart <= vp+vpH {
			// The frame re-transmits this block under its recorded id. If a
			// prior size's id is still held for the URL, delete it before the
			// transmit, then record the current id for later cleanup.
			if held, ok := v.nativeSent[b.URL]; ok && held != 0 && held != b.NativeID {
				sb.WriteString(image.DeleteByID(held))
			}
			if b.NativeID != 0 {
				v.nativeSent[b.URL] = b.NativeID
			}
			continue
		}
		// The block is scrolled out or clipped: delete its placement by the id
		// it was rendered under, freeing the terminal's cached image data, and
		// forget the transmitted id so the frame that scrolls it back in
		// transmits the payload once before referencing again.
		sb.WriteString(image.DeleteByID(b.NativeID))
		delete(v.nativeSent, b.URL)
	}
	if seen == 0 {
		v.nativeSent = make(map[string]uint32)
		if v.nativeClearArticleID == view.ID() {
			return ""
		}
		v.nativeClearArticleID = view.ID()
		return v.sess.ImgNative.Clear()
	}
	v.nativeClearArticleID = 0
	return sb.String()
}

// RewriteNativeReShows rewrites the first line of every fully visible native
// image whose payload was transmitted on an earlier frame (its id is recorded
// in nativeSentPrev, the state before this frame's transmits) from the full
// transmit escape the cached block carries into a placement reference (`a=p`
// naming the same image id), so later frames re-show the cached photo instead
// of re-transmitting its base64 payload. It is a pure per-frame transform of
// the visible viewport lines: the cached block keeps its full transmit, so a
// recompose or a scroll-out/scroll-back cycle restores it, and render stays a
// pure function of state (the snapshot is set by RecomputeNativeClear in the
// update path). A full transmit stays on the first frame, after a scroll-out
// delete cleared the id, and after a re-render at a new size (the prior id was
// deleted and the new id is not yet recorded).
func (v *ImageView) RewriteNativeReShows(lines []string, view ArticleView) {
	if v.sess.ImgNative.Protocol != image.ProtocolKitty {
		return
	}
	vp := view.ViewportY()
	vpH := view.ViewportHeight()
	for _, b := range view.ImageBlocks() {
		if !b.NativeImg {
			continue
		}
		if b.ImgStart < vp || b.CapStart > vp+vpH {
			continue
		}
		// The image was already transmitted before this frame under exactly
		// this render id, so the terminal holds its payload; re-show it by
		// placement reference.
		if held, ok := v.nativeSentPrev[b.URL]; !ok || held == 0 || held != b.NativeID {
			continue
		}
		if idx := b.ImgStart - vp; idx >= 0 && idx < len(lines) {
			lines[idx] = image.KittyPlacementReference(lines[idx])
		}
	}
}

// SuppressClippedNativeTransmits blanks the native-image transmit escape on the
// first line of each native image that is only partially visible in the article
// viewport (its top above the fold or its bottom below it). The placement for
// such an image is deleted by nativeImageClear, but the transmit escape carried
// in the block's first line would re-create it every frame and draw the image
// over the article border; stripping the escape leaves the row blank so the
// partially visible image never paints outside the content area. Fully
// contained images keep their transmit.
func (v *ImageView) SuppressClippedNativeTransmits(lines []string, view ArticleView) {
	if v.sess.ImgNative.Protocol != image.ProtocolKitty {
		return
	}
	vp := view.ViewportY()
	vpH := view.ViewportHeight()
	for _, b := range view.ImageBlocks() {
		if !b.NativeImg {
			continue
		}
		if b.ImgStart >= vp && b.CapStart <= vp+vpH {
			continue
		}
		if idx := b.ImgStart - vp; idx >= 0 && idx < len(lines) {
			lines[idx] = ansi.Strip(lines[idx])
		}
		v.previewClippedImage(lines, b, view)
	}
}

// previewClippedImage fills the visible rows of a partially visible native
// image with its halfblock preview, so the portion of a photo that is in the
// window shows a blocky preview rather than a blank strip (the native placement
// is deleted for a partially visible image because it would draw over the
// article border). The preview renders the cached decoded image at the native
// block's row count, so it aligns with the native's cell box and centers the
// same way. It is memoized per (URL, content width, preview row count), so a
// partially visible image is rendered at most once rather than once per frame
// while the reader scrolls it through the fold — the synchronous UI-thread
// render stays bounded by the image count, not by frame count. When the decoded
// image is not cached it is decoded from the raw photo bytes on demand (and the
// result cached), because the native path serves a width-matched stored block
// and reads the raw photo without ever decoding it into the image cache; only
// when neither is available do the rows stay blank.
func (v *ImageView) previewClippedImage(lines []string, b compose.ImageBlock, view ArticleView) {
	vp := view.ViewportY()
	vpH := view.ViewportHeight()
	vpW := view.ViewportWidth()
	img, ok := v.sess.ImgCache.Get(b.URL)
	if !ok {
		if data, dok := v.sess.ImgPhotos.Get(b.URL); dok {
			if decoded, derr := image.DecodeCapped(data); derr == nil {
				v.sess.ImgCache.Set(b.URL, decoded)
				img, ok = decoded, true
			}
		}
	}
	if !ok {
		return
	}
	rows := b.CapStart - b.ImgStart
	key := fmt.Sprintf("%s|%d|%d", b.URL, vpW, rows)
	pre, ok := v.previewCache[key]
	if !ok {
		var err error
		pre, err = v.sess.ImgRenderer.Render(img, vpW, rows)
		if err != nil {
			return
		}
		v.previewCache[key] = pre
	}
	imgW := image.BlockWidth(pre)
	pad := max(0, (vpW-imgW)/2)
	lo := max(b.ImgStart, vp)
	hi := min(b.CapStart, vp+vpH)
	for r := lo; r < hi; r++ {
		idx := r - vp
		k := r - b.ImgStart
		if idx < 0 || idx >= len(lines) || k >= len(pre) {
			continue
		}
		lines[idx] = strings.Repeat(" ", pad) + pre[k]
	}
}

// ComposeBlock renders the image at url as content lines with the given
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
func (v *ImageView) ComposeBlock(url, attr string, contentW, vpH, headerLines int) ([]string, int, bool, uint32) {
	if !v.Enabled() || url == "" {
		return nil, 0, false, 0
	}
	if v.NativeEnabled() {
		if lines, ok := v.sess.ImgNatives.Get(image.NativeKey(url, contentW, vpH)); ok {
			id := image.NativeRenderID(lines)
			return v.composeBlock(lines, attr, contentW), len(lines), true, id
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
	if lines, ok := v.sess.ImgBlocks.Get(url); ok {
		if w := image.BlockWidth(lines); w == contentW {
			return v.composeBlock(lines, attr, contentW), len(lines), false, 0
		} else if img, ok := v.sess.ImgCache.Get(url); ok {
			b := img.Bounds()
			if w == image.RenderWidth(b.Dx(), b.Dy(), contentW) {
				return v.composeBlock(lines, attr, contentW), len(lines), false, 0
			}
		}
	}
	if img, ok := v.sess.ImgCache.Get(url); ok {
		block, rows := v.fitBlock(func(maxH int) []string {
			lines, err := v.sess.ImgRenderer.Render(img, contentW, maxH)
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
func (v *ImageView) fitBlock(renderFn func(maxH int) []string, attr string, contentW, vpH, headerLines int) ([]string, int) {
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
	captionW := compose.CaptionWidth(v.width, contentW)
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
	return v.composeBlock(rendered, attr, contentW), len(rendered)
}

// composeBlock centers rendered block lines within contentW and appends the
// wrapped attribution lines, each centered beneath the image at the standard
// caption width.
func (v *ImageView) composeBlock(rendered []string, attr string, contentW int) []string {
	imgW := image.BlockWidth(rendered)
	photoPad := max(0, (contentW-imgW)/2)
	out := make([]string, 0, len(rendered)+2)
	for _, l := range rendered {
		out = append(out, strings.Repeat(" ", photoPad)+l)
	}
	if attr != "" {
		style := lipgloss.NewStyle().Foreground(v.palette.Dim)
		captionW := compose.CaptionWidth(v.width, contentW)
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

// FireImageLoad returns a tea.Cmd that loads the images the open article
// renders at or before the load frontier (the lead and the inline images in or
// near the visible viewport, in document order) through the cache hierarchy
// when image rendering is enabled, with one in-flight guard per URL. Images
// beyond the frontier are not fetched, decoded, or rendered until the reader
// scrolls toward them.
func (v *ImageView) FireImageLoad(a store.Article, view ArticleView) tea.Cmd {
	if !v.Enabled() {
		return nil
	}
	v.advanceFrontier(view)
	return v.fireImagesTo(a, view, v.frontier, true)
}

// loadImageCmd starts the image load for the article image at url and position,
// marking the URL in flight so duplicate loads are not started (for example on
// resize), and gating the fetch/decode/render work with the shared concurrency
// bound.
func (v *ImageView) loadImageCmd(a store.Article, url string, position int) tea.Cmd {
	width, vpH, _ := render.ContentGeom(v.width, v.height, v.sess.Config().Display.PaddingX, v.sess.Config().Display.PaddingY)
	v.loading[url] = true
	if v.NativeEnabled() {
		return tea.Batch(
			v.gateCmd(image.BlockCmd(v.sess.ImgCache, v.sess.ImgBlocks, v.sess.Store(), a.ID, url, width, vpH, false, position)),
			v.gateCmd(image.PhotoCmd(v.sess.ImgCache, v.sess.ImgBlocks, v.sess.ImgPhotos, v.sess.Store(), a.ID, url, width, vpH, position)),
		)
	}
	return v.gateCmd(image.BlockCmd(v.sess.ImgCache, v.sess.ImgBlocks, v.sess.Store(), a.ID, url, width, vpH, true, position))
}

// EnsureImageSource fires the image load on resize for every image the open
// article renders at or before the load frontier that has no cached source at
// the current content geometry (for example a stored block rendered at a
// different width, or a native render keyed to an earlier viewport size).
func (v *ImageView) EnsureImageSource(a store.Article, view ArticleView) tea.Cmd {
	if !v.Enabled() {
		return nil
	}
	v.advanceFrontier(view)
	return v.fireImagesTo(a, view, v.frontier, false)
}

// advanceFrontier recomputes the load frontier from the open article's anchor
// rows and current viewport: the frontier is the largest image index whose
// anchor row is within the viewport plus one viewport height of lookahead
// below the fold. It only grows within an article open, so a recompose that
// shifts rows (or a scroll that returns toward the top) can never retract a
// load an earlier position already brought in range.
func (v *ImageView) advanceFrontier(view ArticleView) {
	if view.Article() == nil {
		return
	}
	vp := view.ViewportY()
	limit := vp + 2*view.ViewportHeight()
	k := -1
	for i, row := range view.AnchorRows() {
		if row >= limit {
			break
		}
		k = i
	}
	if k > v.frontier {
		v.frontier = k
	}
}

// AdvanceFrontierLoads advances the load frontier to the current viewport and
// returns a tea.Cmd firing the newly in-range image loads and the deferred
// native renders now in view, or nil when nothing is newly in range. It is the
// scroll hook: every viewport move runs it, and the recompose after a load
// landing runs it too so a deferred native render fires the moment its photo's
// block is within reach.
func (v *ImageView) AdvanceFrontierLoads(view ArticleView) tea.Cmd {
	if !v.Enabled() {
		return nil
	}
	v.advanceFrontier(view)
	a := view.Article()
	if a == nil {
		return nil
	}
	return v.fireImagesTo(*a, view, v.frontier, false)
}

// fireImagesTo returns a tea.Cmd firing the image work for the article's image
// URLs at indices 0..k: deferred native renders now in view, and image loads
// through the cache hierarchy. open selects the open-path policy — fire a load
// for every in-range URL so the native render path runs even when a source is
// already available — versus the scroll/resize policy that skips URLs already
// served at the current geometry (their blocks are already composed). imgLoading
// remains the per-URL in-flight guard, so a frontier advance never double-fires
// a URL whose load or render is still running.
func (v *ImageView) fireImagesTo(a store.Article, view ArticleView, k int, open bool) tea.Cmd {
	width, vpH, _ := render.ContentGeom(v.width, v.height, v.sess.Config().Display.PaddingX, v.sess.Config().Display.PaddingY)
	var cmds []tea.Cmd
	for pos, url := range view.ImageURLs() {
		if pos > k || url == "" {
			continue
		}
		if v.nativePending[url] {
			if v.photoBlockInView(url, view) && !v.loading[url] {
				if nc := v.nativeRenderCmd(url, view); nc != nil {
					cmds = append(cmds, nc)
					delete(v.nativePending, url)
				}
			}
			continue
		}
		if v.loading[url] {
			continue
		}
		if !open && v.hasImageSource(url, width, vpH) {
			continue
		}
		cmds = append(cmds, v.loadImageCmd(a, url, pos))
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// photoBlockInView reports whether the open article's photo at url has its
// block within the viewport plus the lookahead margin (the same condition the
// load frontier uses), so its native render is worth starting.
func (v *ImageView) photoBlockInView(url string, view ArticleView) bool {
	idx := view.URLIndex(url)
	if idx < 0 || idx >= len(view.AnchorRows()) {
		return false
	}
	vp := view.ViewportY()
	return view.AnchorRows()[idx] < vp+2*view.ViewportHeight()
}

// gateCmd wraps an image command so it acquires the shared in-flight slot
// before running its fetch/decode/render work and releases it after, bounding
// the number of concurrent operations across the block, photo, and native
// command types. Firing more commands than the bound just queues them; the
// semaphore, not the batch structure, enforces the limit.
func (v *ImageView) gateCmd(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		v.sem <- struct{}{}
		defer func() { <-v.sem }()
		return cmd()
	}
}

// hasImageSource reports whether a cached source renders the image at the given
// content geometry: a native render at that size, a fetched photo (native), a
// decoded image, or a width-matched stored block.
func (v *ImageView) hasImageSource(url string, contentW, vpH int) bool {
	if _, ok := v.sess.ImgNatives.Get(image.NativeKey(url, contentW, vpH)); ok {
		return true
	}
	if _, ok := v.sess.ImgPhotos.Get(url); ok {
		return true
	}
	if _, ok := v.sess.ImgCache.Get(url); ok {
		return true
	}
	if lines, ok := v.sess.ImgBlocks.Get(url); ok && image.BlockWidth(lines) == contentW {
		return true
	}
	return false
}

// OnBlockLoaded marks the article for recomposition when the message's URL is
// one of its images; the recompose runs via the caller's flush after the
// Update switch. The block cache is always updated so a later open renders the
// block. It returns whether the article view is current and renders the
// message's image (the recompose signal). active reports whether the article
// view is current.
func (v *ImageView) OnBlockLoaded(msg image.BlockMsg, view ArticleView, active bool) bool {
	v.loading[msg.Key] = false
	v.sess.ImgBlocks.Set(msg.Key, msg.Lines)
	if !active {
		return false
	}
	return view.RendersImage(msg.Key)
}

// inlineAttrFor returns the composed caption for an inline image by URL: the
// caption recorded when the block was composed (alt, else figure credit, else
// link text, else the source fallback), so the native render's height fit
// reserves exactly the lines the composed block carries. It falls back to the
// per-image derivation when the article state has no recorded caption. It
// returns "" when url is not one of the open article's inline images.
func (v *ImageView) inlineAttrFor(url string, view ArticleView) string {
	if cap := view.InlineCaps()[url]; cap != "" {
		return cap
	}
	for _, im := range view.InlineImages() {
		if im.URL == url {
			return compose.ResolveImageAttribution(*view.Article(), url, im.Alt, "", false)
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
func (v *ImageView) nativeRenderCmd(url string, view ArticleView) tea.Cmd {
	if view.Article() == nil || !v.NativeEnabled() || url == "" {
		return nil
	}
	v.loading[url] = true
	a := *view.Article()
	attr := ""
	if cap := v.inlineAttrFor(url, view); cap != "" {
		attr = cap
	} else {
		attr = compose.ResolveImageAttribution(a, a.ImageURL, "", "", true)
	}
	width, vpH, _ := render.ContentGeom(v.width, v.height, v.sess.Config().Display.PaddingX, v.sess.Config().Display.PaddingY)
	return v.gateCmd(image.NativeCmd(v.sess.ImgCache, v.sess.ImgNative, v.sess.ImgPhotos, v.sess.ImgNatives, url, width, vpH, attr, view.FitHeaderLines(), compose.CaptionWidth(v.width, width)))
}

// OnPhotoLoaded fires the off-thread native render for the photo at msg's URL
// when it belongs to the open article and its block is in or near view,
// replacing the on-thread render the UI used to perform. A photo whose block is
// out of view keeps its halfblock placeholder: the URL is recorded in
// nativePending and the render starts when the reader scrolls it into view. It
// returns the recompose signal (whether the article view is current and renders
// the message's image) and the tea.Cmd firing the deferred native render, or
// nil.
func (v *ImageView) OnPhotoLoaded(msg image.PhotoMsg, view ArticleView, active bool) (bool, tea.Cmd) {
	v.loading[msg.Key] = false
	if !active {
		return false, nil
	}
	if !view.RendersImage(msg.Key) {
		return false, nil
	}
	if v.photoBlockInView(msg.Key, view) {
		return true, v.nativeRenderCmd(msg.Key, view)
	}
	v.nativePending[msg.Key] = true
	return true, nil
}

// OnNativeLoaded marks the article for recomposition with its freshly rendered
// native block for msg's URL, replacing the placeholder halfblock; the recompose
// runs via the caller's flush after the Update switch. It always clears the
// in-flight guard, even when the article changed or the view departed meanwhile
// (the render is still cached for later). It returns the recompose signal.
func (v *ImageView) OnNativeLoaded(msg image.NativeMsg, view ArticleView, active bool) bool {
	v.loading[msg.Key] = false
	if !active {
		return false
	}
	return view.RendersImage(msg.Key)
}

// OnImageFailed clears the in-flight guard and reports whether the article view
// is current and renders the message's image (so a block that became available
// meanwhile triggers a recompose and appears); otherwise the article already
// renders without an image block and the failure is never surfaced.
func (v *ImageView) OnImageFailed(msg image.FailedMsg, view ArticleView, active bool) bool {
	v.loading[msg.Key] = false
	if !active {
		return false
	}
	if !view.RendersImage(msg.Key) {
		return false
	}
	if _, ok := v.sess.ImgBlocks.Get(msg.Key); ok {
		return true
	}
	return false
}

// ScrollArticle runs a viewport scroll operation then, for single-line moves
// (snap true), applies the two-stage image snap: down brings an inline image in
// at the viewport bottom then moves it to the top, up mirrors it, and a move
// that lands inside a photo range skips it onto its caption (or reveals it).
// Multi-line moves (page, half-page, goto-bottom) pass snap false and land
// wherever they land, even mid-photo, so paging never skips the text between
// images. Every move then advances the load frontier and returns the command
// firing the newly in-range image loads and deferred native renders, or nil.
func (v *ImageView) ScrollArticle(view ArticleView, fn func(), snap bool) tea.Cmd {
	before := view.ViewportY()
	fn()
	after := view.ViewportY()
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
		snapped := compose.SnapYOffset(after, view.ViewportHeight(), before, view.ImageBlocks(), dir, view.NativeImg(), view.LeadShown())
		if snapped != after {
			view.SetViewportY(snapped)
		}
	}
	return v.AdvanceFrontierLoads(view)
}