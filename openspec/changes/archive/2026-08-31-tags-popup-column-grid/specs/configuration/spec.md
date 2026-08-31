## MODIFIED Requirements

### Requirement: Database initialization

The system SHALL accept a `-z` / `--init-db` command-line flag that
reinitializes the database to zero before starting: the database file and its
WAL/SHM sidecars are deleted so the next open creates a fresh store. A missing
database is not an error (the first run simply creates it fresh). When the flag
is set, the system SHALL print a notice to stderr naming the reinitialized
database path before proceeding with normal startup. Existing databases
created before tag names were case-insensitive may contain case-duplicate
tags; reinitializing with `-z` clears them, and development databases may be
wiped at any time since the stored data is transient.

#### Scenario: Init-db wipes the database

- **WHEN** the user runs `yerss -z` and the database file and WAL sidecar exist
- **THEN** the database file and sidecars are deleted, a fresh database is created, and a notice naming the database path is printed to stderr

#### Scenario: Init-db with no existing database

- **WHEN** the user runs `yerss -z` and no database file exists
- **THEN** the command proceeds without error and creates a fresh database