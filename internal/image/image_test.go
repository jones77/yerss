package image

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
)

// pngBytes encodes a solid-color image as PNG bytes.
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

func newImageTestStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(t.TempDir() + "/img.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func TestCaches(t *testing.T) {
	c := NewCache()
	if _, ok := c.Get("u"); ok {
		t.Error("empty cache should miss")
	}
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	c.Set("u", img)
	if got, ok := c.Get("u"); !ok || got != img {
		t.Errorf("decoded cache roundtrip failed: ok=%v", ok)
	}

	b := NewBlocks()
	if _, ok := b.Get("u"); ok {
		t.Error("empty block cache should miss")
	}
	b.Set("u", []string{"a", "b"})
	if got, ok := b.Get("u"); !ok || len(got) != 2 {
		t.Errorf("block cache roundtrip failed: ok=%v", ok)
	}

	p := NewPhotos()
	if _, ok := p.Get("u"); ok {
		t.Error("empty photo cache should miss")
	}
	p.Set("u", []byte("data"))
	if got, ok := p.Get("u"); !ok || string(got) != "data" {
		t.Errorf("photo cache roundtrip failed: ok=%v", ok)
	}

	n := NewNatives()
	if _, ok := n.Get(NativeKey("u", 50, 24)); ok {
		t.Error("empty native cache should miss")
	}
	lines := []string{"line1", "line2"}
	n.Set(NativeKey("u", 50, 24), lines)
	if got, ok := n.Get(NativeKey("u", 50, 24)); !ok || got[0] != "line1" {
		t.Errorf("native cache roundtrip failed: ok=%v", ok)
	}
	// A different render size is a different key.
	if _, ok := n.Get(NativeKey("u", 50, 25)); ok {
		t.Error("native cache key should include the height cap")
	}
	if _, ok := n.Get(NativeKey("u", 51, 24)); ok {
		t.Error("native cache key should include the width")
	}
}

func TestNativeKey(t *testing.T) {
	a := NativeKey("https://e/x.png", 50, 24)
	b := NativeKey("https://e/x.png", 50, 24)
	c := NativeKey("https://e/x.png", 51, 24)
	d := NativeKey("https://e/y.png", 50, 24)
	if a != b {
		t.Errorf("identical keys differ: %q vs %q", a, b)
	}
	if a == c {
		t.Error("different width must produce different keys")
	}
	if a == d {
		t.Error("different url must produce different keys")
	}
}

func TestBlockWidth(t *testing.T) {
	lines := []string{"\x1b[38;2;200;100;50m██\x1b[0m", "\x1b[0m  \x1b[0m"}
	if w := BlockWidth(lines); w != 2 {
		t.Errorf("BlockWidth = %d, want 2", w)
	}
	if w := BlockWidth(nil); w != 0 {
		t.Errorf("BlockWidth(nil) = %d, want 0", w)
	}
}

func TestDecodeCapped(t *testing.T) {
	t.Run("oversized source downscaled to cap", func(t *testing.T) {
		// A 4096x2731 hero photo decodes to a bitmap capped at 2048 wide.
		data := pngBytes(t, 4096, 2731)
		img, err := decodeCapped(data)
		if err != nil {
			t.Fatal(err)
		}
		b := img.Bounds()
		if b.Dx() != maxDecodeWidth {
			t.Errorf("decoded width = %d, want %d", b.Dx(), maxDecodeWidth)
		}
		if wantH := (2731*maxDecodeWidth + 4096/2) / 4096; b.Dy() != wantH {
			t.Errorf("decoded height = %d, want %d (aspect preserved)", b.Dy(), wantH)
		}
	})
	t.Run("at-or-below cap returned unchanged", func(t *testing.T) {
		data := pngBytes(t, 1024, 640)
		img, err := decodeCapped(data)
		if err != nil {
			t.Fatal(err)
		}
		b := img.Bounds()
		if b.Dx() != 1024 || b.Dy() != 640 {
			t.Errorf("natural decode = %dx%d, want 1024x640", b.Dx(), b.Dy())
		}
		exact := pngBytes(t, maxDecodeWidth, 400)
		img, err = decodeCapped(exact)
		if err != nil {
			t.Fatal(err)
		}
		if b := img.Bounds(); b.Dx() != maxDecodeWidth || b.Dy() != 400 {
			t.Errorf("exact-cap decode = %dx%d, want %dx400", b.Dx(), b.Dy(), maxDecodeWidth)
		}
	})
	t.Run("unparseable bytes error", func(t *testing.T) {
		if _, err := decodeCapped([]byte("not an image")); err == nil {
			t.Fatal("unparseable bytes should error")
		}
	})
}

