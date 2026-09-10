## 1. Config plumbing

- [ ] 1.1 Add `ShowSeconds bool` to `config.Display` (default `false`) with `show_seconds` TOML parsing and validation falling back to `false` on unrecognized values
- [ ] 1.2 Add the `show_seconds` line to the seeded config template (`config/bootstrap.go`) and a config-release entry (`config/releases.go`) so existing configs self-update
- [ ] 1.3 Add/update config tests: default false, `true` honored, invalid value falls back to false, seeded template contains the key

## 2. Time layout selection

- [ ] 2.1 Add `LayoutTimeSeconds = "15:04:05"` and `TimeLayout(showSeconds)` / `DateTimeLayout(showSeconds)` helpers to `internal/timeutil`
- [ ] 2.2 Unit-test the helpers for both flag values

## 3. Apply to the views

- [ ] 3.1 List article-row time (`internal/ui/list.go:304`) uses `TimeLayout` from the display option
- [ ] 3.2 List status-bar last-refresh time (`internal/ui/list.go:360`) uses `TimeLayout` from the display option
- [ ] 3.3 Article top-border date (`internal/ui/article.go`) uses `DateTimeLayout` from the display option
- [ ] 3.4 Add tests: list row shows `HH:MM` by default and `HH:MM:SS` when enabled; status bar refresh time follows the option; article top-border date shows `HH:MM` by default and `HH:MM:SS` when enabled

## 4. Verification

- [ ] 4.1 Run `go test ./...` and the linter; confirm the existing article-row, status-bar, and top-border tests still pass
- [ ] 4.2 Exercise both modes in a terminal: default shows minute precision everywhere; `show_seconds = true` shows seconds in the list rows, list status bar, and article top border