## Context

The current performance benchmark uses `telemetrygen` to ingest metrics. However, telemetrygen generates only a single unique timeseries, leading to a single metadata row in `otel_metrics_metadata` during our run. This is not representative of real-world scenarios. Additionally, we lack metrics showing the physical storage utilization (compressed/uncompressed size) of our ClickHouse tables.

## Goals / Non-Goals

**Goals:**
* Configure `telemetrygen` to generate unique timeseries over the benchmark run.
* Add database storage size tracking (rows, compressed size, uncompressed size) per table to the analysis script and benchmark report.

**Non-Goals:**
* Changing the ingestion schema or Go implementation.
* Altering test container configs for standard integration tests.

## Decisions

* **Configure `--unique-timeseries`**: Set `--unique-timeseries` and `--unique-timeseries-duration 1s` flags in `docker-compose.yaml` for both `telemetrygen-gauge` and `telemetrygen-sum`. This forces telemetrygen to vary its resource/scope/metric attributes once every second, producing a low-cardinality set of ~10 metadata rows over a 10s benchmark run.
* **Query `system.parts`**: Retrieve table space metrics using ClickHouse system table:
  ```sql
  SELECT table, sum(rows) as rows, sum(data_compressed_bytes) as compressed, sum(data_uncompressed_bytes) as uncompressed FROM system.parts WHERE database = 'default' AND active GROUP BY table
  ```
  This is a lightweight query that fetches accurate physical disk usage details.

## Risks / Trade-offs

* **[Risk]** Unique timeseries flag increases telemetrygen client memory and CPU usage. → **[Mitigation]** We keep the rate at 2,500/s for 10s, which is well within normal testing limits.
