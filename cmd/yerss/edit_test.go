package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/adrg/xdg"
)

// withTempConfigHome points XDG_CONFIG_HOME at a fresh temp directory and
// refreshes the xdg package's cached base directories so the test controls
// where config.toml and feeds.txt resolve. It returns the temp directory.
func withTempConfigHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	xdg.Reload()
	return dir
}

// writeMarkerEditor writes a shell script that appends each of its arguments
// on its own line to $MARKER and exits with exitCode. It returns the script
// path, meant to be used as $EDITOR.
func writeMarkerEditor(t *testing.T, exitCode int) string {
	t.Helper()
	script := filepath.Join(t.TempDir(), "editor.sh")
	content := "#!/bin/sh\nprintf '%s\\n' \"$@\" >> \"$MARKER\"\nexit " + strconv.Itoa(exitCode) + "\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	return script
}

func setMarkerEnv(t *testing.T) string {
	t.Helper()
	marker := filepath.Join(t.TempDir(), "marker.txt")
	t.Setenv("MARKER", marker)
	return marker
}

func TestResolveEditorEditorWins(t *testing.T) {
	t.Setenv("EDITOR", "vim")
	t.Setenv("VISUAL", "nano")
	got, ok := resolveEditor()
	if !ok || got != "vim" {
		t.Fatalf("expected vim, got %q (ok=%v)", got, ok)
	}
}

func TestResolveEditorVisualFallback(t *testing.T) {
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "nano")
	got, ok := resolveEditor()
	if !ok || got != "nano" {
		t.Fatalf("expected nano, got %q (ok=%v)", got, ok)
	}
}

func TestResolveEditorNone(t *testing.T) {
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "")
	if _, ok := resolveEditor(); ok {
		t.Fatal("expected no editor when both variables are unset")
	}
}

func TestRunEditorNoEditorRefuses(t *testing.T) {
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "")
	withTempConfigHome(t)
	marker := setMarkerEnv(t)
	if code := runEditor(filepath.Join(t.TempDir(), "x.txt")); code != exitNoEditor {
		t.Fatalf("expected exit %d, got %d", exitNoEditor, code)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Error("editor must not run when neither $EDITOR nor $VISUAL is set")
	}
}

func TestRunEditorInvokesTargetPath(t *testing.T) {
	t.Setenv("EDITOR", writeMarkerEditor(t, 0))
	t.Setenv("VISUAL", "nano")
	marker := setMarkerEnv(t)
	target := filepath.Join(t.TempDir(), "feeds.txt")
	if code := runEditor(target); code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("read marker: %v", err)
	}
	if !strings.Contains(string(data), target) {
		t.Errorf("editor not invoked with %q; marker=%q", target, string(data))
	}
}

func TestRunEditorHonorsEditorArguments(t *testing.T) {
	t.Setenv("EDITOR", writeMarkerEditor(t, 0)+" --wait")
	marker := setMarkerEnv(t)
	target := filepath.Join(t.TempDir(), "feeds.txt")
	if code := runEditor(target); code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("read marker: %v", err)
	}
	out := string(data)
	if !strings.Contains(out, "--wait") || !strings.Contains(out, target) {
		t.Errorf("editor arguments not honored; marker=%q", out)
	}
}

func TestRunEditorPropagatesExitCode(t *testing.T) {
	t.Setenv("EDITOR", writeMarkerEditor(t, 7))
	setMarkerEnv(t)
	if code := runEditor(filepath.Join(t.TempDir(), "x.txt")); code != 7 {
		t.Fatalf("expected editor exit code 7, got %d", code)
	}
}

func TestEditFeedsHonorsFeedsFileOverride(t *testing.T) {
	dir := withTempConfigHome(t)
	custom := filepath.Join(dir, "custom-feeds.txt")
	cfgPath := filepath.Join(dir, "yerss", "config.toml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}
	cfgContent := "[data]\nfeeds_file = \"" + custom + "\"\n"
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EDITOR", writeMarkerEditor(t, 0))
	marker := setMarkerEnv(t)

	if code := runEditFeeds(); code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("read marker: %v", err)
	}
	if !strings.Contains(string(data), custom) {
		t.Errorf("editor not invoked with overridden path %q; marker=%q", custom, string(data))
	}
	if _, err := os.Stat(custom); err != nil {
		t.Errorf("overridden feeds file was not ensured to exist: %v", err)
	}
}

func TestEditFeedsFallsBackOnUnparseableConfig(t *testing.T) {
	dir := withTempConfigHome(t)
	cfgPath := filepath.Join(dir, "yerss", "config.toml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, []byte("[data\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EDITOR", writeMarkerEditor(t, 0))
	marker := setMarkerEnv(t)

	if code := runEditFeeds(); code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("read marker: %v", err)
	}
	want := filepath.Join(dir, "yerss", "feeds.txt")
	if !strings.Contains(string(data), want) {
		t.Errorf("editor not invoked with default path %q; marker=%q", want, string(data))
	}
}

func TestRunEditFlagsEditConfigExits(t *testing.T) {
	t.Setenv("EDITOR", writeMarkerEditor(t, 0))
	setMarkerEnv(t)
	code, done := runEditFlags(true, false)
	if !done || code != 0 {
		t.Fatalf("expected -c to exit 0, got code=%d done=%v", code, done)
	}
}

func TestRunEditFlagsBothExit(t *testing.T) {
	t.Setenv("EDITOR", writeMarkerEditor(t, 0))
	marker := setMarkerEnv(t)
	code, done := runEditFlags(true, true)
	if !done || code != 0 {
		t.Fatalf("expected both flags to exit 0, got code=%d done=%v", code, done)
	}
	// Both flags edit config first, then feeds: the marker records two
	// invocations (one per line).
	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("read marker: %v", err)
	}
	if lines := strings.Count(strings.TrimSpace(string(data)), "\n") + 1; lines != 2 {
		t.Errorf("expected two editor invocations, got %d", lines)
	}
}

func TestRunEditFlagsEditFeedsContinues(t *testing.T) {
	t.Setenv("EDITOR", writeMarkerEditor(t, 0))
	setMarkerEnv(t)
	code, done := runEditFlags(false, true)
	if done || code != 0 {
		t.Fatalf("expected -e to continue into startup, got code=%d done=%v", code, done)
	}
}