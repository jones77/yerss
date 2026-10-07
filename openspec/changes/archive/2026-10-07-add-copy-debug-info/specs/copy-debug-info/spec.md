## Purpose

Lets users copy a diagnostic snapshot of the article view so rendering bugs can be reported with enough detail to reproduce them.

## ADDED Requirements

### Requirement: Copy Debug Info command
The system SHALL provide a `Copy Debug Info` command in the article view that copies a diagnostic snapshot of the current article session to the system clipboard.

#### Scenario: Copy debug snapshot
- **WHEN** the user triggers the `Copy Debug Info` command while viewing an article
- **THEN** the system copies a text snapshot to the clipboard and shows a confirmation (e.g. "copied debug info")

### Requirement: Snapshot includes terminal dimensions
The snapshot SHALL include the terminal width and height in `<width>x<height>` form (e.g. `106x31`).

#### Scenario: Dimensions present
- **WHEN** the snapshot is copied
- **THEN** it contains a line with the terminal width and height in `<width>x<height>` form

### Requirement: Snapshot includes article scroll position
The snapshot SHALL include the article's current scroll offset and the total number of rendered lines.

#### Scenario: Scroll position present
- **WHEN** the snapshot is copied
- **THEN** it contains both the current y-offset and the total rendered line count

### Requirement: Snapshot includes article identity
The snapshot SHALL include the article URL and title so the affected content can be located.

#### Scenario: Article identity present
- **WHEN** the snapshot is copied
- **THEN** it contains the article's URL and title

### Requirement: Snapshot includes rendering mode flags
The snapshot SHALL include the active rendering modes that affect layout, specifically ASCII vs Unicode glyph mode and expanded vs compact link mode.

#### Scenario: Rendering modes present
- **WHEN** the snapshot is copied
- **THEN** it indicates the current ASCII/Unicode mode and the current link mode

### Requirement: Copy Debug Info is key-bound and configurable
The `Copy Debug Info` command SHALL be registered in the keybinding catalog so it is assignable via configuration and documented in help.

#### Scenario: Command is bound
- **WHEN** the user presses the configured key in the article view
- **THEN** the system runs the `Copy Debug Info` command
