package articleview

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/config"
)

// Action is the article-view outcome of a key or mouse message that the
// Model-level wiring must handle (navigation away, popups, quit); scrolls,
// mark-read, URL opening, and text copying are handled inside the State.
type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionBack
	ActionOpenLinkPopup
	ActionOpenTagPopup
	ActionOpenURL
	ActionCopyURL
	ActionCopyText
	ActionCopyDebugInfo
	ActionToggleLinkExpand
	ActionHelp
)

// UpdateKey routes an article-view key message. Scrolls are applied to the
// viewport through the image machinery and downward moves mark the article
// read; actions that need Model-level state (popups, quit, back, link-expand
// toggle) are returned for the caller to dispatch. The returned tea.Cmd is the
// URL-open or copy command when the key maps to one.
func (s *State) UpdateKey(msg tea.KeyMsg, keys map[string]config.Action) (Action, tea.Cmd) {
	act, ok := keys[msg.String()]
	if !ok {
		return ActionNone, nil
	}
	var cmd tea.Cmd
	switch act {
	case config.Quit:
		return ActionQuit, nil
	case config.Back, config.CloseArticle:
		return ActionBack, nil
	case config.LinkPopup:
		return ActionOpenLinkPopup, nil
	case config.MoveDown:
		cmd = s.images.ScrollArticle(s, func() { s.Viewport.ScrollDown(1) }, true)
		s.MarkRead(s.sess.Store())
	case config.MoveUp:
		cmd = s.images.ScrollArticle(s, func() { s.Viewport.ScrollUp(1) }, true)
	case config.PageDown:
		cmd = s.images.ScrollArticle(s, func() { s.Viewport.PageDown() }, false)
		s.MarkRead(s.sess.Store())
	case config.PageUp:
		cmd = s.images.ScrollArticle(s, func() { s.Viewport.PageUp() }, false)
	case config.HalfPageDown:
		cmd = s.images.ScrollArticle(s, func() { s.Viewport.HalfPageDown() }, false)
		s.MarkRead(s.sess.Store())
	case config.HalfPageUp:
		cmd = s.images.ScrollArticle(s, func() { s.Viewport.HalfPageUp() }, false)
	case config.Top:
		cmd = s.images.ScrollArticle(s, func() { s.Viewport.GotoTop() }, false)
	case config.Bottom:
		cmd = s.images.ScrollArticle(s, func() { s.Viewport.GotoBottom() }, false)
	case config.OpenURL:
		if s.article.Link != "" {
			return ActionOpenURL, OpenURLCmd(s.article.Link)
		}
	case config.CopyURL:
		return ActionCopyURL, CopyURLCmd(s.article.Link)
	case config.CopyArticleText:
		return ActionCopyText, CopyTextCmd(ansi.Strip(strings.Join(s.lines, "\n")), "copied article text")
	case config.CopyDebugInfo:
		return ActionCopyDebugInfo, CopyDebugInfoCmd(s)
	case config.TagPopup:
		return ActionOpenTagPopup, nil
	case config.LinkExpandToggle:
		return ActionToggleLinkExpand, nil
	case config.Help:
		return ActionHelp, nil
	}
	return ActionNone, cmd
}

// UpdateMouse routes an article-view mouse message. Wheel scrolls the viewport
// through the image machinery (downward wheel marks read); a right press
// returns ActionBack for the Model to navigate back to the list; left press
// starts a text selection or opens a link under the cursor; drag and release
// extend and copy the selection. It returns a copy/open command when the
// action resolves to one.
func (s *State) UpdateMouse(msg tea.MouseMsg) (Action, tea.Cmd) {
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		cmd := s.images.ScrollArticle(s, func() { s.Viewport.ScrollUp(1) }, true)
		return ActionNone, cmd
	case tea.MouseButtonWheelDown:
		cmd := s.images.ScrollArticle(s, func() { s.Viewport.ScrollDown(1) }, true)
		s.MarkRead(s.sess.Store())
		return ActionNone, cmd
	}
	if msg.Shift || msg.Alt || msg.Ctrl {
		return ActionNone, nil
	}
	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonRight {
			return ActionBack, nil
		}
		if msg.Button != tea.MouseButtonLeft {
			return ActionNone, nil
		}
		c, ok := s.ContentCell(msg.X, msg.Y)
		if !ok {
			return ActionNone, nil
		}
		if url := s.LinkAtContentCell(c.X, c.Y); url != "" {
			return ActionNone, OpenURLCmd(url)
		}
		s.Sel = Selection{Active: true, Tracking: true, AnchorX: c.X, AnchorY: c.Y, CurX: c.X, CurY: c.Y}
	case tea.MouseActionMotion:
		if !s.Sel.Tracking {
			return ActionNone, nil
		}
		c := s.ContentCellClamped(msg.X, msg.Y)
		s.Sel.CurX, s.Sel.CurY = c.X, c.Y
	case tea.MouseActionRelease:
		if !s.Sel.Tracking {
			return ActionNone, nil
		}
		text := s.SelectedText()
		s.Sel = Selection{}
		if text != "" {
			return ActionNone, CopyTextCmd(text, "copied selection")
		}
	}
	return ActionNone, nil
}