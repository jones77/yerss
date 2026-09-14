## Why

The article reader renders the article's own source URL as the first content
line regardless of link display mode. Compact mode is meant to hide URLs so the
reader is uncluttered, but the header URL still appears, so compact mode only
hides body-link URLs and not the article's source URL.

## What Changes

- Gate the reader header's URL line on link display mode: in compact mode the
  header URL is not rendered (the title becomes the first content line); in
  expanded mode the header URL renders exactly as it does today.

## Capabilities

### New Capabilities

<!-- None. -->

### Modified Capabilities

- `reader-ui`: change the "Clickable article URL in reader header" requirement so
  the header URL renders only in expanded link mode and is omitted in compact
  mode, with the title becoming the first content line.

## Impact

- **Code:** `internal/ui/article.go` (`computeDerivation`, `renderHeader`,
  `renderHeaderLink`).
- **Tests:** `internal/ui/link_expand_test.go` header test gains a compact-mode
  sibling.
- **Behavior:** Compact mode now hides the header URL as well as body-link URLs.
- **Dependencies:** None.
