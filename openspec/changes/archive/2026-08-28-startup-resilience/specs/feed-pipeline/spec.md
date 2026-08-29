## ADDED Requirements

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
