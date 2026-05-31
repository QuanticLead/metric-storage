## Why

Currently, OTLP metric data-points (Gauges and Sums) are stored in ClickHouse with all their metadata (ResourceAttributes, ScopeAttributes, ScopeName, ScopeVersion, MetricName, MetricDescription, MetricUnit, etc.) duplicated in every row. This leads to massive storage overhead and redundant database writes. We need to extract this metadata into a separate lookup table and store only the timestamp, value, and a reference key in the data-points tables to optimize storage efficiency and query speed.

## What Changes

- Add a metadata lookup table (`otel_metrics_metadata`) to store unique combinations of resource/scope/metric attributes and identifiers.
- Modify the schema of `otel_metrics_gauge` and `otel_metrics_sum` to remove redundant metadata columns and add a metadata reference column (`MetricID` of type `UUID`).
- Update `clickhouse_client.go` to insert new metadata entries and link them from the gauge and sum data-points.
- Implement an in-memory cache on the Go server to avoid checking or writing metadata to ClickHouse for already known metric streams during high-throughput ingestion.
- Maintain backward compatibility or migrate existing test suits to use the new schema.

## Capabilities

### New Capabilities
- `metric-storage`: Defines the requirements for storing and structuring OTLP metrics (Gauges and Sums) efficiently in ClickHouse by separating datapoints and metadata.

### Modified Capabilities

## Impact

- **Database Schema**: Modifies the structure of `otel_metrics_gauge` and `otel_metrics_sum`, and introduces `otel_metrics_metadata`.
- **Go Server Code**: Changes table creation, mapper, and insertion logic in `clickhouse_client.go`, `clickhouse_schema.go`, `metrics_mapper.go`, and `metrics_service.go`.
- **Tests**: Updates unit and integration tests to align with the new schema and metadata-linking behavior.
