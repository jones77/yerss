## 1. Fix the kitty delete-by-id sequence

- [x] 1.1 Change `image.DeleteByID` (internal/image/native.go) to emit the uppercase data-freeing form `\x1b_Ga=d,d=I,i=<id>,q=1\x1b\\` and update its doc comment to state it frees the terminal's cached image data.
- [x] 1.2 Verify the existing exit/scroll-out/resize tests in `internal/ui/article_image_test.go` still pass (they assert via the `DeleteByID` helper), and add a direct assertion that the emitted sequence uses `d=I`.

## 2. Ghostty delete-all fallback on exit

- [x] 2.1 Add env-derived Ghostty detection alongside `DetectProtocol` (protocol.go): `TERM_PROGRAM == "ghostty"` or `GHOSTTY_RESOURCES_DIR` set, exposed as a flag on the kitty-protocol path.
- [x] 2.2 In `computeNativeClear` (internal/ui/image.go), when the terminal is Ghostty and the frame leaves the article view (`view != viewArticle`), append a `d=a` delete-all clear after the per-id deletes so no placement survives even if the by-id delete is ignored.
- [x] 2.3 Ensure the existing `nativeClearArticleID` gate for no-native articles is unchanged and does not interfere with the Ghostty exit fallback.

## 3. Verification

- [x] 3.1 Add tests asserting: the exit frame on a Ghostty-identified terminal contains both the per-id `d=I` deletes and the `d=a` delete-all; the exit frame on a non-Ghostty kitty terminal contains the per-id deletes and no `d=a`.
- [x] 3.2 Run the full test suite (`go test ./...`) and the linter; confirm no regressions in native-clear, scroll-out, or resize tests.
- [x] 3.3 Manual check in Ghostty: open an image-heavy article, back out to the list, and confirm no photo lingers over the list view.