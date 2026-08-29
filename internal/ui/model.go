package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yerss/internal/config"
	"yerss/internal/feed"
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
)

type Model struct {
	cfg   *config.Config
	store *store.Store

	view  viewState
	popup popupMode

	width  int
	height int

	list     listState
	article  articleState
	popupData popupState

	palette Palette
	ascii   bool

	statusMsg     string
	statusExpires time.Time

	lastRefreshedAt time.Time
	refreshing      bool

	keys map[config.View]map[string]config.Action
}

// refreshFinishedMsg reports the outcome of a refresh pass.
type refreshFinishedMsg struct {
	result    *feed.FetchResult
	err       error
	fetchedAt time.Time
}

// urlActionMsg reports the outcome of an open/copy URL command.
type urlActionMsg struct {
	action string
	err    error
}

// New builds a Model bound to a config and store.
func New(cfg *config.Config, st *store.Store) *Model {
	km, err := config.ParseKeybindings(cfg.Keybindings)
	if err != nil {
		km = cfg.Keybindings
	}
	m := &Model{
		cfg:     cfg,
		store:   st,
		view:    viewList,
		palette: resolvePalette(cfg.Display.Theme),
		ascii:   cfg.Display.Ascii || detectAsciiNeeded(),
		width:   80,
		height:  24,
	}
	m.keys = make(map[config.View]map[string]config.Action)
	for _, v := range []config.View{config.ViewList, config.ViewArticle, config.ViewPopup} {
		if eff, err := km.EffectiveKeys(v); err == nil {
			m.keys[v] = eff
		}
	}
	return m
}

// Init loads articles and kicks off a startup refresh when the gate allows.
func (m *Model) Init() tea.Cmd {
	m.loadList()
	last, _ := m.store.LastRefreshedAt()
	m.lastRefreshedAt = last
	if m.hasUnfetchedFeeds() {
		return m.refreshCmd()
	}
	gate := feed.Gate{MinInterval: m.cfg.MinInterval(), Cooldown: m.cfg.Cooldown()}
	if gate.NeedsStartupRefresh(last, time.Now()) {
		return m.refreshCmd()
	}
	return nil
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
		if !verified[u] {
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
		if m.view == viewArticle {
			a := *m.article.article
			m.article = m.newArticleState(a)
		}
		return m, nil
	case tea.KeyMsg:
		if m.popup != noPopup {
			return m.updatePopup(msg)
		}
		if m.view == viewList {
			return m.updateList(msg)
		}
		return m.updateArticle(msg)
	case refreshFinishedMsg:
		m.refreshing = false
		if msg.err != nil {
			m.setStatus("refresh error: " + msg.err.Error())
		} else {
			m.lastRefreshedAt = msg.fetchedAt
			m.setStatus(fmt.Sprintf("Refreshed: %d new, %d updated", msg.result.New, msg.result.Updated))
		}
		m.loadList()
		return m, nil
	case urlActionMsg:
		if msg.err != nil {
			m.setStatus(msg.action + " failed: " + msg.err.Error())
		} else {
			m.setStatus(msg.action + "ed URL")
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
	case popupHelp:
		s = overlay(s, m.renderHelp())
	}
	return s
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
		left = padToWidth(left, bx)
	}
	right := ansi.Cut(baseRow, bx+pw, ansi.StringWidth(baseRow))
	return left + popRow + right
}

// padToWidth pads s with trailing spaces to at least width display columns,
// ignoring ANSI escape sequences.
func padToWidth(s string, width int) string {
	w := ansi.StringWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}