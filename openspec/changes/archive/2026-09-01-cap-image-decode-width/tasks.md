## 1. Decode cap helper

- [x] 1.1 Add `maxDecodeWidth = 2048` package constant in `internal/image`
- [x] 1.2 Add `decodeCapped(data []byte) (image.Image, error)` that decodes, and when the decoded width exceeds the cap, downscales to 2048 wide preserving aspect via the existing `scaleTo`; images at or below the cap return unchanged
- [x] 1.3 Unit-test `decodeCapped`: oversized image returns a bitmap 2048 wide with preserved aspect; at-or-below image returns unchanged; unparseable bytes error

## 2. Thread the cap through the decode sites

- [x] 2.1 Replace the raw `image.Decode` in `BlockCmd`'s stored-photo branch with `decodeCapped`, caching the capped image
- [x] 2.2 Replace the raw `image.Decode` in `BlockCmd`'s fetch branch with `decodeCapped`
- [x] 2.3 Replace the raw `image.Decode` in `PhotoCmd`'s placeholder branch with `decodeCapped`
- [x] 2.4 Replace the raw `image.Decode` in `renderNativeFitted` with `decodeCapped`
- [x] 2.5 Confirm `Photos` cache and `persistImage`/`SetArticleImage` are untouched: stored bytes remain the original compressed source

## 3. Validation

- [x] 3.1 Run `go test ./...` and `golangci-lint run` clean