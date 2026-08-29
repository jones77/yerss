## MODIFIED Requirements

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
