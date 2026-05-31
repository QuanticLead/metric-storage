# metric-storage Specification

## Purpose
TBD - created by archiving change split-metric-metadata. Update Purpose after archive.
## Requirements
### Requirement: Metric Metadata Separation
The system SHALL store metric schema/descriptor details and resource/scope attributes in a dedicated metadata table (`otel_metrics_metadata`) to avoid duplication. The individual metric value data-points (Gauges and Sums) MUST be stored in their respective tables (`otel_metrics_gauge`, `otel_metrics_sum`) containing only the timestamp, value, flags, (and temporality/monotonicity for Sums), plus a lookup reference (`MetricID` of type `UUID`) to the metadata table.

#### Scenario: Metadata table creation
- **WHEN** the metrics store is initialized
- **THEN** it executes DDL statements to create the `otel_metrics_metadata`, `otel_metrics_gauge`, and `otel_metrics_sum` tables if they do not exist.

#### Scenario: Ingest and split new metric streams
- **WHEN** an OTLP export request containing Gauge or Sum metrics is processed
- **THEN** the system computes a unique `MetricID` of type `UUID` (calculated using `xxh3` 128-bit hash of the resource/scope attributes and metric descriptors), inserts a metadata row with that ID into `otel_metrics_metadata`, and inserts the metric data-points into `otel_metrics_gauge` or `otel_metrics_sum` referencing that `MetricID`.

### Requirement: In-Memory Metadata Cache
The Go server SHALL maintain an in-memory cache of successfully inserted or resolved `MetricID` values to prevent redundant database writes.

#### Scenario: Ingesting cached metric streams
- **WHEN** an OTLP metrics stream is received and its computed `MetricID` (UUID) is found in the in-memory cache
- **THEN** the system bypasses the ClickHouse metadata insertion check entirely and directly performs the batch insert of data-points referencing the cached `MetricID`.

