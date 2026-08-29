## MODIFIED Requirements

### Requirement: Configurable keybindings

The system SHALL allow every keybinding to be remapped via the TOML config file. Keybindings SHALL be defined as a mapping from action name to a list of key strings. The system SHALL validate that no key is bound to more than one action and SHALL report an error if a conflict is detected at load time. The default keybindings SHALL include `b` as an alias for the `back` action (alongside `esc`, `enter`, and `h`) and `o` as an alias for the `open_article` action in the list view (alongside `enter` and `l`). In the article view, `o` SHALL remain bound to the `open_url` action (opening the article URL in the browser); the per-view key resolution system ensures `o` resolves to `open_article` in the list view and `open_url` in the article view without conflict.

#### Scenario: Custom keybinding

- **WHEN** the config file maps the "quit" action to `["ctrl+x"]`
- **THEN** pressing Ctrl-X quits the application instead of the default `q`

#### Scenario: Conflicting keybindings rejected

- **WHEN** the config file binds the same key to two different actions
- **THEN** the system reports an error at load time and does not start

#### Scenario: Default b alias for back

- **WHEN** no custom keybindings are configured and the user presses `b` in the article view
- **THEN** the article view returns to the list view, identical to pressing `esc` or `h`

#### Scenario: Default o alias for open_article in list

- **WHEN** no custom keybindings are configured and the user presses `o` on an article row in the list view
- **THEN** the article reader view opens, identical to pressing `enter` or `l`
