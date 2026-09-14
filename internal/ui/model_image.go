package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal/image"
)

// recomputeNativeClear keeps Update's deferred native-clear contract on Model;
// the computation itself lives on the article state, which reports whether the
// native snapshot changed so the frame cache can be invalidated.
func (m *Model) recomputeNativeClear() {
	m.article.RecomputeNativeClear(m.view == viewArticle)
}

// flushRecompose runs the pending article recompose, at most once, when the
// flag is set and the article view is still active, then fires the deferred
// native renders (and newly in-range loads) the recompose's fresh anchors
// brought into view. Image-load message handlers set recomposePending; Update
// calls flushRecompose after the message switch so the per-message work stays
// bounded to cache updates and the recompose — which reuses the cached
// derivation — runs once per image message. It returns the tea.Cmd firing that
// deferred work, or nil.
func (m *Model) flushRecompose() tea.Cmd {
	return m.article.FlushRecompose(m.view == viewArticle)
}

// recomposeArticle runs an in-place recompose of the open article, syncing the
// session's link-expand choice into the article state first.
func (m *Model) recomposeArticle() {
	m.article.LinkExpand = m.linkExpand
	m.article.RecomposeInPlace()
}

// updateImageMessages routes the image-load message types: each handler
// updates the caches and reports whether the open article needs recomposing,
// and the deferred flush runs the recompose at most once per message.
func (m *Model) updateImageMessages(msg tea.Msg) tea.Cmd {
	var needs bool
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case image.BlockMsg:
		needs = m.images.OnBlockLoaded(msg, &m.article, m.view == viewArticle)
	case image.PhotoMsg:
		needs, cmd = m.images.OnPhotoLoaded(msg, &m.article, m.view == viewArticle)
	case image.NativeMsg:
		needs = m.images.OnNativeLoaded(msg, &m.article, m.view == viewArticle)
	case image.FailedMsg:
		needs = m.images.OnImageFailed(msg, &m.article, m.view == viewArticle)
	default:
		return nil
	}
	if needs {
		m.article.SetRecomposePending(true)
	}
	return tea.Batch(cmd, m.article.FlushRecompose(m.view == viewArticle))
}