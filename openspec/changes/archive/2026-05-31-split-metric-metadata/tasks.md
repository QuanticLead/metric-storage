## 1. Database Schema Updates

- [x] 1.1 Update `clickhouse_schema.go` to add the `otel_metrics_metadata` table DDL with `MetricID UUID`.
- [x] 1.2 Update the `otel_metrics_gauge` and `otel_metrics_sum` table schemas in `clickhouse_schema.go` to replace duplicate columns with a single `MetricID UUID` reference.

## 2. Go Logic & In-Memory Cache

- [x] 2.1 Add `github.com/zeebo/xxh3` dependency to `go.mod` (using `go get github.com/zeebo/xxh3`).
- [x] 2.2 Implement a deterministic `xxh3` 128-bit hashing function for metadata objects to produce `uuid.UUID` values.
- [x] 2.3 Add an in-memory cache (using `sync.Map`) to track registered `uuid.UUID` metric identifiers and skip redundant metadata writes to ClickHouse.
- [x] 2.4 Update the mapping models (`GaugeRow`, `SumRow`) in `clickhouse_client.go` and the mapping logic in `metrics_mapper.go` to split metadata from datapoints.

## 3. Storage Ingestion updates

- [x] 3.1 Update `InsertGauge` and `InsertSum` in `clickhouse_client.go` to insert new metadata rows if their `MetricID` is uncached, then write the data-point batches referencing the `MetricID`.

## 4. Tests and Verification

- [x] 4.1 Update the unit test suite in `server_test.go` and integration tests in `integration_test.go` to work with the updated schema and models.
- [x] 4.2 Run tests (`go test ./...`) to verify code compiles and tests pass.
- [x] 4.3 Execute the performance benchmark (`make benchmark-run`) to verify end-to-end telemetrygen ingestion and database write rates.
- [x] 4.4 Create a benchmark report (`benchmark_report.md`) comparing the performance of the baseline solution vs the split metadata solution.
