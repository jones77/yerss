## Context

The article reader converts HTML body content to a display string, soft-wraps
it to the content width, and feeds it to a `bubbles/viewport`. The current
conversion (`internal/ui/html.go`) runs six regex pre-passes over the raw HTML
then calls `jaytaylor/html2text.FromString` with `OmitLinks: true`. The
soft-wrapper (`wrapText`) and display-width primitives (`truncate`, `padRight`)
live in the same file, though only `wrapText` is coupled to the converter's
output format; `truncate`/`padRight` are generic helpers also used by
`border.go` and `list.go`.

`html-to-markdown/v2` exposes a per-tag renderer seam:
`Register.RendererFor(tag, tagType, renderFn, priority)` where `renderFn` is
`func(ctx Context, w Writer, n *html.Node) RenderStatus` and writes directly to
the output buffer. The converter runs a pre-render DOM collapse phase, a render
walk that invokes registered renderers (falling back to commonmark defaults),
and a post-render trim/unescape phase. The base plugin handles `<script>` /
`<style>` / `<head>` removal and whitespace collapsing; the commonmark plugin
handles lists, blockquotes, code blocks, emphasis, and links by default.

## Goals / Non-Goals

**Goals:**
- Replace html2text with html-to-markdown so structured HTML renders as
  readable markdown (lists, blockquotes, code blocks, tables) and inline
  formatting renders as terminal escapes (ANSI bold, OSC 8 links).
- Dissolve the six regex pre-passes into three custom renderers.
- Untangle the conversion, wrapping, and display-width layers into separate
  files / packages with honest boundaries and clean dependency direction.

**Non-Goals:**
- Inline body-image (`<img>` in content) rendering as real images — that is
  `add-article-images` (enclosure lead images only); `<img>` stays `[alt]` /
  `[image]` placeholder here.
- Changing the soft-wrap algorithm, URL-break behavior, or ellipsis truncation
  — `wrapText` moves verbatim.
- Changing the viewport, border, or scroll machinery.
- Flattening the rest of `internal/` (`config`, `feed`, `store`, `ui`) to repo
  root — orthogonal mechanical refactor, separate change if desired.

## Decisions

### D1: Three custom renderers replace six regex hacks

Each hack maps to either a custom renderer or native commonmark handling:

| Hack removed                | Replacement                                      |
|-----------------------------|--------------------------------------------------|
| `replaceImages`             | `RendererFor("img", Inline, renderImage)`         |
| `flattenListParagraphs`     | native (library handles `<li><p>` correctly)      |
| `emboldenStrong`            | `RendererFor("strong"|"b", Inline, renderBold)`   |
| `stripSpans`                | native (per-node whitespace, no spurious spaces)  |
| `renderLists` (40+ lines)   | native commonmark lists + nesting                 |
| `wrapLinks`                 | `RendererFor("a", Inline, renderLink)`            |

The `<a>` renderer emits `OSC8-open [ rendered-children ] OSC8-close (url)`,
matching the exact format `wrapText`'s `linkTokens` already parses. The
`<strong>`/`<b>` renderer emits `\x1b[1m` + rendered children + `\x1b[22m`.
The `<img>` renderer extracts `alt` (or `[image]`) and writes it directly.

**Renderer recursion:** custom renderers call the converter's child-walk to
render inner content (so bold-inside-link composes correctly). The library
provides this via `ctx.RenderChildren(w, n)` or equivalent; the spike (task 1)
confirms the exact API surface.

### D2: EscapeModeDisabled

The converter's output is never re-parsed back to HTML — the terminal displays
the string directly. Smart escaping (`EscapeModeSmart`) would insert backslash
escapes before `*`, `[`, `]`, etc. in prose, producing visible `\*` noise.
`EscapeModeDisabled` passes these characters through raw, which is exactly what
a terminal reader wants. The post-render `UnEscapeContent` phase becomes a
no-op under this mode.

**Risk check:** a literal `[word](fake-url)` in article prose could resemble a
markdown link. `wrapText`'s link detection keys on the OSC 8 open sequence
(`\x1b]8;`), not bare `](`, so prose pseudo-links do not false-trigger the
link-aware wrapping path. Confirmed by reading `linkTokens` in the current
`html.go`.

### D3: New `convert` package at repository root

```
convert/
  convert.go   HTMLToMarkdown(html) string  + 3 custom renderers
  wrap.go      wrapText, linkTokens, truncateURL  (moved verbatim from html.go)

internal/ui/
  article.go   imports yerss/convert
  width.go     truncate, padRight  (generic display-width helpers, kept here)
  ...          border.go, list.go, popup.go, scroll.go, etc. (unchanged)
```

`convert` is the one piece with a coherent standalone identity (HTML →
terminal-markdown). Placing it at repo root (not `internal/convert`) makes the
seam maximally visible and signals it as the natural reusable boundary.
`internal/ui` importing a root-level `convert` is a clean, allowed dependency
direction (internal packages may import non-internal).

