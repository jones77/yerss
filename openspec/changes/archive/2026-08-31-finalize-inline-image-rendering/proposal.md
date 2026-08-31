## Why

The `inline-image-rendering` change's implementation is complete and all
automated tests pass, but its remaining tasks are user-run smoke tests that
were not re-run after the latest fix rounds, and the change (plus the entangled
`standard-caption-width` implementation) is uncommitted and unarchived. This
change captures exactly what remains so a fresh session can finish it: verify
the fixes on the real articles, then sync, archive, and commit both changes and
the shared working tree.

## What Changes

- **Run the pending smoke tests** on the Intercept article ("Ecocide in Iran")
  and the hurricane article
  (`https://theintercept.com/2026/08/29/hurricane-helene-ice-asheville-nc-g20-trump/`),
  then mark the corresponding `inline-image-rendering` tasks complete. The
  specific checks (from that change's `tasks.md`):
  - **9.1 / 6.3** (Intercept): images never overlap the border; space pages
    through text without skipping to images; j/k snap each inline image
    bottom-then-top and out; the TomDispatch logo caption reads its banner text
    with no leftover `](/tomdispatch)`; the links popup lists
    `https://theintercept.com/tomdispatch`.
  - **10.3** (Intercept): j-scrolling shows each photo fully (no blank strip,
    no manual scrolling through empty rows); k-scrolling leaves each photo
    until it scrolls off in one snap.
  - **11.4** (Intercept): the bottom snap shows each photo with its caption at
    the bottom edge (one-line captions included); between the Trump's War
    thumbnail and the book cover the two Fallujah paragraphs read normally and
    the book cover's top shows a blocky preview rather than an empty strip
    before it snaps in.
  - **12.3** (hurricane): scroll-k from the bottom shows each gallery photo once
    (filling the screen) then scrolls it off to the previous image or content,
    with no stuck offset; scroll-j through the whole article stays clean.
  - Also exercise the **halfblock** (`-a`/ASCII and non-native) paths and a
    **resize** during reading if the app supports it, since the snap and
    preview logic is geometry-dependent.
- **Archive `inline-image-rendering`** once the smoke tests pass: sync its two
  delta specs (`specs/inline-image-rendering/spec.md` and
  `specs/reader-ui/spec.md`) into the main specs at `openspec/specs/`, then
  move it to `openspec/changes/archive/2026-08-31-inline-image-rendering/`.
- **Commit the entangled working tree** in the archive commit: the
  implementation files for both `inline-image-rendering` and
  `standard-caption-width` are interleaved in the same uncommitted files
  (everything since `a4b828c`), so committing them at `inline-image-rendering`'s
  archive captures both changes' implementation. Also commit the uncommitted
  `AGENTS.md` additions (schema pointer, markdown-link format, image scroll snap
  invariants).
- **Tune the caption width**: `standard-caption-width` set the caption width to
  `max(1, contentW*7/10)` as a first guess; verify it looks right on a real
  terminal and adjust the fraction if the 70% looks off (it is a deliberate
  "tune me" default).

## Capabilities

### New Capabilities

(none — `skip_specs: true`; this change adds no behavior, it verifies,
archives, and commits existing work. The behavior spec deltas live in the
`inline-image-rendering` change and are synced when it archives.)

### Modified Capabilities

(none)

## Impact

- **Changes to finish**: `openspec/changes/inline-image-rendering/` (active,
  34/40 tasks; pending tasks 6.3, 7.5, 9.1, 10.3, 11.4, 12.3 — 7.5 is
  superseded by the later rounds, so it can be marked done once 9.1 passes).
- **Already archived**: `openspec/changes/archive/2026-08-31-standard-caption-width/`
  (dir + synced `reader-ui` main spec committed in `728a36a`; its
  implementation is still uncommitted and lives in the shared files below).
- **Uncommitted shared implementation files** (since commit `a4b828c`,
  "Prune removed feeds at startup"): `internal/ui/image.go`,
  `internal/ui/article.go`, `internal/ui/article_image_test.go`,
  `internal/ui/links_popup_test.go`, `internal/image/native_cmd.go`,
  `internal/image/image_test.go`, `internal/image/image.go`,
  `internal/image/native.go`, `convert/convert.go`, `convert/credit.go`,
  `convert/convert_test.go`, `convert/credit_test.go`, `AGENTS.md`.
- **Main specs to sync at archive**: `openspec/specs/reader-ui/spec.md`
  (already partially updated by the `standard-caption-width` sync; the
  `inline-image-rendering` deltas add the inline-image rendering, captions,
  and scroll-snap requirements), and a new
  `openspec/specs/inline-image-rendering/spec.md` from that change's delta.
- **Verification commands**: `gofmt -l` on touched files, `go vet ./...`,
  `go build ./...`, `go test ./...`, `openspec validate --all`.
- **Build/run**: `go build -o /tmp/yerss ./cmd/yerss && /tmp/yerss`. The test
  articles are already stored in the local SQLite DB (the default path under
  `~/Library/Application Support/yerss/`); do not fetch them from the network.