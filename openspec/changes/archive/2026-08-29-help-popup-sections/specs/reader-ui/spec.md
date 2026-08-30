## MODIFIED Requirements

### Requirement: Help popup

The system SHALL display a help popup listing each action and its bound keys
when the user presses `?`. Each action SHALL be labeled with its display name
derived from the action identifier by replacing underscores with spaces and
rendering the result in Title Case (for example `half_page_down` →
`Half Page Down`). The action label SHALL appear on the left
and the bound key strings SHALL appear on the right. The popup SHALL present
the bindings grouped under three headings in this order: **Global**, **List
view**, and **Article view**. Actions not specific to a single view — those
valid in all views and those valid in both the list and article views — SHALL
be listed under Global. Actions valid only in the list view SHALL be listed
under List view, and actions valid only in the article view SHALL be listed
under Article view. Within each section the actions SHALL appear in catalog
order, and each section heading SHALL be rendered in the blue role. The List
view and Article view sections SHALL be laid out in a second column to the
right of the Global section, and the popup including its border SHALL NOT
exceed 70 columns wide.

#### Scenario: Action labels are Title Case without underscores

- **WHEN** the help popup is displayed
- **THEN** each action is labeled with underscores replaced by spaces and each word capitalized (for example `Open Article`, `Half Page Down`)

#### Scenario: Help popup lists current bindings

- **WHEN** the user presses `?` in any view
- **THEN** a popup appears listing each action's display name on the left and its currently bound keys on the right

#### Scenario: Help popup groups bindings by scope

- **WHEN** the user presses `?` in any view
- **THEN** the popup lists Quit, Back, Move Up, Page Down, and Tag Popup under Global, Refresh and Open Article under List view, and Link Popup, Open URL, Copy URL, and Copy Article Text under Article view

#### Scenario: Global includes all-view and shared bindings

- **WHEN** the Global section is rendered
- **THEN** it lists every action that is not specific to a single view, including actions valid in all views (such as Quit, Back, Move Up, Move Down, Help) and navigation shared by the list and article views (such as Page Up, Page Down, Top, Bottom, Tag Popup)

#### Scenario: Section headings use the blue role

- **WHEN** a section heading is rendered
- **THEN** it renders in the blue role (ANSI 12 in dark mode, ANSI 4 in light mode)

#### Scenario: List and article sections form a second column

- **WHEN** the help popup is displayed
- **THEN** the List view and Article view sections render in a second column to the right of the Global section

#### Scenario: Help popup stays within seventy columns

- **WHEN** the help popup is rendered with its default keybindings
- **THEN** the popup including its border is at most 70 columns wide