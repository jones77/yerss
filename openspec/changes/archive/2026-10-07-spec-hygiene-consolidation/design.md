## Context

`openspec/specs/` holds 13 capability specs. The image pipeline is fragmented
across nine of them (`native-image-rendering`, `inline-image-rendering`,
`lead-image-dedup`, `image-decode-cap`, `image-block-storage`,
`image-load-scheduling`, `article-recomposition`, plus the image and link
sections of the 1,991-line `reader-ui`). That fragmentation produced three real
contradictions (attribution fallback, storage schema, ASCII border glyph),
repetition (link display mode in three places, header URL in two), and several
clarity defects (wrong `photo: intercept` example, `x`-key catalog gap, refresh
key drift, undocumented Copy Debug Info binding).

The code is the source of truth and already implements the correct behavior;
every finding is spec drift, so this change touches specs only. See
`proposal.md` for the full motivation.

## Goals / Non-Goals

**Goals:**

- Collapse the nine image/link specs into two canonical capabilities and slim
  `reader-ui` into a UI shell that references them.
- Reconcile the three contradictions toward the code-accurate requirement.
- Fix the wrong example and the keybinding/help catalog gaps.

**Non-Goals:**

- No application code changes. No behavior changes.
- No changes to `article-json-export`, `compact-link-mode`, or the untouched
  requirements of `reader-ui`, `feed-pipeline`, `configuration`,
  `copy-debug-info`.
- No new requirements — only relocation and reconciliation of existing ones.

## Decisions

### 1. Two new capabilities split by "how it looks" vs "how it is managed"

The nine image specs collapse into:

- **`image-rendering`** — what the user sees: discovery, canonical source
  identity, dedup, attribution, halfblock and native rendering, placement
  lifecycle, decode width cap.
- **`image-loading`** — how images get there and stay put: unified storage,
  write-once persistence, viewport-driven load frontier, bounded concurrency,
  lazy native render, async lifecycle, derivation cache, recomposition, scroll
  snapping.

*Rationale:* a single mega-spec would recreate the `reader-ui` monolith
problem; keeping the current nine specs preserves the fragmentation. Two specs
along the presentation/lifecycle seam match how the code is organized
(`internal/image`, `internal/convert`, `internal/ui/compose` vs `internal/store`,
`internal/ui/model_image.go`).

### 2. `compact-link-mode` stays canonical; `reader-ui` sheds duplicates

In-body link rendering and the expanded header URL already have a canonical home
in `compact-link-mode` ("Compact link display by default", "Expanded header
shows the article URL above the title"). `reader-ui`'s "Clickable links in
article content" and "Clickable article URL in reader header" are removed as
duplicates, and "Article content rendering" is trimmed of its hyperlink
paragraph (now pointed at `compact-link-mode`). No `compact-link-mode` delta is
needed.

### 3. Contradictions resolve toward the code-accurate spec

| Contradiction | Code truth (source) | Stale spec | Fix |
|---|---|---|---|
| Document-global credit fallback | None (`convert/credit.go` `TestImageCreditIgnoresCreditClassOutsideFigure`) | `reader-ui` "Article lead image rendering" | Unified attribution chain in `image-rendering`, no credit-class fallback |
| Image storage schema | Unified `article_images` table (`image-block-storage`, `AGENTS.md`) | `feed-pipeline` "Additive schema migration" | Remove stale `image_data` migration + persistence sentence |
| ASCII vertical border glyph | `:` vertical, `|` thumb (`render/border.go` `GlyphsFor`) | `configuration` "ASCII border fallback" | Fix glyph list |

### 4. Keybinding/help catalog gaps filled from `config/keys.go`

- `link_expand_toggle` → `x` (article view) added to `configuration`.
- Refresh default reconciled to `R`, `r`, `ctrl+r`, `f5`.
- Copy Debug Info default binding `D` added to `copy-debug-info`.

### 5. Requirement relocation is verbatim where possible

Each requirement and its scenarios moved into the merged specs nearly verbatim;
the only rewritten requirement is the unified "Image attribution" (merging the
lead and inline chains and dropping the credit-class fallback) and the fixed
`photo: theintercept` example.

## Old → new mapping

| Source capability (removed) | Destination |
|---|---|
| `native-image-rendering` (8 reqs) | `image-rendering` |
| `inline-image-rendering` (4 render/dedup reqs) | `image-rendering` |
| `inline-image-rendering` (persistence, viewport) | `image-loading` |
| `lead-image-dedup` (2 reqs) | `image-rendering` |
| `image-decode-cap` (1 req) | `image-rendering` |
| `image-block-storage` (4 reqs) | `image-loading` |
| `image-load-scheduling` (3 reqs) | `image-loading` |
| `article-recomposition` (7 reqs) | `image-loading` |

| Source capability (modified) | Change |
|---|---|
| `reader-ui` | Remove 6 requirements, trim 1 |
| `feed-pipeline` | Remove 1, modify 1 |
| `configuration` | Modify 2 |
| `copy-debug-info` | Modify 1 |

Final tree: 8 specs (`reader-ui`, `feed-pipeline`, `configuration`,
`compact-link-mode`, `image-rendering`, `image-loading`,
`article-json-export`, `copy-debug-info`).

## Risks / Trade-offs

- **[Orphaned behavior during merge]** → Mitigation: a full requirement
  inventory (recorded during exploration) is the cross-check; every removed
  requirement has an explicit Migration target, and every new requirement has at
  least one scenario.
- **[Large delta specs are hard to review]** → Mitigation: the two new
  capability deltas are organized by concern and the removal deltas are
  name-plus-reason only, keeping the diff reviewable.
- **[Sync/archive mismatch]** → Mitigation: `openspec validate` run before
  archive, and the archive stages the archived change dir, synced specs, and any
  touched files explicitly (per `config.yaml`).
