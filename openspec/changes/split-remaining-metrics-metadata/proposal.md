## Why

The database schema uses redundant metadata columns across histogram, exponential histogram, and summary tables, which wastes storage space and leads to inconsistent schema designs compared to the optimized gauge and sum tables. By aligning all metric schemas, we ensure consistency across the database.

## What Changes

- Modify the DDL schemas for `otel_metrics_histogram`, `otel_metrics_exponential_histogram`, and `otel_metrics_summary` in `clickhouse_schema.go` to remove duplicate metadata and index columns.
- Add a reference column `MetricID` (`UUID`) to these three tables.
- Update their `ORDER BY` clauses to use `(MetricID, toUnixTimestamp64Nano(TimeUnix))` to match the pattern established by the gauge and sum tables.
- No changes to Go server mapper/ingestion logic since it does not support ingesting histograms, exponential histograms, or summaries.

## Capabilities

### New Capabilities

*(None)*

### Modified Capabilities

- `metric-storage`: Extend the metadata separation requirement to include `otel_metrics_histogram`, `otel_metrics_exponential_histogram`, and `otel_metrics_summary` tables.

## Impact

- **Database**: Schema definitions for ClickHouse tables `otel_metrics_histogram`, `otel_metrics_exponential_histogram`, and `otel_metrics_summary`.
- **Server**: Only DDL execution is affected; ingestion code is untouched as the server does not support histograms or summaries.
