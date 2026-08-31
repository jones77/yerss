## Purpose

Handles fetching RSS/Atom feeds, the refresh lifecycle (startup gate and manual cooldown), article deduplication, and persistent storage of all downloaded articles in a local SQLite database.

## Requirements

### Requirement: Feed URL source

The system SHALL read feed entries from a plain-text file with one entry per
line. Blank lines and lines whose first non-whitespace character is `#` SHALL
be ignored, so the file may contain comments. The default path SHALL be
`$XDG_CONFIG_HOME/yerss/feeds.txt` (fallback `~/.config/yerss/feeds.txt`). The
path SHALL be configurable via the TOML config file.

Entries SHALL be scheme-less: the `https://` scheme is assumed on every
lookup. When an entry carries a leading `https://`, the system SHALL strip it
and proactively rewrite the file in place with the canonical scheme-less form,
preserving comment and blank lines verbatim; when an entry carries a leading
`http://`, the system SHALL silently strip it and look the feed up as
`https://`. The rewrite is best-effort: a failure leaves the file untouched
while the in-memory entries stay normalized. Each entry SHALL be fetched from
and stored under its canonical `https://` URL, so article deduplication keys
are unchanged for feeds already stored under `https://`.

#### Scenario: Feeds loaded from default path

- **WHEN** the application starts and no custom feeds path is configured
- **THEN** the system reads feed URLs from `$XDG_CONFIG_HOME/yerss/feeds.txt` or `~/.config/yerss/feeds.txt` as fallback

#### Scenario: Custom feeds path

- **WHEN** the TOML config specifies a custom feeds file path
- **THEN** the system reads feed URLs from that path instead of the default

#### Scenario: Comments and blank lines ignored

- **WHEN** the feeds file contains blank lines and lines beginning with `#` interspersed with URL lines
- **THEN** the system skips the blank and comment lines and loads only the URL lines

#### Scenario: Scheme stripped and file rewritten

- **WHEN** the feeds file contains entries written with a leading `https://`
- **THEN** the entries are loaded scheme-less and the file is rewritten in place with the scheme-less form, comments and blank lines preserved

#### Scenario: HTTP entries silently upgraded to HTTPS

- **WHEN** the feeds file contains an entry written with a leading `http://`
- **THEN** the scheme is silently stripped and the feed is looked up as `https://`, with no warning printed

#### Scenario: Canonical lookups keep deduplication stable

- **WHEN** a scheme-less entry is fetched and the store already holds that feed's articles under the `https://` URL
- **THEN** the fetched articles deduplicate against the stored rows instead of being inserted twice

#### Scenario: Rewrite failure still loads normalized entries

- **WHEN** the feeds file carries schemes but cannot be rewritten (for example a read-only directory)
- **THEN** the load succeeds with scheme-less entries and no error is surfaced

### Requirement: Usage output documents the data files

The system SHALL document its three data files in the CLI usage output: the
TOML config file, the feeds.txt file, and the SQLite database — each with its
resolved default path and a one-line description, noting that the config's
`[data]` section can relocate the feeds file and database.

#### Scenario: Usage lists the data files

- **WHEN** the user runs the program with the help flag
- **THEN** the usage output lists the resolved config, feeds, and database paths with short descriptions and the `[data]` relocation note

### Requirement: Startup refresh with 15-minute gate

The system SHALL attempt to refresh all feeds on startup only if more than 15
minutes have elapsed since the last successful refresh. If the last refresh was
within 15 minutes and every configured feed URL has been successfully fetched
before, the system SHALL load articles from the local database without fetching
from the network. However, if the configured feeds file contains a URL for which
no verified feed row exists — meaning no `feeds` row for that URL with a
non-NULL `last_fetched_at`, proving it has never been successfully fetched — the
system SHALL perform a startup refresh regardless of how recently the last
refresh occurred, so that newly added feeds are fetched without requiring a
manual refresh.

#### Scenario: Startup refresh after 15 minutes

- **WHEN** the application starts, every configured feed URL has a verified feed row, and the last refresh was more than 15 minutes ago
- **THEN** the system fetches all feeds from the network and updates the database

#### Scenario: Startup skip refresh within 15 minutes

