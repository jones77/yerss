## MODIFIED Requirements

### Requirement: Restore reader state on start

The system SHALL persist the reader's selection when the application quits and
restore it when the application starts again. The persisted selection SHALL
record the current view (list or article), the row under the cursor (the
article by its ID, or the day group by its calendar-day key when the cursor is
on a day header), and, when the article view is open, the article's scroll
offset together with a content anchor for that offset. On startup the system
SHALL move the list cursor to the row matching the saved selection. When the
saved view is the article view and the saved article still exists, the system
SHALL open that article and restore its scroll position, so that returning to
the list (via `b`, `esc`, or `h`) leaves the same article selected.

The content anchor SHALL identify the reading position independently of the
rendered line count, so it survives a restart even when image blocks have not
materialized yet and the initial composition happens at a different size than
the real terminal. When the first visible line of the viewport lies inside a
composed image block, the anchor SHALL be that block; otherwise the anchor
SHALL be the first text line's leading content together with the number of
lines between that line and the viewport top (the gap). On restore, the system
SHALL re-locate the anchored line in the freshly composed article and position
the viewport so the anchored line sits at its saved gap from the top; when the
anchor is an image block, the system SHALL position the viewport on the block's
placeholder row. When the restored anchor was chosen at a size that differs
from the real terminal size, the system SHALL re-apply the anchor once the true
terminal size is known, so the reading position lands on the same line the user
was reading. When no usable anchor can be located, the system SHALL fall back
to the saved raw offset. When the saved selection no longer exists — for
example the article was pruned or the day group changed — the system SHALL fall
back to the default clamped cursor without error.

#### Scenario: List cursor restored on start

- **WHEN** the user quits with the cursor on an article row in the list view and restarts
- **THEN** the list cursor is on that same article's row

#### Scenario: Day header selection restored on start

- **WHEN** the user quits with the cursor on a day-header row and restarts
- **THEN** the list cursor is on that day group's header row

#### Scenario: Article view reopened on start

- **WHEN** the user quits while reading an article and restarts
- **THEN** the reader reopens that article, and pressing `b`, `esc`, or `h` returns to the list with that article still selected

#### Scenario: Article scroll position restored on start

- **WHEN** the user quits while reading an article scrolled partway and restarts
- **THEN** the reopened article is scrolled to the same position

#### Scenario: Restore lands on the anchored line before blocks materialize

- **WHEN** the user quits while reading partway down an article that contains
  image blocks whose stored blocks do not match the restart content width, and
  restarts
- **THEN** the reopened article lands on the same content line the user was
  reading, even though the image blocks render as placeholders initially

#### Scenario: Restore re-applies the anchor at the real terminal size

- **WHEN** the reader restores an article at a size that differs from the real
  terminal size and the first real window-size message then arrives
- **THEN** the anchor is re-applied at the real size so the reading position
  stays on the line the user was reading

#### Scenario: Restore falls back to the raw offset without an anchor

- **WHEN** the saved article contains no text line or image block that the
  anchor can locate
- **THEN** the system restores the saved raw scroll offset

#### Scenario: Fallback when the saved article no longer exists

- **WHEN** the saved article was removed from the store and the user restarts
- **THEN** the list opens with the default cursor and the application does not error