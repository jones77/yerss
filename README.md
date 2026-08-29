# `yerss`

A Go terminal RSS reader written using gofeed, bubbletea/lipgloss and SQLite.

## Initial Configuration

Add at least one feed URL to `feeds.txt`;
the app won't start without a working feed.

### `feeds.txt`

Add URLS, one per line, to:

`$XDG_CONFIG_HOME/yerss/feeds.txt`;
fallback `~/.config/yerss/feeds.txt`

Blank lines and lines starting with `#` are ignored,
so you can comment out feeds.
The seeded file includes a commented-out example feed.

### `config.toml`

`$XDG_CONFIG_HOME/yerss/config.toml`;
fallback `~/.config/yerss/config.toml`

### Database

`$XDG_DATA_HOME/yerss/yerss.sqlite`;
fallback `~/.local/share/yerss/yerss.sqlite`

## Build

```sh
go build ./cmd/yerss
```

## Run

```sh
go run ./cmd/yerss
```

## Edit config and feeds

Open `feeds.txt` or `config.toml` in your editor (`$EDITOR`, falling back to
`$VISUAL`):

```sh
yerss -e    # or --edit-feeds: edit feeds.txt, then continue into the TUI
yerss -c    # or --edit-config: edit config.toml, then exit
```

`-e` honors a `[data] feeds_file` override from `config.toml` when set. After
`-e` the freshly-edited feeds are re-read and verified on startup. `-c` exits
after editing because config changes (theme, keybindings, refresh) take effect
on the next launch. If neither `$EDITOR` nor `$VISUAL` is set, yerss prints an
error and exits; there is no built-in fallback editor.
