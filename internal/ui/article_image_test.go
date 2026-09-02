package ui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	imgpkg "yerss/internal/image"
	"yerss/internal/store"
	"yerss/internal/ui/compose"
	"yerss/internal/ui/render"
)

// dimsCall records a render request the fake renderer received.
type dimsCall struct{ w, maxH int }

// frameView simulates one full render frame: recompute the native-image clear
// prefix (as Update does after a state change) then compose the view. Tests
// that drive state by hand and then inspect View use it so View itself stays a
// pure function of model state.
func frameView(m *Model) string {
	m.recomputeNativeClear()
	return m.View()
}

// fakeRenderer returns fixed lines and records the width/maxHeight it was
// asked to render at, so tests can assert composition without real images.
type fakeRenderer struct {
	fixed []string
	calls []dimsCall
}

func (f *fakeRenderer) Render(_ image.Image, width, maxHeight int) ([]string, error) {
	f.calls = append(f.calls, dimsCall{w: width, maxH: maxHeight})
	return f.fixed, nil
}

func testImg() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 100, B: 100, A: 255})
		}
	}
	return img
}

func imageArticle() store.Article {
	return store.Article{
		FeedURL:  "https://example.com/feed.xml",
		GUID:     "g-img",
		Title:    "Headline",
		Author:   "Jane",
		Link:     "https://example.com/a",
		ImageURL: "https://example.com/lead.jpg",
		Content:  strings.Repeat("<p>long paragraph body text</p>", 60),
	}
}

func newImageModel(t *testing.T, fixed []string) (*Model, *fakeRenderer) {
	t.Helper()
	m, _ := newTestModel(t)
	fr := &fakeRenderer{fixed: fixed}
	m.sess.ImgRenderer = fr
	m.sess.ImgCache.Set(imageArticle().ImageURL, testImg())
	return m, fr
}

func TestArticleImageBlockComposedBelowHeader(t *testing.T) {
	m, fr := newImageModel(t, []string{"IMG1", "IMG2"})
	m.article = m.newArticleState(imageArticle())

	if len(fr.calls) == 0 {
		t.Fatal("renderer was not called")
	}
	// Header (URL, blank, title, author) is 4 lines; the block sits below it
	// after one blank line and covers the two image lines plus the attribution
	// (which wraps at the standard caption width — the 51-cell caption width for
	// the 74-cell content area holds the whole 14-cell "photo: example" on one
	// line).
	if m.article.imgStart != 5 || m.article.imgEnd != 7 {
		t.Errorf("img range = %d..%d, want 5..7", m.article.imgStart, m.article.imgEnd)
	}
	lines := strippedLines(m.article.lines)
	if !strings.HasPrefix(lines[0], "https://example.com/a") {
		t.Errorf("first content line = %q, want the URL", lines[0])
	}
	if !strings.HasSuffix(lines[5], "IMG1") || !strings.HasSuffix(lines[6], "IMG2") {
		t.Errorf("image lines not below the header: %q", lines[5:7])
	}
	if !strings.Contains(strings.TrimRight(lines[7], " "), "photo: example") {
		t.Errorf("attribution line = %q, want it to contain %q", lines[7], "photo: example")
	}
	if !strings.Contains(strings.Join(lines, "\n"), "Headline") || !strings.Contains(strings.Join(lines, "\n"), "body text") {
		t.Errorf("header/body missing around the image block")
	}
}

func TestArticleImageBlockDisabled(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(m *Model)
	}{
		{"ascii", func(m *Model) { m.ascii = true }},
		{"config off", func(m *Model) { m.sess.Config().Display.Images = "off" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, fr := newImageModel(t, []string{"IMG1"})
			tc.mutate(m)
			m.article = m.newArticleState(imageArticle())
			if len(fr.calls) != 0 {
				t.Errorf("renderer should not be called when images disabled")
			}
			if m.article.imgStart != -1 || m.article.imgEnd != -1 {
				t.Errorf("img range = %d..%d, want -1..-1", m.article.imgStart, m.article.imgEnd)
			}
		})
	}
}

func TestArticleImageBlockNotLoadedYet(t *testing.T) {
	m, fr := newImageModel(t, []string{"IMG1"})
	m.sess.ImgCache = imgpkg.NewCache() // drop the preset image
	m.article = m.newArticleState(imageArticle())
	if len(fr.calls) != 0 {
		t.Errorf("renderer should not be called before the image loads")
	}
	if m.article.imgStart != -1 {
		t.Errorf("imgStart = %d, want -1", m.article.imgStart)
	}
}

func TestArticleImageBlockHeightCap(t *testing.T) {
	// A wide fake photo keeps the fallback attribution on one line, so one
	// render call at the reserved cap is enough.
	wide := strings.Repeat("I", 40)
	m, fr := newImageModel(t, []string{wide, wide})
	m.height = 24
	m.article = m.newArticleState(imageArticle())
	if len(fr.calls) != 1 {
		t.Fatalf("expected one render call, got %d", len(fr.calls))
	}
	header := m.renderHeader(imageArticle(), m.article.viewport.Width)
	headerLines := 0
	if header != "" {
		headerLines = len(strings.Split(header, "\n"))
	}
	// The cap reserves the blank lines around the block, one attribution
	// line, and one line of body text.
	if want := m.article.viewport.Height - headerLines - 4; fr.calls[0].maxH != want {
		t.Errorf("height cap = %d, want %d", fr.calls[0].maxH, want)
	}
	if fr.calls[0].w != m.article.viewport.Width {
		t.Errorf("width = %d, want viewport width %d", fr.calls[0].w, m.article.viewport.Width)
	}
}

func TestImageLoadRecomposesAndPreservesScroll(t *testing.T) {
	m, fr := newImageModel(t, []string{"IMG1", "IMG2"})
	m.sess.ImgCache = imgpkg.NewCache() // no image yet
	m.article = m.newArticleState(imageArticle())
	// Scroll into the body past the insertion point.
	m.article.viewport.ScrollDown(1)
	oldOffset := m.article.viewport.YOffset
	if oldOffset <= 0 {
		t.Fatal("expected to have scrolled past the top")
	}

	m.sess.ImgCache.Set(imageArticle().ImageURL, testImg())
	m.view = viewArticle
	m.onBlockLoaded(imgpkg.BlockMsg{Key: imageArticle().ImageURL, Lines: []string{"IMG1", "IMG2"}})

	if m.article.imgStart <= 0 {
		t.Fatalf("image block not composed after load: imgStart=%d", m.article.imgStart)
	}
	// The bumped offset (past the insertion point) lands inside the lead
	// photo range, so the recompose snap moves it onto the caption.
	if want := m.article.capStart; m.article.viewport.YOffset != want {
		t.Errorf("offset = %d, want %d (bumped then snapped onto the caption)", m.article.viewport.YOffset, want)
	}
	_ = fr
}

func TestImageLoadAtTopInsertsBlockAbove(t *testing.T) {
	m, _ := newImageModel(t, []string{"IMG1"})
	m.sess.ImgCache = imgpkg.NewCache()
	m.article = m.newArticleState(imageArticle())
	if m.article.viewport.YOffset != 0 {
		t.Fatal("expected offset 0 at top")
	}
	m.sess.ImgCache.Set(imageArticle().ImageURL, testImg())
	m.view = viewArticle
	m.onBlockLoaded(imgpkg.BlockMsg{Key: imageArticle().ImageURL, Lines: []string{"IMG1"}})
	if m.article.viewport.YOffset != 0 {
		t.Errorf("offset = %d, want 0 when at the top", m.article.viewport.YOffset)
	}
	if m.article.imgStart <= 0 {
		t.Errorf("image block missing after load: imgStart=%d", m.article.imgStart)
	}
}

func TestImageFailedRendersWithoutBlock(t *testing.T) {
	m, fr := newImageModel(t, []string{"IMG1"})
	m.sess.ImgCache = imgpkg.NewCache()
	m.article = m.newArticleState(imageArticle())
	m.onImageFailed(imgpkg.FailedMsg{Key: imageArticle().ImageURL})
	if m.article.imgStart != -1 {
		t.Errorf("failed image should leave no block, imgStart=%d", m.article.imgStart)
	}
	if len(fr.calls) != 0 {
		t.Errorf("renderer should not be called on failure")
	}
}

func TestOpenArticleFiresImageLoadCmd(t *testing.T) {
	m, _ := newTestModel(t)
	m.sess.Config().Display.Images = "on"
	a := imageArticle()
	a.GUID = "g-open"
	id, err := m.sess.Store().UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	full, err := m.sess.Store().GetArticle(id)
	if err != nil {
		t.Fatal(err)
	}
	m.sess.ImgCache.Set(full.ImageURL, testImg())
	m.loadList()
	m.list.cursor = 1
	cmd := m.openArticle()
	if cmd == nil {
		t.Fatal("openArticle should return an image load command when a URL is present")
	}
	// The command resolves from the in-memory cache without a network request.
	msg := cmd()
	if _, ok := msg.(imgpkg.BlockMsg); !ok {
		t.Fatalf("expected BlockMsg, got %T", msg)
	}
}

func TestOpenArticleNoImageCmdWithoutURL(t *testing.T) {
	m, _ := newTestModel(t)
	insertArticle(t, m.sess.Store(), "one", nil)
	m.loadList()
	if cmd := m.openArticle(); cmd != nil {
		t.Errorf("openArticle without an image URL should return nil, got %T", cmd)
	}
}

func TestRestoreArticleFiresImageLoadCmd(t *testing.T) {
	withLocalZone(t, time.UTC)
	st := openSharedStore(t)
	m1 := modelOn(t, st)
	m1.sess.Config().Display.Images = "on"
	if _, err := m1.sess.Store().UpsertArticle(imageArticle()); err != nil {
		t.Fatal(err)
	}
	m1.loadList()
	m1.list.cursor = 1
	m1.openArticle()
	if m1.view != viewArticle {
		t.Fatalf("expected article view, got %d", m1.view)
	}
	m1.article.viewport.ScrollDown(3)
	m1.persistSelection()

	// A fresh model on the same store simulates a restart: the in-memory
	// image cache is empty, so the restore must fire a load command through
	// the cache hierarchy instead of leaving the article imageless.
	m2 := modelOn(t, st)
	m2.sess.Config().Display.Images = "on"
	m2.loadList()
	cmd := m2.restoreSelection()
	if cmd == nil {
		t.Fatal("restoreSelection should return an image load command for the restored article")
	}
	if m2.view != viewArticle {
		t.Fatalf("restored view = %d, want article", m2.view)
	}
	if m2.article.imgStart != -1 {
		t.Fatalf("image block should be absent until the load lands, imgStart=%d", m2.article.imgStart)
	}

	// The command resolves offline from the preseeded cache, and the load
	// message composes the image block above the preserved reading position.
	m2.sess.ImgCache.Set(imageArticle().ImageURL, testImg())
	msg := cmd()
	lm, ok := msg.(imgpkg.BlockMsg)
	if !ok {
		t.Fatalf("expected BlockMsg, got %T", msg)
	}
	if lm.Key != imageArticle().ImageURL {
		t.Errorf("BlockMsg.Key = %q, want %q", lm.Key, imageArticle().ImageURL)
	}
	wantOffset := m2.article.viewport.YOffset
	m2.onBlockLoaded(lm)
	if m2.article.imgStart <= 0 {
		t.Errorf("image block missing after load: imgStart=%d", m2.article.imgStart)
	}
	if m2.article.viewport.YOffset <= wantOffset {
		t.Errorf("offset = %d, want it bumped past the inserted image lines (was %d)", m2.article.viewport.YOffset, wantOffset)
	}
}

func TestSnapYOffset(t *testing.T) {
	cases := []struct {
		name                            string
		offset, vpH, start, end, cap, d int
		native                          bool
		want                            int
	}{
		// offset is the position AFTER a scroll move; the photo occupies
		// lines [start, cap) and the caption [cap, end]. The photo must
		// never be partially clipped, while the caption scrolls normally.
		{"down onto first photo line", 4, 20, 4, 9, 8, 1, false, 8},
		{"down into photo interior", 6, 20, 4, 9, 8, 1, false, 8},
		{"down at last photo line", 7, 20, 4, 9, 8, 1, false, 8},
		{"down onto caption unchanged", 8, 20, 4, 9, 8, 1, false, 8},
		{"down within caption unchanged", 9, 20, 4, 9, 8, 1, false, 9},
		{"up from caption last line", 9, 20, 4, 9, 8, -1, false, 9},
		{"up from caption first line", 8, 20, 4, 9, 8, -1, false, 8},
		{"up from photo interior", 6, 20, 4, 9, 8, -1, false, 4},
		{"up onto first line", 4, 20, 4, 9, 8, -1, false, 4},
		{"below block unchanged", 10, 20, 4, 9, 8, 1, false, 10},
		{"above block unchanged", 2, 20, 4, 9, 8, -1, false, 2},
		// Down from a fully visible photo (window top at or above the image
		// and the image end within the viewport bottom) skips to the caption.
		{"down from header while full photo on screen skips to caption", 2, 20, 4, 9, 8, 1, false, 8},
		{"down from block top while full photo on screen skips to caption", 4, 20, 4, 9, 8, 1, false, 8},
		// Photo far below the fold: down does not skip.
		{"down photo below fold unchanged", 2, 20, 30, 37, 34, 1, false, 2},
		// Page-down far below the block: no over-skip.
		{"page down below block unchanged", 2, 20, 30, 37, 34, 20, false, 2},
		{"no image block", 5, 20, -1, -1, -1, 1, false, 5},
		{"page down into photo interior", 6, 20, 4, 9, 8, 20, false, 8},
		{"page down onto caption", 8, 20, 4, 9, 8, 20, false, 8},
		// Without a caption cap sits past the block and the down-snap
		// keeps skipping the whole block.
		{"down onto first line no caption", 4, 20, 4, 9, 10, 1, false, 10},
		{"down at last line no caption", 9, 20, 4, 9, 10, 1, false, 10},
		{"up from last line no caption", 9, 20, 4, 9, 10, -1, false, 4},
		// Native: an up move while the photo is fully visible (offset not at
		// the article top) skips straight to the top, because re-transmitting
		// a native photo per header step is costly.
		{"native up over fully visible photo skips to top", 2, 20, 4, 9, 8, -1, true, 0},
		// Halfblock: the up move stays on the current line; the header
		// scrolls up line by line since halfblock re-render is cheap.
		{"halfblock up over fully visible photo unchanged", 2, 20, 4, 9, 8, -1, false, 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var blocks []compose.ImageBlock
			if c.start >= 0 {
				blocks = append(blocks, compose.ImageBlock{ImgStart: c.start, CapStart: c.cap, ImgEnd: c.end, NativeImg: c.native})
			}
			if got := compose.SnapYOffset(c.offset, c.vpH, c.offset, blocks, c.d, c.native); got != c.want {
				t.Errorf("compose.SnapYOffset(%d,%d,%d,%+v,%d,%v) = %d, want %d", c.offset, c.vpH, c.offset, blocks, c.d, c.native, got, c.want)
			}
		})
	}
}

