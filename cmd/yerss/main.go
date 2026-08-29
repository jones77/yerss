package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/pflag"

	"yerss/internal/config"
	"yerss/internal/store"
	"yerss/internal/ui"
)

func main() {
	var editFeeds, editConfig, jsonOut bool
	pflag.BoolVarP(&editFeeds, "edit-feeds", "e", false, "edit feeds.txt in $EDITOR")
	pflag.BoolVarP(&editConfig, "edit-config", "c", false, "edit config.toml in $EDITOR")
	pflag.BoolVarP(&jsonOut, "json", "j", false, "dump articles as JSON and exit")
	pflag.Parse()

	if code, done := runEditFlags(editConfig, editFeeds); done {
		os.Exit(code)
	}

	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "yerss: %v\n", err)
		os.Exit(1)
	}

	st, err := store.Open(cfg.DBPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "yerss: cannot initialize database\n")
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
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "yerss: %v\n", err)
		os.Exit(1)
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