## MODIFIED Requirements

### Requirement: In-Memory Metadata Cache
The Go server SHALL maintain a bounded, thread-safe, in-memory cache of successfully inserted or resolved `MetricID` values to prevent redundant database writes and avoid unbounded memory growth.

#### Scenario: Ingesting cached metric streams
- **WHEN** an OTLP metrics stream is received and its computed `MetricID` (UUID) is found in the in-memory cache
- **THEN** the system bypasses the ClickHouse metadata insertion check entirely and directly performs the batch insert of data-points referencing the cached `MetricID`.

#### Scenario: Ingesting new metric stream when cache is at capacity
- **WHEN** a new `MetricID` is processed and the in-memory cache has reached its configured maximum capacity
- **THEN** the cache evicts the least recently used (LRU) item to accommodate the new `MetricID`.