func TestScrollSnapsOntoCaption(t *testing.T) {
	m, _ := newImageModel(t, []string{"IMG1", "IMG2"})
	m.article = m.newArticleState(imageArticle())
	// The photo rows sit below the header with the wrapped attribution
	// beneath; a downward move from just above them enters the photo's rows
	// and snaps onto the caption instead of past it.
	m.article.viewport.SetYOffset(m.article.imgStart - 1)
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
	if want := m.article.imgStart + 2; m.article.capStart != want {
		t.Fatalf("caption start = %d, want %d (two photo rows above it)", m.article.capStart, want)
	}
	if got := m.article.viewport.YOffset; got != m.article.capStart {
		t.Errorf("offset after down scroll = %d, want %d (the caption)", got, m.article.capStart)
	}

	// Down again scrolls within the caption line by line (no snap loop).
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
	if got := m.article.viewport.YOffset; got != m.article.capStart+1 {
		t.Errorf("offset scrolling within the caption = %d, want %d", got, m.article.capStart+1)
	}

	// Up from a caption line scrolls the caption upward line by line rather
	// than snapping to the photo.
	m.scrollArticle(func() { m.article.viewport.ScrollUp(1) }, true)
	if got := m.article.viewport.YOffset; got != m.article.capStart {
		t.Errorf("offset after up scroll = %d, want %d (the caption's top line)", got, m.article.capStart)
	}
	// Up from a photo row reveals the full image.
	m.article.viewport.SetYOffset(m.article.imgStart + 1)
	m.scrollArticle(func() { m.article.viewport.ScrollUp(1) }, true)
	if got := m.article.viewport.YOffset; got != m.article.imgStart {
		t.Errorf("offset after up scroll from a photo row = %d, want %d (reveal)", got, m.article.imgStart)
	}
}

func TestScrollSkipsFullyVisiblePhoto(t *testing.T) {
	m, _ := newImageModel(t, []string{"IMG1", "IMG2"})
	m.article = m.newArticleState(imageArticle())
	// The height-fitted photo occupies a handful of lines at the top of the
	// article, so at the top of the viewport the full block is on screen.
	if m.article.imgStart <= 0 {
		t.Fatalf("expected an image block on screen, got imgStart=%d", m.article.imgStart)
	}
	// A single downward scroll from above the block snaps to the caption, the
	// first non-photo line.
	m.article.viewport.SetYOffset(0)
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
	if got := m.article.viewport.YOffset; got != m.article.capStart {
		t.Errorf("offset down over a fully visible photo = %d, want %d (the first non-photo line)", got, m.article.capStart)
	}

	// The next downward scroll continues into the body text rather than
	// re-skipping the block.
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
	if got, want := m.article.viewport.YOffset, m.article.capStart+1; got != want {
		t.Errorf("offset after the next down scroll = %d, want %d", got, want)
	}

	// Page-down that lands past the block preserves the full-page scroll and
	// does not snap back onto the photo or caption.
	m.article.viewport.SetYOffset(m.article.imgStart)
	m.scrollArticle(func() { m.article.viewport.PageDown() }, false)
	if got := m.article.viewport.YOffset; got < m.article.imgEnd {
		t.Errorf("page-down offset = %d, want it past the block end %d", got, m.article.imgEnd)
	}

	// Up from just below the block lands on the caption line (the block's
	// last line), scrolling the caption up line by line rather than snapping
	// to the photo, per the narrowed upward reveal rule.
	m.article.viewport.SetYOffset(m.article.imgEnd + 1)
	m.scrollArticle(func() { m.article.viewport.ScrollUp(1) }, true)
	if got := m.article.viewport.YOffset; got != m.article.imgEnd {
		t.Errorf("offset up from below the block = %d, want %d (the caption)", got, m.article.imgEnd)
	}

	// Up from a photo row reveals the full image.
	m.article.viewport.SetYOffset(m.article.imgStart + 1)
	m.scrollArticle(func() { m.article.viewport.ScrollUp(1) }, true)
	if got := m.article.viewport.YOffset; got != m.article.imgStart {
		t.Errorf("offset up from a photo row = %d, want %d (reveal)", got, m.article.imgStart)
	}

	// Native: up from the photo top with a fully visible photo skips straight
	// to the article top, since re-transmitting a native photo per header step
	// is costly.
	m.article.nativeImg = true
	m.article.viewport.SetYOffset(m.article.imgStart)
	m.scrollArticle(func() { m.article.viewport.ScrollUp(1) }, true)
	if got := m.article.viewport.YOffset; got != 0 {
		t.Errorf("native up from the photo top = %d, want 0 (the article top)", got)
	}
}

func TestAttributionWrapsAtCaptionWidth(t *testing.T) {
	// A narrow fake photo (4 cells) with a long credit: the credit wraps at
	// the standard caption width (a fixed fraction of the content width), not
	// the photo's 4-cell width, and the wrapped lines stay within that width.
	longCap := strings.Repeat("credit words ", 7)
	a := imageArticle()
	a.Content = `<figure><img src="https://example.com/lead.jpg"><figcaption>` + longCap + `</figcaption></figure>` +
		strings.Repeat("<p>long paragraph body text</p>", 60)
	m, _ := newImageModel(t, []string{"IMG1", "IMG2"})
	m.article = m.newArticleState(a)

	contentW := m.article.viewport.Width
	captionW := compose.CaptionWidth(contentW)
	if captionW <= 4 {
		t.Fatalf("caption width %d should exceed the 4-cell photo", captionW)
	}
	wantLines := len(imgpkg.LayoutWrapLines(longCap, captionW))
	if got := m.article.imgEnd - m.article.imgStart - 1; got != wantLines {
		t.Fatalf("caption lines = %d, want %d (wrapped at caption width %d)", got, wantLines, captionW)
	}
	lines := strippedLines(m.article.lines)
	for i := 0; i < wantLines; i++ {
		attr := lines[m.article.imgStart+2+i]
		if !strings.Contains(attr, "credit") {
			t.Errorf("caption line %d = %q, want credit text", i, attr)
		}
	}
	// The image plus the wrapped caption fits the viewport budget.
	if m.article.imgEnd-m.article.imgStart+1 >= m.article.viewport.Height {
		t.Errorf("block %d..%d does not fit viewport %d", m.article.imgStart, m.article.imgEnd, m.article.viewport.Height)
	}
}

func TestAttributionDirectlyBeneathImage(t *testing.T) {
	// A 40-cell fake photo leaves the 14-cell fallback attribution
	// untruncated within the photo's width.
	wide := strings.Repeat("I", 40)
	m, _ := newImageModel(t, []string{wide, wide})
	m.article = m.newArticleState(imageArticle())

	lines := strippedLines(m.article.lines)
	// The attribution is the block's last line, immediately after the last
	// image line with no blank between them.
	if !strings.HasSuffix(strings.TrimRight(lines[m.article.imgEnd-1], " "), "IIII") {
		t.Errorf("line above the attribution = %q, want the last image line", lines[m.article.imgEnd-1])
	}
	if !strings.Contains(lines[m.article.imgEnd], "photo: example") {
		t.Errorf("attribution line = %q, want the org fallback", lines[m.article.imgEnd])
	}
	// The attribution is centered within the photo's own span: the photo is
	// padded (contentW-40)/2 and the text (40-14)/2 further.
	contentW := m.article.viewport.Width
	attr := lines[m.article.imgEnd]
	wantPad := (contentW-40)/2 + (40-14)/2
	if got := len(attr) - len(strings.TrimLeft(attr, " ")); got != wantPad {
		t.Errorf("attribution left pad = %d, want %d: %q", got, wantPad, attr)
	}
	// Exactly one blank line separates the attribution from the body.
	if strings.TrimSpace(lines[m.article.imgEnd+1]) != "" {
		t.Errorf("line after the attribution = %q, want a blank line", lines[m.article.imgEnd+1])
	}
	if !strings.Contains(lines[m.article.imgEnd+2], "body text") {
		t.Errorf("body should resume after the blank line: %q", lines[m.article.imgEnd+2])
	}
}

func TestAttributionFromArticleHTML(t *testing.T) {
	wide := strings.Repeat("I", 40)
	m, _ := newImageModel(t, []string{wide})
	a := imageArticle()
	a.Content = `<figure><img src="` + a.ImageURL + `"><figcaption>Photo: Jane Doe</figcaption></figure>` +
		strings.Repeat("<p>long paragraph body text</p>", 60)
	m.article = m.newArticleState(a)

	lines := strippedLines(m.article.lines)
	if !strings.Contains(lines[m.article.imgEnd], "Photo: Jane Doe") {
		t.Errorf("attribution = %q, want the HTML credit", lines[m.article.imgEnd])
	}
	if strings.Contains(strings.Join(lines, "\n"), "photo: example") {
		t.Errorf("org fallback should not appear when an HTML credit exists")
	}
}

// pngBytes encodes a solid-color image as PNG bytes for native renders.
func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// kittyDeleteEsc is the delete-all-visible-placements sequence the frame must
// carry whenever it no longer displays the article's native image.
const kittyDeleteEsc = "\x1b_Ga=d,d=a,q=1\x1b\\"

// nativeRenderLines renders the article's photo through the same NativeCmd
// path the UI uses on PhotoMsg, which caches the finished block in imgNatives
// so a subsequent compose serves it. Tests call it before newArticleState to
// stand in for the PhotoMsg → NativeCmd → NativeMsg flow.
func nativeRenderLines(t *testing.T, m *Model, a store.Article) []string {
	t.Helper()
	width, vpH, _ := render.ContentGeom(m.width, m.height, m.sess.Config().Display.PaddingX, m.sess.Config().Display.PaddingY)
	header := m.renderHeader(a, width)
	headerLines := 0
	if header != "" {
		headerLines = len(strings.Split(header, "\n"))
	}
		msg := imgpkg.NativeCmd(m.sess.ImgNative, m.sess.ImgPhotos, m.sess.ImgNatives, a.ImageURL, width, vpH, compose.ResolveImageAttribution(a, a.ImageURL, "", "", true), headerLines, compose.CaptionWidth(width))()
	nm, ok := msg.(imgpkg.NativeMsg)
	if !ok {
		t.Fatalf("expected NativeMsg, got %T", msg)
	}
	return nm.Lines
}

func TestNativeKittyImageClearsWhenNotDisplayed(t *testing.T) {
	m, _ := newImageModel(t, nil)
	m.sess.ImgNative.Protocol = imgpkg.ProtocolKitty
	m.sess.ImgPhotos.Set(imageArticle().ImageURL, pngBytes(t, 40, 40))
	nativeRenderLines(t, m, imageArticle())
	m.article = m.newArticleState(imageArticle())
	m.view = viewArticle
	if !m.article.nativeImg {
		t.Fatal("expected the composed block to be a native render")
	}

	// While the block intersects the visible window the frame re-transmits
	// the image and must not delete the placement.
	if got := frameView(m); !strings.Contains(got, "\x1b_Ga=T") {
		t.Error("visible native image should carry the transmit escape")
	} else if strings.Contains(got, kittyDeleteEsc) {
		t.Error("visible native image must not be deleted")
	}

	// Landing on the caption (photo rows above the window, caption visible)
	// must also delete the placement: the transmit line is not rendered, so
	// the photo would float over the caption otherwise.
	m.article.viewport.SetYOffset(m.article.capStart)
	leadID := m.article.imageBlocks[0].NativeID
	if leadID == 0 {
		t.Fatal("native block should record its kitty render id")
	}
	if got := frameView(m); !strings.Contains(got, imgpkg.DeleteByID(leadID)) {
		t.Error("caption frame must delete the photo's placement by its render id")
	}

	// Scrolled past the whole block the placement must be deleted too, or
	// the photo would float over the body.
	m.article.viewport.SetYOffset(m.article.imgEnd + 1)
	if got := frameView(m); !strings.Contains(got, imgpkg.DeleteByID(leadID)) {
		t.Error("frame must delete the placement when the image is scrolled out")
	}

	// Backing out to the list must delete the placement too, or the photo
	// persists over the list and stacks over the next article's photo.
	m.backToList()
	if got := frameView(m); !strings.Contains(got, imgpkg.DeleteByID(leadID)) {
		t.Error("list frame must delete the article's placement by its render id")
	}

	// A frame without a native render (the placeholder phase of a fresh
	// article) also carries the delete, so it cannot stack over a stale
	// placement left by an earlier view.
	m.sess.ImgPhotos = imgpkg.NewPhotos()
	m.sess.ImgNatives = imgpkg.NewNatives()
	m.article = m.newArticleState(imageArticle())
	m.view = viewArticle
	if m.article.nativeImg {
		t.Fatal("expected the placeholder state without a native render")
	}
	if got := frameView(m); !strings.Contains(got, kittyDeleteEsc) {
		t.Error("article frame without a native render must delete stray placements")
	}
}

