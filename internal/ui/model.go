package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/glamour/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/config"
	"yerss/internal/feed"
	"yerss/internal/image"
	"yerss/internal/store"
)

type viewState int

const (
	viewList viewState = iota
	viewArticle
)

type popupMode int

const (
	noPopup popupMode = iota
	popupTagsList
	popupTagsArticle
	popupHelp
	popupLinks
)

type Model struct {
	cfg   *config.Config
	store *store.Store

	view  viewState
	popup popupMode

	width  int
	height int

	list      listState
	article   articleState
	popupData popupState

	palette Palette
	ascii   bool

	mdRenderer      *glamour.TermRenderer
	mdRendererW     int
	mdRendererStyle string

	imgRenderer image.Renderer
	imgCache    *image.Cache
	imgBlocks   *image.Blocks
	imgPhotos   *image.Photos
	imgNatives  *image.Natives
	imgNative   image.NativeRenderer
	imgLoading  map[string]bool
	// nativeSent tracks the kitty image id last transmitted per URL, so a
	// frame can delete the prior size's image before a re-render at a new
	// size, and leaving the article can free every image the terminal holds.
	nativeSent map[string]uint32

	statusMsg     string
	statusExpires time.Time

	lastRefreshedAt time.Time
	refreshing      bool

	feedOutcomes []feed.FeedOutcome

	dbSize int64

	keys map[config.View]map[string]config.Action
}

// refreshFinishedMsg reports the outcome of a refresh pass.
type refreshFinishedMsg struct {
	result    *feed.FetchResult
	err       error
	fetchedAt time.Time
}

// urlActionMsg reports the outcome of an open/copy command: opening an article
// URL, or copying text (a URL, the full article text, or a mouse selection) to
// the clipboard. label is the success status line; action names the operation
// for the failure status line.
type urlActionMsg struct {
	action string
	label  string
	err    error
}