func TestBlockCmdInMemoryHit(t *testing.T) {
	st := newImageTestStore(t)
	img := image.NewRGBA(image.Rect(0, 0, 480, 320))
	c := NewCache()
	c.Set("https://example.com/img.png", img)
	blocks := NewBlocks()
	msg := BlockCmd(c, blocks, st, 1, "https://example.com/img.png", 50, 100, true, 0)()
	bm, ok := msg.(BlockMsg)
	if !ok {
		t.Fatalf("expected BlockMsg, got %T", msg)
	}
	if bm.Key != "https://example.com/img.png" {
		t.Errorf("key = %q", bm.Key)
	}
	if BlockWidth(bm.Lines) != 50 {
		t.Errorf("block width = %d, want 50", BlockWidth(bm.Lines))
	}
	// The in-memory path must not persist anything.
	if imgs, _ := st.GetArticleImages(1); len(imgs) != 0 {
		t.Errorf("in-memory hit should not persist, got %+v", imgs)
	}
}

func TestBlockCmdStoredBlockHit(t *testing.T) {
	st := newImageTestStore(t)
	id, err := st.UpsertArticle(store.Article{FeedURL: "f", GUID: "g"})
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 480, 320))
	want, _ := (Halfblocks{}).Render(img, 40, 100)
	if err := st.SetArticleImage(id, store.ArticleImage{Position: 0, URL: "https://example.com/img.png", Block: strings.Join(want, "\n"), Width: BlockWidth(want)}); err != nil {
		t.Fatal(err)
	}
	c := NewCache()
	blocks := NewBlocks()
	msg := BlockCmd(c, blocks, st, id, "https://example.com/img.png", 40, 100, true, 0)()
	bm, ok := msg.(BlockMsg)
	if !ok {
		t.Fatalf("expected BlockMsg, got %T", msg)
	}
	if BlockWidth(bm.Lines) != 40 {
		t.Errorf("stored block width = %d, want 40", BlockWidth(bm.Lines))
	}
	// No network, no decoded cache.
	if _, ok := c.Get("https://example.com/img.png"); ok {
		t.Error("stored-block hit should not decode an image")
	}
}

func TestBlockCmdStoredPhotoRerendersOnWidthMismatch(t *testing.T) {
	st := newImageTestStore(t)
	id, err := st.UpsertArticle(store.Article{FeedURL: "f", GUID: "g"})
	if err != nil {
		t.Fatal(err)
	}
	data := pngBytes(t, 480, 320)
	img := image.NewRGBA(image.Rect(0, 0, 480, 320))
	old, _ := (Halfblocks{}).Render(img, 40, 100)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("width mismatch must re-render from the stored photo, not fetch")
	}))
	defer srv.Close()
	if err := st.SetArticleImage(id, store.ArticleImage{
		Position: 0,
		URL:      srv.URL,
		Block:    strings.Join(old, "\n"),
		Photo:    data,
		Width:    BlockWidth(old),
	}); err != nil {
		t.Fatal(err)
	}

	msg := BlockCmd(NewCache(), NewBlocks(), st, id, srv.URL, 60, 100, true, 0)()
	bm, ok := msg.(BlockMsg)
	if !ok {
		t.Fatalf("expected BlockMsg, got %T", msg)
	}
	if BlockWidth(bm.Lines) != 60 {
		t.Errorf("re-rendered block width = %d, want 60", BlockWidth(bm.Lines))
	}
	imgs, err := st.GetArticleImages(id)
	if err != nil {
		t.Fatal(err)
	}
	if len(imgs) != 1 || BlockWidth(splitLines(imgs[0].Block)) != 60 {
		t.Errorf("persisted block width = %d, want 60", BlockWidth(splitLines(imgs[0].Block)))
	}
	if string(imgs[0].Photo) != string(data) {
		t.Error("re-render must preserve the stored photo bytes")
	}
}

func TestBlockCmdNoFetchModeNeverHitsNetwork(t *testing.T) {
	st := newImageTestStore(t)
	id, _ := st.UpsertArticle(store.Article{FeedURL: "f", GUID: "g"})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no-fetch mode must not hit the network")
	}))
	defer srv.Close()

	msg := BlockCmd(NewCache(), NewBlocks(), st, id, srv.URL, 50, 100, false, 0)()
	if _, ok := msg.(FailedMsg); !ok {
		t.Fatalf("no-fetch miss should fail, got %T", msg)
	}
}

func TestBlockCmdNetworkFetchPersistsBlockAndPhoto(t *testing.T) {
	st := newImageTestStore(t)
	id, err := st.UpsertArticle(store.Article{FeedURL: "f", GUID: "g"})
	if err != nil {
		t.Fatal(err)
	}
	data := pngBytes(t, 4, 4)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	c := NewCache()
	blocks := NewBlocks()
	msg := BlockCmd(c, blocks, st, id, srv.URL, 50, 100, true, 0)()
	if _, ok := msg.(FailedMsg); ok {
		t.Fatal("network fetch should succeed")
	}
	if _, ok := msg.(BlockMsg); !ok {
		t.Fatalf("expected BlockMsg, got %T", msg)
	}
	if _, ok := c.Get(srv.URL); !ok {
		t.Error("network fetch should cache the decoded image")
	}
	imgs, err := st.GetArticleImages(id)
	if err != nil {
		t.Fatal(err)
	}
	if len(imgs) != 1 {
		t.Fatalf("GetArticleImages = %+v, want 1 image", imgs)
	}
	if imgs[0].Block == "" {
		t.Error("network fetch should persist the block")
	}
	if imgs[0].Width == 0 {
		t.Error("network fetch should persist the block width")
	}
	if string(imgs[0].Photo) != string(data) {
		t.Error("network fetch must persist the full raw photo bytes")
	}
}

