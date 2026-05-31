## 1. Schema Migration

- [x] 1.1 Update `clickhouse_schema.go` to modify `createHistogramTableSQL` removing duplicate metadata columns and adding `MetricID UUID CODEC(ZSTD(1))`. Update `ORDER BY` to `(MetricID, toUnixTimestamp64Nano(TimeUnix))` and remove bloom filter indexes.
- [x] 1.2 Update `clickhouse_schema.go` to modify `createExponentialHistogramTableSQL` removing duplicate metadata columns and adding `MetricID UUID CODEC(ZSTD(1))`. Update `ORDER BY` to `(MetricID, toUnixTimestamp64Nano(TimeUnix))` and remove bloom filter indexes.
- [x] 1.3 Update `clickhouse_schema.go` to modify `createSummaryTableSQL` removing duplicate metadata columns and adding `MetricID UUID CODEC(ZSTD(1))`. Update `ORDER BY` to `(MetricID, toUnixTimestamp64Nano(TimeUnix))` and remove bloom filter indexes.


## 2. Validation & Testing

- [x] 2.1 Run tests (`go test ./...`) to ensure that all database tables are created successfully and that the server starts without errors.
- [x] 2.2 Verify Gauge/Sum ingestion tests pass successfully with the updated schema definitions running.

