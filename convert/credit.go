package convert

import (
	"strings"

	"github.com/JohannesKaufmann/dom"
	"golang.org/x/net/html"
)

// ImageCredit extracts a one-line photo credit for the article's lead image
// from its HTML: the figcaption of the figure containing an <img> whose src
// matches imageURL (or of the first figure with a figcaption when none
// matches), falling back to the text of the first element whose class
// indicates a credit line. The text is collapsed to a single
// whitespace-separated line. It returns "" when no credit is found.
func ImageCredit(htmlStr, imageURL string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return ""
	}
	if cap := figureCredit(doc, imageURL); cap != "" {
		return cap
	}
	return classCredit(doc)
}

// figureCredit returns the figcaption text of the figure whose <img> src
// matches imageURL, or of the first figure with a figcaption when no image
// matches.
func figureCredit(n *html.Node, imageURL string) string {
	first := ""
	for _, fig := range findAll(n, "figure") {
		cap := textOf(findFirst(fig, "figcaption"))
		if cap == "" {
			continue
		}
		if first == "" {
			first = cap
		}
		if imageURL != "" && hasImgWithSrc(fig, imageURL) {
			return cap
		}
	}
	return first
}

// classCredit returns the text of the first element whose class attribute
// contains "credit" (for example "photo-credit" or "image-credit").
func classCredit(n *html.Node) string {
	for _, el := range findAll(n, "") {
		if v, ok := dom.GetAttribute(el, "class"); ok && strings.Contains(v, "credit") {
			if text := textOf(el); text != "" {
				return text
			}
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
