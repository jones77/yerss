package ui

import (
	"image"
	"image/color"
	"strings"
	"testing"

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

func TestArticleImageBlockComposedAboveContent(t *testing.T) {
	m, fr := newImageModel(t, []string{"IMG1", "IMG2"})
	m.article = m.newArticleState(imageArticle())

	if len(fr.calls) == 0 {
		t.Fatal("renderer was not called")
	}
	if m.article.imgStart != 0 || m.article.imgEnd != 1 {
		t.Errorf("img range = %d..%d, want 0..1", m.article.imgStart, m.article.imgEnd)
	}
	if len(m.article.lines) < 3 {
		t.Fatalf("too few lines: %d", len(m.article.lines))
	}
	if ansi.Strip(m.article.lines[0]) != "IMG1" || ansi.Strip(m.article.lines[1]) != "IMG2" {
		t.Errorf("image block lines not first: %q", m.article.lines[:2])
	}
	joined := strings.Join(m.article.lines, "\n")
	if !strings.Contains(ansi.Strip(joined), "Headline") || !strings.Contains(ansi.Strip(joined), "body text") {
		t.Errorf("header/body missing after image block")
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
	header := m.renderMarkdown(articleHeaderMarkdown(imageArticle()), m.article.viewport.Width)
	headerLines := 0
	if header != "" {
		headerLines = len(strings.Split(header, "\n"))
	}
	if want := m.article.viewport.Height - headerLines - 1; fr.calls[0].maxH != want {
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

	if m.article.imgStart != 0 {
		t.Fatalf("image block not composed after load: imgStart=%d", m.article.imgStart)
	}
	if want := oldOffset + 2; m.article.viewport.YOffset != want {
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
	if m.article.imgStart != 0 {
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
	// The image block is lines 0..1; from the top, one line down enters the
	// interior and snaps past the block.
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

var _ = tea.Cmd(nil)