func TestPhotoCmdFetchesAndPersistsBlock(t *testing.T) {
	st := newImageTestStore(t)
	id, err := st.UpsertArticle(store.Article{FeedURL: "f", GUID: "g"})
	if err != nil {
		t.Fatal(err)
	}
	data := pngBytes(t, 4, 4)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	c := NewCache()
	blocks := NewBlocks()
	photos := NewPhotos()
	msg := PhotoCmd(c, blocks, photos, st, id, srv.URL, 50, 100, 0)()
	if _, ok := msg.(PhotoMsg); !ok {
		t.Fatalf("expected PhotoMsg, got %T", msg)
	}
	got, ok := photos.Get(srv.URL)
	if !ok || string(got) != string(data) {
		t.Error("photo cmd should cache the raw bytes")
	}
	if _, ok := c.Get(srv.URL); !ok {
		t.Error("photo cmd should cache the decoded image")
	}
	imgs, err := st.GetArticleImages(id)
	if err != nil {
		t.Fatal(err)
	}
	if len(imgs) != 1 || imgs[0].Block == "" {
		t.Error("photo cmd should persist the placeholder block")
	}
	photo, err := st.GetArticleImagePhoto(id, srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(photo) != string(data) {
		t.Error("photo cmd should persist the photo bytes")
	}
}

func TestPhotoCmdStoredPhotoSkipsFetch(t *testing.T) {
	st := newImageTestStore(t)
	id, err := st.UpsertArticle(store.Article{FeedURL: "f", GUID: "g"})
	if err != nil {
		t.Fatal(err)
	}
	url := "https://example.com/img.png"
	data := pngBytes(t, 4, 4)
	if err := st.SetArticleImage(id, store.ArticleImage{Position: 0, URL: url, Photo: data}); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("stored photo must not be re-fetched")
	}))
	defer srv.Close()

	msg := PhotoCmd(NewCache(), NewBlocks(), NewPhotos(), st, id, url, 50, 100, 0)()
	if _, ok := msg.(PhotoMsg); !ok {
		t.Fatalf("expected PhotoMsg, got %T", msg)
	}
}

func TestPhotoCmdAlreadyCachedSkipsFetch(t *testing.T) {
	st := newImageTestStore(t)
	id, _ := st.UpsertArticle(store.Article{FeedURL: "f", GUID: "g"})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("cached photo must not re-fetch")
	}))
	defer srv.Close()

	photos := NewPhotos()
	photos.Set(srv.URL, []byte("data"))
	msg := PhotoCmd(NewCache(), NewBlocks(), photos, st, id, srv.URL, 50, 100, 0)()

	if _, ok := msg.(PhotoMsg); !ok {
		t.Fatalf("expected PhotoMsg, got %T", msg)
	}
}

func TestNativeCmdCacheHitShortCircuits(t *testing.T) {
	withCellPx(t, 8, 16)
	natives := NewNatives()
	url := "https://example.com/img.png"
	key := NativeKey(url, 50, 24)
	want := []string{"cached-line-1", "cached-line-2"}
	natives.Set(key, want)

	// The photos cache is empty: any render attempt would fail, so a
	// NativeMsg proves the command short-circuited on the cache hit.
	msg := NativeCmd(NativeRenderer{Protocol: ProtocolITerm}, NewPhotos(), natives, url, 50, 24, "", 3, 35)()
	nm, ok := msg.(NativeMsg)
	if !ok {
		t.Fatalf("expected NativeMsg on cache hit, got %T", msg)
	}
	if nm.Key != url {
		t.Errorf("key = %q, want %q", nm.Key, url)
	}
	if len(nm.Lines) != len(want) || nm.Lines[0] != want[0] {
		t.Errorf("NativeMsg lines = %v, want %v", nm.Lines, want)
	}
	if got, ok := natives.Get(key); !ok || got[1] != want[1] {
		t.Error("cache should be untouched on a hit")
	}
}

