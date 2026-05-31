## Context

The database schemas for `otel_metrics_histogram`, `otel_metrics_exponential_histogram`, and `otel_metrics_summary` in `clickhouse_schema.go` still contain inline metadata columns (such as `ResourceAttributes`, `ServiceName`, `MetricName`, `Attributes`, etc.) and their corresponding bloom filter indexes. We want to align these remaining schemas with the optimized design implemented for `otel_metrics_gauge` and `otel_metrics_sum` by using a reference column `MetricID` (`UUID`).

## Goals / Non-Goals

**Goals:**
- Update `clickhouse_schema.go` to remove redundant metadata columns from `otel_metrics_histogram`, `otel_metrics_exponential_histogram`, and `otel_metrics_summary`.
- Add `MetricID UUID CODEC(ZSTD(1))` to these three tables.
- Update their `ORDER BY` keys to `(MetricID, toUnixTimestamp64Nano(TimeUnix))`.
- Remove redundant bloom filter indexes on attributes from these tables, as metadata query indexes are defined on `otel_metrics_metadata`.

**Non-Goals:**
- Modifying the Go server mapper or client code to support ingesting histograms, exponential histograms, or summaries. As requested, the server will continue to only ingest Gauge and Sum metrics.

## Decisions

### 1. Schema Alignment and Engine Configuration
- **Decision**: Keep the table engines as `MergeTree()`, partitioned by `toDate(TimeUnix)`, and ordered by `(MetricID, toUnixTimestamp64Nano(TimeUnix))`.
- **Rationale**: This mirrors the exact layout of the gauge and sum tables, ensuring consistent physical sorting and partitioning across all data-point tables.

### 2. Bloom Filter Index Removal
- **Decision**: Remove the bloom filter indexes (`idx_res_attr_key`, `idx_res_attr_value`, etc.) from the value tables.
- **Rationale**: The columns they index are removed from these tables. Any query filtering on resource/scope/metric attributes will perform a join on `MetricID` with `otel_metrics_metadata`, where these attributes are already fully indexed.

## Risks / Trade-offs

- **[Risk]** Inability to test histogram/summary ingestion path.
  - *Mitigation*: Since the server does not support ingesting them, we only need to verify that table creation successfully executes DDLs. The integration tests can verify that the server connects, creates the schema correctly, and can still insert gauge and sum metrics.
