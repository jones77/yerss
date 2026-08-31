// Package image fetches, caches, and renders article lead images. Decoded
// images, rendered halfblock text blocks, and fetched raw photo bytes are each
// cached in memory for the session; only the rendered text block is persisted.
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

// BlockMsg reports an article's lead image rendered as a halfblock text block.
type BlockMsg struct {
	Key   string
	Lines []string
}

// PhotoMsg reports that an article's lead photo has been fetched and cached
// for a native render. The raw bytes are read from the Photos cache.
type PhotoMsg struct {
	Key string
}

// NativeMsg reports that an article's lead photo has been rendered to a native
// inline-image block off the UI goroutine and cached for the session. Lines are
// the rendered native block (image rows only, not yet centered or attributed);
// they are also available from the Natives cache keyed by render size.
type NativeMsg struct {
	Key   string
	Lines []string
}

// FailedMsg reports a lead image that could not be loaded.
type FailedMsg struct {
	Key string
}

// fetchTimeout bounds a single image download. It is a variable so tests can
// shorten it.
var fetchTimeout = 15 * time.Second

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

// Blocks holds rendered halfblock text blocks keyed by URL for the session.
type Blocks struct {
	mu    sync.Mutex
	items map[string][]string
}

// NewBlocks returns an empty in-memory rendered-block cache.
func NewBlocks() *Blocks {
	return &Blocks{items: make(map[string][]string)}
}

// Get returns the cached rendered block for url, if present.
func (b *Blocks) Get(url string) ([]string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	lines, ok := b.items[url]
	return lines, ok
}

// Set stores a rendered block under url.
func (b *Blocks) Set(url string, lines []string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items[url] = lines
}

// Photos holds raw fetched photo bytes keyed by URL for the session.
type Photos struct {
	mu    sync.Mutex
	items map[string][]byte
}

// NewPhotos returns an empty in-memory raw-photo cache.
func NewPhotos() *Photos {
	return &Photos{items: make(map[string][]byte)}
}

// Get returns the cached raw photo bytes for url, if present.
func (p *Photos) Get(url string) ([]byte, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	data, ok := p.items[url]
	return data, ok
}

// Set stores raw photo bytes under url.
func (p *Photos) Set(url string, data []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.items[url] = data
}

// Natives holds native inline-image block renderings keyed by render size for
// the session, so re-viewing an article at the same width and viewport height
// is instant and resizing to a new size re-renders off the UI goroutine.
type Natives struct {
	mu    sync.Mutex
	items map[string][]string
}

// NewNatives returns an empty in-memory rendered-native-block cache.
func NewNatives() *Natives {
	return &Natives{items: make(map[string][]string)}
}

// Get returns the cached native render for key, if present.
func (n *Natives) Get(key string) ([]string, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	lines, ok := n.items[key]
	return lines, ok
}

// Set stores a native render under key.
func (n *Natives) Set(key string, lines []string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.items[key] = lines
}

// NativeKey returns the cache key for a native render of url at the given
// content width and viewport-height cap. Both dimensions matter: changing
// either means a different block size and warrants a re-render.
func NativeKey(url string, width, maxHeight int) string {
	return fmt.Sprintf("%s\x00%d\x00%d", url, width, maxHeight)
}

// BlockWidth reports the display width in cells of a rendered block. A mosaic
// render is exactly the requested width per line, so this is the block's
// render width.
func BlockWidth(lines []string) int {
	w := 0
	for _, l := range lines {
		if lw := ansi.StringWidth(l); lw > w {
			w = lw
		}
	}
	return w
}

