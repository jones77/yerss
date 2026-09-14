package articleview

import (
	"path/filepath"
	"testing"

	"yerss/internal/app"
	"yerss/internal/config"
	"yerss/internal/store"
	"yerss/internal/ui/imageview"
	"yerss/internal/ui/render"
)

// newTestState builds an article reader State bound to a scratch store, with a
// real image machinery so Recompose can drive block composition.
func newTestState(t *testing.T) (State, *store.Store) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	sess := app.New(config.Default(), st)
	iv := imageview.New(sess, render.DarkPalette(), false, 80, 24)
	return NewState(sess, render.DarkPalette(), false, iv, 80, 24), st
}