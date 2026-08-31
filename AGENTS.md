# AGENTS.md

Guidelines for AI agents working in this repository.

## Project

`yerss` is a terminal RSS reader written in Go: a bubbletea/lipgloss TUI, SQLite
for storage, and gofeed for feed parsing. Behavioral requirements live in
`openspec/specs/`.

### Unicode/ASCII glyph pairs

When you add a Unicode glyph (box-drawing, `…`, fold markers, etc.), add its
ASCII fallback in `glyphsFor` (`internal/ui/border.go`) and keep the two in
sync. The `-a`/`--ascii` flag forces the fallback so both paths can be
exercised in testing.

### Investigating articles the user is testing

When the user pastes a link to an article (usually a feed item they opened in
the app), treat it as a pointer to data already on disk, not as a reason to
fetch the open internet. The article's stored content is in the local SQLite
database (`store.Open` in `cmd/yerss/main.go`; the DB path comes from the
`-d`/`--db` flag or `config.Default()`). The schema is defined in code in
`internal/store/store.go` (`Store.migrate` plus the `migrateImageColumns`/
`migrateImageTable` helpers): the `articles` table has `id`, `feed_url`,
`guid`, `title`, `link`, `author`, `published_at`, `content`, `description`,
`read`, `fetched_at`, and the migrated `image_url` (the lead image); the
`article_images` child table stores per-article rendered blocks and photos by
`(article_id, position)`. Query the `articles` table for the row matching the
article's URL, title, feed, or `published_at` to inspect its raw HTML
`content` (e.g. the `<img>` elements and their sizes). Only fetch over the
network when the needed data genuinely is not stored locally.

### Marked-converter markdown format (inline images and links)

`ConvertImages` emits a NUL-delimited sentinel `\x00img:<url>\x00` for each
inline `<img>` in the article body, in document order. The html-to-markdown
converter wraps `<a>` content in markdown links, so a linked image renders as
`[\x00img:<url>\x00 ...link text...](/href)` — the sentinel, any link text
after it (promo banners like the TomDispatch logo carry their label there),
then the destination. `nextSentinel` in `internal/ui/article.go` consumes the
whole link construct (so no stray `[`, link text, or `](...)` renders around
the image block) and reports the link text for use as the image's caption
fallback. Relative destinations (e.g. `/tomdispatch`) are resolved against the
article's origin when harvesting links, so the links popup and open-url get
absolute targets.

### Image scroll snap invariants

`snapYOffset` in `internal/ui/image.go` keeps single-line scrolls (j/k, mouse
wheel) from leaving an inline image partially visible — a partially visible
native placement is deleted (it would draw over the border), so those rows
would render blank. The down bottom-snap fires on partial visibility (an image
whose top is inside the window and whose bottom is below the fold), because a
skip of one image leaves the next image's top already inside the window when
consecutive images are spaced closer than the viewport height. Every snap is
guarded to never return the pre-move offset — a snap that bounces back to it
re-fires identically and loops. A block whose image plus wrapped caption fills
the viewport exactly has no distinct bottom stage (bottom-align equals
top-align), so the up snap scrolls it off instead. The halfblock preview
(`previewClippedImage`) fills a partially visible image's visible rows so a
poking-in photo shows a blocky preview, never a blank strip.

Blocks can fill the viewport even though the fit caps the image height, because
`nativeRenderCmd`/`inlineAttrFor` fit against a caption without the inline link
text while composition uses the full caption — keep the fit caption and the
composed caption in agreement, or a long caption overflows the reserved budget
and inflates the block. When debugging snap behavior, reproduce the article's
real block geometry from the local SQLite DB (a temporary test opening the
store, rendering the native blocks via the real height fit, and simulating the
scrolls) before changing the rules.

## OpenSpec workflow

- Work strictly within `openspec/changes/<change>/tasks.md` for the active change;
  update task checkboxes (`- [ ]` → `- [x]`) in real time as work completes.
- Never mark a task `[x]` without executing tests, the linter, or a build that
  proves the change works.
- Keep changes clean and localized; run `git status` / `git diff` at task
  boundaries to avoid unintended side effects.
- Consult `openspec/changes/<change>/specs/` (delta specs) and the main
  `openspec/specs/` tree for the behavior requirements you are implementing.

## Commit messages

Commit messages SHALL follow the seven rules from
https://cbea.ms/git-commit/ (the source of truth):

1. Separate subject from body with a blank line.
2. Limit the subject line to 50 characters.
3. Capitalize the subject line.
4. Do not end the subject line with a period.
5. Use the imperative mood in the subject line.
6. Wrap the body at 72 characters.
7. Use the body to explain what and why, not how.

Changes are committed at OpenSpec archive time. When archiving, stage the archived
change directory, the synced `openspec/specs/` paths, and the change's
implementation files explicitly (not `git add -A`) and commit them in one message
following the policy above.