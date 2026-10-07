## Why

The read marker is ambiguous. The original requirement ("mark read on open when
the content fits; otherwise on the first downward scroll") left three axes loose:
*when* to mark (open vs. scroll vs. reaching the bottom), *which direction*, and
*which actions* count as movement. The result was behavior — and a reading of
it — that diverged from intent: an article whose content fits the viewport was
marked read on open, but a longer article was only marked on a downward scroll,
so a user who opened a long article and read the visible portion (or moved
upward, or jumped) still saw it bold. The intended rule is simpler: an article
that fits is read on open; an article that does not fit is read on the user's
first movement through it, in any direction.

## What Changes

- An article whose content fits the viewport is marked read when opened (no
  change — this already worked).
- An article whose content does not fit is marked read on the user's first
  movement: a line, page, or half-page scroll in either direction, a jump to
  the top or bottom, or a mouse-wheel scroll in either direction. Movement
  marks read even when the scroll position does not actually change.
- Restoring a saved scroll offset when re-opening an article does not count as
  movement.
- Once read, further movement does not change the read status.

## Why Not

- No change to the list rendering: read titles already render in the plain
  `text` role and unread titles in the bold `bright` role.
- No new toggle/mark-all bindings: `m` and `a` were deliberately removed and
  stay removed.
- No change to read-status persistence (`SetRead`) or the list reload on back.
