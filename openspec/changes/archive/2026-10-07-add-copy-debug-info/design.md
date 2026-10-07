## Context

The article view (`internal/ui/articleview/`) holds all the state needed for a debug snapshot: `s.width`, `s.height`, `s.Viewport.YOffset`, `len(s.lines)`, `s.article.Title`, `s.article.Link`, `s.article.ID`, `s.ascii`, and `s.LinkExpand`. An existing `CopyTextCmd(text, label)` in `open.go` already copies arbitrary text to the clipboard cross-platform and reports the result via `URLActionMsg`. The keybinding system (`internal/config/keys.go`) uses a single `catalog` slice so a new action is propagated to config template, help, and registry automatically.

## Goals / Non-Goals

**Goals:**
- One keypress produces a clipboard snapshot sufficient to reproduce rendering bugs.
- Reuse the existing clipboard command rather than introduce a new copy path.

**Non-Goals:**
- No automatic crash/bug reporting over the network.
- No capturing of user data beyond the identifiers and modes listed in the spec.

## Decisions

**Format the snapshot as a block of `key: value` lines.**
A fixed, stable key structure is trivially parseable by both humans and scripts, and reads naturally in a bug report. Include an `app: yerss` line so the source is self-identifying. Alternative: free-form prose — rejected because it is harder to scan and harder to keep consistent.

**Place the assembly function in the articleview package next to the copy commands.**
A new function (e.g. `debugInfoSnapshot(s)` returning a string, plus a `CopyDebugInfoCmd(s)` that wraps `CopyTextCmd`) keeps the formatting colocated with the state it reads and with the existing copy commands. Alternative: a separate package — overkill for a single-view feature.

**Add a new `Action` constant `CopyDebugInfo` with a default key.**
Pick a default that does not collide with existing article-view bindings (`c` = copy URL, `C` = copy article text). The exact key is chosen at implementation time after checking `EffectiveKeys` for conflicts; the catalog entry uses the `ViewArticle` view.

**Reuse `ansi.Strip` where the snapshot could contain ANSI.** The title is plain text, but applying the existing stripping approach keeps the output clipboard-clean and consistent with the article-text copy path.

## Risks / Trade-offs

- [Default key collision with an existing binding] → check `EffectiveKeys` output and the config template before finalizing the key.
- [Snapshot too verbose or too sparse] → keep it to the fields listed in the spec; extend later if bug reports prove a field is missing.
- [Image rendering mode not a single flag] → include the flags that are cheaply available (`ascii`, `LinkExpand`); if native-vs-halfblock is needed later it can be added without changing the spec contract.
