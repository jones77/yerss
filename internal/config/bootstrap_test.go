package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBootstrapCreatesFiles(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	feedsPath := filepath.Join(dir, "feeds.txt")
	if err := Bootstrap(cfgPath, feedsPath); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	cfgData, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	cfgContent := string(cfgData)
	if !strings.Contains(cfgContent, fmt.Sprintf("# yerss config — schema v%d", ConfigSchemaVersion)) {
		t.Errorf("config missing schema header: %q", cfgContent[:80])
	}
	if !strings.Contains(cfgContent, fmt.Sprintf("(commit: %s)", BuildCommit)) {
		t.Error("config missing commit header")
	}
	if !strings.Contains(cfgContent, "# quit = ") {
		t.Error("config missing keybindings")
	}
	feedsData, err := os.ReadFile(feedsPath)
	if err != nil {
		t.Fatalf("read feeds: %v", err)
	}
	feedsContent := string(feedsData)
	if !strings.Contains(feedsContent, "one RSS URL per line") {
		t.Errorf("feeds.txt missing header: %q", feedsContent)
	}
	if !strings.Contains(feedsContent, "# https://www.dropsitenews.com/feed") {
		t.Errorf("feeds.txt missing commented-out example feed: %q", feedsContent)
	}
}

func TestBootstrapPreservesExisting(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	feedsPath := filepath.Join(dir, "feeds.txt")
	origConfig := "existing config content\n"
	origFeeds := "https://existing.example/feed.xml\n"
	if err := os.WriteFile(cfgPath, []byte(origConfig), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(feedsPath, []byte(origFeeds), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Bootstrap(cfgPath, feedsPath); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	gotCfg, _ := os.ReadFile(cfgPath)
	if string(gotCfg) != origConfig {
		t.Errorf("config was modified: %q", string(gotCfg))
	}
	gotFeeds, _ := os.ReadFile(feedsPath)
	if string(gotFeeds) != origFeeds {
		t.Errorf("feeds.txt was modified: %q", string(gotFeeds))
	}
}

func TestTemplateCoversAllReleaseOptions(t *testing.T) {
	tmpl := seededConfig()
	for _, opt := range ConfigReleases {
		if !strings.Contains(tmpl, opt.Key+" = ") {
			t.Errorf("template missing option %q (section %q, v%d)", opt.Key, opt.Section, opt.Version)
		}
	}
}

func TestMigrateFileAppendsNewOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	writeFile(t, path, "# yerss config — schema v0 (commit: old)\n\n# placeholder\n")
	if err := MigrateFile(path); err != nil {
		t.Fatalf("MigrateFile: %v", err)
	}
	s, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(s)
	if !strings.Contains(content, "# new in v1") {
		t.Errorf("missing banner: %q", content)
	}
	if !strings.Contains(content, "# dir = \"\"") {
		t.Errorf("missing appended option: %q", content)
	}
	if !strings.Contains(content, "# [keybindings]") {
		t.Errorf("missing appended section: %q", content)
	}
	if !strings.Contains(content, "# yerss config — schema v1") {
		t.Errorf("header not bumped: %q", content)
	}
}

func TestMigrateFileIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	writeFile(t, path, "# yerss config — schema v0 (commit: old)\n\n# x\n")
	if err := MigrateFile(path); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(path)
	if err := MigrateFile(path); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(path)
	if string(first) != string(second) {
		t.Errorf("second migrate changed the file:\n---\n%s\n---\n%s", first, second)
	}
}

func TestMigrateFileSkipsPresentOption(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	writeFile(t, path, "# yerss config — schema v0 (commit: old)\n\n[data]\ndir = \"/custom\"\n")
	if err := MigrateFile(path); err != nil {
		t.Fatalf("MigrateFile: %v", err)
	}
	s, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(s)
	if strings.Contains(content, "# dir = \"\"") {
		t.Errorf("dir was re-appended: %q", content)
	}
	if !strings.Contains(content, "dir = \"/custom\"") {
		t.Errorf("live dir lost: %q", content)
	}
	if !strings.Contains(content, "# feeds_file = \"\"") {
		t.Errorf("feeds_file should be appended: %q", content)
	}
}

func TestMigrateFileUntouchedWithoutHeader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	orig := "hello = \"world\"\n"
	writeFile(t, path, orig)
	if err := MigrateFile(path); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != orig {
		t.Errorf("file modified: %q", string(got))
	}
}

func TestMigrateFileCurrentVersionNoRewrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	orig := fmt.Sprintf("# yerss config — schema v%d (commit: dev)\n\n[data]\ndir = \"/custom\"\n", ConfigSchemaVersion)
	writeFile(t, path, orig)
	if err := MigrateFile(path); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != orig {
		t.Errorf("file rewritten: %q", string(got))
	}
}

func TestMigrateFileErrorOnBadPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub")
	writeFile(t, path, "x")
	if err := MigrateFile(filepath.Join(path, "config.toml")); err == nil {
		t.Fatal("expected error for path under a regular file")
	}
}

func TestLoadNonFatalWhenConfigDirUnwritable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root; permission checks are bypassed")
	}
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "yerss")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(cfgDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(cfgDir, 0o755) })

	cfg, err := Load(filepath.Join(cfgDir, "config.toml"))
	if err != nil {
		t.Fatalf("Load should be non-fatal: %v", err)
	}
	if cfg.Display.Theme != "auto" {
		t.Errorf("expected defaults, got theme %q", cfg.Display.Theme)
	}
}