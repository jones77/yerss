package popup

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHelpPopupDismissesOnBAndH(t *testing.T) {
	for _, key := range []string{"b", "h"} {
		s, _ := newTestState(t)
		s.SetKind(KindHelp)
		act := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}, nil)
		if act != ActionBack {
			t.Errorf("%q: action = %v, want ActionBack", key, act)
		}
		if s.Kind() != KindNone {
			t.Errorf("%q: kind = %d, want KindNone", key, s.Kind())
		}
	}
}
