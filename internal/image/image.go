// Package image fetches, caches, and renders article lead images.
package image

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/mosaic"

	"yerss/internal/store"
)

// LoadedMsg reports a successfully decoded lead image.
type LoadedMsg struct {
	Key string
	Img image.Image
}

// FailedMsg reports a lead image that could not be loaded.
type FailedMsg struct {
	Key string
}

// fetchTimeout bounds a single image download. It is a variable so tests can
// shorten it.
var fetchTimeout = 15 * time.Second

// maxImageBytes caps the size of a fetched image. It is a variable so tests
// can shrink it.
var maxImageBytes = 5 << 20

// Cache holds decoded images keyed by URL for the session.
type Cache struct {
	mu    sync.Mutex
	items map[string]image.Image
}

// NewCache returns an empty in-memory decoded-image cache.
func NewCache() *Cache {
	return &Cache{items: make(map[string]image.Image)}
}

// Get returns the cached decoded image for url, if present.
func (c *Cache) Get(url string) (image.Image, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	img, ok := c.items[url]
	return img, ok
}

// Set stores a decoded image under url.
func (c *Cache) Set(url string, img image.Image) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[url] = img
}

// LoadCmd returns a tea.Cmd that resolves url's image through the cache
// hierarchy (in-memory cache, stored database bytes, then a network fetch). On
// a network success the raw bytes are persisted and the decoded image cached;
// on any failure a FailedMsg is emitted. The command never blocks the caller.
func LoadCmd(cache *Cache, st *store.Store, articleID int64, url string) tea.Cmd {
	return func() tea.Msg {
		if img, ok := cache.Get(url); ok {
			return LoadedMsg{Key: url, Img: img}
		}
		if data, err := st.GetImageData(articleID); err == nil && len(data) > 0 {
			if img, _, err := image.Decode(bytes.NewReader(data)); err == nil {
				cache.Set(url, img)
				return LoadedMsg{Key: url, Img: img}
			}
		}
		data, err := fetch(url)
		if err != nil {
			return FailedMsg{Key: url}
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return FailedMsg{Key: url}
		}
		_ = st.SetImageData(articleID, data)
		cache.Set(url, img)
		return LoadedMsg{Key: url, Img: img}
	}
}

// fetch downloads url with guards: an HTTP timeout, an image/* content-type
// allowlist, and a max-bytes cap on the response body.
func fetch(url string) ([]byte, error) {
	client := &http.Client{Timeout: fetchTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("image fetch: status %s", resp.Status)
	}
	if ctype := resp.Header.Get("Content-Type"); !strings.HasPrefix(strings.ToLower(ctype), "image/") {
		return nil, fmt.Errorf("image fetch: content type %q is not an image", ctype)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxImageBytes+1)))
	if err != nil {
		return nil, err
	}
	if len(body) > maxImageBytes {
		return nil, fmt.Errorf("image fetch: response exceeds %d bytes", maxImageBytes)
	}
	return body, nil
}

// Renderer renders a decoded image to terminal text lines at a given width.
type Renderer interface {
	Render(img image.Image, width, maxHeight int) ([]string, error)
}

// Halfblocks renders images as ANSI halfblock characters via x/mosaic.
type Halfblocks struct{}

// Render renders img to halfblock lines, fitting width while preserving aspect
// ratio and never exceeding maxHeight rows. mosaic's Width/Height are in source
// pixels and produce exactly half as many cells/rows, so the cell dimensions
// are doubled before rendering.
func (Halfblocks) Render(img image.Image, width, maxHeight int) ([]string, error) {
	b := img.Bounds()
	srcW, srcH := b.Dx(), b.Dy()
	if srcW < 1 || srcH < 1 {
		return nil, fmt.Errorf("image has no pixels")
	}
	w, h := fitDims(srcW, srcH, width, maxHeight)
	m := mosaic.New()
	m = m.Width(w * 2).Height(h * 2)
	art := m.Render(img)
	return splitLines(art), nil
}

// fitDims computes the output cell dimensions (w cells wide, h rows tall) that
// preserve the source's on-screen aspect ratio while fitting within width and
// maxHeight. A terminal cell is roughly twice as tall as it is wide, and a
// halfblock cell renders two source pixels vertically, so a source of srcW×srcH
// at w cells wide is h = w*srcH/(2*srcW) rows tall.
func fitDims(srcW, srcH, width, maxHeight int) (w, h int) {
	if width < 1 {
		width = 1
	}
	if maxHeight < 1 {
		maxHeight = 1
	}
	w = width
	h = (w*srcH + srcW*2 - 1) / (srcW * 2)
	if h < 1 {
		h = 1
	}
	if h > maxHeight {
		h = maxHeight
		w = (h*2*srcW + srcH - 1) / srcH
		if w < 1 {
			w = 1
		}
	}
	return w, h
}

// splitLines splits a mosaic render into lines, dropping the trailing empty
// segment produced by the final newline.
func splitLines(art string) []string {
	s := strings.TrimSuffix(art, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// LineWidths reports the visible cell width of each rendered line (for tests).
func LineWidths(lines []string) []int {
	ws := make([]int, len(lines))
	for i, l := range lines {
		ws[i] = ansi.StringWidth(l)
	}
	return ws
}