// New builds a Model bound to a config and store.
func New(cfg *config.Config, st *store.Store) *Model {
	km, err := config.ParseKeybindings(cfg.Keybindings)
	if err != nil {
		km = cfg.Keybindings
	}
	m := &Model{
		cfg:         cfg,
		store:       st,
		view:        viewList,
		palette:     resolvePalette(cfg.Display.Theme),
		ascii:       cfg.Display.Ascii || detectAsciiNeeded(),
		imgRenderer: image.Halfblocks{},
		imgCache:    image.NewCache(),
		imgBlocks:   image.NewBlocks(),
		imgPhotos:   image.NewPhotos(),
		imgNatives:  image.NewNatives(),
		imgNative:   image.NativeRenderer{Protocol: image.DetectProtocol()},
		imgLoading:  make(map[string]bool),
		nativeSent:  make(map[string]uint32),
		width:       80,
		height:      24,
	}
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
func (m *Model) SetAscii(v bool) { m.ascii = v }

// Ascii reports whether ASCII fallback glyphs are in use.
func (m *Model) Ascii() bool { return m.ascii }

// Init loads articles, restores the saved reader selection (firing the
// restored article's lead-image load), and kicks off a startup refresh when
// the gate allows.
func (m *Model) Init() tea.Cmd {
	m.loadList()
	imgLoad := m.restoreSelection()
	last, err := m.store.LastRefreshedAt()
	if err != nil {
		m.setStatus("load error: " + err.Error())
		last = time.Time{}
	}
	m.lastRefreshedAt = last
	var refresh tea.Cmd
	if m.hasUnfetchedFeeds() {
		refresh = m.refreshCmd()
	} else {
		gate := feed.Gate{MinInterval: m.cfg.MinInterval(), Cooldown: m.cfg.Cooldown()}
		if gate.NeedsStartupRefresh(last, time.Now()) {
			refresh = m.refreshCmd()
		}
	}
	return tea.Batch(imgLoad, refresh)
}

// hasUnfetchedFeeds reports whether the configured feeds file contains a URL
// that has never been successfully fetched (no feeds row with a non-NULL
// last_fetched_at). It lets the startup refresh bypass the 15-minute gate so a
// feed newly added to feeds.txt is fetched without a manual refresh. A feeds
// file or store error is treated as "no new feeds" so startup falls through to
// the time gate; refreshCmd surfaces the error if a refresh still runs.
func (m *Model) hasUnfetchedFeeds() bool {
	urls, err := feed.LoadFeeds(m.cfg.FeedsFile())
	if err != nil {
		return false
	}
	verified, err := m.store.VerifiedFeedURLs()
	if err != nil {
		return false
	}
	for _, u := range urls {
		if !verified[feed.CanonicalURL(u)] {
			return true
		}
	}
	return false
}

func (m *Model) refreshCmd() tea.Cmd {
	m.refreshing = true
	return func() tea.Msg {
		urls, err := feed.LoadFeeds(m.cfg.FeedsFile())
		if err != nil {
			return refreshFinishedMsg{err: err}
		}
		res := feed.FetchFeeds(m.store, urls)
		return refreshFinishedMsg{result: &res, fetchedAt: time.Now()}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		var load tea.Cmd
		if m.view == viewArticle {
			offset := m.article.viewport.YOffset
			a := *m.article.article
			m.article = m.newArticleState(a)
			m.article.viewport.SetYOffset(offset)
			load = m.ensureImageSource(a)
		}
		return m, load
	case tea.KeyMsg:
		if m.popup != noPopup {
			return m.updatePopup(msg)
		}
		if m.view == viewList {
			return m.updateList(msg)
		}
		return m.updateArticle(msg)
	case tea.MouseMsg:
		if m.popup != noPopup {
			return m, nil
		}
		var cmd tea.Cmd
		if m.view == viewList {
			cmd = m.updateListMouse(msg)
		} else {
			cmd = m.updateArticleMouse(msg)
		}
		return m, cmd
	case image.BlockMsg:
		m.onBlockLoaded(msg)
		return m, nil
	case image.PhotoMsg:
		return m, m.onPhotoLoaded(msg)
	case image.NativeMsg:
		m.onNativeLoaded(msg)
		return m, nil
	case image.FailedMsg:
		m.onImageFailed(msg)
		return m, nil
	case refreshFinishedMsg:
		m.refreshing = false
		if msg.err != nil {
			m.setStatus("refresh error: " + msg.err.Error())
		} else {
			m.lastRefreshedAt = msg.fetchedAt
			m.setStatus(fmt.Sprintf("Refreshed: %d new, %d updated", msg.result.New, msg.result.Updated))
			// The latest pass's per-URL outcomes are what exit diagnostics
			// report; a failed pass leaves the previous outcomes standing.
			m.feedOutcomes = msg.result.Outcomes
		}
		m.loadList()
		return m, nil
	case urlActionMsg:
		if msg.err != nil {
			m.setStatus(msg.action + " failed: " + msg.err.Error())
		} else {
			m.setStatus(msg.label)
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
		s = m.renderArticle()
	}
	switch m.popup {
	case popupTagsList, popupTagsArticle:
		s = overlay(s, m.renderTagPopup())
	case popupLinks:
		s = overlay(s, m.renderLinksPopup())
	case popupHelp:
		s = overlay(s, m.renderHelp())
	}
	// The clear escape is zero-width and position-independent; prepending it
	// to the first line lets the frame repaint carry it to the terminal.
	return m.nativeImageClear() + s
}

// FeedOutcomes returns the per-URL outcomes of the most recent completed
// refresh pass, for exit diagnostics. It is empty until the first pass
// completes.
func (m *Model) FeedOutcomes() []feed.FeedOutcome {
	return m.feedOutcomes
}

func (m *Model) setStatus(msg string) {
	m.statusMsg = msg
	m.statusExpires = time.Now().Add(3 * time.Second)
}

func (m *Model) refreshManual() tea.Cmd {
	gate := feed.Gate{MinInterval: m.cfg.MinInterval(), Cooldown: m.cfg.Cooldown()}
	if ok, remaining := gate.ManualRefreshAllowed(m.lastRefreshedAt, time.Now()); !ok {
		secs := int((remaining + time.Second - 1) / time.Second)
		m.setStatus(fmt.Sprintf("next allowed in %ds", secs))
		return nil
	}
	return m.refreshCmd()
}

// overlay draws pop over base, centered by display width. Centering is computed
// from visible columns with ANSI escapes excluded, so styled popups are centered
// rather than glued to the left edge. Base content outside the popup region is
// preserved.
func overlay(base, pop string) string {
	baseLines := strings.Split(base, "\n")
	popLines := strings.Split(pop, "\n")

	bw := lipgloss.Width(base)
	bh := len(baseLines)
	pw := lipgloss.Width(pop)
	ph := len(popLines)
	if pw == 0 || ph == 0 {
		return base
	}
	if pw > bw {
		pw = bw
	}
	if ph > bh {
		ph = bh
		popLines = popLines[:bh]
	}

	bx := (bw - pw) / 2
	if bx < 0 {
		bx = 0
	}
	by := (bh - ph) / 2
	if by < 0 {
		by = 0
	}

	out := make([]string, len(baseLines))
	copy(out, baseLines)
	for i, l := range popLines {
		t := by + i
		if t < 0 || t >= len(out) {
			continue
		}
		out[t] = spliceStyled(out[t], l, bx, pw)
	}
	return strings.Join(out, "\n")
}

// spliceStyled overlays popRow onto baseRow so that popRow's visible content
// begins at display column bx. The base row's content before bx and after
// bx+pw is preserved, and ANSI escape sequences in both rows are kept intact.
func spliceStyled(baseRow, popRow string, bx, pw int) string {
	left := ansi.Truncate(baseRow, bx, "")
	if ansi.StringWidth(left) < bx {
		left = padRight(left, bx)
	}
	right := ansi.Cut(baseRow, bx+pw, ansi.StringWidth(baseRow))
	return left + popRow + right
}
