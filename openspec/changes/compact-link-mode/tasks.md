## 1. Body link color

- [ ] 1.1 Set `cfg.LinkText.Color` and `cfg.Link.Color` in `GlamourStyleConfig` to the blue role (ANSI 12 dark / 4 light), keeping the existing underline-off and heading tweaks intact
- [ ] 1.2 Add a test asserting a rendered body link's name and URL spans carry the blue color code in both dark and light themes

## 2. Compact-mode href stripping

- [ ] 2.1 Add a `compose.StripLinkHrefs`-style ANSI walker (reusing `ScanANSI`) that drops each link's href span — the span following a name span with the same OSC 8 target — and its leading space
- [ ] 2.2 Unit-test the stripper: single link, two links to the same URL on one line, wide characters, BEL and ESC-backslash terminators, name-text-equals-URL edge case, and link-less text passes through unchanged
- [ ] 2.3 Assert on compact output that `parseLinkSpans` reports exactly the name span (start/end/URL), proving clickability survives

## 3. Toggle state and recompose

- [ ] 3.1 Add `linkExpand bool` (default false) to the `Model`, persist it across article switches, and join it into `derivationKey`
- [ ] 3.2 Run the compact stripping inside `computeDerivation` on body segments when compact (leaving `renderMarkdown` untouched for titles and popups)
- [ ] 3.3 Add the `LinkExpandToggle` action bound to `x` in `ViewArticle` to the config catalog; wire it in `updateArticle` to flip the flag and recompose at the same scroll offset
- [ ] 3.4 Test: pressing `x` toggles name-only ↔ name+URL rendering, keeps the scroll position, and the mode persists across article switches (default compact)

## 4. Header URL wrap and blue role

- [ ] 4.1 Wrap the stripped header URL at the content width (one OSC 8 span per wrapped line) and render each line through `m.styles.status`, keeping the full URL as the hyperlink target
- [ ] 4.2 Update/extend header tests: URL wraps within the content width for a long URL, renders blue, appears exactly once, and stays OSC 8 clickable

## 5. Verification

- [ ] 5.1 Run `go test ./...` and the linter; confirm the existing left-click link test, OSC 8 tests, and list-view `x` expand/collapse behavior still pass
- [ ] 5.2 Exercise both modes in a terminal: compact links are left- and Cmd/Ctrl-clickable and blue; `x` reveals URLs; header URL wraps and is blue