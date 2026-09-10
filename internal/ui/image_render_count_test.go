package ui

import (
	"fmt"
	"image"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	imgpkg "yerss/internal/image"
	"yerss/internal/store"
)

// countingRenderer wraps the real halfblock renderer and counts calls, so a
// test can report how many synchronous mosaic renders the recompose path
// performed on the UI goroutine (the work eliminated by serving cached blocks).
type countingRenderer struct {
	inner imgpkg.Renderer
	calls int
}

func (c *countingRenderer) Render(img image.Image, w, h int) ([]string, error) {
	c.calls++
	return c.inner.Render(img, w, h)
}

// drain runs a tea.Cmd to completion, feeding returned messages back through
// Update and queueing any returned commands, until the queue is empty. It
// mirrors bubbletea's event loop serially.
func drain(m *Model, cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	pending := []tea.Cmd{cmd}
	for len(pending) > 0 {
		c := pending[0]
		pending = pending[1:]
		msg := c()
		switch msg := msg.(type) {
		case tea.BatchMsg:
			pending = append(pending, msg...)
		case nil:
		default:
			_, next := m.Update(msg)
			if next != nil {
				pending = append(pending, next)
			}
		}
	}
}

// photoHeavyArticle returns an article with a lead image and n inline images
// separated by short paragraphs, compact enough that every image's anchor row
// sits within the open viewport's load frontier.
func photoHeavyArticle(n int) store.Article {
	var content strings.Builder
	content.WriteString("<p>intro</p>")
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&content, `<p><img src="https://example.com/img%d.jpg" alt="Photo %d"></p>`, i, i)
		content.WriteString("<p>body</p>")
	}
	return store.Article{
		FeedURL:  "https://example.com/feed.xml",
		GUID:     "g-photo-heavy",
		Title:    "Photo Heavy",
		Author:   "Jane",
		Link:     "https://example.com/a",
		ImageURL: "https://example.com/lead.jpg",
		Content:  content.String(),
	}
}

func TestRecomposeServesCachedBlocksRendersBounded(t *testing.T) {
	// The full open -> fire-image-loads -> message-loop -> recompose pipeline
	// on a synthetic multi-image article must render each block at most a
	// constant number of times per image. Before the cache-serve fix, every
	// image-load recompose re-rendered each already-decoded image on the UI
	// goroutine, giving an O(N^2) render count (measured 349 renders across 33
	// recomposes for the 11-photo Helene article); serving the width-matched
	// cached block keeps the count O(N).
	const images = 8 // lead + 7 inline
	m, st := newTestModel(t)
	m.width, m.height = 120, 44
	cr := &countingRenderer{inner: imgpkg.Halfblocks{}}
	m.sess.ImgRenderer = cr
	m.sess.ImgNative = imgpkg.NativeRenderer{Protocol: imgpkg.ProtocolNone}

	a := photoHeavyArticle(images - 1)
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	for pos := 0; pos < images; pos++ {
		url := a.ImageURL
		if pos > 0 {
			url = fmt.Sprintf("https://example.com/img%d.jpg", pos)
		}
		if err := st.SetArticleImage(id, store.ArticleImage{
			Position: pos,
			URL:      url,
			Photo:    pngBytes(t, 320, 160),
		}); err != nil {
			t.Fatal(err)
		}
	}

	full, err := st.GetArticle(id)
	if err != nil || full == nil {
		t.Fatalf("GetArticle(%d): %v", id, err)
	}
	m.article = m.newArticleState(*full)
	m.view = viewArticle
	drain(m, m.fireImageLoad(*full))

	if len(m.article.imageBlocks) != images {
		t.Fatalf("imageBlocks = %d, want %d (every image should have composed)", len(m.article.imageBlocks), images)
	}
	if got := len(m.article.imageURLs); got != images {
		t.Fatalf("imageURLs = %d, want %d", got, images)
	}
	if m.recomposeCount <= 0 {
		t.Fatal("the message loop should have recomposed the article")
	}

	// O(N) render bound: a recompose never re-renders a block already cached
	// at the current width, so the UI-thread renderer is called at most a
	// constant times the image count, not once per recompose per image.
	if cr.calls > 3*images {
		t.Errorf("UI-thread halfblock renders = %d, want <= %d (3N); recomposes = %d",
			cr.calls, 3*images, m.recomposeCount)
	}
	// Recompose count is bounded by the messages that drove it (one block
	// message per image), so it too stays O(N).
	if m.recomposeCount > 4*images {
		t.Errorf("recomposeCount = %d, want <= %d (4N)", m.recomposeCount, 4*images)
	}
}