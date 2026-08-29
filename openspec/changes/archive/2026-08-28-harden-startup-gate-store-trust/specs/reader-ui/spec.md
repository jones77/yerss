## ADDED Requirements

### Requirement: Open-article bounds safety

When the user opens an article from the list view, the system SHALL verify the
current cursor index is within the bounds of the loaded article list before
accessing it. When the list is empty or the cursor is otherwise out of range
(such as a stale cursor left over from a list that has since shrunk), the
open-article action SHALL be a no-op and SHALL NOT panic. The bounds check SHALL
not depend on any prior clamping having occurred.

#### Scenario: Open article on empty list is a no-op

- **WHEN** the article list is empty and the user presses the open-article key
- **THEN** nothing happens and the application does not panic

#### Scenario: Open article with a stale cursor does not panic

- **WHEN** the cursor index is greater than or equal to the current length of the loaded article list (for example after a refresh shrank the list) and the user presses the open-article key
- **THEN** the action is a no-op and the application does not panic
