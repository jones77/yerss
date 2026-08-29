package ui

import (
	"image"
	"image/color"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

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
	// after one blank line and covers the two image lines plus the
	// attribution line.
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
	if !strings.Contains(lines[7], "phot") {
		t.Errorf("attribution line missing below the image: %q", lines[7])
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
	m, fr := newImageModel(t, []string{"IMG1", "IMG2"})
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
	// The cap reserves the blank lines around the block, the attribution
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
	m.onImageLoaded(imgpkg.LoadedMsg{Key: imageArticle().ImageURL, Img: testImg()})

	if m.article.imgStart <= 0 {
		t.Fatalf("image block not composed after load: imgStart=%d", m.article.imgStart)
	}
	// The inserted lines are the blank above the block, the two image lines,
	// the attribution, and the blank below it.
	if want := oldOffset + 4; m.article.viewport.YOffset != want {
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
	m.onImageLoaded(imgpkg.LoadedMsg{Key: imageArticle().ImageURL, Img: testImg()})
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
	if _, ok := msg.(imgpkg.LoadedMsg); !ok {
		t.Fatalf("expected LoadedMsg, got %T", msg)
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
	lm, ok := msg.(imgpkg.LoadedMsg)
	if !ok {
		t.Fatalf("expected LoadedMsg, got %T", msg)
	}
	if lm.Key != imageArticle().ImageURL {
		t.Errorf("LoadedMsg.Key = %q, want %q", lm.Key, imageArticle().ImageURL)
	}
	wantOffset := m2.article.viewport.YOffset
	m2.onImageLoaded(lm)
	if m2.article.imgStart <= 0 {
		t.Errorf("image block missing after load: imgStart=%d", m2.article.imgStart)
	}
	if m2.article.viewport.YOffset <= wantOffset {
		t.Errorf("offset = %d, want it bumped past the inserted image lines (was %d)", m2.article.viewport.YOffset, wantOffset)
	}
}

func TestSnapYOffset(t *testing.T) {
	cases := []struct {
		name                       string
		offset, vpH, start, end, d int
		want                       int
	}{
		// offset is the position AFTER a scroll move; the image occupies
		// lines [start, end] and must never be partially clipped.
		{"down onto first line", 4, 20, 4, 9, 1, 10},
		{"down into interior", 5, 20, 4, 9, 1, 10},
		{"down at last line", 9, 20, 4, 9, 1, 10},
		{"up from last line", 9, 20, 4, 9, -1, 4},
		{"up into interior", 8, 20, 4, 9, -1, 4},
		{"up onto first line", 4, 20, 4, 9, -1, 4},
		{"below block unchanged", 10, 20, 4, 9, 1, 10},
		{"above block unchanged", 2, 20, 4, 9, 1, 2},
		{"at first line up reveals", 4, 20, 4, 9, -1, 4},
		{"no image block", 5, 20, -1, -1, 1, 5},
		{"page down into interior", 6, 20, 4, 9, 20, 10},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := snapYOffset(c.offset, c.vpH, c.start, c.end, c.d); got != c.want {
				t.Errorf("snapYOffset(%d,%d,%d,%d,%d) = %d, want %d", c.offset, c.vpH, c.start, c.end, c.d, got, c.want)
			}
		})
	}
}

func TestScrollSnapsPastImageBlock(t *testing.T) {
	m, _ := newImageModel(t, []string{"IMG1", "IMG2"})
	m.article = m.newArticleState(imageArticle())
	// The block sits below the header (lines imgStart..imgEnd); a downward
	// move from just above it enters the interior and snaps past the block.
	m.article.viewport.SetYOffset(m.article.imgStart - 1)
	m.scrollArticle(func() { m.article.viewport.ScrollDown(1) })
	if got := m.article.viewport.YOffset; got != m.article.imgEnd+1 {
		t.Errorf("offset after down scroll = %d, want %d", got, m.article.imgEnd+1)
	}

	// Up scroll from past the block reveals the full image.
	m.scrollArticle(func() { m.article.viewport.ScrollUp(1) })
	if got := m.article.viewport.YOffset; got != m.article.imgStart {
		t.Errorf("offset after up scroll = %d, want %d", got, m.article.imgStart)
	}
}

func TestArticleImageCentered(t *testing.T) {
	m, _ := newImageModel(t, []string{"IMG1", "IMG2"})
	m.article = m.newArticleState(imageArticle())

	contentW := m.article.viewport.Width
	lines := strippedLines(m.article.lines)
	img1 := lines[m.article.imgStart]
	attr := lines[m.article.imgEnd]
	// The fake photo is 4 cells wide: it is centered by (contentW-4)/2
	// leading spaces, and the attribution is truncated to the photo's width
	// ("photo: example" -> "phot") and centered within the same span.
	if got := len(img1) - len(strings.TrimLeft(img1, " ")); got != (contentW-4)/2 {
		t.Errorf("image line left pad = %d, want %d: %q", got, (contentW-4)/2, img1)
	}
	if got := len(attr) - len(strings.TrimLeft(attr, " ")); got != (contentW-4)/2 {
		t.Errorf("attribution left pad = %d, want %d: %q", got, (contentW-4)/2, attr)
	}
	if !strings.HasSuffix(attr, "phot") {
		t.Errorf("attribution should be truncated to the photo width: %q", attr)
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

var _ = tea.Cmd(nil)