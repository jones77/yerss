## Why

Compact link mode promises to show only each link's clickable name, never the
visible URL. A link whose URL is longer than the content width violates that
promise: lipgloss re-emits the OSC 8 hyperlink on every wrapped line, so the href
splits into several same-target spans, and `StripLinkHrefs` drops only the first
— leaking the wrapped continuation (e.g. `rula-jebreal-and-the-democratic`) into
the article body. A link name long enough to wrap has the mirrored bug: its
second line is mistaken for the href and dropped.

## What Changes

- Rework `compose.StripLinkHrefs` to group same-target OSC 8 spans into logical
  link units and drop each href unit as a whole — its leading space plus every
  wrapped continuation line — rather than assuming one href span per link.
- Add regression tests for a link whose URL wraps across multiple lines, for a
  link name that itself wraps, and for two adjacent links to the same URL.

## Capabilities

### New Capabilities

<!-- None. -->

### Modified Capabilities

- `compact-link-mode`: add a requirement that a link's visible URL is fully
  hidden in compact mode even when the URL wraps across multiple rendered
  lines, while the link name remains fully shown and clickable.

## Impact

- **Code:** `internal/ui/compose/links.go` (`StripLinkHrefs`).
- **Tests:** `internal/ui/link_expand_test.go`.
- **Behavior:** Compact mode no longer leaks wrapped href continuation text, and
  no longer drops a wrapped link name.
- **Dependencies:** None.
