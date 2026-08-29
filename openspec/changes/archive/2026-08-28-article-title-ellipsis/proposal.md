# Truncate the article top-border title with an ellipsis

## Why

When an article title is too long for the top border, `topBorder` hard-cuts it
so the title butts directly against the right corner (`┌─ 2026-01-02 · Hello
world article titleee╖`) — no ellipsis, and no dash before the right corner,
while the left end has `┌─` (corner + dash). It is asymmetric and reads as a
glitch rather than an intentional truncation.

## What

- Keep the date always fully visible; truncate only the title.
- When the title does not fit, end it with an ellipsis glyph — `…` in Unicode
  mode, `...` in ASCII fallback mode.
- Mirror the left edge on the right: a single horizontal dash adjacent to the
  right corner, so the truncated top border reads
  `┌─ 2026-01-02 · Hello world artic… ─╖`.
- When the title fits with room to spare, keep today's dash-fill behavior.

## Non-goals

- Changing the right-edge scrollbar/thumb, the bottom border, or content padding.
- Right-aligning the title or moving the date.
