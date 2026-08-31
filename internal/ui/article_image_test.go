package ui

import (
	"bytes"
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
)

// dimsCall records a render request the fake renderer received.
type dimsCall struct{ w, maxH int }

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
	m.imgRenderer = fr
	m.imgCache.Set(imageArticle().ImageURL, testImg())
	return m, fr
}

func TestArticleImageBlockComposedBelowHeader(t *testing.T) {
	m, fr := newImageModel(t, []string{"IMG1", "IMG2"})
	m.article = m.newArticleState(imageArticle())

	if len(fr.calls) == 0 {
		t.Fatal("renderer was not called")
	}
	// Header (URL, blank, title, author) is 4 lines; the block sits below it
	// after one blank line and covers the two image lines plus the wrapped
	// attribution (the 4-cell fake photo wraps "photo: example" onto four
	// lines).
	if m.article.imgStart != 5 || m.article.imgEnd != 10 {
		t.Errorf("img range = %d..%d, want 5..10", m.article.imgStart, m.article.imgEnd)
	}
	lines := strippedLines(m.article.lines)
	if !strings.HasPrefix(lines[0], "https://example.com/a") {
		t.Errorf("first content line = %q, want the URL", lines[0])
	}
	if !strings.HasSuffix(lines[5], "IMG1") || !strings.HasSuffix(lines[6], "IMG2") {
		t.Errorf("image lines not below the header: %q", lines[5:7])
	}
	for i, want := range []string{"phot", "o:", "exam", "ple"} {
		if !strings.HasSuffix(strings.TrimRight(lines[7+i], " "), want) {
			t.Errorf("wrapped attribution line %d = %q, want it to end with %q", i, lines[7+i], want)
		}
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
		{"config off", func(m *Model) { m.cfg.Display.Images = "off" }},
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
	m.imgCache = imgpkg.NewCache() // drop the preset image
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
	m.imgCache = imgpkg.NewCache() // no image yet
	m.article = m.newArticleState(imageArticle())
	// Scroll into the body past the insertion point.
	m.article.viewport.ScrollDown(1)
	oldOffset := m.article.viewport.YOffset
	if oldOffset <= 0 {
		t.Fatal("expected to have scrolled past the top")
	}

	m.imgCache.Set(imageArticle().ImageURL, testImg())
	m.view = viewArticle
	m.onBlockLoaded(imgpkg.BlockMsg{Key: imageArticle().ImageURL, Lines: []string{"IMG1", "IMG2"}})

	if m.article.imgStart <= 0 {
		t.Fatalf("image block not composed after load: imgStart=%d", m.article.imgStart)
	}
	// The inserted lines are the two image lines, the four wrapped
	// attribution lines, and the blank below the block; the blank above it
	// is the separator that already divided header and body.
	if want := oldOffset + 7; m.article.viewport.YOffset != want {
		t.Errorf("offset = %d, want %d (bumped by inserted lines)", m.article.viewport.YOffset, want)
	}
	_ = fr
}

func TestImageLoadAtTopInsertsBlockAbove(t *testing.T) {
	m, _ := newImageModel(t, []string{"IMG1"})
	m.imgCache = imgpkg.NewCache()
	m.article = m.newArticleState(imageArticle())
	if m.article.viewport.YOffset != 0 {
		t.Fatal("expected offset 0 at top")
	}
	m.imgCache.Set(imageArticle().ImageURL, testImg())
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
	m.imgCache = imgpkg.NewCache()
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
	m.cfg.Display.Images = "on"
	a := imageArticle()
	a.GUID = "g-open"
	id, err := m.store.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	full, err := m.store.GetArticle(id)
	if err != nil {
		t.Fatal(err)
	}
	m.imgCache.Set(full.ImageURL, testImg())
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
	insertArticle(t, m.store, "one", nil)
	m.loadList()
	if cmd := m.openArticle(); cmd != nil {
		t.Errorf("openArticle without an image URL should return nil, got %T", cmd)
	}
}

func TestRestoreArticleFiresImageLoadCmd(t *testing.T) {
	withLocalZone(t, time.UTC)
	st := openSharedStore(t)
	m1 := modelOn(t, st)
	m1.cfg.Display.Images = "on"
	if _, err := m1.store.UpsertArticle(imageArticle()); err != nil {
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
	m2.cfg.Display.Images = "on"
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
	m2.imgCache.Set(imageArticle().ImageURL, testImg())
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
			if got := snapYOffset(c.offset, c.vpH, c.start, c.end, c.cap, c.d, c.native); got != c.want {
				t.Errorf("snapYOffset(%d,%d,%d,%d,%d,%d,%v) = %d, want %d", c.offset, c.vpH, c.start, c.end, c.cap, c.d, c.native, got, c.want)
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
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) })
	if want := m.article.imgStart + 2; m.article.capStart != want {
		t.Fatalf("caption start = %d, want %d (two photo rows above it)", m.article.capStart, want)
	}
	if got := m.article.viewport.YOffset; got != m.article.capStart {
		t.Errorf("offset after down scroll = %d, want %d (the caption)", got, m.article.capStart)
	}

	// Down again scrolls within the caption line by line (no snap loop).
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) })
	if got := m.article.viewport.YOffset; got != m.article.capStart+1 {
		t.Errorf("offset scrolling within the caption = %d, want %d", got, m.article.capStart+1)
	}

	// Up from a caption line scrolls the caption upward line by line rather
	// than snapping to the photo.
	m.scrollArticle(func() { m.article.viewport.ScrollUp(1) })
	if got := m.article.viewport.YOffset; got != m.article.capStart {
		t.Errorf("offset after up scroll = %d, want %d (the caption's top line)", got, m.article.capStart)
	}
	// Up from a photo row reveals the full image.
	m.article.viewport.SetYOffset(m.article.imgStart + 1)
	m.scrollArticle(func() { m.article.viewport.ScrollUp(1) })
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
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) })
	if got := m.article.viewport.YOffset; got != m.article.capStart {
		t.Errorf("offset down over a fully visible photo = %d, want %d (the first non-photo line)", got, m.article.capStart)
	}

	// The next downward scroll continues into the body text rather than
	// re-skipping the block.
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) })
	if got, want := m.article.viewport.YOffset, m.article.capStart+1; got != want {
		t.Errorf("offset after the next down scroll = %d, want %d", got, want)
	}

	// Page-down that lands past the block preserves the full-page scroll and
	// does not snap back onto the photo or caption.
	m.article.viewport.SetYOffset(m.article.imgStart)
	m.scrollArticle(func() { m.article.viewport.PageDown() })
	if got := m.article.viewport.YOffset; got < m.article.imgEnd {
		t.Errorf("page-down offset = %d, want it past the block end %d", got, m.article.imgEnd)
	}

	// Up from just below the block lands on the caption line (the block's
	// last line), scrolling the caption up line by line rather than snapping
	// to the photo, per the narrowed upward reveal rule.
	m.article.viewport.SetYOffset(m.article.imgEnd + 1)
	m.scrollArticle(func() { m.article.viewport.ScrollUp(1) })
	if got := m.article.viewport.YOffset; got != m.article.imgEnd {
		t.Errorf("offset up from below the block = %d, want %d (the caption)", got, m.article.imgEnd)
	}

	// Up from a photo row reveals the full image.
	m.article.viewport.SetYOffset(m.article.imgStart + 1)
	m.scrollArticle(func() { m.article.viewport.ScrollUp(1) })
	if got := m.article.viewport.YOffset; got != m.article.imgStart {
		t.Errorf("offset up from a photo row = %d, want %d (reveal)", got, m.article.imgStart)
	}

	// Native: up from the photo top with a fully visible photo skips straight
	// to the article top, since re-transmitting a native photo per header step
	// is costly.
	m.article.nativeImg = true
	m.article.viewport.SetYOffset(m.article.imgStart)
	m.scrollArticle(func() { m.article.viewport.ScrollUp(1) })
	if got := m.article.viewport.YOffset; got != 0 {
		t.Errorf("native up from the photo top = %d, want 0 (the article top)", got)
	}
}

