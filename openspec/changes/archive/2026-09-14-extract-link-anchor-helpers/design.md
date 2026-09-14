## Context

`internal/ui/image.go` mixes three concerns: native-image transmit/clear sequencing, halfblock previews and block composition, and image-load scheduling — plus a fourth, unrelated block: five pure link-anchoring helpers used by the link-expand toggle to keep the scroll position stable across a reflow. Those helpers (`anchorableLine`, `stripLeadingURLTokens`, `isURLToken`, `findAnchorLine`, `commonPrefix`, and the `minURLFragmentLen` constant) take only `[]string` lines and `[]articleLink` links and have no receiver on `Model`.

The `compose` package already owns the pure link/reflow logic (links.go's `StripLinkHrefs`, sentinel.go's link scanning, reflow.go's wrapping), so it is the natural home.

## Goals / Non-Goals

**Goals:** relocate the five helpers and their constant into `compose`; keep behavior byte-identical; keep the move mechanical and reviewable.

**Non-Goals:** no package split of `internal/ui` (that is change `split-ui-package`, later); no refactor of the image-load or native-render paths; no signature change beyond the link parameter type.

## Decisions

### D1 — Destination `compose/anchor.go`

A single new file in `compose` holds the moved helpers. Rationale: matches how `compose` already groups link/reflow pure functions by concern (links.go, reflow.go); keeps the diff one-file-in, one-file-out. The two helpers the `ui` package calls across the boundary (`AnchorableLine`, `FindAnchorLine`) are exported; the internal ones (`stripLeadingURLTokens`, `isURLToken`, `commonPrefix`, `minURLFragmentLen`) stay unexported.

### D2 — Parameter `urls []string` instead of `[]articleLink`

The three link-aware helpers (`anchorableLine`, `stripLeadingURLTokens`, `isURLToken`) read only `link.url`, never `link.text`. Changing the parameter to `urls []string` removes `compose`'s need to know about `ui.articleLink` (which `compose` cannot import without a cycle). Call sites pass the URLs via a small `urlsOf` helper in the `ui` package. Rationale: avoids moving the `articleLink` type (used widely across `ui`) for the sake of five helpers that only need URLs.

### D3 — Tests follow the move

`link_expand_test.go` exercises the helpers from package `ui`. Their cases move to a `compose/anchor_test.go` (with `[]string` URLs), preserving coverage; the link-expand integration cases that stay in `ui` keep testing the observable toggle behavior end-to-end.

## Risks / Trade-offs

- [Signature change ripples to two call sites] → Both are in `image.go`'s toggle path; a `urlsOf` helper keeps the diff local.
- [Moving tests risks losing the end-to-end toggle coverage] → Keep the integration cases in `ui`; move only the unit cases.
