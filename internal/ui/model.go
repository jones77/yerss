package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal/app"
	"yerss/internal/config"
	"yerss/internal/feed"
	"yerss/internal/image"
	"yerss/internal/ui/articleview"
	"yerss/internal/ui/imageview"
	"yerss/internal/ui/listview"
	"yerss/internal/ui/popup"
	"yerss/internal/ui/render"
)

type viewState int

const (
	viewList viewState = iota
	viewArticle
)

type Model struct {
	sess *app.Session

	view viewState

	width  int
	height int

	list      listview.State
	article   articleview.State
	popupData popup.State
	// images owns the image-loading, block-composition, and native-image
	// sequencing state and methods (see the imageview package).
	images imageview.ImageView

	palette render.Palette
	ascii   bool

	// linkExpand toggles the article reader's link display mode: false (the
	// default) renders body hyperlinks compact (name only), true renders them
	// expanded (name + URL). It is session-scoped, persisting across article
	// switches, and keys the derivation cache so toggling recomposes. The
	// articleview.State mirrors it in its own LinkExpand field for derivation.
	linkExpand bool

	feedOutcomes []feed.FeedOutcome

	keys map[config.View]map[string]config.Action
}

// refreshFinishedMsg reports the outcome of a refresh pass.
type refreshFinishedMsg struct {
	result    *feed.FetchResult
	err       error
	fetchedAt time.Time
}

// New builds a Model bound to an application session.
func New(sess *app.Session) *Model {
	cfg := sess.Config()
	km, err := config.ParseKeybindings(cfg.Keybindings)
	if err != nil {
		km = cfg.Keybindings
	}
	palette := render.ResolvePalette(cfg.Display.Theme)
	m := &Model{
		sess:    sess,
		view:    viewList,
		palette: palette,
		ascii:   cfg.Display.Ascii || render.DetectAsciiNeeded(),
		width:   80,
		height:  24,
	}
	m.images = *imageview.New(sess, palette, m.ascii, 80, 24)
	m.article = articleview.NewState(sess, palette, m.ascii, &m.images, 80, 24)
	m.article.LinkExpand = m.linkExpand
	m.list = *listview.New(sess, palette, m.ascii, 80, 24)
	m.popupData = *popup.New(sess, palette, m.ascii, 80, 24)
	m.keys = make(map[config.View]map[string]config.Action)
	for _, v := range []config.View{config.ViewList, config.ViewArticle, config.ViewPopup} {
		if eff, err := km.EffectiveKeys(v); err == nil {
			m.keys[v] = eff
		}
	}
	return m
}

// SetAscii forces ASCII fallback glyph rendering, overriding the config setting
// and terminal detection. It takes effect on the next render.
func (m *Model) SetAscii(v bool) {
	m.ascii = v
	m.images.SetAscii(v)
	m.article.SetAscii(v)
}

// Ascii reports whether ASCII fallback glyphs are in use.
func (m *Model) Ascii() bool { return m.ascii }

// SetLinkExpand sets the session's link-expand choice, mirroring it into the
// article state so the next composition and frame render pick it up.
func (m *Model) SetLinkExpand(v bool) {
	m.linkExpand = v
	m.article.LinkExpand = v
}

// Init loads articles, restores the saved reader selection (firing the
// restored article's lead-image load), and kicks off a startup refresh when
// the gate allows. The persisted link-expand mode is applied before the
// selection is restored, so a reopened article composes in the saved mode.
func (m *Model) Init() tea.Cmd {
	if status := m.list.Load(); status != "" {
		m.setStatus(status)
	}
	if expanded, err := m.sess.LoadLinkExpandMode(); err == nil {
		m.SetLinkExpand(expanded)
	}
	imgLoad := m.restoreSelection()
	last, err := m.sess.LastRefreshedAt()
	if err != nil {
		m.setStatus("load error: " + err.Error())
		last = time.Time{}
	}
	m.list.Status.LastRefreshedAt = last
	var refresh tea.Cmd
	if m.sess.NeedsStartupRefresh(last, time.Now()) {
		refresh = m.refreshCmd()
	}
	return tea.Batch(imgLoad, refresh)
}

