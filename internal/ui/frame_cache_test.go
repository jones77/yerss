package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"yerss/internal"

	imgpkg "yerss/internal/image"
	"yerss/internal/ui/articleview"
)

// TestFrameCacheSkipsNoOpBoundaryScrolls asserts the article frame cache: a
// mouse-wheel message at the top or bottom of the article viewport cannot move
// the offset, so the frame must be served from the cache with no re-render; a
// scroll that actually moves the offset re-renders exactly once. The messages
// are driven through Model.Update (so the deferred recomputeNativeClear runs,
// matching the real message loop) followed by View() for the composed frame.
func TestFrameCacheSkipsNoOpBoundaryScrolls(t *testing.T) {
	m, _ := newTestModel(t)
	m.view = viewArticle
	m.article = m.article.Recompose(internal.Article{Title: "long", Content: "<p>x</p>"})
	m.article.Viewport.SetContent(strings.Join(make([]string, 70), "\n"))
	m.article.Viewport.GotoTop()

	frameView(m) // fill the frame cache for the top offset
	m.article.SetFrameRenderCount(0)

	// Wheel up at the very top cannot scroll further; the first such message
	// arms the transient "K: next" hint, so it re-renders exactly once.
	m.Update(mouseWheel(0, false))
	m.View()
	if m.article.Viewport.YOffset != 0 {
		t.Fatalf("wheel-up at top moved offset to %d, want 0", m.article.Viewport.YOffset)
	}
	if m.article.FrameRenderCount() != 1 {
		t.Fatalf("wheel-up at top rendered the frame %d times, want 1 (hint)", m.article.FrameRenderCount())
	}
	if !strings.Contains(m.View(), "K: next") {
		t.Error("top-edge hint should show K: next")
	}

	// A repeated identical no-op is cache-served: the hint is already armed.
	m.article.SetFrameRenderCount(0)
	m.Update(mouseWheel(0, false))
	m.View()
	if m.article.FrameRenderCount() != 0 {
		t.Fatalf("repeat wheel-up at top rendered the frame %d times, want 0", m.article.FrameRenderCount())
	}

	// A real one-line scroll down changes the offset and re-renders exactly once.
	m.Update(mouseWheel(0, true))
	m.View()
	if m.article.Viewport.YOffset != 1 {
		t.Fatalf("wheel-down offset = %d, want 1", m.article.Viewport.YOffset)
	}
	if m.article.FrameRenderCount() != 1 {
		t.Fatalf("real wheel-down rendered the frame %d times, want exactly 1", m.article.FrameRenderCount())
	}
	m.article.SetFrameRenderCount(0)

	// Wheel down at the very bottom cannot scroll further; the first such
	// message arms the "J: previous" hint and re-renders once.
	m.article.Viewport.GotoBottom()
	frameView(m) // fill the frame cache for the bottom offset
	m.article.SetFrameRenderCount(0)
	m.Update(mouseWheel(0, true))
	m.View()
	if m.article.FrameRenderCount() != 1 {
		t.Fatalf("wheel-down at bottom rendered the frame %d times, want 1 (hint)", m.article.FrameRenderCount())
	}
	if !strings.Contains(m.View(), "J: previous") {
		t.Error("bottom-edge hint should show J: previous")
	}
	m.article.SetFrameRenderCount(0)
	m.Update(mouseWheel(0, true))
	m.View()
	if m.article.FrameRenderCount() != 0 {
		t.Fatalf("repeat wheel-down at bottom rendered the frame %d times, want 0", m.article.FrameRenderCount())
	}
}