func TestNativeKittyImageRestoredAfterClear(t *testing.T) {
	// Scrolling back to the image re-renders its line, re-transmitting the
	// photo after the delete, so the clear never loses the image.
	m, _ := newImageModel(t, nil)
	m.sess.ImgNative.Protocol = imgpkg.ProtocolKitty
	m.sess.ImgPhotos.Set(imageArticle().ImageURL, pngBytes(t, 40, 40))
	nativeRenderLines(t, m, imageArticle())
	m.article = m.newArticleState(imageArticle())
	m.view = viewArticle

	m.article.viewport.SetYOffset(m.article.imgEnd + 1)
	if got := frameView(m); !strings.Contains(got, imgpkg.DeleteByID(m.article.imageBlocks[0].NativeID)) {
		t.Fatal("scrolled-out image should have been deleted by its render id")
	}
	m.scrollArticle(func() { m.article.viewport.GotoTop() }, false)
	if got := frameView(m); !strings.Contains(got, "\x1b_Ga=T") || strings.Contains(got, "a=d") {
		t.Error("scrolled-back image must be re-transmitted and not deleted")
	}
}

func TestNativeITermImageNeedsNoClear(t *testing.T) {
	// OSC 1337 inline images are cell-bound, so frames never carry the kitty
	// delete sequence on an iTerm2 terminal.
	m, _ := newImageModel(t, nil)
	m.sess.ImgNative.Protocol = imgpkg.ProtocolITerm
	m.sess.ImgPhotos.Set(imageArticle().ImageURL, pngBytes(t, 40, 40))
	nativeRenderLines(t, m, imageArticle())
	m.article = m.newArticleState(imageArticle())
	m.view = viewArticle
	if got := frameView(m); strings.Contains(got, kittyDeleteEsc) {
		t.Error("iTerm frame should not carry the kitty delete sequence")
	}
	m.backToList()
	if got := frameView(m); strings.Contains(got, kittyDeleteEsc) {
		t.Error("iTerm list frame should not carry the kitty delete sequence")
	}
}

func TestNativePhotoCenteredInArticle(t *testing.T) {
	// The viewport height cap narrows a 40x40 photo well below the content
	// width, so the composed block must be centered like a halfblock block:
	// the escape line carries a blank left pad with an equal (±1) right
	// margin, and the attribution beneath wraps to the photo's own width.
	m, _ := newImageModel(t, nil)
	m.sess.ImgNative.Protocol = imgpkg.ProtocolKitty
	m.sess.ImgPhotos.Set(imageArticle().ImageURL, pngBytes(t, 40, 40))
	nativeRenderLines(t, m, imageArticle())
	m.article = m.newArticleState(imageArticle())
	if !m.article.nativeImg {
		t.Fatal("expected a native render")
	}
	contentW := m.article.viewport.Width
	line := m.article.lines[m.article.imgStart]
	escIdx := strings.Index(line, "\x1b_G")
	if escIdx <= 0 {
		t.Fatalf("native photo should be preceded by a centering pad: %q", line[:min(50, len(line))])
	}
	if pad := line[:escIdx]; strings.TrimLeft(pad, " ") != "" {
		t.Fatalf("pad before the escape should be blank, got %q", pad)
	}
	lpad := escIdx
	if right := contentW - ansi.StringWidth(line); right != lpad && right != lpad+1 {
		t.Errorf("photo not centered: left pad %d, right margin %d (line width %d, content width %d)",
			lpad, right, ansi.StringWidth(line), contentW)
	}
}

func TestNativeBlockPlaceholderBeforePhoto(t *testing.T) {
	m, fr := newImageModel(t, []string{"IMG1", "IMG2"})
	m.sess.ImgNative.Protocol = imgpkg.ProtocolITerm
	m.sess.ImgCache = imgpkg.NewCache() // only the stored block is available
	contentW, _, _ := render.ContentGeom(m.width, m.height, m.sess.Config().Display.PaddingX, m.sess.Config().Display.PaddingY)
	m.sess.ImgBlocks.Set(imageArticle().ImageURL, []string{strings.Repeat("I", contentW), strings.Repeat("I", contentW)})
	m.article = m.newArticleState(imageArticle())

	if m.article.imgStart <= 0 {
		t.Fatalf("stored placeholder not composed: imgStart=%d", m.article.imgStart)
	}
	if len(fr.calls) != 0 {
		t.Errorf("placeholder should come from the stored block, not the renderer")
	}
	lines := strippedLines(m.article.lines)
	if !strings.HasSuffix(lines[m.article.imgStart], strings.Repeat("I", contentW)) ||
		!strings.HasSuffix(lines[m.article.imgStart+1], strings.Repeat("I", contentW)) {
		t.Errorf("stored placeholder lines missing: %q", lines[m.article.imgStart:m.article.imgStart+2])
	}
	for _, l := range m.article.lines {
		if strings.Contains(l, "\x1b]1337") {
			t.Error("no native escape before the photo is fetched")
		}
	}
}

func TestNativePhotoReplacesPlaceholderBlock(t *testing.T) {
	m, fr := newImageModel(t, []string{"IMG1"})
	m.sess.ImgNative.Protocol = imgpkg.ProtocolITerm
	m.sess.ImgPhotos.Set(imageArticle().ImageURL, pngBytes(t, 40, 40))
	nativeRenderLines(t, m, imageArticle())
	m.article = m.newArticleState(imageArticle())

	if len(fr.calls) != 0 {
		t.Errorf("native render should not use the halfblock renderer")
	}
	if m.article.imgStart <= 0 {
		t.Fatalf("native image block missing: imgStart=%d", m.article.imgStart)
	}
	found := false
	for _, l := range m.article.lines {
		if strings.Contains(l, "\x1b]1337;File=") {
			found = true
			break
		}
	}
	if !found {
		t.Error("article lines missing the OSC 1337 native escape")
	}
}

func TestNativePhotoMsgFlowRecomposes(t *testing.T) {
	m, fr := newImageModel(t, nil)
	m.sess.ImgNative.Protocol = imgpkg.ProtocolITerm
	m.sess.ImgCache = imgpkg.NewCache()
	m.sess.ImgBlocks = imgpkg.NewBlocks()
	m.article = m.newArticleState(imageArticle())
	m.view = viewArticle
	if m.article.imgStart != -1 {
		t.Fatalf("no image should be composed before the photo render, imgStart=%d", m.article.imgStart)
	}

	// PhotoMsg lands: the photo bytes are cached and onPhotoLoaded fires the
	// off-thread native render command (no on-thread recompose).
	m.sess.ImgPhotos.Set(imageArticle().ImageURL, pngBytes(t, 40, 40))
	cmd := m.onPhotoLoaded(imgpkg.PhotoMsg{Key: imageArticle().ImageURL})
	if cmd == nil {
		t.Fatal("onPhotoLoaded should fire the native render command")
	}
	if m.article.imgStart != -1 {
		t.Fatalf("PhotoMsg must not recompose on its own, imgStart=%d", m.article.imgStart)
	}

	msg := cmd()
	nm, ok := msg.(imgpkg.NativeMsg)
	if !ok {
		t.Fatalf("expected NativeMsg, got %T", msg)
	}
	if nm.Key != imageArticle().ImageURL {
		t.Errorf("NativeMsg.Key = %q, want %q", nm.Key, imageArticle().ImageURL)
	}
	key := imgpkg.NativeKey(imageArticle().ImageURL, m.article.viewport.Width, m.article.viewport.Height)
	if _, ok := m.sess.ImgNatives.Get(key); !ok {
		t.Fatal("NativeCmd should cache the render under the render-size key")
	}

	// The NativeMsg recomposes the article with the native block.
	beforeCalls := len(fr.calls)
	m.onNativeLoaded(nm)
	if m.article.imgStart <= 0 {
		t.Fatalf("native block missing after NativeMsg: imgStart=%d", m.article.imgStart)
	}
	if !m.article.nativeImg {
		t.Error("composed block should be flagged native")
	}
	found := false
	for _, l := range m.article.lines {
		if strings.Contains(l, "\x1b]1337;File=") {
			found = true
			break
		}
	}
	if !found {
		t.Error("article lines missing the OSC 1337 native escape")
	}
	if m.article.viewport.YOffset != 0 {
		t.Errorf("offset = %d, want 0 (was at top)", m.article.viewport.YOffset)
	}

	// A second recompose (e.g. another load message arriving) serves the
	// cached native render without touching the halfblock renderer.
	m.recomposeArticle()
	if !m.article.nativeImg {
		t.Error("cached-native recompose should stay native")
	}
	if len(fr.calls) != beforeCalls {
		t.Errorf("cached recompose must not re-render the halfblock: calls %d -> %d", beforeCalls, len(fr.calls))
	}
}

func TestStoredBlockServedAtMatchingWidth(t *testing.T) {
	m, fr := newImageModel(t, []string{"IMG1", "IMG2"})
	m.sess.ImgCache = imgpkg.NewCache()
	width, _, _ := render.ContentGeom(m.width, m.height, m.sess.Config().Display.PaddingX, m.sess.Config().Display.PaddingY)
	lines := []string{strings.Repeat("I", width), strings.Repeat("I", width)}
	m.sess.ImgBlocks.Set(imageArticle().ImageURL, lines)
	m.article = m.newArticleState(imageArticle())

	if len(fr.calls) != 0 {
		t.Errorf("stored block should not be re-rendered")
	}
	if m.article.imgStart <= 0 {
		t.Fatalf("stored block not composed: imgStart=%d", m.article.imgStart)
	}
	got := strippedLines(m.article.lines)
	if !strings.HasSuffix(got[m.article.imgStart], strings.Repeat("I", width)) {
		t.Errorf("stored block line not served: %q", got[m.article.imgStart])
	}
}

var _ = tea.Cmd(nil)

func TestInlineImageComposedInPlace(t *testing.T) {
	wide := strings.Repeat("I", 40)
	a := imageArticle()
	a.Content = `<p>intro text</p><p><img src="inline.jpg" alt="Alt caption"></p><p>outro text</p>`
	m, fr := newImageModel(t, []string{wide, wide})
	m.sess.ImgCache.Set("inline.jpg", testImg())
	m.article = m.newArticleState(a)

	if len(m.article.imageBlocks) != 2 {
		t.Fatalf("imageBlocks = %d, want 2 (lead + inline)", len(m.article.imageBlocks))
	}
	lead, inline := m.article.imageBlocks[0], m.article.imageBlocks[1]
	if inline.URL != "inline.jpg" {
		t.Errorf("inline url = %q, want inline.jpg", inline.URL)
	}
	if inline.ImgStart <= lead.ImgEnd {
		t.Errorf("inline imgStart %d should sit past lead block end %d", inline.ImgStart, lead.ImgEnd)
	}
	lines := strippedLines(m.article.lines)
	joined := strings.Join(strippedLines(m.article.lines), "\n")
	if !strings.Contains(joined, "intro text") || !strings.Contains(joined, "outro text") {
		t.Errorf("body text missing around inline image: %q", joined)
	}
	// Image rows appear between intro and outro; the attribution is the alt.
	if !strings.HasSuffix(strings.TrimRight(lines[inline.ImgStart], " "), wide) {
		t.Errorf("inline image row = %q, want the wide block line", lines[inline.ImgStart])
	}
	if inline.CapStart <= inline.ImgStart || inline.CapStart > inline.ImgEnd {
		t.Fatalf("inline capStart %d out of block range %d..%d", inline.CapStart, inline.ImgStart, inline.ImgEnd)
	}
	if !strings.Contains(lines[inline.CapStart], "Alt caption") {
		t.Errorf("inline attribution = %q, want the alt text", lines[inline.CapStart])
	}
	if strings.Contains(lines[inline.CapStart], "photo:") {
		t.Errorf("inline attribution should use alt, got %q", lines[inline.CapStart])
	}
	_ = fr
}

func TestInlineImageSourceFallbackAttribution(t *testing.T) {
	wide := strings.Repeat("I", 40)
	a := imageArticle()
	a.Content = `<p>intro</p><p><img src="inline.jpg"></p><p>outro</p>`
	m, _ := newImageModel(t, []string{wide, wide})
	m.sess.ImgCache.Set("inline.jpg", testImg())
	m.article = m.newArticleState(a)

	if len(m.article.imageBlocks) != 2 {
		t.Fatalf("imageBlocks = %d, want 2", len(m.article.imageBlocks))
	}
	inline := m.article.imageBlocks[1]
	lines := strippedLines(m.article.lines)
	if !strings.Contains(lines[inline.CapStart], "photo: example") {
		t.Errorf("inline attribution = %q, want the source fallback", lines[inline.CapStart])
	}
}

