package convert

import (
	"strings"

	"github.com/JohannesKaufmann/dom"
	"golang.org/x/net/html"
)

// ImageCredit extracts a one-line photo credit for an image from its HTML: the
// figcaption of the figure containing an <img> whose src matches imageURL. The
// caption is collapsed to a single whitespace-separated line. It returns ""
// when no figure matches or the matching figure has no caption. It never
// borrows a caption from another image's figure: an image without its own
// figure renders without an attribution rather than displaying another image's
// (observed in the smoke test, where the first figure's caption attached to
// every image in the article).
func ImageCredit(htmlStr, imageURL string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return ""
	}
	return figureCredit(doc, imageURL)
}

// figureCredit returns the figcaption text of the figure whose <img> src
// matches imageURL, or "" when no figure matches or it has no caption.
func figureCredit(n *html.Node, imageURL string) string {
	for _, fig := range findAll(n, "figure") {
		cap := textOf(findFirst(fig, "figcaption"))
		if cap == "" {
			continue
		}
		if imageURL != "" && hasImgWithSrc(fig, imageURL) {
			return cap
		}
	}
	return ""
}

// findAll collects every element in document order, matching tag when it is
// non-empty and every element otherwise.
func findAll(n *html.Node, tag string) []*html.Node {
	var out []*html.Node
	var walk func(*html.Node)
	walk = func(cur *html.Node) {
		if cur.Type == html.ElementNode && (tag == "" || cur.Data == tag) {
			out = append(out, cur)
		}
		for c := cur.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return out
}

// findFirst returns the first descendant element with the given tag.
func findFirst(n *html.Node, tag string) *html.Node {
	for cur := n.FirstChild; cur != nil; cur = cur.NextSibling {
		if cur.Type == html.ElementNode && cur.Data == tag {
			return cur
		}
		if found := findFirst(cur, tag); found != nil {
			return found
		}
	}
	return nil
}

// hasImgWithSrc reports whether the subtree rooted at n contains an <img>
// whose src attribute equals src.
func hasImgWithSrc(n *html.Node, src string) bool {
	for _, img := range findAll(n, "img") {
		if v, ok := dom.GetAttribute(img, "src"); ok && v == src {
			return true
		}
	}
	return false
}

// textOf collects the text content of the subtree rooted at n, collapsed to a
// single whitespace-separated line.
func textOf(n *html.Node) string {
	if n == nil {
		return ""
	}
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(cur *html.Node) {
		if cur.Type == html.TextNode {
			b.WriteString(cur.Data)
			return
		}
		for c := cur.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}
