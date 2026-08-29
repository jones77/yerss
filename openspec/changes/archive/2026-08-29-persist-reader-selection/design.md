## Context

The reader model (`internal/ui/model.go`) builds the list on `loadList` and
opens articles via `openArticle`/`newArticleState`; the article viewport is a
`bubbles/viewport.Model` with a `YOffset` (clamped by `SetYOffset`). The
SQLite store (`internal/store/store.go`) has `feeds`, `articles`, `categories`,
`article_categories` tables but no key-value storage. Quit is handled in the
three `config.Quit` branches of `updateList`, `updateArticle`, and
`updatePopup`, each returning `tea.Quit`.

## Goals / Non-Goals

**Goals:**
- Persist the list selection, the open article, and its scroll offset across
  restarts.
- Restore the cursor to the exact row (article or day header) when possible.
- Fall back gracefully when the saved selection no longer exists.

**Non-Goals:**
- Persisting the tag filter, collapsed day groups, or status messages.
- Continuous persistence on every navigation (persist-on-quit only).
- Surviving crashes or SIGKILL — only a clean quit records state.

## Decisions

### D1: Persist on quit, not continuously

`persistSelection` runs in each `config.Quit` branch right before `tea.Quit`.
This is a single write per session, matches the request ("the line it was
selecting from when the app was quit"), and avoids wiring writes into every
navigation/click/scroll path. Trade-off: an unclean exit (kill, crash) loses
the selection — accepted as out of scope.

### D2: SQLite `state` key-value table

Add a `state (key TEXT PRIMARY KEY, value TEXT)` table to the existing
migration and store the selection as one JSON blob under the key
`last_selection`. This reuses the store's connection and the file already
opened at startup — no new file or dependency. The `LastSelection` struct is
JSON-serialized; `omitempty` keeps the blob minimal.

**Alternative considered:** a separate state file in the data dir. Rejected —
the DB already exists, is migrated by the store, and avoids a second file to
keep in sync.

### D3: Selection keyed by article ID and day-group key

The cursor can sit on an article row or a day header. An article row is keyed
by its `articles.id`; a day header by its calendar-day key (`20060102`, or
`undated` for undated groups). This survives list reordering and refresh
because it is semantic, not positional. On restore, the matching row is found
in `visibleRows()` and the cursor set to its index; a missing match falls back
to `clampCursor`'s default.

### D4: Restore ordering and article reopen

`Init` calls `loadList` (builds the day groups) then `restoreSelection`. The
restore first positions the cursor, then — when the saved view is `article` and
the article still exists — rebuilds the article state and calls
`viewport.SetYOffset` (which clamps to the current content height) to restore
the scroll position, then switches to the article view. The list cursor stays on
the article's row, so `b`/`esc`/`h` from the restored article returns to the
list with that article selected.

### D5: Fallback on missing selection

A saved article ID that no longer exists, or a day key with no matching group,
simply leaves the cursor clamped and the view in the list; a missing or corrupt
blob returns the zero `LastSelection` and is treated as "no saved state". All
restore paths ignore store errors so startup never fails because of stale
selection data.

## Risks / Trade-offs

- **[Unclean exit loses state]** → Persist-on-quit only records on a clean quit;
  a kill/crash starts fresh. Accepted; continuous persistence is out of scope.
- **[Stale selection after list changes]** → If the article or day group is gone,
  the restore falls back to the clamped cursor with no error. Verified by test.
- **[Restored scroll offset beyond current content]** → `SetYOffset` clamps to
  the content's max offset, so a longer article restored into a shorter one
  lands at the bottom rather than corrupting the viewport.

## Open Questions

None.