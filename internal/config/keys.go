package config

import (
	"fmt"
	"strings"
)

// Action is a fixed set of semantic actions the user can bind keys to.
// Users may remap keys to actions but cannot define new actions.
type Action string

const (
	Quit            Action = "quit"
	Refresh         Action = "refresh"
	OpenArticle     Action = "open_article"
	Back            Action = "back"
	CloseArticle    Action = "close_article"
	LinkPopup       Action = "link_popup"
	MoveUp          Action = "move_up"
	MoveDown        Action = "move_down"
	PageUp          Action = "page_up"
	PageDown        Action = "page_down"
	HalfPageUp      Action = "half_page_up"
	HalfPageDown    Action = "half_page_down"
	Top             Action = "top"
	Bottom          Action = "bottom"
	TagPopup        Action = "tag_popup"
	OpenURL         Action = "open_url"
	CopyURL         Action = "copy_url"
	CopyArticleText Action = "copy_article_text"
	Help            Action = "help"
)

// View identifies a TUI view context for key dispatch and conflict
// validation. The same key may map to different actions in different views.
type View string

const (
	ViewList    View = "list"
	ViewArticle View = "article"
	ViewPopup   View = "popup"
)

// Keymap maps an action to its bound key strings (normalized).
type Keymap map[Action][]string

// actionSpec describes a single action: its identifier, default key strings,
// and the views it is valid in. catalog is the single source of truth for the
// action set, the runtime defaults, per-view key resolution, the help screen,
// and the seeded config template, so none of them can drift apart.
type actionSpec struct {
	action Action
	keys   []string
	views  []View
}

var catalog = []actionSpec{
	{Quit, []string{"q", "ctrl+c"}, []View{ViewList, ViewArticle, ViewPopup}},
	{Refresh, []string{"R", "r", "ctrl+r", "f5"}, []View{ViewList}},
	{OpenArticle, []string{"enter", "l", "o"}, []View{ViewList}},
	{Back, []string{"esc", "h", "b"}, []View{ViewList, ViewArticle, ViewPopup}},
	{CloseArticle, []string{"enter"}, []View{ViewArticle}},
	{LinkPopup, []string{"l", "right"}, []View{ViewArticle}},
	{MoveUp, []string{"up", "k"}, []View{ViewList, ViewArticle, ViewPopup}},
	{MoveDown, []string{"down", "j"}, []View{ViewList, ViewArticle, ViewPopup}},
	{PageUp, []string{"pgup", "ctrl+b"}, []View{ViewList, ViewArticle}},
	{PageDown, []string{"pgdn", "ctrl+f", "space"}, []View{ViewList, ViewArticle}},
	{HalfPageUp, []string{"ctrl+u"}, []View{ViewList, ViewArticle}},
	{HalfPageDown, []string{"ctrl+d"}, []View{ViewList, ViewArticle}},
	{Top, []string{"1", "g", "ctrl+up"}, []View{ViewList, ViewArticle}},
	{Bottom, []string{"G", "ctrl+down"}, []View{ViewList, ViewArticle}},
	{TagPopup, []string{"T", "t"}, []View{ViewList, ViewArticle}},
	{OpenURL, []string{"o"}, []View{ViewArticle, ViewPopup}},
	{CopyURL, []string{"c"}, []View{ViewArticle}},
	{CopyArticleText, []string{"C"}, []View{ViewArticle}},
	{Help, []string{"?"}, []View{ViewList, ViewArticle, ViewPopup}},
}

var allActions = func() map[Action]bool {
	m := make(map[Action]bool, len(catalog))
	for _, spec := range catalog {
		m[spec.action] = true
	}
	return m
}()

// DefaultKeybindings returns the built-in key map.
func DefaultKeybindings() Keymap {
	km := make(Keymap, len(catalog))
	for _, spec := range catalog {
		km[spec.action] = spec.keys
	}
	return km
}

// AllActions returns every bindable action in canonical display order. The
// help popup iterates it so the help screen always matches the catalog.
func AllActions() []Action {
	out := make([]Action, 0, len(catalog))
	for _, spec := range catalog {
		out = append(out, spec.action)
	}
	return out
}

// hasView reports whether views contains v.
func hasView(views []View, v View) bool {
	for _, w := range views {
		if w == v {
			return true
		}
	}
	return false
}

// actionsForView returns the actions valid in view v, in catalog order.
func actionsForView(v View) []Action {
	var out []Action
	for _, spec := range catalog {
		if hasView(spec.views, v) {
			out = append(out, spec.action)
		}
	}
	return out
}

// GlobalActions returns every action that is not specific to a single view, in
// catalog order: actions valid in all views (such as Quit, Back, Move Up, Move
// Down, Help) plus navigation shared by the list and article views (such as
// Page Up, Page Down, Top, Bottom, Tag Popup). The help popup renders these
// under its Global heading.
func GlobalActions() []Action {
	var out []Action
	for _, spec := range catalog {
		inList := hasView(spec.views, ViewList)
		inArticle := hasView(spec.views, ViewArticle)
		if (inList && !inArticle) || (inArticle && !inList) {
			continue
		}
		out = append(out, spec.action)
	}
	return out
}

