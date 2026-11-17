package imageview

import (
	"yerss/internal"
	"yerss/internal/convert"
	"yerss/internal/ui/compose"
)

// ArticleView exposes the open article's render state that the image machinery
// reads and writes. imageview cannot depend on the article state type itself
// (articleview depends on imageview, so the dependency points the other way);
// the article state implements this interface and the Model/UI thread passes it
// to the image methods explicitly, so no image method reaches into a back
// reference for it.
type ArticleView interface {
	// ID is the open article's store id (0 when none is open).
	ID() int64
	// Article is the open article's full store row, or nil when none is open.
	Article() *internal.Article
	// Lines is the composed article content, one display row per entry.
	Lines() []string
	// ViewportY is the viewport's current top offset.
	ViewportY() int
	// ViewportHeight is the viewport's height in rows.
	ViewportHeight() int
	// ViewportWidth is the viewport's width in columns.
	ViewportWidth() int
	// SetViewportY moves the viewport's top offset (used by the scroll snap).
	SetViewportY(y int)
	// ImageBlocks is the current composition's image block spans.
	ImageBlocks() []compose.ImageBlock
	// NativeImg reports whether the lead block carries a native image escape.
	NativeImg() bool
	// LeadShown reports whether the first block was composed at the top of the
	// content (a genuine lead, shown on open).
	LeadShown() bool
	// FitHeaderLines is the mode-independent header budget the image-block
	// height fit reserves.
	FitHeaderLines() int
	// AnchorRows is the per-image anchor row list, parallel to ImageURLs.
	AnchorRows() []int
	// ImageURLs is the article's renderable image URLs in document order.
	ImageURLs() []string
	// InlineCaps maps each inline image URL to its composed caption.
	InlineCaps() map[string]string
	// InlineImages is the article's inline image list from the conversion.
	InlineImages() []convert.InlineImage
	// RendersImage reports whether url is one of the article's renderable
	// image URLs.
	RendersImage(url string) bool
	// URLIndex returns url's position in ImageURLs, or -1.
	URLIndex(url string) int
}
