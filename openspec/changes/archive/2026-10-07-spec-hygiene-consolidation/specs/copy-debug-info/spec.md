## MODIFIED Requirements

### Requirement: Copy Debug Info is key-bound and configurable

The `Copy Debug Info` command SHALL be registered in the keybinding catalog so it is assignable via configuration and documented in help. Its default binding SHALL be uppercase `D` in the article view.

#### Scenario: Command is bound

- **WHEN** no custom keybindings are configured and the user presses `D` in the article view
- **THEN** the system runs the `Copy Debug Info` command

#### Scenario: Command is remappable

- **WHEN** the config file rebinds the `copy_debug_info` action to another key and the user presses that key in the article view
- **THEN** the system runs the `Copy Debug Info` command
