package articleview

import (
	"strings"
	"time"

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
	ActionPreviousArticle
	ActionNextArticle
)

// UpdateKey routes an article-view key message. Scrolls and jumps are applied
// to the viewport through the image machinery and mark the article read (the
// first movement through a non-fitting article marks it); actions that need
// Model-level state (popups, quit, back, link-expand toggle) are returned for
// the caller to dispatch. The returned tea.Cmd is the URL-open or copy command
// when the key maps to one.
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
	case config.NextArticle:
		return ActionNextArticle, nil
	case config.PreviousArticle:
		return ActionPreviousArticle, nil
	case config.LinkPopup:
		return ActionOpenLinkPopup, nil
	case config.MoveDown:
		cmd = s.images.ScrollArticle(s, func() { s.Viewport.ScrollDown(1) }, true)
		s.MarkRead(s.sess.Store())
		cmd = tea.Batch(cmd, s.armNavHint())
	case config.MoveUp:
		cmd = s.images.ScrollArticle(s, func() { s.Viewport.ScrollUp(1) }, true)
		s.MarkRead(s.sess.Store())
		cmd = tea.Batch(cmd, s.armNavHint())
	case config.PageDown:
		cmd = s.images.ScrollArticle(s, func() { s.Viewport.PageDown() }, false)
		s.MarkRead(s.sess.Store())
		cmd = tea.Batch(cmd, s.armNavHint())
	case config.PageUp:
		cmd = s.images.ScrollArticle(s, func() { s.Viewport.PageUp() }, false)
		s.MarkRead(s.sess.Store())
		cmd = tea.Batch(cmd, s.armNavHint())
	case config.HalfPageDown:
		cmd = s.images.ScrollArticle(s, func() { s.Viewport.HalfPageDown() }, false)
		s.MarkRead(s.sess.Store())
		cmd = tea.Batch(cmd, s.armNavHint())
	case config.HalfPageUp:
		cmd = s.images.ScrollArticle(s, func() { s.Viewport.HalfPageUp() }, false)
		s.MarkRead(s.sess.Store())
		cmd = tea.Batch(cmd, s.armNavHint())
	case config.Top:
		cmd = s.images.ScrollArticle(s, func() { s.Viewport.GotoTop() }, false)
		s.MarkRead(s.sess.Store())
		cmd = tea.Batch(cmd, s.armNavHint())
	case config.Bottom:
		cmd = s.images.ScrollArticle(s, func() { s.Viewport.GotoBottom() }, false)
		s.MarkRead(s.sess.Store())
		cmd = tea.Batch(cmd, s.armNavHint())
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
// through the image machinery and mark the article read (in either direction);
// a right press
// returns ActionBack for the Model to navigate back to the list; left press
// starts a text selection or opens a link under the cursor; drag and release
// extend and copy the selection. It returns a copy/open command when the
// action resolves to one.
func (s *State) UpdateMouse(msg tea.MouseMsg) (Action, tea.Cmd) {
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		cmd := s.images.ScrollArticle(s, func() { s.Viewport.ScrollUp(1) }, true)
		s.MarkRead(s.sess.Store())
		return ActionNone, tea.Batch(cmd, s.armNavHint())
	case tea.MouseButtonWheelDown:
		cmd := s.images.ScrollArticle(s, func() { s.Viewport.ScrollDown(1) }, true)
		s.MarkRead(s.sess.Store())
		return ActionNone, tea.Batch(cmd, s.armNavHint())
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

// armNavHint sets the transient navigation hint from the current viewport
// position: "K: next" at the top edge, "J: previous" at the bottom edge, and no
// hint in between or on an article short enough to fit. A set hint returns a
// timer command that expires it.
func (s *State) armNavHint() tea.Cmd {
	total := s.Viewport.TotalLineCount()
	h := s.Viewport.Height
	if total <= h {
		s.navHint = navHintNone
		s.navHintExpires = time.Time{}
		return nil
	}
	maxOffset := total - h
	hint := navHintNone
	switch {
	case s.Viewport.YOffset <= 0:
		hint = navHintNext
	case s.Viewport.YOffset >= maxOffset:
		hint = navHintPrevious
	}
	if hint == navHintNone {
		s.navHint = navHintNone
		s.navHintExpires = time.Time{}
		return nil
	}
	s.navHint = hint
	s.navHintExpires = time.Now().Add(navHintDuration)
	return tea.Tick(navHintDuration, func(time.Time) tea.Msg { return NavHintExpiredMsg{} })
}
