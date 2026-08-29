# yerss

A terminal RSS reader written in Go: a bubbletea/lipgloss TUI, SQLite for
storage, and gofeed for feed parsing.

## Build

```sh
go build ./cmd/yerss
```

## Run

```sh
go run ./cmd/yerss
```

Note: run the `cmd/yerss` package (as above), not a single file. `go run
cmd/yerss/main.go` will not compile — `main.go` depends on sibling files in the
same `package main` (for example `gate.go`), and single-file mode only compiles
the one file.

Pass `-config <path>` to use a config file other than the default.

## Data locations

- Config: `$XDG_CONFIG_HOME/yerss/config.toml` (fallback `~/.config/yerss/config.toml`)
- Feeds: one URL per line in `$XDG_CONFIG_HOME/yerss/feeds.txt` (fallback `~/.config/yerss/feeds.txt`)
- Database: `$XDG_DATA_HOME/yerss/yerss.sqlite` (fallback `~/.local/share/yerss/yerss.sqlite`)

On first run the config directory, a commented `config.toml`, and `feeds.txt` are
created for you. Add at least one feed URL to `feeds.txt`; the app verifies a live
feed before starting the TUI.