func TestNativeCmdRendersAndCaches(t *testing.T) {
	withCellPx(t, 8, 16)
	natives := NewNatives()
	photos := NewPhotos()
	url := "https://example.com/img.png"
	photos.Set(url, pngBytes(t, 400, 320))
	key := NativeKey(url, 50, 24)

	msg := NativeCmd(NativeRenderer{Protocol: ProtocolITerm}, photos, natives, url, 50, 24, "", 3, 35)()
	nm, ok := msg.(NativeMsg)
	if !ok {
		t.Fatalf("expected NativeMsg, got %T", msg)
	}
	if nm.Key != url {
		t.Errorf("key = %q, want %q", nm.Key, url)
	}
	if len(nm.Lines) < 1 || !strings.Contains(nm.Lines[0], "\x1b]1337;File=") {
		t.Fatalf("first line missing OSC 1337 escape: %q", nm.Lines[0][:min(60, len(nm.Lines[0]))])
	}
	if got, ok := natives.Get(key); !ok {
		t.Fatal("render should be cached under the render-size key")
	} else if len(got) != len(nm.Lines) {
		t.Errorf("cached lines = %d, NativeMsg lines = %d", len(got), len(nm.Lines))
	}
}

func TestNativeCmdFitLoopConverges(t *testing.T) {
	withCellPx(t, 8, 16)
	// geometry: 24-row viewport, 3 header lines -> 18-row budget; the block
	// plus its wrapped attribution and one body line must fit.
	mk := func(url string, attr string) tea.Msg {
		return NativeCmd(NativeRenderer{Protocol: ProtocolITerm}, func() *Photos {
			p := NewPhotos()
			p.Set(url, pngBytes(t, 400, 320))
			return p
		}(), NewNatives(), url, 50, 24, attr, 3, 35)()
	}
	noAttr := mk("https://example.com/a.png", "")
	withAttr := mk("https://example.com/b.png", "cccccc dddddd eeeeee ffffff gggggg hhhhhh iiiiii jjjjjj kkkkkk llllll mmmmmm nnnnnn oooooo pppppp")

	plain, ok := noAttr.(NativeMsg)
	if !ok {
		t.Fatalf("no-attr render should be a NativeMsg, got %T", noAttr)
	}
	fitted, ok := withAttr.(NativeMsg)
	if !ok {
		t.Fatalf("with-attr render should be a NativeMsg, got %T", withAttr)
	}
	if len(plain.Lines) != 18 {
		t.Errorf("no-attr block = %d rows, want 18 (the full budget)", len(plain.Lines))
	}
	// A wrapped attribution leaves no extra budget, so the fit loop must have
	// shrunk the photo below the plain budget.
	if len(fitted.Lines) >= len(plain.Lines) {
		t.Errorf("with-attr block = %d rows, want fewer than %d", len(fitted.Lines), len(plain.Lines))
	}
	if len(fitted.Lines) < 1 || BlockWidth(fitted.Lines) > 50 {
		t.Errorf("fitted block invalid: %d rows, width %d", len(fitted.Lines), BlockWidth(fitted.Lines))
	}
}

func TestRenderNativeFittedDecodesOnce(t *testing.T) {
	withCellPx(t, 8, 16)
	decodes := 0
	image.RegisterFormat("yerss-test/fake", "FAKEDECODE", func(r io.Reader) (image.Image, error) {
		decodes++
		return image.NewRGBA(image.Rect(0, 0, 400, 320)), nil
	}, func(r io.Reader) (image.Config, error) {
		return image.Config{Width: 400, Height: 320}, nil
	})
	// A wrapped attribution that shrinks the photo forces a second fit
	// iteration, so a decode-per-candidate bug would decode twice.
	attr := "cccccc dddddd eeeeee ffffff gggggg hhhhhh iiiii jjjjjj kkkkkk llllll mmmmmm nnnnnn oooooo pppppp"
	lines, err := renderNativeFitted(NativeRenderer{Protocol: ProtocolITerm}, []byte("FAKEDECODE"), "https://example.com/fake.png", 50, 24, 3, attr, 35)
	if err != nil {
		t.Fatal(err)
	}
	if decodes != 1 {
		t.Fatalf("decode count = %d, want 1 (the fit loop reuses the decoded image)", decodes)
	}
	if len(lines) < 1 {
		t.Error("render produced no lines")
	}
}

func TestBlockCmdNonImageContentTypeFails(t *testing.T) {
	st := newImageTestStore(t)
	id, _ := st.UpsertArticle(store.Article{FeedURL: "f", GUID: "g"})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html>not an image</html>"))
	}))
	defer srv.Close()

	msg := BlockCmd(NewCache(), NewBlocks(), st, id, srv.URL, 50, 100, true, 0)()
	if _, ok := msg.(FailedMsg); !ok {
		t.Fatalf("non-image content type should fail, got %T", msg)
	}
}

