package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/pflag"

	"yerss/internal/config"
)

func TestResetDatabaseRemovesFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.sqlite")
	for _, p := range []string{path, path + "-wal", path + "-shm"} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := resetDatabase(path); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{path, path + "-wal", path + "-shm"} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s should be removed, stat err = %v", p, err)
		}
	}
}

func TestResetDatabaseMissingIsNoOp(t *testing.T) {
	if err := resetDatabase(filepath.Join(t.TempDir(), "absent.sqlite")); err != nil {
		t.Errorf("missing database should not error, got %v", err)
	}
}

// parseOptions registers the flag set the same way main does, parses args, and
// returns the parsed values so tests can exercise the -i flag without running
// main.
func parseOptions(t *testing.T, args ...string) *options {
	t.Helper()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.SetOutput(io.Discard)
	o := registerFlags(fs)
	if err := fs.Parse(args); err != nil {
		t.Fatalf("parse %v: %v", args, err)
	}
	return o
}

func TestInteractiveFlagParses(t *testing.T) {
	for _, args := range [][]string{{"-i"}, {"--interactive"}} {
		o := parseOptions(t, args...)
		if !o.interactive {
			t.Errorf("%v: expected interactive to be set", args)
		}
	}
}

func TestInteractiveNoOpElsewhere(t *testing.T) {
	// -i must not, by itself, enable any other flag or --init-db.
	o := parseOptions(t, "-i")
	if !o.interactive {
		t.Error("-i should set interactive")
	}
	for name, got := range map[string]bool{
		"json": o.jsonOut, "ascii": o.ascii, "editFeeds": o.editFeeds, "editConfig": o.editConfig, "initDB": o.initDB,
	} {
		if got {
			t.Errorf("-i alone should leave %s false", name)
		}
	}

	// Combining -i with the other flags must not change how those flags parse.
	base := parseOptions(t, "-j", "-a", "-e", "-c")
	withI := parseOptions(t, "-j", "-a", "-e", "-c", "-i")
	if base.jsonOut != withI.jsonOut || base.ascii != withI.ascii ||
		base.editFeeds != withI.editFeeds || base.editConfig != withI.editConfig ||
		base.initDB != withI.initDB {
		t.Errorf("-i must be a no-op for the other flags")
	}
	if !withI.interactive {
		t.Error("-i should be set when combined with other flags")
	}
}

func TestInteractiveDefaultsFalse(t *testing.T) {
	if o := parseOptions(t); o.interactive {
		t.Error("interactive must default to false")
	}
}

func TestUsageTextListsPathsWithoutDescriptions(t *testing.T) {
	out := usageText()
	for _, want := range []string{
		"config", "feeds", "db",
		config.DefaultConfigPath(),
		config.DefaultFeedsPath(),
		config.Default().DBPath(),
		"The config's [data] section can relocate the feeds file and database.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("usage should mention %q", want)
		}
	}
	for _, unwanted := range []string{"settings", "per line", "SQLite"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("usage should not describe files: found %q in %q", unwanted, out)
		}
	}
}
