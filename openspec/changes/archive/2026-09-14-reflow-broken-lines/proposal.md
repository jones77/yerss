## Why

Two text-layout defects in the article view break the reading experience on
narrow terminals. First, a long blockquote whose final word ends in punctuation
(e.g. `on.”`) renders the closing punctuation (`”`) stranded on its own bare
line without the quote bar. Second, in compact link mode a wrapped link whose
URL text spans a line break is left as a short, unpadded line after the URL is
stripped, producing a visible blank gap mid-paragraph. Both stem from glamour's
line wrapping not surviving downstream post-processing in the compose package.

## What Changes

- Repair blockquote re-wrapping so a line consisting only of closing
  punctuation that immediately follows a quote-bar line is absorbed back into
  the blockquote and re-wrapped with the quote bar (instead of being left as a
  bare orphan line).
- Add a re-flow pass for compact link mode that re-wraps paragraphs whose lines
  were left shorter than the content width after link-URL stripping, restoring
  full-width, properly padded lines with no mid-paragraph gaps.
- Both fixes operate on already-rendered styled text and preserve inline styles
  and OSC 8 hyperlink spans; ASCII fallback and theme behavior are unchanged.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `reader-ui`: article content rendering — blockquote closing punctuation must
  not be stranded off the quote bar when the quote's last word wraps.
- `compact-link-mode`: wrapped link URLs are fully hidden in compact mode —
  stripping a wrapped URL must not leave a short, unpadded line in the paragraph.

## Impact

- `internal/ui/compose/markdown.go`: `FixBlockquoteRewrap` / `isOrphanLine`
  handling of trailing-punctuation orphans.
- `internal/ui/compose/`: new re-flow helper (styled-text-aware re-wrapping at
  content width).
- `internal/ui/article.go`: `computeDerivation` applies the compact-mode re-flow
  after `StripLinkHrefs` when links are hidden.
- Tests in `internal/ui/compose/` and `internal/ui/`; no new dependencies
  beyond what the compose package already uses.