func (m *Model) refreshCmd() tea.Cmd {
	m.list.Status.Refreshing = true
	return func() tea.Msg {
		res, err := m.sess.Refresh()
		if err != nil {
			return refreshFinishedMsg{err: err}
		}
		return refreshFinishedMsg{result: res, fetchedAt: time.Now()}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	defer m.recomputeNativeClear()
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetSize(msg.Width, msg.Height)
		m.popupData.SetSize(msg.Width, msg.Height)
		m.images.SetSize(msg.Width, msg.Height)
		m.article.SetSize(msg.Width, msg.Height)
		var load tea.Cmd
		if m.view == viewArticle {
			offset := m.article.Viewport.YOffset
			a := *m.article.Article()
			m.article = m.article.Recompose(a)
			m.article.Viewport.SetYOffset(offset)
			load = m.images.EnsureImageSource(a, &m.article)
		}
		return m, load
	case tea.KeyMsg:
		if m.popupData.Kind() != popup.KindNone {
			return m.updatePopup(msg)
		}
		if m.view == viewList {
			return m.updateList(msg)
		}
		return m.updateArticle(msg)
	case tea.MouseMsg:
		if m.popupData.Kind() != popup.KindNone {
			return m, nil
		}
		var cmd tea.Cmd
		if m.view == viewList {
			cmd = m.updateListMouse(msg)
		} else {
			cmd = m.updateArticleMouse(msg)
		}
		return m, cmd
	case image.BlockMsg, image.PhotoMsg, image.NativeMsg, image.FailedMsg:
		return m, m.updateImageMessages(msg)
	case refreshFinishedMsg:
		m.list.Status.Refreshing = false
		if msg.err != nil {
			m.setStatus("refresh error: " + msg.err.Error())
		} else {
			m.list.Status.LastRefreshedAt = msg.fetchedAt
			m.setStatus(fmt.Sprintf("Refreshed: %d new, %d updated", msg.result.New, msg.result.Updated))
			// The latest pass's per-URL outcomes are what exit diagnostics
			// report; a failed pass leaves the previous outcomes standing.
			m.feedOutcomes = msg.result.Outcomes
		}
		if status := m.list.Load(); status != "" {
			m.setStatus(status)
		}
		return m, nil
	case articleview.URLActionMsg:
		if msg.Err != nil {
			m.setStatus(msg.Action + " failed: " + msg.Err.Error())
		} else {
			m.setStatus(msg.Label)
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) View() string {
	var s string
	switch m.view {
	case viewList:
		s = m.renderList()
	case viewArticle:
		s = m.article.Render()
	}
	if m.popupData.Kind() != popup.KindNone {
		s = render.Overlay(s, m.popupData.Render())
	}
	// The clear escape is zero-width and position-independent; prepending it
	// to the first line lets the frame repaint carry it to the terminal.
	return m.images.NativeClearPrefix() + s
}

// FeedOutcomes returns the per-URL outcomes of the most recent completed
// refresh pass, for exit diagnostics. It is empty until the first pass
// completes.
func (m *Model) FeedOutcomes() []feed.FeedOutcome {
	return m.feedOutcomes
}

func (m *Model) setStatus(msg string) {
	m.list.SetStatus(msg)
}

func (m *Model) refreshManual() tea.Cmd {
	if ok, remaining := m.sess.ManualRefreshAllowed(m.list.Status.LastRefreshedAt, time.Now()); !ok {
		secs := int((remaining + time.Second - 1) / time.Second)
		m.setStatus(fmt.Sprintf("next allowed in %ds", secs))
		return nil
	}
	return m.refreshCmd()
}