func TestInlineImageLinkTextAsCaption(t *testing.T) {
	wide := strings.Repeat("I", 40)
	a := imageArticle()
	a.Content = `<p>intro</p>` +
		`<aside class="promote-banner"><a class="promote-banner__link" href="/tomdispatch">` +
		`<span class="promote-banner__image"><img decoding="async" src="logo.jpg" alt=""></span>` +
		`<div class="promote-banner__text"><p class="promote-banner__eyebrow">Read Our Complete Coverage</p>` +
		`<h2 class="promote-banner__title">TomDispatch</h2></div></a></aside>` +
		`<p>outro</p>`
	m, _ := newImageModel(t, []string{wide, wide})
	m.sess.ImgCache.Set("logo.jpg", testImg())
	m.article = m.newArticleState(a)

	joined := strings.Join(strippedLines(m.article.lines), "\n")
	if strings.Contains(joined, "](") {
		t.Errorf("stray link tail around the linked image: %q", joined)
	}
	var logo *compose.ImageBlock
	for i := range m.article.imageBlocks {
		if m.article.imageBlocks[i].URL == "logo.jpg" {
			logo = &m.article.imageBlocks[i]
		}
	}
	if logo == nil {
		t.Fatalf("logo block not composed: %q", joined)
	}
	lines := strippedLines(m.article.lines)
	caption := strings.Join(lines[logo.CapStart:logo.ImgEnd+1], " ")
	if !strings.Contains(caption, "Read Our Complete Coverage") || !strings.Contains(caption, "TomDispatch") {
		t.Errorf("logo caption = %q, want the link text", caption)
	}
	if strings.Contains(caption, "photo:") {
		t.Errorf("logo caption should use the link text, got %q", caption)
	}
}

func TestInlineImageConsumesWrappingLink(t *testing.T) {
	wide := strings.Repeat("I", 40)
	a := imageArticle()
	// <a><img></a> renders as [SENTINEL](href); the split must consume the
	// whole link so no stray [ or ](...) fragments render around the block.
	a.Content = `<p>intro text</p><p><a href="https://example.com/promo"><img src="inline.jpg"></a></p><p>outro text</p>`
	m, _ := newImageModel(t, []string{wide, wide})
	m.sess.ImgCache.Set("inline.jpg", testImg())
	m.article = m.newArticleState(a)

	joined := strings.Join(strippedLines(m.article.lines), "\n")
	if strings.Contains(joined, "[") || strings.Contains(joined, "](") {
		t.Errorf("stray link fragments around the inline image: %q", joined)
	}
	if !strings.Contains(joined, "intro text") || !strings.Contains(joined, "outro text") {
		t.Errorf("body text missing around the image: %q", joined)
	}
	if len(m.article.imageBlocks) != 2 {
		t.Fatalf("imageBlocks = %d, want 2 (lead + inline)", len(m.article.imageBlocks))
	}
}

func TestInlineImageDoesNotBorrowAnotherImagesCaption(t *testing.T) {
	wide := strings.Repeat("I", 40)
	a := imageArticle()
	a.Content = `<figure><img src="book.jpg"><figcaption>Book cover caption</figcaption></figure>` +
		`<p>intro</p><p><img src="inline.jpg"></p><p>outro</p>`
	m, _ := newImageModel(t, []string{wide, wide})
	m.sess.ImgCache.Set("book.jpg", testImg())
	m.sess.ImgCache.Set("inline.jpg", testImg())
	m.article = m.newArticleState(a)

	if len(m.article.imageBlocks) != 3 {
		t.Fatalf("imageBlocks = %d, want 3 (lead + book + inline)", len(m.article.imageBlocks))
	}
	lines := strippedLines(m.article.lines)
	var bookCap, inlineCap string
	for _, b := range m.article.imageBlocks {
		if b.URL == "book.jpg" {
			bookCap = lines[b.CapStart]
		}
		if b.URL == "inline.jpg" {
			inlineCap = lines[b.CapStart]
		}
	}
	if !strings.Contains(bookCap, "Book cover caption") {
		t.Errorf("book attribution = %q, want its own figure caption", bookCap)
	}
	if strings.Contains(inlineCap, "Book cover caption") {
		t.Errorf("inline attribution = %q, must not borrow the book figure's caption", inlineCap)
	}
	if !strings.Contains(inlineCap, "photo: example") {
		t.Errorf("inline attribution = %q, want the source fallback", inlineCap)
	}
}

func TestInlineImagePhotoGridCaptionOwnership(t *testing.T) {
	// A photo-grid: two per-photo inner figures wrapped by an outer figure
	// whose combined caption belongs to the last photo. The first photo
	// renders with no attribution at all (no borrowed caption, no source
	// fallback); the last photo renders with the combined caption.
	wide := strings.Repeat("I", 40)
	a := imageArticle()
	a.Content = `<p>intro</p>` +
		`<figure class="photo-grid">` +
		`<figure><img src="left.jpg"></figure>` +
		`<figure><img src="right.jpg"></figure>` +
		`<figcaption>Left: 2024 scene. Right: 2026 scene.</figcaption>` +
		`</figure>` +
		`<p>outro</p>`
	m, _ := newImageModel(t, []string{wide, wide})
	m.sess.ImgCache.Set("left.jpg", testImg())
	m.sess.ImgCache.Set("right.jpg", testImg())
	m.article = m.newArticleState(a)

	var left, right *compose.ImageBlock
	for i := range m.article.imageBlocks {
		switch m.article.imageBlocks[i].URL {
		case "left.jpg":
			left = &m.article.imageBlocks[i]
		case "right.jpg":
			right = &m.article.imageBlocks[i]
		}
	}
	if left == nil || right == nil {
		t.Fatalf("want left and right grid blocks, got left=%v right=%v", left, right)
	}
	lines := strippedLines(m.article.lines)
	if left.CapStart != left.ImgEnd+1 {
		t.Errorf("left grid photo capStart %d, want %d (no caption lines composed)", left.CapStart, left.ImgEnd+1)
	}
	caption := strings.Join(lines[right.CapStart:right.ImgEnd+1], " ")
	if !strings.Contains(caption, "Left: 2024 scene. Right: 2026 scene.") {
		t.Errorf("right grid photo caption = %q, want the combined figure caption", caption)
	}
	if strings.Contains(caption, "photo:") {
		t.Errorf("right grid photo should use the figure caption, got %q", caption)
	}
}

func TestInlineImagePlaceholderKeepsParagraphBreak(t *testing.T) {
	a := imageArticle()
	a.Content = `<p>intro text</p><p><img src="inline.jpg" alt="Alt"></p><p>outro text</p>`
	m, _ := newImageModel(t, []string{"IMGL1"})
	m.article = m.newArticleState(a)

	// The inline image has no cached source yet, so it is not composed as a
	// block, but the body keeps a blank line in its place so the paragraphs
	// stay separated until the load lands.
	if len(m.article.imageBlocks) != 1 {
		t.Fatalf("imageBlocks = %d, want 1 (lead only before inline load)", len(m.article.imageBlocks))
	}
	var trimmedParts []string
	for _, l := range strippedLines(m.article.lines) {
		trimmedParts = append(trimmedParts, strings.TrimRight(l, " "))
	}
	trimmed := strings.Join(trimmedParts, "\n")
	if !strings.Contains(trimmed, "intro text\n\noutro text") {
		t.Errorf("paragraph break not preserved while loading: %q", trimmed)
	}
}

func TestInlineImageLoadRecomposesInPlace(t *testing.T) {
	wide := strings.Repeat("I", 40)
	a := imageArticle()
	a.Content = `<p>intro text</p><p><img src="inline.jpg" alt="Alt"></p><p>outro text</p>`
	m, _ := newImageModel(t, []string{wide, wide})
	m.article = m.newArticleState(a)

	// inline.jpg has no cached source yet: only the lead block is composed.
	if len(m.article.imageBlocks) != 1 {
		t.Fatalf("imageBlocks before load = %d, want 1", len(m.article.imageBlocks))
	}

	m.sess.ImgCache.Set("inline.jpg", testImg())
	m.view = viewArticle
	m.onBlockLoaded(imgpkg.BlockMsg{Key: "inline.jpg", Lines: []string{wide, wide}})

	if len(m.article.imageBlocks) != 2 {
		t.Fatalf("imageBlocks after load = %d, want 2", len(m.article.imageBlocks))
	}
	inline := m.article.imageBlocks[1]
	lines := strippedLines(m.article.lines)
	if !strings.HasSuffix(strings.TrimRight(lines[inline.ImgStart], " "), wide) {
		t.Errorf("inline image row after load = %q, want the wide block line", lines[inline.ImgStart])
	}
	if !strings.Contains(lines[inline.CapStart], "Alt") {
		t.Errorf("inline attribution after load = %q, want the alt text", lines[inline.CapStart])
	}
	lineOf := func(sub string) int {
		for i, l := range lines {
			if strings.Contains(l, sub) {
				return i
			}
		}
		return -1
	}
	if iIntro, iOutro := lineOf("intro text"), lineOf("outro text"); iIntro < 0 || iIntro > inline.ImgStart || iOutro < inline.ImgEnd {
		t.Errorf("inline block (lines %d..%d) should sit between intro at %d and outro at %d",
			inline.ImgStart, inline.ImgEnd, iIntro, iOutro)
	}
}

func TestSnapYOffsetBlockList(t *testing.T) {
	// A mid-document inline block between a lead block and trailing body text.
	// The inline block is shorter than the viewport, as fitted images are,
	// exercising the two-stage snap.
	lead := compose.ImageBlock{URL: "lead", ImgStart: 4, CapStart: 6, ImgEnd: 9}
	inline := compose.ImageBlock{URL: "inline", ImgStart: 20, CapStart: 22, ImgEnd: 25}
	blocks := []compose.ImageBlock{lead, inline}

	cases := []struct {
		name      string
		offset    int
		before    int
		vpH       int
		direction int
		native    bool
		want      int
	}{
		{"down into lead photo snaps to its caption", 5, 4, 15, 1, false, 6},
		{"down into inline photo snaps to its caption", 21, 20, 15, 1, false, 22},
		{"up into inline photo reveals", 21, 22, 15, -1, false, 20},
		{"up into lead photo reveals", 5, 6, 15, -1, false, 4},
		// Down: the inline image's top entered the window from below this move
		// (it was fully below the fold before); snap its bottom to the viewport
		// bottom so it is fully visible there, caption included.
		{"down brings inline top into view snaps to its bottom", 6, 5, 15, 1, false, 11},
		// Down from the bottom-aligned position moves the image to the top.
		{"down from bottom-aligned snaps to its top", 12, 11, 15, 1, false, 20},
		// Up: the inline image's bottom entered the window from above this
		// move; snap its top to the viewport top.
		{"up brings inline bottom into view snaps to its top", 24, 25, 15, -1, false, 20},
		// Up from the top-aligned position moves the image to the bottom.
		{"up from top-aligned snaps to its bottom", 19, 20, 15, -1, false, 11},
		// A viewport tall enough to hold the whole inline block: a down move
		// entering the photo from above snaps to its top, never straight to
		// the caption.
		{"down entering fully visible inline snaps to its top", 20, 19, 21, 1, false, 20},
		// The fully-visible skip applies only to the lead block: a down move
		// below a fully visible inline block scrolls normally, preserving the
		// two-stage snap.
		{"down below fully visible inline not skipped", 19, 18, 21, 1, false, 19},
		// The native-up-to-top rule applies only to the lead block; an inline
		// block that is fully visible reveals its top, never the article top.
		{"up over fully visible inline reveals not top", 21, 22, 21, -1, false, 20},
		{"up over fully visible inline native reveals not top", 21, 22, 21, -1, true, 20},
		{"down below the inline block unchanged", 27, 26, 15, 1, false, 27},
		{"up above the inline block unchanged", 2, 3, 15, -1, false, 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := compose.SnapYOffset(c.offset, c.vpH, c.before, blocks, c.direction, c.native); got != c.want {
				t.Errorf("compose.SnapYOffset(%d,%d,%d,%+v,%d,%v) = %d, want %d", c.offset, c.vpH, c.before, blocks, c.direction, c.native, got, c.want)
			}
		})
	}
}

func TestSnapYOffsetWrappedCaptionBottomBoundary(t *testing.T) {
	// An inline image with a caption that wraps to three lines: the bottom
	// snap keys on the wrapped caption's last line (imgEnd), not its first
	// (capStart). A down move that leaves the image's top in the window with
	// only the caption's first line above the fold snaps to the bottom
	// boundary so the full wrapped caption is visible, rather than leaving its
	// tail hanging below the fold.
	lead := compose.ImageBlock{URL: "lead", ImgStart: 4, CapStart: 6, ImgEnd: 9}
	inline := compose.ImageBlock{URL: "inline", ImgStart: 20, CapStart: 23, ImgEnd: 25}
	blocks := []compose.ImageBlock{lead, inline}

	cases := []struct {
		name      string
		offset    int
		before    int
		vpH       int
		direction int
		want      int
	}{
		// Caption first line (23) visible, last line (25) below the fold: the
		// bottom snap fires on the wrapped caption's last line. The pre-change
		// condition keyed on capStart and left the offset unchanged here.
		{"down with wrapped caption tail below fold snaps to its bottom", 9, 8, 15, 1, 11},
		// The next down move rises to the top boundary.
		{"down from bottom boundary rises to its top", 12, 11, 15, 1, 20},
		// Up mirror: an up move cutting the wrapped caption's last line below
		// the fold scrolls the block off rather than leaving the caption cut.
		{"up cutting wrapped caption tail below fold scrolls it off", 9, 10, 15, -1, 5},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := compose.SnapYOffset(c.offset, c.vpH, c.before, blocks, c.direction, false); got != c.want {
				t.Errorf("compose.SnapYOffset(%d,%d,%d) = %d, want %d", c.offset, c.vpH, c.before, got, c.want)
			}
		})
	}
}

