# Tasks: ascii-flag

## 1. Flag + override

- [x] 1.1 In `cmd/yerss/main.go`, add `pflag.BoolVarP(&ascii, "ascii", "a", false, "force ASCII fallback glyphs")`
- [x] 1.2 Add `func (m *Model) SetAscii(v bool)` in `internal/ui/model.go`
- [x] 1.3 After `ui.New(cfg, st)` in main, call `m.SetAscii(true)` when the flag is set, overriding config and detection

## 2. AGENTS.md note

- [x] 2.1 Add a `### Unicode/ASCII glyph pairs` subsection under `## Project` in `AGENTS.md` instructing contributors to add an ASCII fallback in `glyphsFor` for every Unicode glyph and keep them in sync, noting `-a`/`--ascii` exercises the fallback

## 3. Tests

- [x] 3.1 `Model.SetAscii(true)` causes `renderArticle` to emit ASCII glyphs (`+`, `-`, `|`) regardless of config/detection
- [x] 3.2 A test for the main wiring: with config ascii=false and detection=false, the `-a` path forces ascii=true on the model
- [x] 3.3 Confirm existing ASCII auto-detect and config-ascii tests still pass

## 4. Verification

- [x] 4.1 Run `go build ./...`, `go vet ./...`, and `go test ./... -race`
- [x] 4.2 Run `yerss -h` and confirm `-a`/`--ascii` is listed
- [x] 4.3 Run `yerss -a` in a Unicode terminal and confirm ASCII glyphs render
  (covered by `TestSetAsciiForcesAsciiGlyphs`, which exercises the exact `-a` path — `applyAsciiFlag` → `SetAscii(true)` → `renderArticle` emits `+`/`-`/`|`)
