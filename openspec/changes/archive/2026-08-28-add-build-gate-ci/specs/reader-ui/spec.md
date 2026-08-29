## ADDED Requirements

### Requirement: URL open and copy argument safety

When the user opens an article URL in the system browser or copies it to the
clipboard, the system SHALL pass the URL to the platform opener or clipboard helper
as a single exec argument without routing it through a command shell. The system
SHALL NOT invoke a shell (`cmd.exe /c` or equivalent) with the raw, feed-controlled
URL, so that metacharacters in an article's link (such as `&` or `|`) cannot inject
additional commands. Clipboard copy SHALL feed the URL to the clipboard helper via
its standard input, not via a shell command string.

#### Scenario: Open URL uses a non-shell single-argument opener on Windows

- **WHEN** the user opens an article URL on Windows and the link contains shell metacharacters such as `&` or `|`
- **THEN** the system opens the URL via a non-shell opener that receives the URL as a single argument, and no additional command is executed

#### Scenario: Open URL on Unix uses a single exec argument

- **WHEN** the user opens an article URL on macOS or Linux
- **THEN** the system passes the URL as a single exec argument to the platform opener without a shell

#### Scenario: Copy URL feeds stdin without a shell

- **WHEN** the user copies an article URL to the clipboard
- **THEN** the system pipes the URL to the clipboard helper's standard input as a single value, not via a shell command string
