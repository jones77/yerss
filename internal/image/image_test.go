package image

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestCacheGetSet(t *testing.T) {
	c := NewCache()
	if _, ok := c.Get("u"); ok {
		t.Error("empty cache should miss")
	}
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	c.Set("u", img)
	got, ok := c.Get("u")
	if !ok || got != img {
		t.Errorf("cache roundtrip failed: ok=%v", ok)
	}
}

func TestLoadCmdInMemoryHit(t *testing.T) {
	st := newImageTestStore(t)
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	c := NewCache()
	c.Set("https://example.com/img.png", img)
	msg := LoadCmd(c, st, 1, "https://example.com/img.png")()
	lm, ok := msg.(LoadedMsg)
	if !ok {
		t.Fatalf("expected LoadedMsg, got %T", msg)
	}
	if lm.Img != img {
		t.Error("in-memory hit should return the cached image")
	}
}

func TestLoadCmdDBHit(t *testing.T) {
	st := newImageTestStore(t)
	id, err := st.UpsertArticle(store.Article{FeedURL: "f", GUID: "g"})
	if err != nil {
		t.Fatal(err)
	}
	data := pngBytes(t, 4, 4)
	if err := st.SetImageData(id, data); err != nil {
		t.Fatal(err)
	}
	c := NewCache()
	msg := LoadCmd(c, st, id, "https://example.com/img.png")()
	lm, ok := msg.(LoadedMsg)
	if !ok {
		t.Fatalf("expected LoadedMsg from DB, got %T", msg)
	}
	if _, ok := c.Get("https://example.com/img.png"); !ok {
		t.Error("DB hit should populate the in-memory cache")
	}
	_ = lm
}

func TestLoadCmdNetworkFetchPersists(t *testing.T) {
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
	msg := LoadCmd(c, st, id, srv.URL)()
	if _, ok := msg.(FailedMsg); ok {
		t.Fatal("network fetch should succeed")
	}
	lm, ok := msg.(LoadedMsg)
	if !ok {
		t.Fatalf("expected LoadedMsg, got %T", msg)
	}
	if lm.Img.Bounds().Dx() != 4 {
		t.Errorf("decoded image width = %d, want 4", lm.Img.Bounds().Dx())
	}
	stored, err := st.GetImageData(id)
	if err != nil {
		t.Fatal(err)
	}
	if string(stored) != string(data) {
		t.Error("network bytes should be persisted to the database")
	}
	if _, ok := c.Get(srv.URL); !ok {
		t.Error("network fetch should populate the in-memory cache")
	}
}

func TestLoadCmdNonImageContentTypeFails(t *testing.T) {
	st := newImageTestStore(t)
	id, _ := st.UpsertArticle(store.Article{FeedURL: "f", GUID: "g"})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html>not an image</html>"))
	}))
	defer srv.Close()

	msg := LoadCmd(NewCache(), st, id, srv.URL)()
	if _, ok := msg.(FailedMsg); !ok {
		t.Fatalf("non-image content type should fail, got %T", msg)
	}
}

func TestLoadCmdOversizedFails(t *testing.T) {
	st := newImageTestStore(t)
	id, _ := st.UpsertArticle(store.Article{FeedURL: "f", GUID: "g"})
	old := maxImageBytes
	maxImageBytes = 1024
	defer func() { maxImageBytes = old }()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(bytes.Repeat([]byte{0x00}, 4096))
	}))
	defer srv.Close()

	msg := LoadCmd(NewCache(), st, id, srv.URL)()
	if _, ok := msg.(FailedMsg); !ok {
		t.Fatalf("oversized response should fail, got %T", msg)
	}
}

func TestLoadCmdTimeoutFails(t *testing.T) {
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

	msg := LoadCmd(NewCache(), st, id, srv.URL)()
	if _, ok := msg.(FailedMsg); !ok {
		t.Fatalf("timeout should fail, got %T", msg)
	}
}

func TestHalfblocksRenderFitsWidthAndCap(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 80))
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