func TestFetchNoByteCap(t *testing.T) {
	// The old 5 MiB size cap was removed in favor of the content-type
	// allowlist: a body larger than any former cap must be returned in full.
	data := make([]byte, 9<<20)
	for i := range data {
		data[i] = byte(i % 256)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	got, err := fetch(srv.URL)
	if err != nil {
		t.Fatalf("fetch of a body beyond the old cap failed: %v", err)
	}
	if len(got) != len(data) {
		t.Fatalf("fetched %d bytes, want %d", len(got), len(data))
	}
}

func TestBlockCmdTimeoutFails(t *testing.T) {
	st := newImageTestStore(t)
	id, _ := st.UpsertArticle(store.Article{FeedURL: "f", GUID: "g"})
	old := fetchTimeout
	fetchTimeout = 50 * time.Millisecond
	defer func() { fetchTimeout = old }()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		time.Sleep(500 * time.Millisecond)
	}))
	defer srv.Close()

	msg := BlockCmd(NewCache(), NewBlocks(), st, id, srv.URL, 50, 100, true, 0)()
	if _, ok := msg.(FailedMsg); !ok {
		t.Fatalf("timeout should fail, got %T", msg)
	}
}

func TestNativeRendererITerm(t *testing.T) {
	withCellPx(t, 8, 16)
	data := pngBytes(t, 400, 320)
	lines, err := (NativeRenderer{Protocol: ProtocolITerm}).Render(data, "https://example.com/img.png", 50, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) > 20 {
		t.Errorf("height cap violated: %d lines", len(lines))
	}
	if len(lines) < 1 {
		t.Fatal("renderer produced no lines")
	}
	if !strings.Contains(lines[0], "\x1b]1337;File=") {
		t.Errorf("first line missing OSC 1337 escape: %q", lines[0][:min(60, len(lines[0]))])
	}
	for _, l := range lines {
		if ansi.StringWidth(l) != 50 {
			t.Errorf("line width = %d, want 50", ansi.StringWidth(l))
		}
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			t.Errorf("reserved row %d should be blank, got %q", i, lines[i])
		}
	}
}

func TestNativeRendererNarrowBlockWhenHeightCapped(t *testing.T) {
	withCellPx(t, 8, 16)
	// A tall photo capped to 10 rows narrows its box to 8 cells wide; the
	// escape line's padding and every reserved row must span that same 8-cell
	// box, not the requested width, so the caller centers the block like a
	// halfblock render and the attribution wraps to the photo's own width.
	data := pngBytes(t, 40, 100)
	lines, err := (NativeRenderer{Protocol: ProtocolKitty}).Render(data, "https://example.com/tall.png", 50, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 10 {
		t.Fatalf("reserved rows = %d, want 10", len(lines))
	}
	if !strings.HasPrefix(lines[0], "\x1b_G") {
		t.Error("first row should carry the escape")
	}
	for i, l := range lines {
		if w := ansi.StringWidth(l); w != 8 {
			t.Errorf("row %d width = %d, want 8", i, w)
		}
	}
}

func TestNativeRendererKitty(t *testing.T) {
	withCellPx(t, 8, 16)
	data := pngBytes(t, 400, 320)
	lines, err := (NativeRenderer{Protocol: ProtocolKitty}).Render(data, "https://example.com/img.png", 50, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) < 1 || !strings.HasPrefix(lines[0], "\x1b_G") {
		t.Errorf("first line missing kitty escape: %q", lines[0][:min(40, len(lines[0]))])
	}
	// The escape must carry the display box, no-cursor-movement, quiet mode,
	// and a stable placement id.
	for _, want := range []string{"a=T", "f=100", "q=1", "C=1", "c=50", "p=", "i="} {
		if !strings.Contains(lines[0], want) {
			t.Errorf("kitty escape missing %q: %q", want, lines[0][:min(120, len(lines[0]))])
		}
	}
	if len(lines) != 20 {
		t.Errorf("reserved rows = %d, want 20 (the c/r box)", len(lines))
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			t.Errorf("reserved row %d should be blank, got %q", i, lines[i])
		}
	}
}

func TestNativeRendererKittyChunksLargePhoto(t *testing.T) {
	withCellPx(t, 8, 16)
	// A photo noisy enough that its base64 payload exceeds one kitty chunk
	// must be split into m-marked APC continuations.
	data := noisyPNG(t, 640, 480)
	lines, err := (NativeRenderer{Protocol: ProtocolKitty}).Render(data, "https://example.com/big.png", 120, 100)
	if err != nil {
		t.Fatal(err)
	}
	first := lines[0]
	if !strings.Contains(first, ",m=1;") {
		t.Errorf("large payload should be chunked with m=1: %q", first[:min(80, len(first))])
	}
	if !strings.Contains(first, "\x1b_Gm=1;") || !strings.Contains(first, "\x1b_Gm=0;") {
		t.Errorf("large payload should carry m=1 and m=0 continuations")
	}
}

