## 1. Benchmark Configuration

- [x] 1.1 Update `docker-compose.yaml` to add `--unique-timeseries` and `--unique-timeseries-duration 1s` flags to `telemetrygen-gauge` and `telemetrygen-sum` configurations.

## 2. Ingestion Size Reporting

- [x] 2.1 Update `benchmark/analyze.sh` to query ClickHouse `system.parts` table and print rows, compressed size, and uncompressed size for each of the tables: `otel_metrics_metadata`, `otel_metrics_gauge`, `otel_metrics_sum`, `otel_metrics_histogram`.

## 3. Execution & Documentation

- [x] 3.1 Reset to the baseline commit, run the tuned benchmark, and record its statistics (including database storage sizes) under "Run 1: Baseline Solution (Wide Tables)" in `benchmark_report.md`.
- [x] 3.2 Restore our split-metadata implementation, run the tuned benchmark, and append its statistics (including database storage sizes) under "Run 2: Split Metadata Solution (Cached Lookups)" in `benchmark_report.md`.
