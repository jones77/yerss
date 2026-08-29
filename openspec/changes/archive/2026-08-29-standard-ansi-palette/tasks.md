## 1. Palette restructuring

- [x] 1.1 Replace the palette values in `internal/ui/theme.go` with standard
      ANSI colors: `Dim` → 8, `StatusBar` → 12 (dark) / 4 (light), remove
      `Accent` and `Border`, rename `Bold` → `Bright` (15 dark / 0 light), add
      `Text` (7 dark / 0 light)
- [x] 1.2 Update `resolvePalette`/`detectLightBackground` if needed and confirm
      `auto`/`dark`/`light` all build

## 2. Update palette consumers

- [x] 2.1 `internal/ui/border.go`: move `Border` references to `Dim`, keep
      `StatusBar` for inline text/thumb, update the "baby blue" comments to
      "bright blue (12)"
- [x] 2.2 `internal/ui/popup.go`: replace `Accent` references (title text,
      border foreground) with `StatusBar`; replace `Bold` with `Bright`
- [x] 2.3 `internal/ui/list.go`: rename `Bold` → `Bright` for unread titles;
      render the article-row tree rail in the grey role (`Dim`); keep the
      `#333333` selection background
- [x] 2.4 `internal/ui/image.go`: `Dim` reference unchanged but re-verify the
      attribution still renders grey
- [x] 2.5 Update `internal/ui/polish_test.go` references to renamed fields and
      verify escape sequences still match (ANSI numbers vs hex)

## 3. Body text via glamour

- [x] 3.1 In `internal/ui/markdown.go`, override the base document foreground
      to ANSI 7 (dark) / 0 (light) in `glamourStyleConfig`; verify the
      `Document.Color` field takes the ANSI string and styled elements keep
      their theme colors

## 4. Verification

- [x] 4.1 Run the full test suite (`go test ./...`) and the linter; confirm
      nothing references the removed `Accent`/`Border` fields
- [x] 4.2 Run `openspec validate` and confirm the change is complete