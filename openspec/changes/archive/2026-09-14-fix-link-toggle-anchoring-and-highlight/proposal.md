## Why

Pressing `x` to toggle compact/expanded link mode frequently throws the reader to
the wrong place: a blank first-visible line jumps the viewport hundreds of lines
forward on expand, and a URL-only line near the bottom leaves a confusing
mid-article line at the top on collapse. Separately, compact mode emits a stray
double space where a hidden URL used to sit, and mouse-selecting text stops
highlighting at the first link on each line.

## What Changes

- **Anchor the `x` toggle on the first real content word.** When the first
  visible line is blank or a bare URL (both of which do not survive a mode
  change), the toggle derives its anchor from the first non-blank, non-URL text
  line at or below the offset — using its leading word(s) per the
  "first word from the other mode" rubric — and keeps the skipped blank
  separator in place.
- **Pin near-bottom and URL-line toggles to the bottom.** A toggle whose anchor
  is blank or URL-only and lands at or near the article end stays glued to the
  bottom instead of re-anchoring on mid-article text.
- **Eliminate the double space after stripped link names.** Compact reflow trims
  the leading separator space left where a URL was removed, so `claimed  was`
  becomes `claimed was`.
- **Fix fragmented selection highlighting.** The highlight path re-emits the
  active SGR/OSC-8 state at each cut boundary so reverse-video spans the whole
  selection across links instead of resetting at the first link.
- **Resolve the spec conflict** in `reader-ui` "Article content rendering" that
  still mandates unconditional `text url` output, contradicting the compact
  default in `compact-link-mode`.

## Capabilities

### New Capabilities

<!-- none -->

### Modified Capabilities

- `compact-link-mode`: tighten the "Link mode toggle" requirement to cover
  blank/URL-only anchor lines and near-bottom pinning, and add the single-space
  invariant for stripped-link reflow.
- `reader-ui`: revise "Article content rendering" so hyperlink rendering defers
  to the link-mode requirements (removing the stale unconditional `text url`),
  and clarify "In-view text selection" so the highlight spans link boundaries.

## Impact

- `internal/ui/image.go` — `toggleLinkExpand`, `findAnchorLine` (anchor
  derivation and bottom pinning).
- `internal/ui/compose/reflow.go` — `reflowParagraph`/`trimStyledWS` (leading
  whitespace trim).
- `internal/ui/article_selection.go` — `highlightSelection` (escape-balanced
  slicing via `compose.ScanANSI`).
- `openspec/specs/reader-ui/spec.md`, `openspec/specs/compact-link-mode/spec.md`
  — requirement wording.
- No API, dependency, or storage changes.
