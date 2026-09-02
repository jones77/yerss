## Context

The exit frame's kitty cleanup emits `DeleteByID` → `\x1b_Ga=d,d=i,i=<id>,q=1\x1b\\`
(lowercase `d=i`). Per the kitty graphics protocol, lowercase `d=i` deletes the
placement but **retains** the terminal's cached image data; uppercase `d=I`
deletes the placement and frees the cached data. The cleanup spec's intent is
to free the cached data, and Ghostty's kitty-graphics delete handling is
partial — so on Ghostty a placement can survive the by-id delete and float over
the list until the app quits.

See proposal.md — Why for the observed behavior. Current relevant state:

- `image.DeleteByID(id)` (native.go) emits the lowercase `d=i` sequence.
- `NativeRenderer.Clear()` emits `a=d,d=a` (delete-all visible placements).
- `computeNativeClear` (ui/image.go) emits per-id deletes on the article-exit
  frame (view != viewArticle) and the delete-all only for no-native articles.
- Protocol detection (protocol.go) is env-based; `TERM_PROGRAM=ghostty` and
  `GHOSTTY_RESOURCES_DIR` both map to kitty protocol. No terminal identity is
  carried beyond the protocol enum.

## Goals / Non-Goals

**Goals:**

- Make the by-id delete match the spec's intent: free the cached image data
  (`d=I`), on every terminal that honors it.
- On Ghostty — where by-id delete is unreliable — guarantee no placement
  survives leaving the article view via an additional delete-all.
- Keep the scroll-out and resize per-id deletes as-is apart from the case
  change.

**Non-Goals:**

- Changing how native images render or are transmitted.
- Adding a runtime protocol probe/query round-trip to the terminal.
- Touching OSC 1337 (iTerm2/WezTerm) behavior, which correctly emits no deletes.

## Decisions

### D1: Flip `DeleteByID` to the data-freeing form

`image.DeleteByID` becomes `\x1b_Ga=d,d=I,i=<id>,q=1\x1b\\` (uppercase `d=I`).
The kitty spec's uppercase variant deletes the placement and, when the image is
not referenced elsewhere (e.g. scrollback), frees its cached data — exactly the
documented cleanup intent. This applies uniformly to every per-id delete:
scroll-out, resize supersede, and article-exit frames.

**Alternatives considered:**
- *Keep lowercase `d=i`.* Fails the spec's "freeing its cached image data"
  requirement by construction; the retained cache is the accumulation the
  cleanup requirement exists to prevent.
- *Emit both lowercase and uppercase.* Redundant; uppercase subsumes lowercase
  for placement deletion.
- *Use delete-all everywhere instead.* `d=a` only clears visible placements and
  never frees cached data — the opposite of what the cleanup needs.

### D2: Carry terminal identity; add a Ghostty delete-all fallback on exit

Extend the kitty-protocol path with a cheap, env-derived identity flag
(`terminalIsGhostty`), set alongside `DetectProtocol` from the same environment
probe (`TERM_PROGRAM == "ghostty"` or `GHOSTTY_RESOURCES_DIR` set). On the
article-exit frame (`view != viewArticle`), after emitting the per-id deletes,
`computeNativeClear` SHALL append a `d=a` delete-all clear when the terminal is
Ghostty. The delete-all is the belt-and-suspenders guarantee: even if Ghostty
ignores `d=I`, `d=a` removes every visible placement, so no photo survives over
the list. Because the sequence is prepended to the first line of the list frame
(the same position the per-id deletes already occupy), the terminal processes
it before repainting the list.

**Alternatives considered:**
- *Send `d=a` on every terminal.* Changes behavior for kitty proper, which
  handles `d=I` correctly; the redundant clear would also force a repaint
  interaction with kitty's scrollback that the per-id delete already avoids.
  Restricting it to Ghostty keeps the kitty path unchanged.
- *Skip native rendering on Ghostty entirely.* Too broad; renders degrade to
  halfblocks on a supported protocol. The fallback keeps full rendering.
- *Query the terminal for delete support.* Requires a response round-trip the
  app does not currently do; env-based detection mirrors the existing protocol
  detection philosophy.

### D3: Keep the delete-all gate for no-native articles unchanged

The existing `nativeClearArticleID` gate (delete-all only on the transition
into a no-native article) is orthogonal and stays. The Ghostty exit fallback is
keyed on the article-exit transition, not the no-native transition, so both can
fire on distinct frames without interference.

## Risks / Trade-offs

- [`d=I` frees cached data that a re-scroll wanted to re-display cheaply] → The
  article already re-transmits a scrolled-back image via `nativeSent` tracking;
  freeing the terminal cache only costs a re-transmit, not correctness.
- [Ghostty's env identity is wrong in an edge case (e.g. nested terminal,
  SSH)] → The fallback is additive (an extra `d=a` on exit) and harmless even
  when unnecessary; a false negative simply falls back to today's behavior.
- [Tests assert the exact lowercase sequence] → `TestNativeKittyExitDeletesEveryRecordedID`
  and friends assert `DeleteByID(id)` output; they pass unchanged because they
  call the helper, and a new test asserts the uppercase form and the Ghostty
  fallback appear in the exit frame.

## Migration Plan

- Land D1 and D2 together (both are sequence changes in the same exit path).
  Update the `native.go` doc comment that currently claims lowercase `d=i`
  frees cached data. No data or schema migration; run the full `internal/ui`
  and `internal/image` test suites.

## Open Questions

None.