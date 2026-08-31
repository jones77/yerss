## Why

The `?` help key is currently discoverable only by opening the help popup or
reading the README; nothing in the list view's status bar or the article view's
bottom border hints at it. A small `?: help` affordance as the first
right-aligned element in the bottom chrome of both views makes the help popup
discoverable at a glance.

## What Changes

- The list view status bar renders `?: help` as the first right-aligned element,
  separated from the database size by the bullet (`·`, ASCII `.`), so it reads
  `?: help · <size> · <percent>% · <n>/<total>` from left to right.
- The article view bottom border renders `?: help` as the first right-aligned
  element of the position indicator, separated from the percent by the bullet,
  so it reads `?: help · <percent>% · <bottomLine>/<totalLines>`.

## Why Not

- No keymap change: `?` already opens help in the list and article views.
- No help popup change: the popup already lists every binding by scope.