- **WHEN** the application starts, every configured feed URL has a verified feed row, and the last refresh was less than 15 minutes ago
- **THEN** the system loads articles from the database without network fetching

#### Scenario: Startup refresh when new feeds were added

- **WHEN** the application starts and the configured feeds file contains a URL that has no verified feed row (it was added since the last refresh)
- **THEN** the system fetches all feeds from the network on startup regardless of how recently the last refresh occurred, so the newly added feed is fetched

#### Scenario: All-known feeds within 15 minutes still skips refresh

- **WHEN** the application starts, every configured feed URL already has a verified feed row, and the last refresh was less than 15 minutes ago
- **THEN** the system loads articles from the database without network fetching, even if feeds.txt was otherwise touched, because no feed is new

### Requirement: Manual refresh with 60-second cooldown

The system SHALL perform a manual refresh when the user presses `R`, Ctrl-R, or F5 on the list view. Manual refresh SHALL bypass the 15-minute startup gate. Manual refresh SHALL be throttled by a 60-second cooldown: if a refresh was performed less than 60 seconds ago, the system SHALL NOT refresh and SHALL display a transient "next allowed in Xs" message in the status bar indicating how many seconds remain.

#### Scenario: Manual refresh bypasses 15-minute gate

- **WHEN** the user presses R, Ctrl-R, or F5 on the list view and the last refresh was less than 15 minutes ago
- **THEN** the system fetches all feeds from the network and updates the database

#### Scenario: Manual refresh within 60-second cooldown

- **WHEN** the user presses R, Ctrl-R, or F5 and a refresh was performed less than 60 seconds ago
- **THEN** the system does not refresh and displays "next allowed in Xs" in the status bar where X is the remaining seconds

#### Scenario: Manual refresh after cooldown expires

- **WHEN** the user presses R, Ctrl-R, or F5 and 60 or more seconds have elapsed since the last refresh
- **THEN** the system fetches all feeds and updates the database

### Requirement: Bounded feed fetch deadlines

Every feed fetch the system performs — whether for the startup verification gate, the
startup refresh, or a manual refresh — SHALL be bounded by both a per-feed HTTP
timeout and an overall deadline. A feed whose fetch exceeds the per-feed timeout
SHALL be cancelled and recorded as an error in the fetch result; it SHALL NOT block
the completion of the remaining feeds or the refresh pass as a whole. The overall
deadline SHALL cancel any still-in-flight fetches once it elapses. The system SHALL
cap the number of concurrently in-flight feed requests with a fixed bound rather than
spawning one unbounded goroutine per configured feed. The startup verification gate
and the refresh path SHALL share one bounded fetch implementation so their timeout
behavior cannot diverge.

#### Scenario: Hung feed does not stall a manual refresh

- **WHEN** the user triggers a manual refresh and one configured feed's server accepts the TCP connection but never responds
- **THEN** that feed is cancelled at the per-feed timeout, recorded as an error in the fetch result, and the refresh pass completes with the other feeds within the overall deadline

#### Scenario: Overall deadline cancels in-flight fetches

- **WHEN** a fetch pass is in progress and the overall deadline elapses before all feeds have completed
- **THEN** the system cancels every still-in-flight feed fetch and completes the pass with the feeds that already finished

#### Scenario: Concurrency is bounded

- **WHEN** a fetch pass runs against more configured feeds than the concurrency bound
- **THEN** the system holds excess feeds until in-flight slots free up rather than issuing all requests at once

#### Scenario: Gate and refresh share timeout behavior

- **WHEN** the startup verification gate and a manual refresh both fetch the same set of feeds
- **THEN** both paths apply the same bounded-timeout and concurrency behavior because they use one shared fetch implementation

### Requirement: Article deduplication

The system SHALL store articles with a unique constraint on the combination of feed URL and article GUID. When a feed is re-fetched, articles that already exist (matching feed URL and GUID) SHALL NOT be duplicated; their content and metadata SHALL be updated if changed.

#### Scenario: Duplicate article not re-inserted

- **WHEN** a feed is fetched and an article with the same feed URL and GUID already exists in the database
- **THEN** the existing article's content and metadata are updated and no duplicate row is created