// noisyPNG encodes an image whose pixels are pseudo-random, so the PNG payload
// stays large enough to exceed a kitty chunk even after downscaling.
func noisyPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	seed := uint32(1)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			seed = seed*1664525 + 1013904223
			img.Set(x, y, color.RGBA{R: uint8(seed >> 24), G: uint8(seed >> 16), B: uint8(seed >> 8), A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestKittyTransmitChunksLargePayload(t *testing.T) {
	// A payload larger than one chunk must be split with m=1 continuations
	// and a terminal m=0 chunk, with continuation chunks carrying only the m
	// key.
	payload := strings.Repeat("A", kittyChunkSize*3+100)
	got := kittyTransmit(7, 10, 20, payload)
	if !strings.HasPrefix(got, "\x1b_Ga=T,f=100,q=1,i=7,p=7,c=10,r=20,C=1,m=1;") {
		t.Errorf("first chunk must carry the full control data: %q", got[:40])
	}
	chunks := strings.Split(got, "\x1b\\")
	var sizes []int
	var mids []string
	for _, c := range chunks {
		if c == "" {
			continue
		}
		sizes = append(sizes, len(c)-strings.LastIndex(c, ";")-1)
		mids = append(mids, c)
	}
	if len(sizes) != 4 {
		t.Fatalf("got %d chunks, want 4", len(sizes))
	}
	if sizes[0] != kittyChunkSize || sizes[1] != kittyChunkSize || sizes[2] != kittyChunkSize {
		t.Errorf("chunk sizes = %v, want three %d-byte chunks", sizes, kittyChunkSize)
	}
	if sizes[3] != 100 {
		t.Errorf("last chunk size = %d, want 100", sizes[3])
	}
	if !strings.HasPrefix(mids[1], "\x1b_Gm=1;") || !strings.HasPrefix(mids[2], "\x1b_Gm=1;") {
		t.Errorf("continuation chunks must carry only m=1: %q %q", mids[1][:10], mids[2][:10])
	}
	if !strings.HasPrefix(mids[3], "\x1b_Gm=0;") {
		t.Errorf("last chunk must carry m=0: %q", mids[3][:10])
	}
}

func TestKittyTransmitSmallPayloadUnchunked(t *testing.T) {
	payload := "TWFu"
	if got := kittyTransmit(3, 5, 6, payload); got != "\x1b_Ga=T,f=100,q=1,i=3,p=3,c=5,r=6,C=1;"+payload+"\x1b\\" {
		t.Errorf("small payload should be a single m-less chunk: %q", got)
	}
}

func TestNativeRendererNoneFails(t *testing.T) {
	if _, err := (NativeRenderer{Protocol: ProtocolNone}).Render(pngBytes(t, 4, 4), "u", 50, 100); err == nil {
		t.Fatal("no protocol should fail")
	}
}

func TestNativeRendererClear(t *testing.T) {
	// kitty placements survive text erases, so the clear must be an explicit
	// delete of all visible placements, quiet like the transmit.
	if got := (NativeRenderer{Protocol: ProtocolKitty}).Clear(); got != "\x1b_Ga=d,d=a,q=1\x1b\\" {
		t.Errorf("kitty Clear = %q, want the delete-all-visible-placements escape", got)
	}
	// OSC 1337 inline images are cell-bound and erased by the repaint, and
	// the none protocol never places anything: both need no clear sequence.
	for _, p := range []Protocol{ProtocolITerm, ProtocolNone} {
		if got := (NativeRenderer{Protocol: p}).Clear(); got != "" {
			t.Errorf("%v Clear = %q, want empty", p, got)
		}
	}
}

func TestHalfblocksRenderFitsWidthAndCap(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 320))
	lines, err := (Halfblocks{}).Render(img, 50, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 20 {
		t.Errorf("lines = %d, want 20", len(lines))
	}
	for _, l := range lines {
		if ansi.StringWidth(l) != 50 {
			t.Errorf("line width = %d, want 50", ansi.StringWidth(l))
		}
	}
	lines2, err := (Halfblocks{}).Render(img, 50, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines2) > 10 {
		t.Errorf("height cap violated: %d lines", len(lines2))
	}
	for _, l := range lines2 {
		if ansi.StringWidth(l) > 50 {
			t.Errorf("width exceeded: %d", ansi.StringWidth(l))
		}
	}
}

func TestFitDims(t *testing.T) {
	cases := []struct {
		srcW, srcH, width, maxH, w, h int
	}{
		{100, 80, 50, 100, 50, 20},
		{100, 80, 50, 10, 25, 10},
		{100, 80, 0, 100, 1, 1},
	}
	for _, c := range cases {
		w, h := fitDims(c.srcW, c.srcH, c.width, c.maxH)
		if w != c.w || h != c.h {
			t.Errorf("fitDims(%d,%d,%d,%d) = %d,%d, want %d,%d", c.srcW, c.srcH, c.width, c.maxH, w, h, c.w, c.h)
		}
	}
}

func TestDetectProtocolEnv(t *testing.T) {
	old := detectProtocol
	defer func() { detectProtocol = old }()
	cases := []struct {
		name string
		env  map[string]string
		want Protocol
	}{
		{"iTerm2", map[string]string{"TERM_PROGRAM": "iTerm.app"}, ProtocolITerm},
		{"wezterm", map[string]string{"TERM_PROGRAM": "WezTerm"}, ProtocolITerm},
		{"kitty term", map[string]string{"TERM": "xterm-kitty"}, ProtocolKitty},
		{"kitty program", map[string]string{"TERM_PROGRAM": "kitty"}, ProtocolKitty},
		{"kitty window id", map[string]string{"KITTY_WINDOW_ID": "1"}, ProtocolKitty},
		{"kitty pid", map[string]string{"KITTY_PID": "123"}, ProtocolKitty},
		{"ghostty term", map[string]string{"TERM": "xterm-ghostty"}, ProtocolKitty},
		{"ghostty prog", map[string]string{"TERM_PROGRAM": "ghostty"}, ProtocolKitty},
		{"ghostty dir", map[string]string{"GHOSTTY_RESOURCES_DIR": "/x"}, ProtocolKitty},
		{"custom term no kitty", map[string]string{"TERM_PROGRAM": "kitty", "TERM": "xterm-256color"}, ProtocolKitty},
		{"apple terminal", map[string]string{"TERM_PROGRAM": "Apple_Terminal"}, ProtocolNone},
		{"empty", map[string]string{}, ProtocolNone},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, k := range []string{"TMUX", "TMUX_ALLOW_PASSTHROUGH", "TERM_PROGRAM", "TERM", "GHOSTTY_RESOURCES_DIR", "KITTY_WINDOW_ID", "KITTY_PID"} {
				t.Setenv(k, "")
			}
			for k, v := range c.env {
				t.Setenv(k, v)
			}
			if got := detectEnvProtocol(); got != c.want {
				t.Errorf("detectEnvProtocol() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestDetectProtocolTmuxGated(t *testing.T) {
	t.Setenv("TMUX", "/tmp/tmux")
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("TMUX_ALLOW_PASSTHROUGH", "")
	if got := detectEnvProtocol(); got != ProtocolNone {
		t.Errorf("under tmux without passthrough protocol = %v, want none", got)
	}
	t.Setenv("TMUX_ALLOW_PASSTHROUGH", "1")
	if got := detectEnvProtocol(); got != ProtocolITerm {
		t.Errorf("under tmux with passthrough protocol = %v, want iterm", got)
	}
}

// withCellPx pins the terminal cell pixel size for native-renderer tests so
// the encoded payload is deterministic regardless of the test's controlling tty.
func withCellPx(t *testing.T, cw, ch int) {
	t.Helper()
	old := cellPixelSizeFn
	cellPixelSizeFn = func() (int, int) { return cw, ch }
	t.Cleanup(func() { cellPixelSizeFn = old })
}

func TestPixelDimsUsesCellSize(t *testing.T) {
	withCellPx(t, 8, 16)
	r := NativeRenderer{}
	if pxW, pxH := r.pixelDims(50, 20); pxW != 400 || pxH != 320 {
		t.Errorf("8x16 cell: pixelDims = %d,%d, want 400,320", pxW, pxH)
	}
	withCellPx(t, 10, 20)
	if pxW, pxH := r.pixelDims(50, 20); pxW != 500 || pxH != 400 {
		t.Errorf("10x20 cell: pixelDims = %d,%d, want 500,400", pxW, pxH)
	}
	// An oversized cell box is capped to bound the payload.
	pxW, pxH := r.pixelDims(1000, 20)
	if pxW > 4096 || pxH > 4096 {
		t.Errorf("pixelDims not capped: %d,%d", pxW, pxH)
	}
	if pxW != 4096 {
		t.Errorf("capped width = %d, want 4096", pxW)
	}
}

func TestBlockCmdPersistsInlineImageAtPosition(t *testing.T) {
	st := newImageTestStore(t)
	id, err := st.UpsertArticle(store.Article{FeedURL: "f", GUID: "g"})
	if err != nil {
		t.Fatal(err)
	}
	data := pngBytes(t, 480, 320)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	msg := BlockCmd(NewCache(), NewBlocks(), st, id, srv.URL, 50, 100, true, 2)()
	if _, ok := msg.(BlockMsg); !ok {
		t.Fatalf("expected BlockMsg, got %T", msg)
	}
	imgs, err := st.GetArticleImages(id)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, im := range imgs {
		if im.Position == 2 && im.URL == srv.URL && len(im.Photo) > 0 && BlockWidth(splitLines(im.Block)) == 50 {
			found = true
		}
	}
	if !found {
		t.Errorf("inline image not persisted at position 2: %+v", imgs)
	}
}

func TestBlockCmdServesInlineStoredBlockByURL(t *testing.T) {
	st := newImageTestStore(t)
	id, err := st.UpsertArticle(store.Article{FeedURL: "f", GUID: "g"})
	if err != nil {
		t.Fatal(err)
	}
	want, _ := (Halfblocks{}).Render(image.NewRGBA(image.Rect(0, 0, 480, 320)), 40, 100)
	if err := st.SetArticleImage(id, store.ArticleImage{Position: 3, URL: "https://example.com/inline.png", Block: strings.Join(want, "\n"), Width: BlockWidth(want)}); err != nil {
		t.Fatal(err)
	}
	msg := BlockCmd(NewCache(), NewBlocks(), st, id, "https://example.com/inline.png", 40, 100, true, 3)()
	bm, ok := msg.(BlockMsg)
	if !ok {
		t.Fatalf("expected BlockMsg from stored inline block, got %T", msg)
	}
	if BlockWidth(bm.Lines) != 40 {
		t.Errorf("served inline block width = %d, want 40", BlockWidth(bm.Lines))
	}
}

func TestHalfblocksSmallImageRendersAtNaturalSize(t *testing.T) {
	// A source image narrower than smallImageWidth renders at a capped screen
	// width rather than being upscaled to fill the content column, so a tiny
	// site logo does not balloon into a full-column block.
	img := image.NewRGBA(image.Rect(0, 0, 60, 40))
	lines, err := (Halfblocks{}).Render(img, 50, 100)
	if err != nil {
		t.Fatal(err)
	}
	if w := BlockWidth(lines); w != 20 {
		t.Errorf("small image block width = %d, want 20 (the small-image cap)", w)
	}
	// A 200px logo renders at roughly 20 cells, not the full column.
	logo := image.NewRGBA(image.Rect(0, 0, 200, 100))
	lines, err = (Halfblocks{}).Render(logo, 74, 100)
	if err != nil {
		t.Fatal(err)
	}
	if w := BlockWidth(lines); w != 20 {
		t.Errorf("200px logo block width = %d, want 20", w)
	}
	// A regular photo still fills the column.
	big := image.NewRGBA(image.Rect(0, 0, 480, 320))
	lines, err = (Halfblocks{}).Render(big, 50, 100)
	if err != nil {
		t.Fatal(err)
	}
	if w := BlockWidth(lines); w != 50 {
		t.Errorf("photo block width = %d, want 50", w)
	}
}

func TestNativeRendererSmallImageRendersAtNaturalSize(t *testing.T) {
	withCellPx(t, 8, 16)
	data := pngBytes(t, 60, 40)
	lines, err := (NativeRenderer{Protocol: ProtocolKitty}).Render(data, "https://example.com/logo.png", 50, 20)
	if err != nil {
		t.Fatal(err)
	}
	// The capped cell width is 20; the escape must carry that box, not the
	// requested 50.
	if !strings.Contains(lines[0], "c=20") {
		t.Errorf("kitty escape should carry the capped width c=20: %q", lines[0][:min(120, len(lines[0]))])
	}
}

func TestDeleteByID(t *testing.T) {
	got := DeleteByID(7)
	want := "\x1b_Ga=d,d=i,i=7,q=1\x1b\\"
	if got != want {
		t.Errorf("DeleteByID = %q, want %q", got, want)
	}
	if !strings.HasPrefix(got, "\x1b_Ga=d,d=i,i=") || !strings.HasSuffix(got, ",q=1\x1b\\") {
		t.Errorf("DeleteByID shape wrong: %q", got)
	}
	if got := DeleteByID(0); got != "" {
		t.Errorf("DeleteByID(0) = %q, want empty (nothing transmitted)", got)
	}
}

func TestRenderIDDistinctPerSizeAndReused(t *testing.T) {
	url := "https://example.com/img.png"
	a := renderID(url, 400, 300)
	b := renderID(url, 800, 600)
	if a == b {
		t.Errorf("same URL at different pixel sizes must produce distinct ids: both %d", a)
	}
	if got := renderID(url, 400, 300); got != a {
		t.Errorf("same URL and pixel size must reuse the id: got %d, want %d", got, a)
	}
	if got := renderID("https://example.com/other.png", 400, 300); got == a {
		t.Errorf("different URLs at the same size must produce distinct ids: both %d", got)
	}
	for _, id := range []uint32{a, b, renderID(url, 0, 0), renderID("", 10, 10)} {
		if id == 0 {
			t.Error("renderID must never return zero (the protocol reserves image id 0)")
		}
	}
}

func TestNativeRenderIDReadsKittyIDFromLines(t *testing.T) {
	if got := NativeRenderID([]string{"\x1b_Ga=T,f=100,q=1,i=42,p=42,c=2,r=2;AAAA\x1b\\  "}); got != 42 {
		t.Errorf("NativeRenderID = %d, want 42", got)
	}
	if got := NativeRenderID([]string{"\x1b]1337;File=name=AA;size=1;inline=1;width=2;height=2:AAAA\x07"}); got != 0 {
		t.Errorf("iTerm render should have no kitty id, got %d", got)
	}
	if got := NativeRenderID([]string{"plain"}); got != 0 {
		t.Errorf("halfblock line should have no kitty id, got %d", got)
	}
	if got := NativeRenderID(nil); got != 0 {
		t.Errorf("empty lines should have no kitty id, got %d", got)
	}
}
