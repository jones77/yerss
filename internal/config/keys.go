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
	MoveUp          Action = "move_up"
	MoveDown        Action = "move_down"
	PageUp          Action = "page_up"
	PageDown        Action = "page_down"
	HalfPageUp      Action = "half_page_up"
	HalfPageDown    Action = "half_page_down"
	Top             Action = "top"
	Bottom          Action = "bottom"
	TagPopup        Action = "tag_popup"
	ToggleRead      Action = "toggle_read"
	MarkAllRead     Action = "mark_all_read"
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

// DefaultKeybindings returns the built-in key map.
func DefaultKeybindings() Keymap {
	return Keymap{
		Quit:            {"q", "ctrl+c"},
		Refresh:         {"R", "r", "ctrl+r", "f5"},
		OpenArticle:     {"enter", "l", "o"},
		Back:            {"esc", "enter", "h", "b"},
		MoveUp:          {"up", "k"},
		MoveDown:        {"down", "j"},
		PageUp:          {"pgup", "ctrl+b"},
		PageDown:        {"pgdn", "ctrl+f"},
		HalfPageUp:      {"ctrl+u"},
		HalfPageDown:    {"ctrl+d", "space"},
		Top:             {"g", "ctrl+up"},
		Bottom:          {"G", "ctrl+down"},
		TagPopup:        {"T", "t"},
		ToggleRead:      {"m"},
		MarkAllRead:     {"a"},
		OpenURL:         {"o"},
		CopyURL:         {"c"},
		CopyArticleText: {"C"},
		Help:            {"?"},
	}
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

var allActions = func() map[Action]bool {
	m := make(map[Action]bool)
	for _, a := range []Action{
		Quit, Refresh, OpenArticle, Back, MoveUp, MoveDown,
		PageUp, PageDown, HalfPageUp, HalfPageDown, Top, Bottom,
		TagPopup, ToggleRead, MarkAllRead, OpenURL, CopyURL, CopyArticleText, Help,
	} {
		m[a] = true
	}
	return m
}()

func actionsForView(v View) []Action {
	switch v {
	case ViewList:
		return []Action{Quit, Refresh, OpenArticle, Back, MoveUp, MoveDown, PageUp, PageDown, HalfPageUp, HalfPageDown, Top, Bottom, TagPopup, ToggleRead, MarkAllRead, Help}
	case ViewArticle:
		return []Action{Quit, Back, MoveUp, MoveDown, PageUp, PageDown, HalfPageUp, HalfPageDown, Top, Bottom, TagPopup, ToggleRead, OpenURL, CopyURL, CopyArticleText, Help}
	case ViewPopup:
		return []Action{Quit, MoveUp, MoveDown, Back, Help}
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
