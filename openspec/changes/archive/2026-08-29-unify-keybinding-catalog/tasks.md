## 1. Catalog and derivations

- [x] 1.1 Introduce `catalog []actionSpec` in `keys.go` encoding every action's identifier, default keys, and views, in the canonical display order
- [x] 1.2 Rewrite `DefaultKeybindings()` to build its `Keymap` from `catalog`
- [x] 1.3 Rewrite `allActions` to be built from `catalog`
- [x] 1.4 Rewrite `actionsForView(v)` to filter `catalog` by `views`
- [x] 1.5 Test: `DefaultKeybindings()` output is byte-identical to the pre-refactor defaults (add a regression test if none pins it)

## 2. Seeded template and registry

- [x] 2.1 Add a helper that renders the `keybindings` section of the seeded template from `catalog` (e.g. `seededKeybindings()`)
- [x] 2.2 Update `seededConfig()` (bootstrap.go) to emit its `[keybindings]` block via that helper
- [x] 2.3 Add a helper that returns the keybinding rows of `ConfigReleases` from `catalog`, and have the drift-checked registry include them
- [x] 2.4 Test: seeded template now contains `b` (back), `o` (open_article), `t`/`T` (tag_popup), `r`/`R` (refresh), and `copy_article_text`

## 3. Help popup

- [x] 3.1 Export `config.AllActions() []Action` (or equivalent) in the config package
- [x] 3.2 Rewrite `renderHelp` (popup.go) to iterate `config.AllActions()` instead of its hardcoded slice
- [x] 3.3 Test: help popup lists every action in catalog order and shows the current bound keys

## 4. Drift test closure

- [x] 4.1 Extend the existing config drift test to assert catalog ↔ `DefaultKeybindings` ↔ seeded template ↔ `ConfigReleases` are mutually consistent
- [x] 4.2 Confirm `ParseKeybindings` / `Validate` / conflict tests still pass unchanged

## 5. Validation

- [x] 5.1 Run `go test ./...` and `go vet ./...`; fix any failures
- [x] 5.2 `openspec validate --changes unify-keybinding-catalog --strict` passes
- [x] 5.3 Manual smoke: seed a fresh config, confirm `b`/`o`/`t`/`r`/`C` appear commented out; `?` help lists all actions
