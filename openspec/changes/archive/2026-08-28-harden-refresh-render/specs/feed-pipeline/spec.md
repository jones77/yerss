## ADDED Requirements

### Requirement: Bounded feed fetch deadlines

Every feed fetch the system performs — whether for the startup verification gate, the
startup refresh, or a manual refresh — SHALL be bounded by both a per-feed HTTP
timeout and an overall deadline. A feed whose fetch exceeds the per-feed timeout
SHALL be cancelled and recorded as an error in the fetch result; it SHALL NOT block
the completion of the remaining feeds or the refresh pass as a whole. The overall
deadline SHALL cancel any still-in-flight fetches once it elapses. The system SHALL
cap the number of concurrently in-flight feed requests with a fixed bound rather than
spawning one unbounded goroutine per configured feed. The startup verification gate
and the refresh path SHALL share one bounded fetch implementation so their timeout
behavior cannot diverge.

#### Scenario: Hung feed does not stall a manual refresh

- **WHEN** the user triggers a manual refresh and one configured feed's server accepts the TCP connection but never responds
- **THEN** that feed is cancelled at the per-feed timeout, recorded as an error in the fetch result, and the refresh pass completes with the other feeds within the overall deadline

#### Scenario: Overall deadline cancels in-flight fetches

- **WHEN** a fetch pass is in progress and the overall deadline elapses before all feeds have completed
- **THEN** the system cancels every still-in-flight feed fetch and completes the pass with the feeds that already finished

#### Scenario: Concurrency is bounded

- **WHEN** a fetch pass runs against more configured feeds than the concurrency bound
- **THEN** the system holds excess feeds until in-flight slots free up rather than issuing all requests at once

#### Scenario: Gate and refresh share timeout behavior

- **WHEN** the startup verification gate and a manual refresh both fetch the same set of feeds
- **THEN** both paths apply the same bounded-timeout and concurrency behavior because they use one shared fetch implementation
