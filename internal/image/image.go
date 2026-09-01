// Package image fetches, caches, and renders article lead images. Decoded
// images, rendered halfblock text blocks, and fetched raw photo bytes are each
// cached in memory for the session; both the rendered text block and the full
// photo bytes are persisted together to the article_images table.
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

// BlockCmd returns a tea.Cmd that resolves an article's image block at width
// through the cache hierarchy: in-memory decoded image (re-rendered), a stored
// database block whose width matches, a stored database photo (re-rendered at
// the current width), then a network fetch. On reaching the stored photo or a
// network fetch the block and its photo bytes are persisted together at
// position (0 for the lead image, 1..N for inline images); on any failure a
// FailedMsg is emitted. When fetch is false the command never hits the network:
// that is the native-terminal placeholder path, where PhotoCmd owns fetching.
func BlockCmd(cache *Cache, blocks *Blocks, st *store.Store, articleID int64, url string, width, maxHeight int, allowFetch bool, position int) tea.Cmd {
	return func() tea.Msg {
		if lines, ok := blockFromCache(cache, blocks, url, width, maxHeight); ok {
			return BlockMsg{Key: url, Lines: lines}
		}
		if imgs, err := st.GetArticleImages(articleID); err == nil {
			for i := range imgs {
				if imgs[i].URL != url {
					continue
				}
				img := &imgs[i]
				if img.Block != "" && img.Width == width && blockWidthMatches(*img, width) {
					stored := splitLines(img.Block)
					blocks.Set(url, stored)
					return BlockMsg{Key: url, Lines: stored}
				}
				if len(img.Photo) > 0 {
					if decoded, _, derr := image.Decode(bytes.NewReader(img.Photo)); derr == nil {
						cache.Set(url, decoded)
						if lines, err := renderBlock(blocks, url, decoded, width, maxHeight); err == nil {
							_ = st.SetArticleImageBlock(articleID, img.Position, url, strings.Join(lines, "\n"), BlockWidth(lines))
							return BlockMsg{Key: url, Lines: lines}
						}
					}
				}
				break
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
		lines, err := renderBlock(blocks, url, img, width, maxHeight)
		if err != nil {
			return FailedMsg{Key: url}
		}
		persistImage(st, articleID, url, lines, data, position)
		return BlockMsg{Key: url, Lines: lines}
	}
}

// PhotoCmd returns a tea.Cmd that resolves an article's photo bytes for a
// native render, preferring the session Photos cache and the stored photo
// before fetching. When a fresh fetch is needed the photo is also used to render
// and persist the halfblock block so future opens have a placeholder, and both
// the block and photo are persisted together at position (0 for the lead image,
// 1..N for inline images). On any failure a FailedMsg is emitted.
func PhotoCmd(cache *Cache, blocks *Blocks, photos *Photos, st *store.Store, articleID int64, url string, width, maxHeight, position int) tea.Cmd {
	return func() tea.Msg {
		if _, ok := photos.Get(url); ok {
			return PhotoMsg{Key: url}
		}
		if data, err := st.GetArticleImagePhoto(articleID, url); err == nil && len(data) > 0 {
			photos.Set(url, data)
			return PhotoMsg{Key: url}
		}
		data, err := fetch(url)
		if err != nil {
			return FailedMsg{Key: url}
		}
		photos.Set(url, data)
		var blockLines []string
		if _, ok := cache.Get(url); !ok {
			if img, _, derr := image.Decode(bytes.NewReader(data)); derr == nil {
				cache.Set(url, img)
				if _, ok := blocks.Get(url); !ok {
					blockLines, _ = renderBlock(blocks, url, img, width, maxHeight)
				}
			}
		}
		persistImage(st, articleID, url, blockLines, data, position)
		return PhotoMsg{Key: url}
	}
}

// persistImage writes the rendered block and the raw photo bytes together as
// the article image at the given position (0 for the lead image, 1..N for
// inline images). block may be nil when only the photo is available.
func persistImage(st *store.Store, id int64, url string, block []string, photo []byte, position int) {
	_ = st.SetArticleImage(id, store.ArticleImage{
		Position: position,
		URL:      url,
		Block:    strings.Join(block, "\n"),
		Photo:    photo,
		Width:    BlockWidth(block),
	})
}

// renderBlock renders img to a block at width (capped to maxHeight) and caches
// it in blocks.
func renderBlock(blocks *Blocks, url string, img image.Image, width, maxHeight int) ([]string, error) {
	lines, err := (Halfblocks{}).Render(img, width, maxHeight)
	if err != nil {
		return nil, err
	}
	blocks.Set(url, lines)
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
// are doubled before rendering. Small source images (logos, icons) are rendered
// at their natural cell size rather than upscaled to fill width.
func (Halfblocks) Render(img image.Image, width, maxHeight int) ([]string, error) {
	b := img.Bounds()
	srcW, srcH := b.Dx(), b.Dy()
	if srcW < 1 || srcH < 1 {
		return nil, fmt.Errorf("image has no pixels")
	}
	w, h := fitDims(srcW, srcH, renderWidth(srcW, srcH, width), maxHeight)
	m := mosaic.New()
	m = m.Width(w * 2).Height(h * 2)
	art := m.Render(img)
	return splitLines(art), nil
}

// smallImageWidth is the shorter source dimension below which an image is
// treated as a small logo/icon and rendered at a capped screen width rather
// than upscaled to fill the content column. Site logos, icons, and inline
// decorations are narrow or short in at least one dimension (a 300x150 promo
// banner, a 100x100 favicon); news photos are always large in both.
const smallImageWidth = 240

// smallImageMaxCells is the widest a small image renders on screen, so a tiny
// logo (say 300x150 source pixels) occupies a modest, logo-sized block instead
// of a full-column one.
const smallImageMaxCells = 20

// renderWidth returns the cell width an image of srcW×srcH source pixels is
// rendered at: when its shorter side is below smallImageWidth it renders at
// its natural cell width (halfblock: two source pixels per cell) capped to
// smallImageMaxCells and the content width; otherwise it fills contentW. This
// keeps a tiny logo from ballooning into a full-column block.
func renderWidth(srcW, srcH, contentW int) int {
	if contentW < 1 {
		return 1
	}
	if min(srcW, srcH) < smallImageWidth {
		natural := max(1, srcW/2)
		return min(contentW, min(smallImageMaxCells, natural))
	}
	return contentW
}

// blockWidthMatches reports whether a stored block at img.Width renders the
// image at the width it should for the current content width. When the photo
// is stored, its source dimensions determine the expected width (a small image
// renders narrower than the column, so a stale full-width block stored before
// the small-image cap is rejected and re-rendered from the photo); without a
// photo the stored width matching contentW is trusted.
func blockWidthMatches(img store.ArticleImage, contentW int) bool {
	if len(img.Photo) > 0 {
		if cfg, _, err := image.DecodeConfig(bytes.NewReader(img.Photo)); err == nil {
			return img.Width == renderWidth(cfg.Width, cfg.Height, contentW)
		}
	}
	return img.Width == contentW
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
