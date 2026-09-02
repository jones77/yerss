## 1. Canonical source identity

- [x] 1.1 Add `CanonicalSource(url string) string` to `internal/ui/compose` that
      recovers the underlying source from a CDN fetch URL (decode a trailing
      URL-encoded `http(s)://` segment) and otherwise returns the URL unchanged
- [x] 1.2 Add unit tests for `CanonicalSource`: substack variant pairs resolve to
      one source; a plain URL is its own source; a non-CDN URL with no encoded
      segment passes through unchanged

## 2. Compose-time lead suppression

- [x] 2.1 In `newArticleState` (`internal/ui/article.go`), detect whether the
      lead URL's canonical source matches any inline image's canonical source
      and set a `suppressLead` flag
- [x] 2.2 When `suppressLead` is true, skip the top lead `composeImageBlock`
      call and build `imageURLs` from the deduped inline set without prepending
      `a.ImageURL`
- [x] 2.3 When `suppressLead` is true, run `DedupeInline("", inline)` and
      `StripDuplicateSentinels(bodyMD, "")` so the first body occurrence is kept
      and later repeats stripped; keep today's behavior when false

## 3. Attribution and native render

- [x] 3.1 In `nativeRenderCmd` (`internal/ui/image.go`), resolve caption via the
      inline attribution rules when the URL is one of the article's inline
      images, even when it equals the lead URL (i.e. when the lead is suppressed)

## 4. Tests

- [x] 4.1 Update `TestInlineImageLeadDuplicateSkipped` in
      `internal/ui/article_image_test.go` to the new semantics: a lead that
      duplicates a body image renders the body copy, not a top lead
- [x] 4.2 Add a test that the suppressed lead is not fetched/stored at position 0
      (photo bytes stored once at the body image's position)
- [x] 4.3 Add a compose test for canonical-source matching across differing
      substack-style transform URLs
- [x] 4.4 Run `go test ./...` and `go vet ./...` to confirm the change builds and
      passes
