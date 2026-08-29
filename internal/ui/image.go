package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yerss/convert"
	"yerss/internal/image"
	"yerss/internal/store"
)

// imagesEnabled reports whether lead image rendering is active: the config mode
// is not off and ASCII fallback is not forcing images off (the -a flag and
// terminal detection both set m.ascii).
func (m *Model) imagesEnabled() bool {
	return !m.ascii && m.cfg.Display.Images != "off"
}

// articleImageBlock renders the article's lead image (when loaded and
// enabled) as halfblock lines with a dim attribution centered directly
// beneath. The block sits below the reader header with one blank line on
// each side, and its height is capped so the whole block plus at least one
// line of body text fits within the viewport alongside the header
// (headerLines, computed by the caller). The photo is centered as a unit and
// the attribution is wrapped to the photo's own width; a narrow
// (height-capped) photo can push the attribution onto several lines, so the
// photo is re-rendered smaller until the block fits. It returns nil when
// images are disabled, no URL is set, the image is not yet loaded, or
// rendering fails.
func (m *Model) articleImageBlock(a store.Article, contentW, vpH, headerLines int) []string {
	if !m.imagesEnabled() || a.ImageURL == "" {
		return nil
	}
	img, ok := m.imgCache.Get(a.ImageURL)
	if !ok {
		return nil
	}
	attr := articleAttribution(a)

	// Budget: blank line above + attribution lines + blank line below + one
	// body line. The attribution line count is unknown until the photo's
	// width is known (it wraps to the photo), so render optimistically
	// reserving one attribution line, then re-render smaller if the wrapped
	// attribution needs more. Bounded at three iterations.
	base := vpH - headerLines - 3
	maxH := max(1, base-1)
	if attr == "" {
		maxH = max(1, base)
	}
	var rendered, attrLines []string
	for i := 0; ; i++ {
		var err error
		rendered, err = m.imgRenderer.Render(img, contentW, maxH)
		if err != nil {
			return nil
		}
		imgW := 0
		for _, l := range rendered {
			if w := ansi.StringWidth(l); w > imgW {
				imgW = w
			}
		}
		attrLines = nil
		if attr != "" {
			attrLines = wrapLines(attr, imgW)
		}
		want := max(1, base-len(attrLines))
		if want == maxH || i == 2 {
			break
		}
		maxH = want
	}

	imgW := 0
	for _, l := range rendered {
		if w := ansi.StringWidth(l); w > imgW {
			imgW = w
		}
	}
	photoPad := max(0, (contentW-imgW)/2)
	out := make([]string, 0, len(rendered)+len(attrLines))
	for _, l := range rendered {
		out = append(out, strings.Repeat(" ", photoPad)+l)
	}
	style := lipgloss.NewStyle().Foreground(m.palette.Dim)
	for _, al := range attrLines {
		pad := photoPad + max(0, (imgW-ansi.StringWidth(al))/2)
		out = append(out, strings.Repeat(" ", pad)+style.Render(al))
	}
	return out
}

// wrapLines wraps plain text to w display columns, breaking on spaces and
// hard-splitting any word longer than w. Non-empty input yields at least one
// line.
func wrapLines(s string, w int) []string {
	if w < 1 {
		w = 1
	}
	var lines []string
	cur := ""
	for _, word := range strings.Fields(s) {
		for ansi.StringWidth(word) > w {
			var head string
			head, word = splitAtWidth(word, w)
			if cur != "" {
				lines = append(lines, cur)
				cur = ""
			}
			lines = append(lines, head)
		}
		if cur == "" {
			cur = word
		} else if ansi.StringWidth(cur)+1+ansi.StringWidth(word) <= w {
			cur += " " + word
		} else {
			lines = append(lines, cur)
			cur = word
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

// splitAtWidth splits s into the longest prefix of at most w display columns
// and the remainder.
func splitAtWidth(s string, w int) (head, tail string) {
	x := 0
	for i, r := range s {
		if x += ansi.StringWidth(string(r)); x > w {
			return s[:i], s[i:]
		}
	}
	return s, ""
}

// articleAttribution returns the photo attribution for the article's lead
// image: the credit extracted from the article's HTML when present, otherwise
// "photo: <source>" derived by the list view's source-identifier rules. It
// returns "" when neither is available.
func articleAttribution(a store.Article) string {
	if credit := convert.ImageCredit(a.Content, a.ImageURL); credit != "" {
		return credit
	}
	if src := sourceID(a.Link, a.FeedURL); src != "" {
		return "photo: " + src
	}
	return ""
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