func TestSnapYOffsetSkipLeavesNextImagePartial(t *testing.T) {
	// Consecutive inline images taller than the spacing between them, as in
	// the real article: after skipping the first onto its caption, the second
	// image's top is already inside the window (its bottom still below the
	// fold), so the next down move must bottom-align it rather than leave it
	// partially visible (a deleted placement renders blank) awaiting manual
	// scrolling. Modeled on the Intercept smoke-test geometry: 20-row images,
	// one caption line, ~19 rows between image captions, 27-row viewport.
	lead := compose.ImageBlock{URL: "lead", ImgStart: 4, CapStart: 6, ImgEnd: 9}
	one := compose.ImageBlock{URL: "one", ImgStart: 20, CapStart: 40, ImgEnd: 41}
	two := compose.ImageBlock{URL: "two", ImgStart: 60, CapStart: 80, ImgEnd: 81}
	blocks := []compose.ImageBlock{lead, one, two}
	vpH := 27

	cases := []struct {
		name   string
		offset int
		before int
		d      int
		want   int
	}{
		// After one is skipped to its caption, two's top is 20 rows inside the
		// window: the next down move bottom-aligns it.
		{"skip of preceding image leaves next partial snaps to its bottom", 41, 40, 1, 55},
		{"down from that bottom-aligned position moves two to the top", 56, 55, 1, 60},
		{"down from two's top skips it onto its caption", 61, 60, 1, 80},
		// Up mirror: once two is bottom-aligned and an up move cuts its bottom
		// below the fold, the next up move snaps it fully below the fold so it
		// scrolls off cleanly rather than rendering a blank strip. With images
		// closer than the viewport, the scroll-off target lands inside one's
		// photo range, so one is revealed instead.
		{"up cutting two's bottom below fold snaps it off", 52, 53, -1, 20},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := compose.SnapYOffset(c.offset, vpH, c.before, blocks, c.d, false); got != c.want {
				t.Errorf("compose.SnapYOffset(%d,%d,%d) = %d, want %d", c.offset, vpH, c.before, got, c.want)
			}
		})
	}
}

func TestSnapYOffsetUpFromBottomBoundaryScrollsOff(t *testing.T) {
	// An up move from an inline image's BOTTOM boundary (its last line on the
	// viewport's last row) snaps it fully out of view in one press — its top
	// lands at the fold — rather than pausing on the caption or revealing the
	// previous image with this one still partially visible.
	lead := compose.ImageBlock{URL: "lead", ImgStart: 5, CapStart: 19, ImgEnd: 19}
	inline := compose.ImageBlock{URL: "inline", ImgStart: 21, CapStart: 35, ImgEnd: 35}
	blocks := []compose.ImageBlock{lead, inline}

	cases := []struct {
		name      string
		offset    int
		before    int
		vpH       int
		direction int
		want      int
	}{
		// Up from the inline's bottom boundary snaps it off (its top at the
		// fold), one press, never resting on the caption or the gap.
		{"up from the bottom boundary snaps it off", 14, 15, 21, -1, 0},
		// From the inline's TOP boundary the sink stage still moves to its
		// bottom; the off rule fires only on the bottom boundary.
		{"up from the top boundary sinks to its bottom", 20, 21, 21, -1, 15},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := compose.SnapYOffset(c.offset, c.vpH, c.before, blocks, c.direction, false); got != c.want {
				t.Errorf("compose.SnapYOffset(%d,%d,%d) = %d, want %d", c.offset, c.vpH, c.before, got, c.want)
			}
		})
	}
}

func TestSnapYOffsetUpFromBottomBoundaryRevealsPreviousWhenClose(t *testing.T) {
	// With images spaced closer than the viewport height, an up move from a
	// lower image's bottom boundary scrolls it off onto the previous image's
	// top, so the previous image is never left partially visible (its photo
	// range would otherwise catch the scroll-off target).
	lead := compose.ImageBlock{URL: "lead", ImgStart: 4, CapStart: 6, ImgEnd: 9}
	one := compose.ImageBlock{URL: "one", ImgStart: 20, CapStart: 40, ImgEnd: 41}
	two := compose.ImageBlock{URL: "two", ImgStart: 60, CapStart: 80, ImgEnd: 81}
	blocks := []compose.ImageBlock{lead, one, two}
	vpH := 27

	// two's bottom boundary is 55; its off target (33) lands inside one's
	// photo range, so one is revealed at its top instead.
	if got := compose.SnapYOffset(54, vpH, 55, blocks, -1, false); got != 20 {
		t.Errorf("up from two's bottom boundary = %d, want %d (one revealed)", got, 20)
	}
}

func TestSnapYOffsetAdjacentDoublePhotoNotSkipped(t *testing.T) {
	// A double photo — two adjacent inline images, each near viewport height
	// and one blank line apart (imgEnd+2 == imgStart) — must show both images
	// at their flush positions rather than skipping the first. The geometry
	// models the native render fit at a 27-row viewport: block rows run from
	// imgStart (image top) through capStart-1 (photo) with a caption line, and
	// one's caption is directly above two's top.
	lead := compose.ImageBlock{URL: "lead", ImgStart: 4, CapStart: 6, ImgEnd: 9}
	one := compose.ImageBlock{URL: "one", ImgStart: 11, CapStart: 31, ImgEnd: 31}
	two := compose.ImageBlock{URL: "two", ImgStart: 33, CapStart: 53, ImgEnd: 53}
	blocks := []compose.ImageBlock{lead, one, two}
	vpH := 27
	if one.ImgEnd+2 != two.ImgStart {
		t.Fatalf("blocks not adjacent: one end %d + 2 = %d, two start %d", one.ImgEnd, one.ImgEnd+2, two.ImgStart)
	}

	// Down: from the lead's caption, the first image is entered and revealed
	// flush at its top (never skipped to the second image's bottom), then
	// passes onto its caption, then the second image is revealed flush.
	downtab := []struct {
		name        string
		raw, before int
		want        int
	}{
		{"entry snap reveals one top", 11, 10, 11},
		{"one top flush", 12, 11, 31},
		{"one caption", 32, 31, 33},
		{"two top flush", 34, 33, 53},
	}
	for _, c := range downtab {
		if got := compose.SnapYOffset(c.raw, vpH, c.before, blocks, 1, false); got != c.want {
			t.Errorf("down %s = %d, want %d", c.name, got, c.want)
		}
	}

	// Up: from the lower image's bottom boundary the upper image is revealed
	// flush at its top (not skipped), then sinks to its bottom, then off.
	uptab := []struct {
		name         string
		before, want int
	}{
		{"from two bottom reveal one top", 27, 11},
		{"one top sinks to one bottom", 11, 5},
		{"one bottom off", 5, 0},
	}
	for _, c := range uptab {
		raw := c.before - 1
		if got := compose.SnapYOffset(raw, vpH, c.before, blocks, -1, false); got != c.want {
			t.Errorf("up %s from %d = %d, want %d", c.name, c.before, got, c.want)
		}
	}
}

func TestSnapYOffsetCaptionlessGridNextImageFlush(t *testing.T) {
	// A photo-grid's first image is caption-less (capStart == imgEnd+1) and
	// sits one blank line above an adjacent second image sharing the figure
	// caption. Scrolling down past the first image's EXIT (one past its last
	// photo row, the separator gap) would leave the second image resting one
	// line below the viewport top, needing an extra scroll to snap flush; the
	// snap must go straight to the second image's TOP.
	lead := compose.ImageBlock{URL: "lead", ImgStart: 5, CapStart: 35, ImgEnd: 35}
	one := compose.ImageBlock{URL: "one", ImgStart: 39, CapStart: 69, ImgEnd: 68} // caption-less grid photo
	two := compose.ImageBlock{URL: "two", ImgStart: 70, CapStart: 100, ImgEnd: 100}
	blocks := []compose.ImageBlock{lead, one, two}
	vpH := 37
	if one.CapStart != one.ImgEnd+1 {
		t.Fatalf("one not caption-less: capStart %d, imgEnd %d", one.CapStart, one.ImgEnd)
	}
	if two.ImgStart != one.ImgEnd+2 {
		t.Fatalf("two not adjacent: one end %d + 2 = %d, two start %d", one.ImgEnd, one.ImgEnd+2, two.ImgStart)
	}

	// Down from one's top: skip past the caption-less EXIT+gap to two's flush
	// top, so the second image never rests one line below the viewport top.
	if got := compose.SnapYOffset(40, vpH, 39, blocks, 1, false); got != two.ImgStart {
		t.Errorf("down from grid first photo top = %d, want %d (second photo flush)", got, two.ImgStart)
	}
	// The second image then skips onto its (shared) caption.
	if got := compose.SnapYOffset(two.ImgStart+1, vpH, two.ImgStart, blocks, 1, false); got != two.CapStart {
		t.Errorf("down from second photo top = %d, want %d (its caption)", got, two.CapStart)
	}
}

func TestSnapYOffsetAdjacentBlocksSkipGap(t *testing.T) {
	// Two adjacent image blocks (no body text between) are separated by one
	// blank line: the next block's top is exactly two lines past the previous
	// block's last line. A down move landing on that separator line must snap
	// to the next image's top, so the 1-vertical gap between photos is never
	// a resting scroll position.
	lead := compose.ImageBlock{URL: "lead", ImgStart: 5, CapStart: 19, ImgEnd: 19}
	inline := compose.ImageBlock{URL: "inline", ImgStart: 21, CapStart: 35, ImgEnd: 35}
	blocks := []compose.ImageBlock{lead, inline}

	cases := []struct {
		name      string
		offset    int
		before    int
		vpH       int
		direction int
		want      int
	}{
		// Down onto the separator line snaps to the next image's top.
		{"down onto the separator line snaps to the image top", 20, 19, 21, 1, 21},
		{"down onto the separator line snaps regardless of prior offset", 20, 18, 21, 1, 21},
		// The image's top itself is not snapped away.
		{"down at the image top unchanged", 21, 20, 21, 1, 21},
		// Non-adjacent blocks (body text between) are never snapped over: a
		// fully visible short image several lines below the fold stays put.
		{"down at a text line above a distant image unchanged", 20, 19, 21, 1, 20},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bs := blocks
			if c.name == "down at a text line above a distant image unchanged" {
				// The inline is not adjacent: text occupies the separator, and
				// the short image is fully visible so the bottom snap does not
				// preempt the unchanged outcome.
				bs = []compose.ImageBlock{lead, compose.ImageBlock{URL: "inline", ImgStart: 30, CapStart: 34, ImgEnd: 34}}
			}
			if got := compose.SnapYOffset(c.offset, c.vpH, c.before, bs, c.direction, false); got != c.want {
				t.Errorf("compose.SnapYOffset(%d,%d,%d) = %d, want %d", c.offset, c.vpH, c.before, got, c.want)
			}
		})
	}
}

func TestSnapYOffsetFullViewportBlockDoesNotLoop(t *testing.T) {
	// A block whose image plus wrapped caption fills the viewport exactly (a
	// long shared caption, as in a gallery) has its bottom-aligned position
	// equal to its top-aligned one: there is no distinct bottom stage. The
	// up move from the top-aligned position must scroll the image off (or
	// reveal the previous image) rather than snapping back onto itself and
	// looping forever.
	vpH := 22
	lead := compose.ImageBlock{URL: "lead", ImgStart: 4, CapStart: 6, ImgEnd: 9}
	one := compose.ImageBlock{URL: "one", ImgStart: 30, CapStart: 50, ImgEnd: 51} // fills vpH
	two := compose.ImageBlock{URL: "two", ImgStart: 60, CapStart: 80, ImgEnd: 81} // fills vpH
	blocks := []compose.ImageBlock{lead, one, two}

	cases := []struct {
		name   string
		offset int
		before int
		d      int
		want   int
	}{
		// Up from two's top (fills the viewport): no bottom stage, so scroll
		// it off; the target lands inside one, so one is revealed at its top.
		{"up from full-viewport two scrolls it off to one", 59, 60, -1, 30},
		// Up from one's top likewise scrolls it off, landing above the lead.
		{"up from full-viewport one scrolls it off", 29, 30, -1, 8},
		// Down into a full-viewport block still shows it (bottom-align lands
		// on its top), then the next down skips it; no loop.
		{"down into full-viewport one shows it", 31, 29, 1, 30},
		{"down from full-viewport one skips onto its caption", 31, 30, 1, 50},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := compose.SnapYOffset(c.offset, vpH, c.before, blocks, c.d, false); got != c.want {
				t.Errorf("compose.SnapYOffset(%d,%d,%d) = %d, want %d", c.offset, vpH, c.before, got, c.want)
			}
		})
	}
}