#### Scenario: New article inserted

- **WHEN** a feed is fetched and an article has a GUID not seen before for that feed
- **THEN** a new article row is created in the database

### Requirement: SQLite database location and validation

The system SHALL store all articles in a SQLite database at `$XDG_DATA_HOME/yerss/yerss.sqlite` (fallback `~/.local/share/yerss/yerss.sqlite`). The path SHALL be configurable via the TOML config file. The system SHALL refuse to start if it cannot create or write to the database file, and SHALL print a usage message that references the database file path.

#### Scenario: Database created at default path

- **WHEN** the application starts and no custom database path is configured
- **THEN** the system creates or opens the database at `$XDG_DATA_HOME/yerss/yerss.sqlite` or `~/.local/share/yerss/yerss.sqlite` as fallback

#### Scenario: Application refuses to start on DB failure

- **WHEN** the system cannot create or write to the database file
- **THEN** the application prints a usage message referencing the database path and exits without starting the TUI

#### Scenario: Custom database path

- **WHEN** the TOML config specifies a custom database path
- **THEN** the system uses that path instead of the default

### Requirement: Article persistence across restarts

The system SHALL persist every downloaded article in the SQLite database. Articles SHALL survive application restarts. Read status SHALL be persisted and restored on the next launch. Each article SHALL persist a nullable lead image URL captured from the article's `<enclosure>` or `media:thumbnail` metadata when present; articles without such metadata SHALL store a null image URL. When an article's lead image is fetched for display, the system SHALL persist the raw image bytes in the database so that subsequent opens and application restarts render the image without re-fetching from the network.

#### Scenario: Articles survive restart

- **WHEN** the application is restarted after fetching articles
- **THEN** all previously fetched articles are loaded from the database

#### Scenario: Read status persisted

- **WHEN** an article is marked as read and the application is restarted
- **THEN** the article retains its read status on the next launch

#### Scenario: Lead image URL persisted

- **WHEN** an article with an enclosure image URL is fetched and the application is restarted
- **THEN** the article's image URL is restored from the database on the next launch

#### Scenario: Lead image bytes persisted across restart

- **WHEN** an article's lead image has been fetched and stored, and the application is restarted and the article is re-opened
- **THEN** the image renders from the stored bytes without making a network request to the image host

### Requirement: Lead image URL extraction from feeds

During feed parsing, the system SHALL extract a lead image URL for each article from the first image-typed `<enclosure>` or the `media:thumbnail` extension when present. When multiple enclosures exist, the system SHALL select the first one whose MIME type is an image type (`image/*`). When no image enclosure or media thumbnail is present, the article SHALL have no lead image URL. The system SHALL NOT extract image URLs from inline `<img>` elements within the article HTML content; only publisher-attached enclosure/thumbnail metadata is captured.

#### Scenario: Enclosure image URL captured

- **WHEN** a feed item has an enclosure with type `image/jpeg` and URL `https://example.com/lead.jpg`
- **THEN** the article's lead image URL is set to `https://example.com/lead.jpg`

#### Scenario: Media thumbnail URL captured

- **WHEN** a feed item has no image enclosure but has a `media:thumbnail` URL `https://example.com/thumb.png`
- **THEN** the article's lead image URL is set to `https://example.com/thumb.png`

#### Scenario: Non-image enclosure ignored

- **WHEN** a feed item has an enclosure with type `audio/mpeg` and no image enclosure or media thumbnail
- **THEN** the article has no lead image URL

#### Scenario: Inline content image not captured

- **WHEN** a feed item has no enclosure or media thumbnail but its HTML content contains `<img src="https://cdn.example.com/track.png">`
- **THEN** the article has no lead image URL

### Requirement: Additive schema migration for image columns

The system SHALL add two nullable columns to the `articles` table: `image_url TEXT` (the publisher-attached image URL) and `image_data BLOB` (the fetched image bytes). The migration SHALL be idempotent: it SHALL check the existing column set before adding each column so that re-running the migration on an already-migrated database is a no-op. The migration SHALL not drop or alter existing columns or data.

#### Scenario: Columns added on first migration

