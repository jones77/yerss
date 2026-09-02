## Context

The reader composes an open article in `internal/ui/article.go`
(`newArticleState`): it renders the header, then the lead image as a top block
(`article.go:173-177`) when `a.ImageURL != ""`, then the body via
`convert.ConvertImages`, with inline images interleaved in place. Dedup today
lives in `internal/ui/compose/sentinel.go`: `DedupeInline` and
`StripDuplicateSentinels` seed `seen[leadURL]` and drop every inline occurrence
whose URL equals the lead URL, so the lead always shows at top and the body copy
is hidden. Storage in `article_images` is keyed by `(article_id, position)` but
**read by URL** (`internal/image/image.go` `BlockCmd`/`PhotoCmd` call
`GetArticleImages` and match `imgs[i].URL != url`), so rendering does not depend
on which position a photo occupies.

The motivating case (Drop Site News) fails exact-URL dedup because the enclosure
and body image are different substackcdn transform URLs of the same source; on
disk they appear as two rows with identical bytes (position 0 + position 1), and
for a large source the same photo is fetched at two different widths with
different bytes. See `proposal.md` — Why.

## Goals / Non-Goals

**Goals:**

- Suppress the top lead block when it duplicates a body image (canonical source
  match), showing the body copy once in place.
- Reuse the existing dedup machinery, inverting its direction for the
  duplicated case.
- Avoid fetching and storing the suppressed lead a second time.

**Non-Goals:**

- No change to how the enclosure/lead URL is extracted at feed time
  (`feed-pipeline`'s `leadImageURL` stays as is).
- No rewrite of already-stored rows (stale position-0 rows are tolerated; reads
  are by URL).
- No new persistence schema.

## Decisions

**D1 — Canonical source identity key.** Add a function (in `internal/ui/compose`,
e.g. `CanonicalSource(url string) string`) that recovers the underlying image
source from a CDN fetch URL. Generic rule: if the URL is a CDN fetch proxy whose
path carries a URL-encoded `http(s)://...` trailing segment, decode that segment
as the canonical source; otherwise return the URL unchanged. This covers
substackcdn and similar image-CDN proxies. Alternative (byte identity) rejected:
the same photo can have different bytes at different widths (observed: article
151 full-res 1.56 MB vs `w_2400` 288 KB). Alternative (exact URL) rejected: does
not match dropsite.

**D2 — Invert dedup when the lead is duplicated.** In `newArticleState`, compute
the canonical source of `a.ImageURL`. If any inline image's canonical source
matches, set a `suppressLead` flag:

- Skip the top lead `composeImageBlock` call.
- Call `DedupeInline("", inline)` and `StripDuplicateSentinels(bodyMD, "")` (the
  empty-lead case is already handled by their `if leadURL != ""` guards) so the
  first body occurrence is kept and later repeats stripped.
- Build `imageURLs` from the deduped inline set (do not prepend `a.ImageURL`
  separately; it is present via the matching inline URL).

If `suppressLead` is false, keep today's path. This keeps the change localized to
composition; `fireImageLoad`/`ensureImageSource` iterate `imageURLs` and
automatically load only the rendered URLs, so the suppressed lead is not fetched
at position 0.

**D3 — Native/attribution caption resolution.** `nativeRenderCmd`
(`internal/ui/image.go:421`) currently treats `url == a.ImageURL` as the lead and
uses `ResolveImageAttribution(..., isLead=true)`. When the lead is suppressed its
URL is now an inline image, so caption resolution must prefer the inline rules
(`inlineAttrFor`) for it. Implement by checking whether the URL is one of the
article's inline images before falling back to lead attribution. This matches the
inline-image-rendering delta's attribution scenario.

**D4 — Where the decision lives.** Keep the canonical-source match in the compose
package so it is unit-testable and shared by the dedup helpers; keep the
composition branching in `newArticleState`.

## Risks / Trade-offs

- [Stale position-0 row] When the matched body image is not the first inline
  image, its position shifts from 0; a previously-stored position-0 row for the
  old lead URL lingers. Reads are by URL and take the first match
  (`BlockCmd`'s `for i := range imgs`), so the fresh row at the new position may
  not be the one served if a stale row matches first. → Rare (dropsite's
  duplicate is always the first body image, so position stays 0); accept and
  document. If it bites, add a cleanup that removes a stored row whose URL's
  canonical source is now an inline image.
- [Generic encoded-source heuristic misfires] A non-CDN URL that happens to
  contain an encoded URL could be normalized unexpectedly. → Keep the rule
  conservative: only decode a trailing encoded segment when it looks like an
  `http(s)://` source; non-matches fall back to the URL unchanged, preserving
  exact-URL behavior.
- [Caption change] Suppressed lead's caption may differ from the lead-only
  attribution (e.g. `photo: <source>` vs a body figure caption). This is
  intended — the body copy's caption is shown.
