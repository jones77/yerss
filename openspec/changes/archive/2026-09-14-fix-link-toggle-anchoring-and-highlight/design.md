## Context

See `proposal.md` for motivation. The `x` toggle lives in
`toggleLinkExpand`/`recomposeArticle` (`internal/ui/image.go`). It already
handles two anchor cases — a photo block (`blockRow >= 0`) and a non-blank text
line — but falls through to `recomposeArticle`'s total line-count delta
(`oldOffset + inserted`) whenever the anchor is empty or unfindable. That delta
is only correct for insertions at the top of the content, so a blank or URL-only
first line jumps the reader. Compact-mode reflow is
`ReflowShortLines`/`reflowParagraph` in `internal/ui/compose/reflow.go`;
selection highlighting is `highlightSelection` in
`internal/ui/article_selection.go`.

## Goals / Non-Goals

**Goals:**

- Anchor the `x` toggle on a stable content word even when the first visible
  line is blank or URL text, and pin near-bottom toggles to the bottom.
- Make compact reflow produce exactly one space between words.
- Make the selection highlight continuous across links.
- Resolve the `reader-ui` / `compact-link-mode` `text url` spec conflict.

**Non-Goals:**

- Changing image-block anchoring or the recompose-once contract
  (`article-recomposition`).
- Changing the copy path (`selectedText`) — it already strips before cutting.
- Re-flowing blockquotes/lists/indented lines (still excluded).
- Any change to scroll snapping, image sizing, or the header budget.

## Decisions

### 1. Derive the toggle anchor from the first anchorable word, not the raw first line

Replace `anchor = TrimSpace(Strip(lines[oldOffset]))` with a downward scan from
`oldOffset` that finds the first **anchorable** line and takes its leading text,
skipping two kinds of content that do not survive a mode change:

- **Blank lines** (paragraph separators) are stable; the scan counts them in a
  `gap` and keeps walking.
- **URL text** (a leading token that is a URL or a wrapped URL continuation)
  vanishes in compact mode; the scan skips those tokens (or whole URL-only
  lines) and anchors on the first real word that follows.

The anchor is the leading text of that line (the full trimmed line, so
`findAnchorLine`'s prefix/common-prefix passes stay precise). After the
recompose, position the viewport at `findAnchorLine(anchor) - gap`, clamped, so
the skipped blank separator(s) stay on top. If the scan reaches the end of the
article without finding an anchorable word, glue to the bottom.

**URL-token detection** reuses the link list already on the article state
(`m.article.links`): a leading token is URL text when it contains a scheme
(`://`), starts with `www.`, or is a contiguous substring of a known link URL
(this last case catches wrapped continuations such as
`on-of-narco-terrorist-refueling-vessel-aug-28-2026/`).

*Alternatives considered:* (a) keep the total-delta fallback — rejected, it is
the direct cause of the jumps; (b) anchor on the first word of every non-blank
line and rely on `findAnchorLine` failing — rejected, a mid-paragraph word whose
link name moved across modes won't match and would still fall through.

### 2. Keep the exact-bottom path; add a glue-to-bottom when no anchor exists

`recomposeArticle` already `GotoBottom()` when `oldOffset >= oldLen - vpH` (the
exact last screen); that is kept. The new glue-to-bottom is narrower and only
fires when the anchor scan finds nothing (the tail of the article is blank or
URL only), which is the URL-footer case the user hit. No magic margin is
introduced: "near the bottom" is decided by whether a surviving text word exists
below the offset, not by an arbitrary line-count threshold.

### 3. Trim leading styled whitespace in compact reflow

`reflowParagraph` trims only trailing styled whitespace (`trimStyledWS`) and
then joins continuation lines with a single space. The hidden URL leaves its
following separator space behind, so the join produces a double space. The fix
extends the styled-whitespace trimmer to strip **leading** whitespace (and its
adjacent style/OSC spans) symmetrically with the existing trailing trim, applied
to every line of a reflowable paragraph. `reflowablePara` already excludes
blockquotes, lists, and indented lines, so no legitimate leading whitespace is
lost.

*Alternatives considered:* normalize via `strings.Fields` on the stripped text —
rejected, it would drop styling the paragraph still needs.

### 4. Build the selection highlight from scanned segments, not ansi.Cut/Truncate

`ansi.Cut` = `TruncateLeft(Truncate(...))` does not re-emit the SGR/OSC-8 state
active before the cut, so the selected fragment contains orphan close sequences
whose `\x1b[0m` resets the reverse-video early. `highlightSelection` is rebuilt
on `compose.ScanANSI`:

- Track the active SGR and OSC-8 state through the line.
- `before` = bytes up to `from`, with the active state at `from` closed so it is
  self-contained.
- `middle` = re-open the active state, then the selected bytes, re-asserting
  `\x1b[7m` after every interior full SGR reset so the inversion is continuous;
  OSC-8 opens/closes inside the region are preserved so links stay clickable.
- `after` = re-open the active state at `to` plus the remaining bytes.

The region stays wrapped in reverse-on/off; the difference is that internal
resets no longer leak out of the highlight. (Check x/ansi for an existing
re-emit-aware truncate before writing a custom one.)

*Alternatives considered:* strip all styling from the selected region and invert
the plain text — simpler and continuous, but loses link color and clickability
within the selection; rejected in favor of preserving the styled rendering.

## Risks / Trade-offs

- [URL-token substring matching could misclassify a rare word that happens to be
  a URL fragment] → the misclassified line is only a fallback re-anchor; the
  worst case is a conservative glue-to-bottom, never a crash.
- [Re-anchoring on a leading word that appears many times in the article may
  pick an earlier occurrence] → pre-existing limitation of `findAnchorLine`'s
  "first match" passes; kept as-is, not in scope.
