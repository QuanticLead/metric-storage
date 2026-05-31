## Why

While the application bootstraps the OpenTelemetry SDK and instruments gRPC incoming endpoints, the internal database insertions, queue operations, and background worker jobs are uninstrumented. This makes it impossible to trace the lifecycle of a metrics batch from gRPC ingest to ClickHouse disk, or monitor the internal queue size and database write performance under load.

## What Changes

* **Trace Propagation & Context Correlation**: Trace the gRPC `Export` handler execution and background worker batches. Log statements within these contexts will automatically be correlated with `trace_id` and `span_id` due to `otelslog` integration. We will use a detached context tracing pattern to avoid cancelling database writes when gRPC client requests terminate.
* **ClickHouse Query Tracing**: Pass the traced contexts to ClickHouse client operations to enable the database driver's native OpenTelemetry span mapping.
* **Centralized Metric Instrumentation**:
  * Monitor the ingestion queue size dynamically using `com.dash0.homeexercise.ingestion.queue_length` (observable gauge).
  * Track background worker concurrency using `com.dash0.homeexercise.workers.active` (updowncounter).
  * Track in-memory metadata cache performance via `com.dash0.homeexercise.metadata_cache.hits` and `com.dash0.homeexercise.metadata_cache.misses` (counters).

## Capabilities

## New Capabilities

* `opentelemetry-telemetry`: The metrics processor server SHALL report complete internal traces and metrics representing request handling, queue status, caching, and worker state.

## Modified Capabilities

*(None)*

## Impact

* `server.go` (defining tracer, worker instrumentation, and new metrics)
* `metrics_service.go` (tracing the `Export` method)
* `clickhouse_client.go` (propagating context and incrementing metadata cache metrics)
