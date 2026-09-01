## Context

Both functions resolve a caption from `convert.ImageCaption` (or the lead's HTML), then alt text, link text, and `photo: <source>` fallback. See proposal.md.

## Goals / Non-Goals

**Goals:** one derivation path.

**Non-Goals:** no changes to which caption wins.

## Decisions

### D1 — One `resolveImageAttribution(a, url, alt, linkText, isLead)`

The lead path skips the alt/link-text steps (lead images only have credit/source); inline images keep the full chain. Rationale: same chain, explicit difference.

## Risks / Trade-offs

- [Subtle precedence bugs] → Cover with the existing caption tests (photo-grid first image, promo banner, source fallback).
