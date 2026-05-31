## Why

The current benchmark run produces a single metadata entry because telemetrygen sends static resource/scope attributes, which does not reflect realistic production environments. In addition, the analysis report does not capture the actual storage size (compressed and uncompressed bytes) of the ClickHouse tables, which is critical for verifying the storage savings of the split metadata schema.

## What Changes

* **Realistic Metric Generation**: Configure `telemetrygen` with `--unique-timeseries` and `--unique-timeseries-duration 1s` to generate a dynamic, low-cardinality set of ~10 metadata records instead of a single one.
* **Storage Analysis**: Extend `benchmark/analyze.sh` to query ClickHouse system tables and output the rows, compressed bytes, and uncompressed bytes for each table.
* **Report Enrichment**: Update the benchmark report format in `benchmark_report.md` to include database storage size statistics.

## Capabilities

### New Capabilities

*(None)*

### Modified Capabilities

* `performance-benchmark`: Update the performance benchmarking suite to measure and report database storage sizes and support realistic telemetrygen load.

## Impact

* `docker-compose.yaml` (telemetrygen flags)
* `benchmark/analyze.sh` (ClickHouse storage queries)
* `benchmark_report.md` (report schema/attributes)
