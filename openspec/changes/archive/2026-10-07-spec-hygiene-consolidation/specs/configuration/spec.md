## MODIFIED Requirements

### Requirement: Configurable keybindings

The system SHALL allow every keybinding to be remapped via the TOML config file. Keybindings SHALL be defined as a mapping from action name to a list of key strings. The system SHALL validate that no key is bound to more than one action and SHALL report an error if a conflict is detected at load time. Keybinding parsing SHALL be case-sensitive: a lowercase letter and its uppercase counterpart are distinct keys, so both MAY be bound to the same action without conflict. The default keybindings SHALL include `b` as an alias for the `back` action (alongside `esc` and `h`) and `o` as an alias for the `open_article` action in the list view (alongside `enter` and `l`). In the article view, `enter` SHALL be bound to the `close_article` action (closing the article and returning to the list), and `o` SHALL remain bound to the `open_url` action (opening the article URL in the browser); the per-view key resolution system ensures `o` resolves to `open_article` in the list view and `open_url` in the article view without conflict. Shifted-letter default bindings SHALL be reachable via both cases whenever the other case is not already bound to a different action: the tag popup SHALL be bound to both `T` and `t`, and refresh SHALL be bound to `R`, `r`, `ctrl+r`, and `f5`. The `bottom` action SHALL remain bound to uppercase `G` only, because lowercase `g` is already bound to `top`. The `copy_article_text` action SHALL be bound to uppercase `C` only, because lowercase `c` is already bound to `copy_url`. The `expand_toggle` action SHALL be bound to lowercase `x` in the list view, and the `link_expand_toggle` action SHALL be bound to lowercase `x` in the article view; the per-view key resolution system ensures `x` resolves to `expand_toggle` in the list view and `link_expand_toggle` in the article view without conflict.

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

#### Scenario: Expand toggle is bound to x

- **WHEN** no custom keybindings are configured and the user presses `x` in the list view
- **THEN** the `expand_toggle` action runs, expanding or collapsing all day groups

#### Scenario: Link expand toggle is bound to x in the article view

- **WHEN** no custom keybindings are configured and the user presses `x` in the article view
- **THEN** the `link_expand_toggle` action runs, toggling between compact and expanded link display

### Requirement: ASCII border fallback

The system SHALL support a plain-ASCII fallback mode. It SHALL be enabled by the `ascii` option in the TOML config file, and the system SHALL also automatically enable ASCII fallback if the terminal does not support Unicode box-drawing characters. Additionally, the system SHALL accept a `-a` / `--ascii` command-line flag that forces ASCII fallback mode, overriding both the config setting and terminal detection for the duration of that run. When ASCII fallback is in effect, box-drawing characters SHALL be replaced with ASCII equivalents and article lead image rendering SHALL be disabled (halfblock image glyphs are Unicode and incompatible with ASCII-only mode); articles SHALL render with `[alt]`/`[image]` placeholders for inline content images and no lead image block.

#### Scenario: ASCII fallback enabled in config

- **WHEN** the config file sets ascii to true
- **THEN** the article reader border uses ASCII characters — `-` for horizontal borders, `:` for vertical borders, `+` for corners and tees, and `|` for the scrollbar thumb — instead of Unicode box-drawing glyphs, and lead images are not rendered

#### Scenario: ASCII fallback auto-detected

- **WHEN** the terminal does not support Unicode box-drawing characters and ascii is not explicitly configured
- **THEN** the system automatically enables ASCII fallback for border rendering and disables lead image rendering

#### Scenario: ASCII flag forces fallback and overrides config

- **WHEN** the user runs `yerss -a` (or `yerss --ascii`) and the config sets ascii to false and the terminal reports Unicode support
- **THEN** the system renders with ASCII fallback glyphs and no lead images for that run, overriding both the config and terminal detection

#### Scenario: Help flag lists the ascii flag

- **WHEN** the user runs `yerss -h` or `yerss --help`
- **THEN** the usage lists `-a`/`--ascii` and exits 0 without entering the TUI
