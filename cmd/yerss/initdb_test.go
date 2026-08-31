package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"yerss/internal/store"
)

// stubStdin overrides the injectable stdin source and TTY detection for the
// duration of the test and restores them afterwards.
func stubStdin(t *testing.T, input string, tty bool) {
	t.Helper()
	oldReader, oldTTY := stdinReader, stdinIsTTY
	stdinReader = func() io.Reader { return strings.NewReader(input) }
	stdinIsTTY = func() bool { return tty }
	t.Cleanup(func() {
		stdinReader, stdinIsTTY = oldReader, oldTTY
	})
}

// captureStderr runs fn with stderr redirected to a pipe and returns the
// captured output.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stderr = old
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// writeDB creates a real, openable database at path and returns the open store.
func writeDB(t *testing.T, path string) *store.Store {
	t.Helper()
	st, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	return st
}

// insertArticle adds a minimal article with a unique GUID.
func insertArticle(t *testing.T, st *store.Store, guid string) {
	t.Helper()
	if _, err := st.UpsertArticle(store.Article{
		FeedURL:     "https://example.com/feed.xml",
		GUID:        guid,
		Title:       "t",
		Link:        "https://example.com/a",
		PublishedAt: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestFormatSize(t *testing.T) {
	for _, c := range []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1500000, "1.4 MB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	} {
		if got := formatSize(c.in); got != c.want {
			t.Errorf("formatSize(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestConfirmInitDBMissingFileNoPrompt(t *testing.T) {
	stubStdin(t, "", true)
	path := filepath.Join(t.TempDir(), "absent.sqlite")
	var out string
	out = captureStderr(t, func() {
		if code, ok := confirmInitDB(path); !ok || code != 0 {
			t.Fatalf("missing file should proceed without prompting, got code=%d ok=%v", code, ok)
		}
	})
	if out != "" {
		t.Errorf("missing file should print nothing, got %q", out)
	}
}

func TestConfirmInitDBTTYYes(t *testing.T) {
	stubStdin(t, "y\n", true)
	path := filepath.Join(t.TempDir(), "db.sqlite")
	st := writeDB(t, path)
	for _, g := range []string{"a", "b", "c"} {
		insertArticle(t, st, g)
	}
	t.Cleanup(func() { st.Close() })

	out := captureStderr(t, func() {
		if code, ok := confirmInitDB(path); !ok || code != 0 {
			t.Fatalf("expected -i confirm to proceed, got code=%d ok=%v", code, ok)
		}
	})
	if !strings.Contains(out, "[y/N]") {
		t.Errorf("missing [y/N] prompt suffix: %q", out)
	}
	if !strings.Contains(out, "3 articles") {
		t.Errorf("prompt should report the article count: %q", out)
	}
	if !strings.Contains(out, path) {
		t.Errorf("prompt should mention the database path: %q", out)
	}
	if !strings.Contains(out, "delete database (") {
		t.Errorf("prompt should open with the confirmation lead: %q", out)
	}
}

func TestConfirmInitDBDeclined(t *testing.T) {
	for _, input := range []string{"n\n", "\n", "garbage\n"} {
		stubStdin(t, input, true)
		path := filepath.Join(t.TempDir(), "db.sqlite")
		st := writeDB(t, path)
		code, ok := confirmInitDB(path)
		st.Close()
		if ok || code != 0 {
			t.Errorf("input %q: expected refusal with exit 0, got code=%d ok=%v", input, code, ok)
		}
		if _, err := os.Stat(path); err != nil {
			t.Errorf("input %q: database should be untouched, stat err=%v", input, err)
		}
	}
}

func TestConfirmInitDBNonTTY(t *testing.T) {
	stubStdin(t, "y\n", false)
	path := filepath.Join(t.TempDir(), "db.sqlite")
	st := writeDB(t, path)
	code, ok := confirmInitDB(path)
	st.Close()
	if ok || code != 1 {
		t.Fatalf("expected non-TTY refusal with exit 1, got code=%d ok=%v", code, ok)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("non-TTY refusal should leave database untouched, stat err=%v", err)
	}
}

func TestConfirmInitDBKnownSizePrompt(t *testing.T) {
	// A file of known size that is not a valid SQLite database exercises the
	// size formatting and the graceful count-on-failure path (count 0).
	stubStdin(t, "y\n", true)
	path := filepath.Join(t.TempDir(), "db.sqlite")
	if err := os.WriteFile(path, make([]byte, 1500000), 0o644); err != nil {
		t.Fatal(err)
	}
	out := captureStderr(t, func() {
		if code, ok := confirmInitDB(path); !ok || code != 0 {
			t.Fatalf("expected -i confirm to proceed, got code=%d ok=%v", code, ok)
		}
	})
	if !strings.Contains(out, "1.4 MB") {
		t.Errorf("prompt should format a 1500000-byte file as 1.4 MB: %q", out)
	}
	if !strings.Contains(out, "0 articles") {
		t.Errorf("prompt should report 0 articles for an unopenable file: %q", out)
	}
}
