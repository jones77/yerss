## Purpose

Handles fetching RSS/Atom feeds, the refresh lifecycle (startup gate and manual cooldown), article deduplication, and persistent storage of all downloaded articles in a local SQLite database.

## ADDED Requirements

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
