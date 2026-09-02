## 1. Remove the --image-stats flag

- [x] 1.1 Delete `cmd/yerss/imagestats.go`
- [x] 1.2 In `cmd/yerss/main.go`, remove the `imageStats` option field, the `--image-stats` flag registration, and the `if imageStats` early-exit block
- [x] 1.3 Remove `ImageStats` and `ImageSize` types and `Store.ImageStats()` from `internal/store/store.go`, plus the now-unused `bytes`, `image`, and image-format blank imports
- [x] 1.4 Remove `pngBytes`, `TestImageStatsEmpty`, and `TestImageStatsAggregates` from `internal/store/store_test.go`

## 2. Remove the RAM status-bar readout

- [x] 2.1 Remove the `RAM <n>MB` line and update the surrounding comment in `renderStatusBar` (`internal/ui/list.go`)
- [x] 2.2 Remove `TotalBytes()`, `estBytes()`, and the `totalBytes` field from `internal/image/image.go`; simplify `Cache.Set` to stop tracking byte totals
- [x] 2.3 Remove `TestCacheTotalBytes` from `internal/image/image_test.go`
- [x] 2.4 Remove `TestStatusBarShowsRAM` and update the RAM mentions in `internal/ui/daygroups_test.go`

## 3. Validation

- [x] 3.1 Run `go test ./...` and `golangci-lint run` clean
- [x] 3.2 Grep the repo for `image-stats`, `ImageStats`, `RAM <n>`, and `TotalBytes` and confirm no stale references remain outside the archived changes
- [x] 3.3 Run `yerss --help` and confirm `--image-stats` no longer appears