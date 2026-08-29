## MODIFIED Requirements

### Requirement: List view status bar

The system SHALL display a status bar at the bottom of the list view showing the
article count as `<n>/<total>`, where `n` is the number of articles currently
shown (after any active tag filter) and `total` is the total number of articles
in the database, followed by the date the feeds were last refreshed. The status
bar SHALL also display the on-disk size of the article database, formatted as a
human-readable value with binary unit suffixes (B, KB, MB, GB), on the same side
of the status bar as the last-refresh date. The reported size SHALL include the
WAL sidecar file when one is present, so it reflects the actual on-disk
footprint. When a tag filter is active, the status bar SHALL also display the
active filter name.

#### Scenario: Status bar shows filtered/total count, size, and refresh date

- **WHEN** the list view is displayed
- **THEN** the status bar shows the article count as `<n>/<total>` (shown over total), the database size as a human-readable value, and the last-refreshed date

#### Scenario: Status bar shows active filter

- **WHEN** a tag filter is active on the list view
- **THEN** the status bar displays the filtered tag name alongside the `<n>/<total>` count, the database size, and the date

#### Scenario: Database size includes WAL sidecar

- **WHEN** the database has a WAL sidecar file containing uncheckpointed data and the list view is displayed
- **THEN** the status bar reports a size that includes both the main database file and the WAL sidecar file

#### Scenario: Database size formatted with binary unit suffix

- **WHEN** the database file is 1_500_000 bytes
- **THEN** the status bar displays the size as `1.4 MB`

#### Scenario: Database size refreshed after a refresh

- **WHEN** a refresh completes and new articles are stored
- **THEN** the status bar database size reflects the post-refresh on-disk footprint on the next list load