func TestScrollUpThroughFullViewportImageDoesNotLoop(t *testing.T) {
	// An inline image whose block fills the viewport exactly (20 image rows +
	// one caption line == viewport height), as a tall gallery photo. Scrolling
	// up from the bottom must never leave the offset stuck: each up move
	// either snaps or advances.
	a := imageArticle()
	a.Content = `<p>intro</p>` +
		`<figure><img src="inline.jpg" alt=""><figcaption>Short caption</figcaption></figure>` +
		strings.Repeat("<p>long paragraph body text</p>", 60)
	tall := make([]string, 20)
	for i := range tall {
		tall[i] = "IMG"
	}
	m, _ := newImageModel(t, tall)
	m.sess.ImgCache.Set("inline.jpg", testImg())
	m.article = m.newArticleState(a)
	if len(m.article.imageBlocks) < 2 {
		t.Fatalf("want lead + inline, got %d", len(m.article.imageBlocks))
	}
	vpH := m.article.viewport.Height
	inline := m.article.imageBlocks[1]
	// Precondition: the block fills the viewport (no distinct bottom stage).
	if inline.ImgEnd-inline.ImgStart+1 < vpH {
		t.Fatalf("block %d..%d (rows=%d) does not fill viewport %d", inline.ImgStart, inline.ImgEnd, inline.ImgEnd-inline.ImgStart+1, vpH)
	}
	m.article.viewport.GotoBottom()
	stuck := 0
	for step := 0; step < 500 && m.article.viewport.YOffset > 0; step++ {
		before := m.article.viewport.YOffset
		m.scrollArticle(func() { m.article.viewport.ScrollUp(1) }, true)
		if m.article.viewport.YOffset >= before {
			stuck++
			if stuck > 1 {
				t.Fatalf("step %d: up move stuck at offset %d", step, m.article.viewport.YOffset)
			}
		} else {
			stuck = 0
		}
	}
}

func TestSnapAfterRecompose(t *testing.T) {
	lead := compose.ImageBlock{URL: "lead", ImgStart: 4, CapStart: 6, ImgEnd: 9}
	inline := compose.ImageBlock{URL: "inline", ImgStart: 20, CapStart: 22, ImgEnd: 40}
	blocks := []compose.ImageBlock{lead, inline}
	cases := []struct {
		name   string
		offset int
		vpH    int
		want   int
	}{
		{"offset inside the lead photo snaps onto its caption", 5, 15, 6},
		{"offset inside the inline photo snaps to its top", 21, 15, 20},
		{"inline top materialized mid-window snaps forward to it", 10, 15, 20},
		{"offset below the inline block unchanged", 45, 15, 45},
		{"offset above the blocks unchanged", 2, 15, 2},
		{"inline top beyond the window bottom unchanged", 15, 3, 15},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := compose.SnapAfterRecompose(c.offset, c.vpH, blocks); got != c.want {
				t.Errorf("compose.SnapAfterRecompose(%d,%d) = %d, want %d", c.offset, c.vpH, got, c.want)
			}
		})
	}
}

func TestRecomposeArticleSnapsOffsetOutOfInlinePhoto(t *testing.T) {
	wide := strings.Repeat("I", 40)
	a := imageArticle()
	a.Content = `<p>intro text</p><p><img src="inline.jpg" alt="Alt"></p><p>outro text</p>` +
		strings.Repeat("<p>long paragraph body text</p>", 60)
	m, _ := newImageModel(t, []string{wide, wide})
	m.sess.ImgCache.Set("inline.jpg", testImg())
	m.article = m.newArticleState(a)

	if len(m.article.imageBlocks) != 2 {
		t.Fatalf("imageBlocks = %d, want 2 (lead + inline)", len(m.article.imageBlocks))
	}
	inline := m.article.imageBlocks[1]
	// An offset sitting inside the inline photo range (a load could bump the
	// offset there) must snap to the image's top on re-composition, matching
	// motion 1, rather than leaving the photo partially clipped.
	m.article.viewport.SetYOffset(inline.ImgStart + 1)
	m.recomposeArticle()
	if got := m.article.viewport.YOffset; got != inline.ImgStart {
		t.Errorf("offset after recompose = %d, want %d (inline top)", got, inline.ImgStart)
	}
}

func TestRecomposeKeepsBottomWhenOffsetInPhotoRange(t *testing.T) {
	// GotoBottom can land inside the last photo's range — the photo's caption
	// is cut off below the fold — so a re-composition (an image load resolving
	// on open) must not snap the offset back up to that photo's top: the user
	// asked for the article's end and must keep seeing it.
	tall := make([]string, 14)
	for i := range tall {
		tall[i] = "IMG"
	}
	a := imageArticle()
	a.Content = strings.Repeat("<p>lead text</p>", 8) +
		`<p>before</p><p><img src="inline.jpg" alt="Alt"></p>` +
		strings.Repeat("<p>tail text</p>", 8)
	m, _ := newImageModel(t, tall)
	m.sess.ImgCache.Set("inline.jpg", testImg())
	m.article = m.newArticleState(a)
	if len(m.article.imageBlocks) < 2 {
		t.Fatalf("want lead + inline, got %d", len(m.article.imageBlocks))
	}
	vpH := m.article.viewport.Height
	m.article.viewport.GotoBottom()
	bottom := m.article.viewport.YOffset
	inline := m.article.imageBlocks[1]
	// Precondition: the bottom lands inside the inline photo's range (the
	// position the old recompose snap would pull back up to its top).
	if !(inline.ImgStart <= bottom && bottom < inline.CapStart) {
		t.Fatalf("precondition not met: bottom %d not inside inline photo range %d..%d",
			bottom, inline.ImgStart, inline.CapStart)
	}
	m.recomposeArticle()
	if got := m.article.viewport.YOffset; got != len(m.article.lines)-vpH {
		t.Errorf("offset after recompose = %d, want the bottom %d, not the inline top %d",
			got, len(m.article.lines)-vpH, m.article.imageBlocks[1].ImgStart)
	}
}

func TestScrollSnapsInlineImageBottomThenTop(t *testing.T) {
	wide := strings.Repeat("I", 40)
	a := imageArticle()
	a.Content = strings.Repeat("<p>lead text</p>", 30) +
		`<p>before text</p><p><img src="inline.jpg" alt="Alt"></p><p>after text</p>` +
		strings.Repeat("<p>body text</p>", 30)
	m, _ := newImageModel(t, []string{wide, wide})
	m.sess.ImgCache.Set("inline.jpg", testImg())
	m.article = m.newArticleState(a)
	inline := m.article.imageBlocks[1]
	vpH := m.article.viewport.Height

	// Place the inline top exactly at the fold; the next down move snaps its
	// bottom to the viewport bottom (stage one).
	m.article.viewport.SetYOffset(inline.ImgStart - vpH)
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
	if got := m.article.viewport.YOffset; got != inline.ImgEnd-vpH+1 {
		t.Errorf("down into inline = %d, want %d (bottom-aligned)", got, inline.ImgEnd-vpH+1)
	}
	// The next down move snaps its top to the viewport top (stage two).
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
	if got := m.article.viewport.YOffset; got != inline.ImgStart {
		t.Errorf("down again = %d, want %d (top-aligned)", got, inline.ImgStart)
	}
	// The next down move skips it onto its caption.
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
	if got := m.article.viewport.YOffset; got != inline.CapStart {
		t.Errorf("down third = %d, want %d (its caption)", got, inline.CapStart)
	}
}

func TestScrollDownNeverLeavesInlineImageBlank(t *testing.T) {
	// Two tall inline images spaced closer than the viewport height, as in
	// the Intercept smoke-test article: skipping the first leaves the second's
	// top inside the window. A single-line down move must never leave an
	// inline image partially visible for two consecutive moves — the snap
	// resolves the skip's single transient frame on the next press (that frame
	// renders a halfblock preview, never a blank strip), so the user never
	// scrolls through an empty region.
	wide := strings.Repeat("I", 40)
	tall := make([]string, 14)
	for i := range tall {
		tall[i] = wide
	}
	a := imageArticle()
	a.Content = strings.Repeat("<p>lead text</p>", 10) +
		`<p>before one</p><p><img src="one.jpg" alt="One"></p>` +
		strings.Repeat("<p>gap text</p>", 4) +
		`<p>before two</p><p><img src="two.jpg" alt="Two"></p>` +
		strings.Repeat("<p>tail text</p>", 60)
	m, _ := newImageModel(t, tall)
	m.sess.ImgCache.Set("one.jpg", testImg())
	m.sess.ImgCache.Set("two.jpg", testImg())
	m.article = m.newArticleState(a)
	vpH := m.article.viewport.Height
	if len(m.article.imageBlocks) < 3 {
		t.Fatalf("want lead + two inline blocks, got %d", len(m.article.imageBlocks))
	}
	one, two := m.article.imageBlocks[1], m.article.imageBlocks[2]
	// Precondition: the images are close enough that a plain skip onto one's
	// caption would leave two's top inside the window (the transient case).
	if two.ImgStart >= one.CapStart+vpH {
		t.Fatalf("test images not close: two.top %d >= one.caption %d + vpH %d", two.ImgStart, one.CapStart, vpH)
	}

	partial := func(b compose.ImageBlock) bool {
		return b.ImgStart < m.article.viewport.YOffset+vpH && b.CapStart > m.article.viewport.YOffset &&
			!(b.ImgStart >= m.article.viewport.YOffset && b.CapStart <= m.article.viewport.YOffset+vpH)
	}
	stuck := 0
	sawPartial := false
	for step := 0; step < 500 && m.article.viewport.YOffset < len(m.article.lines)-vpH; step++ {
		m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
		any := false
		for _, b := range m.article.imageBlocks[1:] {
			if partial(b) {
				any = true
				break
			}
		}
		if any {
			sawPartial = true
			stuck++
			if stuck > 1 {
				t.Fatalf("step %d: inline image partially visible for %d consecutive moves at offset %d", step, stuck, m.article.viewport.YOffset)
			}
		} else {
			stuck = 0
		}
	}
	if !sawPartial {
		t.Fatal("test never exercised a partially visible inline image; composition changed")
	}
}

func TestScrollUpNeverLeavesInlineImageBlank(t *testing.T) {
	// Up mirror of TestScrollDownNeverLeavesInlineImageBlank: a single-line up
	// move must never leave an inline image partially visible for two
	// consecutive moves — the bottom scroll-off snaps it fully below the fold
	// and the transient frame renders a preview, never a blank strip.
	wide := strings.Repeat("I", 40)
	tall := make([]string, 14)
	for i := range tall {
		tall[i] = wide
	}
	a := imageArticle()
	a.Content = strings.Repeat("<p>lead text</p>", 10) +
		`<p>before one</p><p><img src="one.jpg" alt="One"></p>` +
		strings.Repeat("<p>gap text</p>", 4) +
		`<p>before two</p><p><img src="two.jpg" alt="Two"></p>` +
		strings.Repeat("<p>tail text</p>", 60)
	m, _ := newImageModel(t, tall)
	m.sess.ImgCache.Set("one.jpg", testImg())
	m.sess.ImgCache.Set("two.jpg", testImg())
	m.article = m.newArticleState(a)
	vpH := m.article.viewport.Height
	if len(m.article.imageBlocks) < 3 {
		t.Fatalf("want lead + two inline blocks, got %d", len(m.article.imageBlocks))
	}

	partial := func(b compose.ImageBlock) bool {
		return b.ImgStart < m.article.viewport.YOffset+vpH && b.CapStart > m.article.viewport.YOffset &&
			!(b.ImgStart >= m.article.viewport.YOffset && b.CapStart <= m.article.viewport.YOffset+vpH)
	}
	m.article.viewport.GotoBottom()
	stuck := 0
	for step := 0; step < 500 && m.article.viewport.YOffset > 0; step++ {
		m.scrollArticle(func() { m.article.viewport.ScrollUp(1) }, true)
		any := false
		for _, b := range m.article.imageBlocks[1:] {
			if partial(b) {
				any = true
				break
			}
		}
		if any {
			stuck++
			if stuck > 1 {
				t.Fatalf("step %d: inline image partially visible for %d consecutive moves at offset %d", step, stuck, m.article.viewport.YOffset)
			}
		} else {
			stuck = 0
		}
	}
}

func TestPageDownDoesNotSnapToInlineImage(t *testing.T) {
	wide := strings.Repeat("I", 40)
	a := imageArticle()
	a.Content = strings.Repeat("<p>lead text</p>", 30) +
		`<p>before text</p><p><img src="inline.jpg" alt="Alt"></p><p>after text</p>` +
		strings.Repeat("<p>body text</p>", 30)
	m, _ := newImageModel(t, []string{wide, wide})
	m.sess.ImgCache.Set("inline.jpg", testImg())
	m.article = m.newArticleState(a)
	inline := m.article.imageBlocks[1]

	// A page-down from just above the inline image lands a full page, even
	// mid-photo, rather than snapping to the image: paging must never skip the
	// text between images.
	m.article.viewport.SetYOffset(inline.ImgStart - 1)
	m.scrollArticle(func() { m.article.viewport.PageDown() }, false)
	if got, want := m.article.viewport.YOffset, inline.ImgStart-1+m.article.viewport.Height; got != want {
		t.Errorf("page-down offset = %d, want %d (full page, no snap)", got, want)
	}
}

