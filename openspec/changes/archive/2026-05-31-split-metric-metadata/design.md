## Context

Currently, the OTLP Metric Storage server inserts metric values and all their schema metadata (resource/scope attributes, metric descriptors, units, descriptions) into single wide tables (`otel_metrics_gauge` and `otel_metrics_sum`). This leads to duplicate strings and high disk consumption. We will refactor this to split metadata into a lookup table.

## Goals / Non-Goals

**Goals:**
- Create `otel_metrics_metadata` using a `ReplacingMergeTree` engine, ordered by `MetricID`.
- Update `otel_metrics_gauge` and `otel_metrics_sum` schemas to store only `MetricID`, `StartTimeUnix`, `TimeUnix`, `Value`, `Flags` (and temporality fields for Sums).
- Implement a deterministic `MetricID` generator in Go using xxh3 128-bit hashing on serialized, sorted metadata keys and values, represented as a `UUID`.
- Implement an in-memory cache using `sync.Map` to avoid redundant database insertions for already discovered metric schemas.
- Refactor the mapping and database insertion logic in `metrics_mapper.go` and `clickhouse_client.go` to handle the write path for both tables.

**Non-Goals:**
- Migrating historical data (since this is an exercise/prototype baseline).
- Refactoring the other metric types (Histogram, ExponentialHistogram, Summary) as they are out of scope for Gauge/Sum.

## Decisions

### 1. Unique Metric Stream Identification (MetricID)
- **Option A**: Use FNV-1a 64-bit hash (producing `uint64`).
- **Option B (Chosen)**: Compute a deterministic 128-bit `xxh3` hash (from `github.com/zeebo/xxh3`) based on the sorted representation of the metadata, and convert it to a standard 16-byte `uuid.UUID`.
- **Rationale**: A 128-bit hash guarantees 0% collision probability at any scale. The standard Go `uuid.UUID` is fully comparable and can be used directly as a map/cache key, and is natively supported as a `UUID` type in ClickHouse.

### 2. In-Memory Cache for Deduplication
- **Option A**: Query ClickHouse on every write with `INSERT ... ON DUPLICATE KEY UPDATE` or similar.
- **Option B (Chosen)**: Use Go's `sync.Map` as a write-through cache of known `uuid.UUID` values.
- **Rationale**: OTLP ingestion runs at high throughput. Using `sync.Map` provides lock-free reads for already established metric streams, allowing us to perform metadata inserts only when a new metric stream is first observed.

### 3. ClickHouse Table Engines and Schemas
- **Metadata table**: `ReplacingMergeTree` ordered by `MetricID` (`UUID`).
- **Datapoints tables**: `MergeTree` partitioned by `toDate(TimeUnix)` and ordered by `(MetricID, toUnixTimestamp64Nano(TimeUnix))`.
- **MetricID Schema**:
  ```sql
  MetricID UUID CODEC(ZSTD(1))
  ```

## Risks / Trade-offs

- **[Risk]** Cache Memory Leak. If the number of unique metric streams grows indefinitely, `sync.Map` could consume too much memory.
  - *Mitigation*: The README states that cardinality is "low" (under 100k streams, memory usage of the cache is < 5MB).
- **[Risk]** ClickHouse JOIN performance.
  - *Mitigation*: Since metadata has low cardinality, a join on `MetricID` is extremely fast and fits in memory for hash joins.
