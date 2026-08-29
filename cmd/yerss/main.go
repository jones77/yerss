package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/pflag"

	"yerss/internal/config"
	"yerss/internal/feed"
	"yerss/internal/store"
	"yerss/internal/ui"
)

func main() {
	base := filepath.Base(os.Args[0])

	var editFeeds, editConfig, jsonOut, ascii bool
	pflag.BoolVarP(&editFeeds, "edit-feeds", "e", false, "edit feeds.txt in $EDITOR")
	pflag.BoolVarP(&editConfig, "edit-config", "c", false, "edit config.toml in $EDITOR")
	pflag.BoolVarP(&jsonOut, "json", "j", false, "dump articles as JSON and exit")
	pflag.BoolVarP(&ascii, "ascii", "a", false, "force ASCII fallback glyphs")
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\nFlags:\n", base)
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Files:
  config  %s   display, refresh, and keybinding settings
  feeds   %s   one feed per line; the https:// scheme is assumed
  db      %s   SQLite article store

The config's [data] section can relocate the feeds file and database.
`, config.DefaultConfigPath(), config.DefaultFeedsPath(), config.Default().DBPath())
	}
	pflag.Parse()

	if code, done := runEditFlags(editConfig, editFeeds); done {
		os.Exit(code)
	}

	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", base, err)
		os.Exit(1)
	}

	st, err := store.Open(cfg.DBPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: cannot initialize database\n", base)
		fmt.Fprintf(os.Stderr, "database path: %s\n", cfg.DBPath())
		fmt.Fprintf(os.Stderr, "set [data] dir in your config to use a different location\n")
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	if jsonOut {
		os.Exit(runJSONDump(cfg, st))
	}

	if code := runStartupGate(cfg, st); code != 0 {
		os.Exit(code)
	}

	m := ui.New(cfg, st)
	applyAsciiFlag(m, ascii)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", base, err)
		os.Exit(1)
	}
	// The TUI has left the alternate screen; per-URL fetch diagnostics print
	// after exit so they never disturb the live display.
	feedDiagnostics(os.Stderr, base, m.FeedOutcomes())
}

// feedDiagnostics prints per-URL fetch diagnostics to stderr: an error line
// for each URL that failed to load (the HTTP status code when the server
// answered with one, the error text otherwise) and a warning line for each
// URL whose feed returned no articles. base is the program's basename.
func feedDiagnostics(w io.Writer, base string, outcomes []feed.FeedOutcome) {
	for _, o := range outcomes {
		switch {
		case o.Err != nil:
			if o.Status != 0 {
				fmt.Fprintf(w, "%s: error: %d, url: %s\n", base, o.Status, o.URL)
			} else {
				fmt.Fprintf(w, "%s: error: %v, url: %s\n", base, o.Err, o.URL)
			}
		case o.Articles == 0:
			fmt.Fprintf(w, "%s: warning: no articles returned, url: %s\n", base, o.URL)
		}
	}
}

// applyAsciiFlag forces ASCII fallback on the model when the -a flag is set,
// overriding the config setting and terminal detection for that run.
func applyAsciiFlag(m *ui.Model, force bool) {
	if force {
		m.SetAscii(true)
	}
}

// runEditFlags handles the -c/--edit-config and -e/--edit-feeds flags. It
// returns an exit code and done=true when the process should terminate after
// editing, or done=false to continue into normal startup. -c always exits
// after editing (config changes are read at load time and need a relaunch),
// covering both the -c-only and both-flags cases. -e alone continues into
// normal startup so the freshly-edited feeds are re-read and verified by the
// startup gate.
func runEditFlags(editConfig, editFeeds bool) (code int, done bool) {
	if editConfig {
		if code := runEditConfig(); code != 0 {
			return code, true
		}
	}
	if editFeeds {
		if code := runEditFeeds(); code != 0 {
			return code, true
		}
	}
	if editConfig {
		return 0, true
	}
	return 0, false
}