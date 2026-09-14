package articleview

import (
	"net/url"
	"regexp"
	"strings"

	"yerss/internal/convert"
	"yerss/internal/store"
	"yerss/internal/ui/compose"
)

// Link is one hyperlink harvested from the article's markdown source.
type Link struct {
	Text string
	URL  string
}

// urlsOf returns the link URLs in order, for the compose anchor helpers that
// read only the url field.
func URLsOf(links []Link) []string {
	urls := make([]string, len(links))
	for i, l := range links {
		urls[i] = l.URL
	}
	return urls
}

// harvestLinks extracts the hyperlinks from the article's markdown source in
// document order, deduplicating by URL. It scans the `[text](url)` forms the
// converter emits, tolerating `<...>`-wrapped destinations.
func HarvestLinks(md string) []Link {
	re := regexp.MustCompile(`!?\[([^\]]*)\]\(([^)]+)\)`)
	seen := map[string]bool{}
	var out []Link
	for _, m := range re.FindAllStringSubmatch(md, -1) {
		dest := strings.TrimPrefix(strings.TrimSuffix(m[2], ">"), "<")
		if dest == "" || seen[dest] {
			continue
		}
		seen[dest] = true
		out = append(out, Link{Text: m[1], URL: dest})
	}
	return out
}

// harvestArticleLinks collects every hyperlink rendered in the article body —
// the markdown links, not the article's own header URL — deduplicated by URL,
// with relative destinations resolved to absolute against the article's origin
// so the links popup and open-url work. It scans the same marked markdown the
// body renders (ConvertImages), so a link that wraps an inline image (promo
// banners render as a linked image) is harvested like any other, with the
// image sentinel stripped from the link text.
func HarvestArticleLinks(a store.Article) []Link {
	seen := map[string]bool{}
	var out []Link
	add := func(text, url string) {
		if url == "" || seen[url] {
			return
		}
		seen[url] = true
		out = append(out, Link{Text: text, URL: url})
	}
	md, _ := convert.ConvertImages(a.Content)
	for _, l := range HarvestLinks(md) {
		add(compose.CleanLinkText(compose.SentinelRe.ReplaceAllString(l.Text, "")), ResolveLink(a, l.URL))
	}
	return out
}

// resolveLink resolves a link destination harvested from an article's HTML
// against the article's origin: a relative target like "/tomdispatch" becomes
// absolute (https://theintercept.com/tomdispatch) so opening it navigates to
// the right page. Absolute URLs pass through; targets with no resolvable base
// are returned unchanged.
func ResolveLink(a store.Article, href string) string {
	u, err := url.Parse(href)
	if err != nil || u.IsAbs() {
		return href
	}
	for _, base := range []string{a.Link, a.FeedURL} {
		bu, err := url.Parse(base)
		if err != nil || bu.Scheme == "" || bu.Host == "" {
			continue
		}
		return bu.ResolveReference(u).String()
	}
	return href
}