package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
	if cfg.Display.Images != "auto" {
		t.Errorf("default images = %q, want auto", cfg.Display.Images)
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

func TestImagesOptionParsed(t *testing.T) {
	withXDG(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	writeFile(t, path, `
[display]
images = "off"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Display.Images != "off" {
		t.Errorf("images = %q, want off", cfg.Display.Images)
	}
}

func TestInvalidImagesFallsBackToAuto(t *testing.T) {
	withXDG(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	writeFile(t, path, `
[display]
images = "sometimes"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Display.Images != "auto" {
		t.Errorf("invalid images value = %q, want auto", cfg.Display.Images)
	}
}

func TestScrollbarOptionParsed(t *testing.T) {
	withXDG(t)
	if cfg, err := Load(""); err != nil || cfg.Display.Scrollbar != "single" {
		t.Errorf("default scrollbar = %q (err %v), want single", cfg.Display.Scrollbar, err)
	}
	path := filepath.Join(t.TempDir(), "config.toml")
	writeFile(t, path, `
[display]
scrollbar = "double"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Display.Scrollbar != "double" {
		t.Errorf("scrollbar = %q, want double", cfg.Display.Scrollbar)
	}
}

func TestInvalidScrollbarFallsBackToSingle(t *testing.T) {
	withXDG(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	writeFile(t, path, `
[display]
scrollbar = "fancy"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Display.Scrollbar != "single" {
		t.Errorf("invalid scrollbar value = %q, want single", cfg.Display.Scrollbar)
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

func TestDefaultKeybindingsMatchCatalog(t *testing.T) {
	want := Keymap{
		Quit:            {"q", "ctrl+c"},
		Refresh:         {"R", "r", "ctrl+r", "f5"},
		OpenArticle:     {"enter", "l", "o"},
		Back:            {"esc", "h", "b"},
		CloseArticle:    {"enter"},
		LinkPopup:       {"l", "right"},
		MoveUp:          {"up", "k"},
		MoveDown:        {"down", "j"},
		MoveLeft:        {"left", "h"},
		MoveRight:       {"right", "l"},
		PageUp:          {"pgup", "ctrl+b"},
		PageDown:        {"pgdn", "ctrl+f", "space"},
		HalfPageUp:      {"ctrl+u"},
		HalfPageDown:    {"ctrl+d"},
		Top:             {"1", "g", "ctrl+up"},
		Bottom:          {"G", "ctrl+down"},
		TagPopup:        {"T", "t"},
		ExpandToggle:    {"x"},
		OpenURL:         {"o"},
		CopyURL:         {"c"},
		CopyArticleText: {"C"},
		Help:            {"?"},
	}
	got := DefaultKeybindings()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DefaultKeybindings mismatch:\n got %#v\nwant %#v", got, want)
	}
}

func TestKeybindingCatalogConsistency(t *testing.T) {
	tmpl := seededConfig()
	for _, spec := range catalog {
		line := "# " + string(spec.action) + " = " + tomlKeyArray(spec.keys)
		if !strings.Contains(tmpl, line) {
			t.Errorf("seeded template missing %q", line)
		}
		if spec.action != Quit && len(spec.keys) == 0 {
			t.Errorf("action %q has no keys", spec.action)
		}
	}

	var kbOpts []ConfigOption
	for _, opt := range ConfigReleases {
		if opt.Section == "keybindings" {
			kbOpts = append(kbOpts, opt)
		}
	}
	if len(kbOpts) != len(catalog) {
		t.Fatalf("keybinding options = %d, want %d", len(kbOpts), len(catalog))
	}

	got := AllActions()
	if len(got) != len(catalog) {
		t.Fatalf("AllActions() = %d actions, want %d", len(got), len(catalog))
	}
	for i, spec := range catalog {
		if got[i] != spec.action {
			t.Errorf("AllActions()[%d] = %q, want %q", i, got[i], spec.action)
		}
	}
}

func TestHelpActionGroups(t *testing.T) {
	groups := map[string][]Action{
		"Global":    GlobalActions(),
		"List view": ListActions(),
		"Article view": ArticleActions(),
	}
	seen := map[Action]bool{}
	for label, actions := range groups {
		for _, a := range actions {
			if seen[a] {
				t.Errorf("action %q appears in more than one help group", a)
			}
			seen[a] = true
		}
		if len(actions) == 0 {
			t.Errorf("help group %q is empty", label)
		}
	}
	for _, a := range AllActions() {
		if !seen[a] {
			t.Errorf("action %q missing from every help group", a)
		}
	}
	want := map[string][]Action{
		"Global":       {Quit, Back, MoveUp, MoveDown, MoveLeft, MoveRight, PageUp, PageDown, HalfPageUp, HalfPageDown, Top, Bottom, TagPopup, Help},
		"List view":    {Refresh, OpenArticle, ExpandToggle},
		"Article view": {CloseArticle, LinkPopup, OpenURL, CopyURL, CopyArticleText},
	}
	for label, expected := range want {
		if !reflect.DeepEqual(groups[label], expected) {
			t.Errorf("%sActions() = %v, want %v", label, groups[label], expected)
		}
	}
}

func TestSeededTemplateIncludesAllAliases(t *testing.T) {
	tmpl := seededConfig()
	for _, want := range []string{
		`# back = ["esc", "h", "b"]`,
		`# close_article = ["enter"]`,
		`# open_article = ["enter", "l", "o"]`,
		`# tag_popup = ["T", "t"]`,
		`# refresh = ["R", "r", "ctrl+r", "f5"]`,
		`# copy_article_text = ["C"]`,
	} {
		if !strings.Contains(tmpl, want) {
			t.Errorf("seeded template missing %q", want)
		}
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

func TestDefaultKeybindingsBothCases(t *testing.T) {
	km := DefaultKeybindings()
	listEff, err := km.EffectiveKeys(ViewList)
	if err != nil {
		t.Fatal(err)
	}
	if listEff["t"] != TagPopup || listEff["T"] != TagPopup {
		t.Errorf("both t and T should map to tag_popup, got %v / %v", listEff["t"], listEff["T"])
	}
	if listEff["r"] != Refresh || listEff["R"] != Refresh {
		t.Errorf("both r and R should map to refresh, got %v / %v", listEff["r"], listEff["R"])
	}
	if listEff["g"] != Top || listEff["G"] != Bottom {
		t.Errorf("g should stay top and G bottom, got %v / %v", listEff["g"], listEff["G"])
	}
}

func TestPopupViewGridKeys(t *testing.T) {
	km := DefaultKeybindings()
	eff, err := km.EffectiveKeys(ViewPopup)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"up", "down"} {
		if eff[key] != MoveUp && eff[key] != MoveDown {
			t.Errorf("key %q in popup view should map to a move action, got %v", key, eff[key])
		}
	}
	if eff["left"] != MoveLeft {
		t.Errorf("left in popup view should map to move_left, got %v", eff["left"])
	}
	if eff["h"] != MoveLeft {
		t.Errorf("h in popup view should map to move_left, got %v", eff["h"])
	}
	if eff["right"] != MoveRight {
		t.Errorf("right in popup view should map to move_right, got %v", eff["right"])
	}
	if eff["l"] != MoveRight {
		t.Errorf("l in popup view should map to move_right, got %v", eff["l"])
	}
	if eff["esc"] != Back || eff["b"] != Back {
		t.Errorf("esc and b in popup view should still map to back, got %v / %v", eff["esc"], eff["b"])
	}
	for _, key := range []string{"1", "g", "ctrl+up"} {
		if eff[key] != Top {
			t.Errorf("%q in popup view should map to top, got %v", key, eff[key])
		}
	}
	for _, key := range []string{"G", "ctrl+down"} {
		if eff[key] != Bottom {
			t.Errorf("%q in popup view should map to bottom, got %v", key, eff[key])
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
	if articleEff["enter"] != CloseArticle {
		t.Errorf("enter in article view should map to close_article, got %v", articleEff["enter"])
	}
	if articleEff["c"] != CopyURL {
		t.Errorf("c in article view should map to copy_url, got %v", articleEff["c"])
	}
	if articleEff["C"] != CopyArticleText {
		t.Errorf("C in article view should map to copy_article_text, got %v", articleEff["C"])
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
