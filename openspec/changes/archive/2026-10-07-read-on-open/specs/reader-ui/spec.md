## RENAMED Requirements

- FROM: `### Requirement: Read-status on open and scroll`
- TO: `### Requirement: Read-status on open and movement`

## MODIFIED Requirements

### Requirement: Read-status on open and movement

The system SHALL mark an article as read when the user opens it and the entire
article content fits within the viewport, since no scrolling is possible. If
the article content does not fit entirely within the viewport, the system SHALL
mark it read upon the user's first scroll or jump action through the article —
a line, page, or half-page scroll in either direction, a jump to the top or
bottom, or a mouse-wheel scroll in either direction — regardless of whether the
action actually changes the scroll position. Once marked read, further movement
SHALL NOT change the read status. Restoring a previously saved scroll offset
when an article is re-opened SHALL NOT count as movement.

#### Scenario: Short article marked read on open

- **WHEN** the user opens an article whose full content fits in the viewport
- **THEN** the article's read status is set to true immediately

#### Scenario: Long article not marked read on open

- **WHEN** the user opens an article whose content exceeds the viewport height
- **THEN** the article's read status remains false

#### Scenario: Long article marked read on first movement

- **WHEN** a long article is open and the user scrolls in either direction, or
  jumps to the top or bottom
- **THEN** the article's read status is set to true immediately

#### Scenario: A scroll action marks read regardless of position change

- **WHEN** the user issues a scroll action while reading a long article
- **THEN** the article is marked read even if the scroll position does not
  actually change

#### Scenario: Restoring a saved offset is not movement

- **WHEN** the app re-opens an article at a previously saved scroll offset
- **THEN** the article is not marked read until the user scrolls or jumps

#### Scenario: Further movement does not change read status

- **WHEN** an article is already read and the user scrolls it
- **THEN** the article remains read