func TestScrollSnapsThroughInlineImage(t *testing.T) {
	wide := strings.Repeat("I", 40)
	a := imageArticle()
	a.Content = `<p>intro text</p><p><img src="inline.jpg" alt="Alt"></p><p>outro text</p>` +
		strings.Repeat("<p>long paragraph body text</p>", 60)
	m, _ := newImageModel(t, []string{wide, wide})
	m.sess.ImgCache.Set("inline.jpg", testImg())
	m.article = m.newArticleState(a)

	inline := m.article.imageBlocks[1]
	// A down move from just above the inline photo snaps to its top so the
	// image is shown (motion 1), rather than skipping straight onto the
	// caption.
	m.article.viewport.SetYOffset(inline.ImgStart - 1)
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
	if got := m.article.viewport.YOffset; got != inline.ImgStart {
		t.Errorf("down into inline photo = %d, want %d (its top, motion 1)", got, inline.ImgStart)
	}
	// The next downward move skips the now-shown image onto its caption
	// (motion 2).
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
	if got := m.article.viewport.YOffset; got != inline.CapStart {
		t.Errorf("down again = %d, want %d (its caption, motion 2)", got, inline.CapStart)
	}
	// Down again scrolls within the caption line by line (no snap loop).
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
	if got := m.article.viewport.YOffset; got != inline.CapStart+1 {
		t.Errorf("down within the caption = %d, want %d", got, inline.CapStart+1)
	}
	// Up from a photo row reveals the full inline image.
	m.article.viewport.SetYOffset(inline.ImgStart + 1)
	m.scrollArticle(func() { m.article.viewport.ScrollUp(1) }, true)
	if got := m.article.viewport.YOffset; got != inline.ImgStart {
		t.Errorf("up from an inline photo row = %d, want %d (reveal)", got, inline.ImgStart)
	}
	// Up from just below the inline block lands on its caption line.
	m.article.viewport.SetYOffset(inline.ImgEnd + 1)
	m.scrollArticle(func() { m.article.viewport.ScrollUp(1) }, true)
	if got := m.article.viewport.YOffset; got != inline.ImgEnd {
		t.Errorf("up from below the inline block = %d, want %d (its caption)", got, inline.ImgEnd)
	}
}

func TestScrollSnapsAcrossAdjacentPhotosSkipsGap(t *testing.T) {
	// Two adjacent photos (lead + inline with no body text between) are
	// separated by one blank line. Scrolling down from the lead must snap
	// straight from the lead's caption to the inline's top, never resting on
	// that separator line; scrolling back up must likewise snap, never
	// pausing on the gap.
	wide := strings.Repeat("I", 40)
	tall := make([]string, 14)
	for i := range tall {
		tall[i] = wide
	}
	a := imageArticle()
	a.Content = `<p><img src="inline.jpg" alt="Alt"></p>` +
		strings.Repeat("<p>body text</p>", 60)
	m, _ := newImageModel(t, tall)
	m.sess.ImgCache.Set("inline.jpg", testImg())
	m.article = m.newArticleState(a)
	lead, inline := m.article.imageBlocks[0], m.article.imageBlocks[1]
	gap := inline.ImgStart - 1
	if lead.ImgEnd+2 != inline.ImgStart {
		t.Fatalf("photos not adjacent: lead end %d + 2 = %d, inline start %d", lead.ImgEnd, lead.ImgEnd+2, inline.ImgStart)
	}

	// Down: the lead is pre-consumed (skip onto its caption), then a single
	// move snaps to the inline's top without pausing on the gap.
	m.article.viewport.SetYOffset(0)
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
	if got := m.article.viewport.YOffset; got != lead.CapStart {
		t.Fatalf("down into the lead = %d, want %d (its caption)", got, lead.CapStart)
	}
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
	if got := m.article.viewport.YOffset; got != inline.ImgStart {
		t.Errorf("down from the lead caption = %d, want %d (inline top, gap %d skipped)", got, inline.ImgStart, gap)
	}
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) }, true)
	if got := m.article.viewport.YOffset; got != inline.CapStart {
		t.Errorf("down again = %d, want %d (inline caption)", got, inline.CapStart)
	}

	// Up: from the inline's top the snap steps back through the inline's own
	// stages, never pausing on the separator gap.
	m.article.viewport.SetYOffset(inline.ImgStart)
	m.scrollArticle(func() { m.article.viewport.ScrollUp(1) }, true)
	if got := m.article.viewport.YOffset; got == gap || got == gap+1 {
		t.Errorf("up move rested on the separator gap (offset %d); want a snap", got)
	}
}

func TestScrollUpFromBottomBoundarySnapsPhotoOut(t *testing.T) {
	// A photo resting at its bottom boundary (caption on the viewport's last
	// line) clears out of view in a single scroll-k: its top lands at the
	// fold, with no pause on the caption and no intermediate reveal that
	// leaves the photo partially visible. For a double photo (two adjacent
	// inline images), scrolling the lower image out reveals the upper one
	// flush at its top rather than skipping it, so each image reaches its
	// aligned positions.
	wide := strings.Repeat("I", 40)
	tall := make([]string, 14)
	for i := range tall {
		tall[i] = wide
	}
	a := imageArticle()
	a.Content = `<p><img src="one.jpg" alt="One"></p>` +
		`<p><img src="two.jpg" alt="Two"></p>` +
		strings.Repeat("<p>body text</p>", 60)
	m, _ := newImageModel(t, tall)
	m.sess.ImgCache.Set("one.jpg", testImg())
	m.sess.ImgCache.Set("two.jpg", testImg())
	m.article = m.newArticleState(a)
	vpH := m.article.viewport.Height
	if len(m.article.imageBlocks) < 3 {
		t.Fatalf("want lead + two inline blocks, got %d", len(m.article.imageBlocks))
	}
	one, two := m.article.imageBlocks[1], m.article.imageBlocks[2]
	bottom := two.ImgEnd - vpH + 1

	m.article.viewport.SetYOffset(bottom)
	m.scrollArticle(func() { m.article.viewport.ScrollUp(1) }, true)
	if got := m.article.viewport.YOffset; got != one.ImgStart {
		t.Errorf("up from lower photo bottom boundary = %d, want %d (upper photo revealed flush)", got, one.ImgStart)
	}
}

func TestInlineImageLeadDuplicateSkipped(t *testing.T) {
	wide := strings.Repeat("I", 40)
	a := imageArticle()
	// The lead image URL also appears in the body; the duplicate top lead is
	// suppressed and the body copy renders in place, so the photo appears once
	// at its natural position in the article flow, not as a top-of-article lead.
	a.Content = `<p>intro</p><p><img src="` + a.ImageURL + `"></p><p>outro</p>` +
		strings.Repeat("<p>body text</p>", 20)
	m, _ := newImageModel(t, []string{wide, wide})
	m.article = m.newArticleState(a)

	if len(m.article.imageBlocks) != 1 {
		t.Fatalf("imageBlocks = %d, want 1 (lead suppressed; body copy shown)", len(m.article.imageBlocks))
	}
	if len(m.article.imageURLs) != 1 || m.article.imageURLs[0] != a.ImageURL {
		t.Errorf("imageURLs = %v, want just the lead URL once", m.article.imageURLs)
	}
	lines := strippedLines(m.article.lines)
	introIdx := -1
	for i, l := range lines {
		if strings.Contains(l, "intro") {
			introIdx = i
			break
		}
	}
	if introIdx < 0 {
		t.Fatalf("intro text not found: %q", strings.Join(lines, "\n"))
	}
	block := m.article.imageBlocks[0]
	if block.ImgStart <= introIdx {
		t.Errorf("image block at line %d should render after intro at line %d (body copy, not a top lead)", block.ImgStart, introIdx)
	}
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "outro") {
		t.Errorf("outro text missing around the body image: %q", joined)
	}
}

func TestSuppressedLeadNotFetchedOrStoredSeparately(t *testing.T) {
	wide := strings.Repeat("I", 40)
	a := imageArticle()
	// Dropsite-style: the enclosure and the first body image are different CDN
	// transform URLs of the same source photo. Only the body URL should be
	// fetched/stored; the enclosure URL is suppressed and never appears in
	// imageURLs (which drives the fetch and the article_images position).
	src := "https%3A%2F%2Fsubstack-post-media.s3.amazonaws.com%2Fpublic%2Fimages%2Fac2c7f0f-6fe9-48ad-8b7e-23d3974439ef_1600x900.jpeg"
	a.ImageURL = "https://substackcdn.com/image/fetch/$s_!GO48!,f_auto,q_auto:good,fl_progressive:steep/" + src
	bodyURL := "https://substackcdn.com/image/fetch/$s_!GO48!,w_2400,c_limit,f_auto,q_auto:good,fl_progressive:steep/" + src
	a.Content = `<p>intro</p><p><img src="` + bodyURL + `"></p><p>outro</p>` +
		strings.Repeat("<p>body text</p>", 20)
	m, _ := newImageModel(t, []string{wide, wide})
	m.sess.ImgCache.Set(bodyURL, testImg())
	m.article = m.newArticleState(a)

	if len(m.article.imageBlocks) != 1 {
		t.Fatalf("imageBlocks = %d, want 1 (lead suppressed; body copy shown)", len(m.article.imageBlocks))
	}
	if m.article.imageBlocks[0].URL != bodyURL {
		t.Errorf("block url = %q, want the body URL %q", m.article.imageBlocks[0].URL, bodyURL)
	}
	if len(m.article.imageURLs) != 1 || m.article.imageURLs[0] != bodyURL {
		t.Errorf("imageURLs = %v, want only the body URL (enclosure not fetched/stored)", m.article.imageURLs)
	}
	if m.article.imageURLs[0] == a.ImageURL {
		t.Errorf("enclosure URL %q should be suppressed, not stored at position 0", a.ImageURL)
	}
}

func TestInlineImageRepeatedURLShownOnce(t *testing.T) {
	wide := strings.Repeat("I", 40)
	a := imageArticle()
	a.Content = `<p>one</p><p><img src="a.jpg" alt="A"></p><p>two</p><p><img src="a.jpg" alt="A again"></p><p>three</p>`
	m, _ := newImageModel(t, []string{wide, wide})
	m.sess.ImgCache.Set("a.jpg", testImg())
	m.article = m.newArticleState(a)

	if len(m.article.imageBlocks) != 2 {
		t.Fatalf("imageBlocks = %d, want 2 (lead + one inline occurrence)", len(m.article.imageBlocks))
	}
	if m.article.imageBlocks[1].URL != "a.jpg" {
		t.Errorf("inline block url = %q, want a.jpg", m.article.imageBlocks[1].URL)
	}
	joined := strings.Join(strippedLines(m.article.lines), "\n")
	if !strings.Contains(joined, "one") || !strings.Contains(joined, "two") || !strings.Contains(joined, "three") {
		t.Errorf("body text missing: %q", joined)
	}
}

func TestNativeClippedImageDoesNotPaintOverBorder(t *testing.T) {
	wide := strings.Repeat("I", 40)
	a := imageArticle()
	a.Content = `<p>intro</p><p><img src="inline.jpg" alt="A"></p>` +
		strings.Repeat("<p>long paragraph body text</p>", 60)
	m, _ := newImageModel(t, []string{wide, wide})
	m.sess.ImgNative = imgpkg.NativeRenderer{Protocol: imgpkg.ProtocolKitty}
	m.sess.ImgCache = imgpkg.NewCache()
	m.sess.ImgCache.Set(imageArticle().ImageURL, testImg())
	m.sess.ImgCache.Set("inline.jpg", testImg())
	m.view = viewArticle
	contentW, vpH, _ := render.ContentGeom(m.width, m.height, m.sess.Config().Display.PaddingX, m.sess.Config().Display.PaddingY)
	// A native block taller than the viewport, as a real fitted photo can be
	// when its top sits mid-window: its placement would extend past the fold
	// and draw over the article border.
	tall := make([]string, vpH+40)
	esc := "\x1b_Ga=T,f=100,q=1,i=7,p=7,c=2,r=2;AAAA\x1b\\"
	tall[0] = esc + "  "
	for i := 1; i < len(tall); i++ {
		tall[i] = "  "
	}
	m.sess.ImgNatives.Set(imgpkg.NativeKey("inline.jpg", contentW, vpH), tall)
	m.article = m.newArticleState(a)
	inline := m.article.imageBlocks[1]
	if !inline.NativeImg {
		t.Fatal("inline block should be flagged native")
	}

	// Top visible, bottom below the fold: the frame deletes the placement and
	// suppresses the transmit so the image cannot paint over the border.
	m.article.viewport.SetYOffset(inline.ImgStart)
	view := frameView(m)
	if !strings.Contains(view, imgpkg.DeleteByID(inline.NativeID)) {
		t.Errorf("clipped image placement should be deleted by its render id: %q", view)
	}
	if strings.Contains(view, "a=T,f=100") {
		t.Errorf("clipped image transmit should be suppressed: %q", view)
	}
}

func TestInlineLinkTextCaptionFitsNativeBlock(t *testing.T) {
	// A linked inline image whose link text is long: the composed caption
	// (which includes the link text) drives the native fit, so the fitted
	// image plus the wrapped caption fits the viewport with a distinct bottom
	// stage — not the full-viewport overflow a shortened-caption fit allowed.
	longText := strings.Repeat("Promo banner label text ", 12)
	a := imageArticle()
	a.Content = `<p>intro</p><p><a href="https://example.com/promo"><img src="inline.jpg" alt=""><span>` +
		longText + `</span></a></p>` +
		strings.Repeat("<p>long paragraph body text</p>", 60)
	m, _ := newImageModel(t, []string{"IMG"})
	m.sess.ImgCache.Set("inline.jpg", testImg())
	m.sess.ImgNative = imgpkg.NativeRenderer{Protocol: imgpkg.ProtocolKitty}
	m.sess.ImgPhotos.Set("inline.jpg", pngBytes(t, 40, 40))
	m.view = viewArticle
	m.article = m.newArticleState(a)

	// inlineAttrFor must report the composed caption including the link text,
	// so the native fit reserves its real wrapped line count.
	cap := m.inlineAttrFor("inline.jpg")
	if !strings.Contains(cap, "Promo banner") || strings.Contains(cap, "photo:") {
		t.Fatalf("inlineAttrFor = %q, want the link-text caption", cap)
	}

	// Render the inline photo natively with that caption and re-compose.
	width, vpH, _ := render.ContentGeom(m.width, m.height, m.sess.Config().Display.PaddingX, m.sess.Config().Display.PaddingY)
	msg := imgpkg.NativeCmd(m.sess.ImgNative, m.sess.ImgPhotos, m.sess.ImgNatives, "inline.jpg", width, vpH, cap, m.article.headerLines, compose.CaptionWidth(width))()
	nm, ok := msg.(imgpkg.NativeMsg)
	if !ok {
		t.Fatalf("expected NativeMsg, got %T", msg)
	}
	m.sess.ImgNatives.Set(imgpkg.NativeKey("inline.jpg", width, vpH), nm.Lines)
	m.recomposeArticle()
	inline := m.article.imageBlocks[1]
	if !inline.NativeImg {
		t.Fatal("inline block should be native after the render")
	}
	// The fitted image plus its composed caption fits the viewport budget, so
	// the two-stage snap keeps a distinct bottom stage.
	if inline.ImgEnd-inline.ImgStart+1 >= vpH {
		t.Errorf("native inline block %d..%d does not fit viewport %d", inline.ImgStart, inline.ImgEnd, vpH)
	}
}

