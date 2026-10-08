# Tasks

This is a spec-hygiene change: the delta specs in `specs/` are the deliverable.
No application code changes. Tasks verify completeness and land the deltas into
`openspec/specs/`.

## 1. Verify consolidation completeness

- [x] 1.1 Confirm `image-rendering` delta re-homes every requirement from `native-image-rendering` (8), `inline-image-rendering` (discovery/rendering/dedup), `lead-image-dedup` (2), and `image-decode-cap` (1), with no requirement dropped
- [x] 1.2 Confirm `image-loading` delta re-homes every requirement from `image-block-storage` (4), `image-load-scheduling` (3), and `article-recomposition` (7), plus the persistence/viewport-frontier parts of `inline-image-rendering`
- [x] 1.3 Confirm each re-homed requirement retains at least one scenario

## 2. Resolve the three contradictions

- [x] 2.1 Attribution: confirm no requirement still states a document-global credit fallback (the "element whose class indicates a credit" rule) for the lead or inline image
- [x] 2.2 Storage schema: confirm `feed-pipeline` no longer requires an `articles.image_data` BLOB column or an "additive image-column migration"
- [x] 2.3 ASCII border: confirm `configuration` states `:` for vertical borders and `|` for the scrollbar thumb

## 3. Fix the example and catalog gaps

- [x] 3.1 Confirm the attribution fallback example reads `photo: theintercept` (not `photo: intercept`)
- [x] 3.2 Confirm `configuration` registers the `link_expand_toggle` action (bound to `x` in the article view) and the full refresh key set `R`, `r`, `ctrl+r`, `f5`
- [x] 3.3 Confirm `copy-debug-info` names its default binding `D`

## 4. Apply the deltas to the main spec tree

- [x] 4.1 Sync the change into `openspec/specs/` (`openspec sync-specs` or archive), creating `image-rendering` and `image-loading`, removing the seven merged capabilities, and applying the `reader-ui`, `feed-pipeline`, `configuration`, and `copy-debug-info` deltas
- [x] 4.2 Delete the seven merged capability folders from `openspec/specs/` if the sync does not remove them
- [x] 4.3 Remove the six `reader-ui` requirements and trim its "Article content rendering" hyperlink paragraph

## 5. Validate

- [x] 5.1 Run `openspec validate --change spec-hygiene-consolidation` and fix any delta errors
- [x] 5.2 Grep `openspec/specs/` to confirm no stale cross-references remain to the removed capability names
- [x] 5.3 Run `git status` to confirm only intended spec files changed (no stray or temp files)
