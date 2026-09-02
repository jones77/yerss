package compose

import (
	"net/url"
	"regexp"
	"strings"

	"yerss/internal/convert"
)

// SentinelRe matches the inline-image sentinel token "\x00img:<url>\x00" that
// ConvertImages emits for each <img>, capturing the image URL.
var SentinelRe = regexp.MustCompile(`\x00img:([^\x00]*)\x00`)

// NextSentinel returns the byte span [start, end) of the next inline-image
// sentinel at or after from, together with the image's URL and the text of any
// markdown link the sentinel opens (`<a><img>...text...</a>` renders as
// `[SENTINEL link text](href)`). When the sentinel is link-wrapped, the span is
// extended to consume the whole link construct so no stray `[`, link text, or
// `](...)` fragments render around the image block, and the link text is
// reported for use as the image's caption fallback. ok is false when no
// sentinel remains.
func NextSentinel(md string, from int) (start, end int, url, linkText string, ok bool) {
	loc := SentinelRe.FindStringSubmatchIndex(md[from:])
	if loc == nil {
		return 0, 0, "", "", false
	}
	base := from
	s := base + loc[0]
	e := base + loc[1]
	url = md[base+loc[2] : base+loc[3]]
	// A sentinel immediately preceded by '[' opens the markdown link the <a>
	// renderer produced; '![' (image syntax) is not a link and is left alone.
	if s > 0 && md[s-1] == '[' && (s < 2 || md[s-2] != '!') {
		s--
		if closeIdx := strings.Index(md[e:], "]("); closeIdx >= 0 {
			linkText = CleanLinkText(md[e : e+closeIdx])
			if n := LinkDestEnd(md[e+closeIdx+2:]); n >= 0 {
				e += closeIdx + 2 + n
			} else {
				e += closeIdx + 2
			}
		}
	} else if e < len(md) && strings.HasPrefix(md[e:], "](") {
		// A bare '](' immediately after the sentinel outside a link we opened:
		// consume it as a link tail for safety.
		if n := LinkDestEnd(md[e+2:]); n >= 0 {
			e += 2 + n
		}
	}
	return s, e, url, linkText, true
}

// CleanLinkText turns raw markdown link text (which can carry hard-break
// backslashes and emphasis markers when the converter wraps block text inside
// a link) into a single whitespace-separated caption line.
func CleanLinkText(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	s = strings.NewReplacer("\\", "", "**", "", "*", "", "`", "").Replace(s)
	return strings.TrimSpace(s)
}

// LinkDestEnd returns the length of a markdown link destination beginning at
// s, including its closing ')', or -1 when the destination is unterminated. It
// honors angle-bracketed destinations and balanced parentheses.
func LinkDestEnd(s string) int {
	if strings.HasPrefix(s, "<") {
		for i := 1; i < len(s); i++ {
			if s[i] == '>' {
				if i+1 < len(s) && s[i+1] == ')' {
					return i + 2
				}
				return -1
			}
		}
		return -1
	}
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			if depth == 0 {
				return i + 1
			}
			depth--
		}
	}
	return -1
}

// StripDuplicateSentinels removes every inline-image sentinel whose URL has
// already appeared (the lead image URL, or an earlier occurrence in document
// order), keeping the first occurrence of each image, along with the markdown
// link wrapping a removed sentinel. A skipped duplicate renders no block while
// the surrounding paragraphs keep their separation.
func StripDuplicateSentinels(md, leadURL string) string {
	seen := map[string]bool{}
	if leadURL != "" {
		seen[leadURL] = true
	}
	var b strings.Builder
	pos := 0
	for {
		start, end, url, _, ok := NextSentinel(md, pos)
		if !ok {
			b.WriteString(md[pos:])
			break
		}
		b.WriteString(md[pos:start])
		if !seen[url] {
			b.WriteString(md[start:end])
		}
		seen[url] = true
		pos = end
	}
	return b.String()
}

// DedupeInline drops inline images whose URL is the lead image URL or a URL
// already rendered, keeping the first occurrence in document order, so a photo
// attached as the lead and repeated in the body is not shown twice.
func DedupeInline(leadURL string, inline []convert.InlineImage) []convert.InlineImage {
	seen := map[string]bool{}
	if leadURL != "" {
		seen[leadURL] = true
	}
	kept := make([]convert.InlineImage, 0, len(inline))
	for _, im := range inline {
		if seen[im.URL] {
			continue
		}
		seen[im.URL] = true
		kept = append(kept, im)
	}
	return kept
}

// CanonicalSource recovers the underlying image source from a CDN fetch URL. A
// CDN fetch proxy URL carries a URL-encoded source segment in its path (for
// example substackcdn's ".../fetch/<params>/https%3A%2F%2F..."); the trailing
// encoded http(s) segment is decoded and returned as the canonical source, so
// two transform variants of the same photo resolve to one identity. A URL with
// no recognizable encoded source segment is returned unchanged, so plain image
// URLs keep their exact-URL identity.
func CanonicalSource(rawURL string) string {
	schemeIdx := -1
	for _, prefix := range []string{"https%3A%2F%2F", "http%3A%2F%2F"} {
		if i := strings.LastIndex(rawURL, prefix); i > schemeIdx {
			schemeIdx = i
		}
	}
	if schemeIdx < 0 {
		return rawURL
	}
	decoded, err := url.PathUnescape(rawURL[schemeIdx:])
	if err != nil || !strings.HasPrefix(decoded, "http://") && !strings.HasPrefix(decoded, "https://") {
		return rawURL
	}
	return decoded
}

// LineCountOf returns the number of lines the given parts occupy when joined
// with a newline separator.
func LineCountOf(parts []string) int {
	n := 0
	for _, p := range parts {
		n += len(strings.Split(p, "\n"))
	}
	return n
}

// HeaderLink renders a URL the markdown renderer prints exactly once. The
// glamour renderer emits `[text](url)` as "text url", so the URL is emitted
// bare and left to autolink detection; the angle-bracket form is used only
// when the URL contains characters that would break plain autolink parsing.
func HeaderLink(url string) string {
	if strings.ContainsAny(url, " <>") || strings.Contains(url, ")") {
		return "<" + url + ">"
	}
	return url
}
