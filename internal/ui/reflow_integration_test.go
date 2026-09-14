package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"yerss/internal/store"
)

// TestArticleReflowFixesStrippedLinkAndBlockquote reproduces the two reported
// layout defects at the width the user saw them: a blockquote whose closing
// curly quote strands off the quote bar, and a compact-mode paragraph left
// short where a wrapped URL's href was stripped.
func TestArticleReflowFixesStrippedLinkAndBlockquote(t *testing.T) {
	m, st := newTestModel(t)
	m.width = 76
	m.height = 22
	m.article.SetSize(m.width, m.height)
	content := `<figure class="wp-block-pullquote"><blockquote><p>“The ocean swallows up all the information we would usually rely on.”</p></blockquote></figure>` +
		`<p>Twenty-six of the 227 total killed have been identified. Among those identified by a&nbsp;<a href="https://www.elclip.org/guiria-venezuela-vueltas-narcotrafico-bombardeos-caribe/?lang=en">consortium of journalists</a>&nbsp;from Venezuelan outlets were a bus driver, a mototaxi driver, men involved in fishing or the seafood trade, and a futsal player.</p>`
	a := store.Article{
		FeedURL: "https://example.com/feed.xml",
		GUID:    "reflow-guid",
		Title:   "reflow",
		Content: content,
	}
	id, err := st.UpsertArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	a.ID = id
	m.article = m.article.Recompose(a)

	contentW := m.article.Viewport.Width
	flat := func(lines []string) string {
		vis := strings.ReplaceAll(ansi.Strip(strings.Join(lines, "\n")), "\u00a0", " ")
		return strings.Join(strings.Fields(vis), " ")
	}

	// Compact mode (the default): hrefs hidden and paragraphs re-flowed.
	body := flat(m.article.Lines())
	if strings.Contains(body, "elclip.org") {
		t.Errorf("compact: href text still visible:\n%s", body)
	}
	if !strings.Contains(body, "journalists from Venezuelan outlets were a bus driver") {
		t.Errorf("compact: paragraph not re-flowed after href strip:\n%s", body)
	}
	if !strings.Contains(body, "│ ”") {
		t.Errorf("compact: closing quote not kept on the quote bar:\n%s", body)
	}
	for i, line := range m.article.Lines() {
		vis := strings.TrimSpace(ansi.Strip(line))
		if vis == "”" {
			t.Errorf("compact: closing quote stranded without the quote bar at line %d: %q", i, vis)
		}
		if w := ansi.StringWidth(ansi.Strip(line)); w > contentW {
			t.Errorf("compact line %d exceeds content width: %d > %d: %q", i, w, contentW, vis)
		}
	}

	// Expanded mode: the URL is visible and lines still fit the content width.
	m.SetLinkExpand(true)
	m.article = m.article.Recompose(a)
	expanded := ansi.Strip(strings.Join(m.article.Lines(), "\n"))
	if !strings.Contains(expanded, "elclip.org") {
		t.Errorf("expanded: URL text not visible:\n%s", expanded)
	}
	for i, line := range m.article.Lines() {
		if w := ansi.StringWidth(ansi.Strip(line)); w > contentW {
			t.Errorf("expanded line %d exceeds content width: %d > %d: %q", i, w, contentW, ansi.Strip(line))
		}
	}
}
