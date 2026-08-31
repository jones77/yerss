## 1. Scroll-skip of a fully visible photo (down)

- [x] 1.1 Extend `snapYOffset` in `internal/ui/image.go` with a downward branch: when a down move leaves the photo fully on screen (`yOffset <= imgStart && yOffset+vpH >= imgEnd`), snap to `capStart`; keep the existing land-interior snap and the upward reveal
- [x] 1.2 Update the `snapYOffset` doc comment to describe the fully-visible down skip
- [x] 1.3 Extend the `TestSnapYOffset` table: down from a fully visible photo skips to `capStart`; photo below the fold stays unchanged; page-below past the block stays unchanged

## 2. UI-level scroll skip test (down)

- [x] 2.1 Add `TestScrollSkipsFullyVisiblePhoto` in `internal/ui/article_image_test.go`: one downward scroll over a fully visible photo snaps to the caption, the next key scrolls normally, and page-down lands past the block without over-skipping

## 3. Upward caption scroll and native skip-to-top

- [x] 3.1 Add a `native bool` parameter to `snapYOffset` and narrow the up reveal range from `[imgStart, imgEnd]` to `[imgStart, capStart)` so an up move landing on a caption line is returned unchanged (caption scrolls up line by line)
- [x] 3.2 Add a native-only up rule: when `native && yOffset > 0 && yOffset <= imgStart && yOffset+vpH >= imgEnd`, return `0` (skip to article top)
- [x] 3.3 In `scrollArticle`, pass `m.article.nativeImg` as the `native` argument to `snapYOffset`
- [x] 3.4 Update the `snapYOffset` doc comment to describe the caption up-scroll and the native skip-to-top
- [x] 3.5 Update `TestSnapYOffset`: add a `native` column; change "up from caption first line" (`{8,...,-1}`) to want `8` and "up from caption last line" (`{9,...,-1}`) to want `9`; keep "up from photo interior"/"up onto first line" wanting `4`; add native `{2,20,4,9,8,-1}` → `0` and halfblock `{2,20,4,9,8,-1}` → `2`
- [x] 3.6 Extend `TestScrollSkipsFullyVisiblePhoto`: after the down skip, scroll up from below the block and confirm it lands on the caption line (`imgEnd`) not the photo, then reveal on the next up from a photo row; then set `m.article.nativeImg = true`, place the offset at the photo top, scroll up, and confirm the offset becomes `0`

## 4. Verification

- [x] 4.1 Run `gofmt`, `go vet ./...`, `go build ./...`, `go test ./...` and fix failures
- [x] 4.2 Run `openspec validate photo-scroll-skip`
