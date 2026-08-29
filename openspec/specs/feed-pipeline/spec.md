## Purpose

Handles fetching RSS/Atom feeds, the refresh lifecycle (startup gate and manual cooldown), article deduplication, and persistent storage of all downloaded articles in a local SQLite database.

## Requirements

### Requirement: Feed URL source

The system SHALL read feed URLs from a plain-text file with one URL per line. The default path SHALL be `$XDG_CONFIG_HOME/yerss/feeds.txt` (fallback `~/.config/yerss/feeds.txt`). The path SHALL be configurable via the TOML config file.

#### Scenario: Feeds loaded from default path

- **WHEN** the application starts and no custom feeds path is configured
- **THEN** the system reads feed URLs from `$XDG_CONFIG_HOME/yerss/feeds.txt` or `~/.config/yerss/feeds.txt` as fallback

#### Scenario: Custom feeds path

- **WHEN** the TOML config specifies a custom feeds file path
- **THEN** the system reads feed URLs from that path instead of the default

### Requirement: Startup refresh with 15-minute gate

The system SHALL attempt to refresh all feeds on startup only if more than 15 minutes have elapsed since the last successful refresh. If the last refresh was within 15 minutes, the system SHALL load articles from the local database without fetching from the network.

#### Scenario: Startup refresh after 15 minutes

- **WHEN** the application starts and the last refresh was more than 15 minutes ago
- **THEN** the system fetches all feeds from the network and updates the database

#### Scenario: Startup skip refresh within 15 minutes

- **WHEN** the application starts and the last refresh was less than 15 minutes ago
- **THEN** the system loads articles from the database without network fetching

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

The system SHALL persist every downloaded article in the SQLite database. Articles SHALL survive application restarts. Read status SHALL be persisted and restored on the next launch.

#### Scenario: Articles survive restart

- **WHEN** the application is restarted after fetching articles
- **THEN** all previously fetched articles are loaded from the database

#### Scenario: Read status persisted

- **WHEN** an article is marked as read and the application is restarted
- **THEN** the article retains its read status on the next launch

### Requirement: Automatic tagging from feed categories

The system SHALL extract categories from each article using the feed's category metadata (as provided by gofeed). Each unique category name SHALL be stored as a tag. Articles SHALL be associated with their categories in the database, enabling tag-based filtering and browsing.

#### Scenario: Article with categories tagged

- **WHEN** a feed is fetched and an article has categories "tech" and "news"
- **THEN** the article is associated with tags "tech" and "news" in the database

#### Scenario: Article with no categories

- **WHEN** a feed is fetched and an article has no categories
- **THEN** the article has no tag associations but is still stored and displayed

### Requirement: Startup feed-verification gate

Before launching the terminal UI, the system SHALL run a feed-verification gate.
The gate SHALL read the configured feeds file. If the feeds file is missing or
contains no URL lines (only blanks and comments), the system SHALL print a
message naming the feeds file path and SHALL exit with code 2 without entering
the TUI. If the feeds table contains at least one previously-verified feed (a
feed that has been successfully fetched and persisted in the past), the system
SHALL proceed to the TUI without performing a network check. Otherwise the
system SHALL fetch the configured feeds concurrently, bounded by a 10-second
per-feed timeout and a 30-second overall deadline. While the verification fetch
is in progress, the system SHALL print a `verifying feeds...` notice to stderr.
If at least one feed parses successfully, the system SHALL proceed to the TUI
and the successful fetch SHALL count as the first refresh (so the startup
refresh gate SHALL skip a redundant fetch). If no feed parses successfully, the
system SHALL print the per-feed errors and SHALL exit with code 3 without
entering the TUI.

#### Scenario: Empty feeds file refuses start

- **WHEN** the feeds file exists but contains only blank lines and comments
- **THEN** the system prints a message naming the feeds file path and exits with code 2 without entering the TUI

#### Scenario: Missing feeds file refuses start

- **WHEN** the feeds file does not exist at the resolved path
- **THEN** the system prints a message naming the feeds file path and exits with code 2 without entering the TUI

#### Scenario: Previously-verified feed skips the network check

- **WHEN** the feeds file contains at least one URL and the feeds table already contains a previously-verified feed
- **THEN** the system proceeds to the TUI without performing any network fetch during the gate

#### Scenario: First run with a working feed starts

- **WHEN** the feeds file contains at least one URL, the feeds table is empty, and at least one configured feed parses successfully within the per-feed and overall deadlines
- **THEN** the system persists the verified feed, proceeds to the TUI, and the startup refresh gate skips a redundant fetch because a refresh just occurred

#### Scenario: First run with no working feed refuses start

- **WHEN** the feeds file contains at least one URL, the feeds table is empty, and no configured feed parses successfully within the per-feed and overall deadlines
- **THEN** the system prints the per-feed errors and exits with code 3 without entering the TUI

#### Scenario: Verifying notice printed to stderr

- **WHEN** the gate performs a verification fetch (the feeds table is empty)
- **THEN** the system prints a `verifying feeds...` notice to stderr before the fetch begins