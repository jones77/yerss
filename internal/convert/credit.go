package convert

import (
	"strings"

	"github.com/JohannesKaufmann/dom"
	"golang.org/x/net/html"
)

// ImageCaption returns the one-line caption attributed to imageURL from its
// HTML and whether imageURL sits inside a figure that carries a figcaption. A
// figure's caption belongs to its last <img> in document order, so a photo-grid
// (several photos under one combined caption) attributes the caption only to
// its final photo; an earlier image in the same figure gets an empty caption
// with matched=true, which the caller must honor by rendering no caption rather
// than falling back. The caption is collapsed to a single whitespace-separated
// line.
func ImageCaption(htmlStr, imageURL string) (string, bool) {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return "", false
	}
	return figureCaption(doc, imageURL)
}

// ImageCredit extracts a one-line photo credit for an image from its HTML: the
// figcaption of the figure containing an <img> whose src matches imageURL,
// attributed only when the image is the figure's last <img> (a figure's caption
// belongs to its last image). The caption is collapsed to a single
// whitespace-separated line. It returns "" when no figure matches, the matching
// figure has no caption, or the image is an earlier image in a shared captioned
// figure. It never borrows a caption from another image's figure: an image
// without its own caption renders without an attribution rather than displaying
// another image's (observed in the smoke test, where the first figure's caption
// attached to every image in the article).
func ImageCredit(htmlStr, imageURL string) string {
	cap, _ := ImageCaption(htmlStr, imageURL)
	return cap
}

// figureCaption returns the figcaption text attributed to imageURL and whether
// imageURL lies inside a captioned figure. The caption belongs to the figure's
// last <img>; an image earlier in a shared captioned figure returns "", true so
// the caller can render no caption rather than fall back.
func figureCaption(n *html.Node, imageURL string) (string, bool) {
	for _, fig := range findAll(n, "figure") {
		cap := textOf(findFirst(fig, "figcaption"))
		if cap == "" || imageURL == "" {
			continue
		}
		imgs := findAll(fig, "img")
		for i, img := range imgs {
			if v, ok := dom.GetAttribute(img, "src"); ok && v == imageURL {
				if i == len(imgs)-1 {
					return cap, true
				}
				return "", true
			}
		}
	}
	return "", false
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
