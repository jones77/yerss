package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal/image"
	"yerss/internal/store"
)

// imagesEnabled reports whether lead image rendering is active: the config mode
// is not off and ASCII fallback is not forcing images off (the -a flag and
// terminal detection both set m.ascii).
func (m *Model) imagesEnabled() bool {
	return !m.ascii && m.cfg.Display.Images != "off"
}

// articleImageBlock renders the article's lead image (when loaded and enabled)
// as halfblock lines sized to the content width. The block height is capped so
// the full image plus the header and one line of body fit within the viewport.
// It returns nil when images are disabled, no URL is set, the image is not yet
// loaded, or rendering fails.
func (m *Model) articleImageBlock(a store.Article, contentW, vpH int) []string {
	if !m.imagesEnabled() || a.ImageURL == "" {
		return nil
	}
	img, ok := m.imgCache.Get(a.ImageURL)
	if !ok {
		return nil
	}
	headerLines := 0
	if header := m.renderMarkdown(articleHeaderMarkdown(a), contentW); header != "" {
		headerLines = len(strings.Split(header, "\n"))
	}
	maxH := max(1, vpH-headerLines-1)
	lines, err := m.imgRenderer.Render(img, contentW, maxH)
	if err != nil {
		return nil
	}
	return lines
}

// fireImageLoad returns a tea.Cmd that loads the article's lead image through
// the cache hierarchy when image rendering is enabled and a URL is present.
func (m *Model) fireImageLoad(a store.Article) tea.Cmd {
	if !m.imagesEnabled() || a.ImageURL == "" {
		return nil
	}
	return image.LoadCmd(m.imgCache, m.store, a.ID, a.ImageURL)
}

// onImageLoaded re-composes the open article with its freshly loaded image
// block, preserving the user's reading position: if the offset was already past
// the insertion point (the top of the content), it is bumped by the number of
// inserted image lines.
func (m *Model) onImageLoaded(msg image.LoadedMsg) {
	if m.view != viewArticle {
		return
	}
	a := *m.article.article
	if msg.Key != a.ImageURL {
		return
	}
	oldOffset := m.article.viewport.YOffset
	oldLen := len(m.article.lines)
	m.article = m.newArticleState(a)
	inserted := len(m.article.lines) - oldLen
	if oldOffset > 0 {
		m.article.viewport.SetYOffset(oldOffset + inserted)
	} else {
		m.article.viewport.SetYOffset(0)
	}
}

// onImageFailed is a no-op: the article already renders without an image block
// and the failure is never surfaced to the user.
func (m *Model) onImageFailed(image.FailedMsg) {}

// snapYOffset implements atomic image scroll: after a scroll move lands the
// offset in the image block's line range [imgStart, imgEnd], a downward move
// snaps past the block (imgEnd+1, image vanishes) and an upward move reveals
// the full image (imgStart). A range of -1 marks "no image block". Offsets
// outside the range (or moves that never enter it) are unchanged.
func snapYOffset(yOffset, vpH, imgStart, imgEnd, direction int) int {
	if imgStart < 0 || imgEnd < imgStart {
		return yOffset
	}
	if yOffset >= imgStart && yOffset <= imgEnd {
		if direction > 0 {
			return imgEnd + 1
		}
		if direction < 0 {
			return imgStart
		}
	}
	return yOffset
}

// scrollArticle runs a viewport scroll operation then snaps the offset out of
// the image block interior when the move partially clips the image, using the
// move's direction (down/up) to decide which side of the block to land on.
func (m *Model) scrollArticle(fn func()) {
	before := m.article.viewport.YOffset
	fn()
	after := m.article.viewport.YOffset
	dir := 0
	if after > before {
		dir = 1
	} else if after < before {
		dir = -1
	}
	snapped := snapYOffset(after, m.article.viewport.Height, m.article.imgStart, m.article.imgEnd, dir)
	if snapped != after {
		m.article.viewport.SetYOffset(snapped)
	}
}