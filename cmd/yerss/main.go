package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"yerss/internal/config"
	"yerss/internal/store"
	"yerss/internal/ui"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config.toml (default $XDG_CONFIG_HOME/yerss/config.toml)")
	flag.Parse()

	cfg, err := config.Load(configPath)
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