// TestFrameCacheByteIdenticalWithNativeSelectionLinkMode asserts the article
// frame cache preserves byte-identical output across cache hits when every
// render input is unchanged: a native kitty image (with the transmit state
// settled before the first frame so each frame is a stable placement
// reference), a text selection over a content line, and link-expand mode. It
// then asserts each render-input change in the articleFrameKey — selection,
// link mode, ascii mode — invalidates the cache and re-renders.
func TestFrameCacheByteIdenticalWithNativeSelectionLinkMode(t *testing.T) {
	a := imageArticle()
	a.Content = `<p>long paragraph body text</p><p>see <a href="https://example.com/xx">the link</a> here</p>` +
		strings.Repeat("<p>more body text</p>", 60)

	m, _ := newImageModel(t, nil)
	m.sess.ImgNative.Protocol = imgpkg.ProtocolKitty
	m.sess.ImgPhotos.Set(a.ImageURL, pngBytes(t, 40, 40))
	m.SetLinkExpand(true)
	nativeRenderLines(t, m, a)
	m.article = m.article.Recompose(a)
	m.view = viewArticle
	if !m.article.NativeImg() {
		t.Fatal("expected a native render")
	}
	leadID := m.article.ImageBlocks()[0].NativeID
	if leadID == 0 {
		t.Fatal("native block should record its kitty render id")
	}

	// Settle the native transmit state before the first frame: the terminal
	// already holds the image under its render id, so every frame renders a
	// placement reference and no snapshot change bumps nativeGen mid-sequence.
	m.images.SetNativeSent(a.ImageURL, leadID)

	// A text selection over the first content line, so the highlight path runs.
	m.article.Sel = articleview.Selection{Active: true, Tracking: false, AnchorX: 5, AnchorY: 0, CurX: 20, CurY: 0}

	f1 := frameView(m) // cache fill
	if !strings.Contains(f1, "\x1b_Ga=p,i="+fmt.Sprint(leadID)) {
		t.Fatalf("settled frame should reference the held native image: %q", f1)
	}
	if strings.Contains(f1, "\x1b_Ga=T") {
		t.Fatal("settled frame must not carry the full transmit")
	}
	m.article.SetFrameRenderCount(0)
	f2 := frameView(m) // must be served from the cache
	if m.article.FrameRenderCount() != 0 {
		t.Errorf("stable frame re-rendered %d times, want 0", m.article.FrameRenderCount())
	}
	if f2 != f1 {
		t.Error("cached frame output differs from the first render")
	}

	// A forced cache miss (frameValid dropped) re-renders exactly once and
	// produces byte-identical output.
	m.article.SetFrameValid(false)
	f3 := frameView(m)
	if m.article.FrameRenderCount() != 1 {
		t.Errorf("forced cache miss rendered %d times, want exactly 1", m.article.FrameRenderCount())
	}
	if f3 != f1 {
		t.Error("fresh uncached render differs from the cached output")
	}

	// A selection change is a render input: it invalidates the cache.
	m.article.Sel = articleview.Selection{}
	m.article.SetFrameRenderCount(0)
	frameView(m)
	if m.article.FrameRenderCount() != 1 {
		t.Errorf("selection change rendered %d times, want exactly 1", m.article.FrameRenderCount())
	}

	// The link-mode toggle recomposes the article (contentGen bump) and must
	// re-render; expanded mode shows the href as visible text, compact mode
	// keeps only the clickable name (the href URL survives solely inside the
	// name span's OSC 8 target, which ansi.Strip removes).
	m.article.SetFrameRenderCount(0)
	m.SetLinkExpand(false)
	m.recomposeArticle()
	compact := frameView(m)
	if m.article.FrameRenderCount() != 1 {
		t.Errorf("link-mode toggle rendered %d times, want exactly 1", m.article.FrameRenderCount())
	}
	if !strings.Contains(ansi.Strip(f1), "https://example.com/xx") {
		t.Error("link-expanded frame should show the href as visible text")
	}
	if strings.Contains(ansi.Strip(compact), "https://example.com/xx") {
		t.Error("compact link frame should drop the visible href")
	}

	// Toggling ascii mode changes the frame glyphs and must re-render; the
	// ascii frame is then stable.
	m.SetAscii(true)
	m.article.SetFrameRenderCount(0)
	ascii := frameView(m)
	if m.article.FrameRenderCount() != 1 {
		t.Errorf("ascii-mode toggle rendered %d times, want exactly 1", m.article.FrameRenderCount())
	}
	if ascii == f1 {
		t.Error("ascii mode should change the frame (border glyphs)")
	}
	m.article.SetFrameRenderCount(0)
	if got := frameView(m); got != ascii || m.article.FrameRenderCount() != 0 {
		t.Error("stable ascii frame should be served from the cache byte-identically")
	}
}

// TestFrameCacheRecomposesOnImageLoadAtBoundary asserts an image-load recompose
// invalidates the frame cache even at a scroll boundary: while the lead image
// is unloaded the article renders without a block and a no-op wheel at the top
// is cache-served; when the block load message lands, the recompose (contentGen
// bump) forces the next frame to re-render — exactly once — with the new block
// replacing the placeholder, and the recomposed frame is then cache-stable
// again.
func TestFrameCacheRecomposesOnImageLoadAtBoundary(t *testing.T) {
	m, _ := newImageModel(t, nil)
	m.sess.ImgCache = imgpkg.NewCache() // no image data yet
	m.article = m.article.Recompose(imageArticle())
	m.view = viewArticle
	if m.article.ImgStart() != -1 {
		t.Fatalf("unloaded article should have no block, imgStart=%d", m.article.ImgStart())
	}
	if m.article.Viewport.YOffset != 0 {
		t.Fatalf("expected the article at the top boundary, offset=%d", m.article.Viewport.YOffset)
	}

	// A no-op wheel at the top arms the transient hint and re-renders once;
	// subsequent frames are cache-stable until the image load lands.
	frameView(m)
	m.article.SetFrameRenderCount(0)
	m.Update(mouseWheel(0, false))
	frameView(m)
	if m.article.FrameRenderCount() != 1 {
		t.Fatalf("no-op wheel at a boundary rendered %d times, want 1 (hint)", m.article.FrameRenderCount())
	}

	// Delivering the block message recomposes the article (contentGen bump), so
	// the next frame must re-render and show the loaded block, not the
	// placeholder-free composition. The block lines are width-matched so compose
	// serves them from the block cache.
	contentW := m.article.Viewport.Width
	blockLines := []string{strings.Repeat("I", contentW), strings.Repeat("I", contentW)}
	m.article.SetFrameRenderCount(0)
	m.Update(imgpkg.BlockMsg{Key: imageArticle().ImageURL, Lines: blockLines})
	got := frameView(m)
	if m.article.FrameRenderCount() != 1 {
		t.Errorf("image-load recompose rendered %d times, want exactly 1", m.article.FrameRenderCount())
	}
	if m.article.ImgStart() < 0 {
		t.Fatalf("lead image block not composed after load: imgStart=%d", m.article.ImgStart())
	}
	if !strings.Contains(got, strings.Repeat("I", contentW)) {
		t.Errorf("frame after the load must show the new block: %q", got)
	}

	// The recomposed frame is cache-stable: another boundary no-op stays cached.
	m.article.SetFrameRenderCount(0)
	m.Update(mouseWheel(0, false))
	frameView(m)
	if m.article.FrameRenderCount() != 0 {
		t.Errorf("no-op wheel after recompose re-rendered %d times, want 0", m.article.FrameRenderCount())
	}
}
