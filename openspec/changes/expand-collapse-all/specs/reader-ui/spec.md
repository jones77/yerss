## MODIFIED Requirements

### Requirement: Keyboard navigation

The system SHALL support both vi-style and arrow-key navigation. Page navigation
(PgUp/PgDn, Ctrl-F/Ctrl-B, Space) SHALL scroll by a full viewport page. Half-page
navigation (Ctrl-D/Ctrl-U) SHALL scroll by half a viewport. `1`, `g`, or
Ctrl-Up SHALL move to the top; `G` or Ctrl-Down SHALL move to the bottom. In the
list view, up/down moves selection across visible rows (day headers and expanded
articles); in the article view, up/down scrolls the article by one line. `Enter`,
`l`, or `o` opens the selected article from the list; `Esc`, `Enter`, `h`, or
`b` returns from the article view to the list, reselecting the previously viewed
article. In the list view, `Esc`, `h`, and `b` clear an active tag filter or are
no-ops when no filter is active.

When the list cursor is on a day header, fold keys SHALL act on that day:
`h` SHALL collapse it, `l` SHALL expand it, and `Enter`, Space, or Tab SHALL
toggle its expansion; the article-open behavior of `l`/Enter/`o` SHALL NOT apply
on a header row. When the cursor is on an article row, `Tab` SHALL toggle the
expansion of that article's day group, and the existing bindings (`h`/`b` clear
the filter or are a no-op, `l`/Enter/`o` open the article, Space half-pages)
SHALL apply unchanged.
The system SHALL provide an `x` key in the list view that toggles all day groups
at once: when every day group is expanded it SHALL collapse them all, and
otherwise (some or all collapsed, including a mixed state) it SHALL expand them
all.

#### Scenario: Toggle all collapses an all-expanded list

- **WHEN** every day group is expanded in the list view and the user presses `x`
- **THEN** every day group collapses and only the day headers remain visible

#### Scenario: Toggle all expands a fully collapsed list

- **WHEN** every day group is collapsed in the list view and the user presses `x`
- **THEN** every day group expands and all articles become visible

#### Scenario: Toggle all expands a mixed list

- **WHEN** some day groups are expanded and others are collapsed in the list
  view and the user presses `x`
- **THEN** every day group expands (a mixed state defaults to expand all)