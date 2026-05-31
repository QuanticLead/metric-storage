## MODIFIED Requirements

### Requirement: Metric Metadata Separation
The system SHALL store metric schema/descriptor details and resource/scope attributes in a dedicated metadata table (`otel_metrics_metadata`) to avoid duplication. The individual metric value data-points (Gauges, Sums, Histograms, Exponential Histograms, and Summaries) MUST be stored in their respective tables containing only the metric type-specific data fields plus a lookup reference (`MetricID` of type `UUID`) to the metadata table.

#### Scenario: Metadata table creation
- **WHEN** the metrics store is initialized
- **THEN** it executes DDL statements to create the `otel_metrics_metadata`, `otel_metrics_gauge`, `otel_metrics_sum`, `otel_metrics_histogram`, `otel_metrics_exponential_histogram`, and `otel_metrics_summary` tables if they do not exist.

#### Scenario: Ingest and split new metric streams
- **WHEN** an OTLP export request containing Gauge or Sum metrics is processed
- **THEN** the system computes a unique `MetricID` of type `UUID` (calculated using `xxh3` 128-bit hash of the resource/scope attributes and metric descriptors), inserts a metadata row with that ID into `otel_metrics_metadata`, and inserts the metric data-points into `otel_metrics_gauge` or `otel_metrics_sum` referencing that `MetricID`.
