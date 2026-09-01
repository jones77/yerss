## Purpose

Lets the user inspect the on-disk and estimated in-memory footprint of stored article images, so a future decision about bounding session memory is grounded in real data.

## ADDED Requirements

### Requirement: Image statistics diagnostic

The system SHALL provide an `--image-stats` command-line flag that prints a
read-only report of image storage to stdout and exits 0 without entering the
TUI, without performing any network request, and without running the startup
verification gate. Generating the report SHALL NOT insert, update, or delete
any database row.

#### Scenario: Stats report includes database and image sizes

- **WHEN** the user runs `yerss --image-stats` and the database contains stored article images
- **THEN** stdout reports the database size (including the WAL sidecar), the article count, the number of stored image rows, the total and per-image photo byte sizes (minimum, median, 90th percentile, and maximum), and the total rendered-block bytes

#### Scenario: Stats report estimates decoded memory

- **WHEN** the user runs `yerss --image-stats` and stored photos have parseable dimensions
- **THEN** stdout reports the estimated decoded size (width times height times four bytes) for each photo and the sum across all photos

#### Scenario: Stats is read-only and offline

- **WHEN** the user runs `yerss --image-stats`
- **THEN** the report is generated from the local database without fetching any feed or image and without modifying any database row

#### Scenario: Unparseable photo bytes are counted

- **WHEN** a stored photo's bytes cannot be decoded to obtain dimensions
- **THEN** the report counts it as unparseable and excludes it from the decoded-size estimate rather than failing
