package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

func withXDG(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))
	xdg.Reload()
	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadMissingConfigUsesDefaults(t *testing.T) {
	withXDG(t)
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Display.Theme != "auto" {
		t.Errorf("default theme = %q", cfg.Display.Theme)
	}
	if cfg.Display.PaddingX != 2 || cfg.Display.PaddingY != 1 {
		t.Errorf("default padding = %d/%d", cfg.Display.PaddingX, cfg.Display.PaddingY)
	}
	if cfg.MinInterval() != 15*60e9 {
		t.Errorf("default min interval = %v", cfg.MinInterval())
	}
	if cfg.Cooldown() != 60e9 {
		t.Errorf("default cooldown = %v", cfg.Cooldown())
	}
}

func TestLoadOverrides(t *testing.T) {
	withXDG(t)
	cfgDir := filepath.Join(t.TempDir())
	path := filepath.Join(cfgDir, "config.toml")
	writeFile(t, path, `
[data]
dir = "/tmp/custom-data"

[display]
theme = "light"
ascii = true
padding_x = 4
padding_y = 2
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DataDir() != "/tmp/custom-data" {
		t.Errorf("data dir = %q", cfg.DataDir())
	}
	if cfg.DBPath() != filepath.Join("/tmp/custom-data", "yerss.sqlite") {
		t.Errorf("db path = %q", cfg.DBPath())
	}
	if cfg.Display.Theme != "light" {
		t.Errorf("theme = %q", cfg.Display.Theme)
	}
	if !cfg.Display.Ascii {
		t.Error("ascii should be true")
	}
	if cfg.Display.PaddingX != 4 || cfg.Display.PaddingY != 2 {
		t.Errorf("padding = %d/%d", cfg.Display.PaddingX, cfg.Display.PaddingY)
	}
}

func TestCustomKeybinding(t *testing.T) {
	withXDG(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	writeFile(t, path, `
[keybindings]
quit = ["ctrl+x"]
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	eff, err := cfg.Keybindings.EffectiveKeys(ViewList)
	if err != nil {
		t.Fatal(err)
	}
	if eff["ctrl+x"] != Quit {
		t.Errorf("ctrl+x should map to quit, got %v", eff["ctrl+x"])
	}
	if eff["q"] == Quit {
		t.Error("custom binding should replace the default q binding")
	}
}

func TestConflictingKeybindingsRejected(t *testing.T) {
	withXDG(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	writeFile(t, path, `
[keybindings]
quit = ["q"]
open_article = ["q"]
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected conflict error, got nil")
	}
}

func TestDefaultKeybindingsValid(t *testing.T) {
	km := DefaultKeybindings()
	if _, err := ParseKeybindings(km); err != nil {
		t.Fatalf("defaults should validate: %v", err)
	}
	for _, view := range []View{ViewList, ViewArticle, ViewPopup} {
		if _, err := km.EffectiveKeys(view); err != nil {
			t.Errorf("view %s: %v", view, err)
		}
	}
}

func TestDefaultKeyAliases(t *testing.T) {
	km := DefaultKeybindings()
	listEff, err := km.EffectiveKeys(ViewList)
	if err != nil {
		t.Fatal(err)
	}
	articleEff, err := km.EffectiveKeys(ViewArticle)
	if err != nil {
		t.Fatal(err)
	}
	if listEff["o"] != OpenArticle {
		t.Errorf("o in list view should map to open_article, got %v", listEff["o"])
	}
	if articleEff["o"] != OpenURL {
		t.Errorf("o in article view should map to open_url, got %v", articleEff["o"])
	}
	for _, view := range []View{ViewList, ViewArticle} {
		eff, err := km.EffectiveKeys(view)
		if err != nil {
			t.Fatal(err)
		}
		if eff["b"] != Back {
			t.Errorf("b in view %s should map to back, got %v", view, eff["b"])
		}
	}
	// Existing aliases must remain unchanged.
	if listEff["enter"] != OpenArticle || listEff["l"] != OpenArticle {
		t.Errorf("enter/l in list should still open articles")
	}
	for _, view := range []View{ViewList, ViewArticle} {
		eff, err := km.EffectiveKeys(view)
		if err != nil {
			t.Fatal(err)
		}
		if eff["h"] != Back {
			t.Errorf("h in view %s should map to back, got %v", view, eff["h"])
		}
		if eff["esc"] != Back {
			t.Errorf("esc in view %s should map to back, got %v", view, eff["esc"])
		}
	}
}