// ListActions returns the actions valid only in the list view, in catalog
// order. The help popup renders these under its List view heading.
func ListActions() []Action {
	return exclusiveActions(ViewList)
}

// ArticleActions returns the actions valid only in the article view, in
// catalog order. The help popup renders these under its Article view heading.
func ArticleActions() []Action {
	return exclusiveActions(ViewArticle)
}

// exclusiveActions returns the actions declared valid in view v but not in the
// other reading view, in catalog order.
func exclusiveActions(v View) []Action {
	other := ViewList
	if v == ViewList {
		other = ViewArticle
	}
	var out []Action
	for _, spec := range catalog {
		if hasView(spec.views, v) && !hasView(spec.views, other) {
			out = append(out, spec.action)
		}
	}
	return out
}

// tomlKeyArray renders a key list as a TOML string-array literal, e.g.
// ["q", "ctrl+c"].
func tomlKeyArray(keys []string) string {
	quoted := make([]string, len(keys))
	for i, k := range keys {
		quoted[i] = `"` + k + `"`
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// seededKeybindings renders the commented-out [keybindings] section of the
// seeded config template from the catalog, so the template cannot drift from
// the runtime defaults.
func seededKeybindings() string {
	var b strings.Builder
	b.WriteString("# [keybindings]\n")
	for _, spec := range catalog {
		b.WriteString("# " + string(spec.action) + " = " + tomlKeyArray(spec.keys) + "\n")
	}
	return b.String()
}

// keybindingOptions returns the ConfigReleases entries for the keybinding
// actions, derived from the catalog so the self-update registry cannot drift
// from the runtime defaults.
func keybindingOptions() []ConfigOption {
	opts := make([]ConfigOption, 0, len(catalog))
	for _, spec := range catalog {
		opts = append(opts, ConfigOption{
			Version: 1,
			Section: "keybindings",
			Key:     string(spec.action),
			Default: tomlKeyArray(spec.keys),
		})
	}
	return opts
}

// normalizeKey canonicalizes key string spellings so config values match the
// strings reported by bubbletea. Letter case is preserved so that e.g. "R"
// and "r" remain distinct.
func normalizeKey(s string) string {
	s = strings.TrimSpace(s)
	switch strings.ToLower(strings.ReplaceAll(s, "-", "+")) {
	case "q":
		return s
	case "ctrl+c":
		return "ctrl+c"
	case "ctrl+r":
		return "ctrl+r"
	case "ctrl+f":
		return "ctrl+f"
	case "ctrl+b":
		return "ctrl+b"
	case "ctrl+d":
		return "ctrl+d"
	case "ctrl+u":
		return "ctrl+u"
	case "ctrl+up":
		return "ctrl+up"
	case "ctrl+down":
		return "ctrl+down"
	case "pgup", "pageup", "pg_up":
		return "pgup"
	case "pgdn", "pgdown", "pagedown", "pg_down":
		return "pgdown"
	case "space":
		return " "
	case "enter", "return":
		return "enter"
	case "esc", "escape":
		return "esc"
	case "up":
		return "up"
	case "down":
		return "down"
	case "left":
		return "left"
	case "right":
		return "right"
	case "f5":
		return "f5"
	}
	return s
}

// ParseKeybindings normalizes all key strings and validates that no key is
// bound to two different actions within the same view.
func ParseKeybindings(raw map[Action][]string) (Keymap, error) {
	km := make(Keymap, len(raw))
	for action, keys := range raw {
		if _, ok := allActions[action]; !ok {
			return nil, fmt.Errorf("unknown action %q", action)
		}
		norm := make([]string, 0, len(keys))
		for _, k := range keys {
			nk := normalizeKey(k)
			if nk == "" {
				return nil, fmt.Errorf("action %q has empty keybinding", action)
			}
			norm = append(norm, nk)
		}
		km[action] = norm
	}
	if err := Validate(km); err != nil {
		return nil, err
	}
	return km, nil
}

// Validate checks for conflicting keybindings across all views.
func Validate(km Keymap) error {
	for _, view := range []View{ViewList, ViewArticle, ViewPopup} {
		if _, err := km.EffectiveKeys(view); err != nil {
			return err
		}
	}
	return nil
}

// EffectiveKeys returns the key -> action mapping for a view. The Enter key
// bound to "back" is excluded in the list view where it opens articles, and
// the left arrow always clears the filter in the list view. A key assigned to
// two actions in the same view is reported as a conflict.
func (km Keymap) EffectiveKeys(v View) (map[string]Action, error) {
	m := map[string]Action{}
	for _, act := range actionsForView(v) {
		for _, key := range km[act] {
			if v == ViewList && act == Back && key == "enter" {
				continue
			}
			if prev, ok := m[key]; ok && prev != act {
				return nil, fmt.Errorf("key %q is bound to both %q and %q", key, prev, act)
			}
			m[key] = act
		}
	}
	if v == ViewList {
		if prev, ok := m["left"]; ok && prev != Back {
			return nil, fmt.Errorf("key %q is bound to both %q and %q", "left", prev, Back)
		}
		m["left"] = Back
	}
	return m, nil
}
