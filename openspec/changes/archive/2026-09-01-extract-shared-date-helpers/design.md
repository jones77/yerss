## Context

`dayKey`/`dayStart` live in `internal/ui/geometry.go`; `longDate`/`dayLabel` live in `internal/ui/list.go`; `json.go` reimplements the day key inline. See proposal.md.

## Goals / Non-Goals

**Goals:** one home; identical output; ordinal correctness.

**Non-Goals:** no timezone or locale changes.

## Decisions

### D1 — `internal/timeutil`

`DayStart(t)` (local midnight), `DayKey(t)` ("20060102" or "undated"), `LongDate(t)` (weekday + ordinal day + month + year), `DayLabel(date, now)` (today/yesterday prefix), plus the layout constants `"15:04"`, `"2006-01-02 15:04:05"`, `"20060102"`.

### D2 — Ordinal via modulo 100

Suffix is `th` for 11–13, otherwise by ones digit. Rationale: correct for 111, 112, 113 etc., not just 10–19.

## Risks / Trade-offs

- [json.go bucketing key format] → Uses `timeutil.DayKey`; output JSON keys are unchanged ("20060102").