- **WHEN** the database is opened and the `articles` table does not have an `image_url` or `image_data` column
- **THEN** both columns are added as nullable (`TEXT` and `BLOB` respectively) and existing rows retain their data with null image URLs and image data

#### Scenario: Migration idempotent on already-migrated database

- **WHEN** the database is opened and the `articles` table already has both `image_url` and `image_data` columns
- **THEN** the migration makes no changes to the schema

### Requirement: Automatic tagging from feed categories

The system SHALL extract categories from each article using the feed's
category metadata (as provided by gofeed). Each unique category name SHALL be
stored as a tag. Tag names SHALL be unique case-insensitively but
case-preserving: the first spelling ever stored becomes the canonical name
forever, so `energy`, `Energy`, and `ENERGY` resolve to one tag whose spelling
is the first one seen. Tag matching and filtering SHALL be case-insensitive.
Articles SHALL be associated with their categories in the
database, enabling tag-based filtering and browsing. In addition, every
article SHALL be associated with a source-domain tag derived from its link
host — or its feed URL host when the article has no link — using the
registrable organization label (for example `https://www.nytimes.com` →
`nytimes`, `https://a.lot.of.subdomains.nytimes.com` → `nytimes`). Source tags
SHALL be marked as source tags so the tag popup can group news organizations
ahead of categories, and they SHALL filter the list like category tags. When
the source label duplicates a category name (case-insensitively), it SHALL be
associated once.

#### Scenario: Article with categories tagged

- **WHEN** a feed is fetched and an article has categories "tech" and "news"
- **THEN** the article is associated with tags "tech" and "news" in the database

#### Scenario: Article with no categories

- **WHEN** a feed is fetched and an article has no categories
- **THEN** the article is still stored and displayed, and is associated with its source-domain tag derived from its feed or link host

#### Scenario: Article tagged with its source domain

- **WHEN** a feed is fetched and an article links to `https://www.nytimes.com/story/1` from a feed at `https://feeds.example.com/rss`
- **THEN** the article is additionally associated with the tag `nytimes`, so the tag popup lists it and filtering by it shows that article

#### Scenario: Source-domain tag deduplicates against a category

- **WHEN** an article's source-domain label equals one of its category names (case-insensitively)
- **THEN** the article is associated with that name once, not twice

#### Scenario: Tag names are case-insensitive and case-preserving

- **WHEN** articles are tagged `energy`, then `ENERGY`, then `Energy`
- **THEN** the database stores a single `energy` tag (the first spelling seen) with all three articles associated, and filtering by `ENERGY` or `Energy` finds them

#### Scenario: Case-insensitive tag matching

- **WHEN** the user filters by a tag name whose spelling differs in case from the stored tag
- **THEN** the matching articles are returned, because tag matching is case-insensitive

### Requirement: Startup feed-verification gate

Before launching the terminal UI, the system SHALL run a feed-verification gate.
The gate SHALL read the configured feeds file. If the feeds file is missing or
contains no URL lines (only blanks and comments), the system SHALL print a
message naming the feeds file path and SHALL exit with code 2 without entering
the TUI. If the feeds file exists but cannot be read (for example a permission
denied or I/O error distinct from the file being absent), the system SHALL print a
distinct `cannot read feeds file <path>: <error>` message and SHALL exit with
code 1 without entering the TUI; this read-error case is separate from the
empty/missing case and SHALL NOT reuse the exit 2 "no feeds configured" message.
If the feeds table contains at least one previously-verified feed — defined as a
feed row whose `last_fetched_at` is set (non-NULL), proving the feed was
successfully fetched and persisted in the past — the system SHALL proceed to the
TUI without performing a network check. A feed row whose `last_fetched_at` is NULL
(such as a stub row created only by article persistence) SHALL NOT by itself
satisfy this trust rule. Otherwise the system SHALL fetch the configured feeds
concurrently, bounded by a 10-second per-feed timeout and a 30-second overall
deadline. While the verification fetch is in progress, the system SHALL print a
`verifying feeds...` notice to stderr. If at least one feed parses successfully,
the system SHALL proceed to the TUI and the successful fetch SHALL count as the
first refresh (so the startup refresh gate SHALL skip a redundant fetch). If no
feed parses successfully, the system SHALL print the per-feed errors and SHALL
exit with code 3 without entering the TUI.