// BlockCmd returns a tea.Cmd that resolves an article's lead image block at
// width through the cache hierarchy: in-memory decoded image (re-rendered),
// the stored database block when its width matches, legacy stored photo bytes,
// then a network fetch. On success the block is cached and, when new, persisted
// (which purges any legacy photo bytes); on any failure a FailedMsg is emitted.
// When fetch is false the command never hits the network: that is the
// native-terminal placeholder path, where PhotoCmd owns fetching.
func BlockCmd(cache *Cache, blocks *Blocks, st *store.Store, articleID int64, url string, width, maxHeight int, allowFetch bool) tea.Cmd {
	return func() tea.Msg {
		if lines, ok := blockFromCache(cache, blocks, url, width, maxHeight); ok {
			return BlockMsg{Key: url, Lines: lines}
		}
		if block, err := st.GetImageBlock(articleID); err == nil && block != "" {
			if stored := splitLines(block); BlockWidth(stored) == width {
				blocks.Set(url, stored)
				return BlockMsg{Key: url, Lines: stored}
			}
		}
		if data, err := st.GetImageData(articleID); err == nil && len(data) > 0 {
			if img, _, derr := image.Decode(bytes.NewReader(data)); derr == nil {
				cache.Set(url, img)
				if lines, err := renderBlock(blocks, st, articleID, url, img, width, maxHeight); err == nil {
					return BlockMsg{Key: url, Lines: lines}
				}
			}
		}
		if !allowFetch {
			return FailedMsg{Key: url}
		}
		data, err := fetch(url)
		if err != nil {
			return FailedMsg{Key: url}
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return FailedMsg{Key: url}
		}
		cache.Set(url, img)
		lines, err := renderBlock(blocks, st, articleID, url, img, width, maxHeight)
		if err != nil {
			return FailedMsg{Key: url}
		}
		return BlockMsg{Key: url, Lines: lines}
	}
}

// PhotoCmd returns a tea.Cmd that fetches an article's lead photo into the
// Photos cache for a native render. When the fetched photo is not already
// decoded, it is also used to render and persist the halfblock block so future
// opens have a placeholder. On any failure a FailedMsg is emitted.
func PhotoCmd(cache *Cache, blocks *Blocks, photos *Photos, st *store.Store, articleID int64, url string, width, maxHeight int) tea.Cmd {
	return func() tea.Msg {
		if _, ok := photos.Get(url); ok {
			return PhotoMsg{Key: url}
		}
		data, err := fetch(url)
		if err != nil {
			return FailedMsg{Key: url}
		}
		photos.Set(url, data)
		if _, ok := cache.Get(url); !ok {
			if img, _, derr := image.Decode(bytes.NewReader(data)); derr == nil {
				cache.Set(url, img)
				if _, ok := blocks.Get(url); !ok {
					_, _ = renderBlock(blocks, st, articleID, url, img, width, maxHeight)
				}
			}
		}
		return PhotoMsg{Key: url}
	}
}

// renderBlock renders img to a block at width (capped to maxHeight), caches it
// in blocks, and persists it to the database.
func renderBlock(blocks *Blocks, st *store.Store, articleID int64, url string, img image.Image, width, maxHeight int) ([]string, error) {
	lines, err := (Halfblocks{}).Render(img, width, maxHeight)
	if err != nil {
		return nil, err
	}
	blocks.Set(url, lines)
	_ = st.SetImageBlock(articleID, strings.Join(lines, "\n"))
	return lines, nil
}

// blockFromCache re-renders the block from a cached decoded image when one is
// present, or serves the cached block lines when they match the width.
func blockFromCache(cache *Cache, blocks *Blocks, url string, width, maxHeight int) ([]string, bool) {
	if img, ok := cache.Get(url); ok {
		if lines, err := (Halfblocks{}).Render(img, width, maxHeight); err == nil {
			blocks.Set(url, lines)
			return lines, true
		}
	}
	if lines, ok := blocks.Get(url); ok && BlockWidth(lines) == width {
		return lines, true
	}
	return nil, false
}

// fetch downloads url with guards: an HTTP timeout and an image/* content-type
// allowlist.
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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
	width = max(1, width)
	maxHeight = max(1, maxHeight)
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