func TestAttributionWrapsUnderNarrowPhoto(t *testing.T) {
	// The 4-cell fake photo forces the 14-cell fallback attribution to wrap
	// onto four centered lines within the photo's span.
	m, fr := newImageModel(t, []string{"IMG1", "IMG2"})
	m.article = m.newArticleState(imageArticle())

	// The wrapped attribution needs more lines than the initial cap
	// reserved, so the photo is re-rendered smaller.
	if len(fr.calls) != 2 {
		t.Fatalf("expected two render calls, got %d", len(fr.calls))
	}
	if fr.calls[1].maxH >= fr.calls[0].maxH {
		t.Errorf("re-render should shrink the cap: %d then %d", fr.calls[0].maxH, fr.calls[1].maxH)
	}

	contentW := m.article.viewport.Width
	lines := strippedLines(m.article.lines)
	photoPad := (contentW - 4) / 2
	for i, want := range []string{"phot", "o:", "exam", "ple"} {
		attr := lines[m.article.imgStart+2+i]
		if !strings.HasSuffix(strings.TrimRight(attr, " "), want) {
			t.Errorf("wrapped attribution line %d = %q, want it to end with %q", i, attr, want)
		}
		wantPad := photoPad + max(0, (4-len(want))/2)
		if got := len(attr) - len(strings.TrimLeft(attr, " ")); got != wantPad {
			t.Errorf("wrapped line %d left pad = %d, want %d: %q", i, got, wantPad, attr)
		}
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
	width, vpH, _ := contentGeom(m.width, m.height, m.cfg.Display.PaddingX, m.cfg.Display.PaddingY)
	header := m.renderHeader(a, width)
	headerLines := 0
	if header != "" {
		headerLines = len(strings.Split(header, "\n"))
	}
	msg := imgpkg.NativeCmd(m.imgNative, m.imgPhotos, m.imgNatives, a.ImageURL, width, vpH, articleAttribution(a), headerLines)()
	nm, ok := msg.(imgpkg.NativeMsg)
	if !ok {
		t.Fatalf("expected NativeMsg, got %T", msg)
	}
	return nm.Lines
}

func TestNativeKittyImageClearsWhenNotDisplayed(t *testing.T) {
	m, _ := newImageModel(t, nil)
	m.imgNative.Protocol = imgpkg.ProtocolKitty
	m.imgPhotos.Set(imageArticle().ImageURL, pngBytes(t, 40, 40))
	nativeRenderLines(t, m, imageArticle())
	m.article = m.newArticleState(imageArticle())
	m.view = viewArticle
	if !m.article.nativeImg {
		t.Fatal("expected the composed block to be a native render")
	}

	// While the block intersects the visible window the frame re-transmits
	// the image and must not delete the placement.
	if got := m.View(); !strings.Contains(got, "\x1b_Ga=T") {
		t.Error("visible native image should carry the transmit escape")
	} else if strings.Contains(got, kittyDeleteEsc) {
		t.Error("visible native image must not be deleted")
	}

	// Landing on the caption (photo rows above the window, caption visible)
	// must also delete the placement: the transmit line is not rendered, so
	// the photo would float over the caption otherwise.
	m.article.viewport.SetYOffset(m.article.capStart)
	if got := m.View(); !strings.Contains(got, kittyDeleteEsc) {
		t.Error("caption frame must delete the photo's placement")
	}

	// Scrolled past the whole block the placement must be deleted too, or
	// the photo would float over the body.
	m.article.viewport.SetYOffset(m.article.imgEnd + 1)
	if got := m.View(); !strings.Contains(got, kittyDeleteEsc) {
		t.Error("frame must delete the placement when the image is scrolled out")
	}

	// Backing out to the list must delete the placement too, or the photo
	// persists over the list and stacks over the next article's photo.
	m.backToList()
	if got := m.View(); !strings.Contains(got, kittyDeleteEsc) {
		t.Error("list frame must delete the article's placement")
	}

	// A frame without a native render (the placeholder phase of a fresh
	// article) also carries the delete, so it cannot stack over a stale
	// placement left by an earlier view.
	m.imgPhotos = imgpkg.NewPhotos()
	m.imgNatives = imgpkg.NewNatives()
	m.article = m.newArticleState(imageArticle())
	m.view = viewArticle
	if m.article.nativeImg {
		t.Fatal("expected the placeholder state without a native render")
	}
	if got := m.View(); !strings.Contains(got, kittyDeleteEsc) {
		t.Error("article frame without a native render must delete stray placements")
	}
}

func TestNativeKittyImageRestoredAfterClear(t *testing.T) {
	// Scrolling back to the image re-renders its line, re-transmitting the
	// photo after the delete, so the clear never loses the image.
	m, _ := newImageModel(t, nil)
	m.imgNative.Protocol = imgpkg.ProtocolKitty
	m.imgPhotos.Set(imageArticle().ImageURL, pngBytes(t, 40, 40))
	nativeRenderLines(t, m, imageArticle())
	m.article = m.newArticleState(imageArticle())
	m.view = viewArticle

	m.article.viewport.SetYOffset(m.article.imgEnd + 1)
	if got := m.View(); !strings.Contains(got, kittyDeleteEsc) {
		t.Fatal("scrolled-out image should have been deleted")
	}
	m.scrollArticle(func() { m.article.viewport.GotoTop() })
	if got := m.View(); !strings.Contains(got, "\x1b_Ga=T") || strings.Contains(got, kittyDeleteEsc) {
		t.Error("scrolled-back image must be re-transmitted and not deleted")
	}
}

func TestNativeITermImageNeedsNoClear(t *testing.T) {
	// OSC 1337 inline images are cell-bound, so frames never carry the kitty
	// delete sequence on an iTerm2 terminal.
	m, _ := newImageModel(t, nil)
	m.imgNative.Protocol = imgpkg.ProtocolITerm
	m.imgPhotos.Set(imageArticle().ImageURL, pngBytes(t, 40, 40))
	nativeRenderLines(t, m, imageArticle())
	m.article = m.newArticleState(imageArticle())
	m.view = viewArticle
	if got := m.View(); strings.Contains(got, kittyDeleteEsc) {
		t.Error("iTerm frame should not carry the kitty delete sequence")
	}
	m.backToList()
	if got := m.View(); strings.Contains(got, kittyDeleteEsc) {
		t.Error("iTerm list frame should not carry the kitty delete sequence")
	}
}

func TestNativePhotoCenteredInArticle(t *testing.T) {
	// The viewport height cap narrows a 40x40 photo well below the content
	// width, so the composed block must be centered like a halfblock block:
	// the escape line carries a blank left pad with an equal (±1) right
	// margin, and the attribution beneath wraps to the photo's own width.
	m, _ := newImageModel(t, nil)
	m.imgNative.Protocol = imgpkg.ProtocolKitty
	m.imgPhotos.Set(imageArticle().ImageURL, pngBytes(t, 40, 40))
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
	m.imgNative.Protocol = imgpkg.ProtocolITerm
	m.imgCache = imgpkg.NewCache() // only the stored block is available
	contentW, _, _ := contentGeom(m.width, m.height, m.cfg.Display.PaddingX, m.cfg.Display.PaddingY)
	m.imgBlocks.Set(imageArticle().ImageURL, []string{strings.Repeat("I", contentW), strings.Repeat("I", contentW)})
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
	m.imgNative.Protocol = imgpkg.ProtocolITerm
	m.imgPhotos.Set(imageArticle().ImageURL, pngBytes(t, 40, 40))
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
	m.imgNative.Protocol = imgpkg.ProtocolITerm
	m.imgCache = imgpkg.NewCache()
	m.imgBlocks = imgpkg.NewBlocks()
	m.article = m.newArticleState(imageArticle())
	m.view = viewArticle
	if m.article.imgStart != -1 {
		t.Fatalf("no image should be composed before the photo render, imgStart=%d", m.article.imgStart)
	}

	// PhotoMsg lands: the photo bytes are cached and onPhotoLoaded fires the
	// off-thread native render command (no on-thread recompose).
	m.imgPhotos.Set(imageArticle().ImageURL, pngBytes(t, 40, 40))
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
	if _, ok := m.imgNatives.Get(key); !ok {
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
	m.imgCache = imgpkg.NewCache()
	width, _, _ := contentGeom(m.width, m.height, m.cfg.Display.PaddingX, m.cfg.Display.PaddingY)
	lines := []string{strings.Repeat("I", width), strings.Repeat("I", width)}
	m.imgBlocks.Set(imageArticle().ImageURL, lines)
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
