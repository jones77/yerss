## Context

`renderArticleMarkdown` builds markdown (header + `convert.Convert(body)`) and
hands it to glamour. Two fallbacks and one placeholder write path currently
produce garbled output on error or on special input. See proposal.md - Why.

## Goals / Non-Goals

**Goals:**
- On conversion or render failure, show readable plain text — never raw HTML
  tags or literal `**`/`[text](url)` syntax.
- Escape image alt text so it renders as the literal `[alt]` the spec requires.

**Non-Goals:**
- No change to the happy-path rendering (styling, wrapping, link form).
- No new dependencies; the plain-text fallback reuses `golang.org/x/net/html`
  (already an import of the `convert` package).

## Decisions

**D1. `Convert` falls back to plain text, not raw HTML.**
Replace `return html` on `ConvertString` error with a `plainText(html string)`
helper that walks the `x/net/html` tree and concatenates text-node content
(space-separated at block boundaries), so `<p>Hello</p>` becomes `Hello` and no
tag text leaks through. Alternative considered: return `""` — rejected because
it silently drops the whole article; the reader should still show something
readable.

**D2. `renderMarkdown` falls back to plain text, not raw markdown.**
Replace `return md` on `Render` error with a small `markdownToPlainText(md)`
helper that strips commonmark block and inline markers — leading heading
`#`s, list bullets/numbers, `>` blockquote bars, `**`/`*`/`_` emphasis, backtick
code spans, and `[text](url)`/`[alt]` link/image syntax (keeping the visible
`text` and the URL as readable text). This is a deterministic, testable
transformation used only on the error path. Alternative considered: reusing
`escapeMarkdownText` — rejected because backslash-escaping still leaves `\*\*`
asterisks visibly on screen, which the spec forbids.

**D3. `renderImage` escapes alt text.**
Before writing `[alt]`, backslash-escape the markdown-significant characters
(`\`, `` ` ``, `*`, `_`, `[`, `]`, `<`, `>`) in the alt text. `convert` already
owns this output, so the escape lives in the `convert` package (a small
`escapeAlt` helper), independent of the `ui` package's `escapeMarkdownText`.
Alternative considered: stripping alt text to alphanumerics — rejected, the alt
text should be preserved verbatim.

## Risks / Trade-offs

- **`markdownToPlainText` is approximate** → It only runs on an error path and
  only needs to be "readable", not round-trip perfect; its output is pinned by a
  focused unit test so it cannot silently regress.
- **`plainText` loses inline emphasis/bold in the fallback** → Acceptable: the
  fallback prioritizes readable text over preserved styling, and the spec only
  requires markup removal.
- **Alt escaping changes `[image]` output only for special chars** → Existing
  alt-text tests (`photo of a cat`) are unaffected; add a special-char case.

## Migration Plan

None — behavioral hardening only, no data/config/API impact. Rollback is a
revert.