**Alternative considered:** `internal/convert` for consistency with the rest.
Rejected because `internal/` is ceremony for a single-binary CLI that nobody
imports as a library, and `convert` is the one package where the boundary is
meaningful. Fully flattening `config/feed/store/ui` out of `internal/` is
deferred — it is mechanical import-path churn orthogonal to this change.

**Alternative considered:** keep everything in `internal/ui` as split files.
Rejected because a 29-file god-package is the root cause of the "layers feel
hacky" problem; an honest package boundary is the fix.

### D4: `wrapText` moves to `convert`, with leading-indent preservation

`wrapText`, `linkTokens`, `truncateURL`, and `isTrailingPunct` move from
`html.go` to `convert/wrap.go`, exported as `WrapText`. They are already
OSC-8-aware and markdown-link-aware; they consume the converter's output
format and are stable across the swap. Co-locating them with the converter
keeps the output-format contract in one package rather than splitting an
implicit cross-package contract between `convert` and `internal/ui`.

`WrapText` additionally preserves each paragraph's leading indentation: it
extracts the leading spaces, reduces the effective width by the indent width,
and re-prefixes every wrapped line. Without this, `strings.Fields` (via
`linkTokens`) collapses the leading spaces of nested markdown list items, so
nested bullets render at column 0, indistinguishable from top-level bullets.
Confirmed and fixed against the real Drop Site article structure
(`<li><p>…</p><ul><li>…</li></ul></li>`).

### D5: `truncate` / `padRight` stay in `internal/ui`

These are generic display-width primitives (`ansi.Truncate` / `ansi.StringWidth`
wrappers) used by `border.go` and `list.go` for column padding and title
truncation. They have no relationship to HTML conversion or markdown. They stay
in `internal/ui` (in a small `width.go` or inlined where used).

### D6: Bullet marker stays `*` to match existing output

The commonmark plugin is configured to emit `* ` list markers, matching the
current html2text output and existing test assertions. Switching to `-` or a
Unicode bullet `•` (with ASCII fallback per the project's glyph convention) is
a trivial follow-up — it would require either a commonmark config change or a
4th custom `<li>` renderer plus `glyphsFor` plumbing. Deferred to keep this
change focused on the converter swap and layer untangle.

### D7: Post-render whitespace-only-line normalization

The commonmark list renderer emits continuation indentation on blank lines
inside nested list items, producing whitespace-only lines (e.g. `"  "`) in the
output. A `PostRenderer` collapses whitespace-only lines to empty, so the
terminal shows clean blank lines. Content-bearing trailing whitespace is
preserved: the `<br>` hard-break renderer emits `"  \n"` (two trailing spaces),
which must survive. The normalization only touches lines that are entirely
whitespace, so `<br>` output is unaffected. Confirmed by the spike.

## Risks / Trade-offs

- **[OSC 8 / ANSI escapes mangled by the converter pipeline]** → The escaping
  and collapse phases operate on DOM text nodes (pre-render) and the post-
  render trim/unescape only touches newlines and escaped chars. With
  `EscapeModeDisabled`, raw escapes written by custom renderers to `w` bypass
  both. **This is the core assumption and is de-risked by the spike (task 1)**
  before any other work: a ~20-line test feeds `<a><strong>x</strong></a>`
  through the converter and asserts the OSC-8-wrapped, ANSI-bold `[x](url)`
  output is intact with no stray backslashes. If the spike fails, the design
  pivots to a post-render injection strategy (render markdown first, then
  inject escapes into the finished string) — more complex but viable.
- **[New dependency churn]** → `html-to-markdown/v2` is MIT, 3.8k stars, golden-
  file tested, actively maintained. Pinned in go.mod. Pulls in
  `JohannesKaufmann/dom` and uses `golang.org/x/net/html` (already indirect via
  gofeed). Acceptable for a conversion library that is the project's core
  render path.
- **[Output shape changes break existing test assertions]** → Tests in
  `html_test.go`, `osc8_test.go`, `polish_test.go` assert exact output strings.
  List nesting, bullet spacing, and bold assertions update to match the
  converter's native output. This is expected and is not a regression — the
  new output is more correct. The spike establishes the canonical output shapes
  early so test updates are mechanical.
- **[Markdown chrome noise in terminal]** → Blockquote `> ` prefixes and code
  fences ```` ``` ```` render as literal text in the terminal. This is an
  upgrade (structure is visible) but may feel noisy. Suppressing specific
  elements (e.g. a custom `<pre>` renderer that drops fences) is a deferred
  follow-up; the spike includes a visual check of a real article to gauge
  acceptability.

## Open Questions

- ~~Does `html-to-markdown/v2` expose a `ctx.RenderChildren(w, n)` (or similarly
  named) method for custom renderers to recurse into child nodes?~~ **Resolved
  by spike:** the API is `ctx.RenderChildNodes(ctx, w, n)`. Custom renderers
  write prefix/postfix escapes to `w` and call `ctx.RenderChildNodes` for the
  inner content, so `<a><strong>x</strong></a>` composes bold-inside-link
  correctly.
