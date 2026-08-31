## MODIFIED Requirements

### Requirement: Configurable keybindings

The system SHALL allow every keybinding to be remapped via the TOML config file. Keybindings SHALL be defined as a mapping from action name to a list of key strings. The system SHALL validate that no key is bound to more than one action and SHALL report an error if a conflict is detected at load time. Keybinding parsing SHALL be case-sensitive: a lowercase letter and its uppercase counterpart are distinct keys, so both MAY be bound to the same action without conflict. The default keybindings SHALL include `b` as an alias for the `back` action (alongside `esc` and `h`) and `o` as an alias for the `open_article` action in the list view (alongside `enter` and `l`). In the article view, `enter` SHALL be bound to the `close_article` action (closing the article and returning to the list), and `o` SHALL remain bound to the `open_url` action (opening the article URL in the browser); the per-view key resolution system ensures `o` resolves to `open_article` in the list view and `open_url` in the article view without conflict. Shifted-letter default bindings SHALL be reachable via both cases whenever the other case is not already bound to a different action: the tag popup SHALL be bound to both `T` and `t`, and refresh SHALL be bound to both `R` and `r`. The `bottom` action SHALL remain bound to uppercase `G` only, because lowercase `g` is already bound to `top`. The `copy_article_text` action SHALL be bound to uppercase `C` only, because lowercase `c` is already bound to `copy_url`. The `expand_toggle` action SHALL be bound to lowercase `x` in the list view.

#### Scenario: Expand toggle is bound to x

- **WHEN** no custom keybindings are configured and the user presses `x` in the
  list view
- **THEN** the `expand_toggle` action runs, expanding or collapsing all day groups