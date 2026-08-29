## 1. Clipboard helper refactor

- [x] 1.1 Generalize the clipboard copy in `internal/ui/open.go` into `copyTextCmd(text string) tea.Cmd` that pipes arbitrary text to the platform clipboard helper via stdin (`pbcopy`/`clip`/`xclip -selection clipboard`); make `copyURLCmd` a thin wrapper over it
- [x] 1.2 Extend the status message for text copies so the status line reads e.g. "copied article text" / "copied selection" instead of "copied URL" (extend `urlActionMsg` handling in `internal/ui/model.go`)

## 2. Config: copy_article_text action

- [x] 2.1 Add `CopyArticleText Action = "copy_article_text"` to `internal/config/keys.go`, with default binding `{"C"}`, registered in `allActions` and `actionsForView(ViewArticle)`
- [x] 2.2 Confirm `C` resolves without conflict in the article view (distinct from lowercase `c` → `copy_url`) and add/extend a config test for the default `C` binding

## 3. Selection state and mouse handling

- [x] 3.1 Add a selection struct (anchor cell + current cell in viewport coordinates) to `articleState` in `internal/ui/article.go`
- [x] 3.2 Map mouse coordinates to viewport cells in `updateArticleMouse` (`x = X - 1 - padX`, `y = Y - 1 - effPadY`, clamped to content bounds); ignore presses outside the content area and any event with a non-none modifier
- [x] 3.3 Implement press/drag/release handling: unmodified left-press anchors, motion extends, release finalizes; change `updateArticleMouse` to return a `tea.Cmd` and wire it through `Model.Update`'s `MouseMsg` case
- [x] 3.4 Clear the selection when leaving the article view (back/list navigation)

## 4. Selection rendering

- [x] 4.1 In `renderArticle`, apply an inverted lipgloss style to the selected cell range on each visible content line using `ansi.Cut` (before/selected/after), so ANSI/OSC 8 bytes stay intact
- [x] 4.2 Verify the highlight renders correctly in both Unicode and `--ascii` modes and does not disturb border/scrollbar layout

## 5. Copy behavior

- [x] 5.1 On release, extract the selected text: `ansi.Strip` each line in the row range, slice the cell range with `ansi.Cut`, trim trailing padding spaces, join with `\n`, and copy via `copyTextCmd`; zero-width selection copies nothing
- [x] 5.2 Add the `C` key dispatch in `updateArticle`: `config.CopyArticleText` copies `ansi.Strip(renderArticleContent(*article))` via `copyTextCmd`

## 6. Tests and verification

- [x] 6.1 Add tests for selection mapping and highlight rendering: anchor press, drag extend, release copies, modifier-click is a no-op, zero-width copies nothing, selection clears on leaving the article
- [x] 6.2 Add a test that copied selection text has ANSI and OSC 8 escapes stripped (including a selection spanning a link line)
- [x] 6.3 Add a test that `C` produces a copy command with the full article text and that `c` still maps to `copy_url`
- [x] 6.4 Run `go vet ./...`, `gofmt`, and `go test ./...` (including `-a`/`--ascii`-mode exercises) and confirm the full suite passes