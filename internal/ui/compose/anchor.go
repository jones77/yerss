package compose

import (
	"strings"

	xansi "github.com/charmbracelet/x/ansi"
)

// AnchorableLine scans downward from oldOffset for the first content line whose
// text survives a link-mode change and returns that line's leading text, the
// number of blank separator lines skipped before it, and whether such a line was
// found. Blank lines are counted in gap and skipped; a line whose leading tokens
// are URL text (a link href the compact mode hides) is skipped too, anchoring on
// the first real word that follows. When no anchorable word exists below the
// offset, found is false and the caller glues to the bottom.
func AnchorableLine(lines []string, urls []string, oldOffset int) (anchor string, gap int, found bool) {
	for i := oldOffset; i < len(lines); i++ {
		visible := strings.TrimSpace(xansi.Strip(lines[i]))
		if visible == "" {
			gap++
			continue
		}
		text, ok := stripLeadingURLTokens(visible, urls)
		if !ok {
			continue
		}
		return text, gap, true
	}
	return "", gap, false
}

// stripLeadingURLTokens drops the leading URL tokens of a line's visible text,
// returning the remainder and whether any real (non-URL) word is left. A token
// is URL text when it carries a scheme, starts with www., or is a long-enough
// contiguous substring of a known link URL (a wrapped URL continuation such as
// on-of-narco-terrorist-refueling-vessel-aug-28-2026/). The length floor keeps
// short common words that happen to appear inside a URL from being mistaken for
// a fragment.
func stripLeadingURLTokens(visible string, urls []string) (string, bool) {
	fields := strings.Fields(visible)
	i := 0
	for i < len(fields) && isURLToken(fields[i], urls) {
		i++
	}
	if i == 0 {
		return visible, true
	}
	if i >= len(fields) {
		return "", false
	}
	return strings.Join(fields[i:], " "), true
}

// minURLFragmentLen is the shortest token that substring matching against a
// link URL will treat as a wrapped URL fragment; below it a token is assumed to
// be an ordinary word.
const minURLFragmentLen = 6

// isURLToken reports whether tok is a URL or a wrapped fragment of one of the
// article's harvested links.
func isURLToken(tok string, urls []string) bool {
	if strings.Contains(tok, "://") || strings.HasPrefix(tok, "www.") {
		return true
	}
	if len([]rune(tok)) < minURLFragmentLen {
		return false
	}
	for _, url := range urls {
		if url != "" && strings.Contains(url, tok) {
			return true
		}
	}
	return false
}

// FindAnchorLine locates the content line in lines whose visible text matches
// text. It scans in three passes: an exact trimmed match first; then the first
// line that starts with the text, so a wrapped link name whose line break
// shifted across modes still anchors on its leading text; then the line sharing
// the longest leading run with the text, so a line whose trailing words moved
// (a link grew visibly wider and pushed following text to a new line) still
// anchors on the paragraph that begins it. The first matching line wins within
// each pass, keeping the scan deterministic for repeated text.
func FindAnchorLine(lines []string, text string) (int, bool) {
	for i, ln := range lines {
		if strings.TrimSpace(xansi.Strip(ln)) == text {
			return i, true
		}
	}
	for i, ln := range lines {
		if strings.HasPrefix(strings.TrimSpace(xansi.Strip(ln)), text) {
			return i, true
		}
	}
	firstWord, _, _ := strings.Cut(text, " ")
	best, bestLen, found := 0, 0, false
	for i, ln := range lines {
		l := commonPrefix(strings.TrimSpace(xansi.Strip(ln)), text)
		if l > bestLen {
			best, bestLen, found = i, l, true
		}
	}
	if found && bestLen >= len(firstWord) && bestLen > 0 {
		return best, true
	}
	return 0, false
}

// commonPrefix returns the length of the longest common leading run of the two
// strings in runes.
func commonPrefix(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	n := len(ra)
	if len(rb) < n {
		n = len(rb)
	}
	for i := 0; i < n; i++ {
		if ra[i] != rb[i] {
			return i
		}
	}
	return n
}