## Why

Bug reports for this TUI are hard to reproduce because the rendering bugs depend on exact screen dimensions, scroll position, and article geometry. When a user hits a rendering problem (e.g. a link whose text runs together at 106x31), they currently have to manually transcribe their terminal size and scroll position, which is error-prone and incomplete.

## What Changes

- Add a new `Copy Debug Info` command bound to a key in the article view that copies a structured debug snapshot to the system clipboard.
- The snapshot includes: terminal width × height (e.g. `106x31`), article scroll position (yOffset and total rendered line count), the article URL and title, and rendering mode (ASCII vs Unicode, expanded vs compact link mode, image rendering mode if determinable).
- Reuse the existing clipboard mechanism (`CopyTextCmd`) so the command works identically on macOS (`pbcopy`), Linux (`xclip`), and Windows (`clip`).
- Register the new action in the keybinding catalog so it automatically appears in the config template, help text, and keybinding self-update registry.

## Capabilities

### New Capabilities
- `copy-debug-info`: captures and copies a diagnostic snapshot (terminal size, scroll position, article identity, and rendering mode flags) to the system clipboard from the article view.

### Modified Capabilities
<!-- none -->

## Impact

- `internal/config/keys.go`: new `Action` constant and `catalog` entry (auto-propagates to help/template/registry).
- `internal/ui/articleview/update.go`: new `Action` case in `UpdateKey` dispatch.
- `internal/ui/articleview/`: new command/format function that assembles the snapshot string and calls the existing `CopyTextCmd`.
- No new dependencies; reuses existing clipboard command and viewport/state fields.
