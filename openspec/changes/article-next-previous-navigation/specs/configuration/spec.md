## MODIFIED Requirements

### Requirement: Configurable keybindings

The system SHALL allow every keybinding to be remapped via the TOML config file. Keybindings SHALL be defined as a mapping from action name to a list of key strings. The system SHALL validate that no key is bound to more than one action and SHALL report an error if a conflict is detected at load time. Keybinding parsing SHALL be case-sensitive: a lowercase letter and its uppercase counterpart are distinct keys, so both MAY be bound to the same action without conflict. The default keybindings SHALL include `b` as an alias for the `back` action (alongside `esc` and `h`) and `o` as an alias for the `open_article` action in the list view (alongside `enter` and `l`). In the article view, `enter` SHALL be bound to the `close_article` action (closing the article and returning to the list), and `o` SHALL remain bound to the `open_url` action (opening the article URL in the browser); the per-view key resolution system ensures `o` resolves to `open_article` in the list view and `open_url` in the article view without conflict. Shifted-letter default bindings SHALL be reachable via both cases whenever the other case is not already bound to a different action: the tag popup SHALL be bound to both `T` and `t`, and refresh SHALL be bound to `R`, `r`, `ctrl+r`, and `f5`. The `bottom` action SHALL remain bound to uppercase `G` only, because lowercase `g` is already bound to `top`. The `copy_article_text` action SHALL be bound to uppercase `C` only, because lowercase `c` is already bound to `copy_url`. In the article view, the `next_article` action SHALL be bound to uppercase `K` and the `previous_article` action to uppercase `J`, each distinct by case from the lowercase `k`/`j` `move_up`/`move_down` scroll bindings. The `expand_toggle` action SHALL be bound to lowercase `x` in the list view, and the `link_expand_toggle` action SHALL be bound to lowercase `x` in the article view; the per-view key resolution system ensures `x` resolves to `expand_toggle` in the list view and `link_expand_toggle` in the article view without conflict.

#### Scenario: Custom keybinding

- **WHEN** the config file maps the "quit" action to `["ctrl+x"]`
- **THEN** pressing Ctrl-X quits the application instead of the default `q`

#### Scenario: Conflicting keybindings rejected

- **WHEN** the config file binds the same key to two different actions
- **THEN** the system reports an error at load time and does not start

#### Scenario: Default b alias for back

- **WHEN** no custom keybindings are configured and the user presses `b` in the article view
- **THEN** the article view returns to the list view, identical to pressing `esc` or `h`

#### Scenario: Default enter alias for close_article

- **WHEN** no custom keybindings are configured and the user presses `enter` in the article view
- **THEN** the article view returns to the list view, identical to pressing `esc`, `h`, or `b`

#### Scenario: Default o alias for open_article in list

- **WHEN** no custom keybindings are configured and the user presses `o` on an article row in the list view
- **THEN** the article reader view opens, identical to pressing `enter` or `l`

#### Scenario: Lowercase and uppercase both open the tag popup

- **WHEN** no custom keybindings are configured and the user presses either `t` or `T` in the list view
- **THEN** the tag popup opens

#### Scenario: Lowercase and uppercase both refresh

- **WHEN** no custom keybindings are configured and the user presses `r`, `R`, `ctrl+r`, or `f5` in the list view
- **THEN** a manual refresh is triggered

#### Scenario: Lowercase g stays bound to top

- **WHEN** no custom keybindings are configured and the user presses `1` or `g`
- **THEN** the cursor moves to the top, and `G` moves to the bottom (the two cases remain distinct)

#### Scenario: Uppercase C copies article text

- **WHEN** no custom keybindings are configured and the user presses `C` in the article view
- **THEN** the article text is copied to the clipboard, and lowercase `c` continues to copy the article URL

#### Scenario: Uppercase J and K navigate between articles

- **WHEN** no custom keybindings are configured and the user presses `J` or `K` in the article view
- **THEN** the previous (`J`) or next (`K`) article opens, while lowercase `j`/`k` continue to scroll the article by one line

#### Scenario: Expand toggle is bound to x

- **WHEN** no custom keybindings are configured and the user presses `x` in the
  list view
- **THEN** the `expand_toggle` action runs, expanding or collapsing all day groups

#### Scenario: Link expand toggle is bound to x in the article view

- **WHEN** no custom keybindings are configured and the user presses `x` in the article view
- **THEN** the `link_expand_toggle` action runs, toggling between compact and expanded link display
