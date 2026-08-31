## 1. Verify the pending smoke tests

- [x] 1.1 Re-run 9.1 on the Intercept article: images never overlap the border; space pages through text without skipping to images; j/k snap each inline image bottom-then-top and out; the TomDispatch logo caption reads its banner text with no leftover `](/tomdispatch)`; the links popup lists `https://theintercept.com/tomdispatch`; then mark 6.3 and 7.5 done in `inline-image-rendering/tasks.md`
- [x] 1.2 Re-run 10.3 on the Intercept article: j-scrolling shows each photo fully (no blank strip, no manual scrolling through empty rows); k-scrolling leaves each photo until it scrolls off in one snap
- [x] 1.3 Re-run 11.4 on the Intercept article: the bottom snap shows each photo with its caption at the bottom edge (one-line captions included); the two Fallujah paragraphs between the Trump's War thumbnail and the book cover read normally; the book cover's top shows a blocky preview rather than an empty strip before it snaps in
- [x] 1.4 Re-run 12.3 on the hurricane article from the bottom: scroll-k shows each gallery photo once (filling the screen) then scrolls it off to the previous image or content, with no stuck offset; scroll-j through the whole article stays clean
- [x] 1.5 Exercise the halfblock/ASCII paths (`-a` flag) and a resize during reading; the snap/preview logic is geometry-dependent

## 2. Archive inline-image-rendering

- [x] 2.1 Mark every completed smoke-test task `[x]` in `openspec/changes/inline-image-rendering/tasks.md` (6.3, 7.5, 9.1, 10.3, 11.4, 12.3)
- [x] 2.2 Run `openspec validate inline-image-rendering`
- [x] 2.3 Sync the change's delta specs into the main specs: `specs/inline-image-rendering/spec.md` → new `openspec/specs/inline-image-rendering/spec.md`, and `specs/reader-ui/spec.md` → `openspec/specs/reader-ui/spec.md` (which already carries the standard-caption-width changes); verify nothing is lost and `openspec validate --all` passes
- [x] 2.4 Move the change to `openspec/changes/archive/2026-08-31-inline-image-rendering/`

## 3. Commit the archive and working tree

- [x] 3.1 Stage explicitly (no `git add -A`): the archived `inline-image-rendering` dir, the synced `openspec/specs/` paths, and the shared implementation files — `internal/ui/{image,article,article_image_test,links_popup_test}.go`, `internal/image/{native_cmd,image_test,image,native}.go`, `convert/{convert,credit,convert_test,credit_test}.go`, and `AGENTS.md` — which also carry the `standard-caption-width` implementation
- [x] 3.2 Commit in one message following the seven rules; subject under 50 characters

## 4. Verification

- [x] 4.1 `go test ./...` green after any fix made during the smoke tests
- [x] 4.2 `openspec validate --all` green after the spec sync