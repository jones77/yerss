## MODIFIED Requirements

### Requirement: Stripped link re-wrap preserves single word spacing

The system SHALL re-wrap a compact-mode paragraph that contained a hidden URL so
that word separation between a link name (or its trailing punctuation) and the
following text is exactly one space — never doubled and never missing. The
leading separator space left where a URL was removed SHALL NOT combine with the
re-wrap join space to produce a double space, and the removal SHALL NOT consume
the separator space entirely so the link name fuses with the following word.
Closing punctuation stranded after a hidden URL (a comma or period that followed
the link immediately in the source) SHALL butt against the preceding link name
with no space between them. A rendered line SHALL NOT contain a doubled space, a
missing space, or a space before such stranded punctuation where a link name (or
its punctuation) meets the following text.

#### Scenario: No double space after a stripped link name

- **WHEN** a compact-mode paragraph renders a link name immediately followed by more text whose URL was hidden
- **THEN** the rendered text shows exactly one space between the link name and the following text, with no doubled space

#### Scenario: No double space after stripped link punctuation

- **WHEN** a compact-mode link whose name ends in punctuation is hidden and the paragraph re-wraps
- **THEN** no extra space appears between the punctuation and the following text

#### Scenario: No missing space after a stripped link name

- **WHEN** a compact-mode paragraph renders a link name immediately followed by more text whose URL was hidden
- **THEN** the rendered text shows exactly one space between the link name and the following text, so the link name and the following word are not fused together (e.g. `claimed searching`, not `claimedsearching`)

#### Scenario: No space before punctuation stranded by a hidden URL

- **WHEN** a compact-mode link is followed immediately by punctuation such as a comma, and the URL between the name and the punctuation was hidden
- **THEN** the punctuation butts against the link name with no space between them (e.g. `$195,869.52, for`, not `$195,869.52 , for`)