#### Scenario: Empty feeds file refuses start

- **WHEN** the feeds file exists but contains only blank lines and comments
- **THEN** the system prints a message naming the feeds file path and exits with code 2 without entering the TUI

#### Scenario: Missing feeds file refuses start

- **WHEN** the feeds file does not exist at the resolved path
- **THEN** the system prints a message naming the feeds file path and exits with code 2 without entering the TUI

#### Scenario: Unreadable feeds file refuses start with a distinct error

- **WHEN** the feeds file exists at the resolved path but cannot be read (for example permission denied)
- **THEN** the system prints a `cannot read feeds file <path>: <error>` message and exits with code 1 without entering the TUI, distinct from the exit 2 empty/missing behavior

#### Scenario: Previously-verified feed skips the network check

- **WHEN** the feeds file contains at least one URL and the feeds table already contains a feed row whose `last_fetched_at` is set
- **THEN** the system proceeds to the TUI without performing any network fetch during the gate

#### Scenario: Stub feed row alone does not satisfy the gate

- **WHEN** the feeds file contains at least one URL and the feeds table contains only feed rows whose `last_fetched_at` is NULL (stub rows created by article persistence without a successful feed fetch)
- **THEN** the system does not skip the network check on trust and proceeds to perform the verification fetch

#### Scenario: First run with a working feed starts

- **WHEN** the feeds file contains at least one URL, the feeds table has no previously-verified feed, and at least one configured feed parses successfully within the per-feed and overall deadlines
- **THEN** the system persists the verified feed, proceeds to the TUI, and the startup refresh gate skips a redundant fetch because a refresh just occurred

#### Scenario: First run with no working feed refuses start

- **WHEN** the feeds file contains at least one URL, the feeds table has no previously-verified feed, and no configured feed parses successfully within the per-feed and overall deadlines
- **THEN** the system prints the per-feed errors and exits with code 3 without entering the TUI

#### Scenario: Verifying notice printed to stderr

- **WHEN** the gate performs a verification fetch (no previously-verified feed)
- **THEN** the system prints a `verifying feeds...` notice to stderr before the fetch begins
### Requirement: Exit diagnostics for failed feed URLs

After the TUI exits, the system SHALL print per-URL fetch diagnostics to
stderr, prefixed with the program's basename (`<basename>: `). Each URL from
the most recent completed refresh pass SHALL produce exactly one line: an
`error: <status>, url: <URL>` line when the fetch failed with an HTTP error
status (for example `404`), an `error: <error>, url: <URL>` line when the
fetch failed for any other reason (for example a connection error or invalid
feed), and a `warning: no articles returned, url: <URL>` line when the feed
parsed successfully but returned no articles. Healthy feeds SHALL produce no
output. A later refresh pass replaces the recorded outcomes, and a refresh
that fails to start leaves the previous pass's outcomes standing. The
diagnostics SHALL print after the terminal is restored so they never disturb
the live display.

#### Scenario: HTTP failure reported as an error line

- **WHEN** a refresh pass fetched `https://example.com/rss` and the server answered 404, and the TUI exits
- **THEN** stderr contains `yerss: error: 404, url: https://example.com/rss` (prefixed with the running program's basename)

#### Scenario: Non-HTTP failure reported with the error text

- **WHEN** a refresh pass failed to fetch a URL with a non-HTTP error (for example a connection refusal), and the TUI exits
- **THEN** stderr contains an `<basename>: error: <error text>, url: <URL>` line for that URL

#### Scenario: Empty feed reported as a warning line

- **WHEN** a refresh pass parsed `https://example.com/rss` successfully but the feed returned no articles, and the TUI exits
- **THEN** stderr contains `<basename>: warning: no articles returned, url: https://example.com/rss`

#### Scenario: Healthy feeds are silent

- **WHEN** a refresh pass fetched and parsed a feed that returned articles, and the TUI exits
- **THEN** no diagnostics line is printed for that URL

#### Scenario: Diagnostics reflect the most recent pass

- **WHEN** a URL failed in an earlier refresh pass but succeeded in the most recent one, and the TUI exits
- **THEN** no diagnostics line is printed for that URL