func TestClippedNativeImageShowsHalfblockPreview(t *testing.T) {
	// A partially visible native photo (its top in the window, its bottom
	// below the fold) renders a halfblock preview of its visible rows instead
	// of a blank strip where its deleted placement draws empty.
	a := imageArticle()
	a.Content = `<p>intro</p><p><img src="inline.jpg" alt="A"></p>` +
		strings.Repeat("<p>long paragraph body text</p>", 60)
	m, _ := newImageModel(t, []string{"PREVIEW1", "PREVIEW2"})
	m.sess.ImgNative = imgpkg.NativeRenderer{Protocol: imgpkg.ProtocolKitty}
	m.sess.ImgCache = imgpkg.NewCache()
	m.sess.ImgCache.Set(imageArticle().ImageURL, testImg())
	m.sess.ImgCache.Set("inline.jpg", testImg())
	m.view = viewArticle
	contentW, vpH, _ := render.ContentGeom(m.width, m.height, m.sess.Config().Display.PaddingX, m.sess.Config().Display.PaddingY)
	tall := make([]string, vpH+40)
	esc := "\x1b_Ga=T,f=100,q=1,i=7,p=7,c=2,r=2;AAAA\x1b\\"
	tall[0] = esc + "  "
	for i := 1; i < len(tall); i++ {
		tall[i] = "  "
	}
	m.sess.ImgNatives.Set(imgpkg.NativeKey("inline.jpg", contentW, vpH), tall)
	m.article = m.newArticleState(a)
	inline := m.article.imageBlocks[1]
	if !inline.NativeImg {
		t.Fatal("inline block should be flagged native")
	}

	// Partial at the bottom edge: the visible rows must carry the preview.
	m.article.viewport.SetYOffset(inline.ImgStart)
	view := frameView(m)
	if !strings.Contains(view, "PREVIEW1") {
		t.Errorf("clipped image should show its halfblock preview, got: %q", view)
	}
	// The transmit and placement handling is unchanged.
	if !strings.Contains(view, imgpkg.DeleteByID(inline.NativeID)) {
		t.Errorf("clipped image placement should still be deleted by its render id: %q", view)
	}
	if strings.Contains(view, "a=T,f=100") {
		t.Errorf("clipped image transmit should be suppressed: %q", view)
	}
}

func TestNativeImageClearDeletesScrolledOutByID(t *testing.T) {
	m, _ := newImageModel(t, []string{"IMG1", "IMG2"})
	m.sess.ImgNative = imgpkg.NativeRenderer{Protocol: imgpkg.ProtocolKitty}
	m.view = viewArticle
	a := imageArticle()
	a.Content = `<p>intro</p><p><img src="inline.jpg" alt="A"></p>` +
		strings.Repeat("<p>long paragraph body text</p>", 60)
	m.sess.ImgCache.Set("inline.jpg", testImg())
	m.article = m.newArticleState(a)
	for i := range m.article.imageBlocks {
		m.article.imageBlocks[i].NativeImg = true
		m.article.imageBlocks[i].NativeID = uint32(100 + i)
	}
	lead := m.article.imageBlocks[0]
	inline := m.article.imageBlocks[1]

	// Lead scrolled off while the inline is on screen: the clear must delete
	// the lead's placement by its recorded id and leave the visible inline
	// alone.
	m.article.viewport.SetYOffset(inline.ImgStart)
	got := m.computeNativeClear()
	if !strings.Contains(got, imgpkg.DeleteByID(lead.NativeID)) {
		t.Errorf("clear should delete the scrolled-out lead placement by its render id: %q", got)
	}
	if strings.Contains(got, imgpkg.DeleteByID(inline.NativeID)) {
		t.Errorf("clear must not delete the visible inline placement: %q", got)
	}

	// Everything on screen: no clear at all.
	m.article.viewport.SetYOffset(lead.ImgStart)
	if got := m.computeNativeClear(); strings.Contains(got, "a=d") {
		t.Errorf("clear issued while all images visible: %q", got)
	}
}

func TestNativeKittyResizeDeletesPriorIDBeforeTransmit(t *testing.T) {
	m, _ := newImageModel(t, nil)
	m.sess.ImgNative.Protocol = imgpkg.ProtocolKitty
	// A large photo whose fitted box tracks the viewport geometry, so a
	// viewport change genuinely mints a new render id (a small image renders
	// at a fixed capped width and would reuse the id).
	m.sess.ImgPhotos.Set(imageArticle().ImageURL, pngBytes(t, 480, 320))

	// Render at the initial geometry and compose; one frame transmits the
	// initial size and records its id.
	nativeRenderLines(t, m, imageArticle())
	m.article = m.newArticleState(imageArticle())
	m.view = viewArticle
	if !m.article.nativeImg {
		t.Fatal("expected a native render at the initial size")
	}
	oldID := m.article.imageBlocks[0].NativeID
	if oldID == 0 {
		t.Fatal("native block must record its render id")
	}
	if got := frameView(m); !strings.Contains(got, fmt.Sprintf("i=%d", oldID)) {
		t.Fatalf("initial frame should transmit i=%d: %q", oldID, got)
	}

	// Resize: re-render at a new viewport height (the render cache key and the
	// height fit both depend on it, so the fitted box changes and mints a new
	// id; the frame switches straight to it with no placeholder gap) and
	// re-compose.
	m.height += 10
	nativeRenderLines(t, m, imageArticle())
	m.article = m.newArticleState(imageArticle())
	m.view = viewArticle
	if !m.article.nativeImg {
		t.Fatal("expected a native render at the resized geometry")
	}
	newID := m.article.imageBlocks[0].NativeID
	if newID == 0 || newID == oldID {
		t.Fatalf("resize should mint a new render id: old %d new %d", oldID, newID)
	}

	// The resize frame must delete the prior size's image by its id before
	// the new transmit, so the terminal never holds two sizes of one photo.
	got := frameView(m)
	del := imgpkg.DeleteByID(oldID)
	if i := strings.Index(got, del); i < 0 || i > strings.Index(got, "\x1b_Ga=T") {
		t.Errorf("resize frame must delete the prior id before transmitting the new one: %q", got)
	}
	if !strings.Contains(got, fmt.Sprintf("i=%d", newID)) {
		t.Errorf("resize frame should transmit the new id %d: %q", newID, got)
	}
}

func TestNativeKittyExitDeletesEveryRecordedID(t *testing.T) {
	a := imageArticle()
	a.Content = `<p>intro</p><p><img src="inline.jpg" alt="A"></p>` +
		strings.Repeat("<p>long paragraph body text</p>", 60)
	m, _ := newImageModel(t, nil)
	m.sess.ImgNative = imgpkg.NativeRenderer{Protocol: imgpkg.ProtocolKitty}
	m.sess.ImgPhotos.Set(a.ImageURL, pngBytes(t, 40, 40))
	m.sess.ImgPhotos.Set("inline.jpg", pngBytes(t, 40, 40))

	// Render the lead natively and compose once (the PhotoMsg → NativeCmd →
	// NativeMsg flow), then render the inline natively and re-compose so both
	// blocks record their ids.
	nativeRenderLines(t, m, a)
	m.article = m.newArticleState(a)
	m.view = viewArticle
	width, vpH, _ := render.ContentGeom(m.width, m.height, m.sess.Config().Display.PaddingX, m.sess.Config().Display.PaddingY)
	msg := imgpkg.NativeCmd(m.sess.ImgNative, m.sess.ImgPhotos, m.sess.ImgNatives, "inline.jpg", width, vpH, m.inlineAttrFor("inline.jpg"), m.article.headerLines, compose.CaptionWidth(width))()
	nm, ok := msg.(imgpkg.NativeMsg)
	if !ok {
		t.Fatalf("expected NativeMsg for the inline image, got %T", msg)
	}
	m.sess.ImgNatives.Set(imgpkg.NativeKey("inline.jpg", width, vpH), nm.Lines)
	m.recomposeArticle()
	var ids []uint32
	for _, b := range m.article.imageBlocks {
		if b.NativeImg {
			if b.NativeID == 0 {
				t.Fatalf("native block for %s should record its render id", b.URL)
			}
			ids = append(ids, b.NativeID)
		}
	}
	if len(ids) < 2 {
		t.Fatalf("expected lead and inline native blocks with recorded ids, got %d", len(ids))
	}

	// A frame at the top records the lead id as transmitted; the inline sits
	// below the fold and is deleted by id.
	if got := frameView(m); !strings.Contains(got, fmt.Sprintf("i=%d", ids[0])) {
		t.Fatalf("top frame should transmit the lead id %d: %q", ids[0], got)
	}

	// Scrolling to the inline deletes the scrolled-out lead by id while the
	// inline records its id, so both remain tracked as held by the terminal.
	m.article.viewport.SetYOffset(m.article.imageBlocks[1].ImgStart)
	if got := frameView(m); !strings.Contains(got, imgpkg.DeleteByID(ids[0])) {
		t.Fatalf("scrolled-out lead should be deleted by its recorded id: %q", got)
	}

	// Leaving the article frees every recorded id by that id — a delete-all
	// would only clear visible placements and leave the terminal's cached
	// image data behind.
	m.backToList()
	got := frameView(m)
	for _, id := range ids {
		if !strings.Contains(got, imgpkg.DeleteByID(id)) {
			t.Errorf("exit frame must delete recorded id %d: %q", id, got)
		}
	}
	if !strings.Contains(got, "d=I") {
		t.Errorf("exit frame must use the data-freeing delete form d=I: %q", got)
	}
	if strings.Contains(got, "a=d,d=a") {
		t.Errorf("exit frame should free images by id, not delete-all: %q", got)
	}
}

func TestNativeKittyGhosttyExitEmitsDeleteAllFallback(t *testing.T) {
	// Ghostty's kitty-graphics delete handling is partial, so a placement can
	// survive the per-id by-id deletes; the frame that leaves the article view
	// must additionally emit a delete-all clear so no photo lingers over the
	// list.
	m, _ := newImageModel(t, nil)
	m.sess.ImgNative = imgpkg.NativeRenderer{Protocol: imgpkg.ProtocolKitty, Ghostty: true}
	m.sess.ImgPhotos.Set(imageArticle().ImageURL, pngBytes(t, 40, 40))
	nativeRenderLines(t, m, imageArticle())
	m.article = m.newArticleState(imageArticle())
	m.view = viewArticle
	if !m.article.nativeImg {
		t.Fatal("expected a native render")
	}
	leadID := m.article.imageBlocks[0].NativeID
	if leadID == 0 {
		t.Fatal("native block should record its kitty render id")
	}
	if got := frameView(m); !strings.Contains(got, fmt.Sprintf("i=%d", leadID)) {
		t.Fatalf("article frame should transmit the photo id %d: %q", leadID, got)
	}
	m.backToList()
	got := frameView(m)
	if !strings.Contains(got, imgpkg.DeleteByID(leadID)) {
		t.Errorf("ghostty exit frame must delete the placement by its recorded id: %q", got)
	}
	if !strings.Contains(got, "a=d,d=a") {
		t.Errorf("ghostty exit frame must also carry the delete-all fallback: %q", got)
	}
}

func TestNativeKittyNonGhosttyExitHasNoDeleteAll(t *testing.T) {
	// A kitty-family terminal that handles d=I correctly needs only the per-id
	// deletes; the delete-all fallback must not appear on non-Ghostty terminals.
	m, _ := newImageModel(t, nil)
	m.sess.ImgNative = imgpkg.NativeRenderer{Protocol: imgpkg.ProtocolKitty}
	m.sess.ImgPhotos.Set(imageArticle().ImageURL, pngBytes(t, 40, 40))
	nativeRenderLines(t, m, imageArticle())
	m.article = m.newArticleState(imageArticle())
	m.view = viewArticle
	if !m.article.nativeImg {
		t.Fatal("expected a native render")
	}
	leadID := m.article.imageBlocks[0].NativeID
	if leadID == 0 {
		t.Fatal("native block should record its kitty render id")
	}
	if got := frameView(m); !strings.Contains(got, fmt.Sprintf("i=%d", leadID)) {
		t.Fatalf("article frame should transmit the photo id %d: %q", leadID, got)
	}
	m.backToList()
	got := frameView(m)
	if !strings.Contains(got, imgpkg.DeleteByID(leadID)) {
		t.Errorf("kitty exit frame must delete the placement by its recorded id: %q", got)
	}
	if strings.Contains(got, "a=d,d=a") {
		t.Errorf("kitty exit frame must not carry the delete-all on a non-Ghostty terminal: %q", got)
	}
}
