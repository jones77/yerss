## Purpose

Dumps the application's stored articles as a JSON document grouped by date, for
scripting, export, and piping to other tools without launching the TUI.

## ADDED Requirements

### Requirement: JSON article dump on -j/--json

The system SHALL accept a `-j` / `--json` command-line flag that, instead of
launching the terminal UI, dumps all stored articles as a single JSON object to
stdout and exits 0. The object SHALL be keyed by publication date formatted as
`YYYYMMDD` in the user's local timezone, with keys ordered reverse
chronologically (most recent date first). Each date's value SHALL be an array of
article objects sorted reverse chronologically (most recent first). Each article
object SHALL include all stored fields: `id`, `feed_url`, `guid`, `title`,
`link`, `author`, `published_at`, `content`, `description`, `read`,
`fetched_at`, and `categories` (a JSON array of category name strings). The
system SHALL NOT perform a network fetch or run the startup verification gate in
this mode; it SHALL read directly from the SQLite store. When the database has
no articles, the system SHALL emit `{}` and exit 0. The `-j` mode SHALL take
effect after any `-c`/`--edit-config` and `-e`/`--edit-feeds` editing so those
edits apply to the dump, and SHALL exit 0 without entering the TUI.

#### Scenario: Dump grouped by date, dates reverse chronological

- **WHEN** the user runs `yerss -j` and articles exist on several days
- **THEN** the output is a JSON object keyed by `YYYYMMDD` with the most recent date first

#### Scenario: Articles within a date are reverse chronological

- **WHEN** a date bucket contains multiple articles
- **THEN** the articles in that date's array are ordered most recent first

#### Scenario: All stored fields are present per article

- **WHEN** an article object is emitted
- **THEN** it includes `id`, `feed_url`, `guid`, `title`, `link`, `author`, `published_at`, `content`, `description`, `read`, `fetched_at`, and `categories`

#### Scenario: Date keys use local timezone

- **WHEN** an article's `published_at` is midnight UTC in a timezone that is UTC-5
- **THEN** the article's date key is the previous calendar day's `YYYYMMDD` in the user's local timezone

#### Scenario: Empty database emits an empty object

- **WHEN** the user runs `yerss -j` and the database has no articles
- **THEN** the output is `{}` and the process exits 0

#### Scenario: No network or startup gate in JSON mode

- **WHEN** the user runs `yerss -j`
- **THEN** the system does not fetch feeds or run the startup verification gate; it reads from the store, prints JSON, and exits 0

#### Scenario: Help flag lists the json flag

- **WHEN** the user runs `yerss -h` or `yerss --help`
- **THEN** the usage lists `-j`/`--json` and exits 0 without entering the TUI
