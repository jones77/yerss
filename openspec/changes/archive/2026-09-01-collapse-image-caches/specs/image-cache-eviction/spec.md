## REMOVED Requirements

### Requirement: Bounded in-memory image caches

**Reason**: The reader's session image caches are intentionally unbounded. The database is the retention layer, and the byte-budgeted LRU machinery this requirement described was never implemented and is not wanted.

**Migration**: None. No configuration or data migration is required.

### Requirement: Decoded source resolution cap

**Reason**: Removed together with the bounded-cache requirement. Full-resolution decode is accepted; oversized-source risk will be assessed with the image-stats diagnostic before any cap is reintroduced.

**Migration**: None.
