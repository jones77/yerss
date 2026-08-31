## ADDED Requirements

### Requirement: Interactive confirmation for --init-db

The system SHALL accept an `-i` / `--interactive` command-line flag. The flag
SHALL be a no-op for every other flag and behavior; its only effect is on the
`-z` / `--init-db` flag. When both `-z` and `-i` are set, the system SHALL
prompt before deleting the database, printing the database's on-disk size and
the number of articles it currently holds, and SHALL read a confirmation from
standard input. The system SHALL delete the database only when the user
confirms with `y` or `Y`; any other response (including empty input, `n`,
`N`, or end-of-input) SHALL leave the database untouched. When standard input
is not a terminal (for example a pipe or a continuous-integration environment),
the system SHALL refuse to delete the database and exit with a non-zero status.
`-z` alone SHALL keep the existing behavior and delete the database without a
prompt. When the database file does not exist, `-z -i` SHALL proceed without a
prompt, since there is nothing to delete.

#### Scenario: -i is a no-op on its own

- **WHEN** the user runs `yerss -i` without `-z`
- **THEN** the application starts normally with no prompt and no change in behavior

#### Scenario: -z -i prompts and confirms

- **WHEN** the user runs `yerss -z -i`, the database holds 1203 articles in a 123.4 MB file, and the user answers `y`
- **THEN** the database is deleted and the application starts from zero

#### Scenario: -z -i declined leaves the database

- **WHEN** the user runs `yerss -z -i` and answers `n` (or empty, or end-of-input)
- **THEN** the database is left untouched and the application does not start with a reinitialized database

#### Scenario: -z -i on non-interactive stdin refuses

- **WHEN** the user runs `yerss -z -i` with standard input not attached to a terminal (a pipe or CI)
- **THEN** the system refuses to delete the database and exits non-zero, leaving the database untouched

#### Scenario: -z -i with no database proceeds silently

- **WHEN** the user runs `yerss -z -i` and the database file does not exist
- **THEN** the system does not prompt and starts a fresh database

#### Scenario: -z alone still deletes without prompting

- **WHEN** the user runs `yerss -z` without `-i`
- **THEN** the database is deleted with no prompt, preserving existing behavior

#### Scenario: Help flag lists the interactive flag

- **WHEN** the user runs the program with the help flag
- **THEN** the usage output lists `-i`/`--interactive` alongside the other flags
