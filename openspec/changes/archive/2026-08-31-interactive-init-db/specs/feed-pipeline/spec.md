## MODIFIED Requirements

### Requirement: Usage output documents the data files

The system SHALL document its three data files in the CLI usage output: the
TOML config file, the feeds.txt file, and the SQLite database — each listed with
its resolved default path, and a note that the config's `[data]` section can
relocate the feeds file and database. The usage SHALL list only the resolved
paths; it SHALL NOT attach a description to each file, because the file names
and paths are self-explanatory.

#### Scenario: Usage lists the data files

- **WHEN** the user runs the program with the help flag
- **THEN** the usage output lists the resolved config, feeds, and database paths (without per-file descriptions) and the `[data]` relocation note
