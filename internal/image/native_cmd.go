// Native rendering is the most expensive path in the reader: a high-resolution
// photo is decoded, scaled to the block size, re-encoded as PNG, and base64
// encoded into an inline-image escape. Fire it off the UI goroutine so a frame
// never blocks on it, and cache the finished rendering by render size so
// re-viewing the article is instant.
package image

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

// NativeCmd returns a tea.Cmd that renders an article's lead photo to a native
// inline-image block off the UI goroutine, at the given content width and
// viewport-height cap. It reads the raw photo bytes from the Photos cache (the
// PhotoCmd fetch must precede it) unless the decoded image is already cached in
// cache (keyed by URL), runs the height-fit iteration to keep the block, its
// wrapped attribution, and one line of body text within the viewport budget,
// and caches the rendered lines in the Natives cache keyed by render size. On a
// cache hit it short-circuits to a NativeMsg immediately rather than
// re-rendering. On any failure it emits a FailedMsg so the halfblock block
// remains the final render. captionW is the standard caption width the
// attribution wraps at (independent of the photo's own width), so the fit
// reserves the same line count the UI composes with.
func NativeCmd(cache *Cache, native NativeRenderer, photos *Photos, natives *Natives, url string, width, maxHeight int, attr string, headerLines, captionW int) tea.Cmd {
	return func() tea.Msg {
		key := NativeKey(url, width, maxHeight)
		if lines, ok := natives.Get(key); ok {
			return NativeMsg{Key: url, Lines: lines}
		}
		data, ok := photos.Get(url)
		if !ok {
			return FailedMsg{Key: url}
		}
		lines, err := renderNativeFitted(native, cache, data, url, width, maxHeight, headerLines, attr, captionW)
		if err != nil {
			return FailedMsg{Key: url}
		}
		natives.Set(key, lines)
		return NativeMsg{Key: url, Lines: lines}
	}
}

// renderNativeFitted renders data to a native block at width, capping the
// block height so the block plus its wrapped attribution and one line of body
// text fits within maxHeight after the header, iterating up to three times to
// converge on the fitted height (an attribution that wraps wider than the
// reserved line shrinks the photo one more notch). The photo is decoded once
// and the decoded image is scaled and re-encoded for each candidate height, so
// the source bytes are never re-decoded per candidate. When the halfblock path
// already decoded the photo into cache the cached decoded image is reused and
// the source bytes are not decoded again; only on a cache miss are they
// decoded (via the same width cap the halfblock path applies). The attribution
// wraps at captionW, the standard caption width. It returns the fitted native
// lines (image rows only; the UI centers them and appends the attribution).
// Any render error propagates to the caller.
func renderNativeFitted(native NativeRenderer, cache *Cache, data []byte, url string, width, maxHeight, headerLines int, attr string, captionW int) ([]string, error) {
	src, ok := cache.Get(url)
	if !ok {
		var err error
		src, err = decodeCapped(data)
		if err != nil {
			return nil, err
		}
	}
	base := maxHeight - headerLines - 3
	maxH := max(1, base-1)
	if attr == "" {
		maxH = max(1, base)
	}
	if captionW < 1 {
		captionW = 1
	}
	var rendered []string
	for i := 0; ; i++ {
		lines, err := native.RenderImage(src, url, width, maxH)
		if err != nil {
			return nil, err
		}
		rendered = lines
		attrLines := 0
		if attr != "" {
			attrLines = len(LayoutWrapLines(attr, captionW))
		}
		want := max(1, base-attrLines)
		if want == maxH || i == 2 {
			break
		}
		maxH = want
	}
	return rendered, nil
}

// LayoutWrapLines wraps plain text to w display columns, breaking on spaces and
// hard-splitting any word longer than w. Non-empty input yields at least one
// line. It lives alongside the fit loop so the command and the UI's composeBlock
// agree on how many lines an attribution occupies.
func LayoutWrapLines(s string, w int) []string {
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
