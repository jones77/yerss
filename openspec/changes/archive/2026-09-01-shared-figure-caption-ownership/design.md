## Context

See proposal.md — Why. `figureCredit` (convert/credit.go) attributes a figure's
`figcaption` to any `<img>` in the figure's subtree, so a photo-grid — one
`<figure>` wrapping several photos with a single combined `Left: … Right: …`
caption (verified against the local store, Intercept article id 75) — copies
the caption onto every photo. The caption belongs to the figure's **last** image
in document order; earlier images in the shared figure must render with no
caption at all, including no `photo: <source>` fallback.

## Goals / Non-Goals

**Goals:**
- Attribute a figure's caption only to its last `<img>` (document order).
- A shared-figure non-owner image (an earlier image in a captioned figure)
  renders with no caption — no borrowed caption, no `photo:` fallback.
- Apply uniformly to the lead and inline attribution paths.

**Non-Goals:**
- No split of combined `Left:/Right:` captions across grid photos — the user
  decided the caption belongs to the last photo as a whole, not split.
- No change to the alt/link-text attribution sources for images that do own a
  caption; alt, figure caption, link text, then `photo:` fallback ordering is
  preserved for non-shared-figure images.

## Decisions

**1. `figureCredit` attributes the caption to the figure's last `<img>` only.**

For each captioned `<figure>` in document order, the caption is returned only
when the queried URL is the last `<img>` in that figure's subtree (the outer
photo-grid figure contains both photos, so its last `<img>` is the second
photo). An earlier image that is still *inside* a captioned figure must be
distinguishable from "no figure at all," so the extraction exposes both the
caption and whether the image is in a captioned figure.

New convert API:

```go
// ImageCaption returns the caption attributed to imageURL and whether
// imageURL sits in a captioned figure. A figure's caption belongs to its last
// <img>; an earlier image in a shared figure gets "" with matched=true (the
// caller must render no caption rather than fall back).
func ImageCaption(htmlStr, imageURL string) (caption string, matched bool)
```

`ImageCredit` stays as a thin wrapper returning just the caption, so the lead
path and existing callers/tests are unaffected in signature.

*Alternative considered:* splitting the combined `Left:`/`Right:` caption into
per-photo segments. Rejected — prose-based splitting is fragile across sources
and the user's decision is that the last photo owns the whole caption.
*Alternative considered:* a nearest-figure rule (caption attaches to whichever
photo is DOM-nearest). Rejected — it would drop the caption from both grid
photos; the caption must survive on the last photo.

**2. Attribution paths skip the `photo:` fallback for shared-figure non-owners.**

`inlineAttribution` (internal/ui/article.go) and `articleAttribution`
(internal/ui/image.go) currently fall through to `photo: <source>` whenever the
credit is empty. For a shared-figure non-owner this would render
`photo: theintercept` under the first grid photo — still a caption. Both paths
now check `ImageCaption`'s matched flag first: matched-but-empty means "in a
captioned figure whose caption belongs elsewhere" → return "" (no caption),
bypassing alt, link text, and the source fallback.

Order in `inlineAttribution`:

```
if cap, inFig := convert.ImageCaption(content, url); inFig && cap == "" {
    return ""                      // shared-figure non-owner: no caption at all
}
alt → linkText → "photo: <source>"
```

`articleAttribution` mirrors it for the lead (credit, else photo: fallback,
with the matched-but-empty case returning "").

## Risks / Trade-offs

- **Behavior change for multi-image figures only.** Single-image figures and
  no-figure images behave exactly as before; the existing credit tests lock in
  single-figure and no-borrowing behavior and must pass unmodified.
- **The last-`<img>` rule is a heuristic for shared captions.** A figure whose
  caption genuinely labels the *group* (a map key) will show that caption only
  under its last photo. Accepted: it matches the source convention and the
  user's decision; a truly group-level caption has no per-image home.
- **Nested-figure ordering.** `findAll(fig, "img")` includes nested figures, so
  the outer grid figure's last image is the second photo (correct); the inner
  per-photo figures carry no caption and are skipped. Verified against the real
  article geometry.

## Migration Plan

Pure logic change in `convert/credit.go` and the two attribution call sites
plus tests. No schema, storage, or CLI changes; no rollout steps. Rollback is
the previous commit.

## Open Questions

None — the caption-ownership rule and the "no caption whatsoever" behavior
(bypassing the `photo:` fallback) were decided during exploration.