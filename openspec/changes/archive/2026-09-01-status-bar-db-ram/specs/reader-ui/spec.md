## MODIFIED Requirements

### Requirement: List view status bar

The system SHALL display a status bar at the bottom of the list view. The
left-most elements SHALL be the time and long-form date the feeds were last
refreshed, formatted as `HH:MM Weekday Day-ordinal Month, Year last refresh`
(for example `14:30 Saturday 29th August, 2026 last refresh`), with the
`last refresh` label rendered in the dim role (#707070), or the text
`never refreshed` when no refresh has occurred. The right-most elements SHALL
be a literal `?: help` hint, followed by the bullet and the session's current
in-memory image footprint as `RAM <n>MB`, the bullet and the on-disk size of the
article database as `DB <n>MB`, the scroll position percentage, and the article
count as `<n>/<total>` — where the memory and database sizes are each rendered
as a whole number of megabytes (floored to an integer, with no space between
the number and `MB`) and `<total>` is the total number of articles in the
database — joined by the middle-dot bullet (`·`, ASCII `.`) (for example
`?: help · RAM 45MB · DB 320MB · 100% · 2/2`). The `RAM` value SHALL be the
session's total in-memory decoded-image bytes; the `DB` value SHALL be the
on-disk database size including the WAL sidecar file when one is present.
Values below one megabyte SHALL render as `0MB`. The bullets and the
`last refresh` label SHALL render in the dim role (#707070); the remaining
status bar text SHALL render in the chrome role. When a tag filter is active,
the status bar SHALL also display the active filter name.

#### Scenario: Status bar leads the right side with the help hint

- **WHEN** the list view is displayed
- **THEN** the status bar's right side reads `?: help · RAM <n>MB · DB <n>MB ·
  <percent>% · <n>/<total>`, with `?: help` as the first right-aligned element

#### Scenario: Status bar shows refresh time, memory, size, percentage, and position

- **WHEN** the list view is displayed
- **THEN** the status bar shows the last-refresh time and date on the left as
  `HH:MM Weekday Day-ordinal Month, Year last refresh` and, on the right, the
  session's in-memory image footprint as `RAM <n>MB`, the database size as
  `DB <n>MB`, the percentage, and the `<n>/<total>` position joined by ` · `

#### Scenario: Status bar uses dim bullets and dim refresh label

- **WHEN** the list view is displayed
- **THEN** the bullets between the memory, database size, percentage, and
  position render in the dim role (#707070), the `last refresh` label renders
  in the dim role, and the elements between the bullets render in the chrome
  role

#### Scenario: Status bar shows active filter

- **WHEN** a tag filter is active on the list view
- **THEN** the status bar displays the filtered tag name alongside the refresh
  time, the memory and database sizes, the percentage, and the
  `<n>/<total>` position

#### Scenario: Database size includes WAL sidecar

- **WHEN** the database has a WAL sidecar file containing uncheckpointed data
  and the list view is displayed
- **THEN** the `DB` value reports a size that includes both the main database
  file and the WAL sidecar file

#### Scenario: Memory and database sizes are whole megabytes

- **WHEN** the database file is 1_500_000 bytes and the decoded-image cache
  holds 47_000_000 bytes
- **THEN** the status bar displays `RAM 44MB` and `DB 1MB`, each as a floored
  whole number with no space between the number and `MB`

#### Scenario: Memory size reads zero when the cache is empty

- **WHEN** no images have been decoded in the session and the list view is
  displayed
- **THEN** the status bar displays `RAM 0MB`

#### Scenario: Database size refreshed after a refresh

- **WHEN** a refresh completes and new articles are stored
- **THEN** the status bar `DB` value reflects the post-refresh on-disk
